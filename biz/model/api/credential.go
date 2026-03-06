package api

import "time"

// CreateCredentialReq 创建凭证请求
type CreateCredentialReq struct {
	Name        string `json:"name"`         // 凭证名称（必填）
	Type        string `json:"type"`         // ssh_key, http_basic, http_token（必填）
	Description string `json:"description"`  // 描述信息
	SSHKeyID    uint   `json:"ssh_key_id"`   // 关联数据库 SSH 密钥 ID
	SSHKeyPath  string `json:"ssh_key_path"` // 本地 SSH 密钥文件路径
	Username    string `json:"username"`     // HTTP 用户名
	Secret      string `json:"secret"`       // 密码/Token/Passphrase
	URLPattern  string `json:"url_pattern"`  // URL 自动匹配模式
}

// UpdateCredentialReq 更新凭证请求
type UpdateCredentialReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SSHKeyID    uint   `json:"ssh_key_id"`
	SSHKeyPath  string `json:"ssh_key_path"`
	Username    string `json:"username"`
	Secret      string `json:"secret"`
	URLPattern  string `json:"url_pattern"`
}

// CredentialDTO 凭证响应（脱敏，不含 Secret 明文）
type CredentialDTO struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Description string     `json:"description"`
	SSHKeyID    uint       `json:"ssh_key_id,omitempty"`
	SSHKeyPath  string     `json:"ssh_key_path,omitempty"`
	Username    string     `json:"username,omitempty"`
	HasSecret   bool       `json:"has_secret"`
	URLPattern  string     `json:"url_pattern,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// MatchCredentialReq 根据 URL 匹配凭证请求
type MatchCredentialReq struct {
	URL string `json:"url"`
}

// MatchCredentialResp 凭证匹配响应
type MatchCredentialResp struct {
	Recommended []CredentialDTO `json:"recommended"` // 根据 URL 推荐的凭证
	Others      []CredentialDTO `json:"others"`      // 其他可用凭证
}

// TestCredentialReq 测试凭证连接请求
type TestCredentialReq struct {
	URL string `json:"url"`
}
