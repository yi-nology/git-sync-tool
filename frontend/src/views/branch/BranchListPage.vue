<template>
  <div class="branch-list-page">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="$router.push(`/repos/${repoKey}`)" :icon="ArrowLeft" text>返回</el-button>
        <h2>分支管理</h2>
      </div>
      <div class="header-actions">
        <el-button type="success" @click="$router.push(`/repos/${repoKey}/compare`)">
          <el-icon><Switch /></el-icon> 分支对比 & 合并
        </el-button>
        <el-button @click="handleFetchAll" :loading="fetchLoading">
          <el-icon><Download /></el-icon> 刷新远端 (Fetch)
        </el-button>
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon> 新建分支
        </el-button>
      </div>
    </div>

    <el-tabs v-model="branchType" @tab-change="loadBranches">
      <el-tab-pane label="Local (本地分支)" name="local" />
      <el-tab-pane label="Remote (远端分支)" name="remote" />
    </el-tabs>

    <el-card class="mb-3">
      <el-row :gutter="16">
        <el-col :span="8">
          <el-input v-model="searchQuery" placeholder="搜索分支名/作者..." @keyup.enter="loadBranches" clearable />
        </el-col>
        <el-col :span="4">
          <el-button @click="loadBranches">搜索</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-card>
      <el-table :data="branches" v-loading="loading" stripe>
        <el-table-column prop="name" label="分支名称" min-width="180">
          <template #default="{ row }">
            <span :class="{ 'branch-current': row.is_current }">
              <el-icon v-if="row.is_current" color="#67c23a"><CircleCheck /></el-icon>
              {{ row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="hash" label="最新提交" min-width="200">
          <template #default="{ row }">
            <el-text class="mono-text" size="small">{{ row.hash ? row.hash.substring(0, 8) : '-' }}</el-text>
            <el-text v-if="row.message" size="small" type="info" class="commit-msg" truncated> {{ row.message }}</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="author" label="提交人" width="140">
          <template #default="{ row }">
            <div>{{ row.author }}</div>
            <el-text v-if="row.author_email" size="small" type="info">{{ row.author_email }}</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="date" label="提交时间" width="160">
          <template #default="{ row }">
            {{ formatRelativeTime(row.date) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="140" v-if="branchType === 'local'">
          <template #default="{ row }">
            <template v-if="row.upstream">
              <el-tag v-if="row.ahead > 0" size="small" type="success">{{ row.ahead }}↑</el-tag>
              <el-tag v-if="row.behind > 0" size="small" type="warning">{{ row.behind }}↓</el-tag>
              <el-tag v-if="row.ahead === 0 && row.behind === 0" size="small" type="info">已同步</el-tag>
            </template>
            <el-text v-else type="info" size="small">无上游</el-text>
          </template>
        </el-table-column>
        <el-table-column label="关联" width="160" v-if="branchType === 'remote'">
          <template #default="{ row }">
            <el-tag v-if="getLocalBranch(row.name)" size="small" type="success">
              已关联: {{ getLocalBranch(row.name) }}
            </el-tag>
            <el-text v-else type="info" size="small">无本地关联</el-text>
          </template>
        </el-table-column>
        <el-table-column label="操作" :width="branchType === 'local' ? 340 : 280" fixed="right">
          <template #default="{ row }">
            <el-button-group size="small" v-if="branchType === 'local'">
              <el-button v-if="!row.is_current" @click="handleCheckout(row.name)">
                <el-icon><Select /></el-icon> 切换
              </el-button>
              <el-button @click="handlePush(row.name)">
                <el-icon><Top /></el-icon> 推送
              </el-button>
              <el-button @click="handlePull(row.name)">
                <el-icon><Bottom /></el-icon> 拉取
              </el-button>
              <el-button @click="openTagDialog(row.name)" type="warning">
                <el-icon><PriceTag /></el-icon>
              </el-button>
              <el-button type="primary" @click="goDetail(row.name)">
                <el-icon><View /></el-icon> 详情
              </el-button>
              <el-button @click="openRenameDialog(row)">
                <el-icon><Edit /></el-icon>
              </el-button>
              <el-button v-if="!row.is_current" type="danger" @click="handleDeleteBranch(row.name)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </el-button-group>
            <el-button-group size="small" v-else>
              <el-button type="primary" @click="handleCheckoutRemote(row.name)" v-if="!getLocalBranch(row.name)">
                <el-icon><Download /></el-icon> 检出为本地
              </el-button>
              <el-button @click="handleFfRemote(row.name)" v-if="getLocalBranch(row.name)">
                <el-icon><Bottom /></el-icon> 更新本地
              </el-button>
              <el-button @click="handlePullRemote(row.name)" v-if="getLocalBranch(row.name)">
                <el-icon><RefreshRight /></el-icon> 同步本地
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
      <div class="table-footer">
        <el-text type="info" size="small">共 {{ total }} 个分支</el-text>
      </div>
    </el-card>

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
import { ArrowLeft, Plus, Delete, Edit, Select, Top, Bottom, Switch, Download, CircleCheck, PriceTag, View, RefreshRight } from '@element-plus/icons-vue'
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
const branchType = ref('local')
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
  await loadBranches()
  // Also load local branches for remote association
  try {
    const res = await getBranchList(repoKey, { type: 'local', page_size: 500 })
    localBranches.value = res.list || []
  } catch { /* ignore */ }
  try {
    const repo = await getRepoDetail(repoKey)
    if (repo?.path) {
      const scan = await scanRepo(repo.path)
      remoteNames.value = (scan.remotes || []).map((r: { name: string }) => r.name)
    }
  } catch { /* ignore */ }
})

async function loadBranches() {
  loading.value = true
  try {
    const res = await getBranchList(repoKey, {
      type: branchType.value,
      keyword: searchQuery.value || undefined,
      page_size: 500,
    })
    branches.value = res.list || []
    total.value = res.total || 0
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
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.header-left h2 {
  margin: 0;
  font-size: 20px;
}
.header-actions {
  display: flex;
  gap: 8px;
}
.mb-3 {
  margin-bottom: 12px;
}
.branch-current {
  font-weight: bold;
  color: #67c23a;
}
.mono-text {
  font-family: monospace;
}
.commit-msg {
  margin-left: 8px;
}
.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
.table-footer {
  padding: 12px 0;
  text-align: left;
}
</style>
