import request from '../request'
import type { StatsResponse, LineStatsResponse, LineStatsConfig } from '@/types/stats'

export function getStatsAnalyze(repoKey: string, params?: { branch?: string; author?: string; since?: string; until?: string }) {
  return request.get<unknown, StatsResponse>('/stats/analyze', {
    params: { repo_key: repoKey, ...params },
  })
}

export function getStatsAuthors(repoKey: string) {
  return request.get<unknown, { name: string; email: string }[]>('/stats/authors', { params: { repo_key: repoKey } })
}

export function getStatsBranches(repoKey: string) {
  return request.get<unknown, string[]>('/stats/branches', { params: { repo_key: repoKey } })
}

export function getStatsCommits(repoKey: string, params?: { branch?: string; author?: string; since?: string; until?: string }) {
  return request.get('/stats/commits', { params: { repo_key: repoKey, ...params } })
}

export function getLineStats(repoKey: string, params?: { branch?: string }) {
  return request.get<unknown, LineStatsResponse>('/stats/lines', {
    params: { repo_key: repoKey, ...params },
  })
}

export function getLineStatsConfig(repoKey: string) {
  return request.get<unknown, LineStatsConfig>('/stats/lines/config', { params: { repo_key: repoKey } })
}

export function saveLineStatsConfig(repoKey: string, data: LineStatsConfig) {
  return request.post('/stats/lines/config', { repo_key: repoKey, ...data })
}

export function exportStatsCsv(repoKey: string, params?: Record<string, string>) {
  return request.get('/stats/export/csv', {
    params: { repo_key: repoKey, ...params },
    responseType: 'blob',
  })
}
