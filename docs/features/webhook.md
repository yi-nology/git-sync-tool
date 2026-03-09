# Webhook 集成

Webhook 功能允许外部系统（如 CI/CD 平台）通过 HTTP 请求触发 Git Manage Service 的同步任务。

## 接口信息

| 项目 | 值 |
|------|------|
| Endpoint | `/api/webhooks/task-sync` |
| Method | `POST` |
| Content-Type | `application/json` |

## 安全验证

### 签名验证

所有请求必须包含 `X-Hub-Signature-256` 请求头：

```
X-Hub-Signature-256: sha256=<hex_digest>
```

签名算法：
```python
signature = hmac_sha256(secret_key, request_body)
```

### 频率限制

- 默认限制：100 请求/分钟
- 超出限制返回：`429 Too Many Requests`

### IP 白名单（可选）

在配置文件中设置允许的 IP 列表：

```yaml
webhook:
  ip_whitelist:
    - "192.168.1.0/24"
    - "10.0.0.1"
```

## 请求格式

### 请求头

| Header | 值 | 说明 |
|--------|------|------|
| Content-Type | application/json | 必需 |
| X-Hub-Signature-256 | sha256=... | 签名验证 |

### 请求体

```json
{
  "task_id": 1
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| task_id | uint | 是 | 要触发的同步任务 ID |

## 响应格式

### 成功响应 (200 OK)

```json
{
  "message": "Sync triggered successfully",
  "task_id": 1
}
```

### 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 请求体格式错误或缺少参数 |
| 401 | 签名无效或缺失 |
| 403 | IP 不在白名单中 |
| 404 | 任务不存在 |
| 429 | 请求过于频繁 |

## 调用示例

### Python

```python
import hmac
import hashlib
import json
import requests

SECRET = b'my-secret-key'
URL = 'http://localhost:38080/api/webhooks/task-sync'
DATA = {'task_id': 1}

body = json.dumps(DATA).encode('utf-8')
signature = 'sha256=' + hmac.new(SECRET, body, hashlib.sha256).hexdigest()

headers = {
    'Content-Type': 'application/json',
    'X-Hub-Signature-256': signature
}

response = requests.post(URL, data=body, headers=headers)
print(response.status_code, response.text)
```

### Go

```go
package main

import (
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "net/http"
)

func main() {
    secret := []byte("my-secret-key")
    url := "http://localhost:38080/api/webhooks/task-sync"
    body := []byte(`{"task_id": 1}`)

    mac := hmac.New(sha256.New, secret)
    mac.Write(body)
    signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Hub-Signature-256", signature)

    client := &http.Client{}
    resp, _ := client.Do(req)
    defer resp.Body.Close()
}
```

### cURL

```bash
# 先计算签名
# echo -n '{"task_id": 1}' | openssl dgst -sha256 -hmac "my-secret-key"

curl -X POST http://localhost:38080/api/webhooks/task-sync \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=<calculated_signature>" \
  -d '{"task_id": 1}'
```

### GitHub Actions

```yaml
name: Trigger Sync

on:
  push:
    branches: [main]

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger Git Manage Service
        run: |
          body='{"task_id": 1}'
          signature=$(echo -n "$body" | openssl dgst -sha256 -hmac "${{ secrets.WEBHOOK_SECRET }}" | sed 's/.*= //')
          
          curl -X POST ${{ secrets.GMS_URL }}/api/webhooks/task-sync \
            -H "Content-Type: application/json" \
            -H "X-Hub-Signature-256: sha256=$signature" \
            -d "$body"
```

## 配置说明

### 配置文件

```yaml
webhook:
  secret: my-secret-key      # 签名密钥
  rate_limit: 100            # 频率限制（请求/分钟）
  ip_whitelist: []           # IP 白名单
```

### 获取任务 ID

任务 ID 可以在同步任务列表中查看，或通过 API 获取。

## 最佳实践

1. **密钥管理**: 使用环境变量或密钥管理服务存储 Webhook Secret
2. **错误处理**: 实现重试逻辑，处理临时网络故障
3. **日志记录**: 记录所有 Webhook 调用，便于排查问题
4. **超时设置**: 设置合理的请求超时时间

## 下一步

- [同步任务](/features/sync) - 创建同步任务
- [通知配置](/features/notification) - 同步结果通知
- [API 文档](/api) - 完整的 API 参考
