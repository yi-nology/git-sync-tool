package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type LintRuleDAO struct{}

func NewLintRuleDAO() *LintRuleDAO {
	return &LintRuleDAO{}
}

func (d *LintRuleDAO) Create(rule *po.LintRule) error {
	return DB.Create(rule).Error
}

func (d *LintRuleDAO) BatchCreate(rules []po.LintRule) error {
	return DB.Create(&rules).Error
}

func (d *LintRuleDAO) FindAll() ([]po.LintRule, error) {
	var rules []po.LintRule
	err := DB.Order("priority ASC, id ASC").Find(&rules).Error
	return rules, err
}

func (d *LintRuleDAO) FindEnabled() ([]po.LintRule, error) {
	var rules []po.LintRule
	err := DB.Where("enabled = ?", true).Order("priority ASC, id ASC").Find(&rules).Error
	return rules, err
}

func (d *LintRuleDAO) FindByID(id string) (*po.LintRule, error) {
	var rule po.LintRule
	err := DB.Where("id = ?", id).First(&rule).Error
	return &rule, err
}

func (d *LintRuleDAO) FindByCategory(category string) ([]po.LintRule, error) {
	var rules []po.LintRule
	err := DB.Where("category = ?", category).Order("priority ASC, id ASC").Find(&rules).Error
	return rules, err
}

func (d *LintRuleDAO) FindByIDs(ids []string) ([]po.LintRule, error) {
	var rules []po.LintRule
	err := DB.Where("id IN ?", ids).Order("priority ASC, id ASC").Find(&rules).Error
	return rules, err
}

func (d *LintRuleDAO) Save(rule *po.LintRule) error {
	return DB.Save(rule).Error
}

func (d *LintRuleDAO) Delete(id string) error {
	return DB.Where("id = ?", id).Delete(&po.LintRule{}).Error
}

func (d *LintRuleDAO) ExistsByID(id string) (bool, error) {
	var count int64
	err := DB.Model(&po.LintRule{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (d *LintRuleDAO) Count() (int64, error) {
	var count int64
	err := DB.Model(&po.LintRule{}).Count(&count).Error
	return count, err
}
