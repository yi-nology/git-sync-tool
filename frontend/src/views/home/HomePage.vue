<template>
  <div class="home-page">
    <div class="home-header">
      <h1 class="home-title">Git Manage Service</h1>
      <p class="home-desc">
        一个轻量级的多仓库、多分支自动化同步管理系统。<br />
        提供友好的 Web 界面，支持定时任务、Webhook 触发、多渠道消息通知以及详细的同步日志记录。
      </p>
      <div class="home-actions">
        <el-button type="primary" size="large" @click="$router.push('/repos')">
          开始使用
        </el-button>
        <el-button size="large" tag="a" href="https://github.com/yi-nology/git-manage-service" target="_blank">
          <el-icon><Link /></el-icon>&nbsp;GitHub
        </el-button>
      </div>
    </div>

    <el-divider />

    <h2 class="section-title">功能特性</h2>
    <el-row :gutter="16" class="feature-cards">
      <el-col :xs="24" :sm="12" :md="8" v-for="feat in features" :key="feat.title">
        <el-card shadow="hover" class="feature-card">
          <div class="feature-icon">
            <el-icon :size="32" :color="feat.color"><component :is="feat.icon" /></el-icon>
          </div>
          <h3>{{ feat.title }}</h3>
          <p>{{ feat.desc }}</p>
        </el-card>
      </el-col>
    </el-row>

    <el-divider />

    <h2 class="section-title">技术栈</h2>
    <el-row :gutter="20" class="tech-section">
      <el-col :xs="24" :sm="12">
        <el-card shadow="hover">
          <template #header><strong>后端</strong></template>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="语言">Go 1.24</el-descriptions-item>
            <el-descriptions-item label="HTTP 框架">Hertz (CloudWeGo)</el-descriptions-item>
            <el-descriptions-item label="RPC 框架">Kitex (CloudWeGo)</el-descriptions-item>
            <el-descriptions-item label="ORM">GORM</el-descriptions-item>
            <el-descriptions-item label="Git 引擎">go-git</el-descriptions-item>
            <el-descriptions-item label="数据库">SQLite / MySQL / PostgreSQL</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12">
        <el-card shadow="hover">
          <template #header><strong>前端</strong></template>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="框架">Vue 3 (Composition API)</el-descriptions-item>
            <el-descriptions-item label="语言">TypeScript</el-descriptions-item>
            <el-descriptions-item label="UI 组件库">Element Plus</el-descriptions-item>
            <el-descriptions-item label="构建工具">Vite</el-descriptions-item>
            <el-descriptions-item label="状态管理">Pinia</el-descriptions-item>
            <el-descriptions-item label="图表库">ECharts</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <el-divider />

    <h2 class="section-title">版本信息</h2>
    <el-card shadow="hover" class="version-card">
      <el-descriptions :column="2" border v-if="appInfo">
        <el-descriptions-item label="应用名称">{{ appInfo.app_name }}</el-descriptions-item>
        <el-descriptions-item label="版本号">{{ appInfo.version }}</el-descriptions-item>
        <el-descriptions-item label="构建时间">{{ appInfo.build_time }}</el-descriptions-item>
        <el-descriptions-item label="Git Commit">
          <el-text type="info" size="small" style="font-family: monospace;">{{ appInfo.git_commit }}</el-text>
        </el-descriptions-item>
      </el-descriptions>
      <el-skeleton :rows="2" animated v-else />
    </el-card>

    <div class="home-footer">
      <el-text type="info" size="small">
        Licensed under Apache 2.0
      </el-text>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  Connection,
  Refresh,
  Bell,
  Key,
  Coin,
  Setting,
  Link,
} from '@element-plus/icons-vue'
import { getAppInfo } from '@/api/modules/system'
import type { AppInfo } from '@/api/modules/system'

const appInfo = ref<AppInfo | null>(null)

const features = [
  {
    title: '多仓库管理',
    desc: '统一注册和管理本地 Git 仓库，支持本地扫描与远程克隆。',
    icon: Connection,
    color: '#409EFF',
  },
  {
    title: '自动同步',
    desc: '支持单分支和全分支同步模式，Cron 定时调度与 Webhook 触发。',
    icon: Refresh,
    color: '#67C23A',
  },
  {
    title: '多渠道通知',
    desc: '支持钉钉、企业微信、飞书、蓝信、邮件、自定义 Webhook，按事件类型发送。',
    icon: Bell,
    color: '#E6A23C',
  },
  {
    title: 'SSH 密钥管理',
    desc: '支持数据库统一管理 SSH 密钥，灵活配置仓库认证方式。',
    icon: Key,
    color: '#F56C6C',
  },
  {
    title: '多数据库支持',
    desc: '支持 SQLite、MySQL、PostgreSQL，可选 MinIO 对象存储和 Redis 分布式锁。',
    icon: Coin,
    color: '#909399',
  },
  {
    title: '安全可靠',
    desc: '冲突检测、Fast-Forward 检查及 Force Push 保护，审计日志全程记录。',
    icon: Setting,
    color: '#8B5CF6',
  },
]

onMounted(async () => {
  try {
    appInfo.value = await getAppInfo()
  } catch {
    // 版本信息加载失败不影响页面展示
  }
})
</script>

<style scoped>
.home-page {
  max-width: 1000px;
  margin: 0 auto;
}
.home-header {
  text-align: center;
  padding: 40px 0 20px;
}
.home-title {
  font-size: 32px;
  font-weight: 700;
  margin-bottom: 12px;
  color: #303133;
}
.home-desc {
  font-size: 15px;
  color: #909399;
  line-height: 1.8;
  margin-bottom: 20px;
}
.home-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
}
.section-title {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
}
.feature-cards {
  margin-bottom: 8px;
}
.feature-card {
  text-align: center;
  padding: 12px 8px;
  margin-bottom: 16px;
  height: calc(100% - 16px);
}
.feature-card h3 {
  margin: 12px 0 6px;
  font-size: 16px;
}
.feature-card p {
  color: #909399;
  font-size: 13px;
  line-height: 1.6;
}
.feature-icon {
  margin-bottom: 4px;
}
.tech-section .el-col {
  margin-bottom: 16px;
}
.version-card {
  margin-bottom: 24px;
}
.home-footer {
  text-align: center;
  padding: 16px 0 32px;
}
</style>
