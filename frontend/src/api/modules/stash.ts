import request from '../request'

// Stash管理相关API

export interface StashEntry {
  index: number
  ref: string
  message: string
  branch: string
  date: string
}

// 列出stash
export function listStash(repoKey: string) {
  return request.get<unknown, { stashes: StashEntry[] }>('/stash/list', {
    params: { repo_key: repoKey }
  })
}

// 保存stash
export function saveStash(repoKey: string, message?: string, includeUntracked?: boolean) {
  return request.post('/stash/save', {
    repo_key: repoKey,
    message,
    include_untracked: includeUntracked
  })
}

// 应用stash
export function applyStash(repoKey: string, index: number) {
  return request.post('/stash/apply', {
    repo_key: repoKey,
    index
  })
}

// 弹出stash
export function popStash(repoKey: string, index: number) {
  return request.post('/stash/pop', {
    repo_key: repoKey,
    index
  })
}

// 删除stash
export function dropStash(repoKey: string, index: number) {
  return request.post('/stash/drop', {
    repo_key: repoKey,
    index
  })
}

// 清空stash
export function clearStash(repoKey: string) {
  return request.post('/stash/clear', {
    repo_key: repoKey
  })
}
