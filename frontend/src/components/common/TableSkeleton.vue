<template>
  <div class="table-skeleton">
    <!-- 表头骨架 -->
    <div class="skeleton-header">
      <el-skeleton-item
        v-for="i in columns"
        :key="'header-' + i"
        variant="text"
        :style="{ width: getColumnWidth(i) }"
      />
    </div>

    <!-- 表格行骨架 -->
    <div v-for="row in rows" :key="'row-' + row" class="skeleton-row">
      <el-skeleton-item
        v-for="col in columns"
        :key="'cell-' + row + '-' + col"
        variant="text"
        :style="{ width: getColumnWidth(col) }"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  rows?: number
  columns?: number
  columnWidths?: string[]
}

const props = withDefaults(defineProps<Props>(), {
  rows: 5,
  columns: 5,
  columnWidths: () => ['80px', '150px', '200px', '150px', '120px'],
})

const getColumnWidth = (index: number): string => {
  return props.columnWidths[index - 1] || '100px'
}
</script>

<style scoped>
.table-skeleton {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  overflow: hidden;
}

.skeleton-header {
  display: flex;
  gap: 16px;
  padding: 12px 16px;
  background-color: var(--el-fill-color-light);
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.skeleton-row {
  display: flex;
  gap: 16px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.skeleton-row:last-child {
  border-bottom: none;
}
</style>
