# MCP 服务对接指南

## 什么是 MCP

MCP (Model Context Protocol) 是 Git Manage Service 提供的一种基于 TCP 协议的服务接口，用于与其他系统或工具进行集成。通过 MCP，您可以远程执行 Git 操作、管理同步任务和发送通知等。

## 服务启动

MCP 服务默认在 `cmd/server` 启动时自动运行，端口为 **9000**。

## 对接方式

### 1. 建立 TCP 连接

```bash
# 使用 nc 命令测试连接
nc localhost 9000
```

### 2. 发送请求

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

### 3. 支持的工具

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

### 4. 示例代码

#### Go 示例

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

#### Python 示例

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

#### Node.js 示例

```javascript
const net = require('net');

function mcpRequest(tool, parameters) {
  return new Promise((resolve, reject) => {
    const client = new net.Socket();
    
    client.connect(9000, 'localhost', () => {
      const request = JSON.stringify({
        tool: tool,
        parameters: parameters
      });
      client.write(request);
    });
    
    client.on('data', (data) => {
      const response = JSON.parse(data.toString());
      resolve(response);
      client.destroy();
    });
    
    client.on('error', (error) => {
      reject(error);
      client.destroy();
    });
  });
}

// 示例：获取分支列表
mcpRequest('git_branches', { path: '/path/to/repo' })
  .then(result => console.log(result))
  .catch(error => console.error(error));
```

## 错误处理

当 MCP 服务遇到错误时，会返回 `success: false` 的响应，并在 `message` 字段中包含错误信息。例如：

```json
{
  "success": false,
  "message": "Invalid parameters"
}
```

## 性能建议

1. **连接复用**：对于频繁的操作，建议复用 TCP 连接，而不是每次操作都建立新连接。
2. **请求大小**：单个请求的大小不应超过 4KB，因为服务端的缓冲区大小为 4KB。
3. **并发控制**：避免同时发送过多请求，建议控制并发数在合理范围内。

## 安全考虑

1. **访问控制**：MCP 服务默认监听本地端口 9000，不建议在生产环境中暴露到公网。
2. **认证信息**：通过 MCP 发送的认证信息（如 SSH 密钥）会通过网络传输，请确保网络环境安全。
3. **输入验证**：服务端会对输入参数进行基本验证，但客户端也应该确保发送的参数格式正确。

## 故障排查

### 服务未启动

如果 MCP 服务未启动，可能的原因：

1. 主服务未运行
2. 端口 9000 已被占用
3. 服务启动时出现错误

### 连接失败

如果无法连接到 MCP 服务：

1. 检查服务是否正在运行
2. 检查端口 9000 是否可访问
3. 检查防火墙设置

### 请求失败

如果请求失败：

1. 检查请求格式是否正确
2. 检查参数是否完整
3. 检查目标仓库是否存在且可访问

## 版本兼容性

MCP 服务的接口在不同版本之间可能会有变化，请确保客户端与服务端版本匹配。

## 联系我们

如果在使用 MCP 服务时遇到问题，请通过以下方式联系我们：

- [GitHub Issues](https://github.com/yi-nology/git-manage-service/issues)
- [Discord 社区](https://discord.gg/your-server)
- [邮件支持](mailto:support@example.com)
