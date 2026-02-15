<template>
  <div class="compare-page">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="$router.push(`/repos/${repoKey}/branches`)" :icon="ArrowLeft" text>返回</el-button>
        <h2>分支对比 & 合并</h2>
      </div>
    </div>

    <!-- Control Panel -->
    <el-card class="mb-4">
      <el-row :gutter="16" align="middle">
        <el-col :span="8">
          <div class="form-label">源分支 (Source/Feature)</div>
          <el-select v-model="sourceBranch" placeholder="选择源分支" filterable style="width: 100%">
            <el-option-group label="本地分支">
              <el-option v-for="b in localBranches" :key="b" :label="b" :value="b" />
            </el-option-group>
            <el-option-group label="远程分支">
              <el-option v-for="b in remoteBranches" :key="b" :label="b" :value="b" />
            </el-option-group>
          </el-select>
        </el-col>
        <el-col :span="2" class="text-center">
          <el-icon :size="24" color="#909399"><Right /></el-icon>
        </el-col>
        <el-col :span="8">
          <div class="form-label">目标分支 (Target/Base) <el-text type="info" size="small">仅本地分支</el-text></div>
          <el-select v-model="targetBranch" placeholder="选择目标分支" filterable style="width: 100%">
            <el-option v-for="b in localBranches" :key="b" :label="b" :value="b" />
          </el-select>
        </el-col>
        <el-col :span="6" class="text-right">
          <el-button-group>
            <el-button type="primary" @click="handleCompare" :loading="comparing">
              <el-icon><Switch /></el-icon> 对比
            </el-button>
            <el-button type="success" @click="openMergeDialog" :disabled="!compareResult || !canMerge">
              <el-icon><Connection /></el-icon> 合并
            </el-button>
          </el-button-group>
        </el-col>
      </el-row>
      <!-- 合并提示 -->
      <el-alert
        v-if="targetBranch && isRemoteBranch(targetBranch)"
        title="目标分支不能是远程分支"
        type="warning"
        :closable="false"
        show-icon
        class="mt-3"
        description="Git 合并只能在本地分支上执行，请选择本地分支作为目标分支。"
      />
    </el-card>

    <!-- Summary Stats -->
    <el-row v-if="compareResult" :gutter="16" class="mb-4">
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="变更文件" :value="compareResult.stat.FilesChanged" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="新增行数" :value="compareResult.stat.Insertions" :value-style="{ color: '#67c23a' }" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="删除行数" :value="compareResult.stat.Deletions" :value-style="{ color: '#f56c6c' }" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="download-card">
          <el-button type="info" @click="handleDownloadPatch">
            <el-icon><Download /></el-icon> 导出 Patch
          </el-button>
        </el-card>
      </el-col>
    </el-row>

    <!-- File List & Diff Viewer -->
    <el-row v-if="compareResult" :gutter="16">
      <el-col :span="6">
        <el-card header="变更文件列表">
          <div class="file-list">
            <div
              v-for="f in compareResult.files"
              :key="f.path"
              class="file-item"
              :class="{ active: f.path === currentFile }"
              @click="selectFile(f.path)"
            >
              <el-text size="small" truncated>{{ f.path }}</el-text>
              <div class="file-stat">
                <el-tag size="small" :type="getFileStatusType(f.status)">{{ f.status }}</el-tag>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="18">
        <el-card>
          <template #header>
            <div class="diff-header">
              <el-text>{{ currentFile || '选择文件查看差异' }}</el-text>
              <el-radio-group v-model="diffViewMode" size="small">
                <el-radio-button value="line-by-line">Line</el-radio-button>
                <el-radio-button value="side-by-side">Side</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div id="diff-viewer" v-html="diffHtml" class="diff-content"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty v-if="!compareResult && !comparing" description="请选择分支进行对比" />

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

function getFileStatusType(status: string): '' | 'success' | 'warning' | 'danger' | 'info' {
  if (status === 'A') return 'success'
  if (status === 'D') return 'danger'
  if (status === 'M') return 'warning'
  if (status === 'R') return 'info'
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
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
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
.form-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}
.text-center {
  text-align: center;
  padding-top: 20px;
}
.text-right {
  text-align: right;
  padding-top: 20px;
}
.mb-3 {
  margin-bottom: 12px;
}
.mb-4 {
  margin-bottom: 16px;
}
.mt-3 {
  margin-top: 12px;
}
.file-list {
  max-height: 600px;
  overflow-y: auto;
}
.file-item {
  padding: 6px 8px;
  cursor: pointer;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.file-item:hover {
  background: #ecf5ff;
}
.file-item.active {
  background: #ecf5ff;
  border-left: 3px solid #409eff;
}
.file-stat {
  display: flex;
  gap: 6px;
  font-size: 12px;
  white-space: nowrap;
}
.stat-add {
  color: #67c23a;
}
.stat-del {
  color: #f56c6c;
}
.diff-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.diff-content {
  overflow-x: auto;
}
.download-card {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}
.download-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
