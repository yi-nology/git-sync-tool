# 统一 Git 平台能力层 — 设计文档

> 基于现有 git-manage-service 架构，新增 Provider 抽象层，屏蔽 GitLab/GitHub/Gitea 差异。

---

## 1. 现状分析

| 能力 | 现状 | 缺口 |
|------|------|------|
| 仓库管理 | 完整 CRUD + clone + scan | 无平台感知，纯 git 协议操作 |
| 分支管理 | 17 个 API，含 cherry-pick/rebase | 无平台分支保护规则 |
| 凭证管理 | SSH/HTTP Basic/HTTP Token 三类 | 无平台 Token（OAuth2/PAT）类型 |
| 同步 | 单/全分支 + cron + webhook 触发 | 无双向同步，无镜像模式 |
| Webhook | 仅出站触发同步任务 | **无入站接收**，无事件解析 |
| PR/MR | **完全缺失** | 需要新建整个模块 |

**核心结论**：系统目前在 git 协议层非常完善，但缺少**平台 API 集成层**。设计重点是在不破坏现有分层的前提下，插入 Provider 抽象。

---

## 2. 总体架构

```
                         ┌─────────────────────────┐
                         │     HTTP API (Hz)        │
                         │  /api/v1/cr/*            │
                         │  /api/v1/webhook/*       │
                         └────────┬────────────────┘
                                  │
                    ┌─────────────┴─────────────┐
                    │      Handler Layer         │  ← 参数绑定、校验
                    └─────────────┬─────────────┘
                                  │
              ┌───────────────────┴───────────────────┐
              │           Service Layer                │
              │  ┌──────────────────────────────────┐  │
              │  │      Provider Manager             │  │  ← 新增核心
              │  │  ┌──────┬──────┬──────┬──────┐   │  │
              │  │  │ GitLab│GitHub│Gitea │ Mock │   │  │
              │  │  └──┬───┘──┬───┘──┬───┘──┬───┘   │  │
              │  │     └───────┴───────┴──────┘      │  │
              │  │     统一接口: Provider              │  │
              │  └──────────────────────────────────┘  │
              │                                         │
              │  ┌──────────────┐  ┌────────────────┐  │
              │  │ CR Service   │  │ Webhook Service │  │  ← 新增
              │  └──────────────┘  └────────────────┘  │
              │                                         │
              │  ┌────────────────────────────────────┐ │
              │  │ 现有: Git Service / Sync Service    │ │  ← 不变
              │  └────────────────────────────────────┘ │
              └───────────────────┬─────────────────────┘
                                  │
              ┌───────────────────┴─────────────────────┐
              │              DAO Layer (GORM)            │
              │  provider_config / cr / webhook_event    │  ← 新增表
              └───────────────────┬─────────────────────┘
                                  │
                              SQLite / MySQL
```

**设计原则**：
- 现有 `biz/service/git/`（git 协议层）**完全不动**
- 新增 `biz/service/provider/`（平台 API 层），与 git 层平行
- Handler/Router/DAO 遵循现有分层模式
- Provider 通过接口抽象，每个平台一个实现

---

## 3. Provider 抽象层

### 3.1 核心接口

```go
// biz/service/provider/provider.go

type Platform string

const (
    PlatformGitLab Platform = "gitlab"
    PlatformGitHub Platform = "github"
    PlatformGitea  Platform = "gitea"
)

type Provider interface {
    // 平台信息
    Platform() Platform
    
    // 仓库
    ListRepos(ctx context.Context, opts ListRepoOptions) ([]*PlatformRepo, error)
    GetRepo(ctx context.Context, owner, repo string) (*PlatformRepo, error)
    
    // 分支保护
    ListProtectedBranches(ctx context.Context, owner, repo string) ([]*BranchProtection, error)
    
    // Merge Request / Pull Request
    CreateCR(ctx context.Context, opts CreateCROptions) (*ChangeRequest, error)
    GetCR(ctx context.Context, owner, repo string, number int) (*ChangeRequest, error)
    ListCRs(ctx context.Context, opts ListCROptions) ([]*ChangeRequest, error)
    MergeCR(ctx context.Context, owner, repo string, number int, opts MergeCROptions) (*ChangeRequest, error)
    CloseCR(ctx context.Context, owner, repo string, number int) (*ChangeRequest, error)
    
    // Webhook 注册
    CreateWebhook(ctx context.Context, opts CreateWebhookOptions) (*PlatformWebhook, error)
    DeleteWebhook(ctx context.Context, owner, repo string, webhookID int64) error
    ListWebhooks(ctx context.Context, owner, repo string) ([]*PlatformWebhook, error)
    
    // Webhook 事件解析
    ParseWebhookEvent(r *http.Request) (*NormalizedEvent, error)
    ValidateWebhookSignature(r *http.Request, secret string) error
}
```

