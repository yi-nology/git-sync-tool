import request from '../request'
import type { BranchInfo, CreateBranchReq, MergeReq, MergeCheckResult, CreateTagReq } from '@/types/branch'
import type { PaginationResponse } from '@/types/common'

export function getBranchList(repoKey: string, params?: { page?: number; page_size?: number; keyword?: string; type?: string }) {
  return request.get<unknown, PaginationResponse<BranchInfo>>('/branch/list', {
    params: { repo_key: repoKey, ...params },
  })
}

export function createBranch(data: CreateBranchReq) {
  return request.post('/branch/create', data)
}

export function deleteBranch(repoKey: string, name: string) {
  return request.post('/branch/delete', { repo_key: repoKey, name })
}

export function updateBranch(repoKey: string, name: string, newName: string, desc?: string) {
  return request.post('/branch/update', { repo_key: repoKey, name, new_name: newName, desc })
}

export function checkoutBranch(repoKey: string, name: string) {
  return request.post('/branch/checkout', { repo_key: repoKey, name })
}

export function pushBranch(repoKey: string, name: string, remotes: string[]) {
  return request.post('/branch/push', { repo_key: repoKey, name, remotes })
}

export function pullBranch(repoKey: string, name: string) {
  return request.post('/branch/pull', { repo_key: repoKey, name })
}

export function compareBranches(repoKey: string, base: string, target: string) {
  return request.get<unknown, { stat: { FilesChanged: number; Insertions: number; Deletions: number }; files: { path: string; status: string }[] }>('/branch/compare', {
    params: { repo_key: repoKey, base, target },
  })
}

export function getBranchDiff(repoKey: string, base: string, target: string, file?: string) {
  return request.get<unknown, { diff: string }>('/branch/diff', {
    params: { repo_key: repoKey, base, target, file },
  })
}

export function getBranchPatch(repoKey: string, base: string, target: string) {
  return request.get('/branch/patch', {
    params: { repo_key: repoKey, base, target },
    responseType: 'blob',
  })
}

export function checkMerge(repoKey: string, base: string, target: string) {
  return request.get<unknown, MergeCheckResult>('/branch/merge/check', {
    params: { repo_key: repoKey, base, target },
  })
}

export function mergeBranch(data: MergeReq) {
  return request.post('/branch/merge', data)
}

export function getTagList(repoKey: string) {
  return request.get<unknown, string[]>('/tag/list', { params: { repo_key: repoKey } })
}

export function createTag(data: CreateTagReq) {
  return request.post('/tag/create', data)
}
