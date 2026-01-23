package stats

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/api"
)

// LineCounter 代码行统计器
type LineCounter struct {
	cache sync.Map
}

// LineCacheItem 缓存项
type LineCacheItem struct {
	Status    string               // processing, ready, failed
	Data      *api.LineStatsResponse
	Error     error
	CreatedAt time.Time
	Progress  string
}

// CountConfig 统计配置
type CountConfig struct {
	ExcludeDirs     []string
	ExcludePatterns []string
	ExcludeHidden   bool
	Branch          string // 分支名
	Author          string // 作者筛选
	Since           string // 开始时间 YYYY-MM-DD
	Until           string // 结束时间 YYYY-MM-DD
}

// FileStats 单文件统计
type FileStats struct {
	Path     string
	Language string
	Code     int
	Comment  int
	Blank    int
	Total    int
}

// CodeStats 汇总统计结果
type CodeStats struct {
	TotalFiles   int
	TotalLines   int
	CodeLines    int
	CommentLines int
	BlankLines   int
	ByLanguage   map[string]*api.LanguageStat
}

var lineCounterInstance *LineCounter
var lineCounterOnce sync.Once

// GetLineCounter 获取单例
func GetLineCounter() *LineCounter {
	lineCounterOnce.Do(func() {
		lineCounterInstance = &LineCounter{}
	})
	return lineCounterInstance
}

// GetLineStats 获取代码行统计（支持异步和缓存）
func (lc *LineCounter) GetLineStats(repoPath string, config CountConfig) *api.LineStatsResponse {
	cacheKey := lc.generateCacheKey(repoPath, config)

	// 检查缓存
	if cached, ok := lc.cache.Load(cacheKey); ok {
		item := cached.(*LineCacheItem)
		// 缓存有效期 1 小时
		if time.Since(item.CreatedAt) < time.Hour {
			if item.Status == "processing" {
				return &api.LineStatsResponse{
					Status:   "processing",
					Progress: item.Progress,
				}
			}
			if item.Status == "ready" && item.Data != nil {
				return item.Data
			}
		}
	}

	// 启动后台计算
	cacheItem := &LineCacheItem{
		Status:    "processing",
		CreatedAt: time.Now(),
		Progress:  "正在扫描文件...",
	}
	lc.cache.Store(cacheKey, cacheItem)

	go lc.computeStats(repoPath, config, cacheKey)

	return &api.LineStatsResponse{
		Status:   "processing",
		Progress: "正在扫描文件...",
	}
}

// computeStats 后台计算统计
func (lc *LineCounter) computeStats(repoPath string, config CountConfig, cacheKey string) {
	stats, err := lc.CountLines(repoPath, config)
	if err != nil {
		lc.cache.Store(cacheKey, &LineCacheItem{
			Status:    "failed",
			Error:     err,
			CreatedAt: time.Now(),
		})
		return
	}

	// 转换为响应格式
	languages := make([]*api.LanguageStat, 0, len(stats.ByLanguage))
	for _, lang := range stats.ByLanguage {
		languages = append(languages, lang)
	}

	// 按代码行数排序
	sort.Slice(languages, func(i, j int) bool {
		return languages[i].Code > languages[j].Code
	})

	response := &api.LineStatsResponse{
		Status:       "ready",
		TotalFiles:   stats.TotalFiles,
		TotalLines:   stats.TotalLines,
		CodeLines:    stats.CodeLines,
		CommentLines: stats.CommentLines,
		BlankLines:   stats.BlankLines,
		Languages:    languages,
	}

	lc.cache.Store(cacheKey, &LineCacheItem{
		Status:    "ready",
		Data:      response,
		CreatedAt: time.Now(),
	})
}

