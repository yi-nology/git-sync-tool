// 凭证类型
export type CredentialType = 'ssh_key' | 'http_basic' | 'http_token'

// 凭证 DTO（脱敏响应）
export interface CredentialDTO {
  id: number
  name: string
  type: CredentialType
  description: string
  ssh_key_id?: number
  ssh_key_name?: string   // 关联 SSH 密钥名称
  ssh_key_type?: string   // 关联 SSH 密钥算法类型 (rsa/ed25519/ecdsa/dsa)
  ssh_key_path?: string
  username?: string
  has_secret: boolean
  url_pattern?: string
  last_used_at?: string
  created_at: string
  updated_at: string
}

// 创建凭证请求
export interface CreateCredentialReq {
  name: string
  type: CredentialType
  description?: string
  ssh_key_id?: number
  ssh_key_path?: string
  username?: string
  secret?: string
  url_pattern?: string
}

// 更新凭证请求
export interface UpdateCredentialReq {
  name?: string
  description?: string
  ssh_key_id?: number
  ssh_key_path?: string
  username?: string
  secret?: string
  url_pattern?: string
}

// 测试凭证请求
export interface TestCredentialReq {
  url: string
}

// 测试凭证响应
export interface TestCredentialResp {
  success: boolean
  message: string
}

// 匹配凭证请求
export interface MatchCredentialReq {
  url: string
}

// 匹配凭证响应
export interface MatchCredentialResp {
  recommended: CredentialDTO[]
  others: CredentialDTO[]
}
