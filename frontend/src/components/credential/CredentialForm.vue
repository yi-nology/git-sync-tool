<template>
  <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
    <el-form-item label="名称" prop="name">
      <el-input v-model="form.name" placeholder="例如: GitHub SSH Key" />
    </el-form-item>

    <el-form-item label="类型" prop="type" v-if="!isEdit">
      <el-radio-group v-model="form.type" @change="onTypeChange">
        <el-radio value="ssh_key">SSH 密钥</el-radio>
        <el-radio value="http_basic">HTTP 账号密码</el-radio>
        <el-radio value="http_token">HTTP Token</el-radio>
      </el-radio-group>
    </el-form-item>

    <el-form-item v-if="isEdit" label="类型">
      <el-tag :type="typeTagColor">{{ typeLabel }}</el-tag>
    </el-form-item>

    <el-form-item label="描述">
      <el-input v-model="form.description" type="textarea" :rows="2" placeholder="可选描述信息" />
    </el-form-item>

    <!-- SSH Key 特有字段 -->
    <template v-if="form.type === 'ssh_key'">
      <el-form-item label="密钥来源">
        <el-radio-group v-model="sshSource">
          <el-radio value="database">数据库密钥</el-radio>
          <el-radio value="local">本地文件</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item v-if="sshSource === 'database'" label="SSH 密钥" prop="ssh_key_id">
        <el-select v-model="form.ssh_key_id" placeholder="选择数据库中的 SSH 密钥" style="width: 100%">
          <el-option
            v-for="key in sshKeys"
            :key="key.id"
            :label="key.name"
            :value="key.id"
          >
            <span>{{ key.name }}</span>
            <el-tag v-if="key.key_type" size="small" :type="sshKeyTypeColor(key.key_type)" style="margin-left: 8px">
              {{ sshKeyTypeLabel(key.key_type) }}
            </el-tag>
          </el-option>
        </el-select>
      </el-form-item>

      <el-form-item v-if="sshSource === 'local'" label="密钥路径" prop="ssh_key_path">
        <el-input v-model="form.ssh_key_path" placeholder="例如: ~/.ssh/id_rsa" />
      </el-form-item>

      <el-form-item v-if="sshSource === 'local'" label="密码短语">
        <el-input v-model="form.secret" type="password" show-password placeholder="可选 passphrase" />
      </el-form-item>
    </template>

    <!-- HTTP 特有字段 -->
    <template v-if="form.type === 'http_basic' || form.type === 'http_token'">
      <el-form-item label="用户名" :prop="form.type === 'http_basic' ? 'username' : undefined">
        <el-input v-model="form.username" :placeholder="form.type === 'http_token' ? '通常不需要，或填写用户名' : '用户名'" />
      </el-form-item>

      <el-form-item :label="form.type === 'http_token' ? 'Token' : '密码'" prop="secret">
        <el-input v-model="form.secret" type="password" show-password :placeholder="form.type === 'http_token' ? 'Personal Access Token' : '密码'" />
      </el-form-item>
    </template>

    <el-form-item label="URL 匹配">
      <el-input v-model="form.url_pattern" placeholder="例如: *.github.com（可选，用于智能推荐）" />
      <div class="form-tip">当仓库 URL 匹配此模式时，系统将自动推荐该凭证。支持 * 通配符。</div>
    </el-form-item>

    <el-form-item>
      <el-button type="primary" @click="handleSubmit" :loading="submitting">
        {{ isEdit ? '保存' : '创建' }}
      </el-button>
      <el-button @click="$emit('cancel')">取消</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { createCredential, updateCredential } from '@/api/modules/credential'
import { listDBSSHKeys, type DBSSHKey } from '@/api/modules/sshkey'
import type { CredentialDTO, CreateCredentialReq } from '@/types/credential'

const props = defineProps<{
  credential?: CredentialDTO
}>()

const emit = defineEmits<{
  (e: 'success', cred: CredentialDTO): void
  (e: 'cancel'): void
}>()

const isEdit = computed(() => !!props.credential)
const formRef = ref<FormInstance>()
const submitting = ref(false)
const sshKeys = ref<DBSSHKey[]>([])
const sshSource = ref<'database' | 'local'>('database')

const form = ref<CreateCredentialReq>({
  name: '',
  type: 'ssh_key',
  description: '',
  ssh_key_id: undefined,
  ssh_key_path: '',
  username: '',
  secret: '',
  url_pattern: '',
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入凭证名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择凭证类型', trigger: 'change' }],
}

const typeLabel = computed(() => {
  const map: Record<string, string> = { ssh_key: 'SSH 密钥', http_basic: 'HTTP 账号密码', http_token: 'HTTP Token' }
  return map[form.value.type] || form.value.type
})

const typeTagColor = computed(() => {
  const map: Record<string, string> = { ssh_key: 'success', http_basic: 'warning', http_token: '' }
  return map[form.value.type] || 'info'
})

function onTypeChange() {
  form.value.ssh_key_id = undefined
  form.value.ssh_key_path = ''
  form.value.username = ''
  form.value.secret = ''
}

watch(() => props.credential, (cred) => {
  if (cred) {
    form.value = {
      name: cred.name,
      type: cred.type,
      description: cred.description || '',
      ssh_key_id: cred.ssh_key_id,
      ssh_key_path: cred.ssh_key_path || '',
      username: cred.username || '',
      secret: '',
      url_pattern: cred.url_pattern || '',
    }
    sshSource.value = cred.ssh_key_id ? 'database' : 'local'
  }
}, { immediate: true })

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()

  submitting.value = true
  try {
    const data: CreateCredentialReq = { ...form.value }
    // 根据 SSH 来源清理无关字段
    if (data.type === 'ssh_key') {
      if (sshSource.value === 'database') {
        data.ssh_key_path = ''
        if (!data.ssh_key_id) data.secret = ''
      } else {
        data.ssh_key_id = undefined
      }
      data.username = ''
    } else {
      data.ssh_key_id = undefined
      data.ssh_key_path = ''
    }

    let result: CredentialDTO
    if (isEdit.value && props.credential) {
      result = await updateCredential(props.credential.id, data)
    } else {
      result = await createCredential(data)
    }
    ElMessage.success(isEdit.value ? '凭证已更新' : '凭证已创建')
    emit('success', result)
  } finally {
    submitting.value = false
  }
}

const SSH_KEY_TYPE_LABELS: Record<string, string> = {
  rsa: 'RSA', ed25519: 'Ed25519', ecdsa: 'ECDSA', dsa: 'DSA', unknown: '未知',
}
const SSH_KEY_TYPE_COLORS: Record<string, string> = {
  rsa: 'warning', ed25519: 'success', ecdsa: '', dsa: 'info', unknown: 'info',
}
function sshKeyTypeLabel(t: string): string {
  return SSH_KEY_TYPE_LABELS[t?.toLowerCase()] ?? t?.toUpperCase() ?? ''
}
function sshKeyTypeColor(t: string): string {
  return SSH_KEY_TYPE_COLORS[t?.toLowerCase()] ?? ''
}

onMounted(async () => {
  try {
    sshKeys.value = await listDBSSHKeys()
  } catch {
    // ignore
  }
})
</script>

<style scoped>
.form-tip {
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
  margin-top: var(--spacing-xs);
  line-height: 1.5;
}
</style>