### 3.2 统一数据模型（接口层 DTO）

```go
type PlatformRepo struct {
    ID          int64  `json:"id"`
    FullName    string `json:"full_name"`     // "owner/repo"
    Name        string `json:"name"`
    Owner       string `json:"owner"`
    Description string `json:"description"`
    CloneURL    string `json:"clone_url"`
    SSHURL      string `json:"ssh_url"`
    DefaultBranch string `json:"default_branch"`
    Private     bool   `json:"private"`
    Platform    Platform `json:"platform"`
}

type ChangeRequest struct {
    ID          int64        `json:"id"`
    Number      int          `json:"number"`       // 平台 MR/PR 编号
    Title       string       `json:"title"`
    Description string       `json:"description"`
    State       CRState      `json:"state"`        // opened, merged, closed
    SourceBranch string      `json:"source_branch"`
    TargetBranch string      `json:"target_branch"`
    Author      *CRUser      `json:"author"`
    Reviewers   []*CRUser    `json:"reviewers"`
    Labels      []string     `json:"labels"`
    MergeStatus string       `json:"merge_status"` // mergeable, conflicting, checking
    WebURL      string       `json:"web_url"`      // 平台页面链接
    CreatedAt   time.Time    `json:"created_at"`
    UpdatedAt   time.Time    `json:"updated_at"`
}

type CRState string
const (
    CRStateOpened CRState = "opened"
    CRStateMerged CRState = "merged"
    CRStateClosed CRState = "closed"
)

type CRUser struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    Name     string `json:"name"`
    AvatarURL string `json:"avatar_url"`
}

type ListRepoOptions struct {
    Owner  string `json:"owner"`
    Page   int    `json:"page"`
    PerPage int   `json:"per_page"`
}

type CreateCROptions struct {
    Owner        string `json:"owner"`
    Repo         string `json:"repo"`
    Title        string `json:"title"`
    Description  string `json:"description"`
    SourceBranch string `json:"source_branch"`
    TargetBranch string `json:"target_branch"`
    Labels       []string `json:"labels"`
    RemoveSourceBranch bool `json:"remove_source_branch"`
}

type ListCROptions struct {
    Owner        string `json:"owner"`
    Repo         string `json:"repo"`
    State        CRState `json:"state"`
    SourceBranch string `json:"source_branch"`
    TargetBranch string `json:"target_branch"`
    Page         int    `json:"page"`
    PerPage      int    `json:"per_page"`
}

type MergeCROptions struct {
    MergeCommitMessage string `json:"merge_commit_message"`
    Squash             bool   `json:"squash"`
    RemoveSourceBranch bool   `json:"remove_source_branch"`
}

type BranchProtection struct {
    Pattern         string   `json:"pattern"`
    AllowPush       []string `json:"allow_push"`
    AllowMerge      []string `json:"allow_merge"`
    RequireReview   bool     `json:"require_review"`
    MinReviewers    int      `json:"min_reviewers"`
}

type CreateWebhookOptions struct {
    Owner     string   `json:"owner"`
    Repo      string   `json:"repo"`
    URL       string   `json:"url"`
    Secret    string   `json:"secret"`
    Events    []string `json:"events"` // push, cr, branch
}

type PlatformWebhook struct {
    ID     int64    `json:"id"`
    URL    string   `json:"url"`
    Events []string `json:"events"`
}
```

### 3.3 目录结构

```
biz/service/provider/
├── provider.go           # Provider 接口定义 + 统一 DTO
├── manager.go            # ProviderManager: 注册/获取 provider 实例
├── detect.go             # 从 URL 自动检测平台类型
├── gitlab/
│   ├── provider.go       # GitLab Provider 实现
│   ├── mr.go             # MR 操作
│   ├── webhook.go        # Webhook 注册 + 事件解析
│   └── types.go          # GitLab API 原始类型 → 统一 DTO 转换
├── github/
│   ├── provider.go
│   ├── pr.go
│   ├── webhook.go
│   └── types.go
└── gitea/
    ├── provider.go
    ├── pr.go
    ├── webhook.go
    └── types.go
```

