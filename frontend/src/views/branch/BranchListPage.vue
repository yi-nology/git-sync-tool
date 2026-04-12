<template>
  <div class="branch-list-page">
    <div class="page-header">
      <div class="header-left">
        <button class="back-btn" @click="$router.push(`/repos/${repoKey}`)">
          <el-icon><ArrowLeft /></el-icon> 返回
        </button>
        <h2>分支管理</h2>
      </div>
      <div class="header-actions">
        <button class="action-pill action-pill--green" @click="$router.push(`/repos/${repoKey}/compare`)">
          <el-icon><Switch /></el-icon> 分支对比 & 合并
        </button>
        <button class="action-pill action-pill--outline" @click="handleFetchAll" :disabled="fetchLoading">
          <el-icon><Download /></el-icon> 刷新远端 (Fetch)
        </button>
        <button class="action-pill action-pill--primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon> 新建分支
        </button>
      </div>
    </div>

    <div class="tab-bar">
      <div
        class="tab-item"
        :class="{ active: activeTab === 'local' }"
        @click="handleTabChange('local')"
      >Local (本地分支)</div>
      <div
        v-for="remoteName in remoteNames"
        :key="remoteName"
        class="tab-item"
        :class="{ active: activeTab === `remote-${remoteName}` }"
        @click="handleTabChange(`remote-${remoteName}`)"
      >Remote (远端分支) - {{ remoteName }}</div>
    </div>

    <div class="search-card">
      <el-icon class="search-icon"><Search /></el-icon>
      <input
        v-model="searchQuery"
        placeholder="搜索分支名/作者..."
        class="search-input"
        @keyup.enter="loadBranches"
      />
      <el-icon v-if="searchQuery" class="clear-icon" @click="searchQuery = ''; loadBranches()"><Close /></el-icon>
    </div>

    <div class="branch-table-card" v-loading="loading">
      <div class="table-header">
        <span class="th" style="width:180px">分支名称</span>
        <span class="th" style="width:200px">最新提交</span>
        <span class="th" style="width:140px">提交人</span>
        <span class="th" style="width:140px">提交时间</span>
        <span v-if="activeTab === 'local'" class="th" style="width:160px">上游分支</span>
        <span v-if="activeTab === 'local'" class="th" style="width:120px">状态</span>
        <span v-else class="th" style="width:120px">本地关联</span>
        <span class="th" style="flex:1">操作</span>
      </div>
      <div
        v-for="row in branches"
        :key="row.name"
        class="table-row"
      >
        <span class="td" style="width:180px">
          <template v-if="activeTab === 'local'">
            <span class="branch-name-cell" :class="{ 'branch-current': row.is_current }">
              <el-icon v-if="row.is_current" class="current-icon"><CircleCheck /></el-icon>
              {{ row.name }}
            </span>
          </template>
          <template v-else>
            <span class="branch-name-cell">{{ row.name.replace(`${activeTab.replace('remote-', '')}/`, '') }}</span>
          </template>
        </span>
        <span class="td td-mono" style="width:200px">
          <span class="hash-text">{{ row.hash ? row.hash.substring(0, 8) : '-' }}</span>
          <span v-if="row.message" class="commit-msg">{{ row.message }}</span>
        </span>
        <span class="td" style="width:140px">
          <span class="author-name">{{ row.author }}</span>
          <span v-if="row.author_email" class="author-email">{{ row.author_email }}</span>
        </span>
        <span class="td" style="width:140px">{{ formatRelativeTime(row.date) }}</span>
        <template v-if="activeTab === 'local'">
          <span class="td" style="width:160px">
            <span v-if="row.upstream" class="tag tag--info">{{ row.upstream }}</span>
            <span v-else class="text-muted">无上游</span>
          </span>
          <span class="td" style="width:120px">
            <template v-if="row.upstream">
              <span v-if="row.ahead > 0" class="tag tag--success">{{ row.ahead }}↑</span>
              <span v-if="row.behind > 0" class="tag tag--warning">{{ row.behind }}↓</span>
              <span v-if="row.ahead === 0 && row.behind === 0" class="tag tag--success">已同步</span>
            </template>
            <span v-else class="text-muted">无上游</span>
          </span>
        </template>
        <template v-else>
          <span class="td" style="width:120px">
            <span v-if="getLocalBranch(row.name)" class="tag tag--success">{{ getLocalBranch(row.name) }}</span>
            <span v-else class="text-muted">无关联</span>
          </span>
        </template>
        <span class="td" style="flex:1">
          <template v-if="activeTab === 'local'">
            <el-dropdown @command="(cmd: string) => handleBranchCommand(cmd, row)">
              <button class="row-action-btn row-action-btn--primary">操作</button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item v-if="!row.is_current" command="checkout">
                    <el-icon><Select /></el-icon> 切换
                  </el-dropdown-item>
                  <el-dropdown-item command="push">
                    <el-icon><Top /></el-icon> 推送
                  </el-dropdown-item>
                  <el-dropdown-item command="pull">
                    <el-icon><Bottom /></el-icon> 拉取
                  </el-dropdown-item>
                  <el-dropdown-item command="tag">
                    <el-icon><PriceTag /></el-icon> 打标签
                  </el-dropdown-item>
                  <el-dropdown-item command="detail">
                    <el-icon><View /></el-icon> 详情
                  </el-dropdown-item>
                  <el-dropdown-item command="rename">
                    <el-icon><Edit /></el-icon> 重命名
                  </el-dropdown-item>
                  <el-dropdown-item v-if="!row.is_current" command="delete" divided>
                    <el-text type="danger"><el-icon><Delete /></el-icon> 删除</el-text>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
          <template v-else>
            <el-dropdown @command="(cmd: string) => handleRemoteBranchCommand(cmd, row)">
              <button class="row-action-btn row-action-btn--primary">操作</button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item v-if="!getLocalBranch(row.name)" command="checkout">
                    <el-icon><Download /></el-icon> 检出为本地
                  </el-dropdown-item>
                  <el-dropdown-item v-if="getLocalBranch(row.name)" command="update">
                    <el-icon><Bottom /></el-icon> 更新本地
                  </el-dropdown-item>
                  <el-dropdown-item v-if="getLocalBranch(row.name)" command="sync">
                    <el-icon><RefreshRight /></el-icon> 同步本地
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </span>
      </div>
      <div v-if="branches.length === 0 && !loading" class="empty-table">
        <span class="text-muted">{{ activeTab === 'local' ? '暂无本地分支' : '暂无远端分支' }}</span>
      </div>
    </div>

    <div class="table-footer">
      <span class="pag-info">
        {{ activeTab === 'local' ? `共 ${total} 个本地分支` : `共 ${total} 个远端分支 (${activeTab.replace('remote-', '')})` }}
      </span>
    </div>

    <!-- Create Branch Dialog -->
    <el-dialog v-model="showCreateDialog" title="新建分支" width="480px" destroy-on-close>
      <el-form :model="createForm" label-width="110px">
        <el-form-item label="新分支名称" required>
          <el-input v-model="createForm.name" />
        </el-form-item>
        <el-form-item label="基于 (Base Ref)">
          <el-input v-model="createForm.base_ref" placeholder="默认为当前 HEAD" />
          <div class="form-tip">可以是分支名、标签或 Commit Hash</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- Rename Branch Dialog -->
    <el-dialog v-model="showRenameDialog" title="重命名分支" width="480px" destroy-on-close>
      <el-form :model="renameForm" label-width="90px">
        <el-form-item label="当前名称">
          <el-input :model-value="renameForm.old_name" disabled />
        </el-form-item>
        <el-form-item label="新名称" required>
          <el-input v-model="renameForm.new_name" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRenameDialog = false">取消</el-button>
        <el-button type="primary" @click="handleRename">保存</el-button>
      </template>
    </el-dialog>

    <!-- Push Branch Dialog -->
    <el-dialog v-model="showPushDialog" :title="`推送分支: ${pushBranchName}`" width="480px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="目标远端">
          <el-checkbox-group v-model="pushRemotes">
            <el-checkbox v-for="r in remoteNames" :key="r" :label="r" :value="r" />
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <el-alert type="info" :closable="false" show-icon>
        推送操作将把本地分支更新推送到选定的远端仓库。
      </el-alert>
      <template #footer>
        <el-button @click="showPushDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitPush">确认推送</el-button>
      </template>
    </el-dialog>

    <!-- Create Tag Dialog -->
    <el-dialog v-model="showTagDialog" title="打标签 (Tag)" width="550px" destroy-on-close>
      <el-form :model="tagForm" label-width="100px">
        <el-form-item label="目标引用">
          <el-input :model-value="tagForm.ref" disabled />
        </el-form-item>
        <el-form-item label="版本类型">
          <el-radio-group v-model="tagForm.versionType" @change="handleTagVersionTypeChange">
            <el-radio-button value="patch">Patch</el-radio-button>
            <el-radio-button value="minor">Minor</el-radio-button>
            <el-radio-button value="major">Major</el-radio-button>
            <el-radio-button value="custom">自定义</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="当前版本" v-if="tagNextVersion">
          <el-tag type="info" size="small">{{ tagNextVersion.current || '无' }}</el-tag>
        </el-form-item>
        <el-form-item label="标签名" required>
          <el-input v-model="tagForm.name" :disabled="tagForm.versionType !== 'custom'" placeholder="v1.0.0" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="tagForm.message" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="推送到远端">
          <el-select v-model="tagForm.push_remote" placeholder="不推送" clearable>
            <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showTagDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateTag">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Plus, Delete, Edit, Select, Top, Bottom, Switch, Download, CircleCheck, PriceTag, View, RefreshRight, Search, Close } from '@element-plus/icons-vue'
