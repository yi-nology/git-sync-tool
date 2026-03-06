package db

import (
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type CredentialDAO struct{}

func NewCredentialDAO() *CredentialDAO {
	return &CredentialDAO{}
}

// Create 创建凭证
func (d *CredentialDAO) Create(cred *po.Credential) error {
	return DB.Create(cred).Error
}

// FindAll 查询所有凭证
func (d *CredentialDAO) FindAll() ([]po.Credential, error) {
	var creds []po.Credential
	err := DB.Order("last_used_at DESC NULLS LAST, updated_at DESC").Find(&creds).Error
	return creds, err
}

// FindByID 根据 ID 查询凭证
func (d *CredentialDAO) FindByID(id uint) (*po.Credential, error) {
	var cred po.Credential
	err := DB.First(&cred, id).Error
	return &cred, err
}

// FindByName 根据名称查询凭证
func (d *CredentialDAO) FindByName(name string) (*po.Credential, error) {
	var cred po.Credential
	err := DB.Where("name = ?", name).First(&cred).Error
	return &cred, err
}

// FindByType 按类型查询凭证
func (d *CredentialDAO) FindByType(credType string) ([]po.Credential, error) {
	var creds []po.Credential
	err := DB.Where("type = ?", credType).Order("last_used_at DESC NULLS LAST, updated_at DESC").Find(&creds).Error
	return creds, err
}

// FindBySSHKeyID 查找引用指定 SSH 密钥的凭证
func (d *CredentialDAO) FindBySSHKeyID(sshKeyID uint) ([]po.Credential, error) {
	var creds []po.Credential
	err := DB.Where("type = ? AND ssh_key_id = ?", "ssh_key", sshKeyID).Find(&creds).Error
	return creds, err
}

// Save 更新凭证
func (d *CredentialDAO) Save(cred *po.Credential) error {
	return DB.Save(cred).Error
}

// Delete 删除凭证
func (d *CredentialDAO) Delete(id uint) error {
	return DB.Delete(&po.Credential{}, id).Error
}

// ExistsByName 检查名称是否已存在
func (d *CredentialDAO) ExistsByName(name string) (bool, error) {
	var count int64
	err := DB.Model(&po.Credential{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// ExistsByNameExcludeID 检查名称是否已存在（排除指定 ID）
func (d *CredentialDAO) ExistsByNameExcludeID(name string, excludeID uint) (bool, error) {
	var count int64
	err := DB.Model(&po.Credential{}).Where("name = ? AND id != ?", name, excludeID).Count(&count).Error
	return count > 0, err
}

// UpdateLastUsed 更新最后使用时间
func (d *CredentialDAO) UpdateLastUsed(id uint) error {
	now := time.Now()
	return DB.Model(&po.Credential{}).Where("id = ?", id).Update("last_used_at", &now).Error
}

// FindMatchingURL 根据 URL 匹配凭证（通过 url_pattern 和协议类型）
func (d *CredentialDAO) FindMatchingURL(url string) (recommended []po.Credential, others []po.Credential, err error) {
	var all []po.Credential
	if err = DB.Order("last_used_at DESC NULLS LAST, updated_at DESC").Find(&all).Error; err != nil {
		return
	}

	// 检测 URL 协议类型
	isSSH := isSSHURL(url)

	for _, cred := range all {
		// 检查 url_pattern 匹配
		if cred.URLPattern != "" && matchURLPattern(cred.URLPattern, url) {
			recommended = append(recommended, cred)
			continue
		}

		// 按协议类型匹配
		if isSSH && cred.Type == "ssh_key" {
			recommended = append(recommended, cred)
		} else if !isSSH && (cred.Type == "http_basic" || cred.Type == "http_token") {
			recommended = append(recommended, cred)
		} else {
			others = append(others, cred)
		}
	}
	return
}

// isSSHURL 检测 URL 是否为 SSH 协议
func isSSHURL(url string) bool {
	return strings.HasPrefix(url, "git@") ||
		strings.HasPrefix(url, "ssh://") ||
		strings.Contains(url, "@") && !strings.HasPrefix(url, "http")
}

// matchURLPattern 简单的 URL 模式匹配（支持 * 通配符前缀）
func matchURLPattern(pattern, url string) bool {
	// 从 URL 中提取主机名
	host := extractHost(url)
	if host == "" {
		return false
	}

	// 支持 *.github.com 格式
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // .github.com
		return strings.HasSuffix(host, suffix) || host == pattern[2:]
	}

	return host == pattern
}

// extractHost 从 Git URL 中提取主机名
func extractHost(url string) string {
	// git@github.com:user/repo.git
	if strings.HasPrefix(url, "git@") {
		parts := strings.SplitN(url[4:], ":", 2)
		if len(parts) > 0 {
			return parts[0]
		}
	}
	// ssh://git@github.com/user/repo.git
	if strings.HasPrefix(url, "ssh://") {
		url = url[6:]
		if idx := strings.Index(url, "@"); idx >= 0 {
			url = url[idx+1:]
		}
		if idx := strings.IndexAny(url, ":/"); idx >= 0 {
			return url[:idx]
		}
		return url
	}
	// https://github.com/user/repo.git
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "https://")
		url = strings.TrimPrefix(url, "http://")
		if idx := strings.IndexAny(url, ":/"); idx >= 0 {
			return url[:idx]
		}
		return url
	}
	return ""
}
