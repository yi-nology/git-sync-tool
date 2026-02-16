package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

// BackupDAO 备份记录数据访问对象
type BackupDAO struct{}

// NewBackupDAO 创建备份DAO
func NewBackupDAO() *BackupDAO {
	return &BackupDAO{}
}

// Create 创建备份记录
func (d *BackupDAO) Create(record *po.BackupRecord) error {
	return DB.Create(record).Error
}

// FindByID 根据ID查询备份记录
func (d *BackupDAO) FindByID(id uint) (*po.BackupRecord, error) {
	var record po.BackupRecord
	err := DB.First(&record, id).Error
	return &record, err
}

// FindByRepoID 根据仓库ID查询备份记录
func (d *BackupDAO) FindByRepoID(repoID uint) ([]po.BackupRecord, error) {
	var records []po.BackupRecord
	err := DB.Where("repo_id = ?", repoID).Order("created_at DESC").Find(&records).Error
	return records, err
}

// FindByRepoKey 根据仓库Key查询备份记录
func (d *BackupDAO) FindByRepoKey(repoKey string) ([]po.BackupRecord, error) {
	var records []po.BackupRecord
	err := DB.Where("repo_key = ?", repoKey).Order("created_at DESC").Find(&records).Error
	return records, err
}

// FindLatestByRepoID 查询仓库最新的成功备份
func (d *BackupDAO) FindLatestByRepoID(repoID uint) (*po.BackupRecord, error) {
	var record po.BackupRecord
	err := DB.Where("repo_id = ? AND status = ?", repoID, "success").
		Order("created_at DESC").
		First(&record).Error
	return &record, err
}

// Update 更新备份记录
func (d *BackupDAO) Update(record *po.BackupRecord) error {
	return DB.Save(record).Error
}

// Delete 删除备份记录
func (d *BackupDAO) Delete(id uint) error {
	return DB.Delete(&po.BackupRecord{}, id).Error
}

// DeleteByRepoID 删除仓库的所有备份记录
func (d *BackupDAO) DeleteByRepoID(repoID uint) error {
	return DB.Where("repo_id = ?", repoID).Delete(&po.BackupRecord{}).Error
}

// CountByRepoID 统计仓库的备份数量
func (d *BackupDAO) CountByRepoID(repoID uint) (int64, error) {
	var count int64
	err := DB.Model(&po.BackupRecord{}).Where("repo_id = ?", repoID).Count(&count).Error
	return count, err
}

// FindPage 分页查询备份记录
func (d *BackupDAO) FindPage(page, pageSize int) ([]po.BackupRecord, int64, error) {
	var records []po.BackupRecord
	var total int64

	offset := (page - 1) * pageSize

	if err := DB.Model(&po.BackupRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := DB.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	return records, total, err
}
