<template>
  <div class="spec-editor-page">
    <div class="toolbar">
      <div class="toolbar-left">
        <h2>Spec 编辑器</h2>
        <span v-if="store.currentFile" class="current-file">
          {{ store.currentFile }}
          <el-tag v-if="store.isDirty" type="warning" size="small">未保存</el-tag>
        </span>
      </div>
      <div class="toolbar-right">
        <el-button :icon="Setting" @click="showRuleManager = true">
          规则管理
        </el-button>
        <el-button
          :icon="DocumentChecked"
          :loading="lintingInProgress"
          @click="lintContent"
        >
          检查
        </el-button>
        <el-button
          :icon="Download"
          :disabled="!store.isDirty || hasErrors()"
          :loading="savingInProgress"
          @click="saveCurrentFile"
        >
          保存
        </el-button>
        <el-button
          type="primary"
          :icon="Promotion"
          :disabled="!store.isDirty"
          @click="showCommitDialog = true"
        >
          Commit
        </el-button>
      </div>
    </div>

    <div class="editor-container">
      <div class="file-tree-panel">
        <FileTree />
      </div>
      <div class="editor-panel">
        <div class="monaco-container">
          <SpecMonaco ref="monacoRef" />
        </div>
        <div class="problems-panel">
          <ProblemsPanel @go-to-line="handleGoToLine" />
        </div>
      </div>
    </div>

    <RuleManager v-model="showRuleManager" />
    <CommitDialog v-model="showCommitDialog" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Setting, DocumentChecked, Download, Promotion } from '@element-plus/icons-vue'
import { useSpecStore } from '@/stores/useSpecStore'
import { useSpecEditor } from '@/composables/useSpecEditor'
import FileTree from '@/components/spec/FileTree.vue'
import SpecMonaco from '@/components/spec/SpecMonaco.vue'
import ProblemsPanel from '@/components/spec/ProblemsPanel.vue'
import RuleManager from '@/components/spec/RuleManager.vue'
import CommitDialog from '@/components/spec/CommitDialog.vue'

const store = useSpecStore()
const { lintingInProgress, savingInProgress, lintContent, saveCurrentFile, hasErrors } =
  useSpecEditor()

const monacoRef = ref<InstanceType<typeof SpecMonaco>>()
const showRuleManager = ref(false)
const showCommitDialog = ref(false)

function handleGoToLine(line: number, column?: number) {
  monacoRef.value?.goToLine(line, column)
}
</script>

<style scoped>
.spec-editor-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  color: #d4d4d4;
}

.toolbar {
  height: 60px;
  padding: 0 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #333;
  background: #252526;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.toolbar-left h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}

.current-file {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #888;
  font-size: 14px;
}

.toolbar-right {
  display: flex;
  gap: 12px;
}

.editor-container {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.file-tree-panel {
  width: 280px;
  flex-shrink: 0;
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
}

.problems-panel {
  height: 200px;
  flex-shrink: 0;
}
</style>
