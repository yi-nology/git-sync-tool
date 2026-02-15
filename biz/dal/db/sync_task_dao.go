package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type SyncTaskDAO struct{}

func NewSyncTaskDAO() *SyncTaskDAO {
	return &SyncTaskDAO{}
}

func (d *SyncTaskDAO) Create(task *po.SyncTask) error {
	return DB.Create(task).Error
}

func (d *SyncTaskDAO) FindAllWithRepos() ([]po.SyncTask, error) {
	var tasks []po.SyncTask
	err := DB.Preload("SourceRepo").Preload("TargetRepo").Find(&tasks).Error
	return tasks, err
}

func (d *SyncTaskDAO) FindByRepoKey(repoKey string) ([]po.SyncTask, error) {
	var tasks []po.SyncTask
	err := DB.Preload("SourceRepo").Preload("TargetRepo").
		Where("source_repo_key = ? OR target_repo_key = ?", repoKey, repoKey).
		Find(&tasks).Error
	return tasks, err
}

func (d *SyncTaskDAO) FindByKey(key string) (*po.SyncTask, error) {
	var task po.SyncTask
	err := DB.Preload("SourceRepo").Preload("TargetRepo").
		Where("key = ?", key).First(&task).Error
	return &task, err
}

func (d *SyncTaskDAO) Save(task *po.SyncTask) error {
	return DB.Save(task).Error
}

func (d *SyncTaskDAO) Delete(task *po.SyncTask) error {
	return DB.Delete(task).Error
}

func (d *SyncTaskDAO) CountByRepoKey(repoKey string) (int64, error) {
	var count int64
	err := DB.Model(&po.SyncTask{}).
		Where("source_repo_key = ? OR target_repo_key = ?", repoKey, repoKey).
		Count(&count).Error
	return count, err
}

func (d *SyncTaskDAO) GetKeysByRepoKey(repoKey string) ([]string, error) {
	var taskKeys []string
	err := DB.Model(&po.SyncTask{}).
		Where("source_repo_key = ? OR target_repo_key = ?", repoKey, repoKey).
		Pluck("key", &taskKeys).Error
	return taskKeys, err
}

func (d *SyncTaskDAO) FindEnabledWithCron() ([]po.SyncTask, error) {
	var tasks []po.SyncTask
	err := DB.Where("enabled = ? AND cron != ''", true).Find(&tasks).Error
	return tasks, err
}

func (d *SyncTaskDAO) FindByWebhookToken(token string) (*po.SyncTask, error) {
	var task po.SyncTask
	err := DB.Preload("SourceRepo").Preload("TargetRepo").
		Where("webhook_token = ? AND webhook_token != ''", token).First(&task).Error
	return &task, err
}
