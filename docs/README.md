---
home: true
heroImage: /images/logo.svg
heroText: Git Manage Service
tagline: 轻量级多仓库自动化同步管理系统
actionText: 快速开始 →
actionLink: /getting-started
features:
  - title: 📦 多仓库管理
    details: 轻松注册和管理本地 Git 仓库，支持仓库克隆、导入和可视化浏览
  - title: 🔄 灵活同步规则
    details: 支持任意 Remote 和分支之间的同步，如 origin/main → backup/main
  - title: ⏰ 自动化执行
    details: 内置 Cron 调度器，支持定时同步任务，也可通过 Webhook 触发
  - title: 🔔 多渠道通知
    details: 支持钉钉、企业微信、飞书、蓝信、邮件、自定义 Webhook
  - title: 🔐 SSH 密钥管理
    details: 统一管理 SSH 密钥，支持将密钥存储在数据库中
  - title: 📊 代码质量分析
    details: 提交统计、贡献者排行、代码度量等分析功能
  - title: 📝 Spec 编辑器
    details: 集成 Monaco Editor 的 RPM Spec 文件编辑器，支持实时语法检查
  - title: 🩹 Patch 管理
    details: 生成、管理和应用 Git Patch 文件，支持批量操作
footer: MIT Licensed | Copyright © 2024-present
---

<div class="quick-links">

## 📸 文档预览

![文档首页](/images/docs/docs-home.png)

## 🚀 5 分钟上手

```bash
# 下载
wget https://github.com/yi-nology/git-manage-service/releases/download/v0.7.2/git-manage-service-darwin-arm64.tar.gz

# 解压
tar -xzf git-manage-service-*.tar.gz

# 运行
./git-manage-service

# 访问
open http://localhost:38080
```

[查看完整安装指南 →](/getting-started)

</div>

<div class="features-preview">

## 📸 功能预览

| 仓库管理 | 分支操作 |
|:---:|:---:|
| ![仓库列表](/images/repo-list-with-data.png) | ![分支管理](/images/branch-management.png) |

| 同步任务 | 代码度量 |
|:---:|:---:|
| ![同步任务](/images/sync-tasks.png) | ![Git 度量](/images/git-metrics.png) |

| 审计日志 | 通知配置 |
|:---:|:---:|
| ![审计日志](/images/audit-log.png) | ![通知渠道](/images/notification-channel.png) |

</div>

<div class="tech-stack">

## 🛠 技术栈

| 后端 | 前端 |
|------|------|
| Go 1.25 | Vue 3 |
| CloudWeGo Hertz | Element Plus |
| CloudWeGo Kitex | Pinia |
| SQLite / MySQL / PostgreSQL | ECharts |
| Redis (可选) | Monaco Editor |
| MinIO (可选) | TypeScript |

</div>

<div class="doc-nav">

## 📖 文档导航

| 类型 | 内容 |
|------|------|
| 🚀 [快速开始](/getting-started) | 5 分钟完成安装和基本配置 |
| 📘 [功能指南](/features/repo) | 详细的功能使用说明 |
| 📦 [部署方案](/deployment/binary) | 生产环境部署指南 |
| ⚙️ [配置参考](/configuration) | 完整的配置项说明 |
| 🔌 [API 文档](/api) | HTTP API 接口参考 |

</div>

<style>
.quick-links {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 20px;
  margin: 20px 0;
}

.features-preview table {
  width: 100%;
}

.features-preview img {
  max-width: 100%;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.1);
}

.tech-stack table {
  width: 100%;
}

.doc-nav table {
  width: 100%;
}
</style>
