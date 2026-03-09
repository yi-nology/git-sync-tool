# 通知配置

通知功能支持在同步任务执行完成后，通过多种渠道发送通知消息，帮助你及时了解任务状态。

## 功能概览

![通知渠道配置](images/notification-channel.png)

- **多渠道支持**: 钉钉、企业微信、飞书、蓝信、邮件、自定义 Webhook
- **灵活触发**: 支持 8 种触发事件
- **消息模板**: 使用 Go 模板语法自定义消息内容
- **两级配置**: 渠道级默认模板 + 事件级独立模板

## 支持的通知渠道

| 渠道 | 说明 | 适用场景 |
|------|------|----------|
| 钉钉 | 钉钉机器人 Webhook | 国内团队 |
| 企业微信 | 企微机器人 Webhook | 国内团队 |
| 飞书 | 飞书机器人 Webhook | 国内团队 |
| 蓝信 | 蓝信机器人 Webhook | 政企用户 |
| 邮件 | SMTP 邮件发送 | 通用通知 |
| 自定义 Webhook | 通用 HTTP 回调 | 自定义集成 |

## 触发事件

| 事件 | 代码 | 说明 |
|------|------|------|
| 同步成功 | `sync_success` | 同步任务执行成功 |
| 同步失败 | `sync_failure` | 同步任务执行失败 |
| 同步冲突 | `sync_conflict` | 检测到代码冲突 |
| 备份成功 | `backup_success` | 仓库备份成功 |
| 备份失败 | `backup_failure` | 仓库备份失败 |
| 定时任务开始 | `cron_start` | 定时任务开始执行 |
| 定时任务结束 | `cron_end` | 定时任务执行结束 |
| Webhook 触发 | `webhook_trigger` | 通过 Webhook 触发 |

## 添加通知渠道

### 操作步骤

1. 点击左侧导航 **"系统设置"**
2. 进入 **"通知渠道"** 标签
3. 点击 **"添加渠道"** 按钮
4. 填写渠道配置
5. 保存并测试

### 钉钉配置

1. 在钉钉群中添加自定义机器人
2. 获取 Webhook URL
3. （可选）设置加签密钥

```yaml
类型: 钉钉
名称: 开发群通知
Webhook: https://oapi.dingtalk.com/robot/send?access_token=xxx
密钥: SECxxx（加签时需要）
```

### 企业微信配置

1. 在企微群中添加群机器人
2. 获取 Webhook URL

```yaml
类型: 企业微信
名称: 项目群通知
Webhook: https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx
```

### 飞书配置

1. 在飞书中创建自定义机器人
2. 获取 Webhook URL
3. （可选）设置签名密钥

```yaml
类型: 飞书
名称: 运维群通知
Webhook: https://open.feishu.cn/open-apis/bot/v2/hook/xxx
密钥: xxx（签名时需要）
```

### 邮件配置

```yaml
类型: 邮件
名称: 邮件通知
SMTP 服务器: smtp.example.com
端口: 465
用户名: user@example.com
密码: xxx
发件人: noreply@example.com
收件人: team@example.com
```

### 自定义 Webhook

```yaml
类型: 自定义 Webhook
名称: 自定义通知
URL: https://api.example.com/notify
方法: POST
请求头: |
  Authorization: Bearer xxx
  Content-Type: application/json
```

## 消息模板

### 模板语法

使用 Go 模板语法，格式为双花括号包裹变量名：

```go
标题：[状态文字] 任务标识 同步通知
内容：任务 任务标识 于 时间 执行状态文字
      源远程/源分支 -> 目标远程/目标分支
      耗时: 执行耗时
      错误: 错误信息（如有）
```

### 可用变量

| 变量 | 说明 | 适用事件 |
|------|------|----------|
| `.TaskKey` | 任务标识 | 全部 |
| `.Status` | 状态码 (success/failure) | 全部 |
| `.StatusText` | 状态文字 (成功/失败) | 全部 |
| `.EventType` | 事件类型 | 全部 |
| `.EventLabel` | 事件名称 | 全部 |
| `.Timestamp` | 时间 | 全部 |
| `.RepoKey` | 仓库标识 | 全部 |
| `.SourceRemote` | 源远程仓库 | 同步事件 |
| `.SourceBranch` | 源分支 | 同步事件 |
| `.TargetRemote` | 目标远程仓库 | 同步事件 |
| `.TargetBranch` | 目标分支 | 同步事件 |
| `.ErrorMessage` | 错误信息 | 失败/错误/冲突 |
| `.CommitRange` | 提交范围 | 同步成功 |
| `.Duration` | 执行耗时 | 同步/备份 |
| `.CronExpression` | Cron 表达式 | 定时任务 |
| `.WebhookSource` | Webhook 来源 | Webhook 事件 |
| `.BackupPath` | 备份路径 | 备份事件 |

::: tip 使用方式
在模板中使用变量时，用双花括号包裹变量名。例如：左花括号左花括号 `.TaskKey` 右花括号右花括号。
:::

### 模板示例

#### 简单模板

```go
【Git 同步通知】
任务: 任务标识变量
状态: 状态文字变量
时间: 时间变量
```

#### 详细模板

```go
【同步状态文字】
任务: 任务标识
仓库: 仓库标识
源远程/源分支 -> 目标远程/目标分支
耗时: 执行耗时
提交: 提交范围（如有）
错误: 错误信息（如有）
```

## 两级模板配置

### 渠道级默认模板

在添加通知渠道时设置，适用于所有事件。

### 事件级独立模板

为特定事件设置独立模板，优先级高于渠道级模板。

1. 进入通知渠道详情
2. 点击 **"事件模板"** 标签
3. 为需要的事件设置模板
4. 留空则使用渠道级默认模板

## 测试通知

配置完成后，点击 **"测试"** 按钮发送测试消息，验证配置是否正确。

## 最佳实践

1. **关键事件通知**: 至少配置同步失败和冲突通知
2. **消息简洁**: 保持消息简洁明了，突出关键信息
3. **分级通知**: 重要事件发送多渠道通知，普通事件单渠道即可
4. **定期检查**: 定期检查通知渠道是否正常工作

## 下一步

- [同步任务](/features/sync) - 创建同步任务
- [Webhook 集成](/features/webhook) - 外部触发同步
