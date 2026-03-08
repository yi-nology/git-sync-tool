import request from '../request'
import type { PatchInfoDTO, PatchStatsDTO, GeneratePatchReq, SavePatchReq, ApplyPatchReq } from '@/types/patch'

// 生成 patch
export function generatePatch(data: GeneratePatchReq) {
  return request.post<unknown, { content: string }>('/patch/generate', data)
}

// 保存 patch
export function savePatch(data: SavePatchReq) {
  return request.post<unknown, { path: string; name: string }>('/patch/save', data)
}

// 列出所有 patch
export function listPatches(repoKey: string) {
  return request.get<unknown, PatchInfoDTO[]>('/patch/list', { params: { repo_key: repoKey } })
}

// 获取 patch 内容
export function getPatchContent(path: string) {
  return request.get<unknown, { content: string }>('/patch/content', { params: { path } })
}

// 下载 patch
export function getPatchDownloadUrl(path: string) {
  return `/api/v1/patch/download?path=${encodeURIComponent(path)}`
}

// 应用 patch
export function applyPatch(data: ApplyPatchReq) {
  return request.post<unknown, { message: string }>('/patch/apply', data)
}

// 检查 patch
export function checkPatch(repoKey: string, patchPath: string) {
  return request.post<unknown, PatchStatsDTO>('/patch/check', { repo_key: repoKey, patch_path: patchPath })
}

// 删除 patch
export function deletePatch(repoKey: string, patchPath: string) {
  return request.post<unknown, { message: string }>('/patch/delete', { repo_key: repoKey, patch_path: patchPath })
}
