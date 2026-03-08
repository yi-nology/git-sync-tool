<template>
  <div class="not-found-page">
    <div class="not-found-content">
      <el-icon :size="80" class="not-found-icon"><DocumentDelete /></el-icon>
      <h1 class="not-found-title">404</h1>
      <h2 class="not-found-subtitle">页面未找到</h2>
      <p class="not-found-message">您访问的页面不存在或已被移除</p>
      <el-button type="primary" @click="router.push('/')">
        <el-icon><House /></el-icon>
        返回首页
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { DocumentDelete, House } from '@element-plus/icons-vue'

const router = useRouter()
</script>

<style scoped>
.not-found-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - var(--header-height));
  padding: var(--spacing-xl);
}

.not-found-content {
  text-align: center;
  background: var(--bg-color-page);
  border-radius: var(--border-radius-lg);
  padding: var(--spacing-xl);
  box-shadow: var(--box-shadow-lg);
  border: 1px solid var(--border-color);
  max-width: 500px;
  width: 100%;
}

.not-found-icon {
  color: var(--info-color);
  margin-bottom: var(--spacing-md);
}

.not-found-title {
  font-size: 48px;
  font-weight: 700;
  color: var(--text-color-primary);
  margin: 0 0 var(--spacing-sm) 0;
}

.not-found-subtitle {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--text-color-primary);
  margin: 0 0 var(--spacing-md) 0;
}

.not-found-message {
  font-size: var(--font-size-sm);
  color: var(--text-color-regular);
  margin: 0 0 var(--spacing-lg) 0;
  line-height: 1.5;
}

/* 响应式 */
@media (max-width: 768px) {
  .not-found-content {
    padding: var(--spacing-lg);
  }
  
  .not-found-title {
    font-size: 36px;
  }
  
  .not-found-icon {
    font-size: 64px;
  }
}
</style>
