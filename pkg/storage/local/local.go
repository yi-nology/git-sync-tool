package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ObjectInfo 对象元数据（本地定义，避免循环依赖）
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ContentType  string
}

// LocalStorage 本地文件系统存储实现
type LocalStorage struct {
	basePath string
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// 确保基础路径存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}
	return &LocalStorage{basePath: basePath}, nil
}

// bucketPath 获取桶对应的目录路径
func (l *LocalStorage) bucketPath(bucket string) string {
	return filepath.Join(l.basePath, bucket)
}

// objectPath 获取对象的完整文件路径
func (l *LocalStorage) objectPath(bucket, key string) string {
	return filepath.Join(l.basePath, bucket, key)
}

// PutObject 上传对象到本地文件系统
func (l *LocalStorage) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64) error {
	objPath := l.objectPath(bucket, key)

	// 确保目录存在
	dir := filepath.Dir(objPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入临时文件，然后重命名（保证原子性）
	tmpPath := objPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		file.Close()
		os.Remove(tmpPath) // 清理临时文件
	}()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	// 原子重命名
	if err := os.Rename(tmpPath, objPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// GetObject 从本地文件系统读取对象
func (l *LocalStorage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	objPath := l.objectPath(bucket, key)
	file, err := os.Open(objPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object not found: %s/%s", bucket, key)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// DeleteObject 删除本地文件系统中的对象
func (l *LocalStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	objPath := l.objectPath(bucket, key)
	if err := os.Remove(objPath); err != nil {
		if os.IsNotExist(err) {
			return nil // 对象不存在，视为删除成功
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// ListObjects 列出本地文件系统中指定前缀的所有对象
func (l *LocalStorage) ListObjects(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error) {
	bucketDir := l.bucketPath(bucket)
	var objects []ObjectInfo

	err := filepath.Walk(bucketDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// 获取相对路径作为 key
		relPath, err := filepath.Rel(bucketDir, path)
		if err != nil {
			return err
		}
		// 统一使用斜杠
		relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

		// 检查前缀匹配
		if prefix != "" && !strings.HasPrefix(relPath, prefix) {
			return nil
		}

		objects = append(objects, ObjectInfo{
			Key:          relPath,
			Size:         info.Size(),
			LastModified: info.ModTime(),
		})
		return nil
	})

	if err != nil {
		if os.IsNotExist(err) {
			return []ObjectInfo{}, nil
		}
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	return objects, nil
}

// StatObject 获取本地文件系统中对象的元数据
func (l *LocalStorage) StatObject(ctx context.Context, bucket, key string) (*ObjectInfo, error) {
	objPath := l.objectPath(bucket, key)
	info, err := os.Stat(objPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object not found: %s/%s", bucket, key)
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return &ObjectInfo{
		Key:          key,
		Size:         info.Size(),
		LastModified: info.ModTime(),
	}, nil
}

// BucketExists 检查本地文件系统中桶（目录）是否存在
func (l *LocalStorage) BucketExists(ctx context.Context, bucket string) (bool, error) {
	bucketDir := l.bucketPath(bucket)
	info, err := os.Stat(bucketDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check bucket: %w", err)
	}
	return info.IsDir(), nil
}

// MakeBucket 在本地文件系统中创建桶（目录）
func (l *LocalStorage) MakeBucket(ctx context.Context, bucket string) error {
	bucketDir := l.bucketPath(bucket)
	if err := os.MkdirAll(bucketDir, 0755); err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	return nil
}

// Close 关闭存储（本地存储无需关闭操作）
func (l *LocalStorage) Close() error {
	return nil
}