### 3.4 平台自动检测

```go
// biz/service/provider/detect.go

func DetectPlatform(remoteURL string) (Platform, string, string, error) {
    // 解析 remote URL，提取平台标识 + owner + repo
    // git@github.com:owner/repo.git     → github, owner, repo
    // https://gitlab.com/owner/repo.git → gitlab, owner, repo
    // git@gitea.example.com:org/repo    → gitea, org, repo
    
    patterns := map[string]Platform{
        "github.com":         PlatformGitHub,
        "gitlab.com":         PlatformGitLab,
        "gitlab.example.com": PlatformGitLab, // 自托管，从配置读
    }
    // 同时查 provider_configs 表匹配自定义域名
}
```

### 3.5 Provider Manager

```go
// biz/service/provider/manager.go

type ProviderManager struct {
    providers map[Platform]Provider
    configs   map[int64]*po.ProviderConfig  // 缓存，按 config ID
}

func NewProviderManager(dao *db.ProviderConfigDAO) *ProviderManager

// GetProvider 根据 repo 关联的 provider_config_id 获取已认证的 Provider
func (m *ProviderManager) GetProvider(configID int64) (Provider, error)

// DetectAndCreate 从 remote URL 自动检测平台并创建 Provider
func (m *ProviderManager) DetectAndCreate(remoteURL string, credential *po.Credential) (Provider, error)
```

---

## 4. 数据库模型

### 4.1 provider_config（平台连接配置）— 新表

```go
// biz/model/po/provider_config.go

type ProviderConfig struct {
    gorm.Model
    Name        string `gorm:"uniqueIndex;size:100" json:"name"`         // "公司 GitLab"
    Platform    string `gorm:"size:20;index" json:"platform"`            // gitlab, github, gitea
    BaseURL     string `gorm:"size:500" json:"base_url"`                 // https://gitlab.com (自托管可改)
    CredentialID uint  `gorm:"index" json:"credential_id"`              // 关联凭证 (Token 类型)
    
    // Webhook 接收配置
    WebhookSecret   string `gorm:"size:200" json:"webhook_secret"`        // 验证签名
    WebhookEndpoint string `gorm:"size:500" json:"webhook_endpoint"`      // 回调地址
    
    // 平台特有设置 (JSON)
    SettingsJSON string `gorm:"type:text" json:"-"`                      // 扩展字段
    Settings     map[string]interface{} `gorm:"-" json:"settings"`
}

func (ProviderConfig) TableName() string { return "provider_configs" }
```

### 4.2 扩展现有 Repo 模型

```go
// 在 po.Repo 中新增字段
type Repo struct {
    // ... 现有字段 ...
    
    // 新增: 平台关联
    ProviderConfigID uint   `gorm:"index" json:"provider_config_id"` // 关联平台配置
    PlatformRepoID   string `gorm:"size:100" json:"platform_repo_id"` // 平台侧仓库 ID
    PlatformOwner    string `gorm:"size:200" json:"platform_owner"`   // 平台 owner/org
    PlatformRepo     string `gorm:"size:200" json:"platform_repo"`    // 平台仓库名
}
```

### 4.3 扩展 Credential 模型

```go
// 在 po.Credential 的 Type 字段新增值
// 现有: ssh_key, http_basic, http_token
// 新增: platform_token

type Credential struct {
    // ... 现有字段 ...
    
    // 新增
    Platform      string `gorm:"size:20" json:"platform"`       // 所属平台 (当 Type=platform_token)
    PlatformScope string `gorm:"size:200" json:"platform_scope"` // Token 权限 scope
}
```

### 4.4 change_request（CR/MR 统一记录）— 新表

