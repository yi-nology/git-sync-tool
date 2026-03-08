export interface SpecFileNode {
  name: string
  path: string
  is_dir: boolean
  children?: SpecFileNode[]
  size?: number
  mod_time?: string
}

export interface LintIssue {
  line: number
  column?: number
  end_line?: number
  end_column?: number
  message: string
  severity: 'error' | 'warning' | 'info'
  rule_id: string
  rule_name: string
}

export interface LintResult {
  issues: LintIssue[]
  error_count: number
  warning_count: number
  info_count: number
}

export type RuleCategory = 'required' | 'style' | 'best-practice' | 'custom'

export interface LintRule {
  id: string
  name: string
  description: string
  category: RuleCategory
  enabled: boolean
  severity: 'error' | 'warning' | 'info'
  pattern?: string
}

export interface SaveRequest {
  content: string
  message?: string
}

export interface CommitRequest {
  message: string
  content?: string
}

export interface CommitResponse {
  commit_hash: string
  message: string
}

export interface EditorState {
  currentFile: string | null
  content: string
  originalContent: string
  isDirty: boolean
  cursorPosition: { line: number; column: number } | null
}
