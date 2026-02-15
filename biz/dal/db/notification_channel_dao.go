// biz/dal/db/notification_channel_dao.go - 通知渠道DAO

package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type NotificationChannelDAO struct{}

func NewNotificationChannelDAO() *NotificationChannelDAO {
	return &NotificationChannelDAO{}
}

func (d *NotificationChannelDAO) Create(channel *po.NotificationChannel) error {
	return DB.Create(channel).Error
}

func (d *NotificationChannelDAO) FindAll() ([]po.NotificationChannel, error) {
	var channels []po.NotificationChannel
	err := DB.Find(&channels).Error
	return channels, err
}

func (d *NotificationChannelDAO) FindByType(channelType string) ([]po.NotificationChannel, error) {
	var channels []po.NotificationChannel
	err := DB.Where("type = ?", channelType).Find(&channels).Error
	return channels, err
}

func (d *NotificationChannelDAO) FindByID(id uint) (*po.NotificationChannel, error) {
	var channel po.NotificationChannel
	err := DB.First(&channel, id).Error
	return &channel, err
}

func (d *NotificationChannelDAO) FindEnabled() ([]po.NotificationChannel, error) {
	var channels []po.NotificationChannel
	err := DB.Where("enabled = ?", true).Find(&channels).Error
	return channels, err
}

func (d *NotificationChannelDAO) Save(channel *po.NotificationChannel) error {
	return DB.Save(channel).Error
}

func (d *NotificationChannelDAO) Delete(channel *po.NotificationChannel) error {
	return DB.Delete(channel).Error
}
