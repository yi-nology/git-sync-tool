package storage

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/pkg/storage"
)

// RepoBackupService 仓库备份服务
type RepoBackupService struct {
	storage   storage.Storage
	backupDAO *db.BackupDAO
	repoDAO   *db.RepoDAO
	bucket    string
}

// NewRepoBackupService 创建仓库备份服务
func NewRepoBackupService(storageSvc storage.Storage) *RepoBackupService {
	bucket := configs.GlobalConfig.Storage.BackupBucket
	if bucket == "" {
		bucket = "backups"
	}

	svc := &RepoBackupService{
		storage:   storageSvc,
		backupDAO: db.NewBackupDAO(),
		repoDAO:   db.NewRepoDAO(),
		bucket:    bucket,
	}

	// 确保桶存在
	if storageSvc != nil {
		ctx := context.Background()
		exists, _ := storageSvc.BucketExists(ctx, bucket)
		if !exists {
			if err := storageSvc.MakeBucket(ctx, bucket); err != nil {
				log.Printf("Warning: Failed to create backup bucket: %v", err)
			}
		}
	}

	return svc
}

// getStorageKey 获取备份文件在对象存储中的键名
func (s *RepoBackupService) getStorageKey(repoKey string, timestamp time.Time) string {
	return fmt.Sprintf("repos/%s/%s.tar.gz", repoKey, timestamp.Format("20060102-150405"))
}

// BackupRepo 备份仓库
func (s *RepoBackupService) BackupRepo(ctx context.Context, repoID uint) (*po.BackupRecord, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("object storage not enabled")
	}

	// 获取仓库信息
	repo, err := s.repoDAO.FindByID(repoID)
	if err != nil {
		return nil, fmt.Errorf("failed to find repo: %w", err)
	}

	// 检查仓库路径是否存在
	if _, err := os.Stat(repo.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("repo path does not exist: %s", repo.Path)
	}

	// 创建备份记录
	now := time.Now()
	record := &po.BackupRecord{
		RepoID:     repoID,
		RepoKey:    repo.Key,
		StorageKey: s.getStorageKey(repo.Key, now),
		Status:     "pending",
		StartedAt:  now,
	}
	if err := s.backupDAO.Create(record); err != nil {
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}

	// 执行备份
	size, err := s.doBackup(ctx, repo.Path, record.StorageKey)
	if err != nil {
		record.Status = "failed"
		record.ErrorMsg = err.Error()
		record.CompletedAt = time.Now()
		s.backupDAO.Update(record)
		return record, err
	}

	// 更新备份记录
	record.Status = "success"
	record.Size = size
	record.CompletedAt = time.Now()
	if err := s.backupDAO.Update(record); err != nil {
		log.Printf("Warning: Failed to update backup record: %v", err)
	}

	return record, nil
}

// doBackup 执行备份操作
func (s *RepoBackupService) doBackup(ctx context.Context, repoPath, storageKey string) (int64, error) {
	// 创建 tar.gz 到内存
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// 遍历目录并添加文件
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过 .git/hooks 目录（可能包含不需要的脚本）
		relPath, _ := filepath.Rel(repoPath, path)
		if relPath == ".git/hooks" && info.IsDir() {
			return filepath.SkipDir
		}

		// 创建 tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// 写入文件内容
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to create tar archive: %w", err)
	}

	// 关闭 tar 和 gzip writer
	if err := tw.Close(); err != nil {
		return 0, fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := gw.Close(); err != nil {
		return 0, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	// 上传到对象存储
	size := int64(buf.Len())
	if err := s.storage.PutObject(ctx, s.bucket, storageKey, &buf, size); err != nil {
		return 0, fmt.Errorf("failed to upload backup: %w", err)
	}

	return size, nil
}

// RestoreRepo 从备份恢复仓库
func (s *RepoBackupService) RestoreRepo(ctx context.Context, backupID uint, targetPath string) error {
	if s.storage == nil {
		return fmt.Errorf("object storage not enabled")
	}

	// 获取备份记录
	record, err := s.backupDAO.FindByID(backupID)
	if err != nil {
		return fmt.Errorf("failed to find backup record: %w", err)
	}

	if record.Status != "success" {
		return fmt.Errorf("backup is not successful, status: %s", record.Status)
	}

	// 下载备份文件
	reader, err := s.storage.GetObject(ctx, s.bucket, record.StorageKey)
	if err != nil {
		return fmt.Errorf("failed to download backup: %w", err)
	}
	defer reader.Close()

	// 确保目标目录存在
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 解压到目标路径
	gr, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		targetFilePath := filepath.Join(targetPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetFilePath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// 确保父目录存在
			if err := os.MkdirAll(filepath.Dir(targetFilePath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			file, err := os.OpenFile(targetFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			if _, err := io.Copy(file, tr); err != nil {
				file.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			file.Close()
		}
	}

	return nil
}

// DeleteBackup 删除备份
func (s *RepoBackupService) DeleteBackup(ctx context.Context, backupID uint) error {
	// 获取备份记录
	record, err := s.backupDAO.FindByID(backupID)
	if err != nil {
		return fmt.Errorf("failed to find backup record: %w", err)
	}

	// 从对象存储删除
	if s.storage != nil && record.StorageKey != "" {
		if err := s.storage.DeleteObject(ctx, s.bucket, record.StorageKey); err != nil {
			log.Printf("Warning: Failed to delete backup from storage: %v", err)
		}
	}

	// 从数据库删除记录
	return s.backupDAO.Delete(backupID)
}

// ListBackups 列出仓库的所有备份
func (s *RepoBackupService) ListBackups(ctx context.Context, repoID uint) ([]po.BackupRecord, error) {
	return s.backupDAO.FindByRepoID(repoID)
}

// GetLatestBackup 获取仓库最新的成功备份
func (s *RepoBackupService) GetLatestBackup(ctx context.Context, repoID uint) (*po.BackupRecord, error) {
	return s.backupDAO.FindLatestByRepoID(repoID)
}
