<template>
  <div class="submodule-manager">
    <!-- 工具栏 -->
    <div class="toolbar">
      <el-button type="primary" @click="openAddDialog">
        <el-icon><Plus /></el-icon> 添加 Submodule
      </el-button>
      <el-button @click="handleUpdateAll" :loading="updating">
        <el-icon><Download /></el-icon> 更新全部
      </el-button>
      <el-button @click="loadSubmodules" :loading="loading">
        <el-icon><Refresh /></el-icon> 刷新
      </el-button>
    </div>

    <!-- Submodule 列表 -->
    <el-table :data="submodules" v-loading="loading" empty-text="暂无 Submodule">
      <el-table-column prop="name" label="名称" width="150" />
      <el-table-column prop="path" label="路径" width="200">
        <template #default="{ row }">
          <el-text class="mono-text" size="small">{{ row.path }}</el-text>
        </template>
      </el-table-column>
      <el-table-column prop="url" label="URL" min-width="250" show-overflow-tooltip />
      <el-table-column prop="branch" label="分支" width="100">
        <template #default="{ row }">
          {{ row.branch || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="commit" label="Commit" width="100">
        <template #default="{ row }">
          <el-text class="mono-text" size="small">{{ row.commit?.substring(0, 7) || '-' }}</el-text>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button-group size="small">
            <el-button v-if="row.status === 'uninitialized'" @click="handleInit(row.path)" title="初始化">
              <el-icon><CircleCheck /></el-icon> Init
            </el-button>
            <el-button @click="handleUpdate(row.path)" title="更新">
              <el-icon><Download /></el-icon>
            </el-button>
            <el-button @click="handleSync(row.path)" title="同步URL">
              <el-icon><Connection /></el-icon>
            </el-button>
            <el-popconfirm :title="`确定移除 ${row.name}?`" @confirm="handleRemove(row.path)">
              <template #reference>
                <el-button type="danger" title="移除">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </template>
            </el-popconfirm>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>

    <!-- 添加 Submodule 对话框 -->
    <el-dialog v-model="showAddDialog" title="添加 Submodule" width="550px" destroy-on-close>
      <el-form :model="addForm" label-width="100px">
        <el-form-item label="仓库 URL" required>
          <el-input v-model="addForm.url" placeholder="https://github.com/user/repo.git" />
        </el-form-item>
        <el-form-item label="本地路径" required>
          <el-input v-model="addForm.path" placeholder="相对于仓库根目录的路径" />
        </el-form-item>
        <el-form-item label="跟踪分支">
          <el-input v-model="addForm.branch" placeholder="可选，默认使用默认分支" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAdd" :loading="adding">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Download, Refresh, CircleCheck, Connection, Delete } from '@element-plus/icons-vue'
import { listSubmodules, addSubmodule, initSubmodule, updateSubmodule, syncSubmodule, removeSubmodule } from '@/api/modules/submodule'
import type { SubmoduleInfo } from '@/api/modules/submodule'

const props = defineProps<{
  repoKey: string
}>()

const loading = ref(false)
const updating = ref(false)
const submodules = ref<SubmoduleInfo[]>([])

const showAddDialog = ref(false)
const adding = ref(false)
const addForm = reactive({
  url: '',
  path: '',
  branch: ''
})

onMounted(() => {
  loadSubmodules()
})

async function loadSubmodules() {
  loading.value = true
  try {
    const res = await listSubmodules(props.repoKey)
    submodules.value = res?.submodules || []
  } catch {
    ElMessage.error('加载 Submodule 列表失败')
  } finally {
    loading.value = false
  }
}

function statusType(status: string) {
  const types: Record<string, string> = {
    initialized: 'success',
    uninitialized: 'warning',
    modified: 'danger',
    unknown: 'info'
  }
  return types[status] || 'info'
}

function statusText(status: string) {
  const texts: Record<string, string> = {
    initialized: '已初始化',
    uninitialized: '未初始化',
    modified: '有修改',
    unknown: '未知'
  }
  return texts[status] || status
}

function openAddDialog() {
  addForm.url = ''
  addForm.path = ''
  addForm.branch = ''
  showAddDialog.value = true
}

async function handleAdd() {
  if (!addForm.url || !addForm.path) {
    ElMessage.warning('请填写 URL 和路径')
    return
  }
  adding.value = true
  try {
    await addSubmodule(props.repoKey, addForm.url, addForm.path, addForm.branch || undefined)
    ElMessage.success('Submodule 添加成功')
    showAddDialog.value = false
    await loadSubmodules()
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('添加失败: ' + (err.message || '未知错误'))
  } finally {
    adding.value = false
  }
}

async function handleInit(path: string) {
  try {
    await initSubmodule(props.repoKey, path)
    ElMessage.success('Submodule 初始化成功')
    await loadSubmodules()
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('初始化失败: ' + (err.message || '未知错误'))
  }
}

async function handleUpdate(path: string) {
  try {
    await updateSubmodule(props.repoKey, { path, init: true, recursive: true })
    ElMessage.success('Submodule 更新成功')
    await loadSubmodules()
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('更新失败: ' + (err.message || '未知错误'))
  }
}

async function handleUpdateAll() {
  updating.value = true
  try {
    await updateSubmodule(props.repoKey, { init: true, recursive: true })
    ElMessage.success('所有 Submodule 更新成功')
    await loadSubmodules()
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('更新失败: ' + (err.message || '未知错误'))
  } finally {
    updating.value = false
  }
}

async function handleSync(path: string) {
  try {
    await syncSubmodule(props.repoKey, path)
    ElMessage.success('Submodule URL 同步成功')
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('同步失败: ' + (err.message || '未知错误'))
  }
}

async function handleRemove(path: string) {
  try {
    await removeSubmodule(props.repoKey, path, true)
    ElMessage.success('Submodule 已移除')
    await loadSubmodules()
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('移除失败: ' + (err.message || '未知错误'))
  }
}
</script>

<style scoped>
.submodule-manager {
  padding: 8px 0;
}
.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}
.mono-text {
  font-family: monospace;
}
</style>
