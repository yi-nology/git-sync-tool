<template>
  <el-form
    ref="formRef"
    :model="form"
    :rules="rules"
    :label-width="labelWidth"
    :disabled="disabled"
    class="app-form"
  >
    <slot></slot>
    
    <!-- 操作按钮区域 -->
    <el-form-item v-if="showActions" class="form-actions">
      <el-button @click="$emit('cancel')" :disabled="loading">
        {{ cancelText }}
      </el-button>
      <el-button 
        type="primary" 
        @click="handleSubmit" 
        :loading="loading"
      >
        {{ submitText }}
      </el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'

const props = defineProps({
  form: {
    type: Object,
    required: true
  },
  rules: {
    type: Object as () => FormRules,
    default: () => {}
  },
  labelWidth: {
    type: String,
    default: '120px'
  },
  disabled: {
    type: Boolean,
    default: false
  },
  loading: {
    type: Boolean,
    default: false
  },
  showActions: {
    type: Boolean,
    default: true
  },
  submitText: {
    type: String,
    default: '提交'
  },
  cancelText: {
    type: String,
    default: '取消'
  }
})

const emit = defineEmits<{
  (e: 'submit'): void
  (e: 'cancel'): void
}>()

const formRef = ref<FormInstance>()

/**
 * 处理表单提交
 */
const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    emit('submit')
  } catch (error) {
    // 验证失败，不触发提交
  }
}
</script>

<style scoped>
.app-form {
  background: var(--bg-color-page);
  border-radius: var(--border-radius-md);
  padding: var(--spacing-lg);
  box-shadow: var(--box-shadow-sm);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-sm);
  margin-top: var(--spacing-lg);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--border-color);
}

/* 响应式 */
@media (max-width: 768px) {
  .app-form {
    padding: var(--spacing-md);
  }
  
  .form-actions {
    flex-direction: column;
    align-items: stretch;
  }
  
  .form-actions .el-button {
    width: 100%;
  }
}
</style>
