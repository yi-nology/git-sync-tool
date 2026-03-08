package git

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	ssh2 "golang.org/x/crypto/ssh"
)

// SSHKeyHelper 提供 SSH 密钥处理的辅助功能
type SSHKeyHelper struct {
	// 主机密钥存储
	hostKeys map[string]ssh2.PublicKey
}

// NewSSHKeyHelper 创建 SSHKeyHelper 实例
func NewSSHKeyHelper() *SSHKeyHelper {
	return &SSHKeyHelper{
		hostKeys: make(map[string]ssh2.PublicKey),
	}
}

// ProcessPrivateKey 处理私钥内容，支持带密码的密钥
func (h *SSHKeyHelper) ProcessPrivateKey(privateKey, passphrase string) (string, error) {
	keyContent := privateKey
	if !strings.HasSuffix(keyContent, "\n") {
		keyContent += "\n"
	}

	// 如果有 passphrase，需要解密私钥
	if passphrase != "" {
		// 解析加密的私钥
		rawKey, err := ssh2.ParseRawPrivateKeyWithPassphrase([]byte(keyContent), []byte(passphrase))
		if err != nil {
			return "", fmt.Errorf("failed to parse encrypted private key: %v", err)
		}

		// 重新编码为无密码的 PEM 格式
		pemBytes, err := x509.MarshalPKCS8PrivateKey(rawKey)
		if err != nil {
			return "", fmt.Errorf("failed to marshal private key: %v", err)
		}

		pemBlock := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: pemBytes,
		}
		keyContent = string(pem.EncodeToMemory(pemBlock))
	}

	return keyContent, nil
}

// CreateTempKeyFile 创建临时密钥文件
func (h *SSHKeyHelper) CreateTempKeyFile(keyContent string) (string, error) {
	// 创建临时私钥文件
	tmpFile, err := os.CreateTemp("", "git_ssh_key_*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp key file: %v", err)
	}

	// 写入私钥内容
	if _, err := tmpFile.WriteString(keyContent); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write key file: %v", err)
	}
	tmpFile.Close()

	// 设置文件权限为 600
	if err := os.Chmod(tmpFile.Name(), 0600); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to set key file permissions: %v", err)
	}

	return tmpFile.Name(), nil
}

// BuildSSHCommand 构建 SSH 命令
func (h *SSHKeyHelper) BuildSSHCommand(keyPath string) string {
	return fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null", keyPath)
}

// CleanupTempFile 清理临时文件
func (h *SSHKeyHelper) CleanupTempFile(filePath string) {
	if filePath != "" {
		os.Remove(filePath)
	}
}

// AddHostKey 添加已知的主机密钥
func (h *SSHKeyHelper) AddHostKey(host string, key ssh2.PublicKey) {
	h.hostKeys[host] = key
}

// GetHostKeyCallback 获取主机密钥回调函数
func (h *SSHKeyHelper) GetHostKeyCallback() ssh2.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh2.PublicKey) error {
		// 检查是否是已知主机
		if knownKey, ok := h.hostKeys[hostname]; ok {
			if bytes.Equal(ssh2.Marshal(key), ssh2.Marshal(knownKey)) {
				return nil
			}
			return fmt.Errorf("host key mismatch for %s", hostname)
		}

		// 对于新主机，可以选择接受并存储密钥
		// 这里为了兼容性，暂时接受新主机，但记录警告
		log.Printf("Warning: Accepting new host key for %s", hostname)
		h.hostKeys[hostname] = key
		return nil
	}
}

// BuildSecureSSHCommand 构建安全的 SSH 命令，包含主机密钥验证
func (h *SSHKeyHelper) BuildSecureSSHCommand(keyPath string) string {
	// 这里使用 StrictHostKeyChecking=ask，会在遇到新主机时提示
	// 在生产环境中，应该使用 StrictHostKeyChecking=yes 并预先配置 known_hosts
	return fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=ask -o UserKnownHostsFile=~/.ssh/known_hosts", keyPath)
}
