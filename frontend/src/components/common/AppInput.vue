<template>
  <div class="app-input-container">
    <label v-if="label" class="app-input-label">{{ label }}</label>
    <div class="app-input-wrapper">
      <el-icon v-if="prefixIcon" class="app-input-prefix"><component :is="prefixIcon" /></el-icon>
      <input
        type="text"
        class="app-input"
        :placeholder="placeholder"
        :value="modelValue"
        @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
        :disabled="disabled"
        :readonly="readonly"
      />
      <el-icon v-if="suffixIcon" class="app-input-suffix"><component :is="suffixIcon" /></el-icon>
      <button v-if="showClear && modelValue" class="app-input-clear" @click="$emit('update:modelValue', '')">
        <el-icon><Close /></el-icon>
      </button>
    </div>
    <div v-if="help" class="app-input-help">{{ help }}</div>
    <div v-if="error" class="app-input-error">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ElIcon } from 'element-plus'
import { Close } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  label: {
    type: String,
    default: ''
  },
  placeholder: {
    type: String,
    default: ''
  },
  prefixIcon: {
    type: String,
    default: ''
  },
  suffixIcon: {
    type: String,
    default: ''
  },
  showClear: {
    type: Boolean,
    default: true
  },
  disabled: {
    type: Boolean,
    default: false
  },
  readonly: {
    type: Boolean,
    default: false
  },
  help: {
    type: String,
    default: ''
  },
  error: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue'])
</script>

<style scoped>
.app-input-container {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  margin-bottom: var(--spacing-md);
}

.app-input-label {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--text-color-primary);
}

.app-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.app-input {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  padding-left: var(--prefixIcon ? 36px : var(--spacing-md));
  padding-right: var(--suffixIcon || showClear ? 36px : var(--spacing-md));
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-md);
  font-size: var(--font-size-sm);
  background: var(--bg-color-page);
  color: var(--text-color-primary);
  transition: all var(--transition-fast);
  outline: none;
  box-sizing: border-box;
  height: 32px;
}

.app-input:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.app-input::placeholder {
  color: var(--text-color-placeholder);
}

.app-input:disabled {
  background: var(--border-color-extra-light);
  border-color: var(--border-color-light);
  color: var(--text-color-secondary);
  cursor: not-allowed;
}

.app-input:readonly {
  background: var(--border-color-extra-light);
  cursor: default;
}

.app-input-prefix {
  position: absolute;
  left: var(--spacing-sm);
  color: var(--text-color-secondary);
  font-size: 16px;
  pointer-events: none;
}

.app-input-suffix {
  position: absolute;
  right: var(--spacing-sm);
  color: var(--text-color-secondary);
  font-size: 16px;
  pointer-events: none;
}

.app-input-clear {
  position: absolute;
  right: var(--spacing-sm);
  background: none;
  border: none;
  color: var(--text-color-secondary);
  font-size: 14px;
  cursor: pointer;
  padding: 4px;
  border-radius: var(--border-radius-sm);
  transition: all var(--transition-fast);
  display: flex;
  align-items: center;
  justify-content: center;
}

.app-input-clear:hover {
  background: var(--border-color-extra-light);
  color: var(--text-color-primary);
}

.app-input-help {
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
  margin-top: -4px;
}

.app-input-error {
  font-size: var(--font-size-xs);
  color: var(--danger-color);
  margin-top: -4px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .app-input {
    padding: var(--spacing-xs) var(--spacing-sm);
    padding-left: var(--prefixIcon ? 32px : var(--spacing-sm));
    padding-right: var(--suffixIcon || showClear ? 32px : var(--spacing-sm));
    font-size: var(--font-size-xs);
    height: 28px;
  }
  
  .app-input-label {
    font-size: var(--font-size-xs);
  }
  
  .app-input-help,
  .app-input-error {
    font-size: 10px;
  }
}
</style>
