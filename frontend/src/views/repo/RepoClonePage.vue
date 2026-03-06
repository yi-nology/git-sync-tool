<template>
  <div class="clone-page">
    <h2>克隆远程仓库</h2>

    <el-steps :active="step" align-center class="mb-6">
      <el-step title="仓库地址" description="输入远程 URL" />
      <el-step title="认证配置" description="选择凭证" />
      <el-step title="本地路径" description="确认并克隆" />
    </el-steps>

    <!-- Step 1: URL -->
    <div v-show="step === 0">
      <el-card>
        <el-form label-width="100px">
          <el-form-item label="协议类型">
            <el-radio-group v-model="urlMode" @change="onModeChange">
              <el-radio-button value="ssh">SSH</el-radio-button>
              <el-radio-button value="https">HTTPS</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="远程 URL">
            <el-input
              v-model="form.remote_url"
              :placeholder="urlPlaceholder"
              @blur="onUrlBlur"
              :class="{ 'is-error-input': urlError }"
            >
              <template #prefix>
                <el-tag :type="urlMode === 'ssh' ? 'success' : 'warning'" size="small" class="url-prefix-tag">
                  {{ urlMode === 'ssh' ? 'SSH' : 'HTTPS' }}
                </el-tag>
              </template>
            </el-input>
            <div v-if="urlError" class="field-error">{{ urlError }}</div>
            <div class="url-format-hint">
              <template v-if="urlMode === 'ssh'">格式: <code>git@host:user/repo.git</code> 或 <code>ssh://git@host/path</code></template>
              <template v-else>格式: <code>https://host/user/repo.git</code></template>
            </div>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="goStep2" :disabled="!form.remote_url">下一步</el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>

    <!-- Step 2: Credential -->
    <div v-show="step === 1">
      <el-card>
        <el-form label-width="100px">
          <el-form-item label="认证凭证">
            <CredentialSelector
              v-model="form.credential_id"
              :url="form.remote_url"
              placeholder="选择凭证（公开仓库可不选）"
            />
          </el-form-item>
          <el-form-item>
            <el-text type="info" size="small">
              <template v-if="urlMode === 'ssh'">
                SSH 协议需要配置 SSH 密钥凭证。如果本机已配置 SSH Agent 可跳过。
              </template>
              <template v-else>
                公开仓库可跳过。私有仓库需要配置 HTTP 账号密码或 Token。
              </template>
            </el-text>
          </el-form-item>
          <el-form-item>
            <el-button @click="step = 0">上一步</el-button>
            <el-button type="primary" @click="step = 2">下一步</el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>

    <!-- Step 3: Local path & confirm -->
    <div v-show="step === 2">
      <el-card>
        <el-form :model="form" label-width="100px">
          <el-form-item label="本地路径">
            <el-input v-model="form.local_path" placeholder="/path/to/clone/destination" />
          </el-form-item>
          <el-form-item label="仓库名称">
            <el-input v-model="form.name" placeholder="可选，默认从 URL 推断" />
          </el-form-item>

          <el-divider>确认信息</el-divider>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="协议">
              <el-tag :type="urlMode === 'ssh' ? 'success' : 'warning'" size="small">{{ urlMode === 'ssh' ? 'SSH' : 'HTTPS' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="远程 URL">{{ form.remote_url }}</el-descriptions-item>
            <el-descriptions-item label="本地路径">{{ form.local_path }}</el-descriptions-item>
            <el-descriptions-item label="仓库名称">{{ form.name || '(自动推断)' }}</el-descriptions-item>
            <el-descriptions-item label="认证凭证">
              {{ form.credential_id ? `凭证 #${form.credential_id}` : '无（公开仓库）' }}
            </el-descriptions-item>
          </el-descriptions>

          <el-form-item class="mt-4">
            <el-button @click="step = 1">上一步</el-button>
            <el-button type="primary" @click="handleClone" :loading="cloning" :disabled="!form.remote_url || !form.local_path">
              开始克隆
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>

    <!-- Clone progress -->
    <el-card v-if="taskId" class="mt-4" header="克隆进度">
      <div class="progress-area">
        <el-tag :type="statusTagType" size="small" class="mb-2">{{ statusLabel }}</el-tag>
        <div class="progress-logs">
          <div v-for="(line, i) in progressLines" :key="i" class="log-line">{{ line }}</div>
        </div>
        <el-result v-if="taskStatus === 'done'" icon="success" title="克隆成功">
          <template #extra>
            <el-button type="primary" @click="$router.push('/repos')">查看仓库列表</el-button>
          </template>
        </el-result>
        <el-result v-if="taskStatus === 'failed'" icon="error" title="克隆失败" :sub-title="taskError" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { cloneRepo, getCloneTask } from '@/api/modules/repo'
import type { CloneRepoReq } from '@/types/repo'
import CredentialSelector from '@/components/credential/CredentialSelector.vue'
import { validateGitRemoteUrl, detectGitProtocol, extractRepoName } from '@/utils/git'

type UrlMode = 'ssh' | 'https'

const step = ref(0)
const cloning = ref(false)
const taskId = ref('')
const taskStatus = ref('')
const taskError = ref('')
const progressLines = ref<string[]>([])
let pollTimer: ReturnType<typeof setInterval> | null = null
const urlError = ref('')
const urlMode = ref<UrlMode>('ssh')

const form = ref<CloneRepoReq>({
  remote_url: '',
  local_path: '',
  name: '',
  credential_id: undefined,
})

const urlPlaceholder = computed(() => {
  return urlMode.value === 'ssh'
    ? 'git@github.com:user/repo.git'
    : 'https://github.com/user/repo.git'
})

const statusTagType = computed(() => {
  if (taskStatus.value === 'done') return 'success'
  if (taskStatus.value === 'failed') return 'danger'
  return 'primary'
})

const statusLabel = computed(() => {
  const map: Record<string, string> = {
    running: '克隆中...',
    done: '已完成',
    failed: '失败',
  }
  return map[taskStatus.value] || taskStatus.value || '等待中'
})

function onModeChange() {
  // 切换模式时清空 URL 和错误
  form.value.remote_url = ''
  form.value.name = ''
  urlError.value = ''
}

function onUrlBlur() {
  const url = form.value.remote_url
  if (!url) {
    urlError.value = ''
    return
  }
  // 自动检测协议并同步模式
  const proto = detectGitProtocol(url)
  if (proto === 'ssh') urlMode.value = 'ssh'
  else if (proto === 'http') urlMode.value = 'https'

  // 校验格式
  urlError.value = validateGitRemoteUrl(url)
  // 自动推断名称
  if (!form.value.name) {
    const name = extractRepoName(url)
    if (name) form.value.name = name
  }
}

function goStep2() {
  if (!form.value.remote_url) {
    ElMessage.warning('请输入远程仓库 URL')
    return
  }
  const err = validateGitRemoteUrl(form.value.remote_url)
  if (err) {
    urlError.value = err
    return
  }
  onUrlBlur()
  step.value = 1
}

async function handleClone() {
  if (!form.value.remote_url || !form.value.local_path) {
    ElMessage.warning('请填写远程 URL 和本地路径')
    return
  }

  cloning.value = true
  progressLines.value = []
  taskError.value = ''
  taskStatus.value = 'running'

  try {
    const result = await cloneRepo(form.value)
    taskId.value = result.task_id
    startPolling()
  } catch (e: any) {
    taskStatus.value = 'failed'
    taskError.value = e?.message || '克隆启动失败'
  } finally {
    cloning.value = false
  }
}

function startPolling() {
  pollTimer = setInterval(async () => {
    if (!taskId.value) return
    try {
      const task = await getCloneTask(taskId.value)
      taskStatus.value = task.status
      progressLines.value = task.progress || []
      if (task.error) taskError.value = task.error

      if (task.status === 'done' || task.status === 'failed') {
        stopPolling()
        if (task.status === 'done') {
          ElMessage.success('仓库克隆成功')
        }
      }
    } catch {
      // ignore polling errors
    }
  }, 1500)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.clone-page h2 {
  margin-bottom: 20px;
  font-size: 20px;
}
.mb-6 {
  margin-bottom: 24px;
}
.mb-2 {
  margin-bottom: 8px;
}
.mt-4 {
  margin-top: 16px;
}
.url-prefix-tag {
  margin-right: 4px;
}
.url-format-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
.url-format-hint code {
  background: #f5f7fa;
  padding: 1px 4px;
  border-radius: 2px;
  font-family: monospace;
  font-size: 12px;
}
.progress-area {
  padding: 8px 0;
}
.progress-logs {
  background: #f5f7fa;
  border-radius: 4px;
  padding: 12px;
  margin: 8px 0 16px;
  max-height: 300px;
  overflow-y: auto;
  font-family: monospace;
  font-size: 13px;
}
.log-line {
  line-height: 1.6;
  color: #303133;
}
.field-error {
  color: #f56c6c;
  font-size: 12px;
  margin-top: 4px;
}
.is-error-input :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px #f56c6c inset;
}
</style>
