# Git Manage Service

<p align="center">
  <img src="docs/.vuepress/public/images/logo.svg" alt="Git Manage Service" width="128" height="128">
</p>

<p align="center">
  <strong>轻量级多仓库自动化同步管理系统</strong>
</p>

<p align="center">
  <a href="https://github.com/yi-nology/git-manage-service/releases">
    <img src="https://img.shields.io/github/v/release/yi-nology/git-manage-service?include_prereleases" alt="Release">
  </a>
  <a href="https://github.com/yi-nology/git-manage-service/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/yi-nology/git-manage-service" alt="License">
  </a>
  <a href="https://github.com/yi-nology/git-manage-service/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/yi-nology/git-manage-service/release.yml" alt="Build">
  </a>
</p>

---

## ✨ 功能特性

| 功能 | 说明 |
|------|------|
| 📦 **多仓库管理** | 轻松注册和管理本地 Git 仓库 |
| 🔄 **灵活同步规则** | 支持任意 Remote 和分支之间的同步 |
| ⏰ **自动化执行** | 内置 Cron 调度器，支持定时同步 |
| 🔔 **多渠道通知** | 钉钉、企微、飞书、邮件、自定义 Webhook |
| 🔐 **SSH 密钥管理** | 统一管理 SSH 密钥，安全访问私有仓库 |
| 📊 **代码度量** | 提交统计、贡献者排行、代码质量分析 |
| 🔌 **Webhook 集成** | 支持外部系统触发同步 |
| 📝 **审计日志** | 完整的操作日志记录 |

## 🚀 快速开始

### 方式 1：桌面应用（推荐）🖥️

