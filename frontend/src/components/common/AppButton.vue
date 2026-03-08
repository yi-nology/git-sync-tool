<template>
  <button
    class="app-button"
    :class="[
      `app-button-${type}`,
      { 'app-button-disabled': disabled }
    ]"
    :disabled="disabled"
    @click="$emit('click')"
  >
    <el-icon v-if="icon" :size="16"><component :is="icon" /></el-icon>
    <span v-if="$slots.default"><slot></slot></span>
  </button>
</template>

<script setup lang="ts">
import { ElIcon } from 'element-plus'

const props = defineProps({
  type: {
    type: String,
    default: 'primary',
    validator: (value: string) => ['primary', 'secondary', 'danger', 'success', 'warning'].includes(value)
  },
  icon: {
    type: String,
    default: ''
  },
  disabled: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['click'])
</script>

<style scoped>
.app-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-xs);
  border-radius: var(--border-radius-md);
  font-size: var(--font-size-sm);
  font-weight: 500;
  transition: all var(--transition-fast);
  cursor: pointer;
  border: none;
  padding: var(--spacing-sm) var(--spacing-md);
  outline: none;
  min-width: 80px;
  height: 32px;
}

.app-button-primary {
  background: var(--primary-color);
  color: white;
}

.app-button-primary:hover:not(.app-button-disabled) {
  background: #66b1ff;
  transform: translateY(-1px);
  box-shadow: var(--box-shadow-md);
}

.app-button-secondary {
  background: var(--bg-color);
  color: var(--text-color-primary);
  border: 1px solid var(--border-color);
}

.app-button-secondary:hover:not(.app-button-disabled) {
  background: var(--border-color-extra-light);
  border-color: var(--primary-color);
  transform: translateY(-1px);
}

.app-button-danger {
  background: var(--danger-color);
  color: white;
}

.app-button-danger:hover:not(.app-button-disabled) {
  background: #f78989;
  transform: translateY(-1px);
  box-shadow: var(--box-shadow-md);
}

.app-button-success {
  background: var(--success-color);
  color: white;
}

.app-button-success:hover:not(.app-button-disabled) {
  background: #85ce61;
  transform: translateY(-1px);
  box-shadow: var(--box-shadow-md);
}

.app-button-warning {
  background: var(--warning-color);
  color: white;
}

.app-button-warning:hover:not(.app-button-disabled) {
  background: #ebb563;
  transform: translateY(-1px);
  box-shadow: var(--box-shadow-md);
}

.app-button-disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none !important;
  box-shadow: none !important;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .app-button {
    padding: var(--spacing-xs) var(--spacing-sm);
    font-size: var(--font-size-xs);
    min-width: 60px;
    height: 28px;
  }
}
</style>
