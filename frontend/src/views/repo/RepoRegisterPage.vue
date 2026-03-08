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
        <el-form :model="step1Form" :rules="step1Rules" ref="step1FormRef" label-width="100px">
          <el-form-item label="仓库路径" prop="path">
            <el-input
              v-model="step1Form.path"
              placeholder="/path/to/your/repo"
              clearable
            >
              <template #append>
                <el-button :icon="Folder" @click="handleScan" :loading="scanning">
                  扫描
                </el-button>
              </template>
            </el-input>
          </el-form-item>
        </el-form>
        <el-result v-if="scanError" icon="error" :title="scanError" />
      </el-card>
    </div>

    <!-- Step 2: Info -->
    <div v-show="step === 1">
      <el-card>
        <el-form :model="form" :rules="step2Rules" ref="step2FormRef" label-width="100px">
          <el-form-item label="仓库名称" prop="name">
            <el-input
              v-model="form.name"
              placeholder="仓库名称（必填）"
              clearable
              maxlength="100"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="主远程 URL" prop="remote_url">
            <div class="url-input-group">
              <el-radio-group v-model="urlMode" size="small" @change="onModeChange" class="url-mode-switch">
                <el-radio-button value="ssh">SSH</el-radio-button>
                <el-radio-button value="https">HTTPS</el-radio-button>
              </el-radio-group>
              <el-input
                v-model="form.remote_url"
                :placeholder="urlMode === 'ssh' ? 'git@github.com:user/repo.git' : 'https://github.com/user/repo.git'"
                @blur="validateUrl"
                clearable
              />
            </div>
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
            <el-button type="primary" @click="goStep3">下一步</el-button>
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
import { Folder } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import { scanRepo, createRepo } from '@/api/modules/repo'
import type { ScanResult, RegisterRepoReq } from '@/types/repo'
import CredentialSelector from '@/components/credential/CredentialSelector.vue'
import RemoteCard from '@/components/repo/RemoteCard.vue'
import { validateGitRemoteUrl, detectGitProtocol } from '@/utils/git'
import { useNotification } from '@/composables/useNotification'

const router = useRouter()
const { showSuccess, showError } = useNotification()

const step = ref(0)
const scanning = ref(false)
const scanError = ref('')
const scanResult = ref<ScanResult | null>(null)
const submitting = ref(false)
const remoteCredentials = reactive<Record<string, number | undefined>>({})
const urlMode = ref<'ssh' | 'https'>('ssh')

// Step 1 form
const step1FormRef = ref<FormInstance>()
const step1Form = reactive({
  path: ''
})

const step1Rules: FormRules = {
  path: [
    { required: true, message: '请输入仓库路径', trigger: 'blur' },
    { min: 1, message: '路径不能为空', trigger: 'blur' },
  ]
}

// Step 2 form
const step2FormRef = ref<FormInstance>()
const form = reactive<RegisterRepoReq>({
  name: '',
  path: '',
  remote_url: '',
})

const step2Rules: FormRules = {
  name: [
    { required: true, message: '请输入仓库名称', trigger: 'blur' },
    { min: 1, max: 100, message: '名称长度为 1-100 个字符', trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (!value || !value.trim()) {
          callback(new Error('仓库名称不能为空'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  remote_url: [
    {
      validator: (_rule, value, callback) => {
        if (value && value.trim()) {
          const error = validateGitRemoteUrl(value)
          if (error) {
            callback(new Error(error))
          } else {
            callback()
          }
        } else {
          callback() // 可选字段，允许为空
        }
      },
      trigger: 'blur'
    }
  ]
}

async function handleScan() {
  // 验证表单
  if (!step1FormRef.value) return

  try {
    await step1FormRef.value.validate()
  } catch {
    return
  }

  scanning.value = true
  scanError.value = ''
  try {
    scanResult.value = await scanRepo(step1Form.path)
    form.path = step1Form.path

    // 自动填充名称和远程 URL
    const pathParts = step1Form.path.split('/')
    form.name = pathParts[pathParts.length - 1] || ''
    if (scanResult.value.remotes.length > 0) {
      const origin = scanResult.value.remotes.find(r => r.name === 'origin')
      const autoUrl = origin?.fetch_url || scanResult.value.remotes[0]!.fetch_url
      form.remote_url = autoUrl
      form.remotes = scanResult.value.remotes
      // 自动检测协议模式
      const proto = detectGitProtocol(autoUrl)
      if (proto === 'ssh') urlMode.value = 'ssh'
      else if (proto === 'http') urlMode.value = 'https'
    }
    step.value = 1
    showSuccess('仓库扫描成功')
  } catch (e: any) {
    scanError.value = e?.message || '扫描失败，请检查路径是否为有效的 Git 仓库'
    showError('扫描失败', e)
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
  // 自动触发表单验证
  if (step2FormRef.value) {
    step2FormRef.value.validateField('remote_url')
  }

  // 自动检测协议并同步模式
  if (form.remote_url) {
    const proto = detectGitProtocol(form.remote_url)
    if (proto === 'ssh') urlMode.value = 'ssh'
    else if (proto === 'http') urlMode.value = 'https'
  }
}

function onModeChange() {
  form.remote_url = ''
  if (step2FormRef.value) {
    step2FormRef.value.clearValidate('remote_url')
  }
}

async function goStep3() {
  if (!step2FormRef.value) return

  try {
    await step2FormRef.value.validate()
    step.value = 2
  } catch {
    // 验证失败，Element Plus 会自动显示错误
  }
}

async function handleRegister() {
  if (!step2FormRef.value) return

  try {
    // 再次验证表单
    await step2FormRef.value.validate()
  } catch {
    showError('请完善必填信息')
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
      ...form,
      remote_credentials: Object.keys(rc).length > 0 ? rc : undefined,
    }

    await createRepo(req)
    showSuccess('仓库注册成功')
    router.push('/repos')
  } catch (error: any) {
    showError('注册失败', error)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.register-page {
  padding: 20px;
}

.register-page h2 {
  margin-bottom: 20px;
  font-size: 20px;
  font-weight: 600;
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

.url-input-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.url-mode-switch {
  width: fit-content;
}

.url-format-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.url-format-hint code {
  background-color: #f5f7fa;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: monospace;
}

/* 深色模式 */
:global(.dark) .url-format-hint code {
  background-color: #2c2e30;
}

/* 响应式 */
@media (max-width: 768px) {
  .register-page {
    padding: 16px;
  }

  .url-input-group {
    gap: 12px;
  }
}
</style>
