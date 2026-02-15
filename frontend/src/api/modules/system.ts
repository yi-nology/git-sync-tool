import request from '../request'
import type { SystemConfig, ListDirsResp } from '@/types/stats'

export function getSystemConfig() {
  return request.get<unknown, SystemConfig>('/system/config')
}

export function updateSystemConfig(data: SystemConfig) {
  return request.post('/system/config', data)
}

export function listDirs(path?: string, search?: string) {
  return request.get<unknown, ListDirsResp>('/system/dirs', {
    params: { path: path || '', search: search || '' },
  })
}

export function getSSHKeys() {
  return request.get<unknown, string[]>('/system/ssh-keys')
}

export function testConnection(url: string) {
  return request.post<unknown, { status: string; error?: string }>('/system/test-connection', { url })
}

export function getRepoStatus(repoKey: string) {
  return request.get('/system/repo/status', { params: { repo_key: repoKey } })
}

export function getRepoGitConfig(repoKey: string) {
  return request.get('/system/repo/git-config', { params: { repo_key: repoKey } })
}

export function submitChanges(data: { repo_key: string; message: string; push: boolean; author_name?: string; author_email?: string }) {
  return request.post('/system/repo/submit', data)
}
