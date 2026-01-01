# Git Manage Service - 整体优化与迭代计划

本文档旨在规划 `git-manage-service` 的长期演进路线，从架构重构、用户体验提升到业务功能扩展，旨在打造一个生产级、高可用的 Git 分支管理平台。

## 1. 业务架构与后端重构 (Architecture & Backend)

### 1.1 代码分层与规范 (Layering)
当前代码中 `Handler` 层承担了过多逻辑。
- **Action**: 严格执行 `Handler` -> `Service` -> `DAL` (Data Access Layer) 分层。
- **DTO 分离**: 引入 `model/dto` 和 `model/vo`，将数据库实体 (`model.Repo`) 与 API 请求/响应结构解耦。
- **统一错误处理**: 封装自定义 `AppError`，包含错误码、HTTP 状态码和用户提示信息，替换零散的 `c.JSON(500, err)`。

### 1.2 系统稳定性 (Stability)
- **并发控制**: 优化 Git 操作的并发池 (`gopool`)，防止大量克隆/同步任务耗尽服务器资源。
- **事务管理**: 确保涉及数据库变更和 Git 操作的业务（如删除仓库）具备事务性或补偿机制。
- **结构化日志**: 引入 `zap` 或 `logrus`，替换标准库 log，实现结构化日志记录（TraceID, Level, Context）。

### 1.3 接口标准化 (API Design)
- **RESTful 规范**: 统一 URL 命名和 HTTP Method 使用。
- **分页与过滤**: 为列表接口（如提交记录、同步历史）增加标准的分页 (`page`, `page_size`) 和排序参数。

## 2. 前端与 UI/UX 设计 (Frontend & Experience)

### 2.1 代码组织 (Refactoring)
- **JS 模块化**: 将 `repos.html` 等文件中数千行的 `<script>` 拆分为独立的 `.js` 模块 (ES6 Modules)。
- **通用请求库**: 封装 `request.js`，统一处理 API Base URL、Token 认证、401/500 错误拦截和全局 Loading 状态。

### 2.2 交互体验 (UX)
- **告别 Alert**: 使用 **Toast** (轻量提示) 和 **SweetAlert/Modal** 替代原生的 `alert()` 和 `confirm()`，提升质感。
- **加载反馈**: 在表格数据加载时使用 **骨架屏 (Skeleton)** 或更优雅的 Loading Spinner，而不是简单的文字。
- **响应式布局**: 优化移动端适配，确保在手机/平板上也能查看状态。

### 2.3 视觉升级 (UI)
- **Dashboard**: 首页展示更丰富的数据仪表盘（成功率、活跃仓库、待处理异常）。
- **状态可视化**: 使用更直观的图标和颜色区分同步状态（成功、失败、冲突）。

## 3. 业务功能扩展 (Business Features)

### 3.1 通知与告警 (Notifications)
- **需求**: 当同步任务失败或产生冲突时，主动通知用户。
- **实现**: 集成 Webhook (Outbound)，支持配置 **钉钉、飞书、Slack** 或 **Email** 告警。

### 3.2 审计与安全 (Audit & Security)
- **审计日志**: 记录所有敏感操作（谁、在什么时间、删除了哪个仓库/任务）。
- **只读模式**: 增加系统级配置，允许将系统设为维护模式（暂停所有写操作）。

### 3.3 高级 Git 能力
- **冲突预检**: 在执行同步前，尝试进行内存中的 Merge 预检 (Dry Run)，提前预警冲突。
- **标签同步**: 支持 Tag 的自动同步。

---

## 4. 迭代阶段规划 (Milestones)

### Phase 1: 基础重构 (Foundation)
- [ ] 封装统一的 Result/Error 响应结构。
- [ ] 拆分前端 JS 代码，引入 Toast 组件。
- [ ] 优化 `main.go` 启动逻辑，使用依赖注入风格初始化。

### Phase 2: 体验升级 (Experience)
- [ ] 重写首页 Dashboard，展示核心指标。
- [ ] 优化同步任务列表页，增加筛选、分页。
- [ ] 增加操作日志 (Audit Log) 页面。

### Phase 3: 高级功能 (Advanced)
- [ ] 实现同步失败的外部通知 (Webhook/Email)。
- [ ] 增加 Git 冲突预检功能。
- [ ] 完善 Docker 部署与 CI/CD 配置。
