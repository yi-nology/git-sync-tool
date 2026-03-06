/**
 * Git 远程 URL 格式校验
 *
 * 支持的格式：
 *   SSH SCP:      git@github.com:user/repo.git
 *   SSH Protocol: ssh://[user@]host[:port]/path
 *   HTTPS:        https://host/path
 *   HTTP:         http://host/path
 *   Git Protocol: git://host/path
 *   Local File:   file:///path  或  /absolute/path
 */

// SCP 风格 SSH: git@github.com:user/repo.git  或  user@host:path
const SCP_RE = /^[\w.-]+@[\w.-]+:[\w./_-]+$/

// 标准协议: ssh:// https:// http:// git:// file://
const PROTOCOL_RE = /^(ssh|https?|git|file):\/\/.+$/

// 本地绝对路径
const LOCAL_PATH_RE = /^\/[\w./_-]+$/

/**
 * 校验是否为合法的 Git 远程 URL
 * @returns true 表示合法
 */
export function isValidGitRemoteUrl(url: string): boolean {
  if (!url || !url.trim()) return false
  const trimmed = url.trim()
  return SCP_RE.test(trimmed) || PROTOCOL_RE.test(trimmed) || LOCAL_PATH_RE.test(trimmed)
}

/**
 * 返回 URL 格式错误提示信息，合法则返回空字符串
 */
export function validateGitRemoteUrl(url: string): string {
  if (!url || !url.trim()) return '请输入远程仓库 URL'
  if (isValidGitRemoteUrl(url)) return ''
  return '不支持的 URL 格式。支持: git@host:path、ssh://、https://、http://、git://、file:// 或本地绝对路径'
}

/**
 * 检测 Git URL 使用的协议类型
 */
export function detectGitProtocol(url: string): 'ssh' | 'http' | 'git' | 'file' | 'local' | 'unknown' {
  if (!url) return 'unknown'
  const trimmed = url.trim()
  if (trimmed.startsWith('git@') || trimmed.startsWith('ssh://')) return 'ssh'
  if (SCP_RE.test(trimmed) && !trimmed.includes('://')) return 'ssh'
  if (trimmed.startsWith('https://') || trimmed.startsWith('http://')) return 'http'
  if (trimmed.startsWith('git://')) return 'git'
  if (trimmed.startsWith('file://')) return 'file'
  if (trimmed.startsWith('/')) return 'local'
  return 'unknown'
}

/**
 * 从 Git URL 中提取仓库名
 */
export function extractRepoName(url: string): string {
  if (!url) return ''
  const trimmed = url.trim().replace(/\/+$/, '')
  // SCP 格式: git@host:user/repo.git
  const scpMatch = trimmed.match(/:([^/]+\/)?([^/]+?)(?:\.git)?$/)
  if (scpMatch && !trimmed.includes('://')) {
    return scpMatch[2] || ''
  }
  // 其他格式: 取最后一段路径
  const match = trimmed.match(/\/([^/]+?)(?:\.git)?$/)
  return match?.[1] || ''
}
