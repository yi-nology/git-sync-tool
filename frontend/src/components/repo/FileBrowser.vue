<template>
  <div class="file-browser">
    <!-- 工具栏 -->
    <div class="toolbar">
      <el-select v-model="currentRef" placeholder="选择分支/Tag" style="width: 200px" @change="loadTree">
        <el-option v-for="b in (branches || [])" :key="b" :label="b" :value="b" />
      </el-select>
      <el-breadcrumb separator="/" class="path-breadcrumb">
        <el-breadcrumb-item @click="navigateTo('')">
          <el-icon><HomeFilled /></el-icon>
        </el-breadcrumb-item>
        <el-breadcrumb-item 
          v-for="(part, idx) in pathParts" 
          :key="idx"
          @click="navigateTo(pathParts.slice(0, idx + 1).join('/'))">
          {{ part }}
        </el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <!-- 文件列表 -->
    <el-table :data="entries" v-loading="loading" @row-click="handleRowClick" class="file-table" size="small">
      <el-table-column label="名称" min-width="300">
        <template #default="{ row }">
          <div class="file-name">
            <el-icon v-if="row.type === 'dir'" color="#e6a23c"><Folder /></el-icon>
            <el-icon v-else color="#409eff"><Document /></el-icon>
            <span class="name-text">{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="size" label="大小" width="100">
        <template #default="{ row }">
          {{ row.type === 'file' ? formatSize(row.size) : '-' }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button v-if="row.type === 'file'" size="small" link @click.stop="viewFileHistory(row)">
            <el-icon><Clock /></el-icon>
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 文件内容查看 -->
    <el-dialog v-model="showFileDialog" :title="viewingFile?.name" width="80%" top="5vh" destroy-on-close>
      <div v-if="fileContent" class="file-content">
        <div v-if="fileContent.is_binary" class="binary-notice">
          <el-icon><Warning /></el-icon> 二进制文件，无法预览
        </div>
        <pre v-else class="code-block"><code>{{ fileContent.content }}</code></pre>
      </div>
      <div v-else v-loading="fileLoading" style="min-height: 200px"></div>
    </el-dialog>

    <!-- 文件历史 -->
    <el-dialog v-model="showHistoryDialog" :title="`文件历史: ${historyFile}`" width="700px" destroy-on-close>
      <el-table :data="fileHistoryList" v-loading="historyLoading" size="small" max-height="400">
        <el-table-column prop="short_hash" label="Commit" width="80">
          <template #default="{ row }">
            <el-text class="mono-text" size="small">{{ row.short_hash }}</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="author" label="作者" width="120" />
        <el-table-column prop="date" label="时间" width="150" />
        <el-table-column prop="message" label="消息" show-overflow-tooltip />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { HomeFilled, Folder, Document, Clock, Warning } from '@element-plus/icons-vue'
import { getFileTree, getFileBlob, getFileHistory } from '@/api/modules/file'
import type { TreeEntry, BlobContent, FileCommit } from '@/api/modules/file'

const props = defineProps<{
  repoKey: string
  branches?: string[]
}>()

const loading = ref(false)
const currentRef = ref('')
const currentPath = ref('')
const entries = ref<TreeEntry[]>([])

const showFileDialog = ref(false)
const viewingFile = ref<TreeEntry | null>(null)
const fileContent = ref<BlobContent | null>(null)
const fileLoading = ref(false)

const showHistoryDialog = ref(false)
const historyFile = ref('')
const fileHistoryList = ref<FileCommit[]>([])
const historyLoading = ref(false)

const pathParts = computed(() => {
  return currentPath.value ? currentPath.value.split('/').filter(Boolean) : []
})

onMounted(() => {
  if (props.branches && props.branches.length > 0) {
    currentRef.value = props.branches[0]!
  }
  loadTree()
})

async function loadTree() {
  loading.value = true
  try {
    const res = await getFileTree(props.repoKey, {
      ref: currentRef.value || undefined,
      path: currentPath.value || undefined
    })
    // 排序：目录在前，文件在后
    entries.value = (res.entries || []).sort((a, b) => {
      if (a.type === b.type) return a.name.localeCompare(b.name)
      return a.type === 'dir' ? -1 : 1
    })
  } catch (e) {
    ElMessage.error('加载文件列表失败')
  } finally {
    loading.value = false
  }
}

function handleRowClick(row: TreeEntry) {
  if (row.type === 'dir') {
    currentPath.value = row.path
    loadTree()
  } else {
    viewFile(row)
  }
}

function navigateTo(path: string) {
  currentPath.value = path
  loadTree()
}

async function viewFile(file: TreeEntry) {
  viewingFile.value = file
  fileContent.value = null
  showFileDialog.value = true
  fileLoading.value = true
  try {
    fileContent.value = await getFileBlob(props.repoKey, {
      ref: currentRef.value,
      path: file.path
    })
  } catch {
    ElMessage.error('加载文件内容失败')
  } finally {
    fileLoading.value = false
  }
}

async function viewFileHistory(file: TreeEntry) {
  historyFile.value = file.path
  fileHistoryList.value = []
  showHistoryDialog.value = true
  historyLoading.value = true
  try {
    const res = await getFileHistory(props.repoKey, {
      ref: currentRef.value,
      path: file.path,
      limit: 50
    })
    fileHistoryList.value = res || []
  } catch {
    ElMessage.error('加载文件历史失败')
  } finally {
    historyLoading.value = false
  }
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}
</script>

<style scoped>
.file-browser {
  padding: 8px 0;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}
.path-breadcrumb {
  flex: 1;
}
.path-breadcrumb :deep(.el-breadcrumb__item) {
  cursor: pointer;
}
.path-breadcrumb :deep(.el-breadcrumb__item:hover) {
  color: var(--el-color-primary);
}
.file-table {
  cursor: pointer;
}
.file-name {
  display: flex;
  align-items: center;
  gap: 8px;
}
.name-text {
  color: #303133;
}
.file-content {
  max-height: 70vh;
  overflow: auto;
}
.code-block {
  background: #f5f7fa;
  padding: 16px;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 13px;
  line-height: 1.5;
  margin: 0;
}
.binary-notice {
  text-align: center;
  padding: 40px;
  color: #909399;
}
.mono-text {
  font-family: monospace;
}
</style>
