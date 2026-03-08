export interface GitRemote {
  name: string
  fetch_url: string
  push_url: string
  is_mirror: boolean
}

export interface AuthInfo {
  type: string
  key: string
  secret: string
  source?: string      // "local" or "database"
  ssh_key_id?: number  // Database SSH Key ID (when source="database")
}

export interface RepoDTO {
  id: number
  key: string
  name: string
  path: string
  remote_url: string
  auth_type: string
  auth_key: string
  auth_secret: string
  remote_auths: Record<string, AuthInfo>
  default_credential_id?: number
  remote_credentials?: Record<string, number>
  created_at: string
  updated_at: string
}

export interface RegisterRepoReq {
  name: string
  path: string
  remote_url?: string
  auth_type?: string
  auth_key?: string
  auth_secret?: string
  remotes?: GitRemote[]
  remote_auths?: Record<string, AuthInfo>
  default_credential_id?: number
  remote_credentials?: Record<string, number>
}

export interface CloneRepoReq {
  remote_url: string
  local_path: string
  name?: string
  auth_type?: string
  auth_key?: string
  auth_secret?: string
  ssh_key_id?: number
  credential_id?: number
}

export interface ScanRepoReq {
  path: string
}

export interface ScanResult {
  remotes: GitRemote[]
  branches: TrackingBranch[]
}

export interface TrackingBranch {
  name: string
  upstream_ref: string
}

// 新增：扫描到的仓库信息
export interface ScannedRepo {
  name: string
  path: string
  remotes: GitRemote[]
  current_branch: string
  last_commit: string
  has_changes: boolean
}

// 新增：扫描目录响应
export interface ScanDirectoryResp {
  repos: ScannedRepo[]
  total: number
}

// 新增：批量注册请求项
export interface BatchRepoItem {
  name: string
  path: string
  default_credential_id?: number
}

// 新增：批量注册请求
export interface BatchCreateReq {
  repos: BatchRepoItem[]
}

// 新增：批量注册失败项
export interface BatchFailedItem {
  name: string
  path: string
  reason: string
}

// 新增：批量注册响应
export interface BatchCreateResp {
  success: RepoDTO[]
  failed: BatchFailedItem[]
}
