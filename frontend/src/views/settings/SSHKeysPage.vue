<template>
  <div class="ssh-keys-page">
    <div class="page-header">
      <h2>SSH 密钥管理</h2>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        新增密钥
      </el-button>
    </div>

    <el-card v-loading="loading">
      <el-table :data="sshKeys" stripe>
        <el-table-column prop="name" label="名称" width="180" />
        <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
        <el-table-column prop="key_type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.key_type.toUpperCase() }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="密码保护" width="100" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.has_passphrase" color="#67C23A"><Lock /></el-icon>
            <el-icon v-else color="#909399"><Unlock /></el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="showDetailDialog(row)">查看</el-button>
            <el-button size="small" type="success" @click="showTestDialog(row)">测试</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && sshKeys.length === 0" description="暂无 SSH 密钥" />
    </el-card>

    <!-- 创建密钥对话框 -->
    <el-dialog v-model="createDialogVisible" title="新增 SSH 密钥" width="600px" destroy-on-close>
      <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="createForm.name" placeholder="例如: GitHub Personal Key" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="createForm.description" placeholder="可选，用于备注密钥用途" />
        </el-form-item>
        <el-form-item label="私钥" prop="private_key">
          <el-input 
            v-model="createForm.private_key" 
            type="textarea" 
            :rows="8" 
            placeholder="粘贴 SSH 私钥内容（以 -----BEGIN 开头）"
          />
        </el-form-item>
        <el-form-item label="密码短语" prop="passphrase">
          <el-input 
            v-model="createForm.passphrase" 
            type="password" 
            show-password
            placeholder="如果私钥有密码保护，请输入"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="creating">创建</el-button>
      </template>
    </el-dialog>

    <!-- 查看密钥详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="密钥详情" width="600px">
      <el-descriptions :column="1" border v-if="currentKey">
        <el-descriptions-item label="名称">{{ currentKey.name }}</el-descriptions-item>
        <el-descriptions-item label="描述">{{ currentKey.description || '-' }}</el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag>{{ currentKey.key_type.toUpperCase() }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="密码保护">
          {{ currentKey.has_passphrase ? '是' : '否' }}
        </el-descriptions-item>
        <el-descriptions-item label="公钥">
          <el-input 
            :model-value="currentKey.public_key" 
            type="textarea" 
            :rows="4" 
            readonly
          />
          <el-button size="small" @click="copyPublicKey" style="margin-top: 8px;">
            <el-icon><CopyDocument /></el-icon>
            复制公钥
          </el-button>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(currentKey.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatDate(currentKey.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 测试连接对话框 -->
    <el-dialog v-model="testDialogVisible" title="测试 SSH 连接" width="500px">
      <el-form :model="testForm" label-width="80px">
        <el-form-item label="Git URL">
          <el-input 
            v-model="testForm.url" 
            placeholder="例如: git@github.com:user/repo.git"
          />
        </el-form-item>
      </el-form>
      <div v-if="testResult" class="test-result" :class="testResult.success ? 'success' : 'error'">
        <el-icon v-if="testResult.success"><CircleCheck /></el-icon>
        <el-icon v-else><CircleClose /></el-icon>
        <span>{{ testResult.message }}</span>
      </div>
      <template #footer>
        <el-button @click="testDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleTest" :loading="testing">测试连接</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus, Lock, Unlock, CopyDocument, CircleCheck, CircleClose } from '@element-plus/icons-vue'
import { 
  listDBSSHKeys, 
  createDBSSHKey, 
  deleteDBSSHKey, 
  testDBSSHKey,
  type DBSSHKey,
  type CreateDBSSHKeyReq,
  type TestDBSSHKeyResp
} from '@/api/modules/sshkey'

const loading = ref(false)
const sshKeys = ref<DBSSHKey[]>([])

// 创建表单
const createDialogVisible = ref(false)
const creating = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = ref<CreateDBSSHKeyReq>({
  name: '',
  description: '',
  private_key: '',
  passphrase: '',
})

const createRules: FormRules = {
  name: [{ required: true, message: '请输入密钥名称', trigger: 'blur' }],
  private_key: [{ required: true, message: '请输入私钥内容', trigger: 'blur' }],
}

// 详情对话框
const detailDialogVisible = ref(false)
const currentKey = ref<DBSSHKey | null>(null)

// 测试对话框
const testDialogVisible = ref(false)
const testing = ref(false)
const testForm = ref({ url: '' })
const testResult = ref<TestDBSSHKeyResp | null>(null)
const testKeyId = ref<number>(0)

onMounted(() => {
  fetchSSHKeys()
})

async function fetchSSHKeys() {
  loading.value = true
  try {
    sshKeys.value = await listDBSSHKeys()
  } catch (e) {
    ElMessage.error('获取 SSH 密钥列表失败')
  } finally {
    loading.value = false
  }
}

function showCreateDialog() {
  createForm.value = {
    name: '',
    description: '',
    private_key: '',
    passphrase: '',
  }
  createDialogVisible.value = true
}

async function handleCreate() {
  if (!createFormRef.value) return
  await createFormRef.value.validate()
  
  creating.value = true
  try {
    await createDBSSHKey(createForm.value)
    ElMessage.success('SSH 密钥创建成功')
    createDialogVisible.value = false
    fetchSSHKeys()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || '创建失败')
  } finally {
    creating.value = false
  }
}

function showDetailDialog(key: DBSSHKey) {
  currentKey.value = key
  detailDialogVisible.value = true
}

function showTestDialog(key: DBSSHKey) {
  testKeyId.value = key.id
  testForm.value.url = ''
  testResult.value = null
  testDialogVisible.value = true
}

async function handleTest() {
  if (!testForm.value.url) {
    ElMessage.warning('请输入 Git URL')
    return
  }
  
  testing.value = true
  testResult.value = null
  try {
    testResult.value = await testDBSSHKey(testKeyId.value, { url: testForm.value.url })
  } catch (e: any) {
    testResult.value = {
      success: false,
      message: e?.response?.data?.msg || '测试失败'
    }
  } finally {
    testing.value = false
  }
}

async function handleDelete(key: DBSSHKey) {
  await ElMessageBox.confirm(`确定要删除密钥 "${key.name}" 吗？此操作不可恢复。`, '确认删除', {
    type: 'warning',
  })
  
  try {
    await deleteDBSSHKey(key.id)
    ElMessage.success('删除成功')
    fetchSSHKeys()
  } catch (e) {
    ElMessage.error('删除失败')
  }
}

function copyPublicKey() {
  if (!currentKey.value) return
  navigator.clipboard.writeText(currentKey.value.public_key)
  ElMessage.success('公钥已复制到剪贴板')
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleString('zh-CN')
}
</script>

<style scoped>
.ssh-keys-page {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
}

.test-result {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  border-radius: 4px;
  margin-top: 16px;
}

.test-result.success {
  background-color: #f0f9eb;
  color: #67c23a;
}

.test-result.error {
  background-color: #fef0f0;
  color: #f56c6c;
}
</style>
