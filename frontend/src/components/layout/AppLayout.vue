<template>
  <el-container class="app-layout">
    <el-header class="app-header">
      <div class="header-left">
        <router-link to="/" class="logo">
          <img src="/logo.svg" alt="Logo" class="logo-icon" />
          <span class="logo-text">Git Manage Service</span>
        </router-link>
      </div>
      <el-menu
        :default-active="activeMenu"
        mode="horizontal"
        router
        class="header-menu"
      >
        <el-menu-item index="/">首页</el-menu-item>
        <el-menu-item index="/repos">仓库管理</el-menu-item>
        <el-menu-item index="/audit">审计日志</el-menu-item>
        <el-menu-item index="/settings">设置</el-menu-item>
      </el-menu>
      <div class="header-right">
        <ThemeSwitch />
        <el-tooltip content="快捷键: Cmd/Ctrl + K (搜索)" placement="bottom">
          <el-button :icon="Search" circle />
        </el-tooltip>
      </div>
    </el-header>
    <el-main class="app-main">
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { Connection, Search } from '@element-plus/icons-vue'
import { useUIStore } from '@/stores/useUIStore'
import { useKeyboard } from '@/composables/useKeyboard'
import ThemeSwitch from '@/components/common/ThemeSwitch.vue'

const route = useRoute()
const uiStore = useUIStore()

// 初始化快捷键
useKeyboard()

const activeMenu = computed(() => {
  const path = route.path
  if (path === '/') return '/'
  if (path.startsWith('/repos')) return '/repos'
  if (path.startsWith('/audit')) return '/audit'
  if (path.startsWith('/settings')) return '/settings'
  return path
})

onMounted(() => {
  // 恢复 UI 状态（深色模式等）
  uiStore.restoreState()
  uiStore.applyDarkMode()
})
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  background: var(--el-bg-color-page);
}

.app-header {
  background: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color-lighter);
  display: flex;
  align-items: center;
  padding: 0 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
  margin-right: 40px;
}

.logo {
  display: flex;
  align-items: center;
  text-decoration: none;
  color: var(--el-text-color-primary);
  font-weight: 600;
  font-size: 16px;
  gap: 8px;
  transition: color 0.3s;
}

.logo:hover {
  color: var(--el-color-primary);
}

.logo-text {
  white-space: nowrap;
}

.logo-icon {
  width: 28px;
  height: 28px;
}

.header-menu {
  border-bottom: none !important;
  flex: 1;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
}

.app-main {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
}

/* 深色模式样式 */
:global(.dark) .app-layout {
  background: #1a1a1a;
}

/* 响应式 */
@media (max-width: 768px) {
  .app-header {
    padding: 0 16px;
  }

  .header-left {
    margin-right: 16px;
  }

  .logo-text {
    display: none;
  }

  .header-menu {
    font-size: 14px;
  }

  .header-right {
    gap: 8px;
  }
}
</style>
