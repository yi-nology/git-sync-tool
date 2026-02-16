// biz/service/notification/template.go - 通知模板引擎

package notification

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

// TemplateData 模板变量数据
type TemplateData struct {
	TaskKey        string // 任务标识
	Status         string // success/failure/conflict
	StatusText     string // 成功/失败/冲突
	EventType      string // sync_success, sync_failure, etc.
	EventLabel     string // 同步成功, 同步失败, etc.
	SourceRemote   string // 源远程仓库名
	SourceBranch   string // 源分支名
	TargetRemote   string // 目标远程仓库名
	TargetBranch   string // 目标分支名
	RepoKey        string // 仓库标识
	ErrorMessage   string // 错误信息
	CommitRange    string // 提交范围
	CronExpression string // Cron表达式
	WebhookSource  string // Webhook来源
	BackupPath     string // 备份路径
	Timestamp      string // 格式化时间
	Duration       string // 执行耗时
}

// VariableInfo 模板变量说明
type VariableInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example"`
	Events      string `json:"events"` // 适用事件，逗号分隔，"all" 表示全部
}

// templateFuncMap 模板自定义函数
var templateFuncMap = template.FuncMap{
	"default": func(defaultVal, val string) string {
		if val == "" {
			return defaultVal
		}
		return val
	},
	"truncate": func(maxLen int, s string) string {
		if len(s) <= maxLen {
			return s
		}
		return s[:maxLen] + "..."
	},
}

// RenderTemplate 渲染模板
func RenderTemplate(tmplStr string, data *TemplateData) (string, error) {
	if tmplStr == "" {
		return "", nil
	}

	t, err := template.New("notification").Funcs(templateFuncMap).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("模板解析失败: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("模板渲染失败: %w", err)
	}

	return buf.String(), nil
}

// ValidateTemplate 验证模板语法是否合法
func ValidateTemplate(tmplStr string) error {
	if tmplStr == "" {
		return nil
	}
	_, err := template.New("validate").Funcs(templateFuncMap).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("模板语法错误: %w", err)
	}
	return nil
}

// RenderTitleAndContent 渲染标题和内容，优先使用自定义模板，否则使用默认模板
func RenderTitleAndContent(titleTmpl, contentTmpl string, data *TemplateData) (title, content string) {
	if data == nil {
		return "通知", ""
	}

	// 填充默认值
	fillDefaults(data)

	// 渲染标题
	if titleTmpl == "" {
		titleTmpl = GetDefaultTitleTemplate(data.EventType)
	}
	var err error
	title, err = RenderTemplate(titleTmpl, data)
	if err != nil {
		// fallback
		title = fmt.Sprintf("[%s] %s", data.EventLabel, data.TaskKey)
	}

	// 渲染内容
	if contentTmpl == "" {
		contentTmpl = GetDefaultContentTemplate(data.EventType)
	}
	content, err = RenderTemplate(contentTmpl, data)
	if err != nil {
		// fallback
		content = fmt.Sprintf("任务: %s\n状态: %s\n时间: %s", data.TaskKey, data.StatusText, data.Timestamp)
	}

	return title, content
}

// fillDefaults 填充 TemplateData 的默认值
func fillDefaults(data *TemplateData) {
	if data.Timestamp == "" {
		data.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	}

	if data.StatusText == "" {
		switch data.Status {
		case "success":
			data.StatusText = "成功"
		case "failure":
			data.StatusText = "失败"
		case "conflict":
			data.StatusText = "冲突"
		default:
			data.StatusText = data.Status
		}
	}

	if data.EventLabel == "" {
		data.EventLabel = eventLabels[data.EventType]
		if data.EventLabel == "" {
			data.EventLabel = data.EventType
		}
	}
}

// 事件类型 -> 中文标签
var eventLabels = map[string]string{
	po.TriggerSyncSuccess:     "同步成功",
	po.TriggerSyncFailure:     "同步失败",
	po.TriggerSyncConflict:    "同步冲突",
	po.TriggerWebhookReceived: "Webhook接收",
	po.TriggerWebhookError:    "Webhook错误",
	po.TriggerCronTriggered:   "定时任务触发",
	po.TriggerBackupSuccess:   "备份成功",
	po.TriggerBackupFailure:   "备份失败",
}

// ===================== 默认模板 =====================

// 默认标题模板
var defaultTitleTemplates = map[string]string{
	po.TriggerSyncSuccess:     `[成功] 同步任务 {{.TaskKey}}`,
	po.TriggerSyncFailure:     `[失败] 同步任务 {{.TaskKey}}`,
	po.TriggerSyncConflict:    `[冲突] 同步任务 {{.TaskKey}}`,
	po.TriggerWebhookReceived: `[Webhook] 收到请求: {{.TaskKey}}`,
	po.TriggerWebhookError:    `[Webhook错误] {{.TaskKey}}`,
	po.TriggerCronTriggered:   `[定时] 任务触发: {{.TaskKey}}`,
	po.TriggerBackupSuccess:   `[备份成功] {{.RepoKey}}`,
	po.TriggerBackupFailure:   `[备份失败] {{.RepoKey}}`,
}

