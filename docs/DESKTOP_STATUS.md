# 桌面应用构建说明

## 当前状态

桌面应用功能已添加，但在实现上存在一些技术限制需要解决。

## 问题诊断

### 1. Wails 集成问题

**问题**：
- Wails 需要专门的前端构建配置
- 当前前端使用 Vue3 + Vite，需要适配 Wails 的资源嵌入方式
- desktop 构建标签导致构建时运行桌面应用，端口冲突

**解决方案**：
```bash
# 方案 A：完整 Wails 集成（推荐但需要前端修改）
1. 修改 frontend/vite.config.ts 适配 Wails
2. 创建 Wails 前端绑定
3. 修改前端代码使用 Wails 运行时 API

# 方案 B：混合方案（更简单）
1. 桌面应用作为包装器启动后端服务
2. 使用系统 WebView 显示 localhost:38080
3. 不需要修改前端代码
```

### 2. 构建流程

**当前 CI/CD**：
- `.github/workflows/desktop.yml` 已配置
- 需要修复 Wails 构建参数
- 需要处理前端资源路径

## 推荐方案

### 方案 A：渐进式实现（分步骤）

#### 阶段 1：简单打包器（1-2天）
创建独立启动器：
```bash
# macOS
./script/create-macos-app.sh
# 生成 GitManageService.app（包含后端服务）
# 启动后自动打开浏览器访问 http://localhost:38080
```

#### 阶段 2：WebView 集成（3-5天）
使用系统 WebView：
- macOS: WKWebView
- Windows: WebView2
- Linux: WebKitGTK

#### 阶段 3：完整 Wails 集成（可选，1-2周）
完全重写前端集成，使用 Wails 运行时

### 方案 B：使用 Electron（快速但体积大）

使用 Electron 打包现有 Web 应用：
```bash
# 安装 Electron
npm install --save-dev electron electron-builder

# 创建主进程
# main.js
const { app, BrowserWindow } = require('electron')
const { spawn } = require('child_process')
const path = require('path')

function createWindow() {
  const win = new BrowserWindow({
    width: 1280,
    height: 800,
    webPreferences: {
      nodeIntegration: false
    }
  })
  
  // 启动后端服务
  const backend = spawn('./git-manage-service', ['--mode=all'])
  
  // 等待服务启动
  setTimeout(() => {
    win.loadURL('http://localhost:38080')
  }, 3000)
}

app.whenReady().then(createWindow)
```

## 立即可用的方案

### macOS App Bundle（无需 Wails）

```bash
cd /opt/project/git-manage-service

# 1. 构建 Web 服务
make build

# 2. 创建应用包
chmod +x script/create-macos-app.sh
./script/create-macos-app.sh

# 3. 测试
open /tmp/GitManageService.app
```

### Windows 快捷方式

创建 `GitManageService.bat`:
```batch
@echo off
start git-manage-service.exe --mode=all
timeout /t 3
start http://localhost:38080
```

### Linux Desktop Entry

创建 `git-manage-service.desktop`:
```ini
[Desktop Entry]
Version=0.8.0
Name=Git Manage Service
Exec=/opt/git-manage-service/git-manage-service --mode=all
Icon=git-manage-service
Terminal=false
Type=Application
Categories=Development;
```

## 下一步行动

### 建议 1：先实现简单打包器
- 成本：1-2天
- 效果：用户可以双击 .app/.exe 运行
- 限制：使用系统浏览器，不是独立窗口

### 建议 2：修复 Wails 集成（如果坚持使用）
- 成本：3-5天
- 需要修改前端代码
- 需要深入理解 Wails 架构

### 建议 3：使用 Electron
- 成本：2-3天
- 体积较大（~150MB）
- 但功能完整，开发快速

## 我的建议

**短期（本周）**：
1. 暂时禁用桌面应用 CI/CD
2. 实现简单的应用打包器（脚本创建 .app/.exe）
3. 在 Release 中提供打包好的应用

**中期（下周）**：
1. 评估 Electron vs Wails
2. 选择一个方向深入实现
3. 完善桌面应用功能

**长期（可选）**：
- 完整的原生集成
- 系统托盘
- 自动更新
- 原生菜单

---

需要我帮你：
1. 实现简单打包器？
2. 修复 Wails 集成？
3. 设置 Electron 方案？
