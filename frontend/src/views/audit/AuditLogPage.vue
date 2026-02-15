<template>
  <div class="audit-log-page">
    <div class="page-header">
      <h2><el-icon><Warning /></el-icon> 操作审计日志</h2>
      <el-button @click="loadLogs" :icon="RefreshRight">刷新</el-button>
    </div>

    <el-card>
      <el-table :data="logs" v-loading="loading" stripe border>
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="action" label="操作类型" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="getActionType(row.action)">{{ row.action }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="target" label="目标对象" min-width="200" />
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
import { Warning, RefreshRight } from '@element-plus/icons-vue'
import { getAuditLogs } from '@/api/modules/audit'
import type { AuditLogDTO } from '@/types/stats'
import { formatDate } from '@/utils/format'

const loading = ref(false)
const logs = ref<AuditLogDTO[]>([])
const totalCount = ref(0)
const currentPage = ref(1)
const pageSize = 20

const showDetailDialog = ref(false)
const detailContent = ref('')

function getActionType(action: string): '' | 'success' | 'warning' | 'danger' | 'info' {
  if (action === 'CREATE') return 'success'
  if (action === 'DELETE') return 'danger'
  if (action === 'UPDATE') return 'warning'
  return 'info'
}

onMounted(() => {
  loadLogs()
})

async function loadLogs() {
  loading.value = true
  try {
    const res = await getAuditLogs({ page: currentPage.value, page_size: pageSize })
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