// 默认内容模板
var defaultContentTemplates = map[string]string{
	po.TriggerSyncSuccess: strings.TrimSpace(`
任务: {{.TaskKey}}
状态: {{.StatusText}}
源: {{.SourceRemote}}/{{.SourceBranch}}
目标: {{.TargetRemote}}/{{.TargetBranch}}{{if .CommitRange}}
提交范围: {{.CommitRange}}{{end}}{{if .Duration}}
耗时: {{.Duration}}{{end}}
时间: {{.Timestamp}}
`),

	po.TriggerSyncFailure: strings.TrimSpace(`
任务: {{.TaskKey}}
状态: {{.StatusText}}
源: {{.SourceRemote}}/{{.SourceBranch}}
目标: {{.TargetRemote}}/{{.TargetBranch}}{{if .ErrorMessage}}
错误: {{.ErrorMessage}}{{end}}{{if .Duration}}
耗时: {{.Duration}}{{end}}
时间: {{.Timestamp}}
`),

	po.TriggerSyncConflict: strings.TrimSpace(`
任务: {{.TaskKey}}
状态: 同步冲突
源: {{.SourceRemote}}/{{.SourceBranch}}
目标: {{.TargetRemote}}/{{.TargetBranch}}
说明: 源分支和目标分支存在分叉，无法快进合并{{if .ErrorMessage}}
详情: {{.ErrorMessage}}{{end}}
时间: {{.Timestamp}}
`),

	po.TriggerWebhookReceived: strings.TrimSpace(`
任务: {{.TaskKey}}
状态: Webhook 请求已接收{{if .WebhookSource}}
来源: {{.WebhookSource}}{{end}}
时间: {{.Timestamp}}
`),

	po.TriggerWebhookError: strings.TrimSpace(`
任务: {{.TaskKey}}
状态: Webhook 处理失败{{if .WebhookSource}}
来源: {{.WebhookSource}}{{end}}{{if .ErrorMessage}}
错误: {{.ErrorMessage}}{{end}}
时间: {{.Timestamp}}
`),

	po.TriggerCronTriggered: strings.TrimSpace(`
任务: {{.TaskKey}}
状态: 定时任务已触发{{if .CronExpression}}
Cron表达式: {{.CronExpression}}{{end}}
时间: {{.Timestamp}}
`),

	po.TriggerBackupSuccess: strings.TrimSpace(`
仓库: {{.RepoKey}}
状态: 备份成功{{if .BackupPath}}
备份路径: {{.BackupPath}}{{end}}{{if .Duration}}
耗时: {{.Duration}}{{end}}
时间: {{.Timestamp}}
`),

	po.TriggerBackupFailure: strings.TrimSpace(`
仓库: {{.RepoKey}}
状态: 备份失败{{if .ErrorMessage}}
错误: {{.ErrorMessage}}{{end}}{{if .BackupPath}}
备份路径: {{.BackupPath}}{{end}}
时间: {{.Timestamp}}
`),
}

// GetDefaultTitleTemplate 获取事件类型的默认标题模板
func GetDefaultTitleTemplate(eventType string) string {
	if t, ok := defaultTitleTemplates[eventType]; ok {
		return t
	}
	return `[通知] {{.TaskKey}}`
}

// GetDefaultContentTemplate 获取事件类型的默认内容模板
func GetDefaultContentTemplate(eventType string) string {
	if t, ok := defaultContentTemplates[eventType]; ok {
		return t
	}
	return `任务: {{.TaskKey}}\n状态: {{.StatusText}}\n时间: {{.Timestamp}}`
}

// GetAvailableVariables 获取所有可用模板变量列表（供前端展示）
func GetAvailableVariables() []VariableInfo {
	return []VariableInfo{
		{Name: "TaskKey", Description: "任务标识", Example: "my-sync-task", Events: "all"},
		{Name: "Status", Description: "状态码", Example: "success", Events: "all"},
		{Name: "StatusText", Description: "状态文字", Example: "成功", Events: "all"},
		{Name: "EventType", Description: "事件类型", Example: "sync_success", Events: "all"},
		{Name: "EventLabel", Description: "事件名称", Example: "同步成功", Events: "all"},
		{Name: "Timestamp", Description: "时间", Example: "2026-02-16 10:30:00", Events: "all"},
		{Name: "SourceRemote", Description: "源远程仓库", Example: "origin", Events: "sync_*"},
		{Name: "SourceBranch", Description: "源分支", Example: "main", Events: "sync_*"},
		{Name: "TargetRemote", Description: "目标远程仓库", Example: "backup", Events: "sync_*"},
		{Name: "TargetBranch", Description: "目标分支", Example: "main", Events: "sync_*"},
		{Name: "RepoKey", Description: "仓库标识", Example: "my-repo", Events: "all"},
		{Name: "ErrorMessage", Description: "错误信息", Example: "push failed: ...", Events: "*_failure,*_error,sync_conflict"},
		{Name: "CommitRange", Description: "提交范围", Example: "abc123..def456", Events: "sync_success"},
		{Name: "Duration", Description: "执行耗时", Example: "3.2s", Events: "sync_*,backup_*"},
		{Name: "CronExpression", Description: "Cron表达式", Example: "0 2 * * *", Events: "cron_triggered"},
		{Name: "WebhookSource", Description: "Webhook来源", Example: "github", Events: "webhook_*"},
		{Name: "BackupPath", Description: "备份路径", Example: "/backups/repo.tar.gz", Events: "backup_*"},
	}
}
