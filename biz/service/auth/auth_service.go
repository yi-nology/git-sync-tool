package auth

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	ssh2 "golang.org/x/crypto/ssh"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/service/git"
)

// AuthService 统一认证解析服务
// 支持本地文件密钥、数据库存储的SSH密钥和凭证池
type AuthService struct {
	sshKeyDAO     *db.SSHKeyDAO
	credentialDAO *db.CredentialDAO
}

// NewAuthService 创建认证服务实例
func NewAuthService() *AuthService {
	return &AuthService{
		sshKeyDAO:     db.NewSSHKeyDAO(),
		credentialDAO: db.NewCredentialDAO(),
	}
}

// ResolveAuth 统一解析认证信息
// 支持: local (文件路径), database (数据库密钥), http (用户名密码)
func (s *AuthService) ResolveAuth(authInfo domain.AuthInfo) (transport.AuthMethod, error) {
	if authInfo.Type == "" || authInfo.Type == "none" {
		return nil, nil
	}

	// 处理数据库SSH密钥
	if authInfo.Type == "ssh" && authInfo.Source == "database" && authInfo.SSHKeyID > 0 {
		return s.resolveDBSSHKey(authInfo.SSHKeyID)
	}

	// 处理本地文件SSH密钥
	if authInfo.Type == "ssh" && authInfo.Key != "" {
		return s.resolveLocalSSHKey(authInfo.Key, authInfo.Secret)
	}

	// 处理HTTP认证
	if authInfo.Type == "http" && authInfo.Key != "" {
		return &http.BasicAuth{
			Username: authInfo.Key,
			Password: authInfo.Secret,
		}, nil
	}

	return nil, nil
}

// ResolveAuthFromParams 从基础参数解析（兼容旧接口）
// 当 sshKeyID > 0 时使用数据库密钥，否则使用本地文件路径
func (s *AuthService) ResolveAuthFromParams(authType, authKey, authSecret string, sshKeyID uint) (transport.AuthMethod, error) {
	// 优先使用数据库SSH密钥
	if authType == "ssh" && sshKeyID > 0 {
		return s.resolveDBSSHKey(sshKeyID)
	}

	// 构建AuthInfo并解析
	authInfo := domain.AuthInfo{
		Type:   authType,
		Key:    authKey,
		Secret: authSecret,
		Source: "local",
	}
	return s.ResolveAuth(authInfo)
}

// resolveDBSSHKey 从数据库加载SSH密钥并创建认证方法
func (s *AuthService) resolveDBSSHKey(sshKeyID uint) (transport.AuthMethod, error) {
	// 从数据库加载密钥
	sshKey, err := s.sshKeyDAO.FindByID(sshKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to load SSH key from database: %w", err)
	}

	if sshKey.PrivateKey == "" {
		return nil, fmt.Errorf("SSH key %d has no private key content", sshKeyID)
	}

	// 使用私钥内容创建认证
	publicKeys, err := ssh.NewPublicKeys("git", []byte(sshKey.PrivateKey), sshKey.Passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH private key: %w", err)
	}
	helper := git.NewSSHKeyHelper()
	publicKeys.HostKeyCallback = helper.GetHostKeyCallback()

	return publicKeys, nil
}

// resolveLocalSSHKey 从本地文件加载SSH密钥
func (s *AuthService) resolveLocalSSHKey(keyPath, passphrase string) (transport.AuthMethod, error) {
	publicKeys, err := ssh.NewPublicKeysFromFile("git", keyPath, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to load SSH key from file %s: %w", keyPath, err)
	}
	helper := git.NewSSHKeyHelper()
	publicKeys.HostKeyCallback = helper.GetHostKeyCallback()

	return publicKeys, nil
}

// GetDBSSHKeyContent 获取数据库SSH密钥的私钥内容（用于原生git命令）
// 统一解析并重新编码为无密码的PKCS8 PEM格式，确保ssh命令能正确加载
func (s *AuthService) GetDBSSHKeyContent(sshKeyID uint) (privateKey, passphrase string, err error) {
	sshKey, err := s.sshKeyDAO.FindByID(sshKeyID)
	if err != nil {
		return "", "", fmt.Errorf("failed to load SSH key from database: %w", err)
	}

	if sshKey.PrivateKey == "" {
		return "", "", fmt.Errorf("SSH key %d has no private key content", sshKeyID)
	}

	// 统一解析密钥并重新编码为无密码的标准PEM格式
	// 这样可以避免 passphrase 问题和格式兼容性问题
	decoded, err := normalizePrivateKey(sshKey.PrivateKey, sshKey.Passphrase)
	if err != nil {
		return "", "", fmt.Errorf("failed to normalize private key: %w", err)
	}

	return decoded, "", nil
}

// normalizePrivateKey 解析私钥并重新编码为无密码的标准格式
// 对于 RSA/ECDSA 密钥使用 PKCS8 格式，对于 Ed25519 密钥保持原始 OpenSSH 格式
func normalizePrivateKey(privateKeyPEM, passphrase string) (string, error) {
	keyBytes := []byte(privateKeyPEM)

	var rawKey interface{}
	var err error

	if passphrase != "" {
		rawKey, err = ssh2.ParseRawPrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
	} else {
		rawKey, err = ssh2.ParseRawPrivateKey(keyBytes)
	}
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}

	// 检查是否是 Ed25519 密钥
	if _, ok := rawKey.(*ed25519.PrivateKey); ok {
		// Ed25519 密钥不支持 PKCS8，直接返回原始格式（无密码时）
		// 如果原来有密码，此时已经被解密了，直接返回原始 PEM
		return privateKeyPEM, nil
	}

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(rawKey)
	if err != nil {
		return "", fmt.Errorf("marshal private key to PKCS8: %w", err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}

	return string(pem.EncodeToMemory(pemBlock)), nil
}

