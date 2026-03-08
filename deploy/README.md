# 部署指南

本指南涵盖了 Git Manage Service 的 Docker Compose 和 Kubernetes 部署流程。

## 目录结构

```
deploy/
├── config.yaml              # 应用主配置文件
├── CONFIG_GUIDE.md          # 配置文件详细说明
├── .env.example             # 环境变量示例文件
├── docker-compose/          # Docker Compose 部署配置
│   ├── nginx.conf          # Nginx 反向代理配置
│   ├── mysql/              # MySQL + Redis + MinIO + Nginx
│   │   └── docker-compose.yml
│   ├── postgres/           # PostgreSQL + Redis + MinIO + Nginx
│   │   └── docker-compose.yml
│   └── sqlite/             # SQLite 轻量级部署（无 Nginx）
│       └── docker-compose.yml
└── k8s/                     # Kubernetes 资源清单
    ├── configmap.yaml
    ├── secret.yaml
    ├── mysql.yaml
    ├── deployment.yaml
    └── service.yaml
```

---

## 架构说明

### 🏗️ 部署架构

#### MySQL / PostgreSQL 部署（推荐生产环境）
```
用户请求 (80)
    ↓
┌─────────────┐
│   Nginx     │  ← 反向代理 + 前端静态资源服务
└─────────────┘
    ↓ (API)          ↓ (静态文件)
┌─────────────┐      从 public/ 目录加载
│  后端服务    │  ← Go 服务（仅提供 API）
│  (8080)     │
└─────────────┘
    ↓                ↓                ↓
┌──────────┐  ┌──────────┐  ┌──────────┐
│  MySQL/  │  │  Redis   │  │  MinIO   │
│ Postgres │  │  (缓存)  │  │  (存储)  │
└──────────┘  └──────────┘  └──────────┘
```

**特点**：
- ✅ Nginx 直接提供前端静态资源（性能最优）
- ✅ Go 服务只处理 API 请求（降低负载）
- ✅ 前后端完全分离架构

#### SQLite 部署（适合开发/小型环境）
```
用户请求 (8080)
    ↓
┌─────────────┐
│  后端服务    │  ← Go 服务（API + 前端静态资源）
│  (8080)     │     从 ./public 目录加载
└─────────────┘
    ↓
┌──────────┐
│  SQLite  │  ← 本地文件数据库
└──────────┘
```

**特点**：
- ✅ 单容器部署，简单快捷
- ✅ Go 服务提供完整功能（API + 前端）
- ✅ 适合开发测试和小规模部署

### 🎯 关键特性

**MySQL/PostgreSQL 模式（生产推荐）**：
- **Nginx 直接服务前端**：从 `public/` 目录提供静态资源，性能最优
- **Go 服务专注 API**：只处理业务逻辑，降低服务器负载
- **完全分离架构**：前端由 Nginx 提供，后端专注 API
- **分布式支持**：Redis 分布式锁 + MinIO 对象存储

**SQLite 模式（开发/小型环境）**：
- **一体化服务**：Go 服务同时提供 API 和前端静态资源
- **从 ./public 加载**：后端直接从 public 目录提供前端文件
- **单容器部署**：简化架构，易于开发测试
- **本地存储**：SQLite 数据库 + 本地文件存储

**构建流程**：
1. 前端构建：`npm run build` → `frontend/dist/`
2. 集成到后端：复制到 `public/` 目录
3. Docker 构建：多阶段构建，自动集成前后端

---

## 1. Docker Compose 部署

### 1.1 MySQL 部署（推荐）

**特点**：包含 Nginx、MySQL、Redis、MinIO 完整技术栈
**架构**：Nginx 直接提供前端，Go 服务专注 API

**前置条件**：
```bash
# 确保已构建前端资源到 public/ 目录
cd ../../..
make build-frontend-integrate
# 或者
cd frontend && npm install && npm run build && cd .. && cp -r frontend/dist public
```

**启动服务**：
```bash
# 1. 进入 MySQL 部署目录
cd deploy/docker-compose/mysql

# 2. 启动所有服务
docker-compose up -d

# 3. 查看服务状态
docker-compose ps

# 4. 查看日志
docker-compose logs -f app
```

**访问地址**：
- 前端页面：http://localhost （由 Nginx 提供）
- API 接口：http://localhost/api/v1 （代理到 Go 服务）

- MinIO 控制台：http://localhost:9001

### 1.2 PostgreSQL 部署

**特点**：与 MySQL 类似，但使用 PostgreSQL 数据库
**架构**：Nginx 直接提供前端，Go 服务专注 API

**前置条件**：同 MySQL 部署，需先构建前端

```bash
# 进入 PostgreSQL 部署目录
cd deploy/docker-compose/postgres

# 启动服务
docker-compose up -d
```

### 1.3 SQLite 部署（轻量级）

