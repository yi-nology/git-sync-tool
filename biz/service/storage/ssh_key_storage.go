package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/pkg/storage"
)

// SSHKeyStorageService SSH密钥存储服务
// 支持双写模式：同时写入数据库和对象存储
type SSHKeyStorageService struct {
	storage   storage.Storage
	sshKeyDAO *db.SSHKeyDAO
	bucket    string
}

// NewSSHKeyStorageService 创建SSH密钥存储服务
func NewSSHKeyStorageService(storageSvc storage.Storage) *SSHKeyStorageService {
	bucket := configs.GlobalConfig.Storage.SSHKeyBucket
	if bucket == "" {
		bucket = "ssh-keys"
	}

	svc := &SSHKeyStorageService{
		storage:   storageSvc,
		sshKeyDAO: db.NewSSHKeyDAO(),
		bucket:    bucket,
	}

	// 确保桶存在
	if storageSvc != nil {
		ctx := context.Background()
		exists, _ := storageSvc.BucketExists(ctx, bucket)
		if !exists {
			if err := storageSvc.MakeBucket(ctx, bucket); err != nil {
				log.Printf("Warning: Failed to create SSH key bucket: %v", err)
			}
		}
	}

	return svc
}

// getStorageKey 获取密钥在对象存储中的键名
func (s *SSHKeyStorageService) getStorageKey(keyID uint) string {
	return fmt.Sprintf("key-%d.pem", keyID)
}

// StoreKeyToStorage 将SSH密钥存储到对象存储
func (s *SSHKeyStorageService) StoreKeyToStorage(ctx context.Context, sshKey *po.SSHKey) error {
	if s.storage == nil {
		return nil // 对象存储未启用，跳过
	}

	// 存储私钥
	key := s.getStorageKey(sshKey.ID)
	content := []byte(sshKey.PrivateKey)
	reader := bytes.NewReader(content)

	if err := s.storage.PutObject(ctx, s.bucket, key, reader, int64(len(content))); err != nil {
		return fmt.Errorf("failed to store SSH key to storage: %w", err)
	}

	return nil
}

// LoadKeyFromStorage 从对象存储加载SSH密钥私钥
func (s *SSHKeyStorageService) LoadKeyFromStorage(ctx context.Context, keyID uint) (string, error) {
	if s.storage == nil {
		return "", fmt.Errorf("object storage not enabled")
	}

	key := s.getStorageKey(keyID)
	reader, err := s.storage.GetObject(ctx, s.bucket, key)
	if err != nil {
		return "", fmt.Errorf("failed to load SSH key from storage: %w", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read SSH key content: %w", err)
	}

	return string(content), nil
}

// DeleteKeyFromStorage 从对象存储删除SSH密钥
func (s *SSHKeyStorageService) DeleteKeyFromStorage(ctx context.Context, keyID uint) error {
	if s.storage == nil {
		return nil // 对象存储未启用，跳过
	}

	key := s.getStorageKey(keyID)
	return s.storage.DeleteObject(ctx, s.bucket, key)
}

// CreateWithStorage 创建SSH密钥并同时存储到对象存储（双写）
func (s *SSHKeyStorageService) CreateWithStorage(ctx context.Context, sshKey *po.SSHKey) error {
	// 先写入数据库
	if err := s.sshKeyDAO.Create(sshKey); err != nil {
		return err
	}

	// 再写入对象存储
	if err := s.StoreKeyToStorage(ctx, sshKey); err != nil {
		log.Printf("Warning: Failed to store SSH key %d to object storage: %v", sshKey.ID, err)
		// 不返回错误，数据库已保存成功
	}

	return nil
}

// DeleteWithStorage 删除SSH密钥并同时从对象存储删除
func (s *SSHKeyStorageService) DeleteWithStorage(ctx context.Context, keyID uint) error {
	// 先从数据库删除
	if err := s.sshKeyDAO.Delete(keyID); err != nil {
		return err
	}

	// 再从对象存储删除
	if err := s.DeleteKeyFromStorage(ctx, keyID); err != nil {
		log.Printf("Warning: Failed to delete SSH key %d from object storage: %v", keyID, err)
		// 不返回错误，数据库已删除成功
	}

	return nil
}

// GetPrivateKey 获取SSH私钥内容
// 优先从对象存储读取，失败则回退到数据库
func (s *SSHKeyStorageService) GetPrivateKey(ctx context.Context, keyID uint) (string, error) {
	// 先尝试从对象存储读取
	if s.storage != nil {
		content, err := s.LoadKeyFromStorage(ctx, keyID)
		if err == nil {
			return content, nil
		}
		log.Printf("Warning: Failed to load SSH key %d from storage, fallback to DB: %v", keyID, err)
	}

	// 回退到数据库
	key, err := s.sshKeyDAO.FindByID(keyID)
	if err != nil {
		return "", fmt.Errorf("failed to find SSH key: %w", err)
	}

	return key.PrivateKey, nil
}