```go
// biz/model/po/change_request.go

type ChangeRequest struct {
    gorm.Model
    RepoID          uint   `gorm:"index" json:"repo_id"`                    // 关联本地 repo
    ProviderConfigID uint  `gorm:"index" json:"provider_config_id"`         // 关联平台配置
    PlatformCRID    int64  `gorm:"index" json:"platform_cr_id"`             // 平台侧 MR/PR ID
    CRNumber        int    `gorm:"index" json:"cr_number"`                  // 平台侧编号
    Title           string `gorm:"size:500" json:"title"`
    Description     string `gorm:"type:text" json:"description"`
    State           string `gorm:"size:20;index" json:"state"`              // opened, merged, closed
    SourceBranch    string `gorm:"size:200;index" json:"source_branch"`
    TargetBranch    string `gorm:"size:200;index" json:"target_branch"`
    AuthorName      string `gorm:"size:200" json:"author_name"`
    AuthorUsername  string `gorm:"size:200" json:"author_username"`
    WebURL          string `gorm:"size:500" json:"web_url"`
    MergeStatus     string `gorm:"size:30" json:"merge_status"`             // mergeable, conflicting
    LabelsJSON      string `gorm:"type:text" json:"-"`
    Labels          []string `gorm:"-" json:"labels"`
    MergedAt        *time.Time `json:"merged_at"`
    ClosedAt        *time.Time `json:"closed_at"`
}

func (ChangeRequest) TableName() string { return "change_requests" }
```

### 4.5 webhook_event（标准化事件记录）— 新表

```go
// biz/model/po/webhook_event.go

type WebhookEvent struct {
    gorm.Model
    EventID         string `gorm:"uniqueIndex;size:100" json:"event_id"`    // 去重 ID
    ProviderConfigID uint  `gorm:"index" json:"provider_config_id"`
    EventType       string `gorm:"size:50;index" json:"event_type"`          // cr.opened, cr.merged ...
    Source          string `gorm:"size:20;index" json:"source"`              // gitlab, github, gitea
    
    // 关联资源
    RepoID          uint   `gorm:"index" json:"repo_id"`
    CRID            uint   `gorm:"index" json:"cr_id"`                       // 关联 change_requests 表
    PlatformCRNumber int   `json:"platform_cr_number"`
    
    // 事件详情
    ActorName       string `gorm:"size:200" json:"actor_name"`
    ActorUsername   string `gorm:"size:200" json:"actor_username"`
    PayloadJSON     string `gorm:"type:text" json:"-"`                       // 原始 payload (加密)
    Payload         map[string]interface{} `gorm:"-" json:"payload"`
    
    // 处理状态
    Status          string `gorm:"size:20;index" json:"status"`              // received, processed, failed
    ProcessedAt     *time.Time `json:"processed_at"`
    ErrorMessage    string `gorm:"size:500" json:"error_message"`
}

func (WebhookEvent) TableName() string { return "webhook_events" }
```

### 4.6 webhook_rule（Webhook 路由规则）— 新表

```go
// biz/model/po/webhook_rule.go

type WebhookRule struct {
    gorm.Model
    Name            string `gorm:"size:100" json:"name"`
    ProviderConfigID uint  `gorm:"index" json:"provider_config_id"`          // 0 = 全局
    EventTypePattern string `gorm:"size:100" json:"event_type_pattern"`      // "cr.*", "branch.*"
    RepoPattern     string `gorm:"size:200" json:"repo_pattern"`             // "owner/repo", "*"
    Action          string `gorm:"size:50" json:"action"`                    // sync, notify, script
    ActionConfigJSON string `gorm:"type:text" json:"-"`                     // 动作参数
    ActionConfig    map[string]interface{} `gorm:"-" json:"action_config"`
    Enabled         bool   `gorm:"default:true" json:"enabled"`
}

func (WebhookRule) TableName() string { return "webhook_rules" }
```

### 4.7 ER 关系

```
provider_config ──1:N── repo
       │                  │
       │                  ├──1:N── change_request
       │                  │
       │                  └──1:N── webhook_event
       │
       └──1:N── webhook_event
                      │
                      └──N:1── change_request

webhook_rule ──N:1── provider_config

credential ──1:1── provider_config  (Token 凭证)
credential ──1:N── repo             (现有 SSH/HTTP 凭证)
```

---

## 5. Webhook 标准化

### 5.1 统一事件格式

入站 webhook 经 Provider 解析后，统一输出以下标准事件：

| 标准事件类型 | 触发条件 | 含义 |
|---|---|---|
| `cr.opened` | MR/PR 创建 | 变更请求被创建 |
| `cr.updated` | MR/PR 更新 | 变更请求被更新 |
| `cr.merged` | MR/PR 合并 | 变更请求被合并 |
| `cr.closed` | MR/PR 关闭（未合并） | 变更请求被关闭 |
| `cr.reviewed` | MR/PR 收到 Review | 审批状态变化 |
| `branch.created` | 分支创建 | 新分支出现 |
| `branch.deleted` | 分支删除 | 分支被删除 |
| `tag.created` | 标签创建 | 新标签 |
| `push` | 代码推送 | 有新 commit |

