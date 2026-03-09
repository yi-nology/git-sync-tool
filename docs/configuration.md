# 配置参考

本文档详细介绍 Git Manage Service 的所有配置项。

## 配置文件

默认配置文件路径：`./conf/config.yaml`

配置文件不存在时，系统会使用默认配置并自动创建。

## 完整配置示例

```yaml
# 服务器配置
server:
  port: 38080

# RPC 服务配置
rpc:
  port: 8888

# 数据库配置
database:
  type: sqlite
  path: data/git_sync.db

# Webhook 配置
webhook:
  secret: my-secret-key
  rate_limit: 100
  ip_whitelist: []

# 存储配置
storage:
  type: local
  local_path: ./data

# 分布式锁配置
lock:
  type: memory

# 日志配置
log:
  level: info
  format: json

# 代码质量检查配置
lint:
  enable_rpmlint: false
```

## 配置项详解

### 服务器配置 (server)

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `port` | int | 38080 | HTTP 服务端口 |

### RPC 配置 (rpc)

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `port` | int | 8888 | RPC 服务端口 |

### 数据库配置 (database)

#### SQLite

```yaml
database:
  type: sqlite
  path: data/git_sync.db
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `type` | string | sqlite | 数据库类型 |
| `path` | string | data/git_sync.db | 数据库文件路径 |

#### MySQL

```yaml
database:
  type: mysql
  host: localhost
  port: 3306
  user: root
  password: your_password
  dbname: git_manage_service
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `type` | string | - | 数据库类型：mysql |
| `host` | string | localhost | 数据库主机 |
| `port` | int | 3306 | 数据库端口 |
| `user` | string | root | 数据库用户 |
| `password` | string | - | 数据库密码 |
| `dbname` | string | git_manage_service | 数据库名称 |

#### PostgreSQL

```yaml
database:
  type: postgres
  host: localhost
  port: 5432
  user: postgres
  password: your_password
  dbname: git_manage_service
  sslmode: disable
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `type` | string | - | 数据库类型：postgres |
| `host` | string | localhost | 数据库主机 |
| `port` | int | 5432 | 数据库端口 |
| `user` | string | postgres | 数据库用户 |
| `password` | string | - | 数据库密码 |
| `dbname` | string | git_manage_service | 数据库名称 |
| `sslmode` | string | disable | SSL 模式 |

### Webhook 配置 (webhook)

```yaml
webhook:
  secret: my-secret-key
  rate_limit: 100
  ip_whitelist:
    - "192.168.1.0/24"
    - "10.0.0.1"
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `secret` | string | my-secret-key | Webhook 签名密钥 |
| `rate_limit` | int | 100 | 频率限制（请求/分钟） |
| `ip_whitelist` | []string | [] | IP 白名单 |

### 存储配置 (storage)

#### 本地存储

```yaml
storage:
  type: local
  local_path: ./data
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `type` | string | local | 存储类型 |
| `local_path` | string | ./data | 本地存储路径 |

#### MinIO

```yaml
storage:
  type: minio
  endpoint: localhost:9000
  access_key: minioadmin
  secret_key: minioadmin
  use_ssl: false
  repo_bucket: git-repos
  ssh_key_bucket: ssh-keys
  audit_log_bucket: audit-logs
  backup_bucket: backups
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `type` | string | - | 存储类型：minio |
| `endpoint` | string | localhost:9000 | MinIO 端点 |
| `access_key` | string | minioadmin | Access Key |
| `secret_key` | string | minioadmin | Secret Key |
| `use_ssl` | bool | false | 是否使用 SSL |
| `repo_bucket` | string | git-repos | 仓库存储桶 |
| `ssh_key_bucket` | string | ssh-keys | SSH 密钥存储桶 |
| `audit_log_bucket` | string | audit-logs | 审计日志存储桶 |
| `backup_bucket` | string | backups | 备份存储桶 |

### 分布式锁配置 (lock)

#### 内存锁

```yaml
lock:
  type: memory
```

#### Redis 锁

```yaml
lock:
  type: redis
  redis_addr: localhost:6379
  redis_password: ""
  redis_db: 0
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `type` | string | memory | 锁类型：memory/redis |
| `redis_addr` | string | localhost:6379 | Redis 地址 |
| `redis_password` | string | "" | Redis 密码 |
| `redis_db` | int | 0 | Redis 数据库 |

### 日志配置 (log)

```yaml
log:
  level: info
  format: json
  output: stdout
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `level` | string | info | 日志级别：debug/info/warn/error |
| `format` | string | json | 日志格式：json/text |
| `output` | string | stdout | 输出目标：stdout/file |

### 代码质量检查 (lint)

```yaml
lint:
  enable_rpmlint: false
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `enable_rpmlint` | bool | false | 启用 RPM Lint |

## 环境变量

部分配置可以通过环境变量覆盖：

| 环境变量 | 对应配置 | 说明 |
|----------|----------|------|
| `SERVER_PORT` | server.port | HTTP 端口 |
| `RPC_PORT` | rpc.port | RPC 端口 |
| `DB_TYPE` | database.type | 数据库类型 |
| `DB_PATH` | database.path | SQLite 路径 |
| `DB_HOST` | database.host | 数据库主机 |
| `DB_PORT` | database.port | 数据库端口 |
| `DB_USER` | database.user | 数据库用户 |
| `DB_PASSWORD` | database.password | 数据库密码 |
| `DB_NAME` | database.dbname | 数据库名称 |
| `WEBHOOK_SECRET` | webhook.secret | Webhook 密钥 |
| `REDIS_ADDR` | lock.redis_addr | Redis 地址 |
| `MINIO_ENDPOINT` | storage.endpoint | MinIO 端点 |

## 配置优先级

1. 环境变量（最高优先级）
2. 配置文件
3. 默认值（最低优先级）

## 下一步

- [部署方案](/deployment/binary) - 部署到生产环境
- [API 文档](/api) - HTTP API 参考
