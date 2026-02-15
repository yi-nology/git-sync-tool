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
  config_source: string
  remote_auths: Record<string, AuthInfo>
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
  config_source?: string
  remotes?: GitRemote[]
  remote_auths?: Record<string, AuthInfo>
}

export interface CloneRepoReq {
  remote_url: string
  local_path: string
  name?: string
  auth_type?: string
  auth_key?: string
  auth_secret?: string
  config_source?: string
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