// GetAuthInfoForRemote 从仓库配置获取指定远程的认证信息
func GetAuthInfoForRemote(remoteAuths map[string]domain.AuthInfo, remoteName string, defaultAuthType, defaultAuthKey, defaultAuthSecret string) domain.AuthInfo {
	// 优先使用远程特定的认证配置
	if remoteAuths != nil {
		if auth, ok := remoteAuths[remoteName]; ok {
			return auth
		}
	}

	// 回退到默认认证配置
	return domain.AuthInfo{
		Type:   defaultAuthType,
		Key:    defaultAuthKey,
		Secret: defaultAuthSecret,
		Source: "local",
	}
}

// ResolveCredential 从凭证 ID 解析认证方法（用于 go-git）
func (s *AuthService) ResolveCredential(credentialID uint) (transport.AuthMethod, error) {
	if credentialID == 0 {
		return nil, nil
	}

	cred, err := s.credentialDAO.FindByID(credentialID)
	if err != nil {
		return nil, fmt.Errorf("failed to load credential %d: %w", credentialID, err)
	}

	switch cred.Type {
	case "ssh_key":
		if cred.SSHKeyID > 0 {
			return s.resolveDBSSHKey(cred.SSHKeyID)
		}
		if cred.SSHKeyPath != "" {
			return s.resolveLocalSSHKey(cred.SSHKeyPath, cred.Secret)
		}
		return nil, fmt.Errorf("ssh_key credential %d has no key configured", credentialID)
	case "http_basic", "http_token":
		if cred.Username == "" && cred.Secret == "" {
			return nil, nil
		}
		return &http.BasicAuth{
			Username: cred.Username,
			Password: cred.Secret,
		}, nil
	default:
		return nil, fmt.Errorf("unknown credential type: %s", cred.Type)
	}
}

// GetCredentialKeyContent 获取凭证中 SSH 密钥的原始内容（用于原生 git 命令）
// 返回解密后的私钥和空密码（已转换为 PKCS8 无密码格式）
func (s *AuthService) GetCredentialKeyContent(credentialID uint) (privateKey, passphrase string, err error) {
	if credentialID == 0 {
		return "", "", fmt.Errorf("credential ID is 0")
	}

	cred, err := s.credentialDAO.FindByID(credentialID)
	if err != nil {
		return "", "", fmt.Errorf("failed to load credential %d: %w", credentialID, err)
	}

	if cred.Type != "ssh_key" {
		return "", "", fmt.Errorf("credential %d is not SSH type", credentialID)
	}

	if cred.SSHKeyID > 0 {
		return s.GetDBSSHKeyContent(cred.SSHKeyID)
	}

	// 本地密钥：路径 + passphrase（如果有）
	// 对于原生 git 命令，需要使用路径而非内容
	return "", "", fmt.Errorf("local key credentials should use key path directly, not content")
}

// IsCredentialDBKey 判断凭证是否使用数据库 SSH 密钥
func (s *AuthService) IsCredentialDBKey(credentialID uint) bool {
	if credentialID == 0 {
		return false
	}
	cred, err := s.credentialDAO.FindByID(credentialID)
	if err != nil {
		return false
	}
	return cred.Type == "ssh_key" && cred.SSHKeyID > 0
}

// ResolveCredentialForRemote 为指定远程解析凭证
// 优先级：remote_credentials[remoteName] > default_credential_id > 旧 RemoteAuths > 旧默认认证
func (s *AuthService) ResolveCredentialForRemote(
	remoteCredentials map[string]uint,
	defaultCredentialID uint,
	remoteAuths map[string]domain.AuthInfo,
	remoteName string,
	defaultAuthType, defaultAuthKey, defaultAuthSecret string,
) (transport.AuthMethod, bool, error) {
	// 1. 尝试新凭证系统 - 远程专属凭证
	if remoteCredentials != nil {
		if credID, ok := remoteCredentials[remoteName]; ok && credID > 0 {
			auth, err := s.ResolveCredential(credID)
			isDBKey := s.IsCredentialDBKey(credID)
			return auth, isDBKey, err
		}
	}

	// 2. 尝试新凭证系统 - 默认凭证
	if defaultCredentialID > 0 {
		auth, err := s.ResolveCredential(defaultCredentialID)
		isDBKey := s.IsCredentialDBKey(defaultCredentialID)
		return auth, isDBKey, err
	}

	// 3. 回退到旧系统
	authInfo := GetAuthInfoForRemote(remoteAuths, remoteName, defaultAuthType, defaultAuthKey, defaultAuthSecret)
	isDBKey := authInfo.Type == "ssh" && authInfo.Source == "database" && authInfo.SSHKeyID > 0
	auth, err := s.ResolveAuth(authInfo)
	return auth, isDBKey, err
}

// GetCredentialIDForRemote 获取远程对应的凭证 ID
// 返回 0 表示该远程没有配置凭证（使用旧系统或无认证）
func GetCredentialIDForRemote(remoteCredentials map[string]uint, defaultCredentialID uint, remoteName string) uint {
	if remoteCredentials != nil {
		if credID, ok := remoteCredentials[remoteName]; ok && credID > 0 {
			return credID
		}
	}
	return defaultCredentialID
}
