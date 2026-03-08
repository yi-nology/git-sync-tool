<template>
  <div class="problems-panel">
    <div class="panel-header">
      <div class="panel-title">问题</div>
      <div class="problem-counts">
        <span class="error-count">
          <el-icon><CircleCloseFilled /></el-icon>
          {{ errorCount }}
        </span>
        <span class="warning-count">
          <el-icon><WarningFilled /></el-icon>
          {{ warningCount }}
        </span>
        <span class="info-count">
          <el-icon><InfoFilled /></el-icon>
          {{ infoCount }}
        </span>
      </div>
    </div>
    <el-scrollbar v-if="issues.length > 0" class="problems-list">
      <div
        v-for="issue in issues"
        :key="`${issue.line}-${issue.message}`"
        class="problem-item"
        :class="`severity-${issue.severity}`"
        @click="$emit('go-to-line', issue.line, issue.column)"
      >
        <el-icon class="problem-icon">
          <CircleCloseFilled v-if="issue.severity === 'error'" />
          <WarningFilled v-else-if="issue.severity === 'warning'" />
          <InfoFilled v-else />
        </el-icon>
        <div class="problem-content">
          <div class="problem-message">{{ issue.message }}</div>
          <div class="problem-meta">
            <span class="problem-rule">{{ issue.rule_name }}</span>
            <span class="problem-location">行 {{ issue.line }}</span>
          </div>
        </div>
      </div>
    </el-scrollbar>
    <el-empty v-else description="没有问题" :image-size="60" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { CircleCloseFilled, WarningFilled, InfoFilled } from '@element-plus/icons-vue'
import { useSpecStore } from '@/stores/useSpecStore'

defineEmits<{
  'go-to-line': [line: number, column?: number]
}>()

const store = useSpecStore()

const issues = computed(() => store.lintIssues)

const errorCount = computed(() =>
  issues.value.filter((i) => i.severity === 'error').length
)

const warningCount = computed(() =>
  issues.value.filter((i) => i.severity === 'warning').length
)

const infoCount = computed(() =>
  issues.value.filter((i) => i.severity === 'info').length
)
</script>

<style scoped>
.problems-panel {
  height: 100%;
  background: #1e1e1e;
  border-top: 1px solid #333;
  display: flex;
  flex-direction: column;
}

.panel-header {
  padding: 8px 16px;
  border-bottom: 1px solid #333;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel-title {
  font-weight: 500;
  color: #d4d4d4;
}

.problem-counts {
  display: flex;
  gap: 12px;
  font-size: 13px;
}

.error-count,
.warning-count,
.info-count {
  display: flex;
  align-items: center;
  gap: 4px;
}

.error-count {
  color: #f48771;
}

.warning-count {
  color: #cca700;
}

.info-count {
  color: #75beff;
}

.problems-list {
  flex: 1;
  overflow: auto;
}

.problem-item {
  padding: 8px 16px;
  border-bottom: 1px solid #2d2d2d;
  display: flex;
  gap: 8px;
  cursor: pointer;
  transition: background 0.2s;
}

.problem-item:hover {
  background: #2a2d2e;
}

.problem-item.severity-error .problem-icon {
  color: #f48771;
}

.problem-item.severity-warning .problem-icon {
  color: #cca700;
}

.problem-item.severity-info .problem-icon {
  color: #75beff;
}

.problem-content {
  flex: 1;
  min-width: 0;
}

.problem-message {
  color: #d4d4d4;
  font-size: 13px;
  margin-bottom: 4px;
}

.problem-meta {
  display: flex;
  gap: 8px;
  font-size: 12px;
  color: #888;
}

.problem-rule {
  color: #3794ff;
}

.problem-location {
  color: #888;
}

:deep(.el-empty__description p) {
  color: #888;
}
</style>
