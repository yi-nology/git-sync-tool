<template>
  <div class="app-page">
    <AppBreadcrumb v-if="breadcrumb" :items="breadcrumb" />
    <div class="app-page-header">
      <h1 class="app-page-title">{{ title }}</h1>
      <div class="app-page-actions">
        <slot name="actions"></slot>
      </div>
    </div>
    <div class="app-page-content">
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import AppBreadcrumb from '@/components/common/AppBreadcrumb.vue'

defineProps({
  title: {
    type: String,
    default: ''
  },
  breadcrumb: {
    type: Array,
    default: () => []
  }
})
</script>

<style scoped>
.app-page {
  width: 100%;
  min-height: 100%;
}

.app-page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
  padding-bottom: var(--spacing-md);
  border-bottom: 1px solid var(--border-color);
}

.app-page-title {
  font-size: var(--font-size-xl);
  font-weight: 600;
  color: var(--text-color-primary);
  margin: 0;
}

.app-page-actions {
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
}

.app-page-content {
  width: 100%;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .app-page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }
  
  .app-page-title {
    font-size: var(--font-size-lg);
  }
  
  .app-page-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
