<template>
  <div class="repo-list-page">
    <div class="page-header">
      <h2>仓库列表</h2>
      <el-button type="primary" @click="showAddDialog = true">
        <el-icon><Plus /></el-icon> 注册仓库
      </el-button>
    </div>

    <el-table :data="repoStore.repoList" v-loading="repoStore.loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="名称" min-width="120" />
      <el-table-column prop="path" label="路径" min-width="200">
        <template #default="{ row }">
          <el-text type="info" size="small" class="mono-text">{{ row.path }}</el-text>
        </template>
      </el-table-column>
      <el-table-column prop="config_source" label="配置来源" width="110">
        <template #default="{ row }">
          <el-tag size="small" :type="row.config_source === 'database' ? 'warning' : 'info'">
            {{ row.config_source === 'database' ? '数据库' : '本地文件' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="remote_url" label="远程" min-width="180">
        <template #default="{ row }">
          <el-text size="small" truncated>{{ row.remote_url || '-' }}</el-text>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button-group size="small">
            <el-button type="primary" @click="goToDetail(row.key)">
              <el-icon><View /></el-icon> 详情
            </el-button>
            <el-button type="success" @click="goToBranches(row.key)">
              <el-icon><Share /></el-icon> 分支
            </el-button>
            <el-button type="warning" @click="goToSync(row.key)">
              <el-icon><Refresh /></el-icon> 同步
            </el-button>
            <el-button type="danger" @click="handleDelete(row.key, row.name)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>

    <!-- Add Repo Dialog -->
    <el-dialog v-model="showAddDialog" title="注册仓库" width="700px" destroy-on-close>
      <el-tabs v-model="addMode">
        <el-tab-pane label="接入现有仓库" name="local">
          <el-form :model="localForm" label-width="100px">
            <el-form-item label="本地路径" required>
              <el-input v-model="localForm.path" placeholder="/path/to/repo" @blur="handleScanRepo">
                <template #append>
                  <el-button @click="showFileBrowser = true">浏览...</el-button>
                </template>
              </el-input>
              <div class="form-tip">输入路径后将自动扫描 .git/config</div>
            </el-form-item>
            <el-form-item label="仓库名称" required>
              <el-input v-model="localForm.name" placeholder="my-project" />
            </el-form-item>

            <!-- Scan Results -->
            <div v-if="scanResult" class="scan-result">
              <el-divider content-position="left">解析结果</el-divider>
              <el-form-item label="远程仓库">
                <el-table :data="scanResult.remotes" size="small" border>
                  <el-table-column prop="name" label="Name" width="100" />
                  <el-table-column prop="fetch_url" label="Fetch URL" />
                  <el-table-column prop="push_url" label="Push URL" />
                </el-table>
              </el-form-item>
              <el-form-item label="分支追踪">
                <div v-if="scanResult.branches?.length">
                  <el-tag v-for="b in scanResult.branches" :key="b.name" size="small" class="mr-1 mb-1">
                    {{ b.name }} -> {{ b.upstream_ref }}
                  </el-tag>
                </div>
                <el-text v-else type="info">无追踪分支</el-text>
              </el-form-item>
            </div>

            <el-form-item label="配置来源">
              <el-select v-model="localForm.config_source">
                <el-option label="本地配置文件 (.git/config)" value="local" />
                <el-option label="数据库托管" value="database" />
              </el-select>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="克隆新仓库" name="clone">
          <el-form :model="cloneForm" label-width="100px">
            <el-form-item label="远程 URL" required>
              <el-input v-model="cloneForm.remote_url" placeholder="https://github.com/user/repo.git" @change="autoFillCloneName">
                <template #append>
                  <el-button @click="handleTestConnection" :loading="testingConnection">测试连接</el-button>
                </template>
              </el-input>
              <div v-if="connectionResult" :class="['form-tip', connectionResult.success ? 'text-success' : 'text-danger']">
                {{ connectionResult.message }}
              </div>
            </el-form-item>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="本地路径" required>
                  <el-input v-model="cloneForm.local_path" placeholder="/data/repos/my-repo">
                    <template #append>
                      <el-button @click="showFileBrowser = true">浏览...</el-button>
                    </template>
                  </el-input>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="仓库名称" required>
                  <el-input v-model="cloneForm.name" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item label="认证方式">
              <el-select v-model="cloneForm.auth_type">
                <el-option label="无 (公开仓库)" value="none" />
                <el-option label="SSH 密钥" value="ssh" />
                <el-option label="账号密码 (HTTP/HTTPS)" value="http" />
              </el-select>
            </el-form-item>
            <template v-if="cloneForm.auth_type === 'ssh'">
              <el-form-item label="私钥路径">
                <el-select v-model="cloneForm.auth_key" filterable allow-create placeholder="~/.ssh/id_rsa">
                  <el-option v-for="k in sshKeys" :key="k" :label="k" :value="k" />
                </el-select>
              </el-form-item>
              <el-form-item label="密码">
                <el-input v-model="cloneForm.auth_secret" type="password" placeholder="Passphrase (可选)" show-password />
              </el-form-item>
            </template>
            <template v-if="cloneForm.auth_type === 'http'">
              <el-form-item label="用户名">
                <el-input v-model="cloneForm.auth_key" />
              </el-form-item>
              <el-form-item label="密码/Token">
                <el-input v-model="cloneForm.auth_secret" type="password" show-password />
              </el-form-item>
            </template>

            <!-- Clone Progress -->
            <div v-if="cloneProgress.active" class="clone-progress">
              <el-alert :title="cloneProgress.status === 'failed' ? '克隆失败' : '正在克隆...'" :type="cloneProgress.status === 'failed' ? 'error' : 'info'" :closable="false" show-icon />
              <div class="clone-log">
                <div v-for="(line, i) in cloneProgress.logs" :key="i">{{ line }}</div>
              </div>
            </div>
          </el-form>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitRepo" :loading="submitting">保存</el-button>
      </template>
    </el-dialog>

    <!-- File Browser Dialog -->
    <el-dialog v-model="showFileBrowser" title="选择目录" width="600px" destroy-on-close>
      <div class="file-browser">
        <div class="browser-header">
          <el-button @click="loadDirs(dirState.parent)" :icon="Top">上一级</el-button>
          <el-input :model-value="dirState.current" readonly />
        </div>
        <el-input v-model="dirSearch" placeholder="搜索当前目录..." @input="searchDirs" class="mb-2" />
        <div class="dir-list">
          <div
            v-for="d in dirState.dirs"
            :key="d.path"
            class="dir-item"
            @click="loadDirs(d.path)"
          >
            <el-icon><Folder /></el-icon>
            {{ d.name }}
          </div>
          <el-empty v-if="!dirState.dirs?.length" description="无子目录" :image-size="60" />
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="selectDir">选择当前目录</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Plus, Delete, View, Share, Refresh, Top, Folder } from '@element-plus/icons-vue'
import { useRepoStore } from '@/stores/useRepoStore'
import { createRepo, cloneRepo, deleteRepo, scanRepo, getCloneTask } from '@/api/modules/repo'
import { listDirs, getSSHKeys, testConnection } from '@/api/modules/system'
import type { ScanResult } from '@/types/repo'
import type { ListDirsResp } from '@/types/stats'

const router = useRouter()
const repoStore = useRepoStore()

const showAddDialog = ref(false)
const addMode = ref('local')
const submitting = ref(false)
const showFileBrowser = ref(false)
const dirSearch = ref('')
const sshKeys = ref<string[]>([])

const localForm = ref({
  name: '',
  path: '',
  config_source: 'local',
})

const cloneForm = ref({
  remote_url: '',
  local_path: '',
  name: '',
  auth_type: 'none',
  auth_key: '',
  auth_secret: '',
})

const scanResult = ref<ScanResult | null>(null)
const testingConnection = ref(false)
const connectionResult = ref<{ success: boolean; message: string } | null>(null)

const cloneProgress = ref<{ active: boolean; status: string; logs: string[] }>({
  active: false,
  status: '',
  logs: [],
})

const dirState = ref<ListDirsResp>({ parent: '', current: '', dirs: [] })

onMounted(async () => {
  await repoStore.fetchRepoList()
  try {
    sshKeys.value = await getSSHKeys()
  } catch {
    // ignore
  }
})

function goToDetail(key: string) {
  router.push(`/repos/${key}`)
}
function goToBranches(key: string) {
  router.push(`/repos/${key}/branches`)
}
function goToSync(key: string) {
  router.push(`/repos/${key}/sync`)
}

async function handleDelete(key: string, name: string) {
  try {
    await ElMessageBox.confirm(`确定要删除仓库 "${name}" 吗？如果被同步任务使用将无法删除。`, '确认删除', {
      type: 'warning',
    })
    await deleteRepo(key)
    ElMessage.success('仓库已删除')
    await repoStore.fetchRepoList()
  } catch {
    // cancelled or error handled by request
  }
}

async function handleScanRepo() {
  if (!localForm.value.path) return
  if (!localForm.value.name) {
    const parts = localForm.value.path.split('/')
    localForm.value.name = parts[parts.length - 1] || parts[parts.length - 2] || ''
  }
  try {
    scanResult.value = await scanRepo(localForm.value.path)
  } catch {
    scanResult.value = null
  }
}

function autoFillCloneName() {
  if (!cloneForm.value.remote_url) return
  const parts = cloneForm.value.remote_url.split('/')
  let last = parts[parts.length - 1] || ''
  if (last.endsWith('.git')) last = last.slice(0, -4)
  if (!cloneForm.value.name) cloneForm.value.name = last
}

async function handleTestConnection() {
  if (!cloneForm.value.remote_url) {
    ElMessage.warning('请输入 URL')
    return
  }
  testingConnection.value = true
  connectionResult.value = null
  try {
    const data = await testConnection(cloneForm.value.remote_url)
    connectionResult.value = {
      success: data.status === 'success',
      message: data.status === 'success' ? '连接成功' : `失败: ${data.error}`,
    }
  } catch {
    connectionResult.value = { success: false, message: '请求错误' }
  } finally {
    testingConnection.value = false
  }
}

async function handleSubmitRepo() {
  submitting.value = true
  try {
    if (addMode.value === 'local') {
      if (!localForm.value.name || !localForm.value.path) {
        ElMessage.warning('请填写完整信息')
        return
      }
      const remotes = scanResult.value?.remotes || []
      await createRepo({
        name: localForm.value.name,
        path: localForm.value.path,
        config_source: localForm.value.config_source,
        auth_type: 'none',
        remotes,
        remote_auths: {},
      })
      ElMessage.success('仓库注册成功')
      showAddDialog.value = false
      resetForms()
      await repoStore.fetchRepoList()
    } else {
      if (!cloneForm.value.remote_url || !cloneForm.value.local_path) {
        ElMessage.warning('请填写完整克隆信息')
        return
      }
      const result = await cloneRepo({
        remote_url: cloneForm.value.remote_url,
        local_path: cloneForm.value.local_path,
        name: cloneForm.value.name,
        auth_type: cloneForm.value.auth_type === 'none' ? undefined : cloneForm.value.auth_type,
        auth_key: cloneForm.value.auth_key || undefined,
        auth_secret: cloneForm.value.auth_secret || undefined,
        config_source: 'database',
      })
      startClonePolling(result.task_id)
    }
  } finally {
    submitting.value = false
  }
}

function startClonePolling(taskId: string) {
  cloneProgress.value = { active: true, status: 'running', logs: ['Initializing...'] }
  const interval = setInterval(async () => {
    try {
      const task = await getCloneTask(taskId)
      if (task.progress) cloneProgress.value.logs = task.progress
      if (task.status === 'success') {
        clearInterval(interval)
        cloneProgress.value.status = 'success'
        ElMessage.success('克隆成功')
        showAddDialog.value = false
        resetForms()
        await repoStore.fetchRepoList()
      } else if (task.status === 'failed') {
        clearInterval(interval)
        cloneProgress.value.status = 'failed'
        ElMessage.error('克隆失败: ' + task.error)
      }
    } catch {
      // ignore polling errors
    }
  }, 1000)
}

function resetForms() {
  localForm.value = { name: '', path: '', config_source: 'local' }
  cloneForm.value = { remote_url: '', local_path: '', name: '', auth_type: 'none', auth_key: '', auth_secret: '' }
  scanResult.value = null
  connectionResult.value = null
  cloneProgress.value = { active: false, status: '', logs: [] }
}

async function loadDirs(path?: string) {
  try {
    dirState.value = await listDirs(path || '', dirSearch.value)
  } catch {
    // ignore
  }
}

async function searchDirs() {
  await loadDirs(dirState.value.current)
}

function selectDir() {
  if (addMode.value === 'local') {
    localForm.value.path = dirState.value.current
  } else {
    cloneForm.value.local_path = dirState.value.current
  }
  showFileBrowser.value = false
  if (addMode.value === 'local') handleScanRepo()
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
}
.mono-text {
  font-family: monospace;
}
.mr-1 {
  margin-right: 4px;
}
.mb-1 {
  margin-bottom: 4px;
}
.mb-2 {
  margin-bottom: 8px;
}
.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
.text-success {
  color: #67c23a;
}
.text-danger {
  color: #f56c6c;
}
.scan-result {
  background: #f5f7fa;
  padding: 16px;
  border-radius: 6px;
  margin-bottom: 16px;
}
.clone-progress {
  margin-top: 16px;
}
.clone-log {
  background: #1d1e1f;
  color: #d4d4d4;
  padding: 10px;
  border-radius: 4px;
  max-height: 150px;
  overflow-y: auto;
  font-family: monospace;
  font-size: 12px;
  margin-top: 8px;
}
.file-browser .browser-header {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}
.dir-list {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
}
.dir-item {
  padding: 8px 12px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  border-bottom: 1px solid #f0f0f0;
}
.dir-item:hover {
  background: #ecf5ff;
}
</style>
