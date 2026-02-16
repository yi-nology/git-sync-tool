package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/pkg/storage/local"
	"github.com/yi-nology/git-manage-service/pkg/storage/minio"
)

// localStorageAdapter 本地存储适配器
type localStorageAdapter struct {
	impl *local.LocalStorage
}

func (a *localStorageAdapter) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64) error {
	return a.impl.PutObject(ctx, bucket, key, reader, size)
}

func (a *localStorageAdapter) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return a.impl.GetObject(ctx, bucket, key)
}

func (a *localStorageAdapter) DeleteObject(ctx context.Context, bucket, key string) error {
	return a.impl.DeleteObject(ctx, bucket, key)
}

func (a *localStorageAdapter) ListObjects(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error) {
	objs, err := a.impl.ListObjects(ctx, bucket, prefix)
	if err != nil {
		return nil, err
	}
	result := make([]ObjectInfo, len(objs))
	for i, obj := range objs {
		result[i] = ObjectInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
			ContentType:  obj.ContentType,
		}
	}
	return result, nil
}

func (a *localStorageAdapter) StatObject(ctx context.Context, bucket, key string) (*ObjectInfo, error) {
	obj, err := a.impl.StatObject(ctx, bucket, key)
	if err != nil {
		return nil, err
	}
	return &ObjectInfo{
		Key:          obj.Key,
		Size:         obj.Size,
		LastModified: obj.LastModified,
		ContentType:  obj.ContentType,
	}, nil
}

func (a *localStorageAdapter) BucketExists(ctx context.Context, bucket string) (bool, error) {
	return a.impl.BucketExists(ctx, bucket)
}

func (a *localStorageAdapter) MakeBucket(ctx context.Context, bucket string) error {
	return a.impl.MakeBucket(ctx, bucket)
}

func (a *localStorageAdapter) Close() error {
	return a.impl.Close()
}

// minioStorageAdapter MinIO存储适配器
type minioStorageAdapter struct {
	impl *minio.MinIOStorage
}

func (a *minioStorageAdapter) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64) error {
	return a.impl.PutObject(ctx, bucket, key, reader, size)
}

func (a *minioStorageAdapter) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return a.impl.GetObject(ctx, bucket, key)
}

func (a *minioStorageAdapter) DeleteObject(ctx context.Context, bucket, key string) error {
	return a.impl.DeleteObject(ctx, bucket, key)
}

func (a *minioStorageAdapter) ListObjects(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error) {
	objs, err := a.impl.ListObjects(ctx, bucket, prefix)
	if err != nil {
		return nil, err
	}
	result := make([]ObjectInfo, len(objs))
	for i, obj := range objs {
		result[i] = ObjectInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
			ContentType:  obj.ContentType,
		}
	}
	return result, nil
}

func (a *minioStorageAdapter) StatObject(ctx context.Context, bucket, key string) (*ObjectInfo, error) {
	obj, err := a.impl.StatObject(ctx, bucket, key)
	if err != nil {
		return nil, err
	}
	return &ObjectInfo{
		Key:          obj.Key,
		Size:         obj.Size,
		LastModified: obj.LastModified,
		ContentType:  obj.ContentType,
	}, nil
}

func (a *minioStorageAdapter) BucketExists(ctx context.Context, bucket string) (bool, error) {
	return a.impl.BucketExists(ctx, bucket)
}

func (a *minioStorageAdapter) MakeBucket(ctx context.Context, bucket string) error {
	return a.impl.MakeBucket(ctx, bucket)
}

func (a *minioStorageAdapter) Close() error {
	return a.impl.Close()
}

// NewStorage 根据配置创建存储实例
func NewStorage(cfg configs.StorageConfig) (Storage, error) {
	switch cfg.Type {
	case "minio":
		if cfg.Endpoint == "" {
			return nil, fmt.Errorf("minio endpoint is required")
		}
		if cfg.AccessKey == "" {
			return nil, fmt.Errorf("minio access_key is required")
		}
		if cfg.SecretKey == "" {
			return nil, fmt.Errorf("minio secret_key is required")
		}
		impl, err := minio.NewMinIOStorage(cfg.Endpoint, cfg.AccessKey, cfg.SecretKey, cfg.UseSSL)
		if err != nil {
			return nil, err
		}
		return &minioStorageAdapter{impl: impl}, nil
	case "local", "":
		localPath := cfg.LocalPath
		if localPath == "" {
			localPath = "./storage"
		}
		impl, err := local.NewLocalStorage(localPath)
		if err != nil {
			return nil, err
		}
		return &localStorageAdapter{impl: impl}, nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Type)
	}
}
