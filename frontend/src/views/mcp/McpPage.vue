<template>
  <div class="mcp-page">
    <h1 class="page-title">MCP 配置</h1>
    
    <el-card shadow="hover" class="mcp-intro-card">
      <template #header>
        <div class="card-header">
          <strong>MCP 多模型协作平台</strong>
          <el-switch v-model="mcpEnabled" active-text="开启" inactive-text="关闭" />
        </div>
      </template>
      <p class="intro-text">
        MCP (Multi-Model Collaboration Platform) 是一个多模型协作平台，支持多种 AI 模型的对接与协作。
        通过 MCP，您可以统一管理和使用不同的 AI 模型，实现更高效的工作流。
      </p>
    </el-card>

    <el-divider />

    <h2 class="section-title">对接配置</h2>
    
    <el-card shadow="hover" class="config-card">
      <template #header>
        <strong>MCP 对接参数</strong>
      </template>
      <div class="config-content">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="服务地址">http://localhost:3002</el-descriptions-item>
          <el-descriptions-item label="API 密钥">在 MCP 服务配置中获取</el-descriptions-item>
          <el-descriptions-item label="对接端口">3002</el-descriptions-item>
          <el-descriptions-item label="协议">HTTP</el-descriptions-item>
          <el-descriptions-item label="状态" :span="1">
            <el-tag :type="mcpEnabled ? 'success' : 'info'">
              {{ mcpEnabled ? '已开启' : '已关闭' }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-card>

    <el-card shadow="hover" class="config-card">
      <template #header>
        <strong>配置设置</strong>
      </template>
      <div class="config-form">
        <el-form :model="mcpConfig" label-width="120px">
          <el-form-item label="服务地址">
            <el-input v-model="mcpConfig.serviceUrl" placeholder="请输入 MCP 服务地址" />
          </el-form-item>
          <el-form-item label="API 密钥">
            <el-input v-model="mcpConfig.apiKey" type="password" placeholder="请输入 API 密钥" />
          </el-form-item>
          <el-form-item label="对接端口">
            <el-input-number v-model="mcpConfig.port" :min="1" :max="65535" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveConfig">保存配置</el-button>
            <el-button @click="testConnection">测试连接</el-button>
          </el-form-item>
        </el-form>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'

const mcpEnabled = ref(false)

const mcpConfig = reactive({
  serviceUrl: 'http://localhost:3002',
  apiKey: '',
  port: 3002
})

const saveConfig = () => {
  // 保存配置逻辑
  console.log('保存配置:', mcpConfig)
}

const testConnection = () => {
  // 测试连接逻辑
  console.log('测试连接:', mcpConfig.serviceUrl)
}
</script>

<style scoped>
.mcp-page {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px 0;
}

.page-title {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 20px;
  color: #303133;
}

.section-title {
  font-size: 20px;
  font-weight: 600;
  margin: 30px 0 16px;
  color: #303133;
}

.mcp-intro-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.intro-text {
  line-height: 1.8;
  color: #606266;
}

.config-card {
  margin-bottom: 20px;
}

.config-content {
  margin-top: 10px;
}

.config-form {
  margin-top: 10px;
}

.el-descriptions {
  margin-bottom: 20px;
}
</style>