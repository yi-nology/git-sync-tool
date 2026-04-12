<template>
  <el-container class="app-layout">
    <el-header class="app-header">
      <router-link to="/" class="logo">
        <el-icon class="logo-icon"><Connection /></el-icon>
        <span class="logo-text">Git Manage Service</span>
      </router-link>
      <nav class="header-nav">
        <router-link to="/" class="nav-item" :class="{ active: isActive('/') }">
          <el-icon><HomeFilled /></el-icon><span>首页</span>
        </router-link>
        <router-link to="/repos" class="nav-item" :class="{ active: isActive('/repos') }">
          <el-icon><FolderOpened /></el-icon><span>仓库</span>
        </router-link>
        <router-link to="/audit" class="nav-item" :class="{ active: isActive('/audit') }">
          <el-icon><Warning /></el-icon><span>审计日志</span>
        </router-link>
        <router-link to="/settings" class="nav-item" :class="{ active: isActive('/settings') }">
          <el-icon><Setting /></el-icon><span>设置</span>
        </router-link>
        <router-link to="/mcp" class="nav-item" :class="{ active: isActive('/mcp') }">
          <el-icon><Connection /></el-icon><span>MCP</span>
        </router-link>
      </nav>
      <div class="header-right">
        <ThemeSwitch />
      </div>
    </el-header>
    <el-main class="app-main">
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { Connection, HomeFilled, FolderOpened, Setting, Warning } from '@element-plus/icons-vue'
import { useUIStore } from '@/stores/useUIStore'
import { useKeyboard } from '@/composables/useKeyboard'
import ThemeSwitch from '@/components/common/ThemeSwitch.vue'

const route = useRoute()
const uiStore = useUIStore()

useKeyboard()

function isActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

onMounted(() => {
  uiStore.restoreState()
  uiStore.applyDarkMode()
})
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  background: var(--bg-color);
}

.app-header {
  background: var(--bg-color-page);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  padding: 0 32px;
  height: var(--header-height);
  position: sticky;
  top: 0;
  z-index: 10;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  color: var(--text-color-primary);
  font-weight: 600;
  font-size: 16px;
  font-family: 'Inter', -apple-system, sans-serif;
  margin-right: 32px;
  flex-shrink: 0;
}

.logo-icon {
  color: var(--primary-color);
  font-size: 22px;
}

.header-nav {
  display: flex;
  gap: 4px;
  flex: 1;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: var(--border-radius-md);
  font-size: 13px;
  color: var(--text-color-secondary);
  text-decoration: none;
  transition: all var(--transition-fast);
  font-family: 'Inter', -apple-system, sans-serif;
}

.nav-item:hover {
  color: var(--text-color-primary);
  background: var(--border-color-light);
}

.nav-item.active {
  color: #fff;
  background: var(--primary-color);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
}

.app-main {
  padding: 24px 32px;
  max-width: var(--main-content-max-width);
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
}

@media (max-width: 768px) {
  .app-header {
    padding: 0 16px;
  }

  .logo-text {
    display: none;
  }

  .nav-item span {
    display: none;
  }

  .nav-item {
    padding: 8px 12px;
  }
}
</style>