import { getBranchList, createBranch, deleteBranch, updateBranch, checkoutBranch, pushBranch, pullBranch, createTag } from '@/api/modules/branch'
import { fetchRepo, scanRepo } from '@/api/modules/repo'
import { getRepoDetail } from '@/api/modules/repo'
import { getNextVersion } from '@/api/modules/version'
import type { NextVersionInfo } from '@/api/modules/version'
import type { BranchInfo } from '@/types/branch'
import { formatRelativeTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const repoKey = route.params.repoKey as string

const loading = ref(false)
const fetchLoading = ref(false)
const branches = ref<BranchInfo[]>([])
const localBranches = ref<BranchInfo[]>([])
const total = ref(0)
const activeTab = ref('local')
const searchQuery = ref('')
const remoteNames = ref<string[]>([])

const showCreateDialog = ref(false)
const createForm = ref({ name: '', base_ref: '' })

const showRenameDialog = ref(false)
const renameForm = ref({ old_name: '', new_name: '' })

const showPushDialog = ref(false)
const pushBranchName = ref('')
const pushRemotes = ref<string[]>([])

const showTagDialog = ref(false)
const tagForm = ref({ ref: '', name: '', message: '', push_remote: '', versionType: 'patch' as 'patch' | 'minor' | 'major' | 'custom' })
const tagNextVersion = ref<NextVersionInfo | null>(null)

onMounted(async () => {
  // 先获取远程源列表
  try {
    const repo = await getRepoDetail(repoKey)
    if (repo?.path) {
      const scan = await scanRepo(repo.path)
      remoteNames.value = (scan.remotes || []).map((r: { name: string }) => r.name)
    }
  } catch { /* ignore */ }
  
  // 加载本地分支用于远程关联
  try {
    const res = await getBranchList(repoKey, { type: 'local', page_size: 500 })
    localBranches.value = res.list || []
  } catch { /* ignore */ }
  
  // 加载当前标签页的数据
  await loadBranches()
})

// 标签页切换处理
function handleTabChange(tab: string) {
  activeTab.value = tab
  loadBranches()
}

async function loadBranches() {
  loading.value = true
  try {
    let branchType = 'local'
    let remoteName = ''
    
    if (activeTab.value.startsWith('remote-')) {
      branchType = 'remote'
      remoteName = activeTab.value.replace('remote-', '')
    }
    
    const res = await getBranchList(repoKey, {
      type: branchType,
      keyword: searchQuery.value || undefined,
      page_size: 500,
    })
    
    let filteredBranches = res.list || []
    // 如果是特定远程源的标签页，过滤出对应远程源的分支
    if (remoteName) {
      filteredBranches = filteredBranches.filter(branch => branch.name.startsWith(`${remoteName}/`))
    }
    
    branches.value = filteredBranches
    total.value = filteredBranches.length
  } finally {
    loading.value = false
  }
}

function getLocalBranch(remoteName: string): string | null {
  // Remote branch like "origin/main" -> check if local "main" exists with upstream
  const parts = remoteName.split('/')
  if (parts.length < 2) return null
  const localName = parts.slice(1).join('/')
  const found = localBranches.value.find(b => b.name === localName)
  return found ? found.name : null
}

function goDetail(branchName: string) {
  router.push(`/repos/${repoKey}/branches/${encodeURIComponent(branchName)}`)
}

async function handleFetchAll() {
  fetchLoading.value = true
  try {
    await fetchRepo(repoKey)
    ElMessage.success('远端数据已刷新')
    await loadBranches()
  } finally {
    fetchLoading.value = false
  }
}

async function handleCheckout(name: string) {
  try {
    await checkoutBranch(repoKey, name)
    ElMessage.success(`已切换到 ${name}`)
    await loadBranches()
  } catch { /* handled */ }
}

async function handleCheckoutRemote(remoteName: string) {
  // Checkout remote branch as local: origin/feature -> feature
  const parts = remoteName.split('/')
  const localName = parts.length >= 2 ? parts.slice(1).join('/') : remoteName
  try {
    await createBranch({
      repo_key: repoKey,
      name: localName,
      base_ref: remoteName,
    })
    ElMessage.success(`已检出为本地分支 ${localName}`)
    // Reload local branches cache
    const res = await getBranchList(repoKey, { type: 'local', page_size: 500 })
    localBranches.value = res.list || []
    await loadBranches()
  } catch { /* handled */ }
}

async function handleFfRemote(remoteName: string) {
  const localName = getLocalBranch(remoteName)
  if (!localName) return
  try {
    await pullBranch(repoKey, localName)
    ElMessage.success(`已更新本地分支 ${localName}`)
    await loadBranches()
  } catch { /* handled */ }
}

async function handlePullRemote(remoteName: string) {
  const localName = getLocalBranch(remoteName)
  if (!localName) return
  try {
    await pullBranch(repoKey, localName)
    ElMessage.success(`已同步本地分支 ${localName}`)
    await loadBranches()
  } catch { /* handled */ }
}

function handlePush(name: string) {
  pushBranchName.value = name
  const first = remoteNames.value[0]
  pushRemotes.value = first ? [first] : []
  showPushDialog.value = true
}

async function handleSubmitPush() {
  if (!pushRemotes.value.length) {
    ElMessage.warning('请选择目标远端')
    return
  }
  try {
    await pushBranch(repoKey, pushBranchName.value, pushRemotes.value)
    ElMessage.success('推送成功')
    showPushDialog.value = false
    await loadBranches()
  } catch { /* handled */ }
}

async function handlePull(name: string) {
  try {
    await pullBranch(repoKey, name)
    ElMessage.success('拉取成功')
    await loadBranches()
  } catch { /* handled */ }
}

async function handleCreate() {
  if (!createForm.value.name) {
    ElMessage.warning('请输入分支名称')
    return
  }
  try {
    await createBranch({
      repo_key: repoKey,
      name: createForm.value.name,
      base_ref: createForm.value.base_ref || undefined,
    })
    ElMessage.success('分支创建成功')
    showCreateDialog.value = false
    createForm.value = { name: '', base_ref: '' }
    await loadBranches()
  } catch { /* handled */ }
}

function openRenameDialog(branch: BranchInfo) {
  renameForm.value = { old_name: branch.name, new_name: branch.name }
  showRenameDialog.value = true
}

async function handleRename() {
  if (!renameForm.value.new_name) return
  try {
    await updateBranch(repoKey, renameForm.value.old_name, renameForm.value.new_name)
    ElMessage.success('重命名成功')
    showRenameDialog.value = false
    await loadBranches()
  } catch { /* handled */ }
}

async function handleDeleteBranch(name: string) {
  try {
    await ElMessageBox.confirm(`确定要删除分支 "${name}" 吗？`, '确认删除', { type: 'warning' })
    await deleteBranch(repoKey, name)
    ElMessage.success('分支已删除')
    await loadBranches()
  } catch { /* cancelled or handled */ }
}

function openTagDialog(branchName: string) {
  tagForm.value = { ref: branchName, name: '', message: '', push_remote: '', versionType: 'patch' }
  tagNextVersion.value = null
  showTagDialog.value = true
  getNextVersion(repoKey).then(info => {
    tagNextVersion.value = info
    handleTagVersionTypeChange('patch')
  }).catch(() => { /* ignore */ })
}

function handleTagVersionTypeChange(type: string | number | boolean) {
  if (!tagNextVersion.value) return
  switch (type) {
    case 'patch':
      tagForm.value.name = tagNextVersion.value.next_patch
      break
    case 'minor':
      tagForm.value.name = tagNextVersion.value.next_minor
      break
    case 'major':
      tagForm.value.name = tagNextVersion.value.next_major
      break
    case 'custom':
      tagForm.value.name = ''
      break
  }
}

async function handleCreateTag() {
  if (!tagForm.value.name) {
    ElMessage.warning('请输入标签名')
    return
  }
  try {
    await createTag({
      repo_key: repoKey,
      name: tagForm.value.name,
      ref: tagForm.value.ref,
      message: tagForm.value.message || undefined,
      push_remote: tagForm.value.push_remote || undefined,
    })
    ElMessage.success('标签创建成功')
    showTagDialog.value = false
  } catch { /* handled */ }
}

// 分支操作命令处理
function handleBranchCommand(command: string, row: BranchInfo) {
  switch (command) {
    case 'checkout':
      handleCheckout(row.name)
      break
    case 'push':
      handlePush(row.name)
      break
    case 'pull':
      handlePull(row.name)
      break
    case 'tag':
      openTagDialog(row.name)
      break
    case 'detail':
      goDetail(row.name)
      break
    case 'rename':
      openRenameDialog(row)
      break
    case 'delete':
      handleDeleteBranch(row.name)
      break
  }
}

// 远端分支操作命令处理
function handleRemoteBranchCommand(command: string, row: BranchInfo) {
  switch (command) {
    case 'checkout':
      handleCheckoutRemote(row.name)
      break
    case 'update':
      handleFfRemote(row.name)
      break
    case 'sync':
      handlePullRemote(row.name)
      break
  }
}
</script>

<style scoped>
.branch-list-page {
  padding: var(--spacing-xl);
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-height: 100vh;
  background: var(--bg-color);
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-left h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-color-secondary);
  background: none;
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-sm);
  padding: 6px 12px;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.back-btn:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.header-actions {
  display: flex;
  gap: 8px;
}