```go
type NormalizedEvent struct {
    ID          string                 `json:"id"`           // 唯一事件 ID
    Type        string                 `json:"type"`         // cr.opened, branch.created ...
    Source      Platform               `json:"source"`       // gitlab, github, gitea
    Timestamp   time.Time              `json:"timestamp"`
    Actor       *CRUser                `json:"actor"`
    Repo        *EventRepo             `json:"repo"`
    CR          *ChangeRequest         `json:"cr,omitempty"` // cr.* 事件时填充
    Branch      string                 `json:"branch,omitempty"`
    Tag         string                 `json:"tag,omitempty"`
    RawPayload  json.RawMessage        `json:"raw_payload"`  // 平台原始数据
}

type EventRepo struct {
    FullName string `json:"full_name"`
    Owner    string `json:"owner"`
    Name     string `json:"name"`
}
```

### 5.2 入站 Webhook 流程

```
GitLab/GitHub/Gitea
        │
        │  HTTP POST (平台原始 webhook payload)
        ▼
┌──────────────────────────────┐
│ POST /api/webhooks/receive   │  ← 统一入口
│         │                    │
│  1. 验证签名 (Provider)      │
│  2. 检测来源平台             │  ← Header: X-GitLab-Event / X-GitHub-Event
│  3. 路由到对应 Provider      │
│  4. 解析为 NormalizedEvent   │
│  5. 写入 webhook_events 表   │
│  6. 匹配 webhook_rule        │
│  7. 执行动作 (sync/notify)   │
└──────────────────────────────┘
```

### 5.3 Webhook 路由规则

每个 provider_config 可以配置多条规则：

```json
{
  "name": "MR merged → sync",
  "event_type_pattern": "cr.merged",
  "repo_pattern": "*",
  "action": "sync",
  "action_config": {
    "sync_task_key": "sync-to-mirror"
  }
}
```

```json
{
  "name": "MR opened → notify",
  "event_type_pattern": "cr.opened",
  "repo_pattern": "yi-nology/*",
  "action": "notify",
  "action_config": {
    "channel": "feishu",
    "template": "cr_opened_notification"
  }
}
```

---

## 6. PR/MR 管理模块

### 6.1 Proto 定义

```protobuf
// idl/biz/cr.proto

syntax = "proto3";
package cr;
option go_package = "github.com/yi-nology/git-manage-service/biz/model/cr";

import "api.proto";
import "common.proto";

message ChangeRequest {
    int64 id = 1;
    int32 cr_number = 2;
    string title = 3;
    string description = 4;
    string state = 5;         // opened, merged, closed
    string source_branch = 6;
    string target_branch = 7;
    string author_name = 8;
    string author_username = 9;
    string web_url = 10;
    string merge_status = 11;
    repeated string labels = 12;
    string created_at = 13;
    string updated_at = 14;
}

// ---- 创建 CR ----
message CreateCRRequest {
    string repo_key = 1 [(api.body) = "repo_key"];
    string title = 2 [(api.body) = "title"];
    string description = 3 [(api.body) = "description"];
    string source_branch = 4 [(api.body) = "source_branch"];
    string target_branch = 5 [(api.body) = "target_branch"];
    repeated string labels = 6 [(api.body) = "labels"];
    bool remove_source_branch = 7 [(api.body) = "remove_source_branch"];
}

message CreateCRResponse {
    common.BaseResponse base = 1;
    ChangeRequest cr = 2;
}

// ---- 查询 CR ----
message GetCRRequest {
    string repo_key = 1 [(api.query) = "repo_key"];
    int32 cr_number = 2 [(api.query) = "cr_number"];
}

message ListCRsRequest {
    string repo_key = 1 [(api.query) = "repo_key"];
    string state = 2 [(api.query) = "state"];
    string source_branch = 3 [(api.query) = "source_branch"];
    string target_branch = 4 [(api.query) = "target_branch"];
    int32 page = 5 [(api.query) = "page"];
    int32 page_size = 6 [(api.query) = "page_size"];
}

message ListCRsResponse {
    common.BaseResponse base = 1;
    repeated ChangeRequest items = 2;
    int32 total = 3;
}

// ---- 合并 CR ----
message MergeCRRequest {
    string repo_key = 1 [(api.body) = "repo_key"];
    int32 cr_number = 2 [(api.body) = "cr_number"];
    string merge_commit_message = 3 [(api.body) = "merge_commit_message"];
    bool squash = 4 [(api.body) = "squash"];
    bool remove_source_branch = 5 [(api.body) = "remove_source_branch"];
}

message MergeCRResponse {
    common.BaseResponse base = 1;
    ChangeRequest cr = 2;
}

// ---- 关闭 CR ----
message CloseCRRequest {
    string repo_key = 1 [(api.body) = "repo_key"];
    int32 cr_number = 2 [(api.body) = "cr_number"];
}

message CloseCRResponse {
    common.BaseResponse base = 1;
    ChangeRequest cr = 2;
}

// ---- 同步平台 CR 到本地 ----
message SyncCRsRequest {
    string repo_key = 1 [(api.query) = "repo_key"];
    string state = 2 [(api.query) = "state"];
}

message SyncCRsResponse {
    common.BaseResponse base = 1;
    int32 synced_count = 2;
}

service CRService {
    rpc Create(CreateCRRequest) returns (CreateCRResponse) {
        option (api.post) = "/api/v1/cr/create";
    }
    rpc Get(GetCRRequest) returns (CreateCRResponse) {
        option (api.get) = "/api/v1/cr/detail";
    }
    rpc List(ListCRsRequest) returns (ListCRsResponse) {
        option (api.get) = "/api/v1/cr/list";
    }
    rpc Merge(MergeCRRequest) returns (MergeCRResponse) {
        option (api.post) = "/api/v1/cr/merge";
    }
    rpc Close(CloseCRRequest) returns (CloseCRResponse) {
        option (api.post) = "/api/v1/cr/close";
    }
    rpc Sync(SyncCRsRequest) returns (SyncCRsResponse) {
        option (api.post) = "/api/v1/cr/sync";
    }
}
```