从 [Releases](https://github.com/yi-nology/git-manage-service/releases) 下载桌面应用：

| 平台 | 文件 | 说明 |
|------|------|------|
| **macOS** | `GitManageService-macOS.zip` | Universal Binary (Intel + M1/M2) |
| **Windows** | `GitManageService-Windows.zip` | 包含安装程序 |
| **Linux** | `GitManageService-Linux.tar.gz` | DEB/RPM/AppImage |

```bash
# macOS
unzip GitManageService-macOS.zip
open "Git Manage Service.app"

# Windows
# 双击 git-manage-desktop.exe 或运行安装程序

# Linux
tar -xzf GitManageService-Linux.tar.gz
sudo dpkg -i git-manage-desktop.deb  # DEB
# 或
./git-manage-desktop.AppImage         # AppImage
```

### 方式 2：Web 服务

从 [Releases](https://github.com/yi-nology/git-manage-service/releases) 下载 Web 服务版本：

```bash
# macOS Apple Silicon
wget https://github.com/yi-nology/git-manage-service/releases/download/v0.7.2/git-manage-service-darwin-arm64.tar.gz

# Linux AMD64
wget https://github.com/yi-nology/git-manage-service/releases/download/v0.7.2/git-manage-service-linux-amd64.tar.gz
```

### 运行服务

```bash
# 解压
tar -xzf git-manage-service-*.tar.gz

# 运行
./git-manage-service

# 访问
open http://localhost:38080
```

### Docker 部署

```bash
docker run -d \
  --name git-manage-service \
  -p 38080:38080 \
  -v $(pwd)/data:/app/data \
  ghcr.io/yi-nology/git-manage-service:latest
```

## 📖 文档

| 文档 | 说明 |
|------|------|
| [快速开始](docs/getting-started.md) | 5 分钟完成安装和配置 |
| [功能指南](docs/README.md) | 详细的功能使用说明 |
| [部署方案](docs/deployment/binary.md) | 生产环境部署指南 |
| [配置参考](docs/configuration.md) | 完整的配置项说明 |
| [API 文档](docs/api.md) | HTTP API 接口参考 |
| [Webhook 集成](docs/features/webhook.md) | 外部触发同步 |
| [MCP 对接](docs/features/mcp.md) | MCP 服务对接指南 |

## 📡 MCP 服务对接

### 什么是 MCP

MCP (Model Context Protocol) 是 Git Manage Service 提供的一种基于 TCP 协议的服务接口，用于与其他系统或工具进行集成。通过 MCP，您可以远程执行 Git 操作、管理同步任务和发送通知等。

### 服务启动

MCP 服务默认在 `cmd/server` 启动时自动运行，端口为 **9000**。

### 对接方式

#### 1. 建立 TCP 连接

```bash
# 使用 nc 命令测试连接
nc localhost 9000
```

#### 2. 发送请求

MCP 使用 JSON 格式的请求和响应：

**请求格式：**
```json
{
  "tool": "工具名称",
  "parameters": {
    "参数1": "值1",
    "参数2": "值2"
  }
}
```

**响应格式：**
```json
{
  "success": true,
  "message": "操作成功",
  "data": "可选的返回数据"
}
```

#### 3. 支持的工具

| 工具名称 | 描述 | 参数 |
|---------|------|------|
| `git_clone` | 克隆仓库 | `remote_url`, `local_path`, `auth_type`, `auth_key`, `auth_secret` |
| `git_fetch` | 获取远程更新 | `path`, `remote` |
| `git_push` | 推送代码 | `path`, `target_remote`, `source_hash`, `target_branch`, `options` |
| `git_checkout` | 切换分支 | `path`, `branch` |
| `git_branches` | 获取分支列表 | `path` |
| `git_add` | 添加文件 | `path`, `files` |
| `git_commit` | 提交更改 | `path`, `message`, `author_name`, `author_email` |
| `git_status` | 获取状态 | `path` |
| `git_log` | 获取提交日志 | `path`, `branch`, `since`, `until` |
| `git_auth` | 验证认证信息 | `auth_type`, `auth_key`, `auth_secret` |
| `notification_send` | 发送通知 | `channel_id`, `event`, `message`, `data` |
| `notification_channels` | 获取通知渠道 | 无 |
| `sync_task` | 创建同步任务 | 无 |
| `sync_run` | 运行同步任务 | 无 |

#### 4. 示例代码

**Go 示例：**
```go
package main

import (
	"encoding/json"
	"fmt"
	"net"
)

type ToolRequest struct {
	Tool       string          `json:"tool"`
	Parameters json.RawMessage `json:"parameters"`
}

type ToolResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func main() {
	// 建立连接
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}
	defer conn.Close()

	// 构建请求
	req := ToolRequest{
		Tool: "git_branches",
		Parameters: json.RawMessage(`{"path": "/path/to/repo"}`),
	}

	// 发送请求
	reqData, _ := json.Marshal(req)
	conn.Write(reqData)

	// 接收响应
	buffer := make([]byte, 4096)
	n, _ := conn.Read(buffer)
	respData := buffer[:n]

	// 解析响应
	var resp ToolResponse
	json.Unmarshal(respData, &resp)

	fmt.Println("响应:", resp)
}
```

**Python 示例：**
```python
import json
import socket

def mcp_request(tool, parameters):
    # 建立连接
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect(("localhost", 9000))
    
    # 构建请求
    request = {
        "tool": tool,
        "parameters": parameters
    }
    
    # 发送请求
    sock.sendall(json.dumps(request).encode('utf-8'))
    
    # 接收响应
    response = sock.recv(4096)
    sock.close()
    
    # 解析响应
    return json.loads(response.decode('utf-8'))

# 示例：获取分支列表
result = mcp_request("git_branches", {"path": "/path/to/repo"})
print(result)
```

## 🛠 技术栈

| 后端 | 前端 |
|------|------|
| Go 1.25 | Vue 3 |
| CloudWeGo Hertz | Element Plus |
| CloudWeGo Kitex | Pinia |
| SQLite / MySQL / PostgreSQL | ECharts |
| Redis (可选) | Monaco Editor |
| MinIO (可选) | TypeScript |

## 📸 界面预览

| 仓库管理 | 分支操作 |
|:---:|:---:|
| ![仓库列表](docs/.vuepress/public/images/repo-list-with-data.png) | ![分支管理](docs/.vuepress/public/images/branch-management.png) |

| 同步任务 | 代码度量 |
|:---:|:---:|
| ![同步任务](docs/.vuepress/public/images/sync-tasks.png) | ![Git 度量](docs/.vuepress/public/images/git-metrics.png) |

更多截图请查看 [文档](docs/README.md)。

## 🔧 开发

### 环境要求

- Go 1.25+
- Node.js 18+
- npm 或 yarn

### 本地开发

```bash
# 克隆仓库
git clone https://github.com/yi-nology/git-manage-service.git
cd git-manage-service

# 安装依赖
go mod tidy
cd frontend && npm install && cd ..

# 开发模式运行
make run

# 或完整构建
make build-full
```

### 构建发布

```bash
# 创建版本标签
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# GitHub Actions 会自动构建并发布
```

## 🤝 参与贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

本项目基于 [MIT](LICENSE) 许可证开源。

## 🔗 相关链接

- **GitHub**: [yi-nology/git-manage-service](https://github.com/yi-nology/git-manage-service)
- **Issues**: [问题反馈](https://github.com/yi-nology/git-manage-service/issues)
- **Releases**: [版本下载](https://github.com/yi-nology/git-manage-service/releases)
- **文档**: [在线文档](https://yi-nology.github.io/git-manage-service/)

---

<p align="center">
  如果这个项目对你有帮助，请给一个 ⭐️ Star 支持一下！
</p>
