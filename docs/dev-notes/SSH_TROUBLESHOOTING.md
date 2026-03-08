# SSH 密钥权限问题解决方案

## 问题原因

使用数据库存储的 SSH 密钥克隆仓库时，出现权限错误：
```
git@github.com: Permission denied (publickey).
```

## 解决方案

### 方案 1：使用系统默认 SSH 密钥（推荐）

**步骤 1：检查系统 SSH 密钥**
```bash
# 检查是否有 SSH 密钥
ls -la ~/.ssh/

# 测试 GitHub 连接
ssh -T git@github.com
```

如果显示：`Hi username! You've successfully authenticated` 说明系统密钥正常。

**步骤 2：在克隆仓库时不选择凭证**

在 Web 界面克隆仓库时：
- 不要选择 "数据库 SSH 密钥"
- 留空凭证选择
- 系统会自动使用 `~/.ssh/id_ed25519` 或 `~/.ssh/id_rsa`

### 方案 2：重新添加数据库 SSH 密钥

**步骤 1：获取正确的私钥**
```bash
cat ~/.ssh/id_ed25519
# 或
cat ~/.ssh/id_rsa
```

**步骤 2：在 Web 界面添加 SSH 密钥**
1. 访问：http://localhost:38080/settings
2. 点击 "SSH 密钥管理"
3. 点击 "新增 SSH 密钥"
4. **重要：**
   - 名称：随意（如 "GitHub Key"）
   - 私钥：粘贴完整的私钥内容（包括 `-----BEGIN` 和 `-----END`）
   - 密码：如果你的私钥有密码，必须填写；没有就留空

**步骤 3：验证密钥格式**

正确的私钥格式应该是：
```
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
...（多行 base64 内容）...
-----END OPENSSH PRIVATE KEY-----
```

或传统格式：
```
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
...（多行 base64 内容）...
-----END RSA PRIVATE KEY-----
```

### 方案 3：使用 HTTPS + Personal Access Token（最稳定）

**步骤 1：创建 GitHub Token**
1. 访问：https://github.com/settings/tokens
2. Generate new token (classic)
3. 勾选 `repo` 权限
4. 复制生成的 token

**步骤 2：在 git-manage-service 中使用**
1. 访问：http://localhost:38080/settings
2. 点击 "凭证管理"
3. 新增凭证：
   - 名称：GitHub Token
   - 类型：HTTP Basic Auth
   - 用户名：你的 GitHub 用户名
   - 密码：刚才生成的 Token

**步骤 3：克隆时使用 HTTPS**
- 使用 HTTPS URL：`https://github.com/user/repo.git`
- 选择刚才创建的凭证

## 常见错误

### 1. 私钥格式错误
❌ 错误：缺少 BEGIN/END 标记
❌ 错误：多余的空格或换行
❌ 错误：只复制了部分内容

✅ 正确：完整复制，包括所有行

### 2. 密码填写错误
- 如果私钥有密码：必须填写
- 如果私钥无密码：留空
- 不确定：尝试两种方式

### 3. 权限问题
```bash
# 检查系统密钥权限
ls -la ~/.ssh/
# 应该显示：
# -rw------- (600) id_ed25519
# -rw-r--r-- (644) id_ed25519.pub
```

## 推荐配置

**对于个人使用：**
1. 使用系统默认 SSH 密钥（不配置数据库密钥）
2. 或者使用 HTTPS + Token（最稳定）

**对于团队使用：**
1. 配置数据库 SSH 密钥
2. 确保每个用户都添加了正确的密钥
3. 测试密钥是否可用

## 测试方法

```bash
# 测试系统 SSH
ssh -T git@github.com

# 测试 git 克隆
git clone git@github.com:user/repo.git

# 在 git-manage-service 中测试
# 访问 http://localhost:38080/repos/clone
# 尝试克隆一个公开仓库
```
