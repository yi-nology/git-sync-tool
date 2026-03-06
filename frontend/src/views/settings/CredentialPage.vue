<template>
  <div class="credential-page">
    <div class="page-header">
      <h2>凭证管理</h2>
      <el-button type="primary" @click="showForm = true; editingCredential = undefined">
        <el-icon><Plus /></el-icon>
        新建凭证
      </el-button>
    </div>

    <el-alert
      type="info"
      :closable="false"
      show-icon
      class="mb-4"
    >
      凭证用于 Git 仓库的认证。支持 SSH 密钥（数据库或本地文件）和 HTTP（用户名密码或 Token）。
      配置 URL 匹配模式后，系统将在配置仓库时自动推荐匹配的凭证。
    </el-alert>

    <!-- 凭证列表 -->
    <div v-loading="loading">
      <div v-if="credentials.length === 0 && !loading" class="empty-state">
        <el-empty description="暂无凭证">
          <el-button type="primary" @click="showForm = true">创建第一个凭证</el-button>
        </el-empty>
      </div>

      <div v-else class="credential-list">
        <CredentialCard
          v-for="cred in credentials"
          :key="cred.id"
          :credential="cred"
          @edit="handleEdit"
          @delete="handleDelete"
        />
      </div>
    </div>

    <!-- 测试连接区域 -->
    <el-card v-if="credentials.length > 0" header="测试凭证连接" class="mt-4">
      <el-form inline>
        <el-form-item label="凭证">
          <el-select v-model="testCredId" placeholder="选择凭证" style="width: 200px">
            <el-option v-for="c in credentials" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="远程 URL">
          <el-input v-model="testUrl" placeholder="git@github.com:user/repo.git" style="width: 350px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleTest" :loading="testing" :disabled="!testCredId || !testUrl">
            测试连接
          </el-button>
        </el-form-item>
      </el-form>
      <el-result
        v-if="testResult !== null"
        :icon="testResult.success ? 'success' : 'error'"
        :title="testResult.success ? '连接成功' : '连接失败'"
        :sub-title="testResult.message"
      />
    </el-card>

    <!-- 创建/编辑 Drawer -->
    <el-drawer
      v-model="showForm"
      :title="editingCredential ? '编辑凭证' : '新建凭证'"
      size="500px"
    >
      <CredentialForm
        :credential="editingCredential"
        @success="handleFormSuccess"
        @cancel="showForm = false"
      />
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { listCredentials, deleteCredential, testCredential } from '@/api/modules/credential'
import type { CredentialDTO } from '@/types/credential'
import CredentialCard from '@/components/credential/CredentialCard.vue'
import CredentialForm from '@/components/credential/CredentialForm.vue'

const loading = ref(false)
const credentials = ref<CredentialDTO[]>([])
const showForm = ref(false)
const editingCredential = ref<CredentialDTO | undefined>()

// 测试连接
const testCredId = ref<number>()
const testUrl = ref('')
const testing = ref(false)
const testResult = ref<{ success: boolean; message: string } | null>(null)

async function loadCredentials() {
  loading.value = true
  try {
    credentials.value = await listCredentials()
  } finally {
    loading.value = false
  }
}

function handleEdit(cred: CredentialDTO) {
  editingCredential.value = cred
  showForm.value = true
}

async function handleDelete(cred: CredentialDTO) {
  try {
    await deleteCredential(cred.id)
    ElMessage.success('凭证已删除')
    loadCredentials()
  } catch (e: any) {
    ElMessage.error(e?.message || '删除失败')
  }
}

function handleFormSuccess() {
  showForm.value = false
  editingCredential.value = undefined
  loadCredentials()
}

async function handleTest() {
  if (!testCredId.value || !testUrl.value) return
  testing.value = true
  testResult.value = null
  try {
    testResult.value = await testCredential(testCredId.value, testUrl.value)
  } catch (e: any) {
    testResult.value = { success: false, message: e?.message || '测试失败' }
  } finally {
    testing.value = false
  }
}

onMounted(() => {
  loadCredentials()
})
</script>

<style scoped>
.credential-page h2 {
  margin: 0;
  font-size: 20px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.mb-4 {
  margin-bottom: 16px;
}
.mt-4 {
  margin-top: 16px;
}
.empty-state {
  padding: 40px 0;
}
.credential-list {
  display: grid;
  gap: 12px;
}
</style>
