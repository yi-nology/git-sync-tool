<template>
  <div class="repo-detail-page" v-loading="loading">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="$router.push('/repos')" :icon="ArrowLeft" text>返回</el-button>
        <h2>{{ repo?.name || '仓库详情' }}</h2>
        <el-tag v-if="currentVersion" size="small" type="success">{{ currentVersion }}</el-tag>
      </div>
      <div class="header-actions">
        <el-button type="success" @click="$router.push(`/repos/${repoKey}/branches`)">
          <el-icon><Share /></el-icon> 分支管理
        </el-button>
        <el-button type="primary" @click="$router.push(`/repos/${repoKey}/compare`)">
          <el-icon><Switch /></el-icon> 分支对比
        </el-button>
        <el-button type="warning" @click="$router.push(`/repos/${repoKey}/sync`)">
          <el-icon><Refresh /></el-icon> 同步任务
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab">
      <!-- 基本信息 Tab -->
      <el-tab-pane label="基本信息" name="info">
        <el-card v-if="repo">
          <template #header>
            <div class="card-header-row">
              <span>基本信息</span>
              <el-button type="primary" size="small" @click="openEditDialog">
                <el-icon><Edit /></el-icon> 编辑仓库
              </el-button>
            </div>
          </template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="名称">{{ repo.name }}</el-descriptions-item>
            <el-descriptions-item label="当前版本">
              <el-tag v-if="currentVersion" type="success" size="small">{{ currentVersion }}</el-tag>
              <span v-else>-</span>
            </el-descriptions-item>
            <el-descriptions-item label="本地路径" :span="2">
              <el-text class="mono-text">{{ repo.path }}</el-text>
            </el-descriptions-item>
            <el-descriptions-item label="Repo Key">
              <el-text class="mono-text">{{ repo.key }}</el-text>
              <el-button size="small" link @click="copyKey">复制</el-button>
            </el-descriptions-item>
            <el-descriptions-item label="配置来源">
              <el-tag size="small">{{ repo.config_source === 'database' ? '数据库' : '本地文件' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="远程 URL" :span="2">{{ repo.remote_url || '-' }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ formatDate(repo.created_at) }}</el-descriptions-item>
            <el-descriptions-item label="更新时间">{{ formatDate(repo.updated_at) }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- Scan Info -->
        <el-card v-if="scanData" class="mt-4" header="远程配置 (来自 .git/config)">
          <el-table :data="scanData.remotes" size="small" border>
            <el-table-column prop="name" label="Remote Name" width="120" />
            <el-table-column prop="fetch_url" label="Fetch URL" />
            <el-table-column prop="push_url" label="Push URL" />
            <el-table-column label="Mirror" width="80">
              <template #default="{ row }">
                <el-tag v-if="row.is_mirror" size="small" type="warning">Yes</el-tag>
                <span v-else>-</span>
              </template>
            </el-table-column>
          </el-table>
          <div class="mt-3" v-if="scanData.branches?.length">
            <strong>分支追踪:</strong>
            <el-tag v-for="b in scanData.branches" :key="b.name" size="small" class="ml-1 mt-1">
              {{ b.name }} -> {{ b.upstream_ref }}
            </el-tag>
          </div>
        </el-card>
      </el-tab-pane>

      <!-- 提交统计 Tab -->
      <el-tab-pane label="Git有效提交度量" name="stats">
        <el-card>
          <el-form inline class="filter-form">
            <el-form-item label="分支">
              <el-select v-model="statsFilter.branch" placeholder="全部" clearable @change="loadStats">
                <el-option v-for="b in statsBranches" :key="b" :label="b" :value="b" />
              </el-select>
            </el-form-item>
            <el-form-item label="提交人">
              <el-select v-model="statsFilter.author" placeholder="全部" clearable filterable @change="loadStats">
                <el-option v-for="a in statsAuthors" :key="a" :label="a" :value="a" />
              </el-select>
            </el-form-item>
            <el-form-item label="开始日期">
              <el-date-picker v-model="statsFilter.since" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" />
            </el-form-item>
            <el-form-item label="结束日期">
              <el-date-picker v-model="statsFilter.until" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="loadStats">
                <el-icon><Search /></el-icon> 查询
              </el-button>
              <el-button @click="handleExportCsv('stats')">
                <el-icon><Download /></el-icon> 导出 CSV
              </el-button>
            </el-form-item>
          </el-form>

          <div v-if="statsData">
            <el-row :gutter="16" class="mb-4">
              <el-col :span="12">
                <el-statistic title="总有效行数" :value="statsData.total_lines" />
              </el-col>
              <el-col :span="12">
                <el-statistic title="活跃贡献者" :value="statsData.authors?.length || 0" />
              </el-col>
            </el-row>

            <!-- Authors table -->
            <el-table :data="statsData.authors || []" border size="small" class="mb-4">
              <el-table-column prop="name" label="作者" width="150" />
              <el-table-column prop="email" label="邮箱" width="200" />
              <el-table-column prop="total_lines" label="总行数" width="100" sortable />
            </el-table>

            <!-- Commit history table -->
            <h4>提交历史（最近100条）</h4>
            <el-table :data="commitHistory" border size="small" max-height="400">
              <el-table-column prop="hash" label="Hash" width="100">
                <template #default="{ row }">
                  <el-text class="mono-text" size="small">{{ row.hash?.substring(0, 8) }}</el-text>
                </template>
              </el-table-column>
              <el-table-column prop="author" label="作者" width="120" />
              <el-table-column prop="date" label="时间" width="160">
                <template #default="{ row }">{{ formatRelativeTime(row.date) }}</template>
              </el-table-column>
              <el-table-column prop="message" label="信息" />
            </el-table>
          </div>
          <el-empty v-else description="点击查询按钮加载数据" />
        </el-card>
      </el-tab-pane>

      <!-- 代码行统计 Tab -->
      <el-tab-pane label="真实工程代码度量" name="lines">
        <el-card>
          <el-form inline class="filter-form">
            <el-form-item label="分支">
              <el-select v-model="lineStatsFilter.branch" placeholder="当前工作区" clearable @change="loadLineStats">
                <el-option v-for="b in statsBranches" :key="b" :label="b" :value="b" />
              </el-select>
            </el-form-item>
            <el-form-item label="提交人">
              <el-select v-model="lineStatsFilter.author" placeholder="全部" clearable filterable>
                <el-option v-for="a in statsAuthors" :key="a" :label="a" :value="a" />
              </el-select>
            </el-form-item>
            <el-form-item label="开始日期">
              <el-date-picker v-model="lineStatsFilter.since" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" />
            </el-form-item>
            <el-form-item label="结束日期">
              <el-date-picker v-model="lineStatsFilter.until" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="loadLineStats">
                <el-icon><Search /></el-icon> 查询
              </el-button>
              <el-button @click="openExcludeConfig">
                <el-icon><Setting /></el-icon> 排除配置
              </el-button>
              <el-button @click="handleExportCsv('lines')">
                <el-icon><Download /></el-icon> 导出 CSV
              </el-button>
            </el-form-item>
          </el-form>

          <el-alert type="info" :closable="false" show-icon class="mb-4">
            选择分支/提交人/时间范围后将使用 git blame 分析代码归属，统计速度会较慢
          </el-alert>

          <div v-if="lineStatsData">
            <el-row :gutter="16" class="mb-4">
              <el-col :span="6"><el-statistic title="代码行数" :value="lineStatsData.code_lines" /></el-col>
              <el-col :span="6"><el-statistic title="注释行数" :value="lineStatsData.comment_lines" /></el-col>
              <el-col :span="6"><el-statistic title="空白行数" :value="lineStatsData.blank_lines" /></el-col>
              <el-col :span="6"><el-statistic title="文件总数" :value="lineStatsData.total_files" /></el-col>
            </el-row>

            <el-alert v-if="lineStatsData.status === 'processing'" title="正在统计中..." type="info" :closable="false" show-icon>
              {{ lineStatsData.progress }}
            </el-alert>

            <el-table :data="lineStatsData.languages || []" border size="small" v-if="lineStatsData.languages?.length">
              <el-table-column prop="name" label="语言" width="150" />
              <el-table-column prop="files" label="文件数" width="100" sortable />
              <el-table-column prop="code" label="代码行" width="100" sortable />
              <el-table-column prop="comment" label="注释行" width="100" sortable />
              <el-table-column prop="blank" label="空白行" width="100" sortable />
              <el-table-column label="总行数" width="100" sortable>
                <template #default="{ row }">{{ (row.code || 0) + (row.comment || 0) + (row.blank || 0) }}</template>
              </el-table-column>
            </el-table>
          </div>
          <el-empty v-else description="点击查询按钮加载数据" />
        </el-card>
      </el-tab-pane>

      <!-- 版本历史 Tab -->
      <el-tab-pane label="版本历史" name="versions">
        <el-card>
          <div v-if="versionList.length === 0 && !versionsLoading">
            <el-empty description="暂无版本标签" />
          </div>
          <div v-loading="versionsLoading" class="version-timeline">
            <el-timeline v-if="versionList.length > 0">
              <el-timeline-item
                v-for="v in versionList"
                :key="v.name"
                :timestamp="formatDate(v.date)"
                placement="top"
                type="primary"
                :hollow="false"
                size="large"
              >
                <el-card shadow="hover" class="version-card">
                  <div class="version-header">
                    <el-tag type="success" size="large">{{ v.name }}</el-tag>
                  </div>
                  <div class="version-info">
                    <div><strong>Commit:</strong> <el-text class="mono-text" size="small">{{ v.hash?.substring(0, 8) }}</el-text></div>
                    <div v-if="v.tagger"><strong>作者:</strong> {{ v.tagger }}</div>
                    <div v-if="v.message"><strong>说明:</strong> {{ v.message }}</div>
                  </div>
                </el-card>
              </el-timeline-item>
            </el-timeline>
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- Edit Repo Dialog -->
    <el-dialog v-model="showEditDialog" title="编辑仓库" width="550px" destroy-on-close>
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="本地路径" required>
          <el-input v-model="editForm.path" />
        </el-form-item>
        <el-form-item label="远程 URL">
          <el-input v-model="editForm.remote_url" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveEdit" :loading="editSaving">保存</el-button>
      </template>
    </el-dialog>

    <!-- Exclude Config Dialog -->
    <el-dialog v-model="showExcludeDialog" title="排除配置" width="550px" destroy-on-close>
      <el-form label-width="100px">
        <el-form-item label="排除目录">
          <el-input v-model="excludeDirsText" type="textarea" :rows="4" placeholder="每行一个目录路径" />
        </el-form-item>
        <el-form-item label="排除规则">
          <el-input v-model="excludePatternsText" type="textarea" :rows="4" placeholder="每行一个 glob 规则" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showExcludeDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveExclude">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Share, Switch, Refresh, Edit, Search, Download, Setting } from '@element-plus/icons-vue'
import { getRepoDetail, scanRepo, updateRepo } from '@/api/modules/repo'
import { getStatsAnalyze, getStatsAuthors, getStatsBranches, getStatsCommits, getLineStats, getLineStatsConfig, saveLineStatsConfig, exportStatsCsv } from '@/api/modules/stats'
import { getVersionList, getCurrentVersion } from '@/api/modules/version'
import type { VersionTag } from '@/api/modules/version'
import type { RepoDTO, ScanResult } from '@/types/repo'
import type { StatsResponse, LineStatsResponse } from '@/types/stats'
import { formatDate, formatRelativeTime } from '@/utils/format'

const route = useRoute()
const repoKey = route.params.repoKey as string

const loading = ref(false)
const repo = ref<RepoDTO | null>(null)
const scanData = ref<ScanResult | null>(null)
const activeTab = ref('info')
const currentVersion = ref('')

// Stats
const statsFilter = ref({ branch: '', author: '', since: '', until: '' })
const lineStatsFilter = ref({ branch: '', author: '', since: '', until: '' })
const statsBranches = ref<string[]>([])
const statsAuthors = ref<string[]>([])
const statsData = ref<StatsResponse | null>(null)
const lineStatsData = ref<LineStatsResponse | null>(null)
const commitHistory = ref<{ hash: string; author: string; date: string; message: string }[]>([])

// Versions
const versionList = ref<VersionTag[]>([])
const versionsLoading = ref(false)

// Edit
const showEditDialog = ref(false)
const editSaving = ref(false)
const editForm = ref({ name: '', path: '', remote_url: '' })

// Exclude config
const showExcludeDialog = ref(false)
const excludeDirsText = ref('')
const excludePatternsText = ref('')

onMounted(async () => {
  loading.value = true
  try {
    repo.value = await getRepoDetail(repoKey)
    if (repo.value?.path) {
      try { scanData.value = await scanRepo(repo.value.path) } catch { /* ignore */ }
    }
    try { statsBranches.value = await getStatsBranches(repoKey) } catch { /* ignore */ }
    try { statsAuthors.value = await getStatsAuthors(repoKey) } catch { /* ignore */ }
    try { currentVersion.value = await getCurrentVersion(repoKey) || '' } catch { /* ignore */ }
  } finally {
    loading.value = false
  }
})

// Load versions when tab switches
watch(activeTab, (val) => {
  if (val === 'versions' && versionList.value.length === 0) {
    loadVersions()
  }
})

async function loadStats() {
  try {
    statsData.value = await getStatsAnalyze(repoKey, {
      branch: statsFilter.value.branch || undefined,
      author: statsFilter.value.author || undefined,
      since: statsFilter.value.since || undefined,
      until: statsFilter.value.until || undefined,
    })
    // Load commit history
    const res = await getStatsCommits(repoKey, {
      branch: statsFilter.value.branch || undefined,
      author: statsFilter.value.author || undefined,
      since: statsFilter.value.since || undefined,
      until: statsFilter.value.until || undefined,
    })
    commitHistory.value = (Array.isArray(res) ? res : []).slice(0, 100)
  } catch { /* ignore */ }
}

async function loadLineStats() {
  try {
    lineStatsData.value = await getLineStats(repoKey, {
      branch: lineStatsFilter.value.branch || undefined,
    })
  } catch { /* ignore */ }
}

async function loadVersions() {
  versionsLoading.value = true
  try {
    versionList.value = await getVersionList(repoKey)
  } catch { /* ignore */ }
  finally {
    versionsLoading.value = false
  }
}

function copyKey() {
  if (repo.value?.key) {
    navigator.clipboard.writeText(repo.value.key)
    ElMessage.success('已复制 Repo Key')
  }
}

function openEditDialog() {
  if (!repo.value) return
  editForm.value = {
    name: repo.value.name,
    path: repo.value.path,
    remote_url: repo.value.remote_url || '',
  }
  showEditDialog.value = true
}

async function handleSaveEdit() {
  if (!editForm.value.name || !editForm.value.path) {
    ElMessage.warning('名称和路径不能为空')
    return
  }
  editSaving.value = true
  try {
    await updateRepo({
      key: repoKey,
      name: editForm.value.name,
      path: editForm.value.path,
      remote_url: editForm.value.remote_url || undefined,
    })
    ElMessage.success('保存成功')
    showEditDialog.value = false
    repo.value = await getRepoDetail(repoKey)
  } finally {
    editSaving.value = false
  }
}

async function handleExportCsv(type: string) {
  try {
    const params: Record<string, string> = { type }
    if (type === 'stats') {
      if (statsFilter.value.branch) params.branch = statsFilter.value.branch
      if (statsFilter.value.author) params.author = statsFilter.value.author
      if (statsFilter.value.since) params.since = statsFilter.value.since
      if (statsFilter.value.until) params.until = statsFilter.value.until
    } else {
      if (lineStatsFilter.value.branch) params.branch = lineStatsFilter.value.branch
    }
    const response = await exportStatsCsv(repoKey, params) as unknown as Blob
    const url = window.URL.createObjectURL(response)
    const a = document.createElement('a')
    a.href = url
    a.download = `${repo.value?.name || repoKey}-${type}.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  } catch { ElMessage.error('导出失败') }
}

async function openExcludeConfig() {
  try {
    const config = await getLineStatsConfig(repoKey)
    excludeDirsText.value = (config.exclude_dirs || []).join('\n')
    excludePatternsText.value = (config.exclude_patterns || []).join('\n')
  } catch { /* ignore */ }
  showExcludeDialog.value = true
}

async function handleSaveExclude() {
  try {
    await saveLineStatsConfig(repoKey, {
      exclude_dirs: excludeDirsText.value.split('\n').map(s => s.trim()).filter(Boolean),
      exclude_patterns: excludePatternsText.value.split('\n').map(s => s.trim()).filter(Boolean),
    })
    ElMessage.success('排除配置已保存')
    showExcludeDialog.value = false
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
.card-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.filter-form {
  margin-bottom: 16px;
}
.mono-text {
  font-family: monospace;
}
.mt-3 {
  margin-top: 12px;
}
.mt-4 {
  margin-top: 16px;
}
.mb-4 {
  margin-bottom: 16px;
}
.ml-1 {
  margin-left: 4px;
}
.mt-1 {
  margin-top: 4px;
}
.version-timeline {
  padding: 16px 0;
}
.version-card {
  max-width: 500px;
}
.version-header {
  margin-bottom: 8px;
}
.version-info {
  font-size: 14px;
  line-height: 1.8;
  color: #606266;
}
</style>
