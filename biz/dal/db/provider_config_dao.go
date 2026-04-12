package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ProviderConfigDAO struct{}

func NewProviderConfigDAO() *ProviderConfigDAO { return &ProviderConfigDAO{} }

func (d *ProviderConfigDAO) Create(cfg *po.ProviderConfig) error {
	return DB.Create(cfg).Error
}

func (d *ProviderConfigDAO) FindByID(id uint) (*po.ProviderConfig, error) {
	var cfg po.ProviderConfig
	err := DB.First(&cfg, id).Error
	return &cfg, err
}

func (d *ProviderConfigDAO) FindAll() ([]po.ProviderConfig, error) {
	var configs []po.ProviderConfig
	err := DB.Order("updated_at DESC").Find(&configs).Error
	return configs, err
}

func (d *ProviderConfigDAO) FindByPlatform(platform string) ([]po.ProviderConfig, error) {
	var configs []po.ProviderConfig
	err := DB.Where("platform = ?", platform).Find(&configs).Error
	return configs, err
}

func (d *ProviderConfigDAO) Save(cfg *po.ProviderConfig) error {
	return DB.Save(cfg).Error
}

func (d *ProviderConfigDAO) Delete(id uint) error {
	return DB.Delete(&po.ProviderConfig{}, id).Error
}

func (d *ProviderConfigDAO) ExistsByName(name string) (bool, error) {
	var count int64
	err := DB.Model(&po.ProviderConfig{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}
