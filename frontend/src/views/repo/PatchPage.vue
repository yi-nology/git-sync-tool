<template>
  <div class="patch-page">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="$router.push(`/repos/${repoKey}`)" :icon="ArrowLeft" text>
          返回仓库详情
        </el-button>
        <h2>Patch 管理</h2>
        <el-tag v-if="repoName" size="small" type="info">{{ repoName }}</el-tag>
      </div>
    </div>

    <PatchManager :repo-key="repoKey" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import PatchManager from '@/components/patch/PatchManager.vue'
import { getRepoDetail } from '@/api/modules/repo'

const route = useRoute()
const repoKey = route.params.repoKey as string
const repoName = ref('')

onMounted(async () => {
  try {
    const repo = await getRepoDetail(repoKey)
    repoName.value = repo.name
  } catch (e) {
    console.error('Failed to load repo info:', e)
  }
})
</script>

<style scoped>
.patch-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-left h2 {
  margin: 0;
  font-size: 20px;
}
</style>
