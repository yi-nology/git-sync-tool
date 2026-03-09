# 快速开始

本指南将帮助你在 5 分钟内完成 Git Manage Service 的安装和基本配置。

![文档截图](/git-manage-service/images/docs/docs-getting-started.png)

## 1. 下载安装

### 方式一：下载预编译二进制（推荐）

从 [Releases](https://github.com/yi-nology/git-manage-service/releases) 页面下载适合你系统的版本：

| 平台 | 架构 | 文件名 |
|------|------|--------|
| Linux | AMD64 | `git-manage-service-linux-amd64.tar.gz` |
| Linux | ARM64 | `git-manage-service-linux-arm64.tar.gz` |
| macOS | Intel | `git-manage-service-darwin-amd64.tar.gz` |
| macOS | Apple Silicon | `git-manage-service-darwin-arm64.tar.gz` |
| Windows | AMD64 | `git-manage-service-windows-amd64.exe.zip` |
| Windows | ARM64 | `git-manage-service-windows-arm64.exe.zip` |

### 方式二：Docker

```bash
docker pull ghcr.io/yi-nology/git-manage-service:latest
```

### 方式三：从源码编译

```bash
git clone https://github.com/yi-nology/git-manage-service.git
cd git-manage-service
make build-full
```

## 2. 启动服务

### Linux / macOS

```bash
# 解压
tar -xzf git-manage-service-*.tar.gz

# 添加执行权限
chmod +x git-manage-service-*

# 运行
./git-manage-service-*
```

### Windows

```powershell
# 解压 zip 文件
# 双击运行或在命令行中执行
.\git-manage-service-windows-amd64.exe
```

### Docker

```bash
docker run -d \
  --name git-manage-service \
  -p 38080:38080 \
  -v ./data:/app/data \
  ghcr.io/yi-nology/git-manage-service:latest
```

## 3. 访问界面

浏览器打开: [http://localhost:38080](http://localhost:38080)

![首页](/git-manage-service/images/homepage.png)

## 4. 基本配置

### 4.1 添加仓库

1. 点击左侧导航 **"仓库管理"**
2. 点击 **"注册仓库"** 按钮
3. 输入仓库 **名称** 和 **本地路径**
4. 点击保存

![注册仓库](/git-manage-service/images/repo-register.png)

### 4.2 配置 SSH 密钥（如需访问私有仓库）

1. 点击左侧导航 **"系统设置"**
2. 进入 **"SSH 密钥"** 标签
3. 点击 **"新增密钥"** 按钮
4. 粘贴你的 SSH 私钥内容
5. 保存后，在仓库配置中选择该密钥

![SSH 密钥管理](/git-manage-service/images/ssh-keys.png)

### 4.3 创建同步任务

1. 点击左侧导航 **"同步任务"**
2. 点击 **"新建任务"** 按钮
3. 配置同步规则：
   - **源仓库**: 选择已注册的仓库
   - **源 Remote/分支**: 如 `origin` / `main`
   - **目标 Remote/分支**: 如 `backup` / `main`
   - **Cron 表达式**: 如 `0 */2 * * *`（每 2 小时同步）
4. 保存并启用

![新建同步任务](/git-manage-service/images/sync-task-create.png)

### 4.4 配置通知（可选）

1. 点击左侧导航 **"系统设置"**
2. 进入 **"通知渠道"** 标签
3. 添加通知渠道（钉钉/企微/飞书等）
4. 配置触发事件和消息模板

![通知渠道配置](/git-manage-service/images/notification-channel.png)

## 5. 验证功能

### 手动执行同步

在同步任务列表中，点击 **"运行"** 按钮手动触发一次同步。

### 查看执行日志

1. 点击左侧导航 **"审计日志"**
2. 筛选操作类型为 **"sync"**
3. 查看详细的执行记录

![审计日志](/git-manage-service/images/audit-log.png)

## 下一步

- 📘 [功能指南](/features/repo) - 了解所有功能的详细用法
- 📦 [部署方案](/deployment/binary) - 生产环境部署指南
- ⚙️ [配置参考](/configuration) - 完整的配置项说明
- 🔌 [API 文档](/api) - HTTP API 接口参考

## 常见问题

### 端口被占用？

修改配置文件 `conf/config.yaml`：

```yaml
server:
  port: 8080  # 改为其他端口
```

### 无法访问私有仓库？

确保：
1. SSH 密钥已正确添加
2. 仓库配置中选择了正确的密钥
3. 密钥有访问目标仓库的权限

### 同步失败？

查看审计日志中的错误信息，常见原因：
- 网络问题
- 权限不足
- 分支冲突
- Remote 配置错误
