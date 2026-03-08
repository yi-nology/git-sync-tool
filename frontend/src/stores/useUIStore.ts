import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

/**
 * UI 全局状态管理
 */
export const useUIStore = defineStore('ui', () => {
  // 深色模式
  const isDarkMode = ref(false)

  // 全局加载状态
  const globalLoading = ref(false)
  const globalLoadingText = ref('加载中...')

  // 侧边栏折叠状态
  const isSidebarCollapsed = ref(false)

  // 从 localStorage 恢复状态
  const restoreState = () => {
    const savedDarkMode = localStorage.getItem('ui:darkMode')
    if (savedDarkMode !== null) {
      isDarkMode.value = savedDarkMode === 'true'
    }

    const savedSidebarCollapsed = localStorage.getItem('ui:sidebarCollapsed')
    if (savedSidebarCollapsed !== null) {
      isSidebarCollapsed.value = savedSidebarCollapsed === 'true'
    }
  }

  // 切换深色模式
  const toggleDarkMode = () => {
    isDarkMode.value = !isDarkMode.value
    applyDarkMode()
  }

  // 应用深色模式
  const applyDarkMode = () => {
    const html = document.documentElement
    if (isDarkMode.value) {
      html.classList.add('dark')
    } else {
      html.classList.remove('dark')
    }
    localStorage.setItem('ui:darkMode', String(isDarkMode.value))
  }

  // 设置全局加载
  const setGlobalLoading = (loading: boolean, text = '加载中...') => {
    globalLoading.value = loading
    globalLoadingText.value = text
  }

  // 切换侧边栏
  const toggleSidebar = () => {
    isSidebarCollapsed.value = !isSidebarCollapsed.value
    localStorage.setItem('ui:sidebarCollapsed', String(isSidebarCollapsed.value))
  }

  // 监听深色模式变化
  watch(isDarkMode, () => {
    applyDarkMode()
  })

  return {
    // 状态
    isDarkMode,
    globalLoading,
    globalLoadingText,
    isSidebarCollapsed,

    // 方法
    toggleDarkMode,
    setGlobalLoading,
    toggleSidebar,
    restoreState,
    applyDarkMode,
  }
})