// CountLines 统计目录下所有代码行
func (lc *LineCounter) CountLines(rootPath string, config CountConfig) (*CodeStats, error) {
	stats := &CodeStats{
		ByLanguage: make(map[string]*api.LanguageStat),
	}

	// 预处理排除模式
	excludeDirs := make(map[string]bool)
	for _, dir := range config.ExcludeDirs {
		excludeDirs[strings.ToLower(dir)] = true
	}

	// 检查是否需要按作者/时间过滤
	needBlameFilter := config.Author != "" || config.Since != "" || config.Until != ""
	isGit := lc.isGitRepo(rootPath)

	// 如果需要过滤但不是git仓库，则忽略过滤条件
	if needBlameFilter && !isGit {
		needBlameFilter = false
	}

	// 如果指定了分支，先切换分支工作树（使用git worktree或读取指定分支的文件）
	// 这里我们使用git blame直接指定分支，不需要切换

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 跳过无法访问的文件
		}

		name := d.Name()

		// 排除隐藏文件/目录
		if config.ExcludeHidden && strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 排除指定目录
		if d.IsDir() {
			if excludeDirs[strings.ToLower(name)] {
				return filepath.SkipDir
			}
			return nil
		}

		// 排除指定模式的文件
		if lc.matchesExcludePattern(name, config.ExcludePatterns) {
			return nil
		}

		// 获取语言配置
		langConfig := GetLanguageConfig(path)
		if langConfig == nil {
			return nil // 不支持的文件类型
		}

		// 获取 blame 信息（如果需要过滤）
		var blameInfo map[int]*BlameLineInfo
		if needBlameFilter {
			blameInfo, _ = lc.getGitBlameInfo(rootPath, path, config.Branch)
			// 如果获取blame失败，继续统计但不过滤
		}

		// 分析文件
		fileStats, err := lc.analyzeFileWithFilter(path, langConfig, blameInfo, config)
		if err != nil {
			return nil // 跳过无法读取的文件
		}

		// 如果过滤后没有任何统计，跳过这个文件
		if fileStats.Total == 0 && needBlameFilter {
			return nil
		}

		// 聚合统计
		stats.TotalFiles++
		stats.TotalLines += fileStats.Total
		stats.CodeLines += fileStats.Code
		stats.CommentLines += fileStats.Comment
		stats.BlankLines += fileStats.Blank

		// 按语言统计
		langStat := stats.ByLanguage[fileStats.Language]
		if langStat == nil {
			langStat = &api.LanguageStat{Name: fileStats.Language}
			stats.ByLanguage[fileStats.Language] = langStat
		}
		langStat.Files++
		langStat.Code += fileStats.Code
		langStat.Comment += fileStats.Comment
		langStat.Blank += fileStats.Blank

		return nil
	})

	return stats, err
}

// analyzeFile 分析单个文件（状态机实现）
func (lc *LineCounter) analyzeFile(filePath string, langConfig *CommentConfig) (*FileStats, error) {
	return lc.analyzeFileWithFilter(filePath, langConfig, nil, CountConfig{})
}

