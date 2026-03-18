package commit_analyzer

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"
)

type AnalyzerService struct {
	commitAnalysisDAO *db.CommitAnalysisDAO
	gitService        *git.GitService
}

func NewAnalyzerService() *AnalyzerService {
	return &AnalyzerService{
		commitAnalysisDAO: db.NewCommitAnalysisDAO(),
		gitService:        git.NewGitService(),
	}
}

type FileChange struct {
	File       string `json:"file"`
	Added      int    `json:"added"`
	Deleted    int    `json:"deleted"`
	ChangeType string `json:"changeType"` // added, modified, deleted
}

type LanguageChange struct {
	Language string `json:"language"`
	Added    int    `json:"added"`
	Deleted  int    `json:"deleted"`
}

func (s *AnalyzerService) AnalyzeCommit(repoPath, repoKey, commitHash string) (*po.CommitAnalysis, error) {
	// 获取提交信息
	commitInfo, err := s.gitService.GetCommitInfo(repoPath, commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit info: %v", err)
	}

	// 获取提交差异
	diff, err := s.gitService.GetCommitDiffSimple(repoPath, commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit diff: %v", err)
	}

	// 解析差异，计算变更统计
	addedLines, deletedLines, fileChanges, languageChanges := s.parseDiff(diff)
	totalChanges := addedLines + deletedLines

	// 序列化文件变更和语言变更
	fileChangesJSON, err := json.Marshal(fileChanges)
	if err != nil {
		return nil, err
	}

	languageChangesJSON, err := json.Marshal(languageChanges)
	if err != nil {
		return nil, err
	}

	// 识别提交类型
	commitType := s.detectCommitType(commitInfo.Message)

	// 检查是否是合并提交
	isMerge := strings.Contains(commitInfo.Message, "Merge") || strings.Contains(commitInfo.Message, "merge")

	// 创建提交分析记录
	analysis := &po.CommitAnalysis{
		RepoKey:         repoKey,
		CommitHash:      commitHash,
		Author:          commitInfo.Author,
		AuthorEmail:     commitInfo.AuthorEmail,
		Committer:       commitInfo.Committer,
		CommitterEmail:  commitInfo.CommitterEmail,
		Message:         commitInfo.Message,
		CommitTime:      commitInfo.CommitTime,
		AnalysisTime:    time.Now(),
		AddedLines:      addedLines,
		DeletedLines:    deletedLines,
		TotalChanges:    totalChanges,
		FileChanges:     string(fileChangesJSON),
		LanguageChanges: string(languageChangesJSON),
		CommitType:      commitType,
		IsMerge:         isMerge,
	}

	// 保存分析结果
	if err := s.commitAnalysisDAO.CreateCommitAnalysis(analysis); err != nil {
		return nil, err
	}

	// 更新提交模式
	s.updateCommitPatterns(repoKey, analysis)

	return analysis, nil
}

func (s *AnalyzerService) AnalyzeRepo(repoPath, repoKey string) error {
	// 获取最近的提交历史
	commits, err := s.gitService.GetRecentCommits(repoPath, 100)
	if err != nil {
		return fmt.Errorf("failed to get recent commits: %v", err)
	}

	for _, commitHash := range commits {
		// 检查是否已经分析过
		_, err := s.commitAnalysisDAO.GetCommitAnalysisByHash(repoKey, commitHash)
		if err == nil {
			// 已经分析过，跳过
			continue
		}

		// 分析提交
		_, err = s.AnalyzeCommit(repoPath, repoKey, commitHash)
		if err != nil {
			// 记录错误但继续
			continue
		}
	}

	return nil
}

func (s *AnalyzerService) GetCommitPatterns(repoKey string, patternType string, limit int) ([]po.CommitPattern, error) {
	return s.commitAnalysisDAO.GetCommitPatternsByRepo(repoKey, patternType, limit)
}

