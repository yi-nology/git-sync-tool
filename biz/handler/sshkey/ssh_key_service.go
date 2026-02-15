package sshkey

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/biz/sshkey"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
	"golang.org/x/crypto/ssh"
)

// ListDBSSHKeys 列出数据库中的SSH密钥
// @router /api/v1/system/db-ssh-keys [GET]
func ListDBSSHKeys(ctx context.Context, c *app.RequestContext) {
	dao := db.NewSSHKeyDAO()
	keys, err := dao.FindAll()
	if err != nil {
		response.InternalServerError(c, "Failed to fetch SSH keys: "+err.Error())
		return
	}

	result := make([]*sshkey.DBSSHKey, 0, len(keys))
	for _, key := range keys {
		result = append(result, toDBSSHKeyDTO(&key))
	}

	response.Success(c, result)
}

// CreateDBSSHKey 创建SSH密钥
// @router /api/v1/system/db-ssh-keys [POST]
func CreateDBSSHKey(ctx context.Context, c *app.RequestContext) {
	var req sshkey.CreateDBSSHKeyRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}
	if req.PrivateKey == "" {
		response.BadRequest(c, "private_key is required")
		return
	}

	// 检查名称是否已存在
	dao := db.NewSSHKeyDAO()
	exists, err := dao.ExistsByName(req.Name)
	if err != nil {
		response.InternalServerError(c, "Failed to check key name: "+err.Error())
		return
	}
	if exists {
		response.BadRequest(c, "SSH key with this name already exists")
		return
	}

	// 验证私钥格式并提取信息
	keyType, publicKey, err := parsePrivateKey(req.PrivateKey, req.Passphrase)
	if err != nil {
		response.BadRequest(c, "Invalid private key: "+err.Error())
		return
	}

	// 如果用户没有提供公钥，使用从私钥提取的公钥
	if req.PublicKey == "" {
		req.PublicKey = publicKey
	}

	sshKeyPO := &po.SSHKey{
		Name:        req.Name,
		Description: req.Description,
		PrivateKey:  req.PrivateKey,
		PublicKey:   req.PublicKey,
		Passphrase:  req.Passphrase,
		KeyType:     keyType,
		UserID:      0, // 预留用户ID
	}

	if err := dao.Create(sshKeyPO); err != nil {
		response.InternalServerError(c, "Failed to create SSH key: "+err.Error())
		return
	}

	response.Success(c, toDBSSHKeyDTO(sshKeyPO))
}

// GetDBSSHKey 获取SSH密钥详情
// @router /api/v1/system/db-ssh-keys/:id [GET]
func GetDBSSHKey(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	dao := db.NewSSHKeyDAO()
	key, err := dao.FindByID(uint(id))
	if err != nil {
		response.NotFound(c, "SSH key not found")
		return
	}

	response.Success(c, toDBSSHKeyDTO(key))
}

// UpdateDBSSHKey 更新SSH密钥
// @router /api/v1/system/db-ssh-keys/:id [PUT]
func UpdateDBSSHKey(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	var req sshkey.UpdateDBSSHKeyRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	dao := db.NewSSHKeyDAO()
	key, err := dao.FindByID(uint(id))
	if err != nil {
		response.NotFound(c, "SSH key not found")
		return
	}

	// 更新描述
	if req.Description != "" {
		key.Description = req.Description
	}

	// 更新私钥（如果提供）
	if req.PrivateKey != "" {
		keyType, publicKey, err := parsePrivateKey(req.PrivateKey, req.Passphrase)
		if err != nil {
			response.BadRequest(c, "Invalid private key: "+err.Error())
			return
		}
		key.PrivateKey = req.PrivateKey
		key.PublicKey = publicKey
		key.KeyType = keyType
		key.Passphrase = req.Passphrase
	} else if req.Passphrase != "" {
		// 只更新密码短语，验证现有私钥是否可用
		_, _, err := parsePrivateKey(key.PrivateKey, req.Passphrase)
		if err != nil {
			response.BadRequest(c, "Passphrase does not match the private key")
			return
		}
		key.Passphrase = req.Passphrase
	}

	if err := dao.Update(key); err != nil {
		response.InternalServerError(c, "Failed to update SSH key: "+err.Error())
		return
	}

	response.Success(c, toDBSSHKeyDTO(key))
}

// DeleteDBSSHKey 删除SSH密钥
// @router /api/v1/system/db-ssh-keys/:id [DELETE]
func DeleteDBSSHKey(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	dao := db.NewSSHKeyDAO()
	if _, err := dao.FindByID(uint(id)); err != nil {
		response.NotFound(c, "SSH key not found")
		return
	}

	if err := dao.Delete(uint(id)); err != nil {
		response.InternalServerError(c, "Failed to delete SSH key: "+err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "SSH key deleted successfully"})
}

// TestDBSSHKey 测试SSH密钥连接
// @router /api/v1/system/db-ssh-keys/:id/test [POST]
func TestDBSSHKey(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	var req sshkey.TestDBSSHKeyRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Url == "" {
		response.BadRequest(c, "url is required")
		return
	}

	dao := db.NewSSHKeyDAO()
	key, err := dao.FindByID(uint(id))
	if err != nil {
		response.NotFound(c, "SSH key not found")
		return
	}

	// 测试连接
	gitSvc := git.NewGitService()
	err = gitSvc.TestRemoteConnectionWithDBKey(req.Url, key.PrivateKey, key.Passphrase)
	if err != nil {
		response.Success(c, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	response.Success(c, map[string]interface{}{
		"success": true,
		"message": "Connection successful",
	})
}

// parsePrivateKey 解析私钥，返回密钥类型和公钥
func parsePrivateKey(privateKeyPEM string, passphrase string) (keyType string, publicKey string, err error) {
	privateKeyBytes := []byte(privateKeyPEM)
	var signer ssh.Signer

	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKeyBytes, []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(privateKeyBytes)
	}

	if err != nil {
		return "", "", fmt.Errorf("failed to parse private key: %v", err)
	}

	// 获取公钥
	pubKey := signer.PublicKey()
	publicKey = string(ssh.MarshalAuthorizedKey(pubKey))

	// 获取密钥类型
	keyType = detectKeyType(pubKey.Type())

	return keyType, strings.TrimSpace(publicKey), nil
}

// detectKeyType 检测密钥类型
func detectKeyType(sshKeyType string) string {
	switch sshKeyType {
	case "ssh-rsa":
		return "rsa"
	case "ssh-ed25519":
		return "ed25519"
	case "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521":
		return "ecdsa"
	case "ssh-dss":
		return "dsa"
	default:
		return sshKeyType
	}
}

// getPublicKeyFingerprint 获取公钥指纹（未使用但保留）
func getPublicKeyFingerprint(publicKey string) string {
	parts := strings.Fields(publicKey)
	if len(parts) < 2 {
		return ""
	}

	keyData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(keyData)
	return "SHA256:" + base64.StdEncoding.EncodeToString(hash[:])[:43]
}

// toDBSSHKeyDTO 转换为DTO（不包含私钥）
func toDBSSHKeyDTO(key *po.SSHKey) *sshkey.DBSSHKey {
	return &sshkey.DBSSHKey{
		Id:            uint64(key.ID),
		Name:          key.Name,
		Description:   key.Description,
		PublicKey:     key.PublicKey,
		KeyType:       key.KeyType,
		HasPassphrase: key.HasPassphraseSet(),
		CreatedAt:     key.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     key.UpdatedAt.Format(time.RFC3339),
	}
}
