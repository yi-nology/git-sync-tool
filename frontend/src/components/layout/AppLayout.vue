<template>
  <el-container class="app-layout">
    <el-header class="app-header">
      <div class="header-left">
        <router-link to="/" class="logo">
          <el-icon :size="22"><Connection /></el-icon>
          <span class="logo-text">Git Branch Manager</span>
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
    </el-header>
    <el-main class="app-main">
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Connection } from '@element-plus/icons-vue'

const route = useRoute()

const activeMenu = computed(() => {
  const path = route.path
  if (path === '/') return '/'
  if (path.startsWith('/repos')) return '/repos'
  if (path.startsWith('/audit')) return '/audit'
  if (path.startsWith('/settings')) return '/settings'
  return path
})
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  background: #f5f7fa;
}
.app-header {
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
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
  color: #303133;
  font-weight: 600;
  font-size: 16px;
  gap: 8px;
}
.logo-text {
  white-space: nowrap;
}
.header-menu {
  border-bottom: none !important;
  flex: 1;
}
.app-main {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
}
</style>