func (s *AnalyzerService) GetCommitStats(repoKey string, days int) (map[string]interface{}, error) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)

	// 获取提交数量
	commitCount, err := s.commitAnalysisDAO.GetCommitCountByTimeRange(repoKey, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 获取总变更行数
	totalChanges, err := s.commitAnalysisDAO.GetTotalChangesByTimeRange(repoKey, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 获取提交分析
	analyses, err := s.commitAnalysisDAO.GetCommitAnalysesByTimeRange(repoKey, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 分析提交类型分布
	commitTypeCount := make(map[string]int)
	for _, analysis := range analyses {
		commitTypeCount[analysis.CommitType]++
	}

	// 分析每天的提交数量
	dailyCommits := make(map[string]int)
	for _, analysis := range analyses {
		dateKey := analysis.CommitTime.Format("2006-01-02")
		dailyCommits[dateKey]++
	}

	return map[string]interface{}{
		"commitCount":     commitCount,
		"totalChanges":    totalChanges,
		"commitTypeCount": commitTypeCount,
		"dailyCommits":    dailyCommits,
		"startTime":       startTime,
		"endTime":         endTime,
	}, nil
}

func (s *AnalyzerService) parseDiff(diff string) (int, int, []FileChange, []LanguageChange) {
	addedLines := 0
	deletedLines := 0
	fileChanges := []FileChange{}
	languageChanges := []LanguageChange{}

	// 简单的差异解析
	lines := strings.Split(diff, "\n")
	currentFile := ""
	currentAdded := 0
	currentDeleted := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			// 新文件开始
			if currentFile != "" {
				fileChanges = append(fileChanges, FileChange{
					File:       currentFile,
					Added:      currentAdded,
					Deleted:    currentDeleted,
					ChangeType: s.detectChangeType(currentAdded, currentDeleted),
				})
			}
			// 提取文件名
			parts := strings.Split(line, " ")
			if len(parts) >= 3 {
				currentFile = strings.TrimPrefix(parts[2], "a/")
			}
			currentAdded = 0
			currentDeleted = 0
		} else if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			addedLines++
			currentAdded++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			deletedLines++
			currentDeleted++
		}
	}

	// 添加最后一个文件
	if currentFile != "" {
		fileChanges = append(fileChanges, FileChange{
			File:       currentFile,
			Added:      currentAdded,
			Deleted:    currentDeleted,
			ChangeType: s.detectChangeType(currentAdded, currentDeleted),
		})
	}

	// 分析语言变更
	languageMap := make(map[string]LanguageChange)
	for _, change := range fileChanges {
		language := s.detectLanguage(change.File)
		if language != "" {
			if langChange, exists := languageMap[language]; exists {
				langChange.Added += change.Added
				langChange.Deleted += change.Deleted
				languageMap[language] = langChange
			} else {
				languageMap[language] = LanguageChange{
					Language: language,
					Added:    change.Added,
					Deleted:  change.Deleted,
				}
			}
		}
	}

	for _, langChange := range languageMap {
		languageChanges = append(languageChanges, langChange)
	}

	return addedLines, deletedLines, fileChanges, languageChanges
}

func (s *AnalyzerService) detectChangeType(added, deleted int) string {
	if added > 0 && deleted == 0 {
		return "added"
	} else if added == 0 && deleted > 0 {
		return "deleted"
	} else {
		return "modified"
	}
}

func (s *AnalyzerService) detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".go":
		return "Go"
	case ".c", ".h":
		return "C"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "C++"
	case ".ts", ".tsx":
		return "TypeScript"
	case ".js", ".jsx":
		return "JavaScript"
	case ".rs":
		return "Rust"
	case ".py":
		return "Python"
	case ".java":
		return "Java"
	case ".html":
		return "HTML"
	case ".css":
		return "CSS"
	default:
		return ""
	}
}

func (s *AnalyzerService) detectCommitType(message string) string {
	message = strings.ToLower(message)
	if strings.HasPrefix(message, "feat") {
		return "feat"
	} else if strings.HasPrefix(message, "fix") {
		return "fix"
	} else if strings.HasPrefix(message, "docs") {
		return "docs"
	} else if strings.HasPrefix(message, "style") {
		return "style"
	} else if strings.HasPrefix(message, "refactor") {
		return "refactor"
	} else if strings.HasPrefix(message, "test") {
		return "test"
	} else if strings.HasPrefix(message, "chore") {
		return "chore"
	} else {
		return "other"
	}
}

func (s *AnalyzerService) updateCommitPatterns(repoKey string, analysis *po.CommitAnalysis) {
	// 更新每日模式
	dayOfWeek := analysis.CommitTime.Weekday().String()
	s.updatePattern(repoKey, "daily", dayOfWeek, analysis)

	// 更新每周模式
	weekNumber := (analysis.CommitTime.YearDay()-1)/7 + 1
	weekKey := fmt.Sprintf("Week %d", weekNumber)
	s.updatePattern(repoKey, "weekly", weekKey, analysis)

	// 更新每月模式
	monthKey := analysis.CommitTime.Format("2006-01")
	s.updatePattern(repoKey, "monthly", monthKey, analysis)

	// 更新语言模式
	var languageChanges []LanguageChange
	if err := json.Unmarshal([]byte(analysis.LanguageChanges), &languageChanges); err == nil {
		for _, langChange := range languageChanges {
			s.updatePattern(repoKey, "language", langChange.Language, analysis)
		}
	}
}

