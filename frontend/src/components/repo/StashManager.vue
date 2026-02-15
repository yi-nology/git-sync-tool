<template>
  <div class="stash-manager">
    <!-- 工具栏 -->
    <div class="toolbar">
      <el-button type="primary" @click="openSaveDialog">
        <el-icon><Plus /></el-icon> 保存 Stash
      </el-button>
      <el-button @click="loadStashList" :loading="loading">
        <el-icon><Refresh /></el-icon> 刷新
      </el-button>
      <el-popconfirm title="确定要清空所有 Stash 吗？" @confirm="handleClearAll" v-if="stashList.length > 0">
        <template #reference>
          <el-button type="danger" plain>
            <el-icon><Delete /></el-icon> 清空全部
          </el-button>
        </template>
      </el-popconfirm>
    </div>

    <!-- Stash 列表 -->
    <el-table :data="stashList" v-loading="loading" empty-text="暂无 Stash 记录">
      <el-table-column prop="index" label="#" width="60">
        <template #default="{ row }">
          <el-tag size="small">{{ row.index }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="ref" label="引用" width="120">
        <template #default="{ row }">
          <el-text class="mono-text" size="small">{{ row.ref }}</el-text>
        </template>
      </el-table-column>
      <el-table-column prop="message" label="消息" min-width="200" show-overflow-tooltip />
      <el-table-column prop="branch" label="分支" width="150" />
      <el-table-column prop="date" label="时间" width="160" />
      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button-group size="small">
            <el-button @click="handleApply(row.index)" title="应用 (保留 stash)">
              <el-icon><Select /></el-icon> Apply
            </el-button>
            <el-button type="primary" @click="handlePop(row.index)" title="弹出 (应用并删除)">
              <el-icon><Top /></el-icon> Pop
            </el-button>
            <el-popconfirm :title="`确定删除 stash@{${row.index}}?`" @confirm="handleDrop(row.index)">
              <template #reference>
                <el-button type="danger">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </template>
            </el-popconfirm>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>

    <!-- 保存 Stash 对话框 -->
    <el-dialog v-model="showSaveDialog" title="保存 Stash" width="500px" destroy-on-close>
      <el-form :model="saveForm" label-width="100px">
        <el-form-item label="描述信息">
          <el-input v-model="saveForm.message" placeholder="可选，描述此次 stash 的内容" />
        </el-form-item>
        <el-form-item label="包含未跟踪">
          <el-switch v-model="saveForm.includeUntracked" />
          <span class="form-help">包含未被 Git 跟踪的文件</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSaveDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Refresh, Delete, Select, Top } from '@element-plus/icons-vue'
import { listStash, saveStash, applyStash, popStash, dropStash, clearStash } from '@/api/modules/stash'
import type { StashEntry } from '@/api/modules/stash'

const props = defineProps<{
  repoKey: string
}>()

const loading = ref(false)
const stashList = ref<StashEntry[]>([])

const showSaveDialog = ref(false)
const saving = ref(false)
const saveForm = reactive({
  message: '',
  includeUntracked: false
})

onMounted(() => {
  loadStashList()
})

async function loadStashList() {
  loading.value = true
  try {
    const res = await listStash(props.repoKey)
    stashList.value = res?.stashes || []
  } catch {
    ElMessage.error('加载 Stash 列表失败')
  } finally {
    loading.value = false
  }
}

function openSaveDialog() {
  saveForm.message = ''
  saveForm.includeUntracked = false
  showSaveDialog.value = true
}

async function handleSave() {
  saving.value = true
  try {
    await saveStash(props.repoKey, saveForm.message || undefined, saveForm.includeUntracked)
    ElMessage.success('Stash 保存成功')
    showSaveDialog.value = false
    await loadStashList()
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('保存失败: ' + (err.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

async function handleApply(index: number) {
  try {
    await applyStash(props.repoKey, index)
    ElMessage.success(`stash@{${index}} 已应用`)
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('应用失败: ' + (err.message || '未知错误'))
  }
}

async function handlePop(index: number) {
  try {
    await popStash(props.repoKey, index)
    ElMessage.success(`stash@{${index}} 已弹出`)
    await loadStashList()
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('弹出失败: ' + (err.message || '未知错误'))
  }
}

async function handleDrop(index: number) {
  try {
    await dropStash(props.repoKey, index)
    ElMessage.success(`stash@{${index}} 已删除`)
    await loadStashList()
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('删除失败: ' + (err.message || '未知错误'))
  }
}

async function handleClearAll() {
  try {
    await clearStash(props.repoKey)
    ElMessage.success('所有 Stash 已清空')
    await loadStashList()
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('清空失败: ' + (err.message || '未知错误'))
  }
}
</script>

<style scoped>
.stash-manager {
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
.form-help {
  margin-left: 12px;
  color: #909399;
  font-size: 12px;
}
</style>
