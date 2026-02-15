<template>
  <div class="home-page">
    <div class="hero-section">
      <h1 class="hero-title">Git Branch Manager</h1>
      <p class="hero-desc">
        一个轻量级、可视化的 Git 仓库与分支同步管理工具。<br />
        帮助团队更高效地管理多仓库、多分支的代码同步与集成。
      </p>
      <div class="hero-actions">
        <el-button type="primary" size="large" @click="$router.push('/repos')">
          开始使用
        </el-button>
        <el-button size="large" @click="$router.push('/repos')">
          仓库管理
        </el-button>
      </div>
    </div>

    <el-row :gutter="20" class="feature-cards">
      <el-col :xs="24" :sm="8">
        <el-card shadow="hover" class="feature-card">
          <div class="feature-icon">
            <el-icon :size="40" color="#409EFF"><Connection /></el-icon>
          </div>
          <h3>仓库管理</h3>
          <p>统一管理多个 Git 仓库，支持本地扫描与远程克隆。直观查看分支状态与 Remote 配置。</p>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="8">
        <el-card shadow="hover" class="feature-card">
          <div class="feature-icon">
            <el-icon :size="40" color="#67C23A"><Refresh /></el-icon>
          </div>
          <h3>自动同步</h3>
          <p>配置灵活的同步规则，支持定时任务 (Cron) 与 Webhook 触发。实现分支间的自动合并与推送。</p>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="8">
        <el-card shadow="hover" class="feature-card">
          <div class="feature-icon">
            <el-icon :size="40" color="#E6A23C"><DataLine /></el-icon>
          </div>
          <h3>贡献分析</h3>
          <p>可视化度量代码贡献度，按时间、作者多维度分析代码提交情况与行数变化。</p>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { Connection, Refresh, DataLine } from '@element-plus/icons-vue'
</script>

<style scoped>
.home-page {
  max-width: 1000px;
  margin: 0 auto;
}
.hero-section {
  text-align: center;
  padding: 60px 0 40px;
}
.hero-title {
  font-size: 36px;
  font-weight: 700;
  margin-bottom: 16px;
  color: #303133;
}
.hero-desc {
  font-size: 16px;
  color: #909399;
  line-height: 1.8;
  margin-bottom: 32px;
}
.hero-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
}
.feature-cards {
  padding: 20px 0 40px;
}
.feature-card {
  text-align: center;
  padding: 20px 10px;
  height: 100%;
}
.feature-card h3 {
  margin: 16px 0 8px;
  font-size: 18px;
}
.feature-card p {
  color: #909399;
  font-size: 14px;
  line-height: 1.6;
}
.feature-icon {
  margin-bottom: 8px;
}
</style>
