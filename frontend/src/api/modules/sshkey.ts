import request from '../request'

// SSH密钥DTO
export interface DBSSHKey {
  id: number
  name: string
  description: string
  public_key: string
  key_type: string
  has_passphrase: boolean
  created_at: string
  updated_at: string
}

// 创建SSH密钥请求
export interface CreateDBSSHKeyReq {
  name: string
  description?: string
  private_key: string
  public_key?: string
  passphrase?: string
}

// 更新SSH密钥请求
export interface UpdateDBSSHKeyReq {
  description?: string
  private_key?: string
  passphrase?: string
}

// 测试连接请求
export interface TestDBSSHKeyReq {
  url: string
}

// 测试连接响应
export interface TestDBSSHKeyResp {
  success: boolean
  message: string
}

// 列出所有数据库SSH密钥
export function listDBSSHKeys() {
  return request.get<unknown, DBSSHKey[]>('/system/db-ssh-keys/')
}

// 创建SSH密钥
export function createDBSSHKey(data: CreateDBSSHKeyReq) {
  return request.post<unknown, DBSSHKey>('/system/db-ssh-keys/', data)
}

// 获取SSH密钥详情
export function getDBSSHKey(id: number) {
  return request.get<unknown, DBSSHKey>(`/system/db-ssh-keys/${id}`)
}

// 更新SSH密钥
export function updateDBSSHKey(id: number, data: UpdateDBSSHKeyReq) {
  return request.put<unknown, DBSSHKey>(`/system/db-ssh-keys/${id}`, data)
}

// 删除SSH密钥
export function deleteDBSSHKey(id: number) {
  return request.delete<unknown, { message: string }>(`/system/db-ssh-keys/${id}`)
}

// 测试SSH密钥连接
export function testDBSSHKey(id: number, data: TestDBSSHKeyReq) {
  return request.post<unknown, TestDBSSHKeyResp>(`/system/db-ssh-keys/${id}/test`, data)
}
