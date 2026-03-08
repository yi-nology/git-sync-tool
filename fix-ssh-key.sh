#!/bin/bash

# SSH 密钥修复脚本
echo "🔧 SSH 密钥修复工具"
echo "===================="
echo ""

# 1. 检查私钥
echo "1️⃣ 检查私钥文件："
if [ -f ~/.ssh/id_rsa ]; then
    echo "   ✅ 找到私钥: ~/.ssh/id_rsa"
    KEY_FILE=~/.ssh/id_rsa
elif [ -f ~/.ssh/id_ed25519 ]; then
    echo "   ✅ 找到私钥: ~/.ssh/id_ed25519"
    KEY_FILE=~/.ssh/id_ed25519
else
    echo "   ❌ 未找到私钥文件"
    exit 1
fi

# 2. 检查私钥是否有密码
echo ""
echo "2️⃣ 检查私钥密码："
if ssh-keygen -y -f $KEY_FILE &>/dev/null; then
    echo "   ✅ 私钥无密码"
    HAS_PASSPHRASE=false
else
    echo "   ⚠️  私钥有密码保护"
    HAS_PASSPHRASE=true
fi

# 3. 生成完整的密钥信息
echo ""
echo "3️⃣ 生成密钥信息（用于添加到数据库）："
echo ""
echo "==================== 复制以下内容 ===================="
echo ""
echo "【公钥】（已自动添加到 GitHub）"
echo "----------------------------------------"
cat ${KEY_FILE}.pub
echo "----------------------------------------"
echo ""

echo "【私钥】（需要添加到数据库）"
echo "----------------------------------------"
cat $KEY_FILE
echo "----------------------------------------"
echo ""

if [ "$HAS_PASSPHRASE" = true ]; then
    echo "【密码】（如果私钥有密码，请填写）"
    echo "----------------------------------------"
    echo "(输入你的私钥密码)"
    echo "----------------------------------------"
    echo ""
fi

echo "==================== 复制结束 ===================="
echo ""

# 4. 提供添加步骤
echo "4️⃣ 添加步骤："
echo ""
echo "方法 1：使用 Web 界面（推荐）"
echo "  1. 访问: http://localhost:38080/settings"
echo "  2. 点击 'SSH 密钥管理'"
echo "  3. 点击 '新增 SSH 密钥'"
echo "  4. 填写信息："
echo "     - 名称: GitHub-RSA-Key"
echo "     - 私钥: 粘贴上面的【私钥】完整内容"
echo "     - 密码: 如果有密码就填写，没有就留空"
echo "  5. 点击 '确定' 保存"
echo ""

echo "方法 2：使用 API（测试）"
echo "  curl -X POST http://localhost:38080/api/v1/system/db-ssh-keys/ \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -d '{"
echo "      \"name\": \"GitHub-RSA-Key\","
echo "      \"private_key\": \"-----BEGIN RSA PRIVATE KEY-----\\n...\\n-----END RSA PRIVATE KEY-----\","
echo "      \"passphrase\": \"\""
echo "    }'"
echo ""

# 5. 验证步骤
echo "5️⃣ 验证密钥："
echo ""
echo "  1. 添加完成后，点击 '测试' 按钮"
echo "  2. 或者克隆一个仓库测试："
echo "     - 访问: http://localhost:38080/repos/clone"
echo "     - URL: git@github.com:yi-nology/git-manage-service.git"
echo "     - 凭证: 选择刚添加的密钥"
echo ""

echo "✅ 准备完成！请按照步骤添加密钥。"