### 6.2 Webhook Event Proto

```protobuf
// idl/biz/webhook_event.proto

syntax = "proto3";
package webhook_event;
option go_package = "github.com/yi-nology/git-manage-service/biz/model/webhook_event";

import "api.proto";
import "common.proto";

message WebhookEvent {
    int64 id = 1;
    string event_id = 2;
    string event_type = 3;
    string source = 4;
    string actor_name = 5;
    string repo_full_name = 6;
    int32 platform_cr_number = 7;
    string status = 8;
    string created_at = 9;
    string processed_at = 10;
}

message ListWebhookEventsRequest {
    string event_type = 1 [(api.query) = "event_type"];
    string source = 2 [(api.query) = "source"];
    string status = 3 [(api.query) = "status"];
    int32 page = 4 [(api.query) = "page"];
    int32 page_size = 5 [(api.query) = "page_size"];
}

message ListWebhookEventsResponse {
    common.BaseResponse base = 1;
    repeated WebhookEvent items = 2;
    int32 total = 3;
}

message RetryWebhookEventRequest {
    int64 event_id = 1 [(api.body) = "event_id"];
}

service WebhookEventService {
    rpc List(ListWebhookEventsRequest) returns (ListWebhookEventsResponse) {
        option (api.get) = "/api/v1/webhook/events";
    }
    rpc Retry(RetryWebhookEventRequest) returns (common.BaseResponse) {
        option (api.post) = "/api/v1/webhook/events/retry";
    }
}
```

### 6.3 Provider Config Proto

