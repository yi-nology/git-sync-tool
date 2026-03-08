<template>
  <div class="patch-manager">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span>Patch 管理</span>
            <el-tag v-if="seriesStats" type="info" size="small">
              {{ seriesStats.applied_count }}/{{ seriesStats.total_patches }} 已应用
            </el-tag>
          </div>
          <div class="header-actions">
            <el-button
              v-if="seriesStats && seriesStats.can_apply_next"
              type="success"
              size="small"
              @click="applyNextPatch"
            >
              <el-icon><ArrowRight /></el-icon> 应用下一个 ({{ getNextPatchName() }})
            </el-button>
            <el-button
              v-if="seriesStats && seriesStats.pending_count > 1"
              type="warning"
              size="small"
              @click="applyAllPending"
            >
              <el-icon><Finished /></el-icon> 批量应用 ({{ seriesStats.pending_count }}个)
            </el-button>
            <el-button type="primary" size="small" @click="openGenerateDialog">
              <el-icon><Plus /></el-icon> 生成 Patch
            </el-button>
            <el-button size="small" @click="loadPatches">
              <el-icon><Refresh /></el-icon> 刷新
            </el-button>
          </div>
        </div>
      </template>

      <!-- 进度条 -->
      <div v-if="seriesStats && seriesStats.total_patches > 0" class="progress-section">
        <el-progress
          :percentage="getProgress()"
          :status="seriesStats.conflict_count > 0 ? 'exception' : 'success'"
        />
        <div class="progress-text">
          已应用 {{ seriesStats.applied_count }} / 共 {{ seriesStats.total_patches }} 个 patch
          <span v-if="seriesStats.conflict_count > 0" class="error-text">
            ({{ seriesStats.conflict_count }} 个冲突)
          </span>
        </div>
      </div>

      <!-- Patch 列表 -->
      <el-table :data="patches" v-loading="loading" stripe border size="small">
        <el-table-column prop="sequence" label="序号" width="70" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ String(row.sequence).padStart(3, '0') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="文件名" min-width="200">
          <template #default="{ row }">
            <div class="patch-name">
              <el-icon v-if="row.is_applied" color="#67C23A"><CircleCheck /></el-icon>
              <el-icon v-else-if="row.can_apply" color="#E6A23C"><Clock /></el-icon>
              <el-icon v-else color="#F56C6C"><Warning /></el-icon>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_applied" type="success" size="small">已应用</el-tag>
            <el-tag v-else-if="row.can_apply" type="warning" size="small">待应用</el-tag>
            <el-tag v-else type="danger" size="small">冲突</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="size" label="大小" width="100">
          <template #default="{ row }">
            {{ formatSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="mod_time" label="修改时间" width="160" />
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="viewPatch(row)">查看</el-button>
            <el-button size="small" type="primary" @click="downloadPatch(row)">下载</el-button>
            <el-button 
              size="small" 
              type="success" 
              @click="openApplyDialog(row)"
              :disabled="!row.can_apply && !row.is_applied"
            >
              {{ row.is_applied ? '已应用' : '应用' }}
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无 Patch 文件" :image-size="60">
            <el-button type="primary" size="small" @click="openGenerateDialog">创建第一个 Patch</el-button>
          </el-empty>
        </template>
      </el-table>
    </el-card>

    <!-- 生成 Patch 对话框 -->
    <el-dialog v-model="showGenerateDialog" title="生成 Patch" width="700px" destroy-on-close @open="loadDialogData">
      <el-form :model="generateForm" label-width="120px">
        <el-form-item label="生成方式">
          <el-radio-group v-model="generateMode">
            <el-radio value="range">分支/Tag/Commit 范围</el-radio>
            <el-radio value="commits">选择 Commits</el-radio>
          </el-radio-group>
        </el-form-item>

        <template v-if="generateMode === 'range'">
          <el-form-item label="基准（起点）">
            <el-select
              v-model="generateForm.base"
              filterable
              allow-create
              placeholder="选择或输入基准（分支/Tag/Commit）"
              style="width: 100%"
            >
              <el-option-group label="分支">
                <el-option
                  v-for="branch in branches"
                  :key="branch.name"
                  :label="branch.name"
                  :value="branch.name"
                />
              </el-option-group>
              <el-option-group label="Tags">
                <el-option
                  v-for="tag in tags"
                  :key="tag"
                  :label="tag"
                  :value="tag"
                />
              </el-option-group>
              <el-option-group label="最近 Commits">
                <el-option
                  v-for="commit in recentCommits"
                  :key="commit.hash"
                  :label="`${commit.short_hash} - ${commit.message.slice(0, 50)}`"
                  :value="commit.hash"
                />
              </el-option-group>
            </el-select>
          </el-form-item>
          <el-form-item label="目标（终点）">
            <el-select
              v-model="generateForm.target"
              filterable
              allow-create
              placeholder="选择或输入目标（分支/Tag/Commit）"
              style="width: 100%"
            >
              <el-option-group label="分支">
                <el-option
                  v-for="branch in branches"
                  :key="branch.name"
                  :label="branch.name"
                  :value="branch.name"
                />
              </el-option-group>
              <el-option-group label="Tags">
                <el-option
                  v-for="tag in tags"
                  :key="tag"
                  :label="tag"
                  :value="tag"
                />
              </el-option-group>
              <el-option-group label="最近 Commits">
                <el-option
                  v-for="commit in recentCommits"
                  :key="commit.hash"
                  :label="`${commit.short_hash} - ${commit.message.slice(0, 50)}`"
                  :value="commit.hash"
                />
              </el-option-group>
            </el-select>
          </el-form-item>
        </template>

        <template v-else>
          <el-form-item label="选择 Commits">
            <el-select
              v-model="selectedCommits"
              multiple
              filterable
              placeholder="选择要生成 patch 的 commits（可多选）"
              style="width: 100%"
            >
              <el-option
                v-for="commit in recentCommits"
                :key="commit.hash"
                :label="`${commit.short_hash} - ${commit.message.slice(0, 60)} (${commit.author_name})`"
                :value="commit.hash"
              />
            </el-select>
            <div class="hint">提示：可多选，按 Ctrl/Cmd 点击选择多个</div>
          </el-form-item>
        </template>

        <el-divider />

        <el-form-item label="保存选项">
          <el-checkbox v-model="savePatch">保存到项目</el-checkbox>
        </el-form-item>

        <template v-if="savePatch">
          <el-form-item label="文件名" required>
            <el-input v-model="patchName" placeholder="输入描述，如: feature-login、fix-bug">
              <template #prepend>{{ getNextPatchPrefix() }}</template>
              <template #append>.patch</template>
            </el-input>
            <div class="hint">系统自动生成序号，你只需输入描述部分</div>
          </el-form-item>
          <el-form-item label="保存路径">
            <el-cascader
              v-model="selectedPath"
              :options="pathOptions"
              :props="{ checkStrictly: true, emitPath: false, label: 'name', value: 'path' }"
              filterable
              clearable
              placeholder="选择保存目录（默认: patches/）"
              style="width: 100%"
            />
            <div class="path-hint">留空则保存到仓库的 patches 目录</div>
          </el-form-item>
          <el-form-item>
            <el-checkbox v-model="autoCommit">立即提交到 Git</el-checkbox>
          </el-form-item>
          <el-form-item v-if="autoCommit" label="提交消息">
            <el-input
              v-model="commitMessage"
              placeholder="如: chore: add feature-xxx patch"
            />
            <div class="hint">快捷选项：
              <el-button size="small" link @click="commitMessage = 'chore: add patch for ' + (patchName || 'feature')">chore: add patch</el-button>
              <el-button size="small" link @click="commitMessage = 'feat: add patch'">feat: add patch</el-button>
            </div>
          </el-form-item>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="showGenerateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleGenerate" :loading="generating">
          {{ savePatch ? '生成并保存' : '生成并下载' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 查看 Patch 对话框 -->
    <el-dialog v-model="showViewDialog" title="Patch 内容" width="800px" destroy-on-close>
      <el-input
        v-model="patchContent"
        type="textarea"
        :rows="20"
        readonly
        class="patch-content"
      />
      <template #footer>
        <el-button @click="showViewDialog = false">关闭</el-button>
        <el-button type="primary" @click="copyContent">复制内容</el-button>
      </template>
    </el-dialog>

    <!-- 应用 Patch 对话框 -->
    <el-dialog v-model="showApplyDialog" title="应用 Patch" width="700px" destroy-on-close>
      <el-alert v-if="!patchStats" type="info" :closable="false" class="mb-4">
        应用 Patch 将修改工作区文件，请确保已提交或暂存当前更改。
      </el-alert>

      <el-alert v-else-if="!patchStats.can_apply" type="error" :closable="false" class="mb-4">
        <template #title>此 Patch 无法应用</template>
        <div class="error-detail">{{ patchStats.error }}</div>
      </el-alert>

      <el-alert v-else type="success" :closable="false" class="mb-4">
        <template #title>Patch 可以应用</template>
        <pre class="stat-output">{{ patchStats.stat }}</pre>
      </el-alert>

      <el-form :model="applyForm" label-width="100px">
        <el-form-item label="Patch 文件">
          <el-input :value="applyForm.patchName" readonly />
        </el-form-item>

        <el-form-item label="提交消息">
          <el-input
            v-model="applyForm.commit_message"
            type="textarea"
            :rows="3"
            placeholder="留空则不自动提交，仅应用到工作区"
          />
          <div class="hint">快捷选项：
            <el-button size="small" link @click="applyForm.commit_message = 'feat: apply patch ' + applyForm.patchName">feat: apply patch</el-button>
            <el-button size="small" link @click="applyForm.commit_message = 'fix: apply patch'">fix: apply patch</el-button>
          </div>
        </el-form-item>

        <el-form-item v-if="applyForm.commit_message">
          <el-checkbox v-model="applyForm.sign_off">添加 Signed-off-by</el-checkbox>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showApplyDialog = false">取消</el-button>
        <el-button
          type="primary"
          @click="handleApply"
          :loading="applying"
          :disabled="patchStats && !patchStats.can_apply"
        >
          应用
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, CircleCheck, Clock, Warning, ArrowRight, Finished } from '@element-plus/icons-vue'
import {
  generatePatch,
  savePatch as savePatchApi,
  listPatches,
  getPatchContent,
  getPatchDownloadUrl,
  applyPatch,
  checkPatch,
  deletePatch,
} from '@/api/modules/patch'
import { getBranchList } from '@/api/modules/branch'
import { getTagList } from '@/api/modules/branch'
import { searchCommits } from '@/api/modules/commit'
import { getFileTree } from '@/api/modules/file'
import type { PatchInfoDTO, PatchStatsDTO } from '@/types/patch'
import type { BranchInfo } from '@/types/branch'
import type { CommitDetail } from '@/api/modules/commit'
import { useNotification } from '@/composables/useNotification'

const props = defineProps<{
  repoKey: string
}>()

const { showSuccess, showError } = useNotification()

const loading = ref(false)
const patches = ref<PatchInfoDTO[]>([])
const seriesStats = ref<any>(null)

// 生成
const showGenerateDialog = ref(false)
const generateMode = ref<'range' | 'commits'>('range')
const generateForm = ref({
  base: '',
  target: '',
  commits: [] as string[],
})
const selectedCommits = ref<string[]>([])
const savePatch = ref(true)
const patchName = ref('')
const selectedPath = ref('')
const autoCommit = ref(false)
const commitMessage = ref('')
const generating = ref(false)

// 选择器数据
const branches = ref<BranchInfo[]>([])
const tags = ref<string[]>([])
const recentCommits = ref<CommitDetail[]>([])
const pathOptions = ref<any[]>([])

// 查看
const showViewDialog = ref(false)
const patchContent = ref('')

// 应用
const showApplyDialog = ref(false)
const applyForm = ref({
  patchPath: '',
  patchName: '',
  commit_message: '',
  sign_off: false,
})
const applying = ref(false)
const patchStats = ref<PatchStatsDTO | null>(null)

onMounted(() => {
  loadPatches()
})

async function loadPatches() {
  loading.value = true
  try {
    const result = await listPatches(props.repoKey)
    patches.value = result

    // 计算统计信息
    const applied = patches.value.filter(p => p.is_applied).length
    const pending = patches.value.filter(p => !p.is_applied && p.can_apply).length
    const conflict = patches.value.filter(p => !p.is_applied && !p.can_apply).length

    // 找到下一个待应用的 patch
    const nextIndex = patches.value.findIndex(p => !p.is_applied && p.can_apply)

    seriesStats.value = {
      total_patches: patches.value.length,
      applied_count: applied,
      pending_count: pending,
      conflict_count: conflict,
      can_apply_next: nextIndex >= 0,
      next_patch_index: nextIndex,
    }
  } catch (e: any) {
    showError('加载失败', e)
  } finally {
    loading.value = false
  }
}

async function loadDialogData() {
  // 加载分支列表
  try {
    const res = await getBranchList(props.repoKey, { page_size: 100 })
    branches.value = res.list || []
  } catch (e) {
    console.error('Failed to load branches:', e)
  }

  // 加载 tag 列表
  try {
    tags.value = await getTagList(props.repoKey)
  } catch (e) {
    console.error('Failed to load tags:', e)
  }

  // 加载最近 commits
  try {
    const res = await searchCommits(props.repoKey, { page_size: 50 })
    recentCommits.value = res.commits
  } catch (e) {
    console.error('Failed to load commits:', e)
  }

  // 加载文件树（用于选择保存路径）
  try {
    const res = await getFileTree(props.repoKey, { recursive: true })
    pathOptions.value = buildPathTree(res.entries)
  } catch (e) {
    console.error('Failed to load file tree:', e)
  }
}

function buildPathTree(entries: any[]): any[] {
  const tree: any[] = [{ name: '根目录', path: '' }]

  const dirs = new Set<string>()
  entries.forEach((e: any) => {
    if (e.type === 'dir') {
      const parts = e.path.split('/')
      let path = ''
      parts.forEach((part: string) => {
        path = path ? `${path}/${part}` : part
        dirs.add(path)
      })
    }
  })

  dirs.forEach(path => {
    tree.push({
      name: path,
      path: path,
    })
  })

  return tree
}

function openGenerateDialog() {
  generateForm.value = { base: '', target: '', commits: [] }
  selectedCommits.value = []
  savePatch.value = true
  patchName.value = ''
  selectedPath.value = ''
  autoCommit.value = false
  commitMessage.value = ''
  generateMode.value = 'range'
  showGenerateDialog.value = true
}

async function handleGenerate() {
  // 验证
  if (generateMode.value === 'range') {
    if (!generateForm.value.base || !generateForm.value.target) {
      ElMessage.warning('请选择基准和目标')
      return
    }
  } else {
    if (selectedCommits.value.length === 0) {
      ElMessage.warning('请选择至少一个 Commit')
      return
    }
    generateForm.value.commits = selectedCommits.value
  }

  if (savePatch.value && !patchName.value.trim()) {
    ElMessage.warning('请填写文件名描述')
    return
  }

  if (savePatch.value && autoCommit.value && !commitMessage.value.trim()) {
    ElMessage.warning('请填写提交消息')
    return
  }

  generating.value = true
  try {
    const req: any = { repo_key: props.repoKey }
    if (generateMode.value === 'range') {
      req.base = generateForm.value.base
      req.target = generateForm.value.target
    } else {
      req.commits = generateForm.value.commits
    }

    const result = await generatePatch(req)
    const content = result.content

    if (savePatch.value) {
      // 保存到项目
      const prefix = getNextPatchPrefix()
      const fullName = prefix + (patchName.value.endsWith('.patch') ? patchName.value : patchName.value + '.patch')

      await savePatchApi({
        repo_key: props.repoKey,
        patch_name: fullName,
        patch_content: content,
        custom_path: selectedPath.value || undefined,
        commit_message: autoCommit.value ? commitMessage.value : undefined,
      })
      showSuccess('Patch 已保存' + (autoCommit.value ? '并提交到 Git' : ''))
      loadPatches()
    } else {
      // 下载
      downloadContent(content, patchName.value || 'patch.patch')
    }

    showGenerateDialog.value = false
  } catch (e: any) {
    showError('生成失败', e)
  } finally {
    generating.value = false
  }
}

async function viewPatch(patch: PatchInfoDTO) {
  try {
    const result = await getPatchContent(patch.path)
    patchContent.value = result.content
    showViewDialog.value = true
  } catch (e: any) {
    showError('读取失败', e)
  }
}

function downloadPatch(patch: PatchInfoDTO) {
  const url = getPatchDownloadUrl(patch.path)
  window.open(url, '_blank')
}

function downloadContent(content: string, filename: string) {
  const blob = new Blob([content], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

async function openApplyDialog(patch: PatchInfoDTO) {
  applyForm.value = {
    patchPath: patch.path,
    patchName: patch.name,
    commit_message: '',
    sign_off: false,
  }
  patchStats.value = null
  showApplyDialog.value = true

  // 自动检查 patch 是否可以应用
  try {
    patchStats.value = await checkPatch(props.repoKey, patch.path)
  } catch (e: any) {
    console.error('Failed to check patch:', e)
  }
}

async function handleApply() {
  applying.value = true
  try {
    await applyPatch({
      repo_key: props.repoKey,
      patch_path: applyForm.value.patchPath,
      commit_message: applyForm.value.commit_message || undefined,
      sign_off: applyForm.value.sign_off,
    })
    showSuccess('Patch 已应用')
    showApplyDialog.value = false
  } catch (e: any) {
    showError('应用失败', e)
  } finally {
    applying.value = false
  }
}

async function handleDelete(patch: PatchInfoDTO) {
  try {
    await ElMessageBox.confirm(`确定要删除 "${patch.name}" 吗？`, '确认删除', {
      type: 'warning',
    })
    await deletePatch(props.repoKey, patch.path)
    showSuccess('已删除')
    loadPatches()
  } catch (e: any) {
    if (e !== 'cancel') {
      showError('删除失败', e)
    }
  }
}

function copyContent() {
  navigator.clipboard.writeText(patchContent.value)
  showSuccess('已复制到剪贴板')
}

function getProgress(): number {
  if (!seriesStats.value || seriesStats.value.total_patches === 0) return 0
  return Math.round((seriesStats.value.applied_count / seriesStats.value.total_patches) * 100)
}

function getNextPatchName(): string {
  if (!seriesStats.value || seriesStats.value.next_patch_index < 0) return ''
  const patch = patches.value[seriesStats.value.next_patch_index]
  return patch ? patch.name : ''
}

function getNextPatchPrefix(): string {
  // 自动生成下一个 patch 的序号前缀
  // 例如：如果已有 001-base.patch, 002-feature.patch
  // 则下一个前缀为 003-
  if (patches.value.length === 0) {
    return '001-'
  }

  // 找到最大的序号
  let maxSeq = 0
  patches.value.forEach(p => {
    if (p.sequence > maxSeq) {
      maxSeq = p.sequence
    }
  })

  // 下一个序号
  const nextSeq = maxSeq + 1
  return String(nextSeq).padStart(3, '0') + '-'
}

async function applyNextPatch() {
  if (!seriesStats.value || seriesStats.value.next_patch_index < 0) {
    ElMessage.warning('没有待应用的 patch')
    return
  }

  const patch = patches.value[seriesStats.value.next_patch_index]
  if (patch) {
    await openApplyDialog(patch)
  }
}

async function applyAllPending() {
  if (!seriesStats.value || seriesStats.value.pending_count === 0) {
    ElMessage.warning('没有待应用的 patch')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要批量应用 ${seriesStats.value.pending_count} 个待应用的 patch 吗？`,
      '批量应用 Patch',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消',
      }
    )

    // 依次应用所有待应用的 patch
    const pendingPatches = patches.value.filter(p => !p.is_applied && p.can_apply)
    for (const patch of pendingPatches) {
      try {
        await applyPatch({
          repo_key: props.repoKey,
          patch_path: patch.path,
          commit_message: `feat: apply patch ${patch.name}`,
        })
      } catch (e: any) {
        ElMessage.error(`应用 ${patch.name} 失败: ${e.message || e}`)
        break
      }
    }

    ElMessage.success(`已成功应用 ${pendingPatches.length} 个 patch`)
    loadPatches()
  } catch (e) {
    if (e !== 'cancel') {
      console.error('Batch apply failed:', e)
    }
  }
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}
</script>

<style scoped>
.patch-manager {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.path-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.patch-content :deep(textarea) {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
}

.mb-4 {
  margin-bottom: 16px;
}

.error-detail {
  margin-top: 8px;
  font-size: 12px;
  white-space: pre-wrap;
}

.stat-output {
  margin-top: 8px;
  padding: 8px;
  background: #f5f5f5;
  border-radius: 4px;
  font-size: 12px;
  max-height: 200px;
  overflow: auto;
}

.patch-name {
  display: flex;
  align-items: center;
  gap: 6px;
}

.patch-name .el-icon {
  flex-shrink: 0;
}

.progress-section {
  margin-bottom: 16px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.progress-text {
  margin-top: 8px;
  font-size: 13px;
  color: #606266;
}

.error-text {
  color: #F56C6C;
  margin-left: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
</style>
