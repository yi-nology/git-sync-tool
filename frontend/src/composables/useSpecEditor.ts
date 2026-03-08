import { ref, watch } from 'vue'
import { useSpecStore } from '@/stores/useSpecStore'
import { useNotification } from './useNotification'
import {
  getSpecTree,
  getSpecContent,
  saveSpecContent,
  lintSpec,
  getLintRules,
  updateLintRule,
  commitSpec,
} from '@/api/modules/spec'
import type { LintIssue } from '@/types/spec'

export function useSpecEditor() {
  const store = useSpecStore()
  const { showSuccess, showError, showWarning } = useNotification()
  const lintingInProgress = ref(false)
  const savingInProgress = ref(false)
  const committingInProgress = ref(false)
  let lintTimer: ReturnType<typeof setTimeout> | null = null

  async function loadFileTree() {
    try {
      store.setLoading(true)
      const tree = await getSpecTree()
      store.setFileTree(tree)
    } catch (error) {
      showError('加载文件树失败', error)
    } finally {
      store.setLoading(false)
    }
  }

  async function loadFile(path: string) {
    if (store.isDirty) {
      const confirmed = window.confirm('当前文件未保存，是否继续？')
      if (!confirmed) return
    }

    try {
      store.setLoading(true)
      const { content } = await getSpecContent(path)
      store.setCurrentFile(path)
      store.setContent(content)
      store.setOriginalContent(content)
      store.clearLintIssues()
    } catch (error) {
      showError('加载文件失败', error)
    } finally {
      store.setLoading(false)
    }
  }

  async function lintContent(content?: string) {
    const contentToLint = content || store.content
    if (!contentToLint) return

    try {
      lintingInProgress.value = true
      const enabledRuleIds = store.getEnabledRuleIds()
      const result = await lintSpec(contentToLint, enabledRuleIds)
      store.setLintIssues(result.issues)
    } catch (error) {
      showError('Linting 失败', error)
    } finally {
      lintingInProgress.value = false
    }
  }

  async function saveCurrentFile(message?: string) {
    if (!store.currentFile) {
      showWarning('没有打开的文件')
      return false
    }

    try {
      savingInProgress.value = true
      await saveSpecContent(store.currentFile, {
        content: store.content,
        message,
      })
      store.setOriginalContent(store.content)
      showSuccess('文件保存成功')
      return true
    } catch (error) {
      showError('保存文件失败', error)
      return false
    } finally {
      savingInProgress.value = false
    }
  }

  async function commitChanges(message: string, content?: string) {
    if (!store.currentFile) {
      showWarning('没有打开的文件')
      return false
    }

    try {
      committingInProgress.value = true
      const contentToCommit = content || store.content
      await commitSpec(store.currentFile, {
        message,
        content: contentToCommit,
      })
      store.setOriginalContent(contentToCommit)
      showSuccess('Commit 成功')
      return true
    } catch (error) {
      showError('Commit 失败', error)
      return false
    } finally {
      committingInProgress.value = false
    }
  }

  async function loadRules() {
    try {
      const rules = await getLintRules()
      store.setRules(rules)
    } catch (error) {
      showError('加载规则失败', error)
    }
  }

  async function toggleRule(id: string, enabled: boolean) {
    try {
      await updateLintRule(id, { enabled })
      store.updateRule(id, { enabled })
      showSuccess(enabled ? '规则已启用' : '规则已禁用')
    } catch (error) {
      showError('更新规则失败', error)
    }
  }

  function hasErrors() {
    return store.lintIssues.some((issue) => issue.severity === 'error')
  }

  function getIssuesBySeverity(severity: 'error' | 'warning' | 'info'): LintIssue[] {
    return store.lintIssues.filter((issue) => issue.severity === severity)
  }

  function debouncedLint() {
    if (lintTimer) {
      clearTimeout(lintTimer)
    }
    lintTimer = setTimeout(() => {
      if (store.editorReady && store.content) {
        lintContent()
      }
    }, 500)
  }

  watch(
    () => store.content,
    () => {
      debouncedLint()
    }
  )

  return {
    lintingInProgress,
    savingInProgress,
    committingInProgress,
    loadFileTree,
    loadFile,
    lintContent,
    saveCurrentFile,
    commitChanges,
    loadRules,
    toggleRule,
    hasErrors,
    getIssuesBySeverity,
  }
}
