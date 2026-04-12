<template>
  <div class="register-page">
    <h2>注册本地仓库</h2>

    <!-- 导入模式选择 -->
    <el-card class="step-card">
      <template #header>
        <div class="card-header">
          <span>选择导入方式</span>
        </div>
      </template>
      
      <el-radio-group v-model="importMode" size="large" @change="resetState">
        <el-radio-button value="single">
          <el-icon><FolderChecked /></el-icon>
          单个仓库
        </el-radio-button>
        <el-radio-button value="batch">
          <el-icon><Folders /></el-icon>
          批量扫描
        </el-radio-button>
      </el-radio-group>
    </el-card>

    <!-- Step 1: 选择目录 -->
    <el-card class="step-card">
      <template #header>
        <div class="card-header">
          <span>{{ importMode === 'single' ? '选择仓库目录' : '选择扫描目录' }}</span>
          <el-tag v-if="selectedPath" type="success">已选择</el-tag>
        </div>
      </template>

      <div class="path-section">
        <el-input
          v-model="selectedPath"
          :placeholder="importMode === 'single' ? '选择 Git 仓库根目录...' : '选择包含多个仓库的父目录...'"
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
            {{ importMode === 'single' ? '选择仓库' : '选择目录' }}
          </el-button>
          <el-button 
            v-if="importMode === 'batch'"
            size="large" 
            @click="handleScan" 
            :disabled="!selectedPath" 
            :loading="scanning"
          >
            <el-icon><Search /></el-icon>
            扫描仓库
          </el-button>
          <el-button 
            v-else
            type="success"
            size="large" 
            @click="handleCheckSingleRepo" 
            :disabled="!selectedPath" 
            :loading="scanning"
          >
            <el-icon><Check /></el-icon>
            验证仓库
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 单仓库模式：显示仓库信息 -->
    <el-card class="step-card" v-if="importMode === 'single' && singleRepo">
      <template #header>
        <div class="card-header">
          <span>仓库信息</span>
          <el-tag type="success">有效 Git 仓库</el-tag>
        </div>
      </template>

      <div class="single-repo-info">
        <div class="repo-detail">
          <div class="detail-row">
            <span class="label">仓库名称:</span>
            <el-input v-model="singleRepoName" placeholder="输入仓库名称" style="width: 300px" />
          </div>
          <div class="detail-row">
            <span class="label">仓库路径:</span>
            <span class="value mono">{{ singleRepo.path }}</span>
          </div>
          <div class="detail-row">
            <span class="label">当前分支:</span>
            <el-tag size="small">{{ singleRepo.current_branch || 'unknown' }}</el-tag>
          </div>
          <div class="detail-row" v-if="singleRepo.remotes?.length">
            <span class="label">远程仓库:</span>
            <div class="remote-list">
              <el-tag v-for="r in singleRepo.remotes" :key="r.name" size="small" type="info" class="remote-tag">
                {{ r.name }}: {{ simplifyUrl(r.fetch_url) }}
              </el-tag>
            </div>
          </div>
          <div class="detail-row">
            <span class="label">工作区状态:</span>
            <el-tag :type="singleRepo.has_changes ? 'warning' : 'success'" size="small">
              {{ singleRepo.has_changes ? '有未提交更改' : '干净' }}
            </el-tag>
          </div>
        </div>

        <!-- 凭证配置 -->
        <el-divider />
        <div class="credential-section">
          <p class="hint">配置认证凭证（可选，后续可在仓库详情中修改）</p>
          <CredentialSelector
            v-model="defaultCredentialId"
            placeholder="选择认证凭证（可选）"
            style="width: 100%; max-width: 400px"
          />
        </div>

        <div class="action-buttons">
          <el-button @click="router.push('/repos')">取消</el-button>
          <el-button type="primary" size="large" @click="handleRegisterSingle" :loading="registering">
            <el-icon><Check /></el-icon>
            注册仓库
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 批量模式：选择仓库列表 -->
    <el-card class="step-card" v-if="importMode === 'batch' && scannedRepos.length > 0">
      <template #header>
        <div class="card-header">
          <span>选择要注册的仓库 (共 {{ scannedRepos.length }} 个)</span>
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

    <!-- 批量模式：凭证配置 -->
    <el-card class="step-card" v-if="importMode === 'batch' && selectedRepos.length > 0">
      <template #header>
        <div class="card-header">
          <span>凭证配置（可选）</span>
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
      v-if="!scanning && scannedRepos.length === 0 && selectedPath && hasScanned && importMode === 'batch'"
      description="在该目录下未找到 Git 仓库"
    >
      <el-button @click="handleSelectDir">选择其他目录</el-button>
    </el-empty>
    
    <!-- 单仓库模式：无效仓库 -->
    <el-empty
      v-if="!scanning && !singleRepo && selectedPath && hasScanned && importMode === 'single'"
      description="选择的目录不是有效的 Git 仓库"
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
import { selectDirectory, scanDirectory, batchCreateRepos, createRepo } from '@/api/modules/repo'
import type { ScannedRepo } from '@/types/repo'
import CredentialSelector from '@/components/credential/CredentialSelector.vue'
import { useNotification } from '@/composables/useNotification'

const router = useRouter()
const { showSuccess, showError } = useNotification()

