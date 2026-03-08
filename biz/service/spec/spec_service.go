package spec

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SpecService spec 服务
type SpecService struct{}

// NewSpecService 创建 spec 服务
func NewSpecService() *SpecService {
	return &SpecService{}
}

// ListSpecFiles 列出仓库中的所有 .spec 文件
func (s *SpecService) ListSpecFiles(repoPath string) ([]SpecFileInfo, error) {
	var files []SpecFileInfo

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过 .git 目录
		if strings.Contains(path, ".git") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 只收集 .spec 文件
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".spec") {
			relPath, _ := filepath.Rel(repoPath, path)
			files = append(files, SpecFileInfo{
				Name:    info.Name(),
				Path:    relPath,
				IsDir:   false,
				Size:    info.Size(),
				ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
			})
		}

		return nil
	})

	return files, err
}

// GetSpecContent 获取 spec 文件内容
func (s *SpecService) GetSpecContent(repoPath, specPath string) (string, error) {
	fullPath := filepath.Join(repoPath, specPath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read spec file: %v", err)
	}
	return string(content), nil
}

// SaveSpecContent 保存 spec 文件内容
func (s *SpecService) SaveSpecContent(repoPath, specPath, content, commitMessage string) error {
	fullPath := filepath.Join(repoPath, specPath)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write spec file: %v", err)
	}

	return nil
}

// CreateSpecFile 创建新的 spec 文件
func (s *SpecService) CreateSpecFile(repoPath, dirPath, fileName string) (string, error) {
	fullDir := filepath.Join(repoPath, dirPath)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	fullPath := filepath.Join(fullDir, fileName)
	if _, err := os.Stat(fullPath); err == nil {
		return "", fmt.Errorf("file already exists: %s", fileName)
	}

	// 写入模板内容
	template := s.GetSpecTemplate()
	if err := os.WriteFile(fullPath, []byte(template), 0644); err != nil {
		return "", fmt.Errorf("failed to create spec file: %v", err)
	}

	relPath, _ := filepath.Rel(repoPath, fullPath)
	return relPath, nil
}

// DeleteSpecFile 删除 spec 文件
func (s *SpecService) DeleteSpecFile(repoPath, specPath string) error {
	fullPath := filepath.Join(repoPath, specPath)
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete spec file: %v", err)
	}
	return nil
}

// ValidateSpec 验证 spec 文件（基于规则引擎）
func (s *SpecService) ValidateSpec(content string) SpecValidationResult {
	rules := s.GetBuiltinRules()
	var issues []SpecIssue
	var warnings []SpecIssue

	lines := strings.Split(content, "\n")

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// 应用规则检查
		issues = append(issues, s.applyRule(lines, rule)...)
	}

	// 分离 error 和 warning
	for _, issue := range issues {
		if issue.Severity == "error" {
			issues = append(issues, issue)
		} else {
			warnings = append(warnings, issue)
		}
	}

	return SpecValidationResult{
		Valid:    len(issues) == 0,
		Issues:   issues,
		Warnings: warnings,
		Stats: map[string]string{
			"total_lines": fmt.Sprintf("%d", len(lines)),
			"errors":      fmt.Sprintf("%d", len(issues)),
			"warnings":    fmt.Sprintf("%d", len(warnings)),
		},
	}
}

// applyRule 应用单个规则
func (s *SpecService) applyRule(lines []string, rule SpecRule) []SpecIssue {
	var issues []SpecIssue

	switch rule.Pattern {
	case "required_name":
		if !s.hasSection(lines, "Name:") {
			issues = append(issues, SpecIssue{
				Line:     1,
				Message:  "Missing required field: Name",
				Severity: rule.Severity,
				Rule:     rule.ID,
				RuleDesc: rule.Description,
			})
		}

	case "required_version":
		if !s.hasSection(lines, "Version:") {
			issues = append(issues, SpecIssue{
				Line:     1,
				Message:  "Missing required field: Version",
				Severity: rule.Severity,
				Rule:     rule.ID,
				RuleDesc: rule.Description,
			})
		}

	case "required_release":
		if !s.hasSection(lines, "Release:") {
			issues = append(issues, SpecIssue{
				Line:     1,
				Message:  "Missing required field: Release",
				Severity: rule.Severity,
				Rule:     rule.ID,
				RuleDesc: rule.Description,
			})
		}

	case "required_summary":
		if !s.hasSection(lines, "Summary:") {
			issues = append(issues, SpecIssue{
				Line:     1,
				Message:  "Missing required field: Summary",
				Severity: rule.Severity,
				Rule:     rule.ID,
				RuleDesc: rule.Description,
			})
		}

	case "required_license":
		if !s.hasSection(lines, "License:") {
			issues = append(issues, SpecIssue{
				Line:     1,
				Message:  "Missing required field: License",
				Severity: rule.Severity,
				Rule:     rule.ID,
				RuleDesc: rule.Description,
			})
		}

	case "changelog_format":
		for i, line := range lines {
			if strings.HasPrefix(line, "%changelog") {
				// 检查 changelog 格式
				if i+1 < len(lines) && !strings.HasPrefix(lines[i+1], "*") {
					issues = append(issues, SpecIssue{
						Line:     i + 2,
						Message:  "Changelog entry should start with '*'",
						Severity: rule.Severity,
						Rule:     rule.ID,
						RuleDesc: rule.Description,
					})
				}
			}
		}

	case "no_tabs":
		for i, line := range lines {
			if strings.Contains(line, "\t") {
				issues = append(issues, SpecIssue{
					Line:     i + 1,
					Message:  "Avoid using tabs, use spaces instead",
					Severity: rule.Severity,
					Rule:     rule.ID,
					RuleDesc: rule.Description,
				})
			}
		}

	default:
		// 使用正则表达式匹配
		if rule.Pattern != "" {
			re, err := regexp.Compile(rule.Pattern)
			if err == nil {
				for i, line := range lines {
					if re.MatchString(line) {
						issues = append(issues, SpecIssue{
							Line:     i + 1,
							Message:  fmt.Sprintf("Line matches rule: %s", rule.Name),
							Severity: rule.Severity,
							Rule:     rule.ID,
							RuleDesc: rule.Description,
						})
					}
				}
			}
		}
	}

	return issues
}

