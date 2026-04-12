<template>
  <div class="compare-page">
    <div class="page-header">
      <div class="header-left">
        <button class="back-btn" @click="$router.push(`/repos/${repoKey}/branches`)">
          <el-icon><ArrowLeft /></el-icon> 返回
        </button>
        <h2>分支对比 & 合并</h2>
      </div>
    </div>

    <div class="control-panel">
      <div class="branch-select-group">
        <div class="branch-box">
          <div class="branch-label">源分支 (Source/Feature)</div>
          <el-select v-model="sourceBranch" placeholder="选择源分支" filterable>
            <el-option-group label="本地分支">
              <el-option v-for="b in localBranches" :key="b" :label="b" :value="b" />
            </el-option-group>
            <el-option-group label="远程分支">
              <el-option v-for="b in remoteBranches" :key="b" :label="b" :value="b" />
            </el-option-group>
          </el-select>
        </div>
        <div class="arrow-box">
          <el-icon :size="20" color="#64748B"><Right /></el-icon>
        </div>
        <div class="branch-box">
          <div class="branch-label">目标分支 (Target/Base)</div>
          <el-select v-model="targetBranch" placeholder="选择目标分支" filterable>
            <el-option v-for="b in localBranches" :key="b" :label="b" :value="b" />
          </el-select>
        </div>
      </div>
      <div class="control-actions">
        <button class="action-pill action-pill--primary" @click="handleCompare" :disabled="comparing">
          <el-icon><Switch /></el-icon> 对比
        </button>
        <button class="action-pill action-pill--green" @click="openMergeDialog" :disabled="!compareResult || !canMerge">
          <el-icon><Connection /></el-icon> 合并
        </button>
      </div>
    </div>

    <el-alert
      v-if="targetBranch && isRemoteBranch(targetBranch)"
      title="目标分支不能是远程分支"
      type="warning"
      :closable="false"
      show-icon
      description="Git 合并只能在本地分支上执行，请选择本地分支作为目标分支。"
    />

    <div v-if="compareResult" class="stats-row">
      <div class="stat-card">
        <span class="stat-value">{{ compareResult.stat.FilesChanged }}</span>
        <span class="stat-label">变更文件</span>
      </div>
      <div class="stat-card">
        <span class="stat-value stat-value--green">+{{ compareResult.stat.Insertions }}</span>
        <span class="stat-label">新增行数</span>
      </div>
      <div class="stat-card">
        <span class="stat-value stat-value--red">-{{ compareResult.stat.Deletions }}</span>
        <span class="stat-label">删除行数</span>
      </div>
      <div class="stat-card stat-card--action">
        <button class="action-pill action-pill--outline" @click="handleDownloadPatch">
          <el-icon><Download /></el-icon> 导出 Patch
        </button>
      </div>
    </div>

    <template v-if="compareResult">
      <h3 class="section-title">变更文件列表</h3>

      <div class="file-table-card">
        <div class="table-header">
          <span class="th" style="width:80px">状态</span>
          <span class="th" style="flex:1">文件路径</span>
          <span class="th" style="width:120px">变更</span>
        </div>
        <div
          v-for="f in compareResult.files"
          :key="f.path"
          class="table-row"
          :class="{ active: f.path === currentFile }"
          @click="selectFile(f.path)"
        >
          <span class="td" style="width:80px">
            <span class="status-tag" :class="`status-tag--${getFileStatusClass(f.status)}`">{{ f.status }}</span>
          </span>
          <span class="td td-path" style="flex:1">{{ f.path }}</span>
          <span class="td" style="width:120px">
            <span class="status-tag" :class="`status-tag--${getFileStatusClass(f.status)}`">{{ f.status === 'A' ? 'Added' : f.status === 'D' ? 'Deleted' : f.status === 'M' ? 'Modified' : f.status }}</span>
          </span>
        </div>
      </div>

      <div v-if="currentFile" class="diff-section">
        <div class="diff-header">
          <span class="diff-title">{{ currentFile }}</span>
          <el-radio-group v-model="diffViewMode" size="small">
            <el-radio-button value="line-by-line">Line</el-radio-button>
            <el-radio-button value="side-by-side">Side</el-radio-button>
          </el-radio-group>
        </div>
        <div id="diff-viewer" v-html="diffHtml" class="diff-content"></div>
      </div>
    </template>

    <div v-if="!compareResult && !comparing" class="empty-state">
      <span class="text-muted">请选择分支进行对比</span>
    </div>

    <!-- Merge Dialog -->
    <el-dialog v-model="showMergeDialog" title="合并分支" width="550px" destroy-on-close>
      <p>
        即将合并 <strong>{{ sourceBranch }}</strong> 到 <strong>{{ targetBranch }}</strong>
      </p>

      <div v-if="mergeChecking" class="mb-3">
        <el-icon class="is-loading"><Loading /></el-icon> 正在检测冲突...
      </div>

      <div v-if="mergeCheckResult && !mergeChecking">
        <el-alert
          v-if="mergeCheckResult.success"
          title="可以自动合并"
          type="success"
          :closable="false"
          show-icon
          class="mb-3"
        />
        <el-alert
          v-else
          title="检测到冲突"
          type="error"
          :closable="false"
          show-icon
          class="mb-3"
        >
          <p>无法自动合并。以下文件存在冲突：</p>
          <ul>
            <li v-for="c in mergeCheckResult.conflicts" :key="c">{{ c }}</li>
          </ul>
        </el-alert>
      </div>

      <el-form v-if="mergeCheckResult?.success" :model="mergeForm" label-width="100px">
        <el-form-item label="合并信息">
          <el-input v-model="mergeForm.message" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showMergeDialog = false">取消</el-button>
        <el-button type="success" @click="handleMerge" :disabled="!mergeCheckResult?.success" :loading="merging">
          确认合并
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Right, Switch, Connection, Download, Loading } from '@element-plus/icons-vue'
import { getBranchList, compareBranches, getBranchDiff, getBranchPatch, checkMerge, mergeBranch } from '@/api/modules/branch'
import type { MergeCheckResult, BranchInfo } from '@/types/branch'
import * as Diff2Html from 'diff2html'
import 'diff2html/bundles/css/diff2html.min.css'