**特点**：单容器部署，适合开发测试
**架构**：Go 服务同时提供 API 和前端（从 ./public 加载）

**前置条件**：同样需要构建前端（Docker 构建时会自动包含）

```bash
# 进入 SQLite 部署目录
cd deploy/docker-compose/sqlite

# 启动服务
docker-compose up -d
```

**访问地址**：
- 前端页面：http://localhost:8080 （由 Go 服务提供）
- API 接口：http://localhost:8080/api/v1

**注意**：SQLite 模式不包含 Nginx，Go 服务直接暴露

---

### 🔧 环境变量配置

可以通过 `.env` 文件或环境变量覆盖默认配置：

| 变量名 | 默认值 | 说明 |
| :--- | :--- | :--- |
| `WEBHOOK_SECRET` | `my-secret-key` | Webhook 签名密钥 |
| `DB_TYPE` | `mysql/postgres/sqlite` | 数据库类型 |
| `DB_HOST` | - | 数据库主机地址 |
| `DB_PORT` | - | 数据库端口 |
| `DB_USER` | - | 数据库用户名 |
| `DB_PASSWORD` | - | 数据库密码 |
| `DB_NAME` | - | 数据库名称 |
| `STORAGE_TYPE` | `local/minio` | 存储类型 |
| `LOCK_TYPE` | `memory/redis` | 分布式锁类型 |

---

### 🛠️ 常用命令

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 查看日志
docker-compose logs -f app
docker-compose logs -f nginx

# 重启服务
docker-compose restart app

# 重新构建并启动
docker-compose up -d --build

# 清理所有数据（危险操作）
docker-compose down -v
```

---

## 2. Kubernetes 集群部署

适用于生产环境的高可用部署。详细说明请查看 [k8s/README.md](k8s/README.md)。

### 🚀 一键部署（推荐）

使用自动化脚本快速部署：

```bash
cd deploy/k8s

# 1. 构建前端资源
cd ../../
make build-frontend-integrate
cd deploy/k8s

# 2. 一键部署
./deploy.sh deploy

# 3. 查看状态
./deploy.sh status
```

**脚本功能**：
- ✅ 自动检查前置条件
- ✅ 自动上传前端资源到 PVC
- ✅ 自动部署所有组件
- ✅ 支持重启、卸载、查看日志等操作

### 📋 手动部署步骤

**架构**：
- Nginx Pod（2 副本）：提供前端静态资源
- Backend Pod（2 副本）：Go API 服务
- MySQL/PostgreSQL：数据库
- Ingress：统一入口（可选）

**前置条件**：
```bash
# 1. 构建前端资源
make build-frontend-integrate

# 2. 准备 Docker 镜像
docker build -t git-manage-service:latest .
# 推送到你的 Registry
docker tag git-manage-service:latest your-registry/git-manage-service:latest
docker push your-registry/git-manage-service:latest
```

**部署步骤**：
```bash
cd deploy/k8s

# 1. 创建 Secret 和 ConfigMap
kubectl apply -f secret.yaml
kubectl apply -f configmap.yaml
kubectl apply -f nginx-configmap.yaml

# 2. 部署数据库（可选）
kubectl apply -f mysql.yaml

# 3. 上传前端资源到 PVC
# 创建临时 Pod
kubectl create -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: frontend-uploader
spec:
  containers:
  - name: uploader
    image: busybox
    command: ['sh', '-c', 'sleep 3600']
    volumeMounts:
    - name: frontend
      mountPath: /data
  volumes:
  - name: frontend
    persistentVolumeClaim:
      claimName: git-manage-frontend-pvc
EOF

# 等待就绪并复制前端资源
kubectl wait --for=condition=Ready pod/frontend-uploader
kubectl cp ../../public/. frontend-uploader:/data/
kubectl delete pod frontend-uploader

# 4. 部署应用
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f nginx-deployment.yaml

# 5. 部署 Ingress（可选）
kubectl apply -f ingress.yaml

# 6. 查看状态
kubectl get all -l app=git-manage
```

**访问方式**：
```bash
# 方式一：通过 LoadBalancer
kubectl get svc git-manage-nginx
# 访问：http://<EXTERNAL-IP>

# 方式二：通过端口转发（测试）
kubectl port-forward svc/git-manage-nginx 8080:80
# 访问：http://localhost:8080

# 方式三：通过 Ingress
# 访问：http://git-manage.example.com
```

详细的部署说明、故障排查和生产环境配置，请参考 [k8s/README.md](k8s/README.md)。

---

## 3. 多环境支持

- **开发环境**：直接使用 `docker-compose.yml`，配合 `DB_TYPE=sqlite` 可快速启动。
- **生产环境**：
  - 建议使用 Kubernetes 部署。
  - 将 `config.yaml` 中的 `debug` 设为 `false`。
  - 数据库密码等敏感信息**必须**通过环境变量或 Secret 注入，不要写在 `config.yaml` 明文中。
