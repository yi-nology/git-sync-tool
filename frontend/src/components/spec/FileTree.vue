<template>
  <div class="spec-file-tree">
    <div class="tree-header">
      <el-input
        v-model="filterText"
        placeholder="搜索文件"
        clearable
        :prefix-icon="Search"
      />
    </div>
    <el-scrollbar>
      <el-tree
        ref="treeRef"
        :data="treeData"
        :props="treeProps"
        :filter-node-method="filterNode"
        node-key="path"
        highlight-current
        :expand-on-click-node="false"
        @node-click="handleNodeClick"
      >
        <template #default="{ node, data }">
          <span class="custom-tree-node">
            <el-icon v-if="data.is_dir" class="folder-icon">
              <Folder />
            </el-icon>
            <el-icon v-else class="file-icon">
              <Document />
            </el-icon>
            <span class="node-label">{{ node.label }}</span>
            <span v-if="!data.is_dir && data.size" class="file-size">
              {{ formatSize(data.size) }}
            </span>
          </span>
        </template>
      </el-tree>
    </el-scrollbar>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { ElTree } from 'element-plus'
import { Search, Folder, Document } from '@element-plus/icons-vue'
import { useSpecStore } from '@/stores/useSpecStore'
import { useSpecEditor } from '@/composables/useSpecEditor'
import type { SpecFileNode } from '@/types/spec'

const store = useSpecStore()
const { loadFileTree, loadFile } = useSpecEditor()

const filterText = ref('')
const treeRef = ref<InstanceType<typeof ElTree>>()
const treeData = ref<SpecFileNode[]>([])

const treeProps = {
  label: 'name',
  children: 'children',
}

watch(filterText, (val) => {
  treeRef.value?.filter(val)
})

onMounted(async () => {
  await loadFileTree()
  treeData.value = store.fileTree
})

function filterNode(value: string, data: Record<string, unknown>) {
  if (!value) return true
  const name = data.name as string
  return name.toLowerCase().includes(value.toLowerCase())
}

function handleNodeClick(data: SpecFileNode) {
  if (!data.is_dir) {
    loadFile(data.path)
  }
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}
</script>

<style scoped>
.spec-file-tree {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  border-right: 1px solid #333;
}

.tree-header {
  padding: 16px;
  border-bottom: 1px solid #333;
}

.el-scrollbar {
  flex: 1;
  overflow: auto;
}

.custom-tree-node {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  font-size: 14px;
}

.folder-icon {
  color: #dcb67a;
}

.file-icon {
  color: #519aba;
}

.node-label {
  flex: 1;
}

.file-size {
  color: #888;
  font-size: 12px;
}

:deep(.el-tree) {
  background: transparent;
  color: #d4d4d4;
}

:deep(.el-tree-node__content:hover) {
  background-color: #2a2d2e;
}

:deep(.el-tree-node.is-current > .el-tree-node__content) {
  background-color: #094771;
}
</style>
