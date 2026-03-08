import request from '../request'
import type { RepoDTO, RegisterRepoReq, CloneRepoReq, ScanResult, ScanDirectoryResp, BatchCreateReq, BatchCreateResp } from '@/types/repo'

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

// 新增：选择目录对话框
export function selectDirectory(title?: string) {
  return request.post<unknown, { path: string; cancelled: string }>('/system/select-directory', { title })
}

// 新增：扫描目录下的 Git 仓库
export function scanDirectory(path: string, depth: number = 2, recursive: boolean = true) {
  return request.post<unknown, ScanDirectoryResp>('/repo/scan-directory', { path, depth, recursive })
}

// 新增：批量注册仓库
export function batchCreateRepos(data: BatchCreateReq) {
  return request.post<unknown, BatchCreateResp>('/repo/batch-create', data)
}
