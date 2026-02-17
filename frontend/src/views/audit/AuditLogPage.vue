<template>
  <div class="audit-log-page">
    <div class="page-header">
      <h2><el-icon><Warning /></el-icon> 操作审计日志</h2>
      <el-button @click="loadLogs" :icon="RefreshRight">刷新</el-button>
    </div>

    <el-card>
      <div class="filter-bar">
        <el-select v-model="filterAction" placeholder="操作类型" clearable filterable style="width: 160px" @change="loadLogs">
          <el-option v-for="(label, key) in actionLabelMap" :key="key" :label="label" :value="key" />
        </el-select>
        <el-input v-model="filterTarget" placeholder="目标对象" clearable style="width: 200px" @clear="loadLogs" @keyup.enter="loadLogs" />
        <el-date-picker v-model="filterDateRange" type="daterange" range-separator="~" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" style="width: 280px" @change="loadLogs" />
        <el-button type="primary" @click="loadLogs" :icon="Search">搜索</el-button>
      </div>

      <el-table :data="logs" v-loading="loading" stripe border>
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="action" label="操作类型" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="getActionType(row.action)">{{ getActionLabel(row.action) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="target" label="目标对象" min-width="200">
          <template #default="{ row }">{{ formatTarget(row.target) }}</template>
        </el-table-column>
        <el-table-column label="操作人 / IP" width="200">
          <template #default="{ row }">
            <div>{{ row.operator || '-' }}</div>
            <el-text type="info" size="small">{{ row.ip_address }}</el-text>
          </template>
        </el-table-column>
        <el-table-column label="详情" width="100">
          <template #default="{ row }">
            <el-button v-if="row.details" size="small" link @click="showDetail(row.details)">查看</el-button>
            <span v-else>-</span>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-bar">
        <el-text type="info" size="small">
          显示 {{ (currentPage - 1) * pageSize + 1 }} - {{ Math.min(currentPage * pageSize, totalCount) }} 共 {{ totalCount }} 条
        </el-text>
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="totalCount"
          layout="prev, pager, next"
          @current-change="loadLogs"
          small
        />
      </div>
    </el-card>

    <!-- Detail Dialog -->
    <el-dialog v-model="showDetailDialog" title="操作详情" width="600px">
      <pre class="detail-content">{{ detailContent }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Warning, RefreshRight, Search } from '@element-plus/icons-vue'
import { getAuditLogs } from '@/api/modules/audit'
import { getRepoList } from '@/api/modules/repo'
import type { AuditLogDTO } from '@/types/stats'
import { formatDate } from '@/utils/format'

const repoNameMap = ref<Record<string, string>>({})

const actionLabelMap: Record<string, string> = {
  CREATE: '创建',
  UPDATE: '更新',
  DELETE: '删除',
  FETCH_REPO: '拉取仓库',
  CREATE_BRANCH: '创建分支',
  DELETE_BRANCH: '删除分支',
  UPDATE_BRANCH: '更新分支',
  CHECKOUT_BRANCH: '切换分支',
  PUSH_BRANCH: '推送分支',
  PULL_BRANCH: '拉取分支',
  MERGE_CONFLICT: '合并冲突',
  MERGE_SUCCESS: '合并成功',
  CHERRY_PICK: 'Cherry-Pick',
  REBASE: '变基',
  REBASE_ABORT: '中止变基',
  REBASE_CONTINUE: '继续变基',
  SUBMODULE_ADD: '添加子模块',
  SUBMODULE_INIT: '初始化子模块',
  SUBMODULE_UPDATE: '更新子模块',
  SUBMODULE_SYNC: '同步子模块',
  SUBMODULE_REMOVE: '删除子模块',
  STASH_SAVE: '保存暂存',
  STASH_APPLY: '应用暂存',
  STASH_POP: '弹出暂存',
  STASH_DROP: '丢弃暂存',
  STASH_CLEAR: '清空暂存',
  SYNC: '同步',
  SYNC_ADHOC: '手动同步',
  SUBMIT_CHANGES: '提交变更',
  WEBHOOK_TRIGGER: 'Webhook 触发',
  WEBHOOK_TRIGGER_BY_TOKEN: 'Token 触发',
  NOTIFICATION_CHANNEL_CREATE: '创建通知渠道',
  NOTIFICATION_CHANNEL_UPDATE: '更新通知渠道',
  NOTIFICATION_CHANNEL_DELETE: '删除通知渠道',
}

function getActionLabel(action: string): string {
  return actionLabelMap[action] || action
}

const targetTypeMap: Record<string, string> = {
  repo: '仓库',
  task: '同步任务',
  task_key: '同步任务',
  channel: '通知渠道',
}

function formatTarget(target: string): string {
  const sepIdx = target.indexOf(':')
  if (sepIdx === -1) return target
  const prefix = target.substring(0, sepIdx)
  const value = target.substring(sepIdx + 1)
  const label = targetTypeMap[prefix]
  if (!label) return target
  if (prefix === 'repo') {
    const name = repoNameMap.value[value]
    return name ? `${label}: ${name}` : `${label}: ${value}`
  }
  return `${label}: ${value}`
}

const loading = ref(false)
const logs = ref<AuditLogDTO[]>([])
const totalCount = ref(0)
const currentPage = ref(1)
const pageSize = 20

const filterAction = ref('')
const filterTarget = ref('')
const filterDateRange = ref<[string, string] | null>(null)

const showDetailDialog = ref(false)
const detailContent = ref('')

function getActionType(action: string): '' | 'success' | 'warning' | 'danger' | 'info' {
  if (action.includes('DELETE') || action.includes('REMOVE') || action === 'MERGE_CONFLICT' || action === 'STASH_DROP' || action === 'STASH_CLEAR') return 'danger'
  if (action.includes('CREATE') || action === 'MERGE_SUCCESS' || action === 'SUBMODULE_ADD') return 'success'
  if (action.includes('UPDATE') || action.includes('PUSH') || action === 'SUBMIT_CHANGES' || action === 'REBASE' || action === 'CHERRY_PICK') return 'warning'
  if (action.includes('SYNC') || action.includes('WEBHOOK') || action.includes('FETCH') || action.includes('PULL')) return ''
  return 'info'
}

onMounted(async () => {
  try {
    const repos = await getRepoList()
    const map: Record<string, string> = {}
    for (const r of repos) {
      map[r.key] = r.name
    }
    repoNameMap.value = map
  } catch {
    // 仓库列表加载失败不影响日志展示
  }
  loadLogs()
})

async function loadLogs() {
  loading.value = true
  try {
    const res = await getAuditLogs({
      page: currentPage.value,
      page_size: pageSize,
      action: filterAction.value || undefined,
      target: filterTarget.value || undefined,
      start_date: filterDateRange.value?.[0] || undefined,
      end_date: filterDateRange.value?.[1] || undefined,
    })
    logs.value = res.items || []
    totalCount.value = res.total || 0
  } finally {
    loading.value = false
  }
}

function showDetail(details: string) {
  try {
    detailContent.value = JSON.stringify(JSON.parse(details), null, 2)
  } catch {
    detailContent.value = details
  }
  showDetailDialog.value = true
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.page-header h2 {
  margin: 0;
  font-size: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.pagination-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
}
.filter-bar {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
  flex-wrap: wrap;
  align-items: center;
}
.detail-content {
  background: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  max-height: 500px;
  overflow: auto;
  white-space: pre-wrap;
  font-family: monospace;
  font-size: 13px;
}
</style>
