<template>
  <AppPage title="仓库列表">
    <template #actions>
      <el-dropdown @command="handleAddCommand">
        <AppButton type="primary" icon="Plus">
          添加仓库
        </AppButton>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="register">
              <el-icon><FolderOpened /></el-icon> 注册本地仓库
            </el-dropdown-item>
            <el-dropdown-item command="clone">
              <el-icon><Download /></el-icon> 克隆远程仓库
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </template>

    <!-- 搜索和筛选 -->
    <AppCard v-if="repoStore.repoList.length > 0 || searchText">
      <div class="filter-section">
        <AppInput
          v-model="searchText"
          placeholder="搜索仓库名称或路径..."
          prefixIcon="Search"
          showClear
        />
      </div>
    </AppCard>

    <!-- 骨架屏 -->
    <TableSkeleton
      v-if="repoStore.loading"
      :rows="5"
      :columns="5"
      :column-widths="['60px', '150px', '250px', '200px', '120px']"
    />

    <!-- 表格 -->
    <AppCard v-else>
      <el-table
        :data="paginatedData"
        stripe
        @sort-change="handleSortChange"
        :default-sort="{ prop: 'id', order: 'ascending' }"
      >
        <el-table-column prop="id" label="ID" width="60" sortable />
        <el-table-column prop="name" label="名称" min-width="150" sortable>
          <template #default="{ row }">
            <el-text tag="b">{{ row.name }}</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" min-width="250" show-overflow-tooltip>
          <template #default="{ row }">
            <el-text type="info" size="small" class="mono-text">{{ row.path }}</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="remote_url" label="远程地址" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-text size="small" truncated v-if="row.remote_url">{{ row.remote_url }}</el-text>
            <el-text size="small" type="info" v-else>无远程仓库</el-text>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-dropdown @command="(cmd: string) => handleCommand(cmd, row)">
              <AppButton type="secondary">
                操作
              </AppButton>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="detail">
                    <el-icon><View /></el-icon> 查看详情
                  </el-dropdown-item>
                  <el-dropdown-item command="branches">
                    <el-icon><Share /></el-icon> 分支管理
                  </el-dropdown-item>
                  <el-dropdown-item command="sync">
                    <el-icon><Refresh /></el-icon> 同步任务
                  </el-dropdown-item>
                  <el-dropdown-item command="delete" divided>
                    <el-text type="danger">
                      <el-icon><Delete /></el-icon> 删除仓库
                    </el-text>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>

        <!-- 空状态 -->
        <template #empty>
          <div class="app-empty">
            <el-icon class="app-empty-icon"><Folder /></el-icon>
            <h3 class="app-empty-title">暂无仓库</h3>
            <p class="app-empty-description">添加您的第一个仓库开始管理</p>
            <AppButton type="primary" @click="router.push('/repos/register')">
              添加第一个仓库
            </AppButton>
          </div>
        </template>
      </el-table>
    </AppCard>

    <!-- 分页 -->
    <div class="pagination-section" v-if="filteredRepos.length > 0">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="filteredRepos.length"
        layout="total, sizes, prev, pager, next, jumper"
        background
      />
    </div>
  </AppPage>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import {
  Delete,
  View,
  Share,
  Refresh,
  FolderOpened,
  Download,
  Folder,
} from '@element-plus/icons-vue'
import { useRepoStore } from '@/stores/useRepoStore'
import { deleteRepo } from '@/api/modules/repo'
import { useNotification } from '@/composables/useNotification'
import TableSkeleton from '@/components/common/TableSkeleton.vue'
import AppPage from '@/components/layout/AppPage.vue'
import AppCard from '@/components/common/AppCard.vue'
import AppButton from '@/components/common/AppButton.vue'
import AppInput from '@/components/common/AppInput.vue'

const router = useRouter()
const repoStore = useRepoStore()
const { showSuccess, showError } = useNotification()

// 搜索
const searchText = ref('')

// 分页
const currentPage = ref(1)
const pageSize = ref(10)

// 排序
const sortProp = ref('id')
const sortOrder = ref<'ascending' | 'descending' | null>('ascending')

onMounted(async () => {
  await repoStore.fetchRepoList()
})

// 过滤后的仓库列表
const filteredRepos = computed(() => {
  let list = [...repoStore.repoList]

  // 搜索过滤
  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    list = list.filter(
      (repo) =>
        repo.name.toLowerCase().includes(search) ||
        repo.path.toLowerCase().includes(search) ||
        (repo.remote_url && repo.remote_url.toLowerCase().includes(search))
    )
  }

  // 排序
  if (sortProp.value && sortOrder.value) {
    list.sort((a: any, b: any) => {
      const aVal = a[sortProp.value]
      const bVal = b[sortProp.value]
      const modifier = sortOrder.value === 'ascending' ? 1 : -1

      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return (aVal - bVal) * modifier
      }

      return String(aVal).localeCompare(String(bVal)) * modifier
    })
  }

  return list
})

// 分页数据
const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredRepos.value.slice(start, end)
})

// 排序变化
const handleSortChange = ({ prop, order }: { prop: string; order: string | null }) => {
  sortProp.value = prop
  sortOrder.value = order as 'ascending' | 'descending' | null
}

// 添加命令
function handleAddCommand(command: string) {
  if (command === 'register') {
    router.push('/repos/register')
  } else if (command === 'clone') {
    router.push('/repos/clone')
  }
}

// 操作命令
function handleCommand(command: string, row: any) {
  switch (command) {
    case 'detail':
      router.push(`/repos/${row.key}`)
      break
    case 'branches':
      router.push(`/repos/${row.key}/branches`)
      break
    case 'sync':
      router.push(`/repos/${row.key}/sync`)
      break
    case 'delete':
      handleDelete(row.key, row.name)
      break
  }
}

// 删除仓库
async function handleDelete(key: string, name: string) {
  try {
    await ElMessageBox.confirm(
      `确定要删除仓库 "${name}" 吗？如果被同步任务使用将无法删除。`,
      '确认删除',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消',
      }
    )

    await deleteRepo(key)
    showSuccess('仓库已删除')
    await repoStore.fetchRepoList()

    // 如果当前页没有数据了，返回上一页
    if (paginatedData.value.length === 0 && currentPage.value > 1) {
      currentPage.value--
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      showError('删除失败', error)
    }
  }
}
</script>

<style scoped>
.filter-section {
  display: flex;
  align-items: center;
}

.pagination-section {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--spacing-md);
  padding: var(--spacing-md) 0;
}

.mono-text {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
}

/* 响应式 */
@media (max-width: 768px) {
  .filter-section {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }

  .pagination-section {
    justify-content: center;
  }

  :deep(.el-pagination) {
    flex-wrap: wrap;
    justify-content: center;
  }
}

/* 深色模式 */
:global(.dark) .mono-text {
  color: var(--text-color-regular);
}
</style>
