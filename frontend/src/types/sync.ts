import type { RepoDTO } from './repo'

export interface SyncTaskDTO {
  id: number
  key: string
  source_repo_key: string
  source_remote: string
  source_branch: string
  target_repo_key: string
  target_remote: string
  target_branch: string
  push_options: string
  cron: string
  enabled: boolean
  created_at: string
  updated_at: string
  source_repo?: RepoDTO
  target_repo?: RepoDTO
}

export interface CreateSyncTaskReq {
  source_repo_key: string
  target_repo_key: string
  source_remote: string
  source_branch: string
  target_remote: string
  target_branch: string
  push_options?: string
  cron?: string
  enabled?: boolean
}

export interface UpdateSyncTaskReq extends CreateSyncTaskReq {
  key: string
}

export interface ExecuteSyncReq {
  repo_key: string
  source_remote: string
  source_branch: string
  target_remote: string
  target_branch: string
  push_options?: string
}

export interface SyncRunDTO {
  id: number
  task_key: string
  status: string
  commit_range: string
  error_message: string
  details: string
  start_time: string
  end_time: string
  created_at: string
  updated_at: string
  task?: SyncTaskDTO
}
