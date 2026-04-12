<template>
  <div class="home-page">
    <section class="hero">
      <div class="hero-icon">
        <el-icon :size="32"><Connection /></el-icon>
      </div>
      <h1 class="hero-title">Git Manage Service</h1>
      <p class="hero-desc">
        一个轻量级的多仓库、多分支自动化同步管理系统。<br />
        提供友好的 Web 界面，支持定时任务、Webhook 触发、多渠道消息通知以及详细的同步日志记录。
      </p>
      <div class="hero-actions">
        <el-button type="primary" size="large" @click="$router.push('/repos')">
          开始使用<el-icon class="el-icon--right"><ArrowRight /></el-icon>
        </el-button>
        <el-button size="large" @click="openGitHub">
          <el-icon><Link /></el-icon>GitHub
        </el-button>
      </div>
    </section>

    <div class="section-divider" />

    <section class="features-section">
      <h2 class="section-title">功能特性</h2>
      <div class="features-grid">
        <div v-for="feat in features" :key="feat.title" class="feature-card">
          <div class="feature-icon" :style="{ background: feat.bg }">
            <el-icon :size="24" :color="feat.color"><component :is="feat.icon" /></el-icon>
          </div>
          <h3>{{ feat.title }}</h3>
          <p>{{ feat.desc }}</p>
        </div>
      </div>
    </section>

    <div class="section-divider" />

    <section class="tech-section">
      <h2 class="section-title">技术栈</h2>
      <div class="tech-grid">
        <div class="tech-card">
          <h3>后端</h3>
          <div class="tech-list">
            <div v-for="item in backendTech" :key="item.label" class="tech-row">
              <span class="tech-label">{{ item.label }}</span>
              <span class="tech-value">{{ item.value }}</span>
            </div>
          </div>
        </div>
        <div class="tech-card">
          <h3>前端</h3>
          <div class="tech-list">
            <div v-for="item in frontendTech" :key="item.label" class="tech-row">
              <span class="tech-label">{{ item.label }}</span>
              <span class="tech-value">{{ item.value }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <div class="section-divider" />

    <section class="version-section">
      <h2 class="section-title">版本信息</h2>
      <div class="version-card" v-if="appInfo">
        <div class="version-item">
          <span class="version-label">应用名称</span>
          <span class="version-value">{{ appInfo.app_name }}</span>
        </div>
        <div class="version-item">
          <span class="version-label">版本号</span>
          <span class="version-value highlight">{{ appInfo.version }}</span>
        </div>
        <div class="version-item">
          <span class="version-label">构建时间</span>
          <span class="version-value">{{ appInfo.build_time }}</span>
        </div>
        <div class="version-item">
          <span class="version-label">Git Commit</span>
          <span class="version-value mono">{{ appInfo.git_commit }}</span>
        </div>
      </div>
      <el-skeleton v-else :rows="2" animated />
    </section>

    <footer class="home-footer">
      Licensed under Apache 2.0
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, type Component } from 'vue'
import {
  Connection, Refresh, Bell, Key, Coin, Setting, Link, Edit, Stamp,
  ArrowRight,
} from '@element-plus/icons-vue'
import { getAppInfo } from '@/api/modules/system'
import type { AppInfo } from '@/api/modules/system'

const appInfo = ref<AppInfo | null>(null)

function openGitHub() {
  window.open('https://github.com/yi-nology/git-manage-service', '_blank')
}

interface Feature {
  title: string
  desc: string
  icon: Component
  color: string
  bg: string
}