const importMode = ref<'single' | 'batch'>('single')
const selectedPath = ref('')
const selectingDir = ref(false)
const scanning = ref(false)
const hasScanned = ref(false)

// 单仓库模式
const singleRepo = ref<ScannedRepo | null>(null)
const singleRepoName = ref('')

// 批量模式
const scannedRepos = ref<ScannedRepo[]>([])
const selectedRepos = ref<string[]>([])

const defaultCredentialId = ref<number | undefined>()
const registering = ref(false)

function resetState() {
  selectedPath.value = ''
  singleRepo.value = null
  singleRepoName.value = ''
  scannedRepos.value = []
  selectedRepos.value = []
  hasScanned.value = false
}

// 选择目录
async function handleSelectDir() {
  selectingDir.value = true
  try {
    const res = await selectDirectory(
      importMode.value === 'single' 
        ? '选择 Git 仓库根目录' 
        : '选择包含 Git 仓库的父目录'
    )
    if (res.cancelled !== 'true' && res.path) {
      selectedPath.value = res.path
      hasScanned.value = false
      singleRepo.value = null
      scannedRepos.value = []
      selectedRepos.value = []
      
      // 自动验证/扫描
      if (importMode.value === 'single') {
        await handleCheckSingleRepo()
      } else {
        await handleScan()
      }
    }
  } catch (e: any) {
    showError('选择目录失败', e)
  } finally {
    selectingDir.value = false
  }
}

// 验证单个仓库
async function handleCheckSingleRepo() {
  if (!selectedPath.value) return

  scanning.value = true
  try {
    // 扫描深度为 0，只检查当前目录
    const res = await scanDirectory(selectedPath.value, 0, false)
    const repos = res.repos || []
    
    if (repos.length > 0) {
      singleRepo.value = repos[0]!
      singleRepoName.value = repos[0]!.name
      showSuccess('检测到有效的 Git 仓库')
    } else {
      singleRepo.value = null
      showError('该目录不是有效的 Git 仓库')
    }
    hasScanned.value = true
  } catch (e: any) {
    showError('验证失败', e)
    singleRepo.value = null
    hasScanned.value = true
  } finally {
    scanning.value = false
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
      // 未找到仓库
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
  return simplifyUrl(remote.fetch_url)
}

// 简化 URL 显示
function simplifyUrl(url: string): string {
  if (!url) return ''
  if (url.startsWith('https://github.com/')) {
    return url.replace('https://github.com/', 'github:')
  }
  if (url.startsWith('git@github.com:')) {
    return url.replace('git@github.com:', 'github:')
  }
  return url
}

// 注册单个仓库
async function handleRegisterSingle() {
  if (!singleRepo.value) return
  
  const name = singleRepoName.value.trim()
  if (!name) {
    showError('请输入仓库名称')
    return
  }

  registering.value = true
  try {
    await createRepo({
      name,
      path: singleRepo.value.path,
      default_credential_id: defaultCredentialId.value,
    })
    showSuccess(`仓库 "${name}" 注册成功`)
    router.push('/repos')
  } catch (e: any) {
    showError('注册失败', e)
  } finally {
    registering.value = false
  }
}

// 批量注册仓库
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
  font-size: var(--font-size-xl);
  font-weight: 600;
  color: var(--text-color-primary);
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
  gap: var(--spacing-sm);
}

.path-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.btn-group {
  display: flex;
  gap: 12px;
}

.single-repo-info {
  padding: var(--spacing-sm) 0;
}

.repo-detail {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.detail-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.detail-row .label {
  min-width: 100px;
  color: var(--text-color-secondary);
  font-size: var(--font-size-md);
}

.detail-row .value {
  font-size: var(--font-size-md);
}

.detail-row .value.mono {
  font-family: monospace;
  color: var(--text-color-primary);
}

.remote-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}

.remote-tag {
  font-family: monospace;
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
  gap: var(--spacing-md);
  padding: var(--spacing-md);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.repo-item:hover {
  border-color: var(--primary-color);
  background-color: var(--accent-bg);
}

.repo-item.selected {
  border-color: var(--primary-color);
  background-color: var(--accent-bg);
}

.repo-info {
  flex: 1;
  min-width: 0;
}

.repo-name {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
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
  color: var(--text-color-secondary);
  font-size: var(--font-size-xs);
  font-family: monospace;
}

.repo-remote {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: var(--font-size-sm);
  color: var(--text-color-regular);
}

.repo-status {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
}

.commit-hash {
  font-family: monospace;
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
}

.selection-info {
  margin-top: var(--spacing-md);
  padding: 12px;
  background-color: var(--accent-bg);
  border-radius: var(--border-radius-sm);
  text-align: center;
  color: var(--text-color-regular);
}

.credential-section {
  margin-bottom: 20px;
}

.hint {
  color: var(--text-color-secondary);
  font-size: var(--font-size-sm);
  margin-bottom: 12px;
}

.action-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}

@media (max-width: 768px) {
  .register-page {
    padding: var(--spacing-md);
  }

  .detail-row {
    flex-direction: column;
    align-items: flex-start;
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
    margin-top: var(--spacing-sm);
    padding-top: var(--spacing-sm);
    border-top: 1px solid var(--border-color-light);
  }
}
</style>
