# Wails Desktop Application - 构建指南

## 快速开始

### 本地开发

```bash
# 1. 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 2. 开发模式（热重载）
wails dev

# 3. 构建生产版本
wails build
```

### 构建特定平台

```bash
# macOS Universal Binary (Intel + Apple Silicon)
wails build -platform darwin/universal

# Windows 64-bit
wails build -platform windows/amd64

# Linux 64-bit
wails build -platform linux/amd64
```

## 架构说明

### 文件结构

```
git-manage-service/
├── desktop.go              # 桌面应用入口
├── main.go                 # Web 服务入口
├── frontend/               # Vue3 前端
│   ├── src/               # 源代码
│   ├── dist/              # 构建产物（嵌入到桌面应用）
│   └── wailsjs/           # Wails 前端绑定
├── wails.json             # Wails 配置
└── build/                 # 构建输出
    └── bin/
        ├── git-manage-desktop(.exe)  # 可执行文件
        └── Git Manage Service.app/   # macOS 应用包
```

### 工作原理

1. **启动流程**：
   - 桌面应用启动 → 打开图形窗口
   - 后台启动 HTTP 服务器（端口 38080）
   - 前端通过 HTTP 调用后端 API

2. **前端集成**：
   - 前端构建产物嵌入到二进制文件
   - 使用 Wails 提供的 WebView 显示
   - 前端代码无需修改

3. **后端服务**：
   - 复用 main.go 的服务器逻辑
   - 自动初始化数据库和配置
   - 支持所有 Web 版本功能

## 开发指南

### 前端开发

```bash
cd frontend

# 安装依赖
npm install

# 开发模式（连接到后端 API）
npm run dev

# 构建生产版本
npm run build
```

### 后端开发

修改 `desktop.go` 中的 `startBackend()` 方法来自定义后端启动逻辑。

### 调试

```bash
# 开启调试模式
wails dev -debug

# 打开开发者工具（运行时）
Ctrl+Shift+I (Windows/Linux)
Cmd+Option+I (macOS)
```

## 常见问题

### 1. 前端资源未嵌入

**症状**：运行时报错 `no such file or directory`

**解决**：
```bash
# 确保前端已构建
cd frontend
npm run build

# 重新构建桌面应用
wails build
```

### 2. macOS 无法打开应用

**症状**：`无法打开，因为它来自身份不明的开发者`

**解决**：
```bash
# 允许运行
xattr -cr "Git Manage Service.app"

# 或在系统偏好设置中允许
```

### 3. Windows 缺少 DLL

**症状**：运行时报错缺少 DLL

**解决**：
- 安装 Visual C++ Redistributable
- 或在构建时使用 `-upx` 压缩

### 4. Linux 依赖问题

**症状**：`webkit2gtk not found`

**解决**：
```bash
# Ubuntu/Debian
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev

# Fedora
sudo dnf install gtk3-devel webkit2gtk4.1-devel
```

## CI/CD 自动构建

### 触发方式

1. **自动触发**：推送 tag `v*-desktop`
   ```bash
   git tag v0.8.0-desktop
   git push origin v0.8.0-desktop
   ```

2. **手动触发**：
   - 进入 GitHub Actions
   - 选择 "Desktop Build" workflow
   - 点击 "Run workflow"

### 构建产物

- macOS: `GitManageService-macOS.zip` (Universal Binary)
- Windows: `GitManageService-Windows.zip`
- Linux: `GitManageService-Linux.tar.gz`

## 性能优化

### 减小应用体积

```bash
# 使用 UPX 压缩
wails build -upx

# 裁剪调试信息
wails build -ldflags "-s -w"
```

### 优化前端

编辑 `frontend/vite.config.ts`：
```typescript
export default defineConfig({
  build: {
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
      },
    },
  },
})
```

## 下一步

1. **添加系统托盘**：参考 [Wails 托盘示例](https://wails.io/docs/guides/tray)
2. **自动更新**：集成 [go-update](https://github.com/inconshreveable/go-update)
3. **原生菜单**：使用 Wails Menu API
4. **窗口状态**：保存/恢复窗口位置和大小

## 相关文档

- [Wails 官方文档](https://wails.io/docs/introduction)
- [Wails 示例项目](https://github.com/wailsapp/wails/tree/master/samples)
- [前端集成指南](https://wails.io/docs/guides/frontend)
