<template>
  <div class="settings-page">
    <div class="title-row">
      <div class="title-left">
        <h2 class="page-title">系统设置</h2>
        <p class="page-subtitle">管理系统配置和集成服务</p>
      </div>
    </div>

    <div class="card-grid">
      <div class="settings-card">
        <div class="card-icon card-icon--red">
          <el-icon :size="20"><Key /></el-icon>
        </div>
        <h3 class="card-title">SSH 密钥管理</h3>
        <p class="card-desc">管理用于 Git 仓库认证的 SSH 密钥，支持将密钥存储在数据库中。</p>
        <button class="card-btn" @click="$router.push('/settings/ssh-keys')">管理 SSH 密钥</button>
      </div>

      <div class="settings-card">
        <div class="card-icon card-icon--indigo">
          <el-icon :size="20"><Lock /></el-icon>
        </div>
        <h3 class="card-title">凭证管理</h3>
        <p class="card-desc">统一管理 Git 仓库认证凭证，支持 SSH、HTTP 账号密码和 Token。</p>
        <button class="card-btn" @click="$router.push('/settings/credentials')">管理凭证</button>
      </div>

      <div class="settings-card">
        <div class="card-icon card-icon--amber">
          <el-icon :size="20"><Bell /></el-icon>
        </div>
        <h3 class="card-title">通知渠道管理</h3>
        <p class="card-desc">管理系统通知渠道，支持邮件、钉钉、微信等多种通知方式。</p>
        <button class="card-btn" @click="$router.push('/settings/notification-channels')">管理通知渠道</button>
      </div>

      <div class="settings-card">
        <div class="card-icon card-icon--gray">
          <el-icon :size="20"><Setting /></el-icon>
        </div>
        <h3 class="card-title">系统配置</h3>
        <p class="card-desc">调试模式、日志级别等系统级配置开关。</p>
        <div class="toggle-row">
          <span class="toggle-label">调试模式</span>
          <el-switch v-model="config.debug_mode" />
        </div>
      </div>

      <div class="settings-card">
        <div class="card-icon card-icon--green">
          <el-icon :size="20"><Connection /></el-icon>
        </div>
        <h3 class="card-title">全局 Git 配置</h3>
        <div class="git-form">
          <div class="form-field">
            <label class="field-label">Author Name</label>
            <input v-model="config.author_name" placeholder="输入您的 Git 用户名" class="field-input" />
          </div>
          <div class="form-field">
            <label class="field-label">Author Email</label>
            <input v-model="config.author_email" placeholder="输入您的 Git 邮箱" class="field-input" />
          </div>
          <button class="card-btn" @click="handleSave" :disabled="saving">
            {{ saving ? '保存中...' : '保存配置' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Key, Lock, Bell, Setting, Connection } from '@element-plus/icons-vue'
import { getSystemConfig, updateSystemConfig } from '@/api/modules/system'
import type { SystemConfig } from '@/types/stats'

const loading = ref(false)
const saving = ref(false)
const config = ref<SystemConfig>({
  debug_mode: false,
  author_name: '',
  author_email: '',
})

onMounted(async () => {
  loading.value = true
  try {
    const data = await getSystemConfig()
    config.value = data
  } finally {
    loading.value = false
  }
})

async function handleSave() {
  saving.value = true
  try {
    await updateSystemConfig(config.value)
    ElMessage.success('配置保存成功')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.settings-page {
  padding: var(--spacing-xl);
  display: flex;
  flex-direction: column;
  gap: 24px;
  min-height: 100vh;
  background: var(--bg-color);
}

.title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.title-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.page-subtitle {
  margin: 0;
  font-size: 13px;
  color: var(--text-color-secondary);
}

.card-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.settings-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 24px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
}

.card-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: var(--border-radius-md);
}

.card-icon--red {
  background: #FEF2F2;
  color: var(--danger-color);
}

.card-icon--indigo {
  background: #EEF2FF;
  color: var(--primary-color);
}

.card-icon--amber {
  background: #FFFBEB;
  color: var(--warning-color);
}

.card-icon--gray {
  background: #F3F4F6;
  color: var(--text-color-secondary);
}

.card-icon--green {
  background: #ECFDF5;
  color: var(--success-color);
}

.card-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.card-desc {
  margin: 0;
  font-size: 13px;
  color: var(--text-color-secondary);
  line-height: 1.6;
}

.card-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--primary-color);
  color: #FFFFFF;
  font-size: 13px;
  padding: 8px 16px;
  border-radius: var(--border-radius-md);
  border: none;
  cursor: pointer;
  transition: background var(--transition-fast);
  font-family: var(--font-family);
  align-self: flex-start;
}

.card-btn:hover {
  background: var(--primary-color-hover);
}

.card-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.toggle-label {
  font-size: 13px;
  color: var(--text-color-primary);
}

.git-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.field-label {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.field-input {
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-sm);
  padding: 8px 12px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page);
  outline: none;
  font-family: var(--font-family);
  transition: border-color var(--transition-fast);
}

.field-input:focus {
  border-color: var(--primary-color);
}

.field-input::placeholder {
  color: var(--text-color-placeholder);
}

@media (max-width: 768px) {
  .settings-page {
    padding: var(--spacing-md);
  }
}
</style>
