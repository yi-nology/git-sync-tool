# 二进制部署

这是最简单直接的部署方式，适合单机环境。

![文档截图](/git-manage-service/images/docs/docs-deployment.png)

## 下载安装

### 从 Releases 下载

从 [GitHub Releases](https://github.com/yi-nology/git-manage-service/releases) 下载适合你系统的版本：

| 平台 | 架构 | 文件名 |
|------|------|--------|
| Linux | AMD64 | `git-manage-service-linux-amd64.tar.gz` |
| Linux | ARM64 | `git-manage-service-linux-arm64.tar.gz` |
| macOS | Intel | `git-manage-service-darwin-amd64.tar.gz` |
| macOS | Apple Silicon | `git-manage-service-darwin-arm64.tar.gz` |
| Windows | AMD64 | `git-manage-service-windows-amd64.exe.zip` |
| Windows | ARM64 | `git-manage-service-windows-arm64.exe.zip` |

### 下载命令

```bash
# Linux AMD64
wget https://github.com/yi-nology/git-manage-service/releases/download/v0.7.2/git-manage-service-linux-amd64.tar.gz

# macOS Apple Silicon
wget https://github.com/yi-nology/git-manage-service/releases/download/v0.7.2/git-manage-service-darwin-arm64.tar.gz
```

## 安装步骤

### Linux / macOS

```bash
# 1. 解压
tar -xzf git-manage-service-*.tar.gz

# 2. 添加执行权限
chmod +x git-manage-service-*

# 3. 移动到系统路径（可选）
sudo mv git-manage-service-* /usr/local/bin/git-manage-service
```

### Windows

```powershell
# 1. 解压 zip 文件
# 2. 将 exe 文件移动到合适的位置
```

## 运行服务

### 基本运行

```bash
# 默认模式（HTTP + RPC）
./git-manage-service

# 仅 HTTP 服务
./git-manage-service --mode=http

# 仅 RPC 服务
./git-manage-service --mode=rpc
```

### 指定配置文件

```bash
./git-manage-service --config /path/to/config.yaml
```

### 查看版本

```bash
./git-manage-service --version
```

## 配置

### 配置文件位置

默认配置文件：`./conf/config.yaml`

如果配置文件不存在，系统会使用默认配置并自动创建。

### 主要配置项

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
```

完整配置参考：[配置参考](/configuration)

## 后台运行

### 使用 nohup

```bash
nohup ./git-manage-service > output.log 2>&1 &
```

### 使用 systemd（推荐）

创建服务文件 `/etc/systemd/system/git-manage-service.service`：

```ini
[Unit]
Description=Git Manage Service
After=network.target

[Service]
Type=simple
User=git
WorkingDirectory=/opt/git-manage-service
ExecStart=/opt/git-manage-service/git-manage-service
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

启用服务：

```bash
# 重载配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start git-manage-service

# 开机自启
sudo systemctl enable git-manage-service

# 查看状态
sudo systemctl status git-manage-service
```

### 使用 supervisor

创建配置文件 `/etc/supervisor/conf.d/git-manage-service.conf`：

```ini
[program:git-manage-service]
directory=/opt/git-manage-service
command=/opt/git-manage-service/git-manage-service
autostart=true
autorestart=true
startsecs=3
stderr_logfile=/var/log/git-manage-service.err.log
stdout_logfile=/var/log/git-manage-service.out.log
user=git
```

## 数据目录

默认数据目录结构：

```
./
├── conf/
│   └── config.yaml      # 配置文件
├── data/
│   ├── git_sync.db      # SQLite 数据库
│   └── repos/           # 仓库目录
└── logs/
    └── app.log          # 应用日志
```

## 反向代理

### Nginx 配置

```nginx
server {
    listen 80;
    server_name git-manage.example.com;

    location / {
        proxy_pass http://127.0.0.1:38080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Caddy 配置

```
git-manage.example.com {
    reverse_proxy localhost:38080
}
```

## 升级

```bash
# 1. 停止服务
sudo systemctl stop git-manage-service

# 2. 备份数据
cp -r data data.bak

# 3. 下载新版本
wget https://github.com/yi-nology/git-manage-service/releases/download/vX.X.X/git-manage-service-*.tar.gz

# 4. 解压并替换
tar -xzf git-manage-service-*.tar.gz
mv git-manage-service-* /usr/local/bin/git-manage-service

# 5. 启动服务
sudo systemctl start git-manage-service
```

## 下一步

- [Docker 部署](/deployment/docker) - 使用容器部署
- [Kubernetes 部署](/deployment/kubernetes) - 集群部署
- [配置参考](/configuration) - 完整配置说明
