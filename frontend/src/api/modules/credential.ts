import request from '../request'
import type {
  CredentialDTO,
  CreateCredentialReq,
  UpdateCredentialReq,
  TestCredentialResp,
  MatchCredentialResp,
} from '@/types/credential'

// 列出所有凭证
export function listCredentials() {
  return request.get<unknown, CredentialDTO[]>('/credentials/')
}

// 创建凭证
export function createCredential(data: CreateCredentialReq) {
  return request.post<unknown, CredentialDTO>('/credentials/', data)
}

// 获取凭证详情
export function getCredential(id: number) {
  return request.get<unknown, CredentialDTO>(`/credentials/${id}`)
}

// 更新凭证
export function updateCredential(id: number, data: UpdateCredentialReq) {
  return request.put<unknown, CredentialDTO>(`/credentials/${id}`, data)
}

// 删除凭证
export function deleteCredential(id: number) {
  return request.delete<unknown, { message: string }>(`/credentials/${id}`)
}

// 测试凭证连接
export function testCredential(id: number, url: string) {
  return request.post<unknown, TestCredentialResp>(`/credentials/${id}/test`, { url })
}

// 根据 URL 匹配推荐凭证
export function matchCredentials(url: string) {
  return request.post<unknown, MatchCredentialResp>('/credentials/match', { url })
}
