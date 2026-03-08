<template>
  <div class="patch-manager">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>Patch 管理</span>
          <div class="header-actions">
            <el-button type="primary" size="small" @click="openGenerateDialog">
              <el-icon><Plus /></el-icon> 生成 Patch
            </el-button>
            <el-button size="small" @click="loadPatches">
              <el-icon><Refresh /></el-icon> 刷新
            </el-button>
          </div>
        </div>
      </template>

      <!-- Patch 列表 -->
      <el-table :data="patches" v-loading="loading" stripe border size="small">
        <el-table-column prop="name" label="文件名" min-width="200" />
        <el-table-column prop="size" label="大小" width="100">
          <template #default="{ row }">
            {{ formatSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="mod_time" label="修改时间" width="160" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="viewPatch(row)">查看</el-button>
            <el-button size="small" type="primary" @click="downloadPatch(row)">下载</el-button>
            <el-button size="small" type="success" @click="openApplyDialog(row)">应用</el-button>
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
    <el-dialog v-model="showGenerateDialog" title="生成 Patch" width="600px" destroy-on-close>
      <el-form :model="generateForm" label-width="100px">
        <el-form-item label="生成方式">
          <el-radio-group v-model="generateMode">
            <el-radio value="range">分支/Commit 范围</el-radio>
            <el-radio value="commits">指定 Commits</el-radio>
          </el-radio-group>
        </el-form-item>

        <template v-if="generateMode === 'range'">
          <el-form-item label="基准">
            <el-input v-model="generateForm.base" placeholder="分支名、Tag 或 Commit Hash" />
          </el-form-item>
          <el-form-item label="目标">
            <el-input v-model="generateForm.target" placeholder="分支名、Tag 或 Commit Hash" />
          </el-form-item>
        </template>

        <template v-else>
          <el-form-item label="Commit 列表">
            <el-input
              v-model="commitsText"
              type="textarea"
              :rows="4"
              placeholder="每行一个 Commit Hash，如：&#10;abc123&#10;def456"
            />
          </el-form-item>
        </template>

        <el-divider />

        <el-form-item label="保存选项">
          <el-checkbox v-model="savePatch">保存到项目</el-checkbox>
        </el-form-item>

        <template v-if="savePatch">
          <el-form-item label="文件名" required>
            <el-input v-model="patchName" placeholder="如: feature-xxx.patch">
              <template #append>.patch</template>
            </el-input>
          </el-form-item>
          <el-form-item label="保存路径">
            <el-input v-model="customPath" placeholder="默认: patches/（相对于仓库根目录）">
              <template #prepend>仓库/</template>
            </el-input>
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
import { Plus, Refresh } from '@element-plus/icons-vue'
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
import type { PatchInfoDTO, PatchStatsDTO } from '@/types/patch'
import { useNotification } from '@/composables/useNotification'

const props = defineProps<{
  repoKey: string
}>()

const { showSuccess, showError } = useNotification()

const loading = ref(false)
const patches = ref<PatchInfoDTO[]>([])

// 生成
const showGenerateDialog = ref(false)
const generateMode = ref<'range' | 'commits'>('range')
const generateForm = ref({
  base: '',
  target: '',
  commits: [] as string[],
})
const commitsText = ref('')
const savePatch = ref(true)
const patchName = ref('')
const customPath = ref('')
const autoCommit = ref(false)
const commitMessage = ref('')
const generating = ref(false)

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
    patches.value = await listPatches(props.repoKey)
  } catch (e: any) {
    showError('加载失败', e)
  } finally {
    loading.value = false
  }
}

function openGenerateDialog() {
  generateForm.value = { base: '', target: '', commits: [] }
  commitsText.value = ''
  savePatch.value = true
  patchName.value = ''
  customPath.value = ''
  autoCommit.value = false
  commitMessage.value = ''
  generateMode.value = 'range'
  showGenerateDialog.value = true
}

async function handleGenerate() {
  // 验证
  if (generateMode.value === 'range') {
    if (!generateForm.value.base || !generateForm.value.target) {
      ElMessage.warning('请填写基准和目标')
      return
    }
  } else {
    if (!commitsText.value.trim()) {
      ElMessage.warning('请填写 Commit 列表')
      return
    }
    generateForm.value.commits = commitsText.value.split('\n').map(c => c.trim()).filter(Boolean)
  }

  if (savePatch.value && !patchName.value.trim()) {
    ElMessage.warning('请填写文件名')
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
      await savePatchApi({
        repo_key: props.repoKey,
        patch_name: patchName.value.endsWith('.patch') ? patchName.value : patchName.value + '.patch',
        patch_content: content,
        custom_path: customPath.value || undefined,
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
</style>
