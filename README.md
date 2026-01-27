# Git Manage Service (Git 管理服务)

Git Manage Service 是一个轻量级的多仓库、多分支自动化同步管理系统。它提供了友好的 Web 界面，支持定时任务、Webhook 触发以及详细的同步日志记录。

## 🚀 功能特性

- **多仓库管理**：轻松注册和管理本地 Git 仓库。
- **灵活同步规则**：支持任意 Remote 和分支之间的同步（如 `origin/main` -> `ky/main`）。
- **自动化执行**：内置 Cron 调度器，支持定时同步。
- **Webhook 集成**：支持通过外部系统（如 CI/CD）触发同步。
- **安全可靠**：支持冲突检测、Fast-Forward 检查及 Force Push 保护。
- **可视化界面**：提供直观的 Web UI，查看历史、日志及管理任务。

## 📚 文档

- [产品手册与使用说明](docs/product_manual.md)
- [Webhook 接口文档](docs/webhook.md)

## 🛠 快速开始

### 方式一：下载预编译二进制文件（推荐）

从 [Releases](https://github.com/yi-nology/git-manage-service/releases) 页面下载适合你系统的版本：

- **Linux (AMD64)**: `git-manage-service-linux-amd64.tar.gz`
- **Linux (ARM64)**: `git-manage-service-linux-arm64.tar.gz`
- **macOS (Intel)**: `git-manage-service-darwin-amd64.tar.gz`
- **macOS (Apple Silicon)**: `git-manage-service-darwin-arm64.tar.gz`
- **Windows (AMD64)**: `git-manage-service-windows-amd64.exe.zip`
- **Windows (ARM64)**: `git-manage-service-windows-arm64.exe.zip`

#### Linux / macOS
```bash
# 解压
tar -xzf git-manage-service-*.tar.gz

# 添加执行权限
chmod +x git-manage-service-*

# 运行
./git-manage-service-*
```

#### Windows
```powershell
# 解压 zip 文件
# 双击运行或在命令行中执行
.\git-manage-service-windows-amd64.exe
```

### 方式二：从源码编译
```bash
# 安装依赖
go mod tidy

# 编译
go build -o git-manage-service main.go

# 运行
./git-manage-service
```

### 访问界面
浏览器打开: [http://localhost:8080](http://localhost:8080)

### 查看版本信息
```bash
./git-manage-service --version
```

## 📦 项目结构
```
.
├── biz/            # 业务逻辑 (Service, Handler, Model)
├── docs/           # 项目文档
├── public/         # 前端静态资源
├── test/           # 测试工具
├── main.go         # 入口文件
└── go.mod          # 依赖定义
```

## 🔨 开发者指南

### 创建新版本发布

本项目使用 GitHub Actions 自动构建多平台二进制文件。要创建新的发布版本：

1. **创建版本标签**
```bash
# 创建并推送标签
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

2. **自动构建**
   - GitHub Actions 会自动检测到标签推送
   - 自动构建 6 个平台的二进制文件：
     - Linux (AMD64/ARM64)
     - macOS (Intel/Apple Silicon)
     - Windows (AMD64/ARM64)
   - 自动创建 GitHub Release 并上传构建产物

3. **手动触发**（可选）
   - 访问 GitHub Actions 页面
   - 选择 "Release Build" 工作流
   - 点击 "Run workflow" 按钮手动触发

### 本地构建多平台版本

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o git-manage-service-linux-amd64 main.go

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o git-manage-service-linux-arm64 main.go

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o git-manage-service-darwin-amd64 main.go

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o git-manage-service-darwin-arm64 main.go

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o git-manage-service-windows-amd64.exe main.go

# Windows ARM64
GOOS=windows GOARCH=arm64 go build -o git-manage-service-windows-arm64.exe main.go
```

## 📝 License
MIT