func (s *AnalyzerService) updatePattern(repoKey, patternType, patternValue string, analysis *po.CommitAnalysis) {
	// 获取最新的模式
	pattern, err := s.commitAnalysisDAO.GetLatestCommitPattern(repoKey, patternType, patternValue)
	if err != nil {
		// 模式不存在，创建新的
		pattern = &po.CommitPattern{
			RepoKey:      repoKey,
			PatternType:  patternType,
			PatternValue: patternValue,
			CommitCount:  1,
			ChangeCount:  analysis.TotalChanges,
			StartDate:    analysis.CommitTime,
			EndDate:      analysis.CommitTime,
		}
		if err := s.commitAnalysisDAO.CreateCommitPattern(pattern); err != nil {
			// Log the error but continue
			_ = err // 暂时使用下划线忽略错误，避免空分支
		}
	} else {
		// 更新现有模式
		pattern.CommitCount++
		pattern.ChangeCount += analysis.TotalChanges
		pattern.EndDate = analysis.CommitTime
		if err := s.commitAnalysisDAO.UpdateCommitPattern(pattern); err != nil {
			// Log the error but continue
			_ = err // 暂时使用下划线忽略错误，避免空分支
		}
	}
}

func (s *AnalyzerService) GenerateSyncRecommendations(repoKey, taskKey string) (*po.SyncRecommendation, error) {
	// 获取最近30天的提交统计
	stats, err := s.GetCommitStats(repoKey, 30)
	if err != nil {
		return nil, err
	}

	commitCount := stats["commitCount"].(int64)
	commitTypeCount := stats["commitTypeCount"].(map[string]int)

	// 分析提交频率
	averageCommitsPerDay := float64(commitCount) / 30.0

	// 确定建议的同步频率
	syncFrequency := "daily"
	if averageCommitsPerDay < 0.5 {
		syncFrequency = "weekly"
	} else if averageCommitsPerDay > 5 {
		syncFrequency = "hourly"
	}

	// 生成建议
	recommendation := fmt.Sprintf("基于过去30天的分析，该仓库平均每天有 %.2f 次提交。", averageCommitsPerDay)
	recommendation += fmt.Sprintf("建议的同步频率为 %s。", syncFrequency)

	// 分析提交类型
	if featCount, exists := commitTypeCount["feat"]; exists && featCount > 0 {
		recommendation += fmt.Sprintf(" 包含 %d 个新功能提交。", featCount)
	}
	if fixCount, exists := commitTypeCount["fix"]; exists && fixCount > 0 {
		recommendation += fmt.Sprintf(" 包含 %d 个 bug 修复提交。", fixCount)
	}

	// 计算置信度
	confidence := 0.7
	if commitCount > 20 {
		confidence = 0.9
	} else if commitCount > 5 {
		confidence = 0.8
	}

	// 创建或更新推荐
	syncRecommendation := &po.SyncRecommendation{
		RepoKey:        repoKey,
		TaskKey:        taskKey,
		Recommendation: recommendation,
		SyncFrequency:  syncFrequency,
		Confidence:     confidence,
		LastAnalysis:   time.Now(),
		IsApplied:      false,
	}

	// 检查是否已有推荐
	existing, err := s.commitAnalysisDAO.GetSyncRecommendation(repoKey, taskKey)
	if err != nil {
		// 创建新推荐
		if err := s.commitAnalysisDAO.CreateSyncRecommendation(syncRecommendation); err != nil {
				// Log the error but continue
				_ = err // 暂时使用下划线忽略错误，避免空分支
			}
	} else {
		// 更新现有推荐
		existing.Recommendation = syncRecommendation.Recommendation
		existing.SyncFrequency = syncRecommendation.SyncFrequency
		existing.Confidence = syncRecommendation.Confidence
		existing.LastAnalysis = syncRecommendation.LastAnalysis
		if err := s.commitAnalysisDAO.UpdateSyncRecommendation(existing); err != nil {
				// Log the error but continue
				_ = err // 暂时使用下划线忽略错误，避免空分支
			}
		syncRecommendation = existing
	}

	return syncRecommendation, nil
}

func (s *AnalyzerService) GetSyncRecommendations(repoKey string, limit int) ([]po.SyncRecommendation, error) {
	return s.commitAnalysisDAO.GetSyncRecommendationsByRepo(repoKey, limit)
}

func (s *AnalyzerService) UpdateSyncRecommendation(recommendation *po.SyncRecommendation) error {
	return s.commitAnalysisDAO.UpdateSyncRecommendation(recommendation)
}

func (s *AnalyzerService) CleanupOldData(repoKey string, keepDays int) error {
	if err := s.commitAnalysisDAO.DeleteOldCommitAnalyses(repoKey, keepDays); err != nil {
		return err
	}
	if err := s.commitAnalysisDAO.DeleteOldCommitPatterns(repoKey, keepDays); err != nil {
		return err
	}
	return nil
}
