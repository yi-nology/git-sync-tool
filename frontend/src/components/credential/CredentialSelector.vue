<template>
  <div class="credential-selector">
    <el-select
      v-model="selected"
      :placeholder="placeholder"
      clearable
      filterable
      :loading="loading"
      @change="handleChange"
      @visible-change="onDropdownToggle"
      style="width: 100%"
    >
      <el-option-group v-if="recommended.length > 0" label="推荐凭证">
        <el-option
          v-for="cred in recommended"
          :key="cred.id"
          :label="cred.name"
          :value="cred.id"
        >
          <div class="cred-option">
            <el-tag :type="typeTagColor(cred.type)" size="small" class="cred-tag">{{ typeLabel(cred.type) }}</el-tag>
            <span class="cred-name">{{ cred.name }}</span>
          </div>
        </el-option>
      </el-option-group>
      <el-option-group v-if="others.length > 0" :label="recommended.length > 0 ? '其他凭证' : '可用凭证'">
        <el-option
          v-for="cred in others"
          :key="cred.id"
          :label="cred.name"
          :value="cred.id"
        >
          <div class="cred-option">
            <el-tag :type="typeTagColor(cred.type)" size="small" class="cred-tag">{{ typeLabel(cred.type) }}</el-tag>
            <span class="cred-name">{{ cred.name }}</span>
          </div>
        </el-option>
      </el-option-group>
      <template #empty>
        <div class="empty-hint">
          暂无凭证，<el-link type="primary" @click="$router.push('/settings/credentials')">去创建</el-link>
        </div>
      </template>
    </el-select>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { listCredentials, matchCredentials } from '@/api/modules/credential'
import type { CredentialDTO, CredentialType } from '@/types/credential'

const props = defineProps<{
  modelValue?: number
  url?: string
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: number | undefined): void
}>()

const selected = ref<number | undefined>(props.modelValue)
const loading = ref(false)
const recommended = ref<CredentialDTO[]>([])
const others = ref<CredentialDTO[]>([])

watch(() => props.modelValue, (val) => {
  selected.value = val
})

function handleChange(val: number | undefined) {
  emit('update:modelValue', val)
}

async function loadCredentials() {
  loading.value = true
  try {
    if (props.url) {
      const resp = await matchCredentials(props.url)
      recommended.value = resp.recommended || []
      others.value = resp.others || []
    } else {
      const all = await listCredentials()
      recommended.value = []
      others.value = all || []
    }
  } finally {
    loading.value = false
  }
}

function onDropdownToggle(visible: boolean) {
  if (visible) {
    loadCredentials()
  }
}

function typeLabel(type: CredentialType): string {
  const map: Record<string, string> = {
    ssh_key: 'SSH',
    http_basic: 'HTTP',
    http_token: 'Token',
  }
  return map[type] || type
}

function typeTagColor(type: CredentialType): string {
  const map: Record<string, string> = {
    ssh_key: 'success',
    http_basic: 'warning',
    http_token: '',
  }
  return map[type] || 'info'
}

onMounted(() => {
  loadCredentials()
})
</script>

<style scoped>
.cred-option {
  display: flex;
  align-items: center;
  gap: 8px;
}
.cred-tag {
  flex-shrink: 0;
}
.cred-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.empty-hint {
  padding: 10px;
  text-align: center;
  color: var(--text-color-secondary);
  font-size: var(--font-size-sm);
}
</style>
