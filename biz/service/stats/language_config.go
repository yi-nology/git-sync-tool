package stats

import (
	"path/filepath"
	"strings"
)

// CommentConfig 注释配置
type CommentConfig struct {
	Language         string   // 语言名称
	Extensions       []string // 文件扩展名
	SingleLine       string   // 单行注释符号 (如 //)
	MultiLineStart   string   // 多行注释开始 (如 /*)
	MultiLineEnd     string   // 多行注释结束 (如 */)
	StringDelimiters []string // 字符串定界符
}

// LanguageConfigs 预定义语言配置
var LanguageConfigs = []CommentConfig{
	{
		Language:         "Go",
		Extensions:       []string{".go"},
		SingleLine:       "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\"", "`"},
	},
	{
		Language:         "Java",
		Extensions:       []string{".java"},
		SingleLine:       "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\""},
	},
	{
		Language:         "JavaScript",
		Extensions:       []string{".js", ".jsx", ".mjs"},
		SingleLine:       "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\"", "'", "`"},
	},
	{
		Language:         "TypeScript",
		Extensions:       []string{".ts", ".tsx"},
		SingleLine:       "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\"", "'", "`"},
	},
	{
		Language:         "Python",
		Extensions:       []string{".py"},
		SingleLine:       "#",
		MultiLineStart:   `"""`,
		MultiLineEnd:     `"""`,
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "C/C++",
		Extensions:       []string{".c", ".h", ".cpp", ".hpp", ".cc", ".cxx", ".hh"},
		SingleLine:       "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\""},
	},
	{
		Language:         "C#",
		Extensions:       []string{".cs"},
		SingleLine:       "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\""},
	},
	{
		Language:         "Rust",
		Extensions:       []string{".rs"},
		SingleLine:       "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\""},
	},
	{
		Language:         "Swift",
		Extensions:       []string{".swift"},
		SingleLine:       "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\""},
	},
	{
		Language:         "Kotlin",
		Extensions:       []string{".kt", ".kts"},
		SingleLine:       "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\""},
	},
	{
		Language:         "HTML",
		Extensions:       []string{".html", ".htm"},
		SingleLine:       "",
		MultiLineStart:   "<!--",
		MultiLineEnd:     "-->",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "XML",
		Extensions:       []string{".xml", ".xsl", ".xslt"},
		SingleLine:       "",
		MultiLineStart:   "<!--",
		MultiLineEnd:     "-->",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "CSS",
		Extensions:       []string{".css"},
		SingleLine:       "",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "SCSS",
		Extensions:       []string{".scss", ".sass", ".less"},
		SingleLine:       "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "SQL",
		Extensions:       []string{".sql"},
		SingleLine:       "--",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "Shell",
		Extensions:       []string{".sh", ".bash", ".zsh"},
		SingleLine:       "#",
		MultiLineStart:   "",
		MultiLineEnd:     "",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "PowerShell",
		Extensions:       []string{".ps1", ".psm1"},
		SingleLine:       "#",
		MultiLineStart:   "<#",
		MultiLineEnd:     "#>",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "Ruby",
		Extensions:       []string{".rb"},
		SingleLine:       "#",
		MultiLineStart:   "=begin",
		MultiLineEnd:     "=end",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "PHP",
		Extensions:       []string{".php"},
		SingleLine:       "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "Perl",
		Extensions:       []string{".pl", ".pm"},
		SingleLine:       "#",
		MultiLineStart:   "=pod",
		MultiLineEnd:     "=cut",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "Lua",
		Extensions:       []string{".lua"},
		SingleLine:       "--",
		MultiLineStart:   "--[[",
		MultiLineEnd:     "]]",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "R",
		Extensions:       []string{".r", ".R"},
		SingleLine:       "#",
		MultiLineStart:   "",
		MultiLineEnd:     "",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "Scala",
		Extensions:       []string{".scala"},
		SingleLine:       "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\""},
	},
	{
		Language:         "YAML",
		Extensions:       []string{".yml", ".yaml"},
		SingleLine:       "#",
		MultiLineStart:   "",
		MultiLineEnd:     "",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "TOML",
		Extensions:       []string{".toml"},
		SingleLine:       "#",
		MultiLineStart:   "",
		MultiLineEnd:     "",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "JSON",
		Extensions:       []string{".json"},
		SingleLine:       "",
		MultiLineStart:   "",
		MultiLineEnd:     "",
		StringDelimiters: []string{"\""},
	},
	{
		Language:         "Markdown",
		Extensions:       []string{".md", ".markdown"},
		SingleLine:       "",
		MultiLineStart:   "",
		MultiLineEnd:     "",
		StringDelimiters: []string{},
	},
	{
		Language:         "Dockerfile",
		Extensions:       []string{"Dockerfile"},
		SingleLine:       "#",
		MultiLineStart:   "",
		MultiLineEnd:     "",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "Makefile",
		Extensions:       []string{"Makefile", ".mk"},
		SingleLine:       "#",
		MultiLineStart:   "",
		MultiLineEnd:     "",
		StringDelimiters: []string{"\"", "'"},
	},
	{
		Language:         "Protobuf",
		Extensions:       []string{".proto"},
		SingleLine:       "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\""},
	},
	{
		Language:         "Thrift",
		Extensions:       []string{".thrift"},
		SingleLine:       "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		StringDelimiters: []string{"\""},
	},
	{
		Language:         "Vue",
		Extensions:       []string{".vue"},
		SingleLine:       "//",
		MultiLineStart:   "<!--",
		MultiLineEnd:     "-->",
		StringDelimiters: []string{"\"", "'", "`"},
	},
	{
		Language:         "Svelte",
		Extensions:       []string{".svelte"},
		SingleLine:       "//",
		MultiLineStart:   "<!--",
		MultiLineEnd:     "-->",
		StringDelimiters: []string{"\"", "'", "`"},
	},
}

