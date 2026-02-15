import request from '../request'

export interface VersionTag {
  name: string
  hash: string
  date: string
  message: string
  tagger: string
}

export function getVersionList(repoKey: string) {
  return request.get<unknown, VersionTag[]>('/version/list', { params: { repo_key: repoKey } })
}

export function getCurrentVersion(repoKey: string) {
  return request.get<unknown, string>('/version/current', { params: { repo_key: repoKey } })
}

export function getNextVersion(repoKey: string) {
  return request.get<unknown, { current: string; next: string }>('/version/next', { params: { repo_key: repoKey } })
}
