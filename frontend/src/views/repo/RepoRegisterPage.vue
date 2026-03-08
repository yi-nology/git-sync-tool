<template>
  <div class="register-page">
    <h2>注册本地仓库</h2>

    <!-- Step 1: 选择目录 -->
    <el-card class="step-card">
      <template #header>
        <div class="card-header">
          <span>1. 选择仓库目录</span>
          <el-tag v-if="selectedPath" type="success">已选择</el-tag>
        </div>
      </template>

      <div class="path-section">
        <el-input
          v-model="selectedPath"
          placeholder="点击下方按钮选择目录..."
          readonly
          size="large"
        >
          <template #prefix>
            <el-icon><Folder /></el-icon>
          </template>
        </el-input>

        <div class="btn-group">
          <el-button type="primary" size="large" @click="handleSelectDir" :loading="selectingDir">
            <el-icon><FolderOpened /></el-icon>
            选择目录
          </el-button>
          <el-button size="large" @click="handleScan" :disabled="!selectedPath" :loading="scanning">
            <el-icon><Search /></el-icon>
            扫描仓库
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- Step 2: 选择仓库 -->
    <el-card class="step-card" v-if="scannedRepos.length > 0">
      <template #header>
        <div class="card-header">
          <span>2. 选择要注册的仓库 (共 {{ scannedRepos.length }} 个)</span>
          <div class="header-actions">
            <el-button size="small" @click="selectAll">全选</el-button>
            <el-button size="small" @click="selectNone">取消全选</el-button>
          </div>
        </div>
      </template>

      <div class="repo-list">
        <div
          v-for="repo in scannedRepos"
          :key="repo.path"
          class="repo-item"
          :class="{ selected: selectedRepos.includes(repo.path) }"
          @click="toggleRepo(repo.path)"
        >
          <el-checkbox
            :model-value="selectedRepos.includes(repo.path)"
            @click.stop
            @change="toggleRepo(repo.path)"
          />

          <div class="repo-info">
            <div class="repo-name">
              <el-icon><FolderChecked /></el-icon>
              {{ repo.name }}
            </div>
            <div class="repo-meta">
              <el-tag size="small" type="info">{{ repo.current_branch || 'unknown' }}</el-tag>
              <span class="repo-path">{{ repo.path }}</span>
            </div>
            <div class="repo-remote" v-if="repo.remotes.length > 0">
              <el-icon><Link /></el-icon>
              {{ getMainRemote(repo) }}
            </div>
          </div>

          <div class="repo-status">
            <el-tag v-if="repo.has_changes" size="small" type="warning">有更改</el-tag>
            <el-tag v-else size="small" type="success">干净</el-tag>
            <span class="commit-hash" v-if="repo.last_commit">{{ repo.last_commit }}</span>
          </div>
        </div>
      </div>

      <div class="selection-info">
        已选择 <strong>{{ selectedRepos.length }}</strong> 个仓库
      </div>
    </el-card>

    <!-- Step 3: 凭证配置（可选） -->
    <el-card class="step-card" v-if="selectedRepos.length > 0">
      <template #header>
        <div class="card-header">
          <span>3. 凭证配置（可选）</span>
        </div>
      </template>

      <div class="credential-section">
        <p class="hint">为所有选中的仓库设置默认凭证（可选，后续可在仓库详情中配置）</p>
        <CredentialSelector
          v-model="defaultCredentialId"
          placeholder="选择默认认证凭证（可选）"
          style="width: 100%"
        />
      </div>

      <div class="action-buttons">
        <el-button @click="router.push('/repos')">取消</el-button>
        <el-button type="primary" size="large" @click="handleRegister" :loading="registering">
          <el-icon><Check /></el-icon>
          注册 {{ selectedRepos.length }} 个仓库
        </el-button>
      </div>
    </el-card>

    <!-- 空状态 -->
    <el-empty
      v-if="!scanning && scannedRepos.length === 0 && selectedPath && hasScanned"
      description="在该目录下未找到 Git 仓库"
    >
      <el-button @click="handleSelectDir">选择其他目录</el-button>
    </el-empty>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  Folder,
  FolderOpened,
  FolderChecked,
  Search,
  Link,
  Check,
} from '@element-plus/icons-vue'
import { selectDirectory, scanDirectory, batchCreateRepos } from '@/api/modules/repo'
import type { ScannedRepo } from '@/types/repo'
import CredentialSelector from '@/components/credential/CredentialSelector.vue'
import { useNotification } from '@/composables/useNotification'

const router = useRouter()
const { showSuccess, showError } = useNotification()

const selectedPath = ref('')
const selectingDir = ref(false)
const scanning = ref(false)
const hasScanned = ref(false)
const scannedRepos = ref<ScannedRepo[]>([])
const selectedRepos = ref<string[]>([])
const defaultCredentialId = ref<number | undefined>()
const registering = ref(false)

