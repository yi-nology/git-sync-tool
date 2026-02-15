import request from '../request'
import type { SyncTaskDTO, CreateSyncTaskReq, UpdateSyncTaskReq, ExecuteSyncReq, SyncRunDTO } from '@/types/sync'

export function getSyncTasks(repoKey?: string) {
  return request.get<unknown, SyncTaskDTO[]>('/sync/tasks', {
    params: repoKey ? { repo_key: repoKey } : {},
  })
}

export function getSyncTask(key: string) {
  return request.get<unknown, SyncTaskDTO>('/sync/task', { params: { key } })
}

export function createSyncTask(data: CreateSyncTaskReq) {
  return request.post('/sync/task/create', data)
}

export function updateSyncTask(data: UpdateSyncTaskReq) {
  return request.post('/sync/task/update', data)
}

export function deleteSyncTask(key: string) {
  return request.post('/sync/task/delete', { key })
}

export function runSyncTask(taskKey: string) {
  return request.post('/sync/run', { task_key: taskKey })
}

export function executeSyncOnce(data: ExecuteSyncReq) {
  return request.post<unknown, { task_key: string }>('/sync/execute', data)
}

export function getSyncHistory(repoKey?: string) {
  return request.get<unknown, SyncRunDTO[]>('/sync/history', {
    params: repoKey ? { repo_key: repoKey } : {},
  })
}

export function deleteSyncHistory(id: number) {
  return request.post('/sync/history/delete', { id })
}
