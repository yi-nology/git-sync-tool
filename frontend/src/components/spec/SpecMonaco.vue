<template>
  <div ref="editorContainer" class="spec-monaco-editor"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as monaco from 'monaco-editor'
import { useSpecStore } from '@/stores/useSpecStore'
import { registerSpecLanguage } from '@/monaco/spec-language'
import { useSpecEditor } from '@/composables/useSpecEditor'

const store = useSpecStore()
const { saveCurrentFile } = useSpecEditor()
const editorContainer = ref<HTMLDivElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor | null = null
let markers: monaco.editor.IMarkerData[] = []

onMounted(() => {
  if (!editorContainer.value) return

  registerSpecLanguage()

  editor = monaco.editor.create(editorContainer.value, {
    value: store.content,
    language: 'rpmspec',
    theme: 'vs-dark',
    fontSize: 14,
    minimap: { enabled: true },
    automaticLayout: true,
    scrollBeyondLastLine: false,
    renderWhitespace: 'selection',
    lineNumbers: 'on',
    folding: true,
    foldingStrategy: 'indentation',
    tabSize: 4,
    insertSpaces: true,
    wordWrap: 'on',
  })

  editor.onDidChangeModelContent(() => {
    const value = editor?.getValue() || ''
    store.setContent(value)
  })

  store.setEditorReady(true)

  const handleKeydown = (event: KeyboardEvent) => {
    if ((event.metaKey || event.ctrlKey) && event.key === 's') {
      event.preventDefault()
      saveCurrentFile()
    }
  }

  window.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  if (editor) {
    editor.dispose()
  }
  store.setEditorReady(false)
})

watch(
  () => store.content,
  (newContent) => {
    if (editor && editor.getValue() !== newContent) {
      editor.setValue(newContent)
    }
  }
)

watch(
  () => store.lintIssues,
  (issues) => {
    if (!editor) return

    const model = editor.getModel()
    if (!model) return

    markers = issues.map((issue) => ({
      severity:
        issue.severity === 'error'
          ? monaco.MarkerSeverity.Error
          : issue.severity === 'warning'
            ? monaco.MarkerSeverity.Warning
            : monaco.MarkerSeverity.Info,
      message: issue.message,
      startLineNumber: issue.line,
      startColumn: issue.column || 1,
      endLineNumber: issue.end_line || issue.line,
      endColumn: issue.end_column || (issue.column || 1) + 10,
    }))

    monaco.editor.setModelMarkers(model, 'rpmspec', markers)
  },
  { deep: true }
)

function goToLine(line: number, column?: number) {
  if (!editor) return
  editor.revealLineInCenter(line)
  editor.setPosition({ lineNumber: line, column: column || 1 })
  editor.focus()
}

defineExpose({
  goToLine,
})
</script>

<style scoped>
.spec-monaco-editor {
  width: 100%;
  height: 100%;
  min-height: 400px;
}
</style>
