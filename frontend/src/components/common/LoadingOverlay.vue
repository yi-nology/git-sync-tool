<template>
  <Teleport to="body">
    <Transition name="fade">
      <div v-if="visible" class="loading-overlay">
        <div class="loading-content">
          <el-icon :size="48" class="is-loading"><Loading /></el-icon>
          <p v-if="text" class="loading-text">{{ text }}</p>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { Loading } from '@element-plus/icons-vue'

interface Props {
  visible?: boolean
  text?: string
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  text: '',
})
</script>

<style scoped>
.loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--z-index-modal);
}

:global(.dark) .loading-overlay {
  background: rgba(0, 0, 0, 0.7);
}

.loading-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  background: var(--bg-color-page);
  border-radius: var(--border-radius-lg);
  box-shadow: var(--box-shadow-lg);
  border: 1px solid var(--border-color);
}

.loading-text {
  font-size: var(--font-size-sm);
  color: var(--text-color-regular);
  margin: 0;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity var(--transition-normal);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
