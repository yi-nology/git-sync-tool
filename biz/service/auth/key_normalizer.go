package auth

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	ssh2 "golang.org/x/crypto/ssh"
)

// NormalizePrivateKey 解析私钥并重新编码为无密码的标准格式
// 对于 RSA/ECDSA 密钥使用 PKCS8 格式，对于 Ed25519 密钥保持原始 OpenSSH 格式
func NormalizePrivateKey(privateKeyPEM, passphrase string) (string, error) {
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
		// Ed25519 密钥：直接返回原始格式（无密码时）
		// 如果有密码，已经被解析了，直接返回原始 PEM（此时是无密码的）
		if passphrase == "" {
			return privateKeyPEM, nil
		}
		// 对于有密码的 Ed25519 密钥，尝试转换为 PKCS8
		// Go 1.20+ 支持此操作
	}

	// 尝试转换为 PKCS8 格式
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(rawKey)
	if err != nil {
		// 如果 PKCS8 转换失败（例如旧版本 Go 不支持 ed25519），回退到原始格式
		if passphrase == "" {
			return privateKeyPEM, nil
		}
		return "", fmt.Errorf("marshal private key to PKCS8: %w (consider using a passwordless key)", err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}

	return string(pem.EncodeToMemory(pemBlock)), nil
}
