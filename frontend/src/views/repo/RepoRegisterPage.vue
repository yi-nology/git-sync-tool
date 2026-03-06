<template>
  <div class="register-page">
    <h2>注册本地仓库</h2>

    <el-steps :active="step" align-center class="mb-6">
      <el-step title="扫描仓库" description="输入本地路径" />
      <el-step title="配置信息" description="名称和远程" />
      <el-step title="认证配置" description="选择凭证" />
    </el-steps>

    <!-- Step 1: Scan -->
    <div v-show="step === 0">
      <el-card>
        <el-form label-width="100px">
          <el-form-item label="仓库路径">
            <el-input v-model="repoPath" placeholder="/path/to/your/repo" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleScan" :loading="scanning">扫描仓库</el-button>
          </el-form-item>
        </el-form>
        <el-result v-if="scanError" icon="error" :title="scanError" />
      </el-card>
    </div>

    <!-- Step 2: Info -->
    <div v-show="step === 1">
      <el-card>
        <el-form :model="form" label-width="100px">
          <el-form-item label="仓库名称">
            <el-input v-model="form.name" placeholder="仓库名称" />
          </el-form-item>
          <el-form-item label="主远程 URL">
            <div class="url-input-group">
              <el-radio-group v-model="urlMode" size="small" @change="onModeChange" class="url-mode-switch">
                <el-radio-button value="ssh">SSH</el-radio-button>
                <el-radio-button value="https">HTTPS</el-radio-button>
              </el-radio-group>
              <el-input
                v-model="form.remote_url"
                :placeholder="urlMode === 'ssh' ? 'git@github.com:user/repo.git' : 'https://github.com/user/repo.git'"
                @blur="validateUrl"
                :class="{ 'is-error-input': urlError }"
              />
            </div>
            <div v-if="urlError" class="field-error">{{ urlError }}</div>
            <div class="url-format-hint">
              <template v-if="urlMode === 'ssh'">格式: <code>git@host:user/repo.git</code> 或 <code>ssh://git@host/path</code></template>
              <template v-else>格式: <code>https://host/user/repo.git</code></template>
            </div>
          </el-form-item>

          <el-divider>检测到的远程仓库</el-divider>
          <div v-if="scanResult && scanResult.remotes.length > 0">
            <div v-for="remote in scanResult.remotes" :key="remote.name" class="remote-item">
              <el-tag>{{ remote.name }}</el-tag>
              <span class="remote-url">{{ remote.fetch_url }}</span>
            </div>
          </div>
          <el-empty v-else description="未检测到远程仓库" :image-size="60" />

          <el-form-item class="mt-4">
            <el-button @click="step = 0">上一步</el-button>
            <el-button type="primary" @click="goStep3" :disabled="!!urlError">下一步</el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>

    <!-- Step 3: Credential -->
    <div v-show="step === 2">
      <el-card>
        <el-form label-width="120px">
          <el-form-item label="默认凭证">
            <CredentialSelector
              v-model="form.default_credential_id"
              :url="form.remote_url"
              placeholder="选择默认认证凭证（可选）"
            />
          </el-form-item>

          <template v-if="scanResult && scanResult.remotes.length > 0">
            <el-divider>各远程凭证配置</el-divider>
            <div v-for="remote in scanResult.remotes" :key="remote.name">
              <RemoteCard
                :remote="remote"
                :credential-id="remoteCredentials[remote.name]"
                @update:credential-id="(v) => updateRemoteCred(remote.name, v)"
              />
            </div>
          </template>

          <el-form-item class="mt-4">
            <el-button @click="step = 1">上一步</el-button>
            <el-button type="primary" @click="handleRegister" :loading="submitting">注册仓库</el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { scanRepo, createRepo } from '@/api/modules/repo'
