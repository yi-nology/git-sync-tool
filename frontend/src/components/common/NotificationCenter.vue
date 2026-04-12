<template>
  <Teleport to="body">
    <Transition name="slide-in">
      <div v-if="notifications.length > 0" class="notification-center">
        <div class="notification-header">
          <span>通知中心</span>
          <el-button size="small" text @click="clearAll">清空全部</el-button>
        </div>
        <div class="notification-list">
          <TransitionGroup name="notification-item">
            <div
              v-for="notification in notifications"
              :key="notification.id"
              class="notification-item"
              :class="`notification-${notification.type}`"
            >
              <div class="notification-icon">
                <el-icon>
                  <component :is="getIcon(notification.type)" />
                </el-icon>
              </div>
              <div class="notification-content">
                <div class="notification-title">{{ notification.title }}</div>
                <div v-if="notification.message" class="notification-message">{{ notification.message }}</div>
                <div class="notification-time">{{ formatTime(notification.timestamp) }}</div>
              </div>
              <el-button
                size="small"
                text
                class="notification-close"
                @click="removeNotification(notification.id)"
              >
                <el-icon><Close /></el-icon>
              </el-button>
            </div>
          </TransitionGroup>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElIcon } from 'element-plus'
import { CircleCheckFilled, CircleCloseFilled, WarningFilled, InfoFilled, Close } from '@element-plus/icons-vue'
import { useAppStore } from '@/stores/useAppStore'

const appStore = useAppStore()

const notifications = computed(() => appStore.notifications)

function getIcon(type: string) {
  switch (type) {
    case 'success': return CircleCheckFilled
    case 'error': return CircleCloseFilled
    case 'warning': return WarningFilled
    default: return InfoFilled
  }
}

function formatTime(timestamp: number) {
  const now = Date.now()
  const diff = now - timestamp

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  return `${Math.floor(diff / 86400000)}天前`
}

function removeNotification(id: string) {
  appStore.removeNotification(id)
}

function clearAll() {
  appStore.clearNotifications()
}
</script>

<style scoped>
.notification-center {
  position: fixed;
  top: 70px;
  right: 20px;
  width: 360px;
  max-width: calc(100vw - 40px);
  max-height: calc(100vh - 100px);
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 9998;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.notification-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #ebeef5;
  font-weight: 600;
  font-size: 14px;
}

.notification-list {
  overflow-y: auto;
  padding: 8px;
}

.notification-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  border-radius: var(--border-radius-sm);
  background: var(--bg-color);
  margin-bottom: var(--spacing-sm);
  position: relative;
}

.notification-item.notification-success {
  background: #ECFDF5;
  border-left: 3px solid var(--success-color);
}

.notification-item.notification-error {
  background: #FEF2F2;
  border-left: 3px solid var(--danger-color);
}

.notification-item.notification-warning {
  background: #FFFBEB;
  border-left: 3px solid var(--warning-color);
}

.notification-item.notification-info {
  background: var(--bg-color);
  border-left: 3px solid var(--text-color-secondary);
}

.notification-icon {
  flex-shrink: 0;
  font-size: var(--font-size-lg);
}

.notification-content {
  flex: 1;
  min-width: 0;
}

.notification-title {
  font-weight: 500;
  font-size: var(--font-size-md);
  margin-bottom: var(--spacing-xs);
  color: var(--text-color-primary);
}

.notification-message {
  font-size: var(--font-size-xs);
  color: var(--text-color-regular);
  margin-bottom: var(--spacing-xs);
  line-height: 1.4;
}

.notification-time {
  font-size: 11px;
  color: var(--text-color-secondary);
}

.notification-close {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 4px;
}

.slide-in-enter-active,
.slide-in-leave-active {
  transition: all 0.3s ease;
}

.slide-in-enter-from,
.slide-in-leave-to {
  opacity: 0;
  transform: translateX(30px);
}

.notification-item-enter-active,
.notification-item-leave-active {
  transition: all 0.3s ease;
}

.notification-item-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.notification-item-leave-to {
  opacity: 0;
  transform: translateX(20px);
}

.notification-item-leave-active {
  position: absolute;
  width: calc(100% - 16px);
}
</style>
