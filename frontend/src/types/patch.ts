// Patch 相关类型

export interface PatchInfoDTO {
  name: string
  path: string
  size: number
  mod_time: string
  sequence: number       // 序号
  is_applied: boolean    // 是否已应用
  can_apply: boolean     // 是否可以应用（考虑顺序）
  has_conflict: boolean  // 是否有冲突
}

export interface PatchSeriesDTO {
  repo_key: string
  total_patches: number
  applied_count: number
  pending_count: number
  conflict_count: number
  patches: PatchInfoDTO[]
  can_apply_next: boolean    // 是否可以应用下一个
  next_patch_index: number   // 下一个待应用的 patch 索引
}

export interface PatchStatsDTO {
  stat: string
  can_apply: boolean
  error?: string
  applied_commit?: string
}

export interface GeneratePatchReq {
  repo_key: string
  base?: string
  target?: string
  commits?: string[]
}

export interface SavePatchReq {
  repo_key: string
  patch_name: string
  patch_content: string
  custom_path?: string
  commit_message?: string
  sequence?: number
}

export interface ApplyPatchReq {
  repo_key: string
  patch_path?: string
  patch_content?: string
  sign_off?: boolean
  commit_message?: string
  force?: boolean
}

export interface DeletePatchReq {
  repo_key: string
  patch_path: string
  commit_message?: string
}

export interface ReorderPatchReq {
  repo_key: string
  patch_order: string[]
}