import type { ScanResult, RegisterRepoReq } from '@/types/repo'
import CredentialSelector from '@/components/credential/CredentialSelector.vue'
import RemoteCard from '@/components/repo/RemoteCard.vue'
import { validateGitRemoteUrl, detectGitProtocol } from '@/utils/git'

const router = useRouter()
const step = ref(0)
const repoPath = ref('')
const scanning = ref(false)
const scanError = ref('')
const scanResult = ref<ScanResult | null>(null)
const submitting = ref(false)
const remoteCredentials = reactive<Record<string, number | undefined>>({})
const urlError = ref('')
const urlMode = ref<'ssh' | 'https'>('ssh')

const form = ref<RegisterRepoReq>({
  name: '',
  path: '',
  remote_url: '',
})

async function handleScan() {
  if (!repoPath.value) return
  scanning.value = true
  scanError.value = ''
  try {
    scanResult.value = await scanRepo(repoPath.value)
    form.value.path = repoPath.value

    // 自动填充名称和远程 URL
    const pathParts = repoPath.value.split('/')
    form.value.name = pathParts[pathParts.length - 1] || ''
    if (scanResult.value.remotes.length > 0) {
      const origin = scanResult.value.remotes.find(r => r.name === 'origin')
      const autoUrl = origin?.fetch_url || scanResult.value.remotes[0]!.fetch_url
      form.value.remote_url = autoUrl
      form.value.remotes = scanResult.value.remotes
      // 自动检测协议模式
      const proto = detectGitProtocol(autoUrl)
      if (proto === 'ssh') urlMode.value = 'ssh'
      else if (proto === 'http') urlMode.value = 'https'
    }
    step.value = 1
  } catch (e: any) {
    scanError.value = e?.message || '扫描失败，请检查路径是否为有效的 Git 仓库'
  } finally {
    scanning.value = false
  }
}

function updateRemoteCred(name: string, val: number | undefined) {
  if (val) {
    remoteCredentials[name] = val
  } else {
    delete remoteCredentials[name]
  }
}

function validateUrl() {
  const url = form.value.remote_url
  if (!url) {
    urlError.value = ''
    return
  }
  // 自动检测协议并同步模式
  const proto = detectGitProtocol(url)
  if (proto === 'ssh') urlMode.value = 'ssh'
  else if (proto === 'http') urlMode.value = 'https'
  urlError.value = validateGitRemoteUrl(url)
}

function onModeChange() {
  form.value.remote_url = ''
  urlError.value = ''
}

function goStep3() {
  if (form.value.remote_url) {
    const err = validateGitRemoteUrl(form.value.remote_url)
    if (err) {
      urlError.value = err
      return
    }
  }
  step.value = 2
}

async function handleRegister() {
  if (!form.value.name || !form.value.path) {
    ElMessage.warning('请填写仓库名称和路径')
    return
  }

  submitting.value = true
  try {
    // 组装 remote_credentials
    const rc: Record<string, number> = {}
    for (const [k, v] of Object.entries(remoteCredentials)) {
      if (v) rc[k] = v
    }

    const req: RegisterRepoReq = {
      ...form.value,
      remote_credentials: Object.keys(rc).length > 0 ? rc : undefined,
    }

    await createRepo(req)
    ElMessage.success('仓库注册成功')
    router.push('/repos')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.register-page h2 {
  margin-bottom: 20px;
  font-size: 20px;
}
.mb-6 {
  margin-bottom: 24px;
}
.mt-4 {
  margin-top: 16px;
}
.remote-item {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}
.remote-url {
  color: #606266;
  font-size: 13px;
  word-break: break-all;
}
.field-error {
  color: #f56c6c;
  font-size: 12px;
  margin-top: 4px;
}
.is-error-input :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px #f56c6c inset;
}
.url-input-group {
  display: flex;
  gap: 8px;
  width: 100%;
}
.url-mode-switch {
  flex-shrink: 0;
}
.url-input-group .el-input {
  flex: 1;
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
</style>
