# Desktop Application Build Guide

本文档说明如何将 Git Manage Service 打包成桌面应用程序。

## 🎯 支持的平台

| 平台 | 格式 | 说明 |
|------|------|------|
| **macOS** | `.app` | Universal Binary (Intel + Apple Silicon) |
| **Windows** | `.exe` | Windows 安装程序 (NSIS) |
| **Linux** | `.deb`, `.rpm`, `.AppImage` | 多种 Linux 发行版支持 |

## 📋 前置要求

### macOS
```bash
# 安装 Xcode Command Line Tools
xcode-select --install

# 安装 Go (1.21+)
brew install go

# 安装 Node.js (20+)
brew install node
```

### Windows
```powershell
# 安装 Visual Studio Build Tools (包含 C++ 工具链)
# https://visualstudio.microsoft.com/downloads/

# 安装 Go
# https://golang.org/dl/

# 安装 Node.js
# https://nodejs.org/
```

### Linux (Ubuntu/Debian)
```bash
# 安装依赖
sudo apt-get update
sudo apt-get install -y \
  libgtk-3-dev \
  libwebkit2gtk-4.0-dev \
  build-essential

# 安装 Go
sudo snap install go --classic

# 安装 Node.js
sudo snap install node --classic --channel=20
```

## 🚀 快速开始

### 1. 设置环境

```bash
# 运行设置脚本
chmod +x script/setup-desktop.sh
./script/setup-desktop.sh
```

### 2. 构建应用

#### 构建当前平台
```bash
make desktop
```

#### 构建特定平台
```bash
# macOS
make desktop-darwin

# Windows (需要在 Windows 系统上运行)
make desktop-windows

# Linux (需要在 Linux 系统上运行)
make desktop-linux
```

#### 构建所有平台
```bash
# 注意：需要使用 CI/CD 或在各自平台上构建
make desktop-all
```

## 📦 输出位置

构建完成后，应用会生成在 `build/bin/` 目录：

```
build/bin/
├── Git Manage Service.app    # macOS
├── git-manage-desktop.exe    # Windows
└── git-manage-desktop        # Linux
```

## 🔧 高级配置

### 自定义应用信息

编辑 `wails.json`:

```json
{
  "name": "Git Manage Service",
  "outputfilename": "git-manage-desktop",
  "info": {
    "companyName": "yi-nology",
    "productName": "Git Manage Service",
    "productVersion": "0.7.5",
    "copyright": "Copyright © 2026 yi-nology",
    "comments": "Git repository management desktop application"
  }
}
```

### 添加应用图标

1. 准备图标文件：
   - macOS: `build/appicon.icns`
   - Windows: `build/appicon.ico`
   - Linux: `build/appicon.png`

2. 使用工具生成：
```bash
# 安装 icon 生成工具
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 从 PNG 生成所有格式
wails generate icon -input logo.png
```

### 添加系统托盘

在 `desktop-app.go` 中添加：

```go
import "github.com/wailsapp/wails/v2/pkg/menu"

func (a *App) menu() *menu.Menu {
    // 创建菜单
    return menu.NewMenu()
}
```

## 🎨 开发模式

运行桌面应用开发服务器：

```bash
wails dev
```

这会：
1. 启动前端开发服务器
2. 启动后端服务
3. 打开桌面应用窗口
4. 支持热重载

## 📊 构建选项

### macOS
```bash
# 仅 Intel
wails build -platform darwin/amd64

# 仅 Apple Silicon
wails build -platform darwin/arm64

# Universal Binary
wails build -platform darwin/universal

# 使用 UPX 压缩（需要安装 upx）
wails build -upx
```

### Windows
```bash
# 64位
wails build -platform windows/amd64

# 32位
wails build -platform windows/386

# 生成安装程序
wails build -nsis
```

### Linux
```bash
# DEB 包
wails build -deb

# RPM 包
wails build -rpm

# AppImage
wails build -appimage
```

## 🔄 CI/CD 自动构建

项目包含 GitHub Actions workflow，在推送 tag 时自动构建：

```bash
# 创建并推送 tag
git tag v0.7.5
git push origin v0.7.5

# GitHub Actions 会自动：
# 1. 构建 macOS/Windows/Linux 应用
# 2. 创建 GitHub Release
# 3. 上传所有构建产物
```

手动触发构建：
1. 进入 GitHub Actions
2. 选择 "Desktop Build" workflow
3. 点击 "Run workflow"

## 🐛 常见问题

### macOS: "无法打开，因为它来自身份不明的开发者"

```bash
# 允许运行
xattr -cr "Git Manage Service.app"

# 或在系统偏好设置中允许
```

### Windows: 缺少 DLL

安装 Visual C++ Redistributable:
https://support.microsoft.com/en-us/help/2977003/the-latest-supported-visual-c-downloads

### Linux: webkit2gtk 错误

```bash
# Ubuntu/Debian
sudo apt-get install libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install webkit2gtk3-devel
```

## 📚 相关文档

- [Wails 官方文档](https://wails.io/docs/introduction)
- [应用图标生成](https://wails.io/docs/guides/icons)
- [Windows 签名](https://wails.io/docs/guides/signing)
- [自动更新](https://wails.io/docs/guides/updates)

## 🤝 贡献

如果您遇到问题或有改进建议，请：
1. 查看 [Issues](https://github.com/yi-nology/git-manage-service/issues)
2. 提交新的 Issue
3. 或提交 Pull Request
