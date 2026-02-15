<template>
  <div class="commit-search">
    <!-- 搜索表单 -->
    <el-form inline class="search-form">
      <el-form-item label="分支">
        <el-select v-model="searchParams.ref" placeholder="全部" clearable style="width: 180px">
          <el-option v-for="b in (branches || [])" :key="b" :label="b" :value="b" />
        </el-select>
      </el-form-item>
      <el-form-item label="作者">
        <el-select v-model="searchParams.author" placeholder="全部" clearable filterable style="width: 180px">
          <el-option v-for="a in (authors || [])" :key="a.email" :label="`${a.name} (${a.email})`" :value="a.name" />
        </el-select>
      </el-form-item>
      <el-form-item label="关键词">
        <el-input v-model="searchParams.keyword" placeholder="提交信息关键词" style="width: 180px" clearable />
      </el-form-item>
      <el-form-item label="开始日期">
        <el-date-picker v-model="searchParams.since" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" style="width: 150px" />
      </el-form-item>
      <el-form-item label="结束日期">
        <el-date-picker v-model="searchParams.until" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" style="width: 150px" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch" :loading="loading">
          <el-icon><Search /></el-icon> 搜索
        </el-button>
      </el-form-item>
    </el-form>

    <!-- 搜索结果 -->
    <el-table :data="commits" v-loading="loading" @row-click="viewCommitDetail" class="commit-table" size="small">
      <el-table-column prop="short_hash" label="Commit" width="90">
        <template #default="{ row }">
          <el-text class="mono-text" type="primary">{{ row.short_hash }}</el-text>
        </template>
      </el-table-column>
      <el-table-column prop="author_name" label="作者" width="120" />
      <el-table-column prop="author_date" label="时间" width="160" />
      <el-table-column prop="message" label="提交信息" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.message.split('\n')[0] }}
        </template>
      </el-table-column>
      <el-table-column label="变更" width="150">
        <template #default="{ row }">
          <span class="additions">+{{ row.additions || 0 }}</span>
          <span class="deletions">-{{ row.deletions || 0 }}</span>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination" v-if="total > 0">
      <el-pagination
        v-model:current-page="currentPage"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="handlePageChange"
      />
    </div>

    <!-- 提交详情 -->
    <el-dialog v-model="showDetailDialog" :title="`提交详情: ${selectedCommit?.short_hash}`" width="900px" top="5vh" destroy-on-close>
      <div v-if="selectedCommit" class="commit-detail">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="完整Hash" :span="2">
            <el-text class="mono-text">{{ selectedCommit.hash }}</el-text>
            <el-button size="small" link @click="copyText(selectedCommit.hash)">复制</el-button>
          </el-descriptions-item>
          <el-descriptions-item label="作者">{{ selectedCommit.author_name }} &lt;{{ selectedCommit.author_email }}&gt;</el-descriptions-item>
          <el-descriptions-item label="提交时间">{{ selectedCommit.author_date }}</el-descriptions-item>
          <el-descriptions-item label="提交者">{{ selectedCommit.committer_name }} &lt;{{ selectedCommit.committer_email }}&gt;</el-descriptions-item>
          <el-descriptions-item label="提交者时间">{{ selectedCommit.committer_date }}</el-descriptions-item>
          <el-descriptions-item label="提交信息" :span="2">
            <pre class="commit-message">{{ selectedCommit.message }}</pre>
          </el-descriptions-item>
        </el-descriptions>

        <h4 style="margin: 16px 0 8px">变更文件 ({{ fileChanges.length }})</h4>
        <el-table :data="fileChanges" size="small" max-height="200">
          <el-table-column prop="path" label="文件路径" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="变更" width="120">
            <template #default="{ row }">
              <span class="additions">+{{ row.additions }}</span>
              <span class="deletions">-{{ row.deletions }}</span>
            </template>
          </el-table-column>
        </el-table>

        <h4 style="margin: 16px 0 8px">Diff</h4>
        <div v-loading="diffLoading" class="diff-container">
          <pre v-if="diffContent" class="diff-content">{{ diffContent }}</pre>
          <el-empty v-else-if="!diffLoading" description="无差异内容" />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { searchCommits, getCommitDetail, getCommitDiff } from '@/api/modules/commit'
import type { CommitDetail, FileChange } from '@/api/modules/commit'

const props = defineProps<{
  repoKey: string
  branches?: string[]
  authors?: { name: string; email: string }[]
}>()

const loading = ref(false)
const commits = ref<CommitDetail[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = 20

const searchParams = reactive({
  ref: '',
  author: '',
  keyword: '',
  since: '',
  until: ''
})

const showDetailDialog = ref(false)
const selectedCommit = ref<CommitDetail | null>(null)
const fileChanges = ref<FileChange[]>([])
const diffContent = ref('')
const diffLoading = ref(false)

async function handleSearch() {
  currentPage.value = 1
  await loadCommits()
}

async function loadCommits() {
  loading.value = true
  try {
    const res = await searchCommits(props.repoKey, {
      ...searchParams,
      page: currentPage.value,
      page_size: pageSize
    })
    commits.value = res.commits || []
    total.value = res.total || 0
  } catch {
    ElMessage.error('搜索提交失败')
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) {
  currentPage.value = page
  loadCommits()
}

async function viewCommitDetail(row: CommitDetail) {
  selectedCommit.value = row
  fileChanges.value = []
  diffContent.value = ''
  showDetailDialog.value = true
  diffLoading.value = true

  try {
    const detail = await getCommitDetail(props.repoKey, row.hash)
    selectedCommit.value = detail.commit
    fileChanges.value = detail.files || []
    
    diffContent.value = await getCommitDiff(props.repoKey, row.hash)
  } catch {
    ElMessage.error('加载提交详情失败')
  } finally {
    diffLoading.value = false
  }
}

function statusType(status: string) {
  const types: Record<string, string> = {
    added: 'success',
    modified: 'warning',
    deleted: 'danger',
    renamed: 'info'
  }
  return types[status] || 'info'
}

function copyText(text: string) {
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制')
}
</script>

<style scoped>
.commit-search {
  padding: 8px 0;
}
.search-form {
  margin-bottom: 16px;
}
.commit-table {
  cursor: pointer;
}
.mono-text {
  font-family: monospace;
}
.additions {
  color: #67c23a;
  margin-right: 8px;
}
.deletions {
  color: #f56c6c;
}
.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
.commit-message {
  margin: 0;
  white-space: pre-wrap;
  font-family: inherit;
}
.diff-container {
  max-height: 300px;
  overflow: auto;
  background: #f5f7fa;
  border-radius: 4px;
}
.diff-content {
  margin: 0;
  padding: 12px;
  font-size: 12px;
  line-height: 1.5;
  font-family: monospace;
  white-space: pre;
}
</style>