```protobuf
// idl/biz/provider.proto

syntax = "proto3";
package provider;
option go_package = "github.com/yi-nology/git-manage-service/biz/model/provider";

import "api.proto";
import "common.proto";

message ProviderConfig {
    int64 id = 1;
    string name = 2;
    string platform = 3;      // gitlab, github, gitea
    string base_url = 4;
    int64 credential_id = 5;
    string webhook_secret = 6;
    string created_at = 7;
    string updated_at = 8;
}

message CreateProviderConfigRequest {
    string name = 1 [(api.body) = "name"];
    string platform = 2 [(api.body) = "platform"];
    string base_url = 3 [(api.body) = "base_url"];
    int64 credential_id = 4 [(api.body) = "credential_id"];
    string webhook_secret = 5 [(api.body) = "webhook_secret"];
}

message UpdateProviderConfigRequest {
    int64 id = 1 [(api.body) = "id"];
    string name = 2 [(api.body) = "name"];
    string base_url = 3 [(api.body) = "base_url"];
    int64 credential_id = 4 [(api.body) = "credential_id"];
    string webhook_secret = 5 [(api.body) = "webhook_secret"];
}

message DeleteProviderConfigRequest {
    int64 id = 1 [(api.body) = "id"];
}

message GetProviderConfigRequest {
    int64 id = 1 [(api.query) = "id"];
}

message ListProviderConfigsRequest {
    string platform = 1 [(api.query) = "platform"];
}

message ListProviderConfigsResponse {
    common.BaseResponse base = 1;
    repeated ProviderConfig items = 2;
}

message TestProviderConfigRequest {
    int64 id = 1 [(api.body) = "id"];
}

message TestProviderConfigResponse {
    common.BaseResponse base = 1;
    bool connected = 2;
    string platform = 3;
    string user_name = 4;
}

service ProviderConfigService {
    rpc Create(CreateProviderConfigRequest) returns (common.BaseResponse) {
        option (api.post) = "/api/v1/provider/create";
    }
    rpc Get(GetProviderConfigRequest) returns (ListProviderConfigsResponse) {
        option (api.get) = "/api/v1/provider/detail";
    }
    rpc List(ListProviderConfigsRequest) returns (ListProviderConfigsResponse) {
        option (api.get) = "/api/v1/provider/list";
    }
    rpc Update(UpdateProviderConfigRequest) returns (common.BaseResponse) {
        option (api.post) = "/api/v1/provider/update";
    }
    rpc Delete(DeleteProviderConfigRequest) returns (common.BaseResponse) {
        option (api.post) = "/api/v1/provider/delete";
    }
    rpc Test(TestProviderConfigRequest) returns (TestProviderConfigResponse) {
        option (api.post) = "/api/v1/provider/test";
    }
}
```

---

## 7. 新增目录结构总览

```
biz/
├── service/
│   └── provider/                  # 新增
│       ├── provider.go            # 接口定义 + DTO
│       ├── manager.go             # Provider 实例管理
│       ├── detect.go              # URL → 平台检测
│       ├── gitlab/                # 新增
│       │   ├── provider.go
│       │   ├── mr.go
│       │   ├── webhook.go
│       │   └── types.go
│       ├── github/                # 新增
│       │   ├── provider.go
│       │   ├── pr.go
│       │   ├── webhook.go
│       │   └── types.go
│       └── gitea/                 # 新增
│           ├── provider.go
│           ├── pr.go
│           ├── webhook.go
│           └── types.go
├── handler/
│   ├── cr/                        # 新增
│   │   └── cr_service.go
│   ├── provider/                  # 新增
│   │   └── provider_config_service.go
│   ├── webhook_event/             # 新增
│   │   └── webhook_event_service.go
│   └── webhook/
│       └── webhook_service.go     # 扩展: 新增 /receive 入口
├── router/
│   ├── cr/                        # 新增 (Hz 生成)
│   ├── provider/                  # 新增 (Hz 生成)
│   └── webhook_event/             # 新增 (Hz 生成)
├── model/
│   ├── cr/                        # 新增 (Hz 生成)
│   ├── provider/                  # 新增 (Hz 生成)
│   └── webhook_event/             # 新增 (Hz 生成)
├── model/po/                      # 新增
│   ├── provider_config.go
│   ├── change_request.go
│   ├── webhook_event.go
│   └── webhook_rule.go
├── dal/db/                        # 新增
│   ├── provider_config_dao.go
│   ├── change_request_dao.go
│   ├── webhook_event_dao.go
│   └── webhook_rule_dao.go
├── service/
│   ├── cr_service.go              # 新增
│   └── webhook_event_service.go   # 新增
└── middleware/
    └── webhook_auth.go            # 新增: 签名验证中间件

idl/biz/
├── provider.proto                 # 新增
├── cr.proto                       # 新增
└── webhook_event.proto            # 新增
```

---

## 8. API 汇总

### 8.1 平台配置

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/provider/create` | 添加平台连接 |
| GET | `/api/v1/provider/detail` | 查询配置详情 |
| GET | `/api/v1/provider/list` | 列出所有平台配置 |
| POST | `/api/v1/provider/update` | 更新配置 |
| POST | `/api/v1/provider/delete` | 删除配置 |
| POST | `/api/v1/provider/test` | 测试连通性 |

### 8.2 CR/MR 管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/cr/create` | 创建 CR (跨平台) |
| GET | `/api/v1/cr/detail` | 查询 CR 详情 |
| GET | `/api/v1/cr/list` | 列出 CR |
| POST | `/api/v1/cr/merge` | 合并 CR |
| POST | `/api/v1/cr/close` | 关闭 CR |
| POST | `/api/v1/cr/sync` | 从平台同步 CR 到本地 |