.action-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  padding: 8px 16px;
  border-radius: var(--border-radius-md);
  border: none;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-family);
}

.action-pill--green {
  background: #ECFDF5;
  color: var(--success-color);
}
.action-pill--green:hover {
  background: #D1FAE5;
}

.action-pill--primary {
  background: var(--primary-color);
  color: #FFFFFF;
}
.action-pill--primary:hover {
  background: var(--primary-color-hover);
}

.action-pill--outline {
  background: transparent;
  color: var(--text-color-primary);
  border: 1px solid var(--border-color);
}
.action-pill--outline:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.tab-bar {
  display: flex;
  border-bottom: 1px solid var(--border-color);
}

.tab-item {
  padding: 10px 20px;
  font-size: 14px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
  border-bottom: 2px solid transparent;
}

.tab-item:hover {
  color: var(--primary-color);
}

.tab-item.active {
  color: var(--primary-color);
  font-weight: 500;
  border-bottom-color: var(--primary-color);
}

.search-card {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
  padding: 12px 16px;
}

.search-icon {
  color: var(--text-color-secondary);
  font-size: 16px;
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  color: var(--text-color-primary);
  font-family: var(--font-family);
}

.search-input::placeholder {
  color: var(--text-color-placeholder);
}

.clear-icon {
  color: var(--text-color-secondary);
  font-size: 16px;
  cursor: pointer;
  flex-shrink: 0;
  transition: color var(--transition-fast);
}

