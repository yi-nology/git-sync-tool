<template>
  <div class="sync-page">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <el-button @click="$router.push(`/repos/${repoKey}`)" :icon="ArrowLeft" text>返回</el-button>
        <h2>Git 同步管理</h2>
      </div>
      <div class="header-right">
        <el-button type="success" @click="handleBatchSync" :disabled="selectedTasks.length === 0">
          <el-icon><Refresh /></el-icon> 同步选中 ({{ selectedTasks.length }})
        </el-button>
        <el-button type="primary" @click="openQuickSync">
          <el-icon><Plus /></el-icon> 快速同步
        </el-button>
        <el-button @click="openAddTask">
          <el-icon><Setting /></el-icon> 新建规则
        </el-button>
      </div>
    </div>

    <!-- Quick Sync Panel -->
    <el-card v-if="showQuickPanel" class="quick-panel">
      <template #header>
        <div class="panel-header">
          <span>⚡ 快速同步</span>
          <el-button text @click="showQuickPanel = false"><el-icon><Close /></el-icon></el-button>
        </div>
      </template>
      <el-form :model="quickForm" inline class="quick-form">
        <el-form-item label="源">
          <el-select v-model="quickForm.sourceRemote" style="width: 100px">
            <el-option label="Local" value="local" />
            <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
          </el-select>
          <el-input v-model="quickForm.sourceBranch" placeholder="分支" style="width: 100px; margin-left: 8px" />
        </el-form-item>
        
        <el-icon class="arrow-icon"><Right /></el-icon>
        
        <el-form-item label="目标">
          <el-select v-model="quickForm.targetRemote" style="width: 100px">
            <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
          </el-select>
          <el-input v-model="quickForm.targetBranch" placeholder="分支" style="width: 100px; margin-left: 8px" />
        </el-form-item>

        <el-form-item>
          <el-checkbox v-model="quickForm.gitTags">--tags</el-checkbox>
          <el-checkbox v-model="quickForm.gitForce">--force</el-checkbox>
          <el-checkbox v-model="quickForm.gitPrune">--prune</el-checkbox>
        </el-form-item>

        <el-form-item>
          <el-button @click="handlePreview" :loading="previewing">预览</el-button>
          <el-button type="primary" @click="handleQuickSync" :loading="syncing">执行</el-button>
        </el-form-item>
      </el-form>
      
      <el-alert v-if="previewResult" :title="previewResult.command" type="info" :closable="false" class="preview-result">
        <div v-if="previewResult.commits_to_push">
          <strong>Commits:</strong> {{ Array.isArray(previewResult.commits_to_push) ? previewResult.commits_to_push.length : previewResult.commits_to_push }}
        </div>
        <div v-if="previewResult.tags_to_push">
          <strong>Tags:</strong> {{ Array.isArray(previewResult.tags_to_push) ? previewResult.tags_to_push.length : previewResult.tags_to_push }}
        </div>
        <div v-if="previewResult.warning" style="color: #e6a23c">{{ previewResult.warning }}</div>
      </el-alert>
    </el-card>

    <!-- Task List -->
    <div v-loading="loading" class="task-list">
      <el-empty v-if="tasks.length === 0 && !loading" description="暂无同步规则">
        <el-button type="primary" @click="openAddTask">创建第一条规则</el-button>
      </el-empty>

      <el-card v-for="task in tasks" :key="task.key" class="task-card" :class="{ disabled: !task.enabled }">
        <div class="task-content">
          <!-- Checkbox -->
          <el-checkbox v-model="selectedTasks" :value="task.key" class="task-checkbox" />
          
          <!-- Direction -->
          <div class="direction-flow">
            <div class="endpoint source">
              <span class="label">{{ task.source_remote }}</span>
              <span class="branch">{{ task.source_branch }}</span>
            </div>
            <div class="flow-arrow">
              <el-icon><Right /></el-icon>
              <el-tag v-if="task.sync_mode === 'all-branch'" size="small" type="info">全分支</el-tag>
            </div>
            <div class="endpoint target">
              <span class="label">{{ task.target_remote }}</span>
              <span class="branch">{{ task.target_branch }}</span>
            </div>
          </div>

          <!-- Status -->
          <div class="task-status">
            <el-tag :type="task.enabled ? 'success' : 'info'" size="small">
              {{ task.enabled ? '✅ 已启用' : '⏸️ 已暂停' }}
            </el-tag>
            <span v-if="task.cron" class="cron">
              <el-icon><AlarmClock /></el-icon> {{ task.cron }}
            </span>
          </div>

          <!-- Git Options -->
          <div class="git-options">
            <el-tag v-if="task.git_tags" size="small" effect="plain">--tags</el-tag>
            <el-tag v-if="task.git_force" size="small" type="warning" effect="plain">--force</el-tag>
            <el-tag v-if="task.git_prune" size="small" effect="plain">--prune</el-tag>
            <el-tag v-if="task.git_no_verify" size="small" effect="plain">--no-verify</el-tag>
          </div>

          <!-- Actions -->
          <div class="task-actions">
            <el-button size="small" type="success" @click="handleRun(task.key)" :icon="CaretRight" round>执行</el-button>
            <el-button size="small" @click="showHistory(task.key)" :icon="Clock" round>历史</el-button>
            <el-dropdown trigger="click">
              <el-button size="small" :icon="More" round />
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="openEditTask(task)">编辑</el-dropdown-item>
                  <el-dropdown-item @click="toggleEnabled(task)">
                    {{ task.enabled ? '暂停' : '启用' }}
                  </el-dropdown-item>
                  <el-dropdown-item divided @click="handleDelete(task.key)" style="color: #f56c6c">删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </el-card>
    </div>

    <!-- Task Dialog -->
    <el-dialog v-model="showTaskDialog" :title="editingTask ? '编辑同步规则' : '新建同步规则'" width="700px" destroy-on-close>
      <el-form :model="taskForm" label-width="100px">
        <el-form-item label="同步模式">
          <el-radio-group v-model="taskForm.sync_mode">
            <el-radio value="single">单分支同步</el-radio>
            <el-radio value="all-branch">全分支同步</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="源 (Source)">
              <el-select v-model="taskForm.source_remote" style="width: 100%" @change="onSourceRemoteChange">
                <el-option label="Local (本地)" value="local" />
                <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="taskForm.sync_mode !== 'all-branch'" label="源分支">
              <el-select
                v-model="taskForm.source_branch"
                filterable
                style="width: 100%"
                placeholder="选择源分支"
                :loading="branchLoading"
              >
                <el-option v-for="b in sourceBranches" :key="b" :label="b" :value="b" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="目标 (Target)">
              <el-select v-model="taskForm.target_remote" style="width: 100%">
                <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="taskForm.sync_mode !== 'all-branch'" label="目标分支">
              <el-select
                v-model="taskForm.target_branch"
                filterable
                allow-create
                default-first-option
                style="width: 100%"
                placeholder="选择或输入目标分支（回车新建）"
                :loading="branchLoading"
              >
                <el-option v-for="b in targetBranches" :key="b" :label="b" :value="b" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-alert v-if="taskForm.sync_mode === 'all-branch'" title="全分支模式将自动同步源端所有分支到目标端对应分支" type="info" :closable="false" show-icon class="mb-3" />
        <el-alert v-if="taskForm.source_remote === taskForm.target_remote" title="源和目标不能相同" type="warning" :closable="false" show-icon class="mb-3" />

        <el-divider content-position="left">Git 选项</el-divider>
        
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item>
              <el-checkbox v-model="taskForm.git_tags">--tags 推送所有标签</el-checkbox>
            </el-form-item>
            <el-form-item>
              <el-checkbox v-model="taskForm.git_prune">--prune 清理已删除分支</el-checkbox>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item>
              <el-checkbox v-model="taskForm.git_force">--force 强制推送 ⚠️</el-checkbox>
            </el-form-item>
            <el-form-item>
              <el-checkbox v-model="taskForm.git_no_verify">--no-verify 跳过钩子</el-checkbox>
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">定时任务</el-divider>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Cron">
              <el-input v-model="taskForm.cron" placeholder="0 2 * * * (留空禁用)" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="启用">
              <el-switch v-model="taskForm.enabled" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="showTaskDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveTask" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <!-- History Dialog -->
    <el-dialog v-model="showHistoryDialog" title="同步历史" width="900px">
      <el-table :data="historyList" size="small" border>
        <el-table-column prop="start_time" label="时间" width="160">
          <template #default="{ row }">{{ formatDate(row.start_time) }}</template>
        </el-table-column>
        <el-table-column prop="trigger_source" label="触发" width="100">
          <template #default="{ row }">
            <el-tag :type="getTriggerTagType(row.trigger_source)" size="small">{{ getTriggerLabel(row.trigger_source) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusColor(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="100">
          <template #default="{ row }">
            {{ row.end_time ? (new Date(row.end_time).getTime() - new Date(row.start_time).getTime()) + 'ms' : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="详情">
          <template #default="{ row }">
            <el-button size="small" link @click="showLog(row.details)">日志</el-button>
            <span v-if="row.error_message" class="error-msg">{{ row.error_message }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- Log Dialog -->
    <el-dialog v-model="showLogDialog" title="执行详情" width="700px">
      <pre class="log-content">{{ logContent }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft, Plus, Setting, Refresh, Close, Right, CaretRight, Clock,
  AlarmClock, More
} from '@element-plus/icons-vue'
import {
  getSyncTasks, createSyncTask, updateSyncTask, deleteSyncTask,
  runSyncTask, getSyncHistory, previewSync, batchSync
} from '@/api/modules/sync'
import { getRepoDetail, scanRepo } from '@/api/modules/repo'
import { getBranchList } from '@/api/modules/branch'
import type { BranchInfo } from '@/types/branch'
import type { SyncTaskDTO, SyncRunDTO, PreviewSyncResponse } from '@/types/sync'
import { formatDate, getStatusColor } from '@/utils/format'

const route = useRoute()
const repoKey = route.params.repoKey as string

const loading = ref(false)
const saving = ref(false)
const syncing = ref(false)
const previewing = ref(false)
const branchLoading = ref(false)
const tasks = ref<SyncTaskDTO[]>([])
const remoteNames = ref<string[]>([])
const allBranches = ref<BranchInfo[]>([])
const selectedTasks = ref<string[]>([])

// Branches available for the currently selected source remote
const sourceBranches = computed(() => {
  const remote = taskForm.value.source_remote
  if (remote === 'local') {
    return allBranches.value.filter(b => b.type === 'local').map(b => b.name)
  }
  const prefix = remote + '/'
  return allBranches.value
    .filter(b => b.type === 'remote' && b.name.startsWith(prefix))
    .map(b => b.name.slice(prefix.length))
})

// Target branch options: local branches (user can also type custom)
const targetBranches = computed(() => {
  return allBranches.value.filter(b => b.type === 'local').map(b => b.name)
})

// Quick sync
const showQuickPanel = ref(false)
const quickForm = ref({
  sourceRemote: 'local',
  sourceBranch: 'main',
  targetRemote: '',
  targetBranch: 'main',
  gitTags: false,
  gitForce: false,
  gitPrune: false,
})
const previewResult = ref<PreviewSyncResponse | null>(null)

// Task dialog
const showTaskDialog = ref(false)
const editingTask = ref<SyncTaskDTO | null>(null)
const taskForm = ref({
  source_remote: 'local',
  source_branch: 'main',
  target_remote: '',
  target_branch: 'main',
  cron: '',
  enabled: true,
  sync_mode: 'single',
  git_tags: false,
  git_force: false,
  git_prune: false,
  git_no_verify: false,
})

// History dialog
const showHistoryDialog = ref(false)
const historyList = ref<SyncRunDTO[]>([])
const showLogDialog = ref(false)
const logContent = ref('')

function getTriggerTagType(source: string) {
  switch (source) {
    case 'cron': return 'warning'
    case 'webhook': return 'success'
    case 'manual': return 'primary'
    default: return 'info'
  }
}

function getTriggerLabel(source: string) {
  switch (source) {
    case 'cron': return '定时'
    case 'webhook': return 'Webhook'
    case 'manual': return '手动'
    default: return source || '手动'
  }
}

function onSourceRemoteChange() {
  // Reset source branch when remote changes (old value may not exist in new list)
  taskForm.value.source_branch = ''
}

onMounted(async () => {
  await loadTasks()
  try {
    const repo = await getRepoDetail(repoKey)
    if (repo?.path) {
      const scan = await scanRepo(repo.path)
      remoteNames.value = (scan.remotes || []).map((r) => r.name)
      quickForm.value.targetRemote = remoteNames.value[0] || ''
      taskForm.value.target_remote = remoteNames.value[0] || ''
    }
  } catch { /* ignore */ }
  // Load branches for dropdown
  branchLoading.value = true
  try {
    const result = await getBranchList(repoKey, { page_size: 500 })
    allBranches.value = result?.list || []
  } catch { /* ignore */ } finally {
    branchLoading.value = false
  }
})

async function loadTasks() {
  loading.value = true
  try {
    tasks.value = (await getSyncTasks(repoKey)) || []
  } finally {
    loading.value = false
  }
}

function openQuickSync() {
  showQuickPanel.value = !showQuickPanel.value
  previewResult.value = null
}

function openAddTask() {
  editingTask.value = null
  taskForm.value = {
    source_remote: 'local',
    source_branch: 'main',
    target_remote: remoteNames.value[0] || '',
    target_branch: 'main',
    cron: '',
    enabled: true,
    sync_mode: 'single',
    git_tags: false,
    git_force: false,
    git_prune: false,
    git_no_verify: false,
  }
  showTaskDialog.value = true
}

function openEditTask(task: SyncTaskDTO) {
  editingTask.value = task
  taskForm.value = {
    source_remote: task.source_remote,
    source_branch: task.source_branch,
    target_remote: task.target_remote,
    target_branch: task.target_branch,
    cron: task.cron,
    enabled: task.enabled,
    sync_mode: task.sync_mode || 'single',
    git_tags: task.git_tags,
    git_force: task.git_force,
    git_prune: task.git_prune,
    git_no_verify: task.git_no_verify,
  }
  showTaskDialog.value = true
}

async function handleSaveTask() {
  if (taskForm.value.source_remote === taskForm.value.target_remote) {
    ElMessage.warning('源和目标不能相同')
    return
  }
  saving.value = true
  try {
    if (editingTask.value) {
      await updateSyncTask({
        key: editingTask.value.key,
        source_repo_key: repoKey,
        target_repo_key: repoKey,
        ...taskForm.value,
      })
    } else {
      await createSyncTask({
        source_repo_key: repoKey,
        target_repo_key: repoKey,
        ...taskForm.value,
      })
    }
    ElMessage.success('保存成功')
    showTaskDialog.value = false
    await loadTasks()
  } finally {
    saving.value = false
  }
}

async function handlePreview() {
  previewing.value = true
  try {
    previewResult.value = await previewSync({
      repo_key: repoKey,
      source_remote: quickForm.value.sourceRemote,
      source_branch: quickForm.value.sourceBranch,
      target_remote: quickForm.value.targetRemote,
      target_branch: quickForm.value.targetBranch,
      git_tags: quickForm.value.gitTags,
      git_force: quickForm.value.gitForce,
      git_prune: quickForm.value.gitPrune,
    })
  } catch (e: any) {
    ElMessage.error(e.message || '预览失败')
  } finally {
    previewing.value = false
  }
}

async function handleQuickSync() {
  if (quickForm.value.gitForce) {
    try {
      await ElMessageBox.confirm('--force 会覆盖远端提交，确定继续？', '危险操作', { type: 'warning' })
    } catch {
      return
    }
  }
  
  syncing.value = true
  try {
    // Create a temporary task and run it
    const result = await createSyncTask({
      source_repo_key: repoKey,
      target_repo_key: repoKey,
      source_remote: quickForm.value.sourceRemote,
      source_branch: quickForm.value.sourceBranch,
      target_remote: quickForm.value.targetRemote,
      target_branch: quickForm.value.targetBranch,
      git_tags: quickForm.value.gitTags,
      git_force: quickForm.value.gitForce,
      git_prune: quickForm.value.gitPrune,
      enabled: false,
    }) as any
    if (result?.task_key) {
      await runSyncTask(result.task_key)
      ElMessage.success('同步已触发')
    }
  } catch (e: any) {
    ElMessage.error(e.message || '同步失败')
  } finally {
    syncing.value = false
  }
}

async function handleRun(key: string) {
  try {
    await runSyncTask(key)
    ElMessage.success('任务已触发')
  } catch { /* handled */ }
}

async function handleBatchSync() {
  if (selectedTasks.value.length === 0) return
  
  try {
    await ElMessageBox.confirm(`确定同步 ${selectedTasks.value.length} 个规则？`, '批量同步', { type: 'info' })
    await batchSync(selectedTasks.value)
    ElMessage.success('批量同步已触发')
    selectedTasks.value = []
  } catch { /* cancelled */ }
}

async function toggleEnabled(task: SyncTaskDTO) {
  try {
    await updateSyncTask({
      key: task.key,
      source_repo_key: repoKey,
      target_repo_key: repoKey,
      source_remote: task.source_remote,
      source_branch: task.source_branch,
      target_remote: task.target_remote,
      target_branch: task.target_branch,
      enabled: !task.enabled,
    })
    ElMessage.success(task.enabled ? '已暂停' : '已启用')
    await loadTasks()
  } catch { /* handled */ }
}

async function handleDelete(key: string) {
  try {
    await ElMessageBox.confirm('确定删除该同步规则吗？', '确认删除', { type: 'warning' })
    await deleteSyncTask(key)
    ElMessage.success('删除成功')
    await loadTasks()
  } catch { /* cancelled */ }
}

async function showHistory(taskKey: string) {
  try {
    const all = await getSyncHistory()
    historyList.value = all.filter((h) => h.task_key === taskKey)
    showHistoryDialog.value = true
  } catch { /* handled */ }
}

function showLog(details: string) {
  logContent.value = details || '无详情'
  showLogDialog.value = true
}
</script>

<style scoped>
.sync-page {
  padding: 0;
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
  gap: 8px;
}

.header-left h2 {
  margin: 0;
  font-size: 20px;
}

.header-right {
  display: flex;
  gap: 8px;
}

/* Quick Panel */
.quick-panel {
  margin-bottom: 20px;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.quick-form {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.arrow-icon {
  font-size: 20px;
  color: #409eff;
  margin: 0 8px;
}

.preview-result {
  margin-top: 16px;
}

/* Task List */
.task-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-card {
  transition: all 0.3s;
  border-left: 4px solid #409eff;
}

.task-card.disabled {
  opacity: 0.6;
  border-left-color: #c0c4cc;
}

.task-content {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 8px 0;
}

.task-checkbox {
  flex-shrink: 0;
}

.direction-flow {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.endpoint {
  display: flex;
  flex-direction: column;
  padding: 8px 16px;
  border-radius: 6px;
  min-width: 100px;
}

.endpoint.source {
  background: #f0f9eb;
  border: 1px solid #e1f3d8;
}

.endpoint.target {
  background: #fdf6ec;
  border: 1px solid #faecd8;
}

.endpoint .label {
  font-weight: 600;
  font-size: 14px;
  color: #303133;
}

.endpoint .branch {
  font-size: 12px;
  color: #909399;
  font-family: monospace;
}

.flow-arrow {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  color: #409eff;
}

.task-status {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: flex-end;
}

.cron {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #909399;
}

.git-options {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.task-actions {
  display: flex;
  gap: 8px;
}

.text-center {
  text-align: center;
  padding: 40px 0;
}

.mb-3 {
  margin-bottom: 12px;
}

.log-content {
  background: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  max-height: 400px;
  overflow-y: auto;
  white-space: pre-wrap;
  font-size: 13px;
  font-family: monospace;
}

.error-msg {
  color: #f56c6c;
  font-size: 12px;
  margin-left: 8px;
}
</style>
