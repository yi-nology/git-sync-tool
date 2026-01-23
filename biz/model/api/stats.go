package api

type AuthorStat struct {
	Name       string         `json:"name"`
	Email      string         `json:"email"`
	TotalLines int            `json:"total_lines"`
	FileTypes  map[string]int `json:"file_types"`
	// Date -> Lines mapping (e.g. "2023-01-01" -> 10)
	TimeTrend map[string]int `json:"time_trend"`
}

type StatsResponse struct {
	TotalLines int           `json:"total_lines"`
	Authors    []*AuthorStat `json:"authors"`
}

// LanguageStat 单个编程语言的统计信息
type LanguageStat struct {
	Name    string `json:"name"`
	Files   int    `json:"files"`
	Code    int    `json:"code"`
	Comment int    `json:"comment"`
	Blank   int    `json:"blank"`
}

// LineStatsResponse 代码行统计响应
type LineStatsResponse struct {
	Status       string         `json:"status,omitempty"`        // processing, ready, failed
	Progress     string         `json:"progress,omitempty"`      // 进度信息
	TotalFiles   int            `json:"total_files,omitempty"`   // 文件总数
	TotalLines   int            `json:"total_lines,omitempty"`   // 总行数
	CodeLines    int            `json:"code_lines,omitempty"`    // 代码行数
	CommentLines int            `json:"comment_lines,omitempty"` // 注释行数
	BlankLines   int            `json:"blank_lines,omitempty"`   // 空白行数
	Languages    []*LanguageStat `json:"languages,omitempty"`    // 按语言统计
}

// LineStatsConfig 排除配置
type LineStatsConfig struct {
	ExcludeDirs     []string `json:"exclude_dirs"`
	ExcludePatterns []string `json:"exclude_patterns"`
}

// LineStatsConfigRequest 配置请求
type LineStatsConfigRequest struct {
	RepoKey         string   `json:"repo_key"`
	ExcludeDirs     []string `json:"exclude_dirs"`
	ExcludePatterns []string `json:"exclude_patterns"`
}
