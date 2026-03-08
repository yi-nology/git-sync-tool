// Patch 相关类型

export interface PatchInfoDTO {
  name: string
  path: string
  size: number
  mod_time: string
}

export interface PatchStatsDTO {
  stat: string
  can_apply: boolean
  error?: string
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
}

export interface ApplyPatchReq {
  repo_key: string
  patch_path?: string
  patch_content?: string
  sign_off?: boolean
  commit_message?: string
}

export interface DeletePatchReq {
  repo_key: string
  patch_path: string
}
