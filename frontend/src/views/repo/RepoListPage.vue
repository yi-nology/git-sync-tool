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
                  <el-table-column label="认证" width="80" align="center">
                    <template #default="{ row, $index }">
                      <el-button size="small" :icon="Lock" circle
                        :type="remoteAuths[row.name]?.type && remoteAuths[row.name]?.type !== 'none' ? 'success' : 'default'"
                        @click="openRemoteAuth($index, row.name)" title="配置认证" />
                    </template>
                  </el-table-column>
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
              <el-form-item label="密钥来源">
                <el-radio-group v-model="cloneForm.ssh_source">
                  <el-radio value="local">本地文件</el-radio>
                  <el-radio value="database">数据库密钥</el-radio>
                </el-radio-group>
              </el-form-item>
              <template v-if="cloneForm.ssh_source === 'local'">
                <el-form-item label="SSH 密钥">
                  <el-select v-model="cloneForm.auth_key" filterable allow-create placeholder="~/.ssh/id_rsa" style="width: 100%">
                    <el-option v-for="k in sshKeys" :key="k" :label="k" :value="k" />
                  </el-select>
                </el-form-item>
                <el-form-item label="密钥密码">
                  <el-input v-model="cloneForm.auth_secret" type="password" placeholder="Passphrase (可选)" show-password />
                </el-form-item>
              </template>
              <template v-if="cloneForm.ssh_source === 'database'">
                <el-form-item label="选择密钥">
                  <el-select v-model="cloneForm.ssh_key_id" placeholder="请选择数据库密钥" style="width: 100%">
                    <el-option v-for="k in dbSSHKeyList" :key="k.id" :label="`${k.name} (${k.key_type})`" :value="k.id" />
                  </el-select>
                </el-form-item>
                <div v-if="cloneDbKeyInfo" style="padding: 0 100px; color: #909399; font-size: 12px; margin-bottom: 12px;">
                  {{ cloneDbKeyInfo }}
                </div>
              </template>
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
          <el-form-item label="密钥来源">
            <el-radio-group v-model="remoteAuthForm.source">
              <el-radio value="local">本地文件</el-radio>
              <el-radio value="database">数据库密钥</el-radio>
            </el-radio-group>
          </el-form-item>
          <template v-if="remoteAuthForm.source === 'local'">
            <el-form-item label="SSH 密钥">
              <el-select v-model="remoteAuthForm.key" filterable allow-create placeholder="手动输入路径..." style="width: 100%">
                <el-option v-for="k in sshKeys" :key="k" :label="k" :value="k" />
              </el-select>
            </el-form-item>
            <el-form-item label="密钥密码">
              <el-input v-model="remoteAuthForm.secret" type="password" show-password placeholder="Passphrase (可选)" />
            </el-form-item>
          </template>
          <template v-if="remoteAuthForm.source === 'database'">
            <el-form-item label="选择密钥">
              <el-select v-model="remoteAuthForm.ssh_key_id" placeholder="请选择数据库密钥" style="width: 100%">
                <el-option v-for="k in dbSSHKeyList" :key="k.id" :label="`${k.name} (${k.key_type})`" :value="k.id" />
              </el-select>
            </el-form-item>
            <div v-if="selectedDbKeyInfo" style="padding: 0 110px; color: #909399; font-size: 12px;">
              {{ selectedDbKeyInfo }}
            </div>
          </template>
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
        <el-button type="primary" @click="saveRemoteAuth">确定</el-button>
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
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Plus, Delete, View, Share, Refresh, Top, Folder, Lock } from '@element-plus/icons-vue'
import { useRepoStore } from '@/stores/useRepoStore'
import { createRepo, cloneRepo, deleteRepo, scanRepo, getCloneTask } from '@/api/modules/repo'
import { listDirs, getSSHKeys, testConnection } from '@/api/modules/system'
import { listDBSSHKeys } from '@/api/modules/sshkey'
import type { ScanResult } from '@/types/repo'
import type { AuthInfo } from '@/types/repo'
import type { ListDirsResp } from '@/types/stats'
import type { DBSSHKey } from '@/api/modules/sshkey'

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
})

