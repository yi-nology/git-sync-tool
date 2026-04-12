<template>
  <div class="branch-detail-page" v-loading="loading">
    <div class="page-header">
      <div class="header-left">
        <button class="back-btn" @click="$router.push(`/repos/${repoKey}/branches`)">
          <el-icon><ArrowLeft /></el-icon> 返回
        </button>
        <h2>{{ branchName }}</h2>
        <span v-if="isCurrent" class="current-tag">当前分支</span>
      </div>
      <div class="header-actions">
        <button class="action-pill action-pill--primary" @click="$router.push(`/repos/${repoKey}/compare`)">
          <el-icon><Switch /></el-icon> 对比/合并
        </button>
        <button class="action-pill action-pill--green" @click="handlePush">
          <el-icon><Top /></el-icon> 推送远端
        </button>
        <button v-if="hasUncommitted" class="action-pill action-pill--amber" @click="showSubmitDialog = true">
          <el-icon><Upload /></el-icon> 提交变更
        </button>
        <button class="action-pill action-pill--outline" @click="loadData">
          <el-icon><Refresh /></el-icon> 刷新
        </button>
        <button v-if="!isCurrent" class="action-pill action-pill--danger" @click="handleDelete">
          <el-icon><Delete /></el-icon> 删除分支
        </button>
      </div>
    </div>

    <div v-if="hasUncommitted" class="uncommitted-alert">
      <span class="alert-title">检测到未提交的变更</span>
      <pre class="status-text">{{ repoStatus }}</pre>
    </div>

    <div class="stats-row">
      <div class="stat-card">
        <span class="stat-value">{{ statsData?.total_lines || 0 }}</span>
        <span class="stat-label">总代码行数</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ commits.length }}</span>
        <span class="stat-label">提交总数</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ statsData?.authors?.length || 0 }}</span>
        <span class="stat-label">贡献者数</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ fileTypeCount }}</span>
        <span class="stat-label">文件类型</span>
      </div>
    </div>

    <h3 class="section-title">最近提交</h3>

    <div class="commit-table-card" v-if="commits.length > 0">
      <div class="table-header">
        <span class="th" style="width:100px">Hash</span>
        <span class="th" style="flex:1">信息</span>
        <span class="th" style="width:120px">作者</span>
        <span class="th" style="width:140px">时间</span>
      </div>
      <div v-for="c in commits" :key="c.hash" class="table-row">
        <span class="td" style="width:100px">
          <span class="hash-text">{{ c.hash?.substring(0, 8) }}</span>
        </span>
        <span class="td td-message" style="flex:1">{{ c.message }}</span>
        <span class="td" style="width:120px">
          <span class="author-name">{{ c.author }}</span>
        </span>
        <span class="td" style="width:140px">{{ formatRelativeTime(c.date) }}</span>
      </div>
    </div>
    <div v-else class="empty-table">
      <span class="text-muted">暂无提交记录</span>
    </div>

    <!-- Push Dialog -->
    <el-dialog v-model="showPushDialog" :title="`推送分支: ${branchName}`" width="480px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="目标远端">
          <el-checkbox-group v-model="pushRemotes">
            <el-checkbox v-for="r in remoteNames" :key="r" :label="r" :value="r" />
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPushDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitPush">确认推送</el-button>
      </template>
    </el-dialog>

    <!-- Submit Changes Dialog -->
    <el-dialog v-model="showSubmitDialog" title="提交变更" width="550px" destroy-on-close>
      <el-form :model="submitForm" label-width="110px">
        <el-form-item label="Author Name">
          <el-input v-model="submitForm.author_name" />
        </el-form-item>
        <el-form-item label="Author Email">
          <el-input v-model="submitForm.author_email" />
        </el-form-item>
        <el-form-item label="Commit 信息" required>
          <el-input v-model="submitForm.message" type="textarea" :rows="3" placeholder="请输入提交信息" />
        </el-form-item>
        <el-form-item label="提交后推送">
          <el-switch v-model="submitForm.push" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSubmitDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitChanges" :loading="submitting">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Switch, Top, Upload, Refresh, Delete } from '@element-plus/icons-vue'
import { pushBranch, deleteBranch, getBranchList } from '@/api/modules/branch'
import { getRepoDetail, scanRepo } from '@/api/modules/repo'
import { getStatsAnalyze, getStatsCommits } from '@/api/modules/stats'
import { getRepoStatus, getRepoGitConfig, submitChanges } from '@/api/modules/system'
import type { StatsResponse } from '@/types/stats'
import { formatRelativeTime } from '@/utils/format'

interface CommitInfo {
  hash: string
  message: string
  author: string
  date: string
}

const route = useRoute()
const router = useRouter()
const repoKey = route.params.repoKey as string
const branchName = route.params.branchName as string

const loading = ref(false)
const isCurrent = ref(false)
const statsData = ref<StatsResponse | null>(null)
const commits = ref<CommitInfo[]>([])
const remoteNames = ref<string[]>([])
const repoStatus = ref('')
const hasUncommitted = ref(false)

const showPushDialog = ref(false)
const pushRemotes = ref<string[]>([])

