# Docker 部署

Docker 部署适合容器化环境，提供了更好的隔离性和可移植性。

## 快速开始

### 使用 Docker Run

```bash
docker run -d \
  --name git-manage-service \
  -p 38080:38080 \
  -v $(pwd)/data:/app/data \
  ghcr.io/yi-nology/git-manage-service:latest
```

### 使用 Docker Compose（推荐）

```bash
# 克隆仓库
git clone https://github.com/yi-nology/git-manage-service.git
cd git-manage-service

# SQLite 方案（最简单）
cd deploy/docker-compose/sqlite
docker-compose up -d
```

## 部署方案

项目提供三种 Docker Compose 配置：

### SQLite 方案（单机）

适合：开发环境、小型团队

```bash
cd deploy/docker-compose/sqlite
docker-compose up -d
```

服务：
- Git Manage Service (38080)

### MySQL 方案（生产）

适合：生产环境、中等规模

```bash
cd deploy/docker-compose/mysql
docker-compose up -d
```

服务：
- Git Manage Service (38080)
- MySQL (3306)
- Redis (6379)
- MinIO (9000, 9001)

### PostgreSQL 方案（生产）

适合：生产环境、PostgreSQL 用户

```bash
cd deploy/docker-compose/postgres
docker-compose up -d
```

服务：
- Git Manage Service (38080)
- PostgreSQL (5432)
- Redis (6379)
- MinIO (9000, 9001)

## 配置说明

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SERVER_PORT` | HTTP 服务端口 | 38080 |
| `RPC_PORT` | RPC 服务端口 | 8888 |
| `DB_TYPE` | 数据库类型 | sqlite |
| `DB_PATH` | SQLite 数据库路径 | data/git_sync.db |
| `DB_HOST` | 数据库主机 | localhost |
| `DB_PORT` | 数据库端口 | 3306 |
| `DB_USER` | 数据库用户 | root |
| `DB_PASSWORD` | 数据库密码 | - |
| `DB_NAME` | 数据库名称 | git_manage |
| `REDIS_ADDR` | Redis 地址 | localhost:6379 |
| `MINIO_ENDPOINT` | MinIO 端点 | localhost:9000 |

### 数据持久化

```yaml
volumes:
  - ./data:/app/data           # 数据目录
  - ./repos:/app/repos         # 仓库目录
  - ./conf:/app/conf           # 配置目录
```

### 端口映射

```yaml
ports:
  - "38080:38080"  # HTTP
  - "8888:8888"    # RPC（可选）
```

## Docker Compose 示例

### SQLite 版本

```yaml
version: '3.8'

services:
  git-manage-service:
    image: ghcr.io/yi-nology/git-manage-service:latest
    container_name: git-manage-service
    restart: unless-stopped
    ports:
      - "38080:38080"
    volumes:
      - ./data:/app/data
      - ./repos:/app/repos
    environment:
      - TZ=Asia/Shanghai
```

### MySQL 版本

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: git-manage-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: your_password
      MYSQL_DATABASE: git_manage
    volumes:
      - mysql_data:/var/lib/mysql
    ports:
      - "3306:3306"

  redis:
    image: redis:7-alpine
    container_name: git-manage-redis
    restart: unless-stopped
    ports:
      - "6379:6379"

  minio:
    image: minio/minio
    container_name: git-manage-minio
    restart: unless-stopped
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    volumes:
      - minio_data:/data
    ports:
      - "9000:9000"
      - "9001:9001"

  git-manage-service:
    image: ghcr.io/yi-nology/git-manage-service:latest
    container_name: git-manage-service
    restart: unless-stopped
    depends_on:
      - mysql
      - redis
      - minio
    ports:
      - "38080:38080"
    volumes:
      - ./repos:/app/repos
    environment:
      - TZ=Asia/Shanghai
      - DB_TYPE=mysql
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_USER=root
      - DB_PASSWORD=your_password
      - DB_NAME=git_manage
      - REDIS_ADDR=redis:6379
      - MINIO_ENDPOINT=minio:9000

volumes:
  mysql_data:
  minio_data:
```

## 常用命令

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 查看日志
docker-compose logs -f git-manage-service

# 重启服务
docker-compose restart git-manage-service

# 进入容器
docker exec -it git-manage-service sh

# 更新镜像
docker-compose pull git-manage-service
docker-compose up -d git-manage-service
```

## 健康检查

```yaml
healthcheck:
  test: ["CMD", "wget", "-q", "--spider", "http://localhost:38080/api/health"]
  interval: 30s
  timeout: 10s
  retries: 3
```

## 资源限制

```yaml
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 2G
    reservations:
      cpus: '0.5'
      memory: 512M
```

## 下一步

- [Kubernetes 部署](/deployment/kubernetes) - 集群部署
- [配置参考](/configuration) - 完整配置说明
