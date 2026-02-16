package storage

import (
	"context"
	"io"
	"time"
)

// ObjectInfo 对象元数据
type ObjectInfo struct {
	Key          string    // 对象键名
	Size         int64     // 对象大小（字节）
	LastModified time.Time // 最后修改时间
	ContentType  string    // 内容类型
}

// Storage 对象存储接口
type Storage interface {
	// PutObject 上传对象
	PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64) error

	// GetObject 下载对象，调用者需要关闭返回的 ReadCloser
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)

	// DeleteObject 删除对象
	DeleteObject(ctx context.Context, bucket, key string) error

	// ListObjects 列出指定前缀的所有对象
	ListObjects(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error)

	// StatObject 获取对象元数据
	StatObject(ctx context.Context, bucket, key string) (*ObjectInfo, error)

	// BucketExists 检查桶是否存在
	BucketExists(ctx context.Context, bucket string) (bool, error)

	// MakeBucket 创建桶
	MakeBucket(ctx context.Context, bucket string) error

	// Close 关闭存储连接
	Close() error
}
