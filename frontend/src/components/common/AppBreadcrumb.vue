<template>
  <div class="app-breadcrumb">
    <router-link v-for="(item, index) in items" :key="index" :to="(item as BreadcrumbItem).path" class="app-breadcrumb-item" :class="{ active: index === items.length - 1 }">
      {{ (item as BreadcrumbItem).label }}
    </router-link>
    <span v-for="(_, index) in items.slice(0, -1)" :key="`separator-${index}`" class="app-breadcrumb-separator">
      /
    </span>
  </div>
</template>

<script setup lang="ts">
interface BreadcrumbItem {
  label: string
  path: string
}

defineProps<{
  items: BreadcrumbItem[]
}>()
</script>

<style scoped>
.app-breadcrumb {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-size: var(--font-size-sm);
  margin-bottom: var(--spacing-md);
  padding: var(--spacing-sm) 0;
  border-bottom: 1px solid var(--border-color-light);
}

.app-breadcrumb-item {
  color: var(--text-color-secondary);
  text-decoration: none;
  transition: color var(--transition-fast);
}

.app-breadcrumb-item:hover {
  color: var(--primary-color);
  text-decoration: none;
}

.app-breadcrumb-item.active {
  color: var(--text-color-primary);
  font-weight: 500;
  pointer-events: none;
}

.app-breadcrumb-separator {
  color: var(--text-color-placeholder);
  margin-left: 4px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .app-breadcrumb {
    font-size: var(--font-size-xs);
    gap: var(--spacing-xs);
  }
  
  .app-breadcrumb-item {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 100px;
  }
}
</style>
