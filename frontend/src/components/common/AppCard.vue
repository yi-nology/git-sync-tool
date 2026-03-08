<template>
  <div class="app-card">
    <div v-if="title || $slots.header" class="app-card-header">
      <h3 v-if="title" class="app-card-title">{{ title }}</h3>
      <slot name="header"></slot>
    </div>
    <div class="app-card-body">
      <slot></slot>
    </div>
    <div v-if="$slots.footer" class="app-card-footer">
      <slot name="footer"></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps({
  title: {
    type: String,
    default: ''
  }
})
</script>

<style scoped>
.app-card {
  background: var(--bg-color-page);
  border-radius: var(--border-radius-md);
  box-shadow: var(--box-shadow-sm);
  border: 1px solid var(--border-color);
  transition: all var(--transition-normal);
  overflow: hidden;
  margin-bottom: var(--spacing-md);
}

.app-card:hover {
  box-shadow: var(--box-shadow-md);
  transform: translateY(-2px);
}

.app-card-header {
  padding: var(--spacing-md);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--bg-color);
}

.app-card-title {
  font-size: var(--font-size-md);
  font-weight: 600;
  color: var(--text-color-primary);
  margin: 0;
}

.app-card-body {
  padding: var(--spacing-md);
}

.app-card-footer {
  padding: var(--spacing-md);
  border-top: 1px solid var(--border-color);
  background: var(--bg-color);
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-sm);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .app-card {
    border-radius: var(--border-radius-sm);
  }
  
  .app-card-header,
  .app-card-body,
  .app-card-footer {
    padding: var(--spacing-sm);
  }
  
  .app-card-title {
    font-size: var(--font-size-sm);
  }
}
</style>