// hasSection 检查是否有指定的 section
func (s *SpecService) hasSection(lines []string, prefix string) bool {
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return true
		}
	}
	return false
}

// GetBuiltinRules 获取内置规则
func (s *SpecService) GetBuiltinRules() []SpecRule {
	return []SpecRule{
		// 必需字段
		{
			ID:          "required-name",
			Name:        "Required: Name",
			Description: "Spec file must have a Name field",
			Severity:    "error",
			Pattern:     "required_name",
			Enabled:     true,
			Category:    "required",
		},
		{
			ID:          "required-version",
			Name:        "Required: Version",
			Description: "Spec file must have a Version field",
			Severity:    "error",
			Pattern:     "required_version",
			Enabled:     true,
			Category:    "required",
		},
		{
			ID:          "required-release",
			Name:        "Required: Release",
			Description: "Spec file must have a Release field",
			Severity:    "error",
			Pattern:     "required_release",
			Enabled:     true,
			Category:    "required",
		},
		{
			ID:          "required-summary",
			Name:        "Required: Summary",
			Description: "Spec file must have a Summary field",
			Severity:    "error",
			Pattern:     "required_summary",
			Enabled:     true,
			Category:    "required",
		},
		{
			ID:          "required-license",
			Name:        "Required: License",
			Description: "Spec file must have a License field",
			Severity:    "error",
			Pattern:     "required_license",
			Enabled:     true,
			Category:    "required",
		},

		// 格式规范
		{
			ID:          "changelog-format",
			Name:        "Changelog Format",
			Description: "Changelog entries should follow RPM format",
			Severity:    "warning",
			Pattern:     "changelog_format",
			Enabled:     true,
			Category:    "style",
		},
		{
			ID:          "no-tabs",
			Name:        "No Tabs",
			Description: "Use spaces instead of tabs for consistency",
			Severity:    "info",
			Pattern:     "no_tabs",
			Enabled:     true,
			Category:    "style",
		},

		// 最佳实践
		{
			ID:          "buildroot-usage",
			Name:        "BuildRoot Usage",
			Description: "BuildRoot is deprecated in modern RPM",
			Severity:    "warning",
			Pattern:     "(?i)^BuildRoot:",
			Enabled:     false,
			Category:    "best-practice",
		},
		{
			ID:          "defattr-usage",
			Name:        "%defattr Usage",
			Description: "%defattr is usually not needed in modern RPM",
			Severity:    "info",
			Pattern:     "%defattr",
			Enabled:     false,
			Category:    "best-practice",
		},
	}
}

// GetSpecTemplate 获取 spec 文件模板
func (s *SpecService) GetSpecTemplate() string {
	return `Name:           
Version:        
Release:        1%{?dist}
Summary:        

License:        
URL:            
Source0:        

BuildRequires:  
Requires:       

%description


%prep
%setup -q

%build

%install
rm -rf $RPM_BUILD_ROOT

%clean
rm -rf $RPM_BUILD_ROOT

%files
%doc

%changelog
* $(date +"%a %b %d %Y") Your Name <your.email@example.com> - VERSION-1
- Initial package
`
}

// SpecFileInfo spec 文件信息
type SpecFileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size,omitempty"`
	ModTime string `json:"mod_time,omitempty"`
}

// SpecValidationResult 验证结果
type SpecValidationResult struct {
	Valid    bool              `json:"valid"`
	Issues   []SpecIssue       `json:"issues"`
	Warnings []SpecIssue       `json:"warnings"`
	Stats    map[string]string `json:"stats"`
}

// SpecIssue spec 问题
type SpecIssue struct {
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
	Rule      string `json:"rule"`
	RuleDesc  string `json:"rule_desc"`
	QuickFix  string `json:"quick_fix,omitempty"`
}

// SpecRule spec 规则
type SpecRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Pattern     string `json:"pattern"`
	Enabled     bool   `json:"enabled"`
	Category    string `json:"category"`
	AutoFix     bool   `json:"auto_fix"`
}

// ReadSpecLines 按行读取 spec 文件（用于编辑器）
func (s *SpecService) ReadSpecLines(content string) []string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}