.clear-icon:hover {
  color: var(--text-color-primary);
}

.branch-table-card {
  border-radius: var(--border-radius-lg);
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  overflow: hidden;
}

.table-header {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  background: var(--accent-bg);
}

.th {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color-secondary);
}

.table-row {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-color);
  transition: background var(--transition-fast);
}

.table-row:last-child {
  border-bottom: none;
}

.table-row:hover {
  background: var(--border-color-extra-light);
}

.td {
  font-size: 13px;
  color: var(--text-color-secondary);
}

.branch-name-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.branch-current {
  color: var(--success-color);
}

.current-icon {
  color: var(--success-color);
}

.td-mono {
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hash-text {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  color: var(--primary-color);
}

.commit-msg {
  margin-left: 6px;
  color: var(--text-color-secondary);
  font-size: 12px;
}

.author-name {
  display: block;
  color: var(--text-color-primary);
  font-size: 13px;
}

.author-email {
  display: block;
  font-size: 12px;
  color: var(--text-color-secondary);
}

.tag {
  display: inline-block;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--border-radius-sm);
}

.tag--success {
  background: #ECFDF5;
  color: var(--success-color);
}

.tag--info {
  background: var(--accent-bg);
  color: var(--primary-color);
}

.tag--warning {
  background: #FFFBEB;
  color: var(--warning-color);
}

.text-muted {
  font-size: 12px;
  color: var(--text-color-placeholder);
}

.row-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border-radius: var(--border-radius-sm);
  border: 1px solid var(--border-color);
  background: transparent;
  font-size: 13px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.row-action-btn:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.row-action-btn--primary {
  background: var(--primary-color);
  color: #FFFFFF;
  border: none;
}

.row-action-btn--primary:hover {
  background: var(--primary-color-hover);
}

.empty-table {
  padding: 32px;
  text-align: center;
}

.table-footer {
  padding: 8px 0;
  text-align: left;
}

.pag-info {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.form-tip {
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
  margin-top: var(--spacing-xs);
}

@media (max-width: 768px) {
  .branch-list-page {
    padding: var(--spacing-md);
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .branch-table-card {
    overflow-x: auto;
  }

  .table-header,
  .table-row {
    min-width: 900px;
  }
}
</style>
