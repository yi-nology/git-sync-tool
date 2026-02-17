<template>
  <div class="sync-task-page">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="$router.push(`/repos/${repoKey}`)" :icon="ArrowLeft" text>返回</el-button>
        <h2>同步任务</h2>
      </div>
      <el-button type="primary" @click="openAddTask">
        <el-icon><Plus /></el-icon> 新建同步规则
      </el-button>
    </div>

    <div v-if="tasks.length === 0 && !loading" class="text-center">
      <el-empty description="暂无同步规则" />
    </div>

    <div v-loading="loading">
      <el-card v-for="task in tasks" :key="task.key" class="task-card" :class="{ disabled: !task.enabled }">
        <div class="task-header">
          <div class="task-title">
            <el-tag size="small" :type="getSyncTagType(task)">{{ getSyncTypeLabel(task.source_remote, task.target_remote) }}</el-tag>
            <strong>{{ task.source_remote }}/{{ task.source_branch }}</strong>
            <el-icon><Right /></el-icon>
            <strong>{{ task.target_remote }}/{{ task.target_branch }}</strong>
          </div>
          <div class="task-actions">
            <el-button size="small" type="success" @click="handleRun(task.key)" :icon="CaretRight" circle title="立即执行" />
            <el-button size="small" type="info" @click="showHistory(task.key)" :icon="Clock" circle title="历史记录" />
            <el-button size="small" @click="openEditTask(task)" :icon="Edit" circle title="编辑" />
            <el-button size="small" type="danger" @click="handleDelete(task.key)" :icon="Delete" circle title="删除" />
          </div>
        </div>
        <div class="task-meta">
          <span v-if="task.cron"><el-icon><AlarmClock /></el-icon> {{ task.cron }}</span>
          <span>
            <el-icon><Open v-if="task.enabled" /><TurnOff v-else /></el-icon>
            {{ task.enabled ? '已启用' : '已禁用' }}
          </span>
          <span v-if="task.push_options"><el-icon><Monitor /></el-icon> {{ task.push_options }}</span>
        </div>
      </el-card>
    </div>

    <!-- Task Dialog -->
    <el-dialog v-model="showTaskDialog" :title="editingTask ? '编辑同步规则' : '新建同步规则'" width="650px" destroy-on-close>
      <el-form :model="taskForm" label-width="100px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="源 (Source)">
              <el-select v-model="taskForm.source_remote" style="width: 100%">
                <el-option label="Local (本地)" value="local" />
                <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
              </el-select>
            </el-form-item>
            <el-form-item label="源分支">
              <el-input v-model="taskForm.source_branch" placeholder="main" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="目标 (Target)">
              <el-select v-model="taskForm.target_remote" style="width: 100%">
                <el-option label="Local (本地)" value="local" />
                <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
              </el-select>
            </el-form-item>
            <el-form-item label="目标分支">
              <el-input v-model="taskForm.target_branch" placeholder="main" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-alert v-if="taskForm.source_remote === taskForm.target_remote" title="源和目标不能相同" type="warning" :closable="false" show-icon class="mb-3" />

        <el-form-item label="Push 选项">
          <el-input v-model="taskForm.push_options" placeholder="例如: --force" />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="定时 (Cron)">
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
    <el-dialog v-model="showHistoryDialog" title="同步历史" width="800px">
      <el-table :data="historyList" size="small" border>
        <el-table-column prop="start_time" label="时间" width="160">
          <template #default="{ row }">{{ formatDate(row.start_time) }}</template>
        </el-table-column>
        <el-table-column prop="trigger_source" label="触发来源" width="100">
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
    <el-dialog v-model="showLogDialog" title="执行详情" width="600px">
      <pre class="log-content">{{ logContent }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Plus, Right, CaretRight, Clock, Edit, Delete, AlarmClock, Open, TurnOff, Monitor } from '@element-plus/icons-vue'
import { getSyncTasks, createSyncTask, updateSyncTask, deleteSyncTask, runSyncTask, getSyncHistory } from '@/api/modules/sync'
import { getRepoDetail, scanRepo } from '@/api/modules/repo'
import type { SyncTaskDTO, SyncRunDTO } from '@/types/sync'
import { formatDate, getStatusColor, getSyncTypeLabel } from '@/utils/format'

const route = useRoute()
const repoKey = route.params.repoKey as string

const loading = ref(false)
const saving = ref(false)
const tasks = ref<SyncTaskDTO[]>([])
const remoteNames = ref<string[]>([])

const showTaskDialog = ref(false)
const editingTask = ref<SyncTaskDTO | null>(null)
const taskForm = ref({
  source_remote: 'local',
  source_branch: 'main',
  target_remote: '',
  target_branch: 'main',
  push_options: '',
  cron: '',
  enabled: true,
})

const showHistoryDialog = ref(false)
const historyList = ref<SyncRunDTO[]>([])
const showLogDialog = ref(false)
const logContent = ref('')

function getSyncTagType(task: SyncTaskDTO) {
  if (task.source_remote === 'local') return 'success'
  if (task.target_remote === 'local') return 'warning'
  return 'primary'
}

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

onMounted(async () => {
  await loadTasks()
  try {
    const repo = await getRepoDetail(repoKey)
    if (repo?.path) {
      const scan = await scanRepo(repo.path)
      remoteNames.value = (scan.remotes || []).map((r) => r.name)
    }
  } catch { /* ignore */ }
})

async function loadTasks() {
  loading.value = true
  try {
    tasks.value = await getSyncTasks(repoKey)
  } finally {
    loading.value = false
  }
}

function openAddTask() {
  editingTask.value = null
  taskForm.value = {
    source_remote: 'local',
    source_branch: 'main',
    target_remote: remoteNames.value[0] || '',
    target_branch: 'main',
    push_options: '',
    cron: '',
    enabled: true,
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
    push_options: task.push_options,
    cron: task.cron,
    enabled: task.enabled,
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

async function handleRun(key: string) {
  try {
    await runSyncTask(key)
    ElMessage.success('任务已触发')
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
.task-card {
  margin-bottom: 12px;
  border-left: 4px solid #409eff;
}
.task-card.disabled {
  opacity: 0.6;
}
.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.task-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
}
.task-actions {
  display: flex;
  gap: 4px;
}
.task-meta {
  display: flex;
  gap: 16px;
  margin-top: 10px;
  font-size: 13px;
  color: #909399;
}
.task-meta span {
  display: flex;
  align-items: center;
  gap: 4px;
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
