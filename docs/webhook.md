# Webhook 接口文档

分支管理工具支持通过 Webhook 触发多仓同步。

## 1. 接口信息

- **Endpoint**: `/api/webhooks/task-sync`
- **Method**: `POST`
- **Content-Type**: `application/json`

## 2. 安全验证

### 2.1 签名验证
所有请求必须包含 `X-Hub-Signature-256` 头，值为请求体的 HMAC-SHA256 签名。

算法：`hmac_sha256(secret_key, request_body)`
格式：`sha256=<hex_digest>`

### 2.2 频率限制
- 限制：100 请求/分钟
- 超出返回：429 Too Many Requests

### 2.3 IP 白名单 (可选)
可通过配置环境变量限制允许的 IP 来源。

## 3. 请求参数

### Header
| Key | Value | 说明 |
|---|---|---|
| X-Hub-Signature-256 | sha256=... | 签名 |
| Content-Type | application/json | |

### Body
```json
{
  "task_id": 1
}
```

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| task_id | uint | 是 | 需要触发的多仓同步 ID |

## 4. 响应

### 成功 (200 OK)
```json
{
  "message": "Sync triggered successfully",
  "task_id": 1
}
```

### 错误
- **400 Bad Request**: 请求体格式错误或缺少参数
- **401 Unauthorized**: 签名无效或缺失
- **403 Forbidden**: IP 不在白名单
- **429 Too Many Requests**: 请求过于频繁

## 5. 调用示例

### Python
```python
import hmac
import hashlib
import json
import requests

SECRET = b'my-secret-key'
URL = 'http://localhost:8080/api/webhooks/task-sync'
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
	"fmt"
	"net/http"
)

func main() {
	secret := []byte("my-secret-key")
	url := "http://localhost:8080/api/webhooks/task-sync"
	body := []byte(`{"task_id": 1}`)

	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signature)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println("Status:", resp.Status)
}
```

### Bash (curl)
```bash
# Calculate signature first (e.g. using openssl)
# echo -n '{"task_id": 1}' | openssl dgst -sha256 -hmac "my-secret-key"
# (Output: 7d6...)

curl -X POST http://localhost:8080/api/webhooks/task-sync \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=7d6..." \
  -d '{"task_id": 1}'
```
