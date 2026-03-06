<template>
  <div class="repo-list-page">
    <div class="page-header">
      <h2>仓库列表</h2>
      <el-dropdown @command="handleAddCommand">
        <el-button type="primary">
          <el-icon><Plus /></el-icon> 添加仓库
          <el-icon class="el-icon--right"><ArrowDown /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="register">
              <el-icon><FolderOpened /></el-icon> 注册本地仓库
            </el-dropdown-item>
            <el-dropdown-item command="clone">
              <el-icon><Download /></el-icon> 克隆远程仓库
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <el-table :data="repoStore.repoList" v-loading="repoStore.loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="名称" min-width="120" />
      <el-table-column prop="path" label="路径" min-width="200">
        <template #default="{ row }">
          <el-text type="info" size="small" class="mono-text">{{ row.path }}</el-text>
        </template>
      </el-table-column>
      <el-table-column prop="remote_url" label="远程" min-width="180">
        <template #default="{ row }">
          <el-text size="small" truncated>{{ row.remote_url || '-' }}</el-text>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button-group size="small">
            <el-button type="primary" @click="goToDetail(row.key)">
              <el-icon><View /></el-icon> 详情
            </el-button>
            <el-button type="success" @click="goToBranches(row.key)">
              <el-icon><Share /></el-icon> 分支
            </el-button>
            <el-button type="warning" @click="goToSync(row.key)">
              <el-icon><Refresh /></el-icon> 同步
            </el-button>
            <el-button type="danger" @click="handleDelete(row.key, row.name)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Plus, Delete, View, Share, Refresh, FolderOpened, Download, ArrowDown } from '@element-plus/icons-vue'
import { useRepoStore } from '@/stores/useRepoStore'
import { deleteRepo } from '@/api/modules/repo'

const router = useRouter()
const repoStore = useRepoStore()

onMounted(async () => {
  await repoStore.fetchRepoList()
})

function handleAddCommand(command: string) {
  if (command === 'register') {
    router.push('/repos/register')
  } else if (command === 'clone') {
    router.push('/repos/clone')
  }
}

function goToDetail(key: string) {
  router.push(`/repos/${key}`)
}
function goToBranches(key: string) {
  router.push(`/repos/${key}/branches`)
}
function goToSync(key: string) {
  router.push(`/repos/${key}/sync`)
}

async function handleDelete(key: string, name: string) {
  try {
    await ElMessageBox.confirm(`确定要删除仓库 "${name}" 吗？如果被同步任务使用将无法删除。`, '确认删除', {
      type: 'warning',
    })
    await deleteRepo(key)
    ElMessage.success('仓库已删除')
    await repoStore.fetchRepoList()
  } catch {
    // cancelled or error handled by request
  }
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.page-header h2 {
  margin: 0;
  font-size: 20px;
}
.mono-text {
  font-family: monospace;
}
</style>
