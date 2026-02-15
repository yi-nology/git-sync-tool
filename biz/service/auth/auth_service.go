package auth

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	ssh2 "golang.org/x/crypto/ssh"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
)

// AuthService 统一认证解析服务
// 支持本地文件密钥和数据库存储的SSH密钥
type AuthService struct {
	sshKeyDAO *db.SSHKeyDAO
}

// NewAuthService 创建认证服务实例
func NewAuthService() *AuthService {
	return &AuthService{
		sshKeyDAO: db.NewSSHKeyDAO(),
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
	publicKeys.HostKeyCallback = ssh2.InsecureIgnoreHostKey()

	return publicKeys, nil
}

// resolveLocalSSHKey 从本地文件加载SSH密钥
func (s *AuthService) resolveLocalSSHKey(keyPath, passphrase string) (transport.AuthMethod, error) {
	publicKeys, err := ssh.NewPublicKeysFromFile("git", keyPath, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to load SSH key from file %s: %w", keyPath, err)
	}
	publicKeys.HostKeyCallback = ssh2.InsecureIgnoreHostKey()

	return publicKeys, nil
}

// GetDBSSHKeyContent 获取数据库SSH密钥的私钥内容（用于原生git命令）
func (s *AuthService) GetDBSSHKeyContent(sshKeyID uint) (privateKey, passphrase string, err error) {
	sshKey, err := s.sshKeyDAO.FindByID(sshKeyID)
	if err != nil {
		return "", "", fmt.Errorf("failed to load SSH key from database: %w", err)
	}

	if sshKey.PrivateKey == "" {
		return "", "", fmt.Errorf("SSH key %d has no private key content", sshKeyID)
	}

	return sshKey.PrivateKey, sshKey.Passphrase, nil
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
