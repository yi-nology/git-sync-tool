import request from '../request'
import type { RepoDTO, RegisterRepoReq, CloneRepoReq, ScanResult } from '@/types/repo'

export function getRepoList() {
  return request.get<unknown, RepoDTO[]>('/repo/list')
}

export function getRepoDetail(repoKey: string) {
  return request.get<unknown, RepoDTO>('/repo/detail', { params: { key: repoKey } })
}

export function createRepo(data: RegisterRepoReq) {
  return request.post<unknown, RepoDTO>('/repo/create', data)
}

export function updateRepo(data: RegisterRepoReq & { key: string }) {
  return request.post('/repo/update', data)
}

export function deleteRepo(key: string) {
  return request.post('/repo/delete', { key })
}

export function cloneRepo(data: CloneRepoReq) {
  return request.post<unknown, { task_id: string }>('/repo/clone', data)
}

export function fetchRepo(repoKey: string) {
  return request.post('/repo/fetch', { repo_key: repoKey })
}

export function scanRepo(path: string) {
  return request.post<unknown, ScanResult>('/repo/scan', { path })
}

export function getCloneTask(taskId: string) {
  return request.get<unknown, { status: string; progress: string[]; error: string }>('/repo/task', {
    params: { task_id: taskId },
  })
}
