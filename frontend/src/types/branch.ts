export interface BranchInfo {
  name: string
  type: 'local' | 'remote'
  is_current: boolean
  hash: string
  author: string
  author_email: string
  date: string
  message: string
  upstream: string
  ahead: number
  behind: number
}

export interface CreateBranchReq {
  repo_key: string
  name: string
  base_ref?: string
}

export interface MergeReq {
  repo_key: string
  source: string
  target: string
  message?: string
  no_ff?: boolean
  squash?: boolean
}

export interface MergeCheckResult {
  success: boolean
  conflicts: string[]
  merge_id?: string
  report_url?: string
}

export interface DiffStats {
  files_changed: number
  insertions: number
  deletions: number
  file_list: DiffFile[]
}

export interface DiffFile {
  path: string
  status: string
  insertions: number
  deletions: number
}

export interface CreateTagReq {
  repo_key: string
  name: string
  ref: string
  message?: string
  push_remote?: string
}
