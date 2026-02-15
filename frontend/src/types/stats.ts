export interface AuthorStat {
  name: string
  email: string
  total_lines: number
  file_types: Record<string, number>
  time_trend: Record<string, number>
}

export interface StatsResponse {
  total_lines: number
  authors: AuthorStat[]
}

export interface LanguageStat {
  name: string
  files: number
  code: number
  comment: number
  blank: number
}

export interface LineStatsResponse {
  status: string
  progress: string
  total_files: number
  total_lines: number
  code_lines: number
  comment_lines: number
  blank_lines: number
  languages: LanguageStat[]
}

export interface LineStatsConfig {
  exclude_dirs: string[]
  exclude_patterns: string[]
}

export interface AuditLogDTO {
  id: number
  action: string
  target: string
  operator: string
  details: string
  ip_address: string
  user_agent: string
  created_at: string
}

export interface SystemConfig {
  debug_mode: boolean
  author_name: string
  author_email: string
}

export interface DirItem {
  name: string
  path: string
}

export interface ListDirsResp {
  parent: string
  current: string
  dirs: DirItem[]
}
