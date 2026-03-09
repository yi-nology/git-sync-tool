<template>
  <div class="spec-editor-container">
    <div class="toolbar">
      <div class="toolbar-left">
        <h3>Spec 编辑器</h3>
        <span v-if="currentFile" class="current-file">
          {{ currentFile }}
          <el-tag v-if="isDirty" type="warning" size="small">未保存</el-tag>
        </span>
      </div>
      <div class="toolbar-right">
        <el-button size="small" @click="loadFileTree">
          <el-icon><Refresh /></el-icon>
        </el-button>
        <el-button size="small" :loading="lintingInProgress" @click="lintContent">
          <el-icon><DocumentChecked /></el-icon> 检查
        </el-button>
        <el-button
          size="small"
          :disabled="!isDirty || hasErrors()"
          :loading="savingInProgress"
          @click="saveCurrentFile"
        >
          <el-icon><Download /></el-icon> 保存
        </el-button>
      </div>
    </div>

    <div class="editor-layout">
      <div class="file-tree-panel">
        <div class="tree-header">
          <el-input
            v-model="filterText"
            placeholder="搜索文件"
            clearable
            size="small"
            :prefix-icon="Search"
          />
        </div>
        <el-scrollbar>
          <el-tree
            ref="treeRef"
            :data="fileTree"
            :props="{ label: 'name', children: 'children' }"
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
              </span>
            </template>
          </el-tree>
          <div v-if="fileTree.length === 0 && !loading" class="empty-tree-container">
            <el-empty description="此仓库暂无 .spec 文件">
              <el-button type="primary" @click="showInitDialog = true">
                <el-icon><Plus /></el-icon> 初始化 Spec 文件
              </el-button>
            </el-empty>
          </div>
        </el-scrollbar>
      </div>

      <div class="editor-panel">
        <div class="monaco-container" ref="monacoContainer">
          <div v-if="!content" class="empty-editor">
            <el-empty description="请选择 .spec 文件" />
          </div>
        </div>
        <div class="problems-panel">
          <div class="problems-header">
            <span>问题 ({{ lintIssues.length }})</span>
            <el-tag v-if="errorCount > 0" type="danger" size="small">{{ errorCount }} 错误</el-tag>
            <el-tag v-if="warningCount > 0" type="warning" size="small">{{ warningCount }} 警告</el-tag>
          </div>
          <el-scrollbar>
            <div v-if="lintIssues.length === 0" class="no-problems">
              <el-icon><CircleCheck /></el-icon>
              <span>没有问题</span>
            </div>
            <div
              v-for="issue in lintIssues"
              :key="`${issue.line}-${issue.message}`"
              class="problem-item"
              :class="`problem-${issue.severity}`"
              @click="goToLine(issue.line, issue.column)"
            >
              <el-icon v-if="issue.severity === 'error'" color="#f56c6c"><CircleClose /></el-icon>
              <el-icon v-else-if="issue.severity === 'warning'" color="#e6a23c"><WarningFilled /></el-icon>
              <el-icon v-else color="#909399"><InfoFilled /></el-icon>
              <span class="problem-line">Line {{ issue.line }}</span>
              <span class="problem-message">{{ issue.message }}</span>
            </div>
          </el-scrollbar>
        </div>
      </div>
    </div>

    <!-- 初始化 Spec 文件对话框 -->
    <el-dialog
      v-model="showInitDialog"
      title="初始化 Spec 文件"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form :model="initForm" label-width="100px" :rules="initFormRules" ref="initFormRef">
        <el-form-item label="文件名" prop="filename">
          <el-input v-model="initForm.filename" placeholder="例如: mypackage.spec">
            <template #append>.spec</template>
          </el-input>
          <div class="form-tip">建议使用软件包名称作为文件名</div>
        </el-form-item>
        
        <el-form-item label="Name" prop="name">
          <el-input v-model="initForm.name" placeholder="软件包名称" />
        </el-form-item>
        
        <el-form-item label="Version" prop="version">
          <el-input v-model="initForm.version" placeholder="例如: 1.0.0" />
        </el-form-item>
        
        <el-form-item label="Release" prop="release">
          <el-input v-model="initForm.release" placeholder="例如: 1" />
        </el-form-item>
        
        <el-form-item label="Summary" prop="summary">
          <el-input v-model="initForm.summary" placeholder="软件包简要描述" />
        </el-form-item>
        
        <el-form-item label="License" prop="license">
          <el-select v-model="initForm.license" placeholder="选择许可证" filterable>
            <el-option label="MIT" value="MIT" />
            <el-option label="Apache License 2.0" value="Apache-2.0" />
            <el-option label="GNU General Public License v3.0" value="GPL-3.0" />
            <el-option label="BSD 2-Clause" value="BSD-2-Clause" />
            <el-option label="BSD 3-Clause" value="BSD-3-Clause" />
            <el-option label="Mozilla Public License 2.0" value="MPL-2.0" />
            <el-option label="ISC" value="ISC" />
            <el-option label="Unlicense" value="Unlicense" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="URL" prop="url">
          <el-input v-model="initForm.url" placeholder="项目主页 URL" />
        </el-form-item>
        
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="initForm.description"
            type="textarea"
            :rows="3"
            placeholder="软件包详细描述"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showInitDialog = false">取消</el-button>
        <el-button type="primary" @click="handleInitSpec" :loading="initInProgress">
          创建并打开
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  Folder,
  Document,
  Refresh,
  DocumentChecked,
  Download,
  CircleCheck,
  CircleClose,
  WarningFilled,
  InfoFilled,
  Plus,
} from '@element-plus/icons-vue'
import * as monaco from 'monaco-editor'
import {
  getSpecTree,
  getSpecContent,
  saveSpecContent,
  lintSpec,
  createSpecFile,
} from '@/api/modules/spec'
import type { SpecFileNode, LintIssue } from '@/types/spec'

