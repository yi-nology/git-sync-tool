import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { SpecFileNode, LintIssue, LintRule } from '@/types/spec'

export const useSpecStore = defineStore('spec', () => {
  const fileTree = ref<SpecFileNode[]>([])
  const currentFile = ref<string | null>(null)
  const content = ref('')
  const originalContent = ref('')
  const isDirty = ref(false)
  const lintIssues = ref<LintIssue[]>([])
  const rules = ref<LintRule[]>([])
  const loading = ref(false)
  const editorReady = ref(false)

  function setFileTree(tree: SpecFileNode[]) {
    fileTree.value = tree
  }

  function setCurrentFile(path: string | null) {
    currentFile.value = path
  }

  function setContent(newContent: string) {
    content.value = newContent
    isDirty.value = newContent !== originalContent.value
  }

  function setOriginalContent(original: string) {
    originalContent.value = original
    isDirty.value = content.value !== original
  }

  function resetContent() {
    content.value = originalContent.value
    isDirty.value = false
  }

  function setLintIssues(issues: LintIssue[]) {
    lintIssues.value = issues
  }

  function clearLintIssues() {
    lintIssues.value = []
  }

  function setRules(newRules: LintRule[]) {
    rules.value = newRules
  }

  function updateRule(id: string, data: Partial<LintRule>) {
    const index = rules.value.findIndex((r) => r.id === id)
    if (index !== -1 && rules.value[index]) {
      const rule = rules.value[index]
      Object.assign(rule, data)
    }
  }

  function setLoading(value: boolean) {
    loading.value = value
  }

  function setEditorReady(value: boolean) {
    editorReady.value = value
  }

  function getEnabledRuleIds() {
    return rules.value.filter((r) => r.enabled).map((r) => r.id)
  }

  return {
    fileTree,
    currentFile,
    content,
    originalContent,
    isDirty,
    lintIssues,
    rules,
    loading,
    editorReady,
    setFileTree,
    setCurrentFile,
    setContent,
    setOriginalContent,
    resetContent,
    setLintIssues,
    clearLintIssues,
    setRules,
    updateRule,
    setLoading,
    setEditorReady,
    getEnabledRuleIds,
  }
})