### 8.3 Webhook 事件

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/webhooks/receive` | 接收平台 webhook (统一入口) |
| GET | `/api/v1/webhook/events` | 查询事件列表 |
| POST | `/api/v1/webhook/events/retry` | 重试失败事件 |

### 8.4 现有 API 不变

所有 `/api/v1/repo/*`、`/api/v1/branch/*`、`/api/v1/sync/*`、`/api/v1/credential/*` 等现有 API 完全不动。

---

## 9. 实现优先级与里程碑

### Phase 1: 基础设施（1 周）

```
[1] provider_config 表 + DAO + CRUD API
[2] Provider 接口定义 + Manager
[3] 平台 URL 自动检测
[4] Credential 扩展 (platform_token 类型)
[5] 数据库迁移
```

### Phase 2: GitLab Provider + CR（2 周）

```
[1] GitLab Provider 实现 (MR CRUD + Webhook 注册)
[2] CR 模块 (Proto → Handler → Service → DAO)
[3] CR 同步 (从 GitLab 拉取 MR 列表到本地)
[4] 集成测试
```

### Phase 3: GitHub/Gitea Provider（2 周）

```
[1] GitHub Provider 实现 (PR CRUD + Webhook)
[2] Gitea Provider 实现 (PR CRUD + Webhook)
[3] CR 跨平台统一查询
```

### Phase 4: Webhook 标准化（1 周）

```
[1] 统一入站 endpoint + 签名验证
[2] 事件解析 (各平台 → NormalizedEvent)
[3] webhook_event 存储 + 去重
[4] webhook_rule 路由 + 动作执行
[5] 事件重试机制
```

### Phase 5: 质量保障

```
[1] Webhook 丢失率 = 0: 事件去重 + 失败重试 + 死信队列
[2] CR 成功率 ≥ 99.9%: 幂等合并 + 冲突预检 + 重试
[3] 分支创建 < 500ms: 本地 git 操作已满足，平台 API 异步
```

---

## 10. 关键设计决策

| 决策 | 选择 | 原因 |
|------|------|------|
| Provider 用接口还是 HTTP client 封装 | **Go 接口** | 编译期类型安全，mock 友好 |
| CR 数据存本地还是只做透传 | **双写：本地存一份 + 平台操作** | 支持离线查询、统计分析、事件关联 |
| Webhook 入站是一个还是多个 endpoint | **统一 `/api/webhooks/receive`** | 通过 Header 区分平台，简化配置 |
| 平台 Token 放哪里 | **复用现有 Credential 体系** | 加密存储已实现，加 `platform_token` 类型即可 |
| 现有 git 操作层是否改造 | **不改** | git 协议操作和平台 API 操作是两个正交关注点 |
| Webhook 事件处理 | **异步 + 规则引擎** | 不阻塞响应，规则匹配决定后续动作 |

---

## 11. 与现有系统的集成点

```
                        ┌──────────────────────────────────┐
                        │         新增模块                   │
                        │  Provider / CR / Webhook Event    │
                        └───┬──────────┬──────────┬────────┘
                            │          │          │
                 ┌──────────┘          │          └──────────┐
                 ▼                     ▼                     ▼
          ┌─────────────┐   ┌─────────────────┐   ┌──────────────┐
          │ Credential  │   │   Repo Model    │   │  Notification │
          │ (扩展类型)   │   │ (扩展平台字段)  │   │  (复用现有)   │
          └─────────────┘   └─────────────────┘   └──────────────┘
                                                     ▲
          ┌──────────────────────────────────────────┘
          │  Webhook Rule 触发:
          │  - sync → 调用现有 SyncService
          │  - notify → 调用现有 NotificationService
          └──────────────────────────────────────────
```

**复用点**：
- **Credential 系统**：直接新增 `platform_token` 类型，Token 加密存储已实现
- **Repo 模型**：扩展字段即可，不影响现有逻辑
- **Notification 系统**：webhook rule 的 `notify` 动作直接复用
- **Sync 系统**：webhook rule 的 `sync` 动作直接调用现有同步能力
- **Audit 系统**：所有 CR/Webhook 操作自动产生审计日志
