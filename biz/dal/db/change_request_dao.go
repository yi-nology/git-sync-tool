package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ChangeRequestDAO struct{}

func NewChangeRequestDAO() *ChangeRequestDAO { return &ChangeRequestDAO{} }

func (d *ChangeRequestDAO) Create(cr *po.ChangeRequest) error {
	return DB.Create(cr).Error
}

func (d *ChangeRequestDAO) FindByID(id uint) (*po.ChangeRequest, error) {
	var cr po.ChangeRequest
	err := DB.First(&cr, id).Error
	return &cr, err
}

func (d *ChangeRequestDAO) FindByRepoAndNumber(repoID uint, crNumber int) (*po.ChangeRequest, error) {
	var cr po.ChangeRequest
	err := DB.Where("repo_id = ? AND cr_number = ?", repoID, crNumber).First(&cr).Error
	return &cr, err
}

func (d *ChangeRequestDAO) FindByRepo(repoID uint, state, sourceBranch, targetBranch string, page, pageSize int) ([]po.ChangeRequest, int64, error) {
	q := DB.Model(&po.ChangeRequest{}).Where("repo_id = ?", repoID)
	if state != "" {
		q = q.Where("state = ?", state)
	}
	if sourceBranch != "" {
		q = q.Where("source_branch = ?", sourceBranch)
	}
	if targetBranch != "" {
		q = q.Where("target_branch = ?", targetBranch)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var crs []po.ChangeRequest
	offset := (page - 1) * pageSize
	err := q.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&crs).Error
	return crs, total, err
}

func (d *ChangeRequestDAO) Save(cr *po.ChangeRequest) error {
	return DB.Save(cr).Error
}

func (d *ChangeRequestDAO) DeleteByRepo(repoID uint) error {
	return DB.Where("repo_id = ?", repoID).Delete(&po.ChangeRequest{}).Error
}
