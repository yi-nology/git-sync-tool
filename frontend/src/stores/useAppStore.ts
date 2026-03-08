import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface Notification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message?: string
  timestamp: number
}

export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref(false)
  const globalLoading = ref(false)
  const loadingText = ref('')
  const notifications = ref<Notification[]>([])
  const notificationId = ref(0)

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setGlobalLoading(loading: boolean, text = '') {
    globalLoading.value = loading
    loadingText.value = text
  }

  function addNotification(notification: Omit<Notification, 'id' | 'timestamp'>) {
    const id = `notification-${notificationId.value++}`
    notifications.value.unshift({
      ...notification,
      id,
      timestamp: Date.now(),
    })

    // 自动删除通知（5秒后）
    setTimeout(() => {
      removeNotification(id)
    }, 5000)

    // 限制通知数量
    if (notifications.value.length > 10) {
      notifications.value = notifications.value.slice(0, 10)
    }
  }

  function removeNotification(id: string) {
    const index = notifications.value.findIndex(n => n.id === id)
    if (index !== -1) {
      notifications.value.splice(index, 1)
    }
  }

  function clearNotifications() {
    notifications.value = []
  }

  const hasNotifications = computed(() => notifications.value.length > 0)

  return {
    sidebarCollapsed,
    toggleSidebar,
    globalLoading,
    loadingText,
    setGlobalLoading,
    notifications,
    hasNotifications,
    addNotification,
    removeNotification,
    clearNotifications,
  }
})