// analyzeFileWithFilter 分析单个文件，支持按blame信息过滤
func (lc *LineCounter) analyzeFileWithFilter(filePath string, langConfig *CommentConfig, blameInfo map[int]*BlameLineInfo, config CountConfig) (*FileStats, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats := &FileStats{
		Path:     filePath,
		Language: langConfig.Language,
	}

	scanner := bufio.NewScanner(file)
	// 增大缓冲区以处理长行
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	inMultiLineComment := false
	lineNum := 0
	needFilter := len(blameInfo) > 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// 检查是否应该统计这一行（基于blame过滤）
		if needFilter {
			info := blameInfo[lineNum]
			if !lc.shouldCountLine(info, config) {
				// 仍然需要跟踪多行注释状态
				if inMultiLineComment {
					if langConfig.MultiLineEnd != "" && strings.Contains(trimmed, langConfig.MultiLineEnd) {
						inMultiLineComment = false
					}
				} else if langConfig.MultiLineStart != "" && strings.Contains(trimmed, langConfig.MultiLineStart) {
					startIdx := strings.Index(trimmed, langConfig.MultiLineStart)
					afterStart := trimmed[startIdx+len(langConfig.MultiLineStart):]
					if langConfig.MultiLineEnd == "" || !strings.Contains(afterStart, langConfig.MultiLineEnd) {
						inMultiLineComment = true
					}
				}
				continue
			}
		}

		// 空白行
		if trimmed == "" {
			stats.Blank++
			stats.Total++
			continue
		}

		stats.Total++

		// 处理多行注释状态
		if inMultiLineComment {
			stats.Comment++
			if langConfig.MultiLineEnd != "" && strings.Contains(trimmed, langConfig.MultiLineEnd) {
				inMultiLineComment = false
			}
			continue
		}

		// 检测多行注释开始
		if langConfig.MultiLineStart != "" && strings.Contains(trimmed, langConfig.MultiLineStart) {
			// 检查是否在同一行结束
			startIdx := strings.Index(trimmed, langConfig.MultiLineStart)
			afterStart := trimmed[startIdx+len(langConfig.MultiLineStart):]

			if langConfig.MultiLineEnd != "" && strings.Contains(afterStart, langConfig.MultiLineEnd) {
				// 同一行结束的注释
				// 检查注释前是否有代码
				beforeComment := strings.TrimSpace(trimmed[:startIdx])
				if beforeComment == "" {
					stats.Comment++
				} else {
					stats.Code++
				}
			} else {
				// 多行注释开始
				inMultiLineComment = true
				// 检查注释前是否有代码
				beforeComment := strings.TrimSpace(trimmed[:startIdx])
				if beforeComment == "" {
					stats.Comment++
				} else {
					stats.Code++
				}
			}
			continue
		}

		// 检测单行注释
		if langConfig.SingleLine != "" {
			commentIdx := lc.findCommentStart(trimmed, langConfig)
			if commentIdx == 0 {
				// 整行都是注释
				stats.Comment++
				continue
			} else if commentIdx > 0 {
				// 有代码也有注释，算作代码行
				stats.Code++
				continue
			}
		}

		// 普通代码行
		stats.Code++
	}

	return stats, scanner.Err()
}

// findCommentStart 查找单行注释开始位置（考虑字符串）
func (lc *LineCounter) findCommentStart(line string, langConfig *CommentConfig) int {
	if langConfig.SingleLine == "" {
		return -1
	}

	inString := false
	stringChar := byte(0)

	for i := 0; i < len(line); i++ {
		ch := line[i]

		// 处理字符串状态
		if inString {
			if ch == stringChar && (i == 0 || line[i-1] != '\\') {
				inString = false
			}
			continue
		}

		// 检测字符串开始
		for _, delim := range langConfig.StringDelimiters {
			if len(delim) == 1 && ch == delim[0] {
				inString = true
				stringChar = ch
				break
			}
		}

		if inString {
			continue
		}

		// 检测单行注释
		if strings.HasPrefix(line[i:], langConfig.SingleLine) {
			return i
		}
	}

	return -1
}

// matchesExcludePattern 检查文件名是否匹配排除模式
func (lc *LineCounter) matchesExcludePattern(filename string, patterns []string) bool {
	lowerFilename := strings.ToLower(filename)
	for _, pattern := range patterns {
		lowerPattern := strings.ToLower(pattern)
		if matched, _ := filepath.Match(lowerPattern, lowerFilename); matched {
			return true
		}
	}
	return false
}

// generateCacheKey 生成缓存键
func (lc *LineCounter) generateCacheKey(repoPath string, config CountConfig) string {
	// 将配置序列化为字符串
	configStr := strings.Join(config.ExcludeDirs, ",") + "|" +
		strings.Join(config.ExcludePatterns, ",") + "|" +
		config.Branch + "|" + config.Author + "|" + config.Since + "|" + config.Until

	hash := md5.Sum([]byte(configStr))
	return "lines:" + repoPath + ":" + hex.EncodeToString(hash[:8])
}

