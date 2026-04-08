<template>
  <el-card shadow="hover" class="credential-card">
    <template #header>
      <div class="card-header">
        <div class="card-title">
          <el-tag :type="typeTagColor" size="small">{{ typeLabel }}</el-tag>
          <span class="name">{{ credential.name }}</span>
        </div>
        <div class="card-actions">
          <el-button text type="primary" size="small" @click="$emit('edit', credential)">编辑</el-button>
          <el-popconfirm
            title="确定删除该凭证？"
            confirm-button-text="删除"
            cancel-button-text="取消"
            @confirm="$emit('delete', credential)"
          >
            <template #reference>
              <el-button text type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </div>
      </div>
    </template>

    <el-descriptions :column="2" size="small">
      <el-descriptions-item label="类型">{{ typeLabel }}</el-descriptions-item>
      <el-descriptions-item label="密钥/密码">
        <el-tag v-if="credential.has_secret" type="success" size="small">已配置</el-tag>
        <el-tag v-else type="info" size="small">未配置</el-tag>
      </el-descriptions-item>

      <el-descriptions-item v-if="credential.ssh_key_id" label="SSH 密钥">
        <span>{{ credential.ssh_key_name || `#${credential.ssh_key_id}` }}</span>
        <el-tag v-if="credential.ssh_key_type" size="small" :type="sshKeyTypeColor(credential.ssh_key_type)" style="margin-left: 6px">
          {{ sshKeyTypeLabel(credential.ssh_key_type) }}
        </el-tag>
      </el-descriptions-item>
      <el-descriptions-item v-if="credential.ssh_key_path" label="密钥路径">
        {{ credential.ssh_key_path }}
      </el-descriptions-item>
      <el-descriptions-item v-if="credential.username" label="用户名">
        {{ credential.username }}
      </el-descriptions-item>
      <el-descriptions-item v-if="credential.url_pattern" label="URL 匹配">
        <el-tag size="small">{{ credential.url_pattern }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item v-if="credential.last_used_at" label="最后使用">
        {{ formatTime(credential.last_used_at) }}
      </el-descriptions-item>
    </el-descriptions>

    <div v-if="credential.description" class="description">{{ credential.description }}</div>
  </el-card>
</template>

<script setup lang="ts">
import type { CredentialDTO } from '@/types/credential'
import { computed } from 'vue'

const props = defineProps<{
  credential: CredentialDTO
}>()

defineEmits<{
  (e: 'edit', cred: CredentialDTO): void
  (e: 'delete', cred: CredentialDTO): void
}>()

const typeLabel = computed(() => {
  const map: Record<string, string> = {
    ssh_key: 'SSH 密钥',
    http_basic: 'HTTP 账号密码',
    http_token: 'HTTP Token',
  }
  return map[props.credential.type] || props.credential.type
})

const typeTagColor = computed(() => {
  const map: Record<string, string> = {
    ssh_key: 'success',
    http_basic: 'warning',
    http_token: '',
  }
  return map[props.credential.type] || 'info'
})

function formatTime(t: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString()
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
</script>

<style scoped>
.credential-card {
  margin-bottom: 12px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
}
.card-title .name {
  font-weight: 600;
  font-size: 14px;
}
.card-actions {
  display: flex;
  gap: 4px;
}
.description {
  margin-top: 8px;
  font-size: 13px;
  color: #606266;
}
</style>