interface Props {
  repoKey: string
}

const props = defineProps<Props>()

const filterText = ref('')
const fileTree = ref<SpecFileNode[]>([])
const currentFile = ref('')
const content = ref('')
const originalContent = ref('')
const isDirty = ref(false)
const lintIssues = ref<LintIssue[]>([])
const loading = ref(false)
const lintingInProgress = ref(false)
const savingInProgress = ref(false)
const treeRef = ref()
const monacoContainer = ref<HTMLElement>()
let editorInstance: monaco.editor.IStandaloneCodeEditor | null = null

const errorCount = ref(0)
const warningCount = ref(0)

// 防抖定时器
let lintDebounceTimer: ReturnType<typeof setTimeout> | null = null

// 初始化 Spec 文件相关
const showInitDialog = ref(false)
const initInProgress = ref(false)
const initFormRef = ref()
const initForm = ref({
  filename: '',
  name: '',
  version: '1.0.0',
  release: '1',
  summary: '',
  license: 'MIT',
  url: '',
  description: '',
})

const initFormRules = {
  filename: [
    { required: true, message: '请输入文件名', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_-]+$/, message: '文件名只能包含字母、数字、下划线和短横线', trigger: 'blur' },
  ],
  name: [{ required: true, message: '请输入软件包名称', trigger: 'blur' }],
  version: [{ required: true, message: '请输入版本号', trigger: 'blur' }],

// 暴露刷新方法给父组件
defineExpose({
  refresh: loadFileTree,
  clearEditor: () => {
    currentFile.value = ''
    content.value = ''
    originalContent.value = ''
    isDirty.value = false
    lintIssues.value = []
    errorCount.value = 0
    warningCount.value = 0
    if (editorInstance) {
      editorInstance.setValue('')
    }
  }
})
  release: [{ required: true, message: '请输入发布号', trigger: 'blur' }],
  summary: [{ required: true, message: '请输入简要描述', trigger: 'blur' }],
  license: [{ required: true, message: '请选择许可证', trigger: 'change' }],
}

watch(filterText, (val) => {
  treeRef.value?.filter(val)
})

onMounted(async () => {
  await loadFileTree()
  await nextTick()
  initMonaco()
})

onBeforeUnmount(() => {
  if (editorInstance) {
    editorInstance.dispose()
  }
  if (lintDebounceTimer) {
    clearTimeout(lintDebounceTimer)
  }
})

async function loadFileTree() {
  try {
    loading.value = true
    const tree = await getSpecTree(props.repoKey)
    
    // 确保返回的是数组
    fileTree.value = Array.isArray(tree) ? tree : []
    
    // 不再自动弹出引导，只在文件树为空时显示空状态提示
  } catch (error) {
    ElMessage.error('加载文件树失败')
    console.error(error)
    fileTree.value = []
  } finally {
    loading.value = false
  }
}