const showSubmitDialog = ref(false)
const submitting = ref(false)
const submitForm = ref({
  author_name: '',
  author_email: '',
  message: '',
  push: false,
})

const fileTypeCount = computed(() => {
  if (!statsData.value?.authors) return 0
  const types = new Set<string>()
  for (const a of statsData.value.authors) {
    if (a.file_types) {
      Object.keys(a.file_types).forEach((t) => types.add(t))
    }
  }
  return types.size
})

onMounted(() => loadData())

async function loadData() {
  loading.value = true
  try {
    // Check if this is the current branch
    try {
      const res = await getBranchList(repoKey, { type: 'local', page_size: 500 })
      const branch = (res.list || []).find((b) => b.name === branchName)
      isCurrent.value = branch?.is_current || false
    } catch { /* ignore */ }

    // Load stats
    try {
      statsData.value = await getStatsAnalyze(repoKey, { branch: branchName })
    } catch { /* ignore */ }

    // Load commits
    try {
      const res = await getStatsCommits(repoKey, { branch: branchName })
      commits.value = (Array.isArray(res) ? res : []).slice(0, 20)
    } catch { /* ignore */ }

    // Load remotes
    try {
      const repo = await getRepoDetail(repoKey)
      if (repo?.path) {
        const scan = await scanRepo(repo.path)
        remoteNames.value = (scan.remotes || []).map((r: { name: string }) => r.name)
      }
    } catch { /* ignore */ }

    // Check uncommitted changes
    try {
      const status = await getRepoStatus(repoKey) as unknown as { status: string }
      repoStatus.value = status?.status || ''
      hasUncommitted.value = !!repoStatus.value && repoStatus.value.trim() !== ''
    } catch { /* ignore */ }

    // Load git config for submit form
    try {
      const config = await getRepoGitConfig(repoKey) as unknown as { name: string; email: string }
      submitForm.value.author_name = config?.name || ''
      submitForm.value.author_email = config?.email || ''
    } catch { /* ignore */ }
  } finally {
    loading.value = false
  }
}

function handlePush() {
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
    await pushBranch(repoKey, branchName, pushRemotes.value)
    ElMessage.success('推送成功')
    showPushDialog.value = false
  } catch { /* handled */ }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`确定要删除分支 "${branchName}" 吗？`, '确认删除', { type: 'warning' })
    await deleteBranch(repoKey, branchName)
    ElMessage.success('分支已删除')
    router.push(`/repos/${repoKey}/branches`)
  } catch { /* cancelled */ }
}

async function handleSubmitChanges() {
  if (!submitForm.value.message) {
    ElMessage.warning('请输入提交信息')
    return
  }
  submitting.value = true
  try {
    await submitChanges({
      repo_key: repoKey,
      message: submitForm.value.message,
      push: submitForm.value.push,
      author_name: submitForm.value.author_name || undefined,
      author_email: submitForm.value.author_email || undefined,
    })
    ElMessage.success('提交成功')
    showSubmitDialog.value = false
    await loadData()
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.branch-detail-page {
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

.current-tag {
  display: inline-flex;
  align-items: center;
  font-size: 11px;
  color: var(--success-color);
  background: #ECFDF5;
  padding: 2px 8px;
  border-radius: var(--border-radius-sm);
}

.header-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
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

.action-pill--primary {
  background: var(--primary-color);
  color: #FFFFFF;
}
.action-pill--primary:hover {
  background: var(--primary-color-hover);
}

.action-pill--green {
  background: #ECFDF5;
  color: var(--success-color);
}
.action-pill--green:hover {
  background: #D1FAE5;
}

.action-pill--amber {
  background: #FFFBEB;
  color: var(--warning-color);
}
.action-pill--amber:hover {
  background: #FEF3C7;
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

.action-pill--danger {
  background: #FEF2F2;
  color: var(--danger-color);
}
.action-pill--danger:hover {
  background: #FEE2E2;
}

.uncommitted-alert {
  background: #FFFBEB;
  border: 1px solid var(--warning-color);
  border-radius: var(--border-radius-lg);
  padding: 12px 16px;
}

.alert-title {
  font-weight: 600;
  font-size: 13px;
  color: var(--warning-color);
  display: block;
  margin-bottom: 8px;
}

.status-text {
  margin: 0;
  white-space: pre-wrap;
  font-family: monospace;
  font-size: 12px;
  max-height: 200px;
  overflow-y: auto;
  color: var(--text-color-secondary);
}

.stats-row {
  display: flex;
  gap: 16px;
}

.stat-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 20px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.stat-label {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.section-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.commit-table-card {
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

.hash-text {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
  color: var(--primary-color);
}

.td-message {
  color: var(--text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.author-name {
  color: var(--text-color-primary);
  font-size: 13px;
}

.empty-table {
  padding: 40px;
  text-align: center;
}

.text-muted {
  font-size: 13px;
  color: var(--text-color-placeholder);
}

@media (max-width: 768px) {
  .branch-detail-page {
    padding: var(--spacing-md);
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .stats-row {
    flex-wrap: wrap;
  }

  .stat-card {
    min-width: calc(50% - 12px);
  }

  .commit-table-card {
    overflow-x: auto;
  }

  .table-header,
  .table-row {
    min-width: 600px;
  }
}
</style>
