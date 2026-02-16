package minio

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ObjectInfo 对象元数据（本地定义，避免循环依赖）
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ContentType  string
}

// MinIOStorage MinIO 对象存储实现
type MinIOStorage struct {
	client *minio.Client
}

// NewMinIOStorage 创建 MinIO 存储实例
func NewMinIOStorage(endpoint, accessKey, secretKey string, useSSL bool) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &MinIOStorage{client: client}, nil
}

// PutObject 上传对象到 MinIO
func (m *MinIOStorage) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64) error {
	opts := minio.PutObjectOptions{}
	if size < 0 {
		// 未知大小，使用 -1 让 MinIO 自动处理
		size = -1
	}

	_, err := m.client.PutObject(ctx, bucket, key, reader, size, opts)
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}
	return nil
}

// GetObject 从 MinIO 下载对象
func (m *MinIOStorage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	obj, err := m.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return obj, nil
}

// DeleteObject 删除 MinIO 中的对象
func (m *MinIOStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	err := m.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// ListObjects 列出 MinIO 中指定前缀的所有对象
func (m *MinIOStorage) ListObjects(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error) {
	var objects []ObjectInfo

	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}

	for obj := range m.client.ListObjects(ctx, bucket, opts) {
		if obj.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", obj.Err)
		}
		objects = append(objects, ObjectInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
			ContentType:  obj.ContentType,
		})
	}

	return objects, nil
}

// StatObject 获取 MinIO 中对象的元数据
func (m *MinIOStorage) StatObject(ctx context.Context, bucket, key string) (*ObjectInfo, error) {
	info, err := m.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}

	return &ObjectInfo{
		Key:          info.Key,
		Size:         info.Size,
		LastModified: info.LastModified,
		ContentType:  info.ContentType,
	}, nil
}

// BucketExists 检查 MinIO 中桶是否存在
func (m *MinIOStorage) BucketExists(ctx context.Context, bucket string) (bool, error) {
	exists, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket: %w", err)
	}
	return exists, nil
}

// MakeBucket 在 MinIO 中创建桶
func (m *MinIOStorage) MakeBucket(ctx context.Context, bucket string) error {
	err := m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	if err != nil {
		// 检查桶是否已存在
		exists, errExists := m.client.BucketExists(ctx, bucket)
		if errExists == nil && exists {
			return nil // 桶已存在，视为成功
		}
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	return nil
}

// Close 关闭 MinIO 连接（MinIO 客户端无需显式关闭）
func (m *MinIOStorage) Close() error {
	return nil
}
