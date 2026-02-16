package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type RepoDAO struct{}

func NewRepoDAO() *RepoDAO {
	return &RepoDAO{}
}

func (d *RepoDAO) Create(repo *po.Repo) error {
	return DB.Create(repo).Error
}

func (d *RepoDAO) FindAll() ([]po.Repo, error) {
	var repos []po.Repo
	err := DB.Find(&repos).Error
	return repos, err
}

func (d *RepoDAO) FindByKey(key string) (*po.Repo, error) {
	var repo po.Repo
	err := DB.Where("key = ?", key).First(&repo).Error
	return &repo, err
}

func (d *RepoDAO) FindByPath(path string) (*po.Repo, error) {
	var repo po.Repo
	err := DB.Where("path = ?", path).First(&repo).Error
	return &repo, err
}

func (d *RepoDAO) Save(repo *po.Repo) error {
	return DB.Save(repo).Error
}

func (d *RepoDAO) Delete(repo *po.Repo) error {
	return DB.Delete(repo).Error
}

// FindByID 根据ID查询仓库
func (d *RepoDAO) FindByID(id uint) (*po.Repo, error) {
	var repo po.Repo
	err := DB.First(&repo, id).Error
	return &repo, err
}
