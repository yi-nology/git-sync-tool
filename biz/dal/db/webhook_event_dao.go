package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type WebhookEventDAO struct{}

func NewWebhookEventDAO() *WebhookEventDAO { return &WebhookEventDAO{} }

func (d *WebhookEventDAO) Create(event *po.WebhookEvent) error {
	return DB.Create(event).Error
}

func (d *WebhookEventDAO) FindByEventID(eventID string) (*po.WebhookEvent, error) {
	var event po.WebhookEvent
	err := DB.Where("event_id = ?", eventID).First(&event).Error
	return &event, err
}

func (d *WebhookEventDAO) List(eventType, source, status string, page, pageSize int) ([]po.WebhookEvent, int64, error) {
	q := DB.Model(&po.WebhookEvent{})
	if eventType != "" {
		q = q.Where("event_type = ?", eventType)
	}
	if source != "" {
		q = q.Where("source = ?", source)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var events []po.WebhookEvent
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&events).Error
	return events, total, err
}

func (d *WebhookEventDAO) Save(event *po.WebhookEvent) error {
	return DB.Save(event).Error
}

type WebhookRuleDAO struct{}

func NewWebhookRuleDAO() *WebhookRuleDAO { return &WebhookRuleDAO{} }

func (d *WebhookRuleDAO) Create(rule *po.WebhookRule) error {
	return DB.Create(rule).Error
}

func (d *WebhookRuleDAO) FindAll() ([]po.WebhookRule, error) {
	var rules []po.WebhookRule
	err := DB.Where("enabled = ?", true).Find(&rules).Error
	return rules, err
}

func (d *WebhookRuleDAO) FindByProviderConfigID(providerConfigID uint) ([]po.WebhookRule, error) {
	var rules []po.WebhookRule
	err := DB.Where("provider_config_id = ? OR provider_config_id = 0", providerConfigID).Where("enabled = ?", true).Find(&rules).Error
	return rules, err
}

func (d *WebhookRuleDAO) Save(rule *po.WebhookRule) error {
	return DB.Save(rule).Error
}

func (d *WebhookRuleDAO) Delete(id uint) error {
	return DB.Delete(&po.WebhookRule{}, id).Error
}
