<template>
  <el-card shadow="hover" class="remote-card">
    <template #header>
      <div class="remote-header">
        <div class="remote-info">
          <el-tag type="primary" size="small">{{ remote.name }}</el-tag>
          <span v-if="remote.is_mirror" class="mirror-badge">镜像</span>
        </div>
      </div>
    </template>

    <div class="remote-urls">
      <div class="url-row">
        <span class="url-label">Fetch:</span>
        <el-text class="url-value" truncated>{{ remote.fetch_url }}</el-text>
      </div>
      <div v-if="remote.push_url && remote.push_url !== remote.fetch_url" class="url-row">
        <span class="url-label">Push:</span>
        <el-text class="url-value" truncated>{{ remote.push_url }}</el-text>
      </div>
    </div>

    <el-divider />

    <div class="credential-section">
      <span class="section-label">认证凭证</span>
      <CredentialSelector
        v-model="credentialId"
        :url="remote.fetch_url"
        placeholder="选择凭证（可选）"
        @update:model-value="handleCredentialChange"
      />
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { GitRemote } from '@/types/repo'
import CredentialSelector from '@/components/credential/CredentialSelector.vue'

const props = defineProps<{
  remote: GitRemote
  credentialId?: number
}>()

const emit = defineEmits<{
  (e: 'update:credentialId', value: number | undefined): void
}>()

const credentialId = ref<number | undefined>(props.credentialId)

watch(() => props.credentialId, (val) => {
  credentialId.value = val
})

function handleCredentialChange(val: number | undefined) {
  emit('update:credentialId', val)
}
</script>

<style scoped>
.remote-card {
  margin-bottom: 12px;
}
.remote-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.remote-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
.mirror-badge {
  font-size: 12px;
  color: #909399;
}
.remote-urls {
  font-size: 13px;
}
.url-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}
.url-label {
  color: #909399;
  flex-shrink: 0;
  width: 40px;
}
.url-value {
  flex: 1;
  min-width: 0;
}
.credential-section {
  display: flex;
  align-items: center;
  gap: 12px;
}
.section-label {
  font-size: 13px;
  color: #606266;
  flex-shrink: 0;
}
</style>
