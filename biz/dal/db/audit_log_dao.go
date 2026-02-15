package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm"
)

type AuditLogDAO struct{}

func NewAuditLogDAO() *AuditLogDAO {
	return &AuditLogDAO{}
}

func (d *AuditLogDAO) Create(log *po.AuditLog) error {
	return DB.Create(log).Error
}

func (d *AuditLogDAO) FindLatest(limit int) ([]po.AuditLog, error) {
	var logs []po.AuditLog
	err := DB.Order("created_at desc").Limit(limit).Find(&logs).Error
	return logs, err
}

func (d *AuditLogDAO) Count() (int64, error) {
	var count int64
	err := DB.Model(&po.AuditLog{}).Count(&count).Error
	return count, err
}

func (d *AuditLogDAO) FindPage(page, pageSize int) ([]po.AuditLog, error) {
	return d.FindPageWithFilters(page, pageSize, "", "", "", "")
}

func (d *AuditLogDAO) FindPageWithFilters(page, pageSize int, action, target, startDate, endDate string) ([]po.AuditLog, error) {
	var logs []po.AuditLog
	offset := (page - 1) * pageSize
	query := d.applyFilters(DB, action, target, startDate, endDate)
	// Exclude 'details' column for list view to improve performance
	err := query.Select("id", "action", "target", "operator", "ip_address", "user_agent", "created_at").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error
	return logs, err
}

func (d *AuditLogDAO) CountWithFilters(action, target, startDate, endDate string) (int64, error) {
	var count int64
	query := d.applyFilters(DB.Model(&po.AuditLog{}), action, target, startDate, endDate)
	err := query.Count(&count).Error
	return count, err
}

func (d *AuditLogDAO) applyFilters(query *gorm.DB, action, target, startDate, endDate string) *gorm.DB {
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if target != "" {
		query = query.Where("target LIKE ?", "%"+target+"%")
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}
	return query
}

func (d *AuditLogDAO) FindByID(id uint) (*po.AuditLog, error) {
	var log po.AuditLog
	err := DB.First(&log, id).Error
	return &log, err
}
