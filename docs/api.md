# API 文档

本文档介绍 Git Manage Service 的 HTTP API 接口。

## API 概览

- **Base URL**: `http://localhost:38080/api/v1`
- **Content-Type**: `application/json`
- **认证方式**: 无（内部服务）

## 通用响应格式

### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 错误响应

```json
{
  "code": 400,
  "message": "错误描述",
  "data": null
}
```

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 仓库管理

### 获取仓库列表

```
GET /api/v1/repos
```

**响应示例**：

```json
{
  "code": 0,
  "data": {
    "repos": [
      {
        "id": 1,
        "name": "my-project",
        "path": "/home/git/repos/my-project",
        "current_branch": "main",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 1
  }
}
```

### 获取仓库详情

```
GET /api/v1/repos/:id
```

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 仓库 ID |

### 创建仓库（注册）

```
POST /api/v1/repos
```

**请求体**：

```json
{
  "name": "my-project",
  "path": "/home/git/repos/my-project",
  "ssh_key_id": 1
}
```

### 删除仓库

```
DELETE /api/v1/repos/:id
```

## 分支管理

### 获取分支列表

```
GET /api/v1/repos/:id/branches
```

**查询参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| type | string | 分支类型：local/remote/all |

**响应示例**：

```json
{
  "code": 0,
  "data": {
    "branches": [
      {
        "name": "main",
        "type": "local",
        "is_current": true,
        "last_commit": "abc123"
      }
    ]
  }
}
```

### 切换分支

```
POST /api/v1/repos/:id/branches/checkout
```

**请求体**：

```json
{
  "branch": "develop"
}
```

### 创建分支

```
POST /api/v1/repos/:id/branches
```

**请求体**：

```json
{
  "name": "feature/new-feature",
  "base_branch": "main"
}
```

### 删除分支

```
DELETE /api/v1/repos/:id/branches/:branch
```

## 同步任务

### 获取任务列表

```
GET /api/v1/sync/tasks
```

### 获取任务详情

```
GET /api/v1/sync/tasks/:id
```

### 创建同步任务

```
POST /api/v1/sync/tasks
```

**请求体**：

```json
{
  "name": "main-to-backup",
  "repo_id": 1,
  "source_remote": "origin",
  "source_branch": "main",
  "target_remote": "backup",
  "target_branch": "main",
  "cron_expression": "0 */2 * * *",
  "enabled": true
}
```

### 执行同步任务

```
POST /api/v1/sync/tasks/:id/run
```

### 更新同步任务

```
PUT /api/v1/sync/tasks/:id
```

### 删除同步任务

```
DELETE /api/v1/sync/tasks/:id
```

## Webhook

### 触发同步

```
POST /api/webhooks/task-sync
```

**请求头**：

| Header | 说明 |
|--------|------|
| Content-Type | application/json |
| X-Hub-Signature-256 | HMAC-SHA256 签名 |

**请求体**：

```json
{
  "task_id": 1
}
```

详见 [Webhook 文档](/features/webhook)

## SSH 密钥

### 获取密钥列表

```
GET /api/v1/ssh-keys
```

### 创建密钥

```
POST /api/v1/ssh-keys
```

**请求体**：

```json
{
  "name": "github-deploy-key",
  "type": "rsa",
  "private_key": "-----BEGIN RSA PRIVATE KEY-----\n...",
  "description": "用于访问 GitHub 仓库"
}
```

### 删除密钥

```
DELETE /api/v1/ssh-keys/:id
```

### 测试密钥

```
POST /api/v1/ssh-keys/:id/test
```

**请求体**：

```json
{
  "repo_url": "git@github.com:org/repo.git"
}
```

## 通知渠道

### 获取渠道列表

```
GET /api/v1/notification/channels
```

### 创建通知渠道

```
POST /api/v1/notification/channels
```

**请求体**：

```json
{
  "type": "dingtalk",
  "name": "开发群通知",
  "config": {
    "webhook": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
    "secret": "SECxxx"
  },
  "events": ["sync_success", "sync_failure"],
  "template": "【同步{{.StatusText}}】任务: {{.TaskKey}}"
}
```

### 测试通知

```
POST /api/v1/notification/channels/:id/test
```

## 审计日志

### 获取日志列表

```
GET /api/v1/audit/logs
```

**查询参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| type | string | 操作类型 |
| object_type | string | 对象类型 |
| start_time | string | 开始时间 |
| end_time | string | 结束时间 |
| page | int | 页码 |
| page_size | int | 每页数量 |

## 系统信息

### 健康检查

```
GET /api/health
```

**响应示例**：

```json
{
  "status": "ok"
}
```

### 获取版本信息

```
GET /api/v1/system/version
```

**响应示例**：

```json
{
  "code": 0,
  "data": {
    "version": "v0.7.2",
    "build_time": "2024-01-01T00:00:00Z",
    "git_commit": "abc123"
  }
}
```

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 500 | 服务器内部错误 |

## 分页

列表接口支持分页：

```
GET /api/v1/repos?page=1&page_size=20
```

响应格式：

```json
{
  "code": 0,
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

## 下一步

- [Webhook 集成](/features/webhook) - 外部触发同步
- [配置参考](/configuration) - 服务配置
