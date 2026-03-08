import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RepoDTO } from '@/types/repo'
import { getRepoList, getRepoDetail } from '@/api/modules/repo'

export const useRepoStore = defineStore('repo', () => {
  const repoList = ref<RepoDTO[]>([])
  const currentRepo = ref<RepoDTO | null>(null)
  const loading = ref(false)

  async function fetchRepoList() {
    loading.value = true
    try {
      const data = await getRepoList()
      // 确保 repoList 始终是数组
      repoList.value = Array.isArray(data) ? data : []
    } catch (error) {
      console.error('[RepoStore] Failed to fetch repo list:', error)
      repoList.value = []
    } finally {
      loading.value = false
    }
  }

  async function fetchRepoDetail(key: string) {
    loading.value = true
    try {
      currentRepo.value = await getRepoDetail(key)
    } catch (error) {
      console.error('[RepoStore] Failed to fetch repo detail:', error)
      currentRepo.value = null
    } finally {
      loading.value = false
    }
  }

  function getRepoByKey(key: string) {
    return repoList.value.find((r) => r.key === key)
  }

  return { repoList, currentRepo, loading, fetchRepoList, fetchRepoDetail, getRepoByKey }
})
