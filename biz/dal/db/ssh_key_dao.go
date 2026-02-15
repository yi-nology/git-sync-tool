package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type SSHKeyDAO struct{}

func NewSSHKeyDAO() *SSHKeyDAO {
	return &SSHKeyDAO{}
}

// Create 创建SSH密钥
func (d *SSHKeyDAO) Create(sshKey *po.SSHKey) error {
	return DB.Create(sshKey).Error
}

// FindAll 查询所有SSH密钥
func (d *SSHKeyDAO) FindAll() ([]po.SSHKey, error) {
	var keys []po.SSHKey
	err := DB.Find(&keys).Error
	return keys, err
}

// FindByID 根据ID查询SSH密钥
func (d *SSHKeyDAO) FindByID(id uint) (*po.SSHKey, error) {
	var key po.SSHKey
	err := DB.First(&key, id).Error
	return &key, err
}

// FindByName 根据名称查询SSH密钥
func (d *SSHKeyDAO) FindByName(name string) (*po.SSHKey, error) {
	var key po.SSHKey
	err := DB.Where("name = ?", name).First(&key).Error
	return &key, err
}

// FindByUserID 根据用户ID查询SSH密钥（预留多用户支持）
func (d *SSHKeyDAO) FindByUserID(userID uint) ([]po.SSHKey, error) {
	var keys []po.SSHKey
	err := DB.Where("user_id = ?", userID).Find(&keys).Error
	return keys, err
}

// Update 更新SSH密钥
func (d *SSHKeyDAO) Update(sshKey *po.SSHKey) error {
	return DB.Save(sshKey).Error
}

// Delete 删除SSH密钥
func (d *SSHKeyDAO) Delete(id uint) error {
	return DB.Delete(&po.SSHKey{}, id).Error
}

// ExistsByName 检查名称是否已存在
func (d *SSHKeyDAO) ExistsByName(name string) (bool, error) {
	var count int64
	err := DB.Model(&po.SSHKey{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// ExistsByNameExcludeID 检查名称是否已存在（排除指定ID）
func (d *SSHKeyDAO) ExistsByNameExcludeID(name string, excludeID uint) (bool, error) {
	var count int64
	err := DB.Model(&po.SSHKey{}).Where("name = ? AND id != ?", name, excludeID).Count(&count).Error
	return count > 0, err
}