const route = useRoute()
const repoKey = route.params.repoKey as string

const allBranches = ref<BranchInfo[]>([])
const sourceBranch = ref('')
const targetBranch = ref('')
const comparing = ref(false)
const compareResult = ref<{ stat: { FilesChanged: number; Insertions: number; Deletions: number }; files: { path: string; status: string }[] } | null>(null)

const currentFile = ref('')
const diffHtml = ref('')
const diffViewMode = ref<'line-by-line' | 'side-by-side'>('line-by-line')

const showMergeDialog = ref(false)
const mergeChecking = ref(false)
const mergeCheckResult = ref<MergeCheckResult | null>(null)
const merging = ref(false)
const mergeForm = ref({ message: '' })

// 分离本地和远程分支
const localBranches = computed(() => 
  allBranches.value.filter(b => b.type === 'local').map(b => b.name)
)
const remoteBranches = computed(() => 
  allBranches.value.filter(b => b.type === 'remote').map(b => b.name)
)

// 判断是否是远程分支
function isRemoteBranch(name: string): boolean {
  return remoteBranches.value.includes(name)
}

// 是否可以合并：目标分支必须是本地分支
const canMerge = computed(() => {
  return targetBranch.value && !isRemoteBranch(targetBranch.value)
})

onMounted(async () => {
  try {
    const res = await getBranchList(repoKey, { page_size: 1000 })
    allBranches.value = res.list || []
  } catch { /* ignore */ }
})

watch(diffViewMode, () => {
  if (currentFile.value) selectFile(currentFile.value)
})

function getFileStatusClass(status: string): string {
  if (status === 'A') return 'added'
  if (status === 'D') return 'deleted'
  if (status === 'M') return 'modified'
  if (status === 'R') return 'renamed'
  return ''
}

async function handleCompare() {
  if (!sourceBranch.value || !targetBranch.value) {
    ElMessage.warning('请选择源分支和目标分支')
    return
  }
  comparing.value = true
  compareResult.value = null
  currentFile.value = ''
  diffHtml.value = ''
  try {
    compareResult.value = await compareBranches(repoKey, sourceBranch.value, targetBranch.value)
  } finally {
    comparing.value = false
  }
}

