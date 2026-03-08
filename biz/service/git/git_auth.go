package git

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/yi-nology/git-manage-service/biz/model/domain"
	conf "github.com/yi-nology/git-manage-service/pkg/configs"
)

func (s *GitService) getAuth(authType, authKey, authSecret string) (transport.AuthMethod, error) {
	if authType == "http" && authKey != "" {
		return &http.BasicAuth{
			Username: authKey,
			Password: authSecret,
		}, nil
	} else if authType == "ssh" && authKey != "" {
		publicKeys, err := ssh.NewPublicKeysFromFile("git", authKey, "")
		if err != nil {
			return nil, err
		}
		helper := NewSSHKeyHelper()
		publicKeys.HostKeyCallback = helper.GetHostKeyCallback()
		return publicKeys, nil
	}
	return nil, nil
}

// getAuthFromInfo 从AuthInfo结构获取认证方法，支持本地密钥和数据库密钥
func (s *GitService) getAuthFromInfo(authInfo domain.AuthInfo) (transport.AuthMethod, error) {
	if authInfo.Type == "ssh" {
		if authInfo.Source == "database" && authInfo.SSHKeyID > 0 {
			// 从数据库加载密钥 - 需要在调用方提供私钥内容
			return nil, fmt.Errorf("database key loading should be handled by caller with GetAuthFromDBKey")
		}
		// Source == "local" 或为空，使用文件路径
		if authInfo.Key != "" {
			publicKeys, err := ssh.NewPublicKeysFromFile("git", authInfo.Key, authInfo.Secret)
			if err != nil {
				return nil, err
			}
			helper := NewSSHKeyHelper()
			publicKeys.HostKeyCallback = helper.GetHostKeyCallback()
			return publicKeys, nil
		}
	} else if authInfo.Type == "http" && authInfo.Key != "" {
		return &http.BasicAuth{
			Username: authInfo.Key,
			Password: authInfo.Secret,
		}, nil
	}
	return nil, nil
}

// GetAuthFromDBKey 从数据库密钥内容创建认证方法
func (s *GitService) GetAuthFromDBKey(privateKey, passphrase string) (transport.AuthMethod, error) {
	publicKeys, err := ssh.NewPublicKeys("git", []byte(privateKey), passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}
	helper := NewSSHKeyHelper()
	publicKeys.HostKeyCallback = helper.GetHostKeyCallback()

	// 配置更广泛的 SSH 算法支持以提高兼容性
	publicKeys.HostKeyCallbackHelper = ssh.HostKeyCallbackHelper{
		HostKeyCallback: helper.GetHostKeyCallback(),
	}

	return publicKeys, nil
}

// TestRemoteConnectionWithDBKey 使用数据库密钥测试远程连接
func (s *GitService) TestRemoteConnectionWithDBKey(url, privateKey, passphrase string) error {
	// 优先使用原生 git 命令（更可靠）
	err := s.testConnectionWithGitCommand(url, privateKey, passphrase)
	if err == nil {
		return nil
	}

	// 原生命令失败，尝试使用 go-git
	auth, err := s.GetAuthFromDBKey(privateKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to prepare auth: %v", err)
	}

	ep, err := transport.NewEndpoint(url)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	// 使用 memory storage 初始化
	storer := memory.NewStorage()
	r, err := git.Init(storer, nil)
	if err != nil {
		return fmt.Errorf("failed to init memory repo: %v", err)
	}

	remote, err := r.CreateRemote(&config.RemoteConfig{
		Name: "test",
		URLs: []string{ep.String()},
	})
	if err != nil {
		return fmt.Errorf("failed to create remote: %v", err)
	}

	_, err = remote.List(&git.ListOptions{
		Auth: auth,
	})
	if err != nil {
		return fmt.Errorf("connection failed: %v (git command also failed: %v)", err, err)
	}

	return nil
}

