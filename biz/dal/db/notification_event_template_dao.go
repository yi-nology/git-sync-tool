// biz/dal/db/notification_event_template_dao.go - 事件模板DAO

package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm"
)

type NotificationEventTemplateDAO struct{}

func NewNotificationEventTemplateDAO() *NotificationEventTemplateDAO {
	return &NotificationEventTemplateDAO{}
}

// FindByChannelID 查询渠道的所有事件模板
func (d *NotificationEventTemplateDAO) FindByChannelID(channelID uint) ([]po.NotificationEventTemplate, error) {
	var templates []po.NotificationEventTemplate
	err := DB.Where("channel_id = ?", channelID).Find(&templates).Error
	return templates, err
}

// FindByChannelAndEvent 查询特定渠道+事件的模板
func (d *NotificationEventTemplateDAO) FindByChannelAndEvent(channelID uint, eventType string) (*po.NotificationEventTemplate, error) {
	var tmpl po.NotificationEventTemplate
	err := DB.Where("channel_id = ? AND event_type = ?", channelID, eventType).First(&tmpl).Error
	if err != nil {
		return nil, err
	}
	return &tmpl, nil
}

// ReplaceByChannelID 替换渠道的所有事件模板（在给定事务中先删后插）
func (d *NotificationEventTemplateDAO) ReplaceByChannelID(tx *gorm.DB, channelID uint, templates []po.NotificationEventTemplate) error {
	// 硬删除该渠道的所有旧模板（Unscoped 避免软删除导致 UNIQUE 约束冲突）
	if err := tx.Unscoped().Where("channel_id = ?", channelID).Delete(&po.NotificationEventTemplate{}).Error; err != nil {
		return err
	}
	// 批量插入新模板
	if len(templates) > 0 {
		return tx.Create(&templates).Error
	}
	return nil
}

// DeleteByChannelID 删除渠道的所有事件模板
func (d *NotificationEventTemplateDAO) DeleteByChannelID(tx *gorm.DB, channelID uint) error {
	return tx.Unscoped().Where("channel_id = ?", channelID).Delete(&po.NotificationEventTemplate{}).Error
}
