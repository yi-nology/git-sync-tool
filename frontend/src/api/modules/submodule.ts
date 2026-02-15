import request from '../request'

// Submodule管理相关API

export interface SubmoduleInfo {
  name: string
  path: string
  url: string
  branch: string
  commit: string
  status: 'initialized' | 'uninitialized' | 'modified' | 'unknown'
}

export interface SubmoduleStatusItem {
  path: string
  commit: string
  status: string  // +, -, U, 空
  description: string
}

// 列出submodule
export function listSubmodules(repoKey: string) {
  return request.get<unknown, { submodules: SubmoduleInfo[] }>('/submodule/list', {
    params: { repo_key: repoKey }
  })
}

// 获取submodule状态
export function getSubmoduleStatus(repoKey: string, recursive?: boolean) {
  return request.get<unknown, { items: SubmoduleStatusItem[] }>('/submodule/status', {
    params: { repo_key: repoKey, recursive }
  })
}

// 添加submodule
export function addSubmodule(repoKey: string, url: string, path: string, branch?: string) {
  return request.post('/submodule/add', {
    repo_key: repoKey,
    url,
    path,
    branch
  })
}

// 初始化submodule
export function initSubmodule(repoKey: string, path?: string) {
  return request.post('/submodule/init', {
    repo_key: repoKey,
    path
  })
}

// 更新submodule
export function updateSubmodule(repoKey: string, params: { path?: string; init?: boolean; recursive?: boolean; remote?: boolean }) {
  return request.post('/submodule/update', {
    repo_key: repoKey,
    ...params
  })
}

// 同步submodule URL
export function syncSubmodule(repoKey: string, path?: string, recursive?: boolean) {
  return request.post('/submodule/sync', {
    repo_key: repoKey,
    path,
    recursive
  })
}

// 移除submodule
export function removeSubmodule(repoKey: string, path: string, force?: boolean) {
  return request.post('/submodule/remove', {
    repo_key: repoKey,
    path,
    force
  })
}