// testConnectionWithGitCommand 使用原生 git 命令测试连接（更可靠）
func (s *GitService) testConnectionWithGitCommand(url, privateKey, passphrase string) error {
	helper := NewSSHKeyHelper()

	// 处理私钥内容
	keyContent, err := helper.ProcessPrivateKey(privateKey, passphrase)
	if err != nil {
		return err
	}

	// 创建临时密钥文件
	tmpFile, err := helper.CreateTempKeyFile(keyContent)
	if err != nil {
		return err
	}
	defer helper.CleanupTempFile(tmpFile)

	// 构建 GIT_SSH_COMMAND
	sshCmd := helper.BuildSSHCommand(tmpFile)

	// 执行 git ls-remote
	cmd := exec.Command("git", "ls-remote", "--heads", url)
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git ls-remote failed: %v, output: %s", err, string(output))
	}

	return nil
}

// TestRemoteConnectionWithLocalKey 使用本地SSH密钥文件测试远程连接
func (s *GitService) TestRemoteConnectionWithLocalKey(url, keyPath, passphrase string) error {
	publicKeys, err := ssh.NewPublicKeysFromFile("git", keyPath, passphrase)
	if err != nil {
		return fmt.Errorf("failed to load SSH key from file %s: %v", keyPath, err)
	}
	helper := NewSSHKeyHelper()
	publicKeys.HostKeyCallback = helper.GetHostKeyCallback()

	remote := git.NewRemote(nil, &config.RemoteConfig{
		Name: "test",
		URLs: []string{url},
	})
	_, err = remote.List(&git.ListOptions{Auth: publicKeys})
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	return nil
}

// TestRemoteConnectionWithHTTP 使用HTTP认证测试远程连接
func (s *GitService) TestRemoteConnectionWithHTTP(url, username, password string) error {
	auth := &http.BasicAuth{
		Username: username,
		Password: password,
	}

	remote := git.NewRemote(nil, &config.RemoteConfig{
		Name: "test",
		URLs: []string{url},
	})
	_, err := remote.List(&git.ListOptions{Auth: auth})
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	return nil
}

func (s *GitService) detectSSHAuth(urlStr string) transport.AuthMethod {
	// Simple check for SSH
	// git@... or ssh://...
	if !strings.HasPrefix(urlStr, "git@") && !strings.HasPrefix(urlStr, "ssh://") && !strings.Contains(urlStr, "git") {
		// Try parsing endpoint to be sure
		ep, err := transport.NewEndpoint(urlStr)
		if err != nil || ep.Protocol != "ssh" {
			return nil
		}
	}

	user := "git"
	ep, err := transport.NewEndpoint(urlStr)
	if err == nil && ep.User != "" {
		user = ep.User
	}

	if conf.DebugMode {
		log.Printf("[DEBUG] detectSSHAuth for %s (user: %s)", urlStr, user)
	}

	// 1. Try common key paths first (if they are unencrypted)
	home, err := os.UserHomeDir()
	if err == nil {
		keyPaths := []string{
			filepath.Join(home, ".ssh", "id_rsa"),
			filepath.Join(home, ".ssh", "id_ed25519"),
			filepath.Join(home, ".ssh", "id_ecdsa"),
		}

		for _, path := range keyPaths {
			if _, err := os.Stat(path); err == nil {
				// Try to load with empty password
				auth, err := ssh.NewPublicKeysFromFile(user, path, "")
				if err == nil {
					helper := NewSSHKeyHelper()
					auth.HostKeyCallback = helper.GetHostKeyCallback()
					if conf.DebugMode {
						log.Printf("[DEBUG] Using SSH Key: %s", path)
					}
					return auth
				} else if conf.DebugMode {
					log.Printf("[DEBUG] Failed to load key %s (maybe encrypted?): %v", path, err)
				}
			}
		}
	}

	// 2. Try SSH Agent
	if auth, err := ssh.NewSSHAgentAuth(user); err == nil {
		helper := NewSSHKeyHelper()
		auth.HostKeyCallback = helper.GetHostKeyCallback()
		if conf.DebugMode {
			log.Printf("[DEBUG] Using SSH Agent Auth")
		}
		return auth
	}

	if conf.DebugMode {
		log.Printf("[DEBUG] No SSH auth found")
	}
	return nil
}
