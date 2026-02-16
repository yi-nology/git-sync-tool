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
    <el-dialog v-model="showDialog" :title="editingId ? '编辑通知渠道' : '添加通知渠道'" width="600px" destroy-on-close>
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

        <el-divider content-position="left">消息模板</el-divider>
        <el-form-item label="标题模板">
          <el-input
            v-model="form.title_template"
            placeholder="留空使用默认模板，例如：[{{.StatusText}}] 同步任务 {{.TaskKey}}"
            clearable
            style="font-family: monospace"
          />
        </el-form-item>
        <el-form-item label="内容模板">
          <el-input
            v-model="form.content_template"
            type="textarea"
            :rows="5"
            placeholder="留空使用默认模板，支持 {{.Var}} 语法"
            clearable
            style="font-family: monospace"
          />
        </el-form-item>
        <el-form-item>
          <el-collapse>
            <el-collapse-item title="可用模板变量参考">
              <div class="template-vars-table">
                <table>
                  <thead><tr><th>变量</th><th>说明</th><th>示例</th><th>适用事件</th></tr></thead>
                  <tbody>
                    <tr v-for="v in templateVariables" :key="v.name">
                      <td class="var-name" @click="copyVar(v.name)"><code>{{`{{.${v.name}}}`}}</code></td>
                      <td>{{ v.description }}</td>
                      <td><code>{{ v.example }}</code></td>
                      <td>{{ v.events }}</td>
                    </tr>
                  </tbody>
                </table>
                <div class="vars-tip">点击变量名可复制</div>
              </div>
            </el-collapse-item>
          </el-collapse>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Refresh, Promotion, Edit, Delete } from '@element-plus/icons-vue'
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
  notify_on_failure: true,
  title_template: '',
  content_template: ''
})

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
  // 默认勾选失败类事件
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
    // 向后兼容：根据旧字段推断
    triggerEventsMap.sync_failure = true
    triggerEventsMap.backup_failure = true
  }
}

const templateVariables = [
  { name: 'TaskKey', description: '任务标识', example: 'my-sync-task', events: '全部' },
  { name: 'Status', description: '状态码', example: 'success', events: '全部' },
  { name: 'StatusText', description: '状态文字', example: '成功', events: '全部' },
  { name: 'EventType', description: '事件类型', example: 'sync_success', events: '全部' },
  { name: 'EventLabel', description: '事件名称', example: '同步成功', events: '全部' },
  { name: 'Timestamp', description: '时间', example: '2026-02-16 10:30:00', events: '全部' },
  { name: 'SourceRemote', description: '源远程仓库', example: 'origin', events: '同步事件' },
  { name: 'SourceBranch', description: '源分支', example: 'main', events: '同步事件' },
  { name: 'TargetRemote', description: '目标远程仓库', example: 'backup', events: '同步事件' },
  { name: 'TargetBranch', description: '目标分支', example: 'main', events: '同步事件' },
  { name: 'RepoKey', description: '仓库标识', example: 'my-repo', events: '全部' },
  { name: 'ErrorMessage', description: '错误信息', example: 'push failed', events: '失败/错误/冲突' },
  { name: 'CommitRange', description: '提交范围', example: 'abc..def', events: '同步成功' },
  { name: 'Duration', description: '执行耗时', example: '3.2s', events: '同步/备份' },
  { name: 'CronExpression', description: 'Cron表达式', example: '0 2 * * *', events: '定时任务' },
  { name: 'WebhookSource', description: 'Webhook来源', example: 'github', events: 'Webhook事件' },
  { name: 'BackupPath', description: '备份路径', example: '/backups/repo.tar.gz', events: '备份事件' },
]

function copyVar(name: string) {
  const text = `{{.${name}}}`
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success(`已复制 ${text}`)
  }).catch(() => {
    ElMessage.info(`请手动复制: ${text}`)
  })
}

const configForm = reactive({
  // Email
  smtp_host: '',
  smtp_port: '',
  username: '',
  password: '',
  from: '',
  to: '',
  // DingTalk & WeChat & Feishu
  webhook_url: '',
  secret: '',
  // Lanxin
  sign: '',
  // Security mode
  security_type: 'none',
  keywords: '',
  // Custom Webhook
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

// 切换类型时重置配置
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

function openAddDialog() {
  editingId.value = null
  form.name = ''
  form.type = ''
  form.enabled = true
  form.notify_on_success = false
  form.notify_on_failure = true
  form.title_template = ''
  form.content_template = ''
  resetTriggerEvents()
  Object.keys(configForm).forEach(key => {
    (configForm as Record<string, string>)[key] = ''
  })
  configForm.method = 'POST'
  configForm.content_type = 'application/json'
  showDialog.value = true
}

async function openEditDialog(channel: NotificationChannel) {
  editingId.value = channel.id
  form.name = channel.name
  form.type = channel.type
  form.enabled = channel.enabled
  form.notify_on_success = channel.notify_on_success
  form.notify_on_failure = channel.notify_on_failure
  form.title_template = channel.title_template || ''
  form.content_template = channel.content_template || ''
  loadTriggerEvents(channel.trigger_events)

  // 解析配置
  try {
    const config = JSON.parse(channel.config || '{}')
    Object.keys(config).forEach(key => {
      if (key in configForm) {
        (configForm as Record<string, string>)[key] = config[key]
      }
    })
    // 向后兼容：旧配置没有 security_type，根据 secret/sign 推断
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
      title_template: form.title_template,
      content_template: form.content_template
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
.template-vars-table table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}
.template-vars-table th,
.template-vars-table td {
  padding: 4px 8px;
  border: 1px solid var(--el-border-color-lighter);
  text-align: left;
}
.template-vars-table th {
  background: var(--el-fill-color-lighter);
  font-weight: 500;
}
.template-vars-table .var-name {
  cursor: pointer;
  color: var(--el-color-primary);
}
.template-vars-table .var-name:hover {
  text-decoration: underline;
}
.template-vars-table code {
  font-family: monospace;
  font-size: 12px;
}
.vars-tip {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}
</style>