// BlameLineInfo git blame 行信息
type BlameLineInfo struct {
	Author    string
	Email     string
	Timestamp int64
}

// isGitRepo 检查目录是否是 git 仓库
func (lc *LineCounter) isGitRepo(repoPath string) bool {
	gitDir := filepath.Join(repoPath, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// getGitBlameInfo 获取文件的 git blame 信息
func (lc *LineCounter) getGitBlameInfo(repoPath, filePath, branch string) (map[int]*BlameLineInfo, error) {
	// 构建 git blame 命令
	args := []string{"blame", "--line-porcelain"}
	if branch != "" {
		args = append(args, branch, "--")
	}
	
	// 获取相对路径
	relPath, err := filepath.Rel(repoPath, filePath)
	if err != nil {
		relPath = filePath
	}
	args = append(args, relPath)
	
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	return lc.parseBlameOutput(string(output))
}

// parseBlameOutput 解析 git blame --line-porcelain 输出
func (lc *LineCounter) parseBlameOutput(output string) (map[int]*BlameLineInfo, error) {
	result := make(map[int]*BlameLineInfo)
	lines := strings.Split(output, "\n")
	
	var currentLine int
	var currentInfo *BlameLineInfo
	
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		
		// 提交哈希行 (40字符的hex + 行号信息)
		if len(line) >= 40 && isHexString(line[:40]) {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				lineNum, _ := strconv.Atoi(parts[2])
				currentLine = lineNum
				currentInfo = &BlameLineInfo{}
			}
			continue
		}
		
		if currentInfo == nil {
			continue
		}
		
		// 解析作者信息
		if strings.HasPrefix(line, "author ") {
			currentInfo.Author = strings.TrimPrefix(line, "author ")
		} else if strings.HasPrefix(line, "author-mail ") {
			email := strings.TrimPrefix(line, "author-mail ")
			email = strings.Trim(email, "<>")
			currentInfo.Email = email
		} else if strings.HasPrefix(line, "author-time ") {
			timestamp, _ := strconv.ParseInt(strings.TrimPrefix(line, "author-time "), 10, 64)
			currentInfo.Timestamp = timestamp
		} else if strings.HasPrefix(line, "\t") {
			// 代码行内容行，表示当前blame块结束
			if currentLine > 0 && currentInfo != nil {
				result[currentLine] = currentInfo
			}
			currentInfo = nil
		}
	}
	
	return result, nil
}

// isHexString 检查字符串是否为十六进制
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// shouldCountLine 判断某一行是否应该被统计（基于作者和时间过滤）
func (lc *LineCounter) shouldCountLine(info *BlameLineInfo, config CountConfig) bool {
	if info == nil {
		return true // 无blame信息时默认统计
	}
	
	// 作者过滤
	if config.Author != "" {
		authorMatch := strings.Contains(strings.ToLower(info.Author), strings.ToLower(config.Author)) ||
			strings.Contains(strings.ToLower(info.Email), strings.ToLower(config.Author))
		if !authorMatch {
			return false
		}
	}
	
	// 时间范围过滤
	if config.Since != "" {
		sinceTime, err := time.Parse("2006-01-02", config.Since)
		if err == nil && info.Timestamp < sinceTime.Unix() {
			return false
		}
	}
	
	if config.Until != "" {
		untilTime, err := time.Parse("2006-01-02", config.Until)
		if err == nil {
			// until 日期包含当天，所以加1天
			untilTime = untilTime.Add(24 * time.Hour)
			if info.Timestamp >= untilTime.Unix() {
				return false
			}
		}
	}
	
	return true
}

// ClearCache 清除缓存
func (lc *LineCounter) ClearCache(repoPath string) {
	// 遍历并删除匹配的缓存
	lc.cache.Range(func(key, value interface{}) bool {
		if strings.HasPrefix(key.(string), "lines:"+repoPath+":") {
			lc.cache.Delete(key)
		}
		return true
	})
}