// extensionToConfig 扩展名到配置的映射（用于快速查找）
var extensionToConfig map[string]*CommentConfig

func init() {
	extensionToConfig = make(map[string]*CommentConfig)
	for i := range LanguageConfigs {
		cfg := &LanguageConfigs[i]
		for _, ext := range cfg.Extensions {
			extensionToConfig[strings.ToLower(ext)] = cfg
		}
	}
}

// GetLanguageConfig 根据文件路径获取语言配置
func GetLanguageConfig(filePath string) *CommentConfig {
	// 先尝试扩展名
	ext := strings.ToLower(filepath.Ext(filePath))
	if cfg, ok := extensionToConfig[ext]; ok {
		return cfg
	}

	// 再尝试文件名（如 Dockerfile, Makefile）
	filename := filepath.Base(filePath)
	if cfg, ok := extensionToConfig[filename]; ok {
		return cfg
	}

	return nil
}

// GetSupportedExtensions 获取所有支持的扩展名
func GetSupportedExtensions() []string {
	var exts []string
	for ext := range extensionToConfig {
		exts = append(exts, ext)
	}
	return exts
}

// DefaultExcludeDirs 默认排除目录
var DefaultExcludeDirs = []string{
	"node_modules",
	"vendor",
	".git",
	".svn",
	".hg",
	"dist",
	"build",
	"target",
	"out",
	"bin",
	"obj",
	".idea",
	".vscode",
	"__pycache__",
	".pytest_cache",
	".mypy_cache",
	"coverage",
	".coverage",
	".nyc_output",
	"bower_components",
	"jspm_packages",
	".gradle",
	".maven",
	"Pods",
	"DerivedData",
}

// DefaultExcludePatterns 默认排除文件模式
var DefaultExcludePatterns = []string{
	"*.min.js",
	"*.min.css",
	"*.map",
	"*.lock",
	"*.sum",
	"*.pb.go",
	"*.gen.go",
	"*_generated.go",
	"*.generated.*",
	"package-lock.json",
	"yarn.lock",
	"pnpm-lock.yaml",
	"Cargo.lock",
	"Gemfile.lock",
	"poetry.lock",
	"composer.lock",
}
