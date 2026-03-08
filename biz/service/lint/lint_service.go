package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/configs"
)

type LintService struct{}

func NewLintService() *LintService {
	return &LintService{}
}

type LintRequest struct {
	Content string   `json:"content"`
	Rules   []string `json:"rules,omitempty"`
}

type LintResult struct {
	File   string      `json:"file"`
	Issues []LintIssue `json:"issues"`
	Stats  LintStats   `json:"stats"`
}

type LintIssue struct {
	RuleID    string `json:"ruleId"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Line      int    `json:"line"`
	Column    int    `json:"column,omitempty"`
	EndLine   int    `json:"endLine,omitempty"`
	EndColumn int    `json:"endColumn,omitempty"`
}

type LintStats struct {
	ErrorCount   int `json:"errorCount"`
	WarningCount int `json:"warningCount"`
	InfoCount    int `json:"infoCount"`
}

func (s *LintService) Lint(content string, ruleIDs []string) (*LintResult, error) {
	var rules []po.LintRule
	var err error

	if len(ruleIDs) > 0 {
		rules, err = db.NewLintRuleDAO().FindByIDs(ruleIDs)
	} else {
		rules, err = db.NewLintRuleDAO().FindEnabled()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load lint rules: %v", err)
	}

	result := &LintResult{
		File:   "",
		Issues: []LintIssue{},
		Stats:  LintStats{},
	}

	lines := strings.Split(content, "\n")

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		issues := s.applyRule(lines, rule)
		result.Issues = append(result.Issues, issues...)

		for _, issue := range issues {
			switch issue.Severity {
			case "error":
				result.Stats.ErrorCount++
			case "warning":
				result.Stats.WarningCount++
			case "info":
				result.Stats.InfoCount++
			}
		}
	}

	if configs.GlobalConfig.Lint.EnableRpmlint {
		if rpmlintIssues := s.runRpmlint(content, ""); len(rpmlintIssues) > 0 {
			result.Issues = append(result.Issues, rpmlintIssues...)
			for _, issue := range rpmlintIssues {
				switch issue.Severity {
				case "error":
					result.Stats.ErrorCount++
				case "warning":
					result.Stats.WarningCount++
				case "info":
					result.Stats.InfoCount++
				}
			}
		}
	}

	return result, nil
}

func (s *LintService) applyRule(lines []string, rule po.LintRule) []LintIssue {
	var issues []LintIssue

	switch rule.Pattern {
	case "required_name":
		if !s.hasField(lines, "Name:") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing required field: Name",
				Line:     1,
			})
		}

	case "required_version":
		if !s.hasField(lines, "Version:") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing required field: Version",
				Line:     1,
			})
		}

	case "required_release":
		if !s.hasField(lines, "Release:") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing required field: Release",
				Line:     1,
			})
		}

	case "required_summary":
		if !s.hasField(lines, "Summary:") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing required field: Summary",
				Line:     1,
			})
		}

	case "required_license":
		if !s.hasField(lines, "License:") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing required field: License",
				Line:     1,
			})
		}

	case "required_url":
		if !s.hasField(lines, "URL:") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing recommended field: URL",
				Line:     1,
			})
		}

	case "required_description":
		if !s.hasSection(lines, "%description") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing required section: %description",
				Line:     1,
			})
		}

	case "required_prep":
		if !s.hasSection(lines, "%prep") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing recommended section: %prep",
				Line:     1,
			})
		}

	case "required_build":
		if !s.hasSection(lines, "%build") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing recommended section: %build",
				Line:     1,
			})
		}

	case "required_install":
		if !s.hasSection(lines, "%install") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing recommended section: %install",
				Line:     1,
			})
		}

	case "required_files":
		if !s.hasSection(lines, "%files") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing required section: %files",
				Line:     1,
			})
		}

	case "empty_sections":
		sections := []string{"%description", "%prep", "%build", "%install", "%files"}
		for _, section := range sections {
			sectionLine := s.findSectionLine(lines, section)
			if sectionLine > 0 {
				// 检查段落是否为空（只有空行或下一个段落）
				isEmpty := true
				for i := sectionLine; i < len(lines); i++ {
					line := strings.TrimSpace(lines[i])
					if line == "" {
						continue
					}
					if strings.HasPrefix(line, "%") && i > sectionLine {
						break
					}
					if !strings.HasPrefix(line, "%") {
						isEmpty = false
						break
					}
				}
				if isEmpty {
					issues = append(issues, LintIssue{
						RuleID:   rule.ID,
						Severity: rule.Severity,
						Message:  fmt.Sprintf("Section %s is empty", section),
						Line:     sectionLine + 1,
					})
				}
			}
		}

	case "buildroot_usage":
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "BuildRoot:") {
				if !strings.Contains(line, "%{_tmppath}") {
					issues = append(issues, LintIssue{
						RuleID:   rule.ID,
						Severity: rule.Severity,
						Message:  "BuildRoot should use %{_tmppath} macro",
						Line:     i + 1,
					})
				}
			}
		}

	case "macro_consistency":
		// 检查大括号宏和不带大括号宏的混用
		reBraces := regexp.MustCompile(`%\{[a-zA-Z_][a-zA-Z0-9_]*\}`)
		reNoBraces := regexp.MustCompile(`%[a-zA-Z_][a-zA-Z0-9_]*[^{a-zA-Z0-9_]`)
		for i, line := range lines {
			braces := reBraces.FindAllString(line, -1)
			noBraces := reNoBraces.FindAllString(line, -1)
			if len(braces) > 0 && len(noBraces) > 0 {
				issues = append(issues, LintIssue{
					RuleID:   rule.ID,
					Severity: rule.Severity,
					Message:  "Inconsistent macro usage: use either %{macro} or %macro consistently",
					Line:     i + 1,
				})
				break
			}
		}

	case "changelog_format":
		for i, line := range lines {
			if strings.HasPrefix(line, "%changelog") {
				if i+1 < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i+1]), "*") {
					issues = append(issues, LintIssue{
						RuleID:   rule.ID,
						Severity: rule.Severity,
						Message:  "Changelog entry should start with '*'",
						Line:     i + 2,
					})
				}
			}
		}

	case "no_tabs":
		for i, line := range lines {
			if strings.Contains(line, "\t") {
				issues = append(issues, LintIssue{
					RuleID:   rule.ID,
					Severity: rule.Severity,
					Message:  "Avoid using tabs, use spaces instead",
					Line:     i + 1,
				})
			}
		}

	default:
		if rule.Pattern != "" {
			re, err := regexp.Compile(rule.Pattern)
			if err == nil {
				for i, line := range lines {
					loc := re.FindStringIndex(line)
					if loc != nil {
						issues = append(issues, LintIssue{
							RuleID:    rule.ID,
							Severity:  rule.Severity,
							Message:   fmt.Sprintf("Line matches rule: %s", rule.Name),
							Line:      i + 1,
							Column:    loc[0] + 1,
							EndLine:   i + 1,
							EndColumn: loc[1],
						})
					}
				}
			}
		}
	}

	return issues
}

func (s *LintService) hasField(lines []string, prefix string) bool {
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return true
		}
	}
	return false
}

func (s *LintService) hasSection(lines []string, section string) bool {
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), section) {
			return true
		}
	}
	return false
}

func (s *LintService) findSectionLine(lines []string, section string) int {
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), section) {
			return i + 1
		}
	}
	return 0
}

func (s *LintService) runRpmlint(content string, specFilePath string) []LintIssue {
	_, err := exec.LookPath("rpmlint")
	if err != nil {
		return []LintIssue{}
	}

	tmpFile, err := os.CreateTemp("", "spec-*.spec")
	if err != nil {
		return []LintIssue{}
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		tmpFile.Close()
		return []LintIssue{}
	}
	tmpFile.Close()

	if specFilePath == "" {
		specFilePath = tmpFile.Name()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "rpmlint", "-f", "json", specFilePath)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if len(exitErr.Stderr) > 0 {
			}
		}
	}

	var rpmlintResult []struct {
		File     string `json:"file"`
		Line     int    `json:"line"`
		Message  string `json:"message"`
		Severity string `json:"severity"`
	}

	if err := json.Unmarshal(output, &rpmlintResult); err != nil {
		return []LintIssue{}
	}

	var issues []LintIssue
	for _, item := range rpmlintResult {
		severity := "info"
		if strings.Contains(strings.ToLower(item.Severity), "error") {
			severity = "error"
		} else if strings.Contains(strings.ToLower(item.Severity), "warning") {
			severity = "warning"
		}

		issues = append(issues, LintIssue{
			RuleID:   "rpmlint",
			Severity: severity,
			Message:  item.Message,
			Line:     item.Line,
		})
	}

	return issues
}
