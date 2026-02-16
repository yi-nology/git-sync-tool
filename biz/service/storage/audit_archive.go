package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/pkg/storage"
)

// AuditArchiveService 审计日志归档服务
type AuditArchiveService struct {
	storage     storage.Storage
	auditLogDAO *db.AuditLogDAO
	bucket      string
}

// NewAuditArchiveService 创建审计日志归档服务
func NewAuditArchiveService(storageSvc storage.Storage) *AuditArchiveService {
	bucket := configs.GlobalConfig.Storage.AuditLogBucket
	if bucket == "" {
		bucket = "audit-logs"
	}

	svc := &AuditArchiveService{
		storage:     storageSvc,
		auditLogDAO: db.NewAuditLogDAO(),
		bucket:      bucket,
	}

	// 确保桶存在
	if storageSvc != nil {
		ctx := context.Background()
		exists, _ := storageSvc.BucketExists(ctx, bucket)
		if !exists {
			if err := storageSvc.MakeBucket(ctx, bucket); err != nil {
				log.Printf("Warning: Failed to create audit log bucket: %v", err)
			}
		}
	}

	return svc
}

// AuditLogArchive 归档记录
type AuditLogArchive struct {
	ArchiveID  string        `json:"archive_id"`
	StartDate  string        `json:"start_date"`
	EndDate    string        `json:"end_date"`
	TotalCount int           `json:"total_count"`
	ArchivedAt time.Time     `json:"archived_at"`
	Logs       []po.AuditLog `json:"logs"`
}

// getStorageKey 获取归档文件在对象存储中的键名
func (s *AuditArchiveService) getStorageKey(startDate, endDate time.Time) string {
	return fmt.Sprintf("archives/%s_to_%s.json",
		startDate.Format("20060102"),
		endDate.Format("20060102"))
}

// ArchiveLogs 归档指定日期范围内的审计日志
func (s *AuditArchiveService) ArchiveLogs(ctx context.Context, startDate, endDate time.Time) error {
	if s.storage == nil {
		return fmt.Errorf("object storage not enabled")
	}

	// 查询指定日期范围内的审计日志
	logs, err := s.auditLogDAO.FindByDateRange(startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to query audit logs: %w", err)
	}

	if len(logs) == 0 {
		log.Printf("No audit logs found for date range %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		return nil
	}

	// 创建归档记录
	archive := AuditLogArchive{
		ArchiveID:  fmt.Sprintf("archive_%s_%s", startDate.Format("20060102"), endDate.Format("20060102")),
		StartDate:  startDate.Format("2006-01-02"),
		EndDate:    endDate.Format("2006-01-02"),
		TotalCount: len(logs),
		ArchivedAt: time.Now(),
		Logs:       logs,
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(archive, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal archive: %w", err)
	}

	// 上传到对象存储
	storageKey := s.getStorageKey(startDate, endDate)
	if err := s.storage.PutObject(ctx, s.bucket, storageKey, bytes.NewReader(data), int64(len(data))); err != nil {
		return fmt.Errorf("failed to upload archive: %w", err)
	}

	log.Printf("Archived %d audit logs to %s/%s", len(logs), s.bucket, storageKey)
	return nil
}

// ArchiveOldLogs 归档并删除超过指定天数的审计日志
func (s *AuditArchiveService) ArchiveOldLogs(ctx context.Context, daysOld int) error {
	if s.storage == nil {
		return fmt.Errorf("object storage not enabled")
	}

	endDate := time.Now().AddDate(0, 0, -daysOld)
	startDate := endDate.AddDate(0, -1, 0) // 归档前一个月的数据

	// 先归档
	if err := s.ArchiveLogs(ctx, startDate, endDate); err != nil {
		return err
	}

	// 删除已归档的日志（可选，需要用户确认）
	// 这里只是标记，实际删除需要谨慎操作
	log.Printf("Audit logs older than %d days have been archived. Consider cleaning up database records.", daysOld)
	return nil
}

// ListArchives 列出所有归档文件
func (s *AuditArchiveService) ListArchives(ctx context.Context) ([]storage.ObjectInfo, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("object storage not enabled")
	}

	return s.storage.ListObjects(ctx, s.bucket, "archives/")
}

// GetArchive 获取指定的归档内容
func (s *AuditArchiveService) GetArchive(ctx context.Context, archiveKey string) (*AuditLogArchive, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("object storage not enabled")
	}

	reader, err := s.storage.GetObject(ctx, s.bucket, archiveKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get archive: %w", err)
	}
	defer reader.Close()

	var archive AuditLogArchive
	if err := json.NewDecoder(reader).Decode(&archive); err != nil {
		return nil, fmt.Errorf("failed to decode archive: %w", err)
	}

	return &archive, nil
}

// DeleteArchive 删除归档文件
func (s *AuditArchiveService) DeleteArchive(ctx context.Context, archiveKey string) error {
	if s.storage == nil {
		return fmt.Errorf("object storage not enabled")
	}

	return s.storage.DeleteObject(ctx, s.bucket, archiveKey)
}