async function selectFile(path: string) {
  currentFile.value = path
  try {
    const res = await getBranchDiff(repoKey, sourceBranch.value, targetBranch.value, path)
    diffHtml.value = Diff2Html.html(res.diff || '', {
      drawFileList: false,
      matching: 'lines',
      outputFormat: diffViewMode.value,
    })
  } catch {
    diffHtml.value = '<p>加载差异失败</p>'
  }
}

async function handleDownloadPatch() {
  try {
    const response = await getBranchPatch(repoKey, sourceBranch.value, targetBranch.value)
    const blob = response.data instanceof Blob ? response.data : new Blob([response.data], { type: 'application/octet-stream' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${sourceBranch.value}-to-${targetBranch.value}.patch`
    a.click()
    window.URL.revokeObjectURL(url)
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('导出 Patch 失败: ' + (err.message || '未知错误'))
  }
}

async function openMergeDialog() {
  showMergeDialog.value = true
  mergeChecking.value = true
  mergeCheckResult.value = null
  mergeForm.value.message = `Merge ${sourceBranch.value} into ${targetBranch.value}`
  try {
    mergeCheckResult.value = await checkMerge(repoKey, sourceBranch.value, targetBranch.value)
  } finally {
    mergeChecking.value = false
  }
}

async function handleMerge() {
  merging.value = true
  try {
    await mergeBranch({
      repo_key: repoKey,
      source: sourceBranch.value,
      target: targetBranch.value,
      message: mergeForm.value.message,
    })
    ElMessage.success('合并成功')
    showMergeDialog.value = false
    await handleCompare()
  } finally {
    merging.value = false
  }
}
</script>

<style scoped>
.compare-page {
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

.control-panel {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
}

.branch-select-group {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}

.branch-box {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}

.branch-label {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.branch-box :deep(.el-select) {
  width: 100%;
}

.arrow-box {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 8px;
}

.control-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.action-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  padding: 10px 20px;
  border-radius: var(--border-radius-md);
  border: none;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-family);
}

.action-pill:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-pill--primary {
  background: var(--primary-color);
  color: #FFFFFF;
}
.action-pill--primary:hover:not(:disabled) {
  background: var(--primary-color-hover);
}

.action-pill--green {
  background: #ECFDF5;
  color: var(--success-color);
}
.action-pill--green:hover:not(:disabled) {
  background: #D1FAE5;
}

.action-pill--outline {
  background: transparent;
  color: var(--text-color-primary);
  border: 1px solid var(--border-color);
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
  padding: 16px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
}

.stat-card--action {
  justify-content: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.stat-value--green {
  color: var(--success-color);
}

.stat-value--red {
  color: var(--danger-color);
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

.file-table-card {
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
  cursor: pointer;
  transition: background var(--transition-fast);
}

.table-row:last-child {
  border-bottom: none;
}

.table-row:hover {
  background: var(--border-color-extra-light);
}

.table-row.active {
  background: var(--accent-bg);
}

.td {
  font-size: 13px;
  color: var(--text-color-secondary);
}

.status-tag {
  display: inline-block;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--border-radius-sm);
}

.status-tag--added {
  background: #ECFDF5;
  color: var(--success-color);
}

.status-tag--deleted {
  background: #FEF2F2;
  color: var(--danger-color);
}

.status-tag--modified {
  background: #EEF2FF;
  color: var(--primary-color);
}

.status-tag--renamed {
  background: #FFFBEB;
  color: var(--warning-color);
}

.td-path {
  color: var(--text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.td-mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 13px;
}

.stat-add {
  color: var(--success-color);
}

.stat-del {
  color: var(--danger-color);
}

.diff-section {
  border-radius: var(--border-radius-lg);
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  overflow: hidden;
}

.diff-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-color);
}

.diff-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.diff-content {
  overflow-x: auto;
  padding: 0;
}

.empty-state {
  padding: 40px;
  text-align: center;
}

.text-muted {
  font-size: 13px;
  color: var(--text-color-placeholder);
}

@media (max-width: 768px) {
  .compare-page {
    padding: var(--spacing-md);
  }

  .control-panel {
    flex-direction: column;
    align-items: stretch;
  }

  .branch-select-group {
    flex-direction: column;
  }

  .arrow-box {
    transform: rotate(90deg);
  }

  .stats-row {
    flex-wrap: wrap;
  }

  .stat-card {
    min-width: calc(50% - 12px);
  }
}
</style>