const cloneForm = ref({
  remote_url: '',
  local_path: '',
  name: '',
  auth_type: 'none',
  auth_key: '',
  auth_secret: '',
  ssh_source: 'local' as 'local' | 'database',
  ssh_key_id: 0,
})

const scanResult = ref<ScanResult | null>(null)
const testingConnection = ref(false)
const connectionResult = ref<{ success: boolean; message: string } | null>(null)

// Remote Auth state
const showRemoteAuthDialog = ref(false)
const remoteAuthName = ref('')
const remoteAuthIndex = ref(-1)
const remoteAuthForm = ref<AuthInfo>({ type: 'none', key: '', secret: '', source: 'local', ssh_key_id: 0 })
const dbSSHKeyList = ref<DBSSHKey[]>([])
const remoteAuths = ref<Record<string, AuthInfo>>({})

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
  try {
    dbSSHKeyList.value = await listDBSSHKeys()
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
        auth_type: 'none',
        remotes,
        remote_auths: remoteAuths.value,
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
      const isDbKey = cloneForm.value.auth_type === 'ssh' && cloneForm.value.ssh_source === 'database'
      const result = await cloneRepo({
        remote_url: cloneForm.value.remote_url,
        local_path: cloneForm.value.local_path,
        name: cloneForm.value.name,
        auth_type: cloneForm.value.auth_type === 'none' ? undefined : cloneForm.value.auth_type,
        auth_key: isDbKey ? undefined : (cloneForm.value.auth_key || undefined),
        auth_secret: isDbKey ? undefined : (cloneForm.value.auth_secret || undefined),
        ssh_key_id: isDbKey ? cloneForm.value.ssh_key_id : undefined,
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
  localForm.value = { name: '', path: '' }
  cloneForm.value = { remote_url: '', local_path: '', name: '', auth_type: 'none', auth_key: '', auth_secret: '', ssh_source: 'local', ssh_key_id: 0 }
  scanResult.value = null
  connectionResult.value = null
  cloneProgress.value = { active: false, status: '', logs: [] }
  remoteAuths.value = {}
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

// Remote Auth functions
function openRemoteAuth(index: number, remoteName: string) {
  remoteAuthIndex.value = index
  remoteAuthName.value = remoteName
  const existing = remoteAuths.value[remoteName]
  remoteAuthForm.value = existing
    ? { ...existing }
    : { type: 'none', key: '', secret: '', source: 'local', ssh_key_id: 0 }
  showRemoteAuthDialog.value = true
  getSSHKeys().then(keys => { sshKeys.value = keys }).catch(() => {})
  listDBSSHKeys().then(keys => { dbSSHKeyList.value = keys }).catch(() => {})
}

function saveRemoteAuth() {
  const name = remoteAuthName.value
  if (remoteAuthForm.value.type === 'none') {
    delete remoteAuths.value[name]
  } else {
    if (remoteAuthForm.value.type === 'ssh' && remoteAuthForm.value.source === 'database') {
      const dbKey = dbSSHKeyList.value.find(k => k.id === remoteAuthForm.value.ssh_key_id)
      if (dbKey) remoteAuthForm.value.key = dbKey.name
    }
    remoteAuths.value[name] = { ...remoteAuthForm.value }
  }
  showRemoteAuthDialog.value = false
}

const selectedDbKeyInfo = computed(() => {
  if (remoteAuthForm.value.source !== 'database' || !remoteAuthForm.value.ssh_key_id) return ''
  const key = dbSSHKeyList.value.find(k => k.id === remoteAuthForm.value.ssh_key_id)
  return key ? (key.description || `创建于 ${key.created_at}`) : ''
})

const cloneDbKeyInfo = computed(() => {
  if (!cloneForm.value.ssh_key_id) return ''
  const key = dbSSHKeyList.value.find(k => k.id === cloneForm.value.ssh_key_id)
  return key ? (key.description || `创建于 ${key.created_at}`) : ''
})
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
