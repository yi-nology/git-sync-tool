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
              <el-select v-model="statsFilter.branch" placeholder="全部" clearable @change="loadStats" style="width: 220px">
                <el-option v-for="b in statsBranches" :key="b" :label="b" :value="b" />
              </el-select>
            </el-form-item>
            <el-form-item label="提交人">
              <el-select v-model="statsFilter.author" placeholder="全部" clearable filterable @change="loadStats" style="width: 220px">
                <el-option v-for="a in statsAuthors" :key="a.email" :label="`${a.name}(${a.email})`" :value="a.name" />
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

            <GitStatsCharts :stats-data="statsData" />

            <!-- Commit history table -->
            <el-card shadow="never" class="mt-4">
              <template #header><span style="font-weight:600;font-size:14px">提交历史（最近100条）</span></template>
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
            </el-card>
          </div>
          <el-empty v-else description="点击查询按钮加载数据" />
        </el-card>
      </el-tab-pane>

      <!-- 代码行统计 Tab -->
      <el-tab-pane label="真实工程代码度量" name="lines">
        <el-card>
          <el-form inline class="filter-form">
            <el-form-item label="分支">
              <el-select v-model="lineStatsFilter.branch" placeholder="当前工作区" clearable @change="loadLineStats" style="width: 220px">
                <el-option v-for="b in statsBranches" :key="b" :label="b" :value="b" />
              </el-select>
            </el-form-item>
            <el-form-item label="提交人">
              <el-select v-model="lineStatsFilter.author" placeholder="全部" clearable filterable style="width: 220px">
                <el-option v-for="a in statsAuthors" :key="a.email" :label="`${a.name}(${a.email})`" :value="a.name" />
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

            <LineStatsCharts :line-stats-data="lineStatsData" />
          </div>
          <el-empty v-else description="点击查询按钮加载数据" />
        </el-card>
      </el-tab-pane>

      <!-- 版本历史 Tab -->
      <el-tab-pane label="版本历史" name="versions">
        <el-card>
          <template #header>
            <div class="card-header-row">
              <span>版本标签管理</span>
              <div class="header-actions">
                <el-button size="small" @click="handleFetchTags" :loading="fetchTagsLoading">
                  <el-icon><Download /></el-icon> 拉取远端 Tags
                </el-button>
                <el-button size="small" type="primary" @click="openCreateTagDialog">
                  <el-icon><Plus /></el-icon> 创建 Tag
                </el-button>
                <el-button size="small" @click="loadVersions">
                  <el-icon><Refresh /></el-icon>
                </el-button>
              </div>
            </div>
          </template>
          <div v-if="versionList.length === 0 && !versionsLoading">
            <el-empty description="暂无版本标签">
              <el-button type="primary" @click="openCreateTagDialog">创建第一个 Tag</el-button>
            </el-empty>
          </div>
          <el-table v-else :data="versionList" v-loading="versionsLoading" stripe border size="small">
            <el-table-column prop="name" label="标签名称" width="160">
              <template #default="{ row }">
                <el-tag type="success" size="small">{{ row.name }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="hash" label="Commit" width="120">
              <template #default="{ row }">
                <el-text class="mono-text" size="small">{{ row.hash?.substring(0, 8) }}</el-text>
              </template>
            </el-table-column>
            <el-table-column prop="tagger" label="作者" width="120" />
            <el-table-column prop="date" label="日期" width="160">
              <template #default="{ row }">{{ formatDate(row.date) }}</template>
            </el-table-column>
            <el-table-column prop="message" label="说明" min-width="200" show-overflow-tooltip />
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button-group size="small">
                  <el-button @click="handlePushTag(row.name)">
                    <el-icon><Top /></el-icon> 推送
                  </el-button>
                  <el-button @click="handleCopyHash(row.hash)">
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                  <el-button type="danger" @click="handleDeleteTag(row.name)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </el-button-group>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 文件浏览 Tab -->
      <el-tab-pane label="文件浏览" name="files">
        <FileBrowser :repo-key="repoKey" :branches="allRefs" />
      </el-tab-pane>

      <!-- Commit 搜索 Tab -->
      <el-tab-pane label="Commit 搜索" name="commits">
        <CommitSearch :repo-key="repoKey" :branches="allRefs" :authors="statsAuthors" />
      </el-tab-pane>

      <!-- Stash 管理 Tab -->
      <el-tab-pane label="Stash 管理" name="stash">
        <StashManager :repo-key="repoKey" />
      </el-tab-pane>

      <!-- Submodule 管理 Tab -->
      <el-tab-pane label="Submodule" name="submodules">
        <SubmoduleManager :repo-key="repoKey" />
      </el-tab-pane>
    </el-tabs>

    <!-- Edit Repo Dialog -->
    <el-dialog v-model="showEditDialog" title="编辑仓库" width="750px" destroy-on-close>
      <el-tabs v-model="editActiveTab">
        <el-tab-pane label="基本信息" name="basic">
          <el-form :model="editForm" label-width="100px">
            <el-form-item label="名称" required>
              <el-input v-model="editForm.name" />
            </el-form-item>
            <el-form-item label="本地路径" required>
              <el-input v-model="editForm.path" />
            </el-form-item>
            <el-form-item label="远程 URL">
              <el-input v-model="editForm.remote_url" placeholder="自动从 Remotes 填充" />
            </el-form-item>
            <el-form-item label="配置来源">
              <el-select v-model="editForm.config_source" style="width: 100%">
                <el-option label="本地配置文件 (.git/config)" value="local" />
                <el-option label="数据库 (使用下方配置)" value="database" />
              </el-select>
              <div style="color: #909399; font-size: 12px; margin-top: 4px;">
                决定同步任务使用本地 git 配置还是数据库中存储的远程地址与凭证。
              </div>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="远程仓库 (Remotes) 配置" name="remote">
          <div style="margin-bottom: 12px;">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
              <span style="font-weight: 600;">远程仓库 (Remotes)</span>
              <el-button size="small" type="primary" @click="addEditRemote">+ 新增</el-button>
            </div>
            <el-table :data="editRemotes" border size="small">
              <el-table-column label="Name" width="120">
                <template #default="{ row }">
                  <el-input v-model="row.name" size="small" placeholder="origin" />
                </template>
              </el-table-column>
              <el-table-column label="URL">
                <template #default="{ row }">
                  <el-input v-model="row.fetch_url" size="small" placeholder="Fetch URL" class="mb-1" />
                  <el-input v-model="row.push_url" size="small" placeholder="Push URL (选填)" />
                </template>
              </el-table-column>
              <el-table-column label="Mirror" width="70" align="center">
                <template #default="{ row }">
                  <el-checkbox v-model="row.is_mirror" />
                </template>
              </el-table-column>
              <el-table-column label="操作" width="160" align="center">
                <template #default="{ row, $index }">
                  <el-button size="small" :icon="Connection" circle @click="testEditRemote($index)" title="测试连接" :loading="row._testing" />
                  <el-button size="small" :icon="Lock" circle :type="row._auth?.type && row._auth.type !== 'none' ? 'success' : 'default'" @click="openEditRemoteAuth($index)" title="配置认证" />
                  <el-button size="small" :icon="Delete" circle type="danger" @click="editRemotes.splice($index, 1)" title="删除" />
                </template>
              </el-table-column>
            </el-table>
          </div>

          <!-- Tracking Branches -->
          <div v-if="editTrackingBranches.length > 0" style="margin-top: 16px;">
            <span style="font-weight: 600;">分支追踪</span>
            <div style="margin-top: 8px;">
              <el-tag v-for="b in editTrackingBranches" :key="b.name" size="small" style="margin: 2px 4px;">
                {{ b.name }} -> {{ b.upstream_ref }}
              </el-tag>
            </div>
          </div>
          <div v-else style="margin-top: 16px; color: #909399; font-size: 13px;">
            无追踪分支
          </div>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveEdit" :loading="editSaving">保存</el-button>
      </template>
    </el-dialog>

    <!-- Remote Auth Dialog -->
    <el-dialog v-model="showRemoteAuthDialog" :title="`配置认证: ${remoteAuthName}`" width="480px" destroy-on-close>
      <el-form label-width="110px">
        <el-form-item label="认证方式">
          <el-select v-model="remoteAuthForm.type" style="width: 100%">
            <el-option label="无 (None)" value="none" />
            <el-option label="SSH 密钥" value="ssh" />
            <el-option label="用户名/密码 (HTTP)" value="http" />
          </el-select>
        </el-form-item>
        <template v-if="remoteAuthForm.type === 'ssh'">
          <el-form-item label="SSH 密钥">
            <el-select v-model="remoteAuthForm.key" filterable allow-create placeholder="手动输入路径..." style="width: 100%">
              <el-option v-for="k in sshKeyList" :key="k" :label="k" :value="k" />
            </el-select>
          </el-form-item>
          <el-form-item label="密钥密码">
            <el-input v-model="remoteAuthForm.secret" type="password" show-password placeholder="Passphrase (可选)" />
          </el-form-item>
        </template>
        <template v-if="remoteAuthForm.type === 'http'">
          <el-form-item label="用户名">
            <el-input v-model="remoteAuthForm.key" placeholder="用户名" />
          </el-form-item>
          <el-form-item label="密码 / Token">
            <el-input v-model="remoteAuthForm.secret" type="password" show-password placeholder="密码或 Token" />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="showRemoteAuthDialog = false">取消</el-button>
        <el-button type="primary" @click="saveEditRemoteAuth">确定</el-button>
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

    <!-- Create Tag Dialog -->
    <el-dialog v-model="showCreateTagDialog" title="创建版本标签" width="550px" destroy-on-close>
      <el-form :model="createTagForm" label-width="100px">
        <el-form-item label="版本类型">
          <el-radio-group v-model="createTagForm.versionType" @change="handleVersionTypeChange">
            <el-radio-button value="patch">Patch (修复)</el-radio-button>
            <el-radio-button value="minor">Minor (功能)</el-radio-button>
            <el-radio-button value="major">Major (大版本)</el-radio-button>
            <el-radio-button value="custom">自定义</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="当前版本" v-if="nextVersionInfo">
          <el-tag type="info">{{ nextVersionInfo.current || '无' }}</el-tag>
        </el-form-item>
        <el-form-item label="标签名称" required>
          <el-input v-model="createTagForm.name" :disabled="createTagForm.versionType !== 'custom'" placeholder="v1.0.0" />
        </el-form-item>
        <el-form-item label="目标引用">
          <el-input v-model="createTagForm.ref" placeholder="HEAD (默认当前分支最新提交)" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="createTagForm.message" type="textarea" :rows="2" placeholder="版本说明" />
        </el-form-item>
        <el-form-item label="推送到远端">
          <el-select v-model="createTagForm.push_remote" placeholder="不推送" clearable>
            <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateTagDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateTag" :loading="createTagLoading">创建</el-button>
      </template>
    </el-dialog>

    <!-- Push Tag Dialog -->
    <el-dialog v-model="showPushTagDialog" :title="`推送标签: ${pushTagName}`" width="480px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="目标远端">
          <el-select v-model="pushTagRemote" placeholder="选择远端">
            <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPushTagDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitPushTag" :loading="pushTagLoading">确认推送</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Share, Switch, Refresh, Edit, Search, Download, Setting, Plus, Top, Delete, CopyDocument, Connection, Lock } from '@element-plus/icons-vue'
import { getRepoDetail, scanRepo, updateRepo, fetchRepo } from '@/api/modules/repo'
import { getSSHKeys, testConnection } from '@/api/modules/system'
import { getStatsAnalyze, getStatsAuthors, getStatsBranches, getStatsCommits, getLineStats, getLineStatsConfig, saveLineStatsConfig, exportStatsCsv } from '@/api/modules/stats'
import { getVersionList, getCurrentVersion, getNextVersion } from '@/api/modules/version'
import type { VersionTag, NextVersionInfo } from '@/api/modules/version'
import { createTag, deleteTag, pushTag } from '@/api/modules/branch'
import type { RepoDTO, ScanResult, GitRemote, AuthInfo, TrackingBranch } from '@/types/repo'
import type { StatsResponse, LineStatsResponse } from '@/types/stats'
import { formatDate, formatRelativeTime } from '@/utils/format'
import GitStatsCharts from '@/components/stats/GitStatsCharts.vue'
import LineStatsCharts from '@/components/stats/LineStatsCharts.vue'
import FileBrowser from '@/components/repo/FileBrowser.vue'
import CommitSearch from '@/components/repo/CommitSearch.vue'
import StashManager from '@/components/repo/StashManager.vue'
import SubmoduleManager from '@/components/repo/SubmoduleManager.vue'

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
const statsAuthors = ref<{ name: string; email: string }[]>([])
const statsData = ref<StatsResponse | null>(null)
const lineStatsData = ref<LineStatsResponse | null>(null)
const commitHistory = ref<{ hash: string; author: string; date: string; message: string }[]>([])

// Versions
const versionList = ref<VersionTag[]>([])
const versionsLoading = ref(false)
const fetchTagsLoading = ref(false)
const remoteNames = ref<string[]>([])

// 合并分支和 tags 供文件浏览和 commit 搜索使用
const allRefs = computed(() => {
  const tags = versionList.value.map(v => v.name)
  return [...statsBranches.value, ...tags]
})

// Create Tag
const showCreateTagDialog = ref(false)
const createTagLoading = ref(false)
const nextVersionInfo = ref<NextVersionInfo | null>(null)
const createTagForm = ref({
  versionType: 'patch' as 'patch' | 'minor' | 'major' | 'custom',
  name: '',
  ref: 'HEAD',
  message: '',
  push_remote: '',
})

// Push Tag
const showPushTagDialog = ref(false)
const pushTagName = ref('')
const pushTagRemote = ref('origin')
const pushTagLoading = ref(false)

// Edit
const showEditDialog = ref(false)
const editSaving = ref(false)
const editActiveTab = ref('basic')
const editForm = ref({ name: '', path: '', config_source: 'local', remote_url: '' })

interface EditRemoteRow extends GitRemote {
  _auth?: AuthInfo
  _testing?: boolean
}
const editRemotes = ref<EditRemoteRow[]>([])
const editTrackingBranches = ref<TrackingBranch[]>([])

// Remote Auth
const showRemoteAuthDialog = ref(false)
const remoteAuthName = ref('')
const remoteAuthIndex = ref(-1)
const remoteAuthForm = ref<AuthInfo>({ type: 'none', key: '', secret: '' })
const sshKeyList = ref<string[]>([])

// Exclude config
const showExcludeDialog = ref(false)
const excludeDirsText = ref('')
const excludePatternsText = ref('')

onMounted(async () => {
  loading.value = true
  try {
    repo.value = await getRepoDetail(repoKey)
    if (repo.value?.path) {
      try {
        scanData.value = await scanRepo(repo.value.path)
        remoteNames.value = (scanData.value?.remotes || []).map((r: { name: string }) => r.name)
      } catch { /* ignore */ }
    }
    try {
      statsBranches.value = await getStatsBranches(repoKey)
      if (statsBranches.value.length > 0) {
        statsFilter.value.branch = statsBranches.value[0]!
        lineStatsFilter.value.branch = statsBranches.value[0]!
      }
    } catch { /* ignore */ }
    try { statsAuthors.value = await getStatsAuthors(repoKey) } catch { /* ignore */ }
    try { currentVersion.value = await getCurrentVersion(repoKey) || '' } catch { /* ignore */ }
    // 加载 tags 用于文件浏览和 commit 搜索
    try { versionList.value = await getVersionList(repoKey) } catch { /* ignore */ }
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

async function handleFetchTags() {
  fetchTagsLoading.value = true
  try {
    await fetchRepo(repoKey)
    ElMessage.success('远端 Tags 拉取成功')
    await loadVersions()
  } catch {
    ElMessage.error('拉取远端 Tags 失败')
  } finally {
    fetchTagsLoading.value = false
  }
}

async function openCreateTagDialog() {
  createTagForm.value = { versionType: 'patch', name: '', ref: 'HEAD', message: '', push_remote: '' }
  nextVersionInfo.value = null
  showCreateTagDialog.value = true
  try {
    nextVersionInfo.value = await getNextVersion(repoKey)
    handleVersionTypeChange(createTagForm.value.versionType)
  } catch { /* ignore */ }
}

function handleVersionTypeChange(type: string | number | boolean) {
  if (!nextVersionInfo.value) return
  switch (type) {
    case 'patch':
      createTagForm.value.name = nextVersionInfo.value.next_patch
      break
    case 'minor':
      createTagForm.value.name = nextVersionInfo.value.next_minor
      break
    case 'major':
      createTagForm.value.name = nextVersionInfo.value.next_major
      break
    case 'custom':
      createTagForm.value.name = ''
      break
  }
}

async function handleCreateTag() {
  if (!createTagForm.value.name) {
    ElMessage.warning('标签名称不能为空')
    return
  }
  createTagLoading.value = true
  try {
    await createTag({
      repo_key: repoKey,
      name: createTagForm.value.name,
      ref: createTagForm.value.ref || 'HEAD',
      message: createTagForm.value.message,
      push_remote: createTagForm.value.push_remote || undefined,
    })
    ElMessage.success(`标签 ${createTagForm.value.name} 创建成功`)
    showCreateTagDialog.value = false
    await loadVersions()
    try { currentVersion.value = await getCurrentVersion(repoKey) || '' } catch { /* ignore */ }
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('创建标签失败: ' + (err.message || '未知错误'))
  } finally {
    createTagLoading.value = false
  }
}

function handlePushTag(tagName: string) {
  pushTagName.value = tagName
  pushTagRemote.value = remoteNames.value[0] || 'origin'
  showPushTagDialog.value = true
}

async function handleSubmitPushTag() {
  pushTagLoading.value = true
  try {
    await pushTag({ repo_key: repoKey, tag_name: pushTagName.value, remote_name: pushTagRemote.value })
    ElMessage.success(`标签 ${pushTagName.value} 已推送到 ${pushTagRemote.value}`)
    showPushTagDialog.value = false
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('推送标签失败: ' + (err.message || '未知错误'))
  } finally {
    pushTagLoading.value = false
  }
}

async function handleDeleteTag(tagName: string) {
  try {
    await ElMessageBox.confirm(`确定要删除标签 "${tagName}" 吗？`, '删除标签', {
      confirmButtonText: '仅删除本地',
      cancelButtonText: '取消',
      type: 'warning',
      distinguishCancelAndClose: true,
    })
    await deleteTag({ repo_key: repoKey, name: tagName })
    ElMessage.success(`标签 ${tagName} 已删除`)
    await loadVersions()
    try { currentVersion.value = await getCurrentVersion(repoKey) || '' } catch { /* ignore */ }
  } catch (action) {
    if (action === 'cancel' || action === 'close') return
    const err = action as { message?: string }
    ElMessage.error('删除标签失败: ' + (err.message || '未知错误'))
  }
}

function handleCopyHash(hash: string) {
  if (hash) {
    navigator.clipboard.writeText(hash)
    ElMessage.success('已复制 Commit Hash')
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
    config_source: repo.value.config_source || 'local',
    remote_url: repo.value.remote_url || '',
  }
  editActiveTab.value = 'basic'
  editRemotes.value = []
  editTrackingBranches.value = []
  showEditDialog.value = true

  // Scan repo to populate remotes & tracking branches
  if (repo.value.path) {
    const storedAuths = repo.value.remote_auths || {}
    scanRepo(repo.value.path).then(result => {
      editRemotes.value = (result.remotes || []).map(r => ({
        ...r,
        _auth: storedAuths[r.name] || { type: 'none', key: '', secret: '' },
        _testing: false,
      }))
      editTrackingBranches.value = result.branches || []
      // Auto-fill remote_url from first remote if empty
      if (!editForm.value.remote_url && editRemotes.value.length > 0) {
        editForm.value.remote_url = editRemotes.value[0]!.fetch_url
      }
    }).catch(() => { /* ignore scan failure */ })
  }
}

async function handleSaveEdit() {
  if (!editForm.value.name || !editForm.value.path) {
    ElMessage.warning('名称和路径不能为空')
    return
  }
  editSaving.value = true
  try {
    const remotes: GitRemote[] = editRemotes.value
      .filter(r => r.name && r.fetch_url)
      .map(r => ({
        name: r.name,
        fetch_url: r.fetch_url,
        push_url: r.push_url || r.fetch_url,
        is_mirror: r.is_mirror,
      }))

    const remoteAuths: Record<string, AuthInfo> = {}
    editRemotes.value.forEach(r => {
      if (r.name && r._auth && r._auth.type !== 'none') {
        remoteAuths[r.name] = { type: r._auth.type, key: r._auth.key, secret: r._auth.secret }
      }
    })

    await updateRepo({
      key: repoKey,
      name: editForm.value.name,
      path: editForm.value.path,
      remote_url: editForm.value.remote_url || undefined,
      config_source: editForm.value.config_source,
      remotes,
      remote_auths: remoteAuths,
    })
    ElMessage.success('保存成功')
    showEditDialog.value = false
    repo.value = await getRepoDetail(repoKey)
    // Refresh scan data
    if (repo.value?.path) {
      try {
        scanData.value = await scanRepo(repo.value.path)
        remoteNames.value = (scanData.value?.remotes || []).map((r: { name: string }) => r.name)
      } catch { /* ignore */ }
    }
  } finally {
    editSaving.value = false
  }
}

function addEditRemote() {
  editRemotes.value.push({
    name: '',
    fetch_url: '',
    push_url: '',
    is_mirror: false,
    _auth: { type: 'none', key: '', secret: '' },
    _testing: false,
  })
}

async function testEditRemote(index: number) {
  const row = editRemotes.value[index]
  if (!row || !row.fetch_url) {
    ElMessage.warning('请输入 Fetch URL')
    return
  }
  row._testing = true
  try {
    const result = await testConnection(row.fetch_url)
    if (result.status === 'success') {
      ElMessage.success(`${row.name || 'Remote'} 连接成功`)
    } else {
      ElMessage.error('连接失败: ' + (result.error || '未知错误'))
    }
  } catch {
    ElMessage.error('连接测试请求失败')
  } finally {
    row._testing = false
  }
}

function openEditRemoteAuth(index: number) {
  const row = editRemotes.value[index]
  if (!row) return
  remoteAuthIndex.value = index
  remoteAuthName.value = row.name || 'New Remote'
  remoteAuthForm.value = { ...(row._auth || { type: 'none', key: '', secret: '' }) }
  showRemoteAuthDialog.value = true
  // Load SSH keys
  getSSHKeys().then(keys => { sshKeyList.value = keys }).catch(() => {})
}

function saveEditRemoteAuth() {
  const row = editRemotes.value[remoteAuthIndex.value]
  if (row) {
    row._auth = { ...remoteAuthForm.value }
  }
  showRemoteAuthDialog.value = false
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