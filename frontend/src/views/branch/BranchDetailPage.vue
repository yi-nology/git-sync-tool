<template>
  <div class="branch-detail-page" v-loading="loading">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="$router.push(`/repos/${repoKey}/branches`)" :icon="ArrowLeft" text>返回</el-button>
        <h2>{{ branchName }}</h2>
        <el-tag v-if="isCurrent" type="success" size="small">当前分支</el-tag>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="$router.push(`/repos/${repoKey}/compare`)">
          <el-icon><Switch /></el-icon> 对比/合并
        </el-button>
        <el-button type="success" @click="handlePush">
          <el-icon><Top /></el-icon> 推送远端
        </el-button>
        <el-button v-if="hasUncommitted" type="warning" @click="showSubmitDialog = true">
          <el-icon><Upload /></el-icon> 提交变更
        </el-button>
        <el-button @click="loadData">
          <el-icon><Refresh /></el-icon> 刷新
        </el-button>
        <el-button v-if="!isCurrent" type="danger" @click="handleDelete">
          <el-icon><Delete /></el-icon> 删除分支
        </el-button>
      </div>
    </div>

    <!-- Uncommitted changes alert -->
    <el-alert
      v-if="hasUncommitted"
      title="检测到未提交的变更"
      type="warning"
      :closable="false"
      show-icon
      class="mb-4"
    >
      <pre class="status-text">{{ repoStatus }}</pre>
    </el-alert>

    <!-- Stats Cards -->
    <el-row :gutter="16" class="mb-4">
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="总代码行数" :value="statsData?.total_lines || 0" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="提交总数" :value="commits.length" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="贡献者数" :value="statsData?.authors?.length || 0" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="文件类型" :value="fileTypeCount" />
        </el-card>
      </el-col>
    </el-row>

    <!-- Main Content: Commits + Contributors -->
    <el-row :gutter="16">
      <el-col :span="16">
        <el-card header="最近提交">
          <div v-if="commits.length === 0">
            <el-empty description="暂无提交记录" />
          </div>
          <div v-else class="commit-list">
            <div v-for="c in commits" :key="c.hash" class="commit-item">
              <div class="commit-main">
                <el-text class="commit-message" truncated>{{ c.message }}</el-text>
                <div class="commit-meta">
                  <el-text type="info" size="small">
                    <el-icon><User /></el-icon> {{ c.author }}
                  </el-text>
                  <el-text type="info" size="small">{{ formatRelativeTime(c.date) }}</el-text>
                  <el-text class="mono-text" size="small" type="info">{{ c.hash?.substring(0, 8) }}</el-text>
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card header="贡献者排行">
          <div v-if="!statsData?.authors?.length">
            <el-empty description="暂无贡献者数据" />
          </div>
          <div v-else class="author-list">
            <div v-for="a in statsData.authors" :key="a.email" class="author-item">
              <div class="author-info">
                <el-text class="author-name">{{ a.name }}</el-text>
                <el-text type="info" size="small">{{ a.email }}</el-text>
              </div>
              <el-tag size="small" type="success">{{ a.total_lines }} 行</el-tag>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

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
import { ArrowLeft, Switch, Top, Upload, Refresh, Delete, User } from '@element-plus/icons-vue'
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
  flex-wrap: wrap;
}
.mb-4 {
  margin-bottom: 16px;
}
.commit-list {
  max-height: 600px;
  overflow-y: auto;
}
.commit-item {
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}
.commit-item:last-child {
  border-bottom: none;
}
.commit-main {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.commit-message {
  font-weight: 500;
  font-size: 14px;
}
.commit-meta {
  display: flex;
  gap: 16px;
  align-items: center;
}
.mono-text {
  font-family: monospace;
}
.author-list {
  max-height: 500px;
  overflow-y: auto;
}
.author-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}
.author-item:last-child {
  border-bottom: none;
}
.author-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.author-name {
  font-weight: 500;
}
.status-text {
  margin: 0;
  white-space: pre-wrap;
  font-family: monospace;
  font-size: 12px;
  max-height: 200px;
  overflow-y: auto;
}
</style>
