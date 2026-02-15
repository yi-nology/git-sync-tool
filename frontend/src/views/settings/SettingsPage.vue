<template>
  <div class="settings-page">
    <h2>系统设置</h2>

    <el-card header="SSH 密钥管理" class="mb-4">
      <p>管理用于 Git 仓库认证的 SSH 密钥，支持将密钥存储在数据库中。</p>
      <el-button type="primary" @click="$router.push('/settings/ssh-keys')">
        <el-icon><Key /></el-icon>
        管理 SSH 密钥
      </el-button>
    </el-card>

    <el-card header="通知渠道管理" class="mb-4">
      <NotificationManager />
    </el-card>

    <el-card header="系统配置" class="mb-4">
      <el-form :model="config" label-width="140px" v-loading="loading">
        <el-form-item label="调试模式">
          <el-switch v-model="config.debug_mode" />
          <div class="form-tip">开启后，系统将在后台日志中输出详细的 Git 命令执行信息。</div>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card header="全局 Git 配置" class="mb-4">
      <el-form :model="config" label-width="140px">
        <el-form-item label="Author Name">
          <el-input v-model="config.author_name" />
        </el-form-item>
        <el-form-item label="Author Email">
          <el-input v-model="config.author_email" type="email" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSave" :loading="saving">保存配置</el-button>
        </el-form-item>
      </el-form>
      <el-text type="info" size="small">设置全局默认的提交作者信息 (git config --global)。如果在仓库中未单独设置，将使用此配置。</el-text>
    </el-card>

    <el-card header="API 文档">
      <p>查看在线 Swagger 文档：<a href="/swagger.html" target="_blank">Swagger UI</a></p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Key } from '@element-plus/icons-vue'
import { getSystemConfig, updateSystemConfig } from '@/api/modules/system'
import type { SystemConfig } from '@/types/stats'
import NotificationManager from '@/components/settings/NotificationManager.vue'

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
.settings-page h2 {
  margin-bottom: 20px;
  font-size: 20px;
}
.mb-4 {
  margin-bottom: 16px;
}
.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