const features: Feature[] = [
  { title: '多仓库管理', desc: '统一注册和管理本地 Git 仓库，支持本地扫描与远程克隆。', icon: Connection, color: '#6366F1', bg: '#EEF2FF' },
  { title: '自动同步', desc: '支持单分支和全分支同步模式，Cron 定时调度与 Webhook 触发。', icon: Refresh, color: '#10B981', bg: '#ECFDF5' },
  { title: '多渠道通知', desc: '支持钉钉、企业微信、飞书、蓝信、邮件、自定义 Webhook。', icon: Bell, color: '#F59E0B', bg: '#FFFBEB' },
  { title: 'SSH 密钥管理', desc: '支持数据库统一管理 SSH 密钥，灵活配置仓库认证方式。', icon: Key, color: '#EF4444', bg: '#FEF2F2' },
  { title: 'Spec 编辑器', desc: '集成 Monaco Editor 的 RPM Spec 文件编辑器，支持实时语法检查。', icon: Edit, color: '#06B6D4', bg: '#ECFEFF' },
  { title: 'Patch 管理', desc: '生成、管理和应用 Git Patch 文件，支持批量操作和自动化工作流。', icon: Stamp, color: '#8B5CF6', bg: '#F5F3FF' },
  { title: '多数据库支持', desc: '支持 SQLite、MySQL、PostgreSQL，可选 MinIO 对象存储。', icon: Coin, color: '#64748B', bg: '#F1F5F9' },
  { title: '安全可靠', desc: '冲突检测、Fast-Forward 检查及 Force Push 保护，审计日志全程记录。', icon: Setting, color: '#8B5CF6', bg: '#F5F3FF' },
  { title: '支持 MCP', desc: '集成 MCP 多模型协作平台，支持多种 AI 模型对接与协作。', icon: Connection, color: '#10B981', bg: '#ECFDF5' },
]

const backendTech = [
  { label: '语言', value: 'Go 1.24' },
  { label: 'HTTP 框架', value: 'Hertz (CloudWeGo)' },
  { label: 'RPC 框架', value: 'Kitex (CloudWeGo)' },
  { label: 'ORM', value: 'GORM' },
  { label: 'Git 引擎', value: 'go-git' },
  { label: '数据库', value: 'SQLite / MySQL / PostgreSQL' },
]

const frontendTech = [
  { label: '框架', value: 'Vue 3 (Composition API)' },
  { label: '语言', value: 'TypeScript' },
  { label: 'UI 组件库', value: 'Element Plus' },
  { label: '构建工具', value: 'Vite' },
  { label: '状态管理', value: 'Pinia' },
  { label: '图表库', value: 'ECharts' },
]

onMounted(async () => {
  try {
    appInfo.value = await getAppInfo()
  } catch { /* ignore */ }
})
</script>

<style scoped>
.home-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 0 20px;
}

.hero {
  text-align: center;
  padding: 48px 0 32px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.hero-icon {
  width: 64px;
  height: 64px;
  border-radius: var(--border-radius-xl);
  background: var(--accent-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary-color);
}

.hero-title {
  font-size: 36px;
  font-weight: 700;
  color: var(--text-color-primary);
  font-family: 'Inter', -apple-system, sans-serif;
}

.hero-desc {
  font-size: 15px;
  color: var(--text-color-secondary);
  line-height: 1.8;
  max-width: 560px;
}

.hero-actions {
  display: flex;
  gap: 12px;
  margin-top: 4px;
}

.section-divider {
  height: 1px;
  background: var(--border-color);
  margin: 8px 0 32px;
}

.section-title {
  font-size: 22px;
  font-weight: 600;
  color: var(--text-color-primary);
  margin-bottom: 20px;
  font-family: 'Inter', -apple-system, sans-serif;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.feature-card {
  text-align: center;
  padding: 24px 16px;
  border-radius: var(--border-radius-lg);
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  transition: all var(--transition-normal);
}

.feature-card:hover {
  box-shadow: var(--box-shadow-md);
  transform: translateY(-2px);
}

.feature-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--border-radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 12px;
}

.feature-card h3 {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color-primary);
  margin-bottom: 6px;
}

.feature-card p {
  font-size: 13px;
  color: var(--text-color-secondary);
  line-height: 1.6;
}

.tech-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.tech-card {
  padding: 24px;
  border-radius: var(--border-radius-lg);
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
}

.tech-card h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color-primary);
  margin-bottom: 16px;
}

.tech-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.tech-row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
}

.tech-label {
  color: var(--text-color-secondary);
}

.tech-value {
  color: var(--text-color-primary);
  font-weight: 500;
}

.version-card {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  padding: 24px;
  border-radius: var(--border-radius-lg);
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
}

.version-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.version-label {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.version-value {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.version-value.highlight {
  color: var(--primary-color);
}

.version-value.mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}

.home-footer {
  text-align: center;
  padding: 24px 0 32px;
  font-size: 12px;
  color: var(--text-color-secondary);
}

@media (max-width: 768px) {
  .features-grid {
    grid-template-columns: 1fr;
  }
  .tech-grid {
    grid-template-columns: 1fr;
  }
  .version-card {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
