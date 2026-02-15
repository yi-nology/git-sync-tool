import request from '../request'

// 文件浏览相关API

export interface TreeEntry {
  name: string
  path: string
  type: 'file' | 'dir'
  size: number
  mode: string
  hash: string
}

export interface BlobContent {
  content: string
  encoding: 'utf-8' | 'base64'
  size: number
  is_binary: boolean
  mime_type: string
}

export interface FileCommit {
  hash: string
  short_hash: string
  message: string
  author: string
  date: string
}

// 获取目录树
export function getFileTree(repoKey: string, params: { ref?: string; path?: string; recursive?: boolean }) {
  return request.get<unknown, { entries: TreeEntry[]; current_ref: string; current_path: string }>('/file/tree', {
    params: { repo_key: repoKey, ...params }
  })
}

// 获取文件内容
export function getFileBlob(repoKey: string, params: { ref?: string; path: string }) {
  return request.get<unknown, BlobContent>('/file/blob', {
    params: { repo_key: repoKey, ...params }
  })
}

// 获取文件历史
export function getFileHistory(repoKey: string, params: { ref?: string; path: string; limit?: number }) {
  return request.get<unknown, FileCommit[]>('/file/history', {
    params: { repo_key: repoKey, ...params }
  })
}