// 选择目录
async function handleSelectDir() {
  selectingDir.value = true
  try {
    const res = await selectDirectory('选择 Git 仓库所在目录')
    if (res.cancelled !== 'true' && res.path) {
      selectedPath.value = res.path
      scannedRepos.value = []
      selectedRepos.value = []
      hasScanned.value = false
      // 自动开始扫描
      await handleScan()
    }
  } catch (e: any) {
    showError('选择目录失败', e)
  } finally {
    selectingDir.value = false
  }
}

// 扫描目录
async function handleScan() {
  if (!selectedPath.value) return

  scanning.value = true
  try {
    const res = await scanDirectory(selectedPath.value, 2, true)
    scannedRepos.value = res.repos || []
    // 默认全选
    selectedRepos.value = scannedRepos.value.map(r => r.path)
    hasScanned.value = true

    if (res.total === 0) {
      // 未找到仓库，提示用户
    } else {
      showSuccess(`找到 ${res.total} 个 Git 仓库`)
    }
  } catch (e: any) {
    showError('扫描失败', e)
  } finally {
    scanning.value = false
  }
}

// 切换仓库选择
function toggleRepo(path: string) {
  const idx = selectedRepos.value.indexOf(path)
  if (idx >= 0) {
    selectedRepos.value.splice(idx, 1)
  } else {
    selectedRepos.value.push(path)
  }
}

// 全选
function selectAll() {
  selectedRepos.value = scannedRepos.value.map(r => r.path)
}

// 取消全选
function selectNone() {
  selectedRepos.value = []
}

// 获取主要远程
function getMainRemote(repo: ScannedRepo): string {
  const origin = repo.remotes.find(r => r.name === 'origin')
  const remote = origin || repo.remotes[0]
  if (!remote) return '无远程'
  // 简化 URL 显示
  const url = remote.fetch_url
  if (url.startsWith('https://github.com/')) {
    return url.replace('https://github.com/', 'github:')
  }
  if (url.startsWith('git@github.com:')) {
    return url.replace('git@github.com:', 'github:')
  }
  return url
}

// 注册仓库
async function handleRegister() {
  if (selectedRepos.value.length === 0) {
    showError('请至少选择一个仓库')
    return
  }

  registering.value = true
  try {
    const repos = selectedRepos.value.map(path => {
      const repo = scannedRepos.value.find(r => r.path === path)!
      return {
        name: repo.name,
        path: repo.path,
        default_credential_id: defaultCredentialId.value,
      }
    })

    const res = await batchCreateRepos({ repos })

    // 处理失败项（可能为 null）
    const failedList = res.failed || []
    const successList = res.success || []

    if (failedList.length > 0) {
      const failedNames = failedList.map(f => f.name).join(', ')
      showError(`${failedList.length} 个仓库注册失败: ${failedNames}`)
    }

    if (successList.length > 0) {
      showSuccess(`成功注册 ${successList.length} 个仓库`)
      router.push('/repos')
    }

    // 如果全部失败，不跳转
    if (failedList.length > 0 && successList.length === 0) {
      // 保持在当前页面
    }
  } catch (e: any) {
    showError('注册失败', e)
  } finally {
    registering.value = false
  }
}
</script>

<style scoped>
.register-page {
  padding: 20px;
  max-width: 900px;
  margin: 0 auto;
}

.register-page h2 {
  margin-bottom: 24px;
  font-size: 20px;
  font-weight: 600;
}

.step-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.path-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.btn-group {
  display: flex;
  gap: 12px;
}

.repo-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
}

.repo-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.repo-item:hover {
  border-color: var(--el-color-primary-light-5);
  background-color: var(--el-fill-color-light);
}

.repo-item.selected {
  border-color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-9);
}

.repo-info {
  flex: 1;
  min-width: 0;
}

.repo-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 15px;
  margin-bottom: 6px;
}

.repo-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 4px;
}

.repo-path {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  font-family: monospace;
}

.repo-remote {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.repo-status {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
}

.commit-hash {
  font-family: monospace;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.selection-info {
  margin-top: 16px;
  padding: 12px;
  background-color: var(--el-fill-color-light);
  border-radius: 6px;
  text-align: center;
  color: var(--el-text-color-regular);
}

.credential-section {
  margin-bottom: 20px;
}

.hint {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  margin-bottom: 12px;
}

.action-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 响应式 */
@media (max-width: 768px) {
  .register-page {
    padding: 16px;
  }

  .repo-item {
    flex-direction: column;
    align-items: flex-start;
  }

  .repo-status {
    flex-direction: row;
    width: 100%;
    justify-content: space-between;
    align-items: center;
    margin-top: 8px;
    padding-top: 8px;
    border-top: 1px solid var(--el-border-color-lighter);
  }
}
</style>
