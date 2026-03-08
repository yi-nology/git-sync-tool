<template>
  <div class="error-boundary">
    <slot v-if="!hasError">
      <!-- 正常内容 -->
    </slot>
    <div v-else class="error-content">
      <el-icon :size="48" class="error-icon"><Warning /></el-icon>
      <h3 class="error-title">发生错误</h3>
      <p class="error-message">{{ errorMessage }}</p>
      <el-button type="primary" @click="handleRetry">
        重试
      </el-button>
      <el-button @click="handleReset">
        重置
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onErrorCaptured } from 'vue'
import { Warning } from '@element-plus/icons-vue'

const props = defineProps({
  fallbackMessage: {
    type: String,
    default: '操作过程中发生错误，请稍后重试'
  }
})

const emit = defineEmits<{
  (e: 'retry'): void
  (e: 'reset'): void
}>()

const hasError = ref(false)
const errorMessage = ref(props.fallbackMessage)

/**
 * 捕获错误
 */
const onError = (error: Error, _instance: any, info: string) => {
  hasError.value = true
  errorMessage.value = error.message || props.fallbackMessage
  console.error('Error captured by ErrorBoundary:', error, info)
  return false // 阻止错误继续传播
}

onErrorCaptured(onError)

/**
 * 处理重试
 */
const handleRetry = () => {
  emit('retry')
}

/**
 * 处理重置
 */
const handleReset = () => {
  hasError.value = false
  errorMessage.value = props.fallbackMessage
  emit('reset')
}
</script>

<style scoped>
.error-boundary {
  min-height: 200px;
}

.error-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl);
  text-align: center;
  background: var(--bg-color-page);
  border-radius: var(--border-radius-md);
  box-shadow: var(--box-shadow-sm);
  border: 1px solid var(--border-color);
  margin: var(--spacing-md) 0;
}

.error-icon {
  color: var(--danger-color);
  margin-bottom: var(--spacing-md);
}

.error-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--text-color-primary);
  margin: 0 0 var(--spacing-sm) 0;
}

.error-message {
  font-size: var(--font-size-sm);
  color: var(--text-color-regular);
  margin: 0 0 var(--spacing-lg) 0;
  max-width: 400px;
  line-height: 1.5;
}

/* 响应式 */
@media (max-width: 768px) {
  .error-content {
    padding: var(--spacing-lg);
  }
}
</style>
