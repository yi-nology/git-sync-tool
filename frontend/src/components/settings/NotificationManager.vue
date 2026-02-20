<template>
  <div class="notification-manager">
    <!-- 工具栏 -->
    <div class="toolbar">
      <el-button type="primary" @click="openAddDialog">
        <el-icon><Plus /></el-icon> 添加渠道
      </el-button>
      <el-button @click="loadChannels" :loading="loading">
        <el-icon><Refresh /></el-icon> 刷新
      </el-button>
    </div>

    <!-- 渠道列表 -->
    <el-table :data="channels" v-loading="loading" empty-text="暂无通知渠道">
      <el-table-column prop="name" label="名称" min-width="150" />
      <el-table-column prop="type" label="类型" width="120">
        <template #default="{ row }">
          <el-tag size="small">{{ typeLabels[row.type] || row.type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
            {{ row.enabled ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="触发时机" min-width="280">
        <template #default="{ row }">
          <template v-if="parseTriggerEvents(row.trigger_events).length > 0">
            <el-tag v-for="event in parseTriggerEvents(row.trigger_events)" :key="event" :type="triggerEventTagType(event)" size="small" class="mr-1 mb-1">
              {{ triggerEventLabels[event] || event }}
            </el-tag>
          </template>
          <template v-else>
            <el-tag v-if="row.notify_on_success" type="success" size="small" class="mr-1">成功</el-tag>
            <el-tag v-if="row.notify_on_failure" type="danger" size="small">失败</el-tag>
          </template>
        </template>
      </el-table-column>
      <el-table-column prop="updated_at" label="更新时间" width="160" />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button-group size="small">
            <el-button @click="handleTest(row.id)" :loading="testingId === row.id" title="测试">
              <el-icon><Promotion /></el-icon>
            </el-button>
            <el-button @click="openEditDialog(row)" title="编辑">
              <el-icon><Edit /></el-icon>
            </el-button>
            <el-popconfirm title="确定删除此渠道?" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button type="danger" title="删除">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </template>
            </el-popconfirm>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>

    <!-- 添加/编辑渠道对话框 -->
    <el-dialog v-model="showDialog" :title="editingId ? '编辑通知渠道' : '添加通知渠道'" width="640px" destroy-on-close>
      <el-form :model="form" label-width="100px">
        <el-form-item label="渠道名称" required>
          <el-input v-model="form.name" placeholder="例如：开发团队钉钉群" />
        </el-form-item>
        <el-form-item label="渠道类型" required>
          <el-select v-model="form.type" style="width: 100%" :disabled="!!editingId">
            <el-option label="邮件 (Email)" value="email" />
            <el-option label="钉钉机器人" value="dingtalk" />
            <el-option label="企业微信机器人" value="wechat" />
            <el-option label="蓝信机器人" value="lanxin" />
            <el-option label="飞书机器人" value="feishu" />
            <el-option label="自定义 Webhook" value="webhook" />
          </el-select>
        </el-form-item>

        <!-- Email 配置 -->
        <template v-if="form.type === 'email'">
          <el-divider content-position="left">邮件配置</el-divider>
          <el-form-item label="SMTP 服务器">
            <el-row :gutter="12">
              <el-col :span="16">
                <el-input v-model="configForm.smtp_host" placeholder="smtp.example.com" />
              </el-col>
              <el-col :span="8">
                <el-input v-model="configForm.smtp_port" placeholder="端口 587" />
              </el-col>
            </el-row>
          </el-form-item>
          <el-form-item label="用户名">
            <el-input v-model="configForm.username" placeholder="发件邮箱账号" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="configForm.password" type="password" show-password placeholder="邮箱密码或授权码" />
          </el-form-item>
          <el-form-item label="发件人">
            <el-input v-model="configForm.from" placeholder="Git管理服务 <noreply@example.com>" />
          </el-form-item>
          <el-form-item label="收件人">
            <el-input v-model="configForm.to" placeholder="多个邮箱用逗号分隔" />
          </el-form-item>
        </template>

        <!-- 钉钉配置 -->
        <template v-if="form.type === 'dingtalk'">
          <el-divider content-position="left">钉钉配置</el-divider>
          <el-form-item label="Webhook URL">
            <el-input v-model="configForm.webhook_url" placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx" />
          </el-form-item>
          <el-form-item label="安全模式">
            <el-radio-group v-model="configForm.security_type">
              <el-radio value="none">无</el-radio>
              <el-radio value="sign">签名</el-radio>
              <el-radio value="keyword">关键字</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item v-if="configForm.security_type === 'sign'" label="签名密钥">
            <el-input v-model="configForm.secret" placeholder="SEC开头的密钥" />
          </el-form-item>
          <el-form-item v-if="configForm.security_type === 'keyword'" label="关键字">
            <el-input v-model="configForm.keywords" placeholder="消息中需要包含的关键字" />
          </el-form-item>
        </template>

        <!-- 企业微信配置 -->
        <template v-if="form.type === 'wechat'">
          <el-divider content-position="left">企业微信配置</el-divider>
          <el-form-item label="Webhook URL">
            <el-input v-model="configForm.webhook_url" placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx" />
          </el-form-item>
        </template>

        <!-- 蓝信配置 -->
        <template v-if="form.type === 'lanxin'">
          <el-divider content-position="left">蓝信配置</el-divider>
          <el-form-item label="Webhook URL">
            <el-input v-model="configForm.webhook_url" placeholder="蓝信机器人 Webhook 地址" />
          </el-form-item>
          <el-form-item label="安全模式">
            <el-radio-group v-model="configForm.security_type">
              <el-radio value="none">无</el-radio>
              <el-radio value="sign">签名</el-radio>
              <el-radio value="keyword">关键字</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item v-if="configForm.security_type === 'sign'" label="签名密钥">
            <el-input v-model="configForm.sign" placeholder="签名密钥" />
          </el-form-item>
          <el-form-item v-if="configForm.security_type === 'keyword'" label="关键字">
            <el-input v-model="configForm.keywords" placeholder="消息中需要包含的关键字" />
          </el-form-item>
        </template>

        <!-- 飞书配置 -->
        <template v-if="form.type === 'feishu'">
          <el-divider content-position="left">飞书配置</el-divider>
          <el-form-item label="Webhook URL">
            <el-input v-model="configForm.webhook_url" placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/xxx" />
          </el-form-item>
          <el-form-item label="安全模式">
            <el-radio-group v-model="configForm.security_type">
              <el-radio value="none">无</el-radio>
              <el-radio value="sign">签名</el-radio>
              <el-radio value="keyword">关键字</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item v-if="configForm.security_type === 'sign'" label="签名密钥">
            <el-input v-model="configForm.secret" placeholder="飞书签名密钥" />
          </el-form-item>
          <el-form-item v-if="configForm.security_type === 'keyword'" label="关键字">
            <el-input v-model="configForm.keywords" placeholder="消息中需要包含的关键字" />
          </el-form-item>
        </template>

        <!-- 自定义 Webhook 配置 -->
        <template v-if="form.type === 'webhook'">
          <el-divider content-position="left">Webhook 配置</el-divider>
          <el-form-item label="URL">
            <el-input v-model="configForm.url" placeholder="https://your-server.com/webhook" />
          </el-form-item>
          <el-form-item label="请求方法">
            <el-select v-model="configForm.method" style="width: 100%">
              <el-option label="POST" value="POST" />
              <el-option label="GET" value="GET" />
            </el-select>
          </el-form-item>
          <el-form-item label="Content-Type">
            <el-select v-model="configForm.content_type" style="width: 100%">
              <el-option label="application/json" value="application/json" />
              <el-option label="application/x-www-form-urlencoded" value="application/x-www-form-urlencoded" />
            </el-select>
          </el-form-item>
        </template>

        <el-divider content-position="left">通知选项</el-divider>
        <el-form-item label="启用渠道">
          <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item label="触发时机">
          <div class="trigger-events-grid">
            <div class="trigger-group">
              <div class="trigger-group-title">同步事件</div>
              <el-checkbox v-model="triggerEventsMap.sync_success">同步成功</el-checkbox>
              <el-checkbox v-model="triggerEventsMap.sync_failure">同步失败</el-checkbox>
              <el-checkbox v-model="triggerEventsMap.sync_conflict">同步冲突</el-checkbox>
            </div>
            <div class="trigger-group">
              <div class="trigger-group-title">Webhook 事件</div>
              <el-checkbox v-model="triggerEventsMap.webhook_received">Webhook 接收</el-checkbox>
              <el-checkbox v-model="triggerEventsMap.webhook_error">Webhook 处理错误</el-checkbox>
            </div>
            <div class="trigger-group">
              <div class="trigger-group-title">定时任务</div>
              <el-checkbox v-model="triggerEventsMap.cron_triggered">定时任务触发</el-checkbox>
            </div>
            <div class="trigger-group">
              <div class="trigger-group-title">备份事件</div>
              <el-checkbox v-model="triggerEventsMap.backup_success">备份成功</el-checkbox>
              <el-checkbox v-model="triggerEventsMap.backup_failure">备份失败</el-checkbox>
            </div>
          </div>
        </el-form-item>

        <!-- ========= 按触发时机独立配置消息模板 ========= -->
        <el-divider content-position="left">消息模板（按时机独立配置）</el-divider>

        <template v-if="enabledEvents.length > 0">
          <!-- 事件 Tab 切换 -->
          <el-form-item>
            <el-tabs v-model="activeEventTab" type="card" class="event-template-tabs">
              <el-tab-pane
                v-for="event in enabledEvents"
                :key="event"
                :name="event"
                :label="triggerEventLabels[event]"
              />
            </el-tabs>
          </el-form-item>

          <!-- 变量选择面板 -->
          <el-form-item>
            <div class="variable-panel">
              <div v-for="category in variableCategories" :key="category.label" class="var-group">
                <div class="var-group-header">
                  <el-tag :type="category.type" size="small" effect="dark">{{ category.label }}</el-tag>
                </div>
                <div class="var-buttons">
                  <el-tooltip
                    v-for="varName in category.vars"
                    :key="varName"
                    :content="`${formatVar(varName)}\n示例: ${getVarByName(varName)?.example || ''}`"
                    placement="top"
                    :show-after="300"
                    raw-content
                  >
                    <el-button
                      size="small"
                      :type="category.type || 'primary'"
                      plain
                      class="var-btn"
                      @click="insertVariable(varName)"
                    >{{ getVarByName(varName)?.description || varName }}</el-button>
                  </el-tooltip>
                </div>
              </div>
              <div class="active-editor-hint">
                <el-icon :size="14"><InfoFilled /></el-icon>
                <span>点击变量将插入到
                  <el-tag :type="activeEditor === 'title' ? 'primary' : 'success'" size="small" effect="plain">
                    {{ activeEditor === 'title' ? '标题模板' : '内容模板' }}
                  </el-tag>
                  <el-tag type="info" size="small" effect="plain" style="margin-left:4px">{{ triggerEventLabels[activeEventTab] }}</el-tag>
                </span>
              </div>
            </div>
          </el-form-item>

          <!-- 标题模板编辑器 -->
          <el-form-item label="标题模板">
            <el-input
              ref="titleInputRef"
              v-model="currentTitleTemplate"
              :placeholder="`留空使用「${triggerEventLabels[activeEventTab]}」的内置默认模板`"
              clearable
              class="template-input"
              :class="{ 'editor-active': activeEditor === 'title' }"
              @focus="handleEditorFocus('title')"
            />
          </el-form-item>

          <!-- 内容模板编辑器 -->
          <el-form-item label="内容模板">
            <el-input
              ref="contentInputRef"
              v-model="currentContentTemplate"
              type="textarea"
              :rows="6"
              :placeholder="`留空使用「${triggerEventLabels[activeEventTab]}」的内置默认模板`"
              clearable
              class="template-input"
              :class="{ 'editor-active': activeEditor === 'content' }"
              @focus="handleEditorFocus('content')"
            />
          </el-form-item>

          <!-- 实时预览 -->
          <el-form-item>
            <el-collapse v-model="previewCollapse" class="preview-collapse">
              <el-collapse-item title="实时预览" name="preview">
                <div class="template-preview">
                  <div class="preview-section">
                    <div class="preview-label">标题</div>
                    <div class="preview-title">{{ previewTitle }}</div>
                  </div>
                  <el-divider style="margin: 12px 0" />
                  <div class="preview-section">
                    <div class="preview-label">内容</div>
                    <pre class="preview-content">{{ previewContent }}</pre>
                  </div>
                  <div class="preview-hint">
                    <el-icon :size="12"><InfoFilled /></el-icon>
                    预览使用示例数据渲染，留空字段将使用内置默认模板
                  </div>
                </div>
              </el-collapse-item>
            </el-collapse>
          </el-form-item>
        </template>

        <template v-else>
          <el-form-item>
            <div class="empty-template-hint">
              <el-icon :size="32" color="var(--el-text-color-placeholder)"><InfoFilled /></el-icon>
              <p>请先在上方选择触发时机，再配置对应的消息模板</p>
            </div>
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, nextTick, onMounted, watch } from 'vue'
import { ElMessage, type InputInstance } from 'element-plus'
import { Plus, Refresh, Promotion, Edit, Delete, InfoFilled } from '@element-plus/icons-vue'
import { listChannels, createChannel, updateChannel, deleteChannel, testChannel } from '@/api/modules/notification'
import type { NotificationChannel } from '@/api/modules/notification'

const loading = ref(false)
const channels = ref<NotificationChannel[]>([])
const testingId = ref<number | null>(null)

const showDialog = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)

const form = reactive({
  name: '',
  type: '' as string,
  enabled: true,
  notify_on_success: false,
  notify_on_failure: true
})

// ========= 触发时机 =========

const triggerEventsMap = reactive<Record<string, boolean>>({
  sync_success: false,
  sync_failure: true,
  sync_conflict: true,
  webhook_received: false,
  webhook_error: true,
  cron_triggered: false,
  backup_success: false,
  backup_failure: true
})

const allTriggerEventKeys = [
  'sync_success', 'sync_failure', 'sync_conflict',
  'webhook_received', 'webhook_error',
  'cron_triggered',
  'backup_success', 'backup_failure'
]

const triggerEventLabels: Record<string, string> = {
  sync_success: '同步成功',
  sync_failure: '同步失败',
  sync_conflict: '同步冲突',
  webhook_received: 'Webhook 接收',
  webhook_error: 'Webhook 错误',
  cron_triggered: '定时任务触发',
  backup_success: '备份成功',
  backup_failure: '备份失败'
}

function triggerEventTagType(event: string): string {
  if (event.includes('success')) return 'success'
  if (event.includes('failure') || event.includes('error') || event.includes('conflict')) return 'danger'
  if (event.includes('received') || event.includes('triggered')) return 'warning'
  return 'info'
}

function parseTriggerEvents(raw: string): string[] {
  if (!raw) return []
  try {
    const arr = JSON.parse(raw)
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

function getTriggerEventsJson(): string {
  const events = allTriggerEventKeys.filter(k => triggerEventsMap[k])
  return JSON.stringify(events)
}

function resetTriggerEvents() {
  allTriggerEventKeys.forEach(k => {
    triggerEventsMap[k] = false
  })
  triggerEventsMap.sync_failure = true
  triggerEventsMap.sync_conflict = true
  triggerEventsMap.webhook_error = true
  triggerEventsMap.backup_failure = true
}

function loadTriggerEvents(raw: string) {
  allTriggerEventKeys.forEach(k => {
    triggerEventsMap[k] = false
  })
  const events = parseTriggerEvents(raw)
  if (events.length > 0) {
    events.forEach(e => {
      if (e in triggerEventsMap) {
        triggerEventsMap[e] = true
      }
    })
  } else {
    triggerEventsMap.sync_failure = true
    triggerEventsMap.backup_failure = true
  }
}

// ========= 事件级模板 =========

const eventTemplates = reactive<Record<string, { title_template: string; content_template: string }>>({})
const activeEventTab = ref('')

const enabledEvents = computed(() => allTriggerEventKeys.filter(k => triggerEventsMap[k]))

function ensureEventTemplate(event: string) {
  if (!eventTemplates[event]) {
    eventTemplates[event] = { title_template: '', content_template: '' }
  }
}

// 当启用的事件变化时，同步 Tab 和模板数据
watch(enabledEvents, (events) => {
  for (const event of events) {
    ensureEventTemplate(event)
  }
  if (!events.includes(activeEventTab.value) && events.length > 0) {
    activeEventTab.value = events[0]!
  }
  if (events.length === 0) {
    activeEventTab.value = ''
  }
}, { immediate: true })

// 当前 Tab 的标题模板（可写 computed）
const currentTitleTemplate = computed({
  get: () => eventTemplates[activeEventTab.value]?.title_template || '',
  set: (val: string) => {
    ensureEventTemplate(activeEventTab.value)
    eventTemplates[activeEventTab.value]!.title_template = val
  }
})

// 当前 Tab 的内容模板（可写 computed）
const currentContentTemplate = computed({
  get: () => eventTemplates[activeEventTab.value]?.content_template || '',
  set: (val: string) => {
    ensureEventTemplate(activeEventTab.value)
    eventTemplates[activeEventTab.value]!.content_template = val
  }
})

// ========= 模板变量 =========

const templateVariables = [
  { name: 'TaskKey', description: '任务标识', example: 'my-sync-task', events: '全部' },
  { name: 'Status', description: '状态码', example: 'success', events: '全部' },
  { name: 'StatusText', description: '状态文字', example: '成功', events: '全部' },
  { name: 'EventType', description: '事件类型', example: 'sync_success', events: '全部' },
  { name: 'EventLabel', description: '事件名称', example: '同步成功', events: '全部' },
  { name: 'Timestamp', description: '时间', example: '2026-02-16 10:30:00', events: '全部' },
  { name: 'RepoKey', description: '仓库标识', example: 'my-repo', events: '全部' },
  { name: 'SourceRemote', description: '源远程仓库', example: 'origin', events: '同步事件' },
  { name: 'SourceBranch', description: '源分支', example: 'main', events: '同步事件' },
  { name: 'TargetRemote', description: '目标远程仓库', example: 'backup', events: '同步事件' },
  { name: 'TargetBranch', description: '目标分支', example: 'main', events: '同步事件' },
  { name: 'CommitRange', description: '提交范围', example: 'abc..def', events: '同步成功' },
  { name: 'Duration', description: '执行耗时', example: '3.2s', events: '同步/备份' },
  { name: 'SyncMode', description: '同步模式', example: 'all-branch', events: '同步事件' },
  { name: 'BranchCount', description: '总分支数', example: '10', events: '同步事件' },
  { name: 'SuccessCount', description: '成功数', example: '8', events: '同步事件' },
  { name: 'FailedCount', description: '失败数', example: '2', events: '同步事件' },
  { name: 'ErrorMessage', description: '错误信息', example: 'push failed', events: '失败/错误/冲突' },
  { name: 'CronExpression', description: 'Cron表达式', example: '0 2 * * *', events: '定时任务' },
  { name: 'WebhookSource', description: 'Webhook来源', example: 'github', events: 'Webhook事件' },
  { name: 'BackupPath', description: '备份路径', example: '/backups/repo.tar.gz', events: '备份事件' },
]

const variableCategories = [
  {
    label: '通用',
    type: '' as const,
    vars: ['TaskKey', 'Status', 'StatusText', 'EventType', 'EventLabel', 'Timestamp', 'RepoKey']
  },
  {
    label: '同步事件',
    type: 'success' as const,
    vars: ['SourceRemote', 'SourceBranch', 'TargetRemote', 'TargetBranch', 'CommitRange', 'Duration', 'SyncMode', 'BranchCount', 'SuccessCount', 'FailedCount']
  },
  {
    label: '错误信息',
    type: 'danger' as const,
    vars: ['ErrorMessage']
  },
  {
    label: '特殊变量',
    type: 'warning' as const,
    vars: ['CronExpression', 'WebhookSource', 'BackupPath']
  }
]

const activeEditor = ref<'title' | 'content'>('content')
const titleInputRef = ref<InputInstance | null>(null)
const contentInputRef = ref<InputInstance | null>(null)
const previewCollapse = ref(['preview'])

// 基础预览数据
const basePreviewData: Record<string, string> = {
  TaskKey: 'my-sync-task', RepoKey: 'my-repo',
  SourceRemote: 'origin', SourceBranch: 'main',
  TargetRemote: 'backup', TargetBranch: 'main',
  CommitRange: 'abc123..def456', Duration: '3.2s',
  SyncMode: 'single-branch', BranchCount: '10',
  SuccessCount: '8', FailedCount: '2',
  ErrorMessage: 'push failed: connection refused',
  CronExpression: '0 2 * * *', WebhookSource: 'github',
  BackupPath: '/backups/repo.tar.gz', Timestamp: '2026-02-20 10:30:00'
}

// 根据当前活动 Tab 生成上下文相关的预览数据
function getPreviewData(): Record<string, string> {
  const event = activeEventTab.value
  const isSuccess = event.includes('success')
  const isFailure = event.includes('failure') || event.includes('error') || event.includes('conflict')
  return {
    ...basePreviewData,
    EventType: event,
    EventLabel: triggerEventLabels[event] || event,
    Status: isSuccess ? 'success' : isFailure ? 'failure' : 'info',
    StatusText: isSuccess ? '成功' : isFailure ? '失败' : '通知',
    ErrorMessage: isFailure ? 'push failed: connection refused' : ''
  }
}

function formatVar(name: string): string {
  return '{{.' + name + '}}'
}

function getVarByName(name: string) {
  return templateVariables.find(v => v.name === name)
}

function handleEditorFocus(editorType: 'title' | 'content') {
  activeEditor.value = editorType
}

function insertVariable(varName: string) {
  const varText = `{{.${varName}}}`
  const targetRef = activeEditor.value === 'title' ? titleInputRef.value : contentInputRef.value

  if (!targetRef) {
    navigator.clipboard.writeText(varText).then(() => {
      ElMessage.success(`已复制 ${varText}`)
    }).catch(() => {
      ElMessage.info(`请手动复制: ${varText}`)
    })
    return
  }

  const inputEl = (targetRef as any).textarea || (targetRef as any).input
  if (!inputEl) return

  const startPos = inputEl.selectionStart ?? 0
  const endPos = inputEl.selectionEnd ?? 0
  const currentValue = activeEditor.value === 'title' ? currentTitleTemplate.value : currentContentTemplate.value

  const newValue = currentValue.slice(0, startPos) + varText + currentValue.slice(endPos)

  if (activeEditor.value === 'title') {
    currentTitleTemplate.value = newValue
  } else {
    currentContentTemplate.value = newValue
  }

  nextTick(() => {
    const newCursorPos = startPos + varText.length
    inputEl.focus()
    inputEl.setSelectionRange(newCursorPos, newCursorPos)
  })

  ElMessage.success({ message: `已插入 ${varText}`, duration: 1000 })
}

function renderTemplate(tmplStr: string): string {
  if (!tmplStr) return ''
  const data = getPreviewData()
  let result = tmplStr
  for (const [key, value] of Object.entries(data)) {
    result = result.replace(new RegExp(`\\{\\{\\.${key}\\}\\}`, 'g'), value)
  }
  return result
}

// 每个事件类型的内置默认标题模板
const defaultEventTitleTemplates: Record<string, string> = {
  sync_success: '[成功] 同步任务 {{.TaskKey}}',
  sync_failure: '[失败] 同步任务 {{.TaskKey}}',
  sync_conflict: '[冲突] 同步任务 {{.TaskKey}}',
  webhook_received: '[Webhook] 收到请求: {{.TaskKey}}',
  webhook_error: '[Webhook错误] {{.TaskKey}}',
  cron_triggered: '[定时] 任务触发: {{.TaskKey}}',
  backup_success: '[备份成功] {{.RepoKey}}',
  backup_failure: '[备份失败] {{.RepoKey}}'
}

// 每个事件类型的内置默认内容模板
const defaultEventContentTemplates: Record<string, string> = {
  sync_success: '任务: {{.TaskKey}}\n状态: {{.StatusText}}\n源: {{.SourceRemote}}/{{.SourceBranch}}\n目标: {{.TargetRemote}}/{{.TargetBranch}}\n时间: {{.Timestamp}}',
  sync_failure: '任务: {{.TaskKey}}\n状态: {{.StatusText}}\n源: {{.SourceRemote}}/{{.SourceBranch}}\n目标: {{.TargetRemote}}/{{.TargetBranch}}\n错误: {{.ErrorMessage}}\n时间: {{.Timestamp}}',
  sync_conflict: '任务: {{.TaskKey}}\n状态: 同步冲突\n源: {{.SourceRemote}}/{{.SourceBranch}}\n目标: {{.TargetRemote}}/{{.TargetBranch}}\n时间: {{.Timestamp}}',
  webhook_received: '任务: {{.TaskKey}}\n状态: Webhook 请求已接收\n来源: {{.WebhookSource}}\n时间: {{.Timestamp}}',
  webhook_error: '任务: {{.TaskKey}}\n状态: Webhook 处理失败\n错误: {{.ErrorMessage}}\n时间: {{.Timestamp}}',
  cron_triggered: '任务: {{.TaskKey}}\n状态: 定时任务已触发\nCron: {{.CronExpression}}\n时间: {{.Timestamp}}',
  backup_success: '仓库: {{.RepoKey}}\n状态: 备份成功\n备份路径: {{.BackupPath}}\n时间: {{.Timestamp}}',
  backup_failure: '仓库: {{.RepoKey}}\n状态: 备份失败\n错误: {{.ErrorMessage}}\n时间: {{.Timestamp}}'
}

const previewTitle = computed(() => {
  const event = activeEventTab.value
  const tmpl = currentTitleTemplate.value || defaultEventTitleTemplates[event] || '[通知] {{.TaskKey}}'
  return renderTemplate(tmpl)
})

const previewContent = computed(() => {
  const event = activeEventTab.value
  const tmpl = currentContentTemplate.value || defaultEventContentTemplates[event] || '任务: {{.TaskKey}}\n状态: {{.StatusText}}\n时间: {{.Timestamp}}'
  return renderTemplate(tmpl)
})

// ========= 渠道配置表单 =========

const configForm = reactive({
  smtp_host: '',
  smtp_port: '',
  username: '',
  password: '',
  from: '',
  to: '',
  webhook_url: '',
  secret: '',
  sign: '',
  security_type: 'none',
  keywords: '',
  url: '',
  method: 'POST',
  content_type: 'application/json'
})

const typeLabels: Record<string, string> = {
  email: '邮件',
  dingtalk: '钉钉',
  wechat: '企业微信',
  lanxin: '蓝信',
  feishu: '飞书',
  webhook: 'Webhook'
}

onMounted(() => {
  loadChannels()
})

watch(() => form.type, () => {
  if (!editingId.value) {
    Object.keys(configForm).forEach(key => {
      (configForm as Record<string, string>)[key] = ''
    })
    configForm.method = 'POST'
    configForm.content_type = 'application/json'
    configForm.security_type = 'none'
  }
})

async function loadChannels() {
  loading.value = true
  try {
    channels.value = await listChannels()
  } catch {
    ElMessage.error('加载通知渠道失败')
  } finally {
    loading.value = false
  }
}

function clearEventTemplates() {
  Object.keys(eventTemplates).forEach(k => delete eventTemplates[k])
}

function openAddDialog() {
  editingId.value = null
  form.name = ''
  form.type = ''
  form.enabled = true
  form.notify_on_success = false
  form.notify_on_failure = true
  resetTriggerEvents()
  clearEventTemplates()
  activeEventTab.value = ''
  Object.keys(configForm).forEach(key => {
    (configForm as Record<string, string>)[key] = ''
  })
  configForm.method = 'POST'
  configForm.content_type = 'application/json'
  configForm.security_type = 'none'
  showDialog.value = true
}

async function openEditDialog(channel: NotificationChannel) {
  editingId.value = channel.id
  form.name = channel.name
  form.type = channel.type
  form.enabled = channel.enabled
  form.notify_on_success = channel.notify_on_success
  form.notify_on_failure = channel.notify_on_failure
  loadTriggerEvents(channel.trigger_events)

  // 加载事件级模板
  clearEventTemplates()
  if (channel.event_templates_json) {
    try {
      const templates: Array<{ event_type: string; title_template: string; content_template: string }> = JSON.parse(channel.event_templates_json)
      for (const t of templates) {
        eventTemplates[t.event_type] = {
          title_template: t.title_template || '',
          content_template: t.content_template || ''
        }
      }
    } catch { /* ignore */ }
  }

  // 解析渠道配置
  try {
    const config = JSON.parse(channel.config || '{}')
    Object.keys(config).forEach(key => {
      if (key in configForm) {
        (configForm as Record<string, string>)[key] = config[key]
      }
    })
    if (!config.security_type) {
      if ((form.type === 'dingtalk' || form.type === 'feishu') && config.secret) {
        configForm.security_type = 'sign'
      } else if (form.type === 'lanxin' && config.sign) {
        configForm.security_type = 'sign'
      } else {
        configForm.security_type = 'none'
      }
    }
  } catch { /* ignore */ }

  showDialog.value = true
}

function getConfigJson(): string {
  const configKeys: Record<string, string[]> = {
    email: ['smtp_host', 'smtp_port', 'username', 'password', 'from', 'to'],
    dingtalk: ['webhook_url', 'security_type', 'secret', 'keywords'],
    wechat: ['webhook_url'],
    lanxin: ['webhook_url', 'security_type', 'sign', 'keywords'],
    feishu: ['webhook_url', 'security_type', 'secret', 'keywords'],
    webhook: ['url', 'method', 'content_type']
  }

  const keys = configKeys[form.type] || []
  const config: Record<string, string> = {}
  keys.forEach(key => {
    const value = (configForm as Record<string, string>)[key]
    if (value) config[key] = value
  })
  return JSON.stringify(config)
}

function buildEventTemplatesJson(): string {
  const list: Array<{ event_type: string; title_template: string; content_template: string }> = []
  for (const event of enabledEvents.value) {
    const tmpl = eventTemplates[event]
    if (tmpl && (tmpl.title_template || tmpl.content_template)) {
      list.push({
        event_type: event,
        title_template: tmpl.title_template,
        content_template: tmpl.content_template
      })
    }
  }
  return list.length > 0 ? JSON.stringify(list) : ''
}

async function handleSave() {
  if (!form.name || !form.type) {
    ElMessage.warning('请填写名称和类型')
    return
  }
  saving.value = true
  try {
    const params = {
      name: form.name,
      type: form.type,
      config: getConfigJson(),
      enabled: form.enabled,
      notify_on_success: form.notify_on_success,
      notify_on_failure: form.notify_on_failure,
      trigger_events: getTriggerEventsJson(),
      title_template: '',
      content_template: '',
      event_templates_json: buildEventTemplatesJson()
    }
    if (editingId.value) {
      await updateChannel({ ...params, id: editingId.value })
      ElMessage.success('渠道更新成功')
    } else {
      await createChannel(params)
      ElMessage.success('渠道创建成功')
    }
    showDialog.value = false
    await loadChannels()
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('保存失败: ' + (err.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await deleteChannel(id)
    ElMessage.success('渠道已删除')
    await loadChannels()
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('删除失败: ' + (err.message || '未知错误'))
  }
}

async function handleTest(id: number) {
  testingId.value = id
  try {
    const result = await testChannel(id, '这是一条测试消息 - Git管理服务')
    if (result.success) {
      ElMessage.success('测试消息发送成功')
    } else {
      ElMessage.error('测试失败: ' + (result.error || '未知错误'))
    }
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('测试失败: ' + (err.message || '未知错误'))
  } finally {
    testingId.value = null
  }
}
</script>

<style scoped>
.notification-manager {
  padding: 8px 0;
}
.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}
.mr-1 {
  margin-right: 4px;
}
.mb-1 {
  margin-bottom: 4px;
}
.trigger-events-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  width: 100%;
}
.trigger-group {
  padding: 8px 12px;
  background: var(--el-fill-color-lighter);
  border-radius: 6px;
}
.trigger-group-title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
  font-weight: 500;
}
.trigger-group .el-checkbox {
  display: block;
  margin-left: 0;
  margin-bottom: 2px;
}

/* 事件模板 Tab */
.event-template-tabs {
  width: 100%;
}
.event-template-tabs :deep(.el-tabs__header) {
  margin-bottom: 0;
}
.event-template-tabs :deep(.el-tabs__item) {
  font-size: 13px;
  padding: 0 16px;
}

/* 变量选择面板 */
.variable-panel {
  width: 100%;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  padding: 14px;
  background: var(--el-fill-color-blank);
}
.var-group {
  margin-bottom: 12px;
}
.var-group:last-of-type {
  margin-bottom: 0;
}
.var-group-header {
  margin-bottom: 8px;
}
.var-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.var-btn {
  font-size: 12px;
  padding: 4px 10px;
  height: auto;
  border-radius: 4px;
  transition: all 0.2s ease;
}
.var-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.12);
}
.active-editor-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 12px;
  padding: 6px 10px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

/* 模板编辑器 */
.template-input {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}
.template-input.editor-active :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px var(--el-color-primary) inset, 0 0 0 3px var(--el-color-primary-light-8);
}
.template-input.editor-active :deep(.el-textarea__inner) {
  box-shadow: 0 0 0 1px var(--el-color-primary) inset, 0 0 0 3px var(--el-color-primary-light-8);
}

/* 实时预览 */
.preview-collapse {
  width: 100%;
  border: none;
}
.preview-collapse :deep(.el-collapse-item__header) {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-secondary);
  height: 36px;
  line-height: 36px;
}
.template-preview {
  background: var(--el-fill-color-lighter);
  border-radius: 6px;
  padding: 14px;
}
.preview-section {
  margin-bottom: 4px;
}
.preview-label {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  font-weight: 500;
  margin-bottom: 4px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.preview-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  padding: 8px 12px;
  background: var(--el-bg-color);
  border-radius: 4px;
  border-left: 3px solid var(--el-color-primary);
}
.preview-content {
  font-size: 13px;
  line-height: 1.6;
  color: var(--el-text-color-regular);
  padding: 10px 12px;
  background: var(--el-bg-color);
  border-radius: 4px;
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: inherit;
}
.preview-hint {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--el-text-color-placeholder);
  margin-top: 10px;
}

/* 空状态提示 */
.empty-template-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
  width: 100%;
  text-align: center;
}
.empty-template-hint p {
  margin-top: 12px;
  color: var(--el-text-color-placeholder);
  font-size: 13px;
}
</style>
