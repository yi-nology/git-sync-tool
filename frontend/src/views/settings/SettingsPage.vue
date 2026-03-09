<template>
  <AppPage title="系统设置">
    <div class="settings-grid">
      <!-- SSH 密钥管理 -->
      <AppCard title="SSH 密钥管理">
        <div class="settings-card-content">
          <p class="settings-description">管理用于 Git 仓库认证的 SSH 密钥，支持将密钥存储在数据库中。</p>
          <AppButton type="primary" icon="Key" @click="$router.push('/settings/ssh-keys')">
            管理 SSH 密钥
          </AppButton>
        </div>
      </AppCard>

      <!-- 凭证管理 -->
      <AppCard title="凭证管理">
        <div class="settings-card-content">
          <p class="settings-description">统一管理 Git 仓库认证凭证，支持 SSH 密钥、HTTP 账号密码和 Token。配置 URL 匹配模式后可自动推荐。</p>
          <AppButton type="primary" icon="Lock" @click="$router.push('/settings/credentials')">
            管理凭证
          </AppButton>
        </div>
      </AppCard>

      <!-- 通知渠道管理 -->
      <AppCard title="通知渠道管理">
        <div class="settings-card-content">
          <p class="settings-description">管理系统通知渠道，支持邮件、钉钉、微信等多种通知方式，可配置通知触发事件和消息模板。</p>
          <AppButton type="primary" icon="Bell" @click="$router.push('/settings/notification-channels')">
            管理通知渠道
          </AppButton>
        </div>
      </AppCard>

      <!-- 系统配置 -->
      <AppCard title="系统配置">
        <div class="settings-form">
          <div class="form-item">
            <label class="form-label">调试模式</label>
            <el-switch v-model="config.debug_mode" />
            <div class="form-tip">开启后，系统将在后台日志中输出详细的 Git 命令执行信息。</div>
          </div>
        </div>
      </AppCard>

      <!-- 全局 Git 配置 -->
      <AppCard title="全局 Git 配置">
        <div class="settings-form">
          <AppInput
            v-model="config.author_name"
            label="Author Name"
            placeholder="输入您的 Git 用户名"
          />
          <AppInput
            v-model="config.author_email"
            label="Author Email"
            placeholder="输入您的 Git 邮箱"
            type="email"
          />
          <div class="form-actions">
            <AppButton type="primary" @click="handleSave" :disabled="saving">
              {{ saving ? '保存中...' : '保存配置' }}
            </AppButton>
          </div>
          <p class="settings-info">设置全局默认的提交作者信息 (git config --global)。如果在仓库中未单独设置，将使用此配置。</p>
        </div>
      </AppCard>
    </div>
  </AppPage>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Key, Lock, Bell } from '@element-plus/icons-vue'
import { getSystemConfig, updateSystemConfig } from '@/api/modules/system'
import type { SystemConfig } from '@/types/stats'
import AppPage from '@/components/layout/AppPage.vue'
import AppCard from '@/components/common/AppCard.vue'
import AppButton from '@/components/common/AppButton.vue'
import AppInput from '@/components/common/AppInput.vue'

// 图标被模板中的 icon prop 引用
void Key
void Lock
void Bell

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
.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: var(--spacing-md);
  margin-top: var(--spacing-md);
}

.settings-card-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.settings-description {
  color: var(--text-color-regular);
  line-height: 1.5;
  margin: 0;
}

.settings-form {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.form-label {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--text-color-primary);
}

.form-tip {
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
  margin-top: 4px;
}

.form-actions {
  display: flex;
  gap: var(--spacing-sm);
  margin-top: var(--spacing-sm);
}

.settings-info {
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
  margin-top: var(--spacing-sm);
  line-height: 1.4;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .settings-grid {
    grid-template-columns: 1fr;
  }
  
  .settings-card-content {
    gap: var(--spacing-sm);
  }
  
  .form-actions {
    justify-content: flex-start;
  }
}
</style>