async function loadFile(path: string) {
  if (isDirty.value) {
    const confirmed = await ElMessageBox.confirm('当前文件未保存，是否继续？', '提示', {
      type: 'warning',
    }).catch(() => false)
    if (!confirmed) return
  }

  try {
    loading.value = true
    const { content: fileContent } = await getSpecContent(path, props.repoKey)
    currentFile.value = path
    content.value = fileContent
    originalContent.value = fileContent
    isDirty.value = false
    lintIssues.value = []

    if (editorInstance) {
      editorInstance.setValue(fileContent)
    }
  } catch (error) {
    ElMessage.error('加载文件失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

function initMonaco() {
  if (!monacoContainer.value) return

  // 注册 RPM Spec 语言
  monaco.languages.register({ id: 'rpmspec' })
  monaco.languages.setMonarchTokensProvider('rpmspec', {
    keywords: [
      'Name', 'Version', 'Release', 'Summary', 'License', 'URL', 'Source0',
      'Patch0', 'BuildArch', 'BuildRoot', 'BuildRequires', 'Requires',
      'Provides', 'Obsoletes', 'Conflicts', '%description', '%prep',
      '%build', '%install', '%clean', '%files', '%changelog', '%package',
      '%post', '%postun', '%pre', '%preun'
    ],
    tokenizer: {
      root: [
        [/#.*/, 'comment'],
        [/%\w+/, 'keyword'],
        [/\$\w+/, 'variable'],
        [/\$\{\w+\}/, 'variable'],
        [/%\{\w+\}/, 'variable'],
        [/[<>=]+/, 'operator'],
        [/\d+\.\d+\.\d+/, 'number'],
        [/\d+/, 'number'],
        [/"([^"]*)"/, 'string'],
        [/'([^']*)'/, 'string'],
      ]
    }
  })

  editorInstance = monaco.editor.create(monacoContainer.value, {
    value: content.value,
    language: 'rpmspec',
    theme: 'vs-dark',
    automaticLayout: true,
    fontSize: 14,
    lineNumbers: 'on',
    minimap: { enabled: true },
    scrollBeyondLastLine: false,
  })

  editorInstance.onDidChangeModelContent(() => {
    content.value = editorInstance?.getValue() || ''
    isDirty.value = content.value !== originalContent.value
    
    // 实时 Linting（防抖 500ms）
    if (lintDebounceTimer) {
      clearTimeout(lintDebounceTimer)
    }
    lintDebounceTimer = setTimeout(() => {
      if (content.value) {
        lintContent()
      }
    }, 500)
  })
}

async function lintContent() {
  if (!content.value) return

  try {
    lintingInProgress.value = true
    const result = await lintSpec(content.value)
    lintIssues.value = result.issues || []

    errorCount.value = lintIssues.value.filter(i => i.severity === 'error').length
    warningCount.value = lintIssues.value.filter(i => i.severity === 'warning').length

    // 更新 Monaco markers
    if (editorInstance) {
      const model = editorInstance.getModel()
      if (model) {
        const markers = lintIssues.value.map(issue => ({
          severity: issue.severity === 'error' ? monaco.MarkerSeverity.Error :
                    issue.severity === 'warning' ? monaco.MarkerSeverity.Warning :
                    monaco.MarkerSeverity.Info,
          message: issue.message,
          startLineNumber: issue.line,
          startColumn: issue.column || 1,
          endLineNumber: issue.end_line || issue.line,
          endColumn: issue.end_column || (issue.column || 1) + 10,
        }))
        monaco.editor.setModelMarkers(model, 'rpmspec', markers)
      }
    }

    if (lintIssues.value.length === 0) {
      ElMessage.success('没有发现问题')
    } else {
      ElMessage.warning(`发现 ${lintIssues.value.length} 个问题`)
    }
  } catch (error) {
    ElMessage.error('Linting 失败')
    console.error(error)
  } finally {
    lintingInProgress.value = false
  }
}

async function saveCurrentFile() {
  if (!currentFile.value) return

  // 保存前验证：检查是否有 error 级别的问题
  if (errorCount.value > 0) {
    ElMessage.warning(`发现 ${errorCount.value} 个错误，请先修复后再保存`)
    return
  }

  try {
    savingInProgress.value = true
    await saveSpecContent(currentFile.value, {
      content: content.value,
      message: `chore(spec): update ${currentFile.value}`,
    }, props.repoKey)

    originalContent.value = content.value
    isDirty.value = false
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
    console.error(error)
  } finally {
    savingInProgress.value = false
  }
}

function hasErrors() {
  return errorCount.value > 0
}

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

function goToLine(line: number, column?: number) {
  if (editorInstance) {
    editorInstance.revealLineInCenter(line)
    editorInstance.setPosition({ lineNumber: line, column: column || 1 })
    editorInstance.focus()
  }
}

async function handleInitSpec() {
  if (!initFormRef.value) return

  try {
    await initFormRef.value.validate()
  } catch {
    return
  }

  try {
    initInProgress.value = true

    const filename = initForm.value.filename.endsWith('.spec')
      ? initForm.value.filename
      : `${initForm.value.filename}.spec`

    // 生成 Spec 文件内容
    const specContent = generateSpecTemplate(initForm.value)

    // 调用后端 API 创建文件
    const result = await createSpecFile({
      repo_key: props.repoKey,
      path: '.',
      name: filename,
      content: specContent,
    })

    ElMessage.success('Spec 文件创建成功')

    // 关闭对话框
    showInitDialog.value = false

    // 重新加载文件树
    await loadFileTree()

    // 打开新创建的文件
    await loadFile(result.path)

    // 重置表单
    initForm.value = {
      filename: '',
      name: '',
      version: '1.0.0',
      release: '1',
      summary: '',
      license: 'MIT',
      url: '',
      description: '',
    }
  } catch (error) {
    ElMessage.error('创建 Spec 文件失败')
    console.error(error)
  } finally {
    initInProgress.value = false
  }
}

function generateSpecTemplate(form: typeof initForm.value): string {
  return `Name:           ${form.name}
Version:        ${form.version}
Release:        ${form.release}%{?dist}
Summary:        ${form.summary}

License:        ${form.license}
URL:            ${form.url || 'https://example.com'}
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  gcc
BuildRequires:  make

%description
${form.description || 'This is the description of ' + form.name}

%prep
%setup -q

%build
%configure
make %{?_smp_mflags}

%install
rm -rf %{buildroot}
%make_install

%files
%doc README.md
%license LICENSE
%{_bindir}/%{name}

%changelog
* $(date '+%a %b %d %Y') Your Name <your.email@example.com> - ${form.version}-${form.release}
- Initial package
`
}
</script>

<style scoped>
.spec-editor-container {
  height: calc(100vh - 140px);
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  color: #d4d4d4;
  border-radius: 4px;
  overflow: hidden;
}

.toolbar {
  height: 50px;
  padding: 0 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #333;
  background: #252526;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-left h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
}

.current-file {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #888;
  font-size: 13px;
}

.toolbar-right {
  display: flex;
  gap: 8px;
}

.editor-layout {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.file-tree-panel {
  width: 250px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #333;
}

.tree-header {
  padding: 12px;
  border-bottom: 1px solid #333;
}

.file-tree-panel :deep(.el-tree) {
  background: transparent;
  color: #d4d4d4;
}

.file-tree-panel :deep(.el-tree-node__content:hover) {
  background-color: #2a2d2e;
}

.file-tree-panel :deep(.el-tree-node.is-current > .el-tree-node__content) {
  background-color: #37373d;
}

.custom-tree-node {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
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

.editor-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.monaco-container {
  flex: 1;
  min-height: 0;
  position: relative;
}

.empty-editor {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1e1e1e;
}

.problems-panel {
  height: 200px;
  border-top: 1px solid #333;
  display: flex;
  flex-direction: column;
}

.problems-header {
  height: 36px;
  padding: 0 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  background: #252526;
  border-bottom: 1px solid #333;
  font-size: 13px;
  font-weight: 500;
}

.problems-panel .el-scrollbar {
  flex: 1;
}

.no-problems {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px;
  color: #4caf50;
  font-size: 14px;
}

.problem-item {
  padding: 6px 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  border-bottom: 1px solid #2d2d2d;
}

.problem-item:hover {
  background: #2a2d2e;
}

.problem-error {
  border-left: 3px solid #f56c6c;
}

.problem-warning {
  border-left: 3px solid #e6a23c;
}

.problem-info {
  border-left: 3px solid #909399;
}

.problem-line {
  color: #888;
  min-width: 60px;
}

.problem-message {
  flex: 1;
}

.empty-tree-container {
  padding: 24px;
  text-align: center;
}

.empty-tree-container :deep(.el-empty__description) {
  color: #888;
}

.form-tip {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}

/* 引导对话框样式 */
:deep(.init-guide-dialog) {
  .el-message-box__content {
    padding: 20px 0;
  }
  
  .el-message-box__message p {
    font-size: 14px;
    color: #606266;
  }
}
</style>
