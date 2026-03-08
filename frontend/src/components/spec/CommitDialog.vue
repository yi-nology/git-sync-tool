<template>
  <el-dialog
    v-model="visible"
    title="提交变更"
    width="600px"
    :close-on-click-modal="false"
  >
    <el-form :model="form" label-width="100px">
      <el-form-item label="Commit 模板">
        <el-select v-model="selectedTemplate" placeholder="选择模板" style="width: 100%">
          <el-option
            v-for="template in commitTemplates"
            :key="template.value"
            :label="template.label"
            :value="template.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="Commit 消息">
        <el-input
          v-model="form.message"
          type="textarea"
          :rows="5"
          placeholder="输入 commit 消息"
        />
      </el-form-item>

      <el-form-item label="变更预览">
        <div class="diff-preview">
          <div class="diff-stats">
            <span class="added">+{{ addedLines }}</span>
            <span class="removed">-{{ removedLines }}</span>
          </div>
          <el-scrollbar max-height="300px">
            <pre class="diff-content">{{ diffPreview }}</pre>
          </el-scrollbar>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="committing" @click="handleCommit">
          提交
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useSpecStore } from '@/stores/useSpecStore'
import { useSpecEditor } from '@/composables/useSpecEditor'

const visible = defineModel<boolean>()

const store = useSpecStore()
const { commitChanges, committingInProgress } = useSpecEditor()

const form = ref({
  message: '',
})

const selectedTemplate = ref('')

const commitTemplates = [
  { label: '更新版本', value: 'Update version to ' },
  { label: '修复问题', value: 'Fix: ' },
  { label: '添加功能', value: 'Feat: ' },
  { label: '重构代码', value: 'Refactor: ' },
  { label: '更新文档', value: 'Docs: ' },
]

watch(selectedTemplate, (val) => {
  if (val) {
    form.value.message = val
  }
})

const diffPreview = computed(() => {
  const original = store.originalContent.split('\n')
  const current = store.content.split('\n')

  const diff: string[] = []
  const maxLines = Math.max(original.length, current.length)

  for (let i = 0; i < maxLines; i++) {
    const origLine = original[i] || ''
    const currLine = current[i] || ''

    if (origLine !== currLine) {
      if (origLine && !currLine) {
        diff.push(`- ${origLine}`)
      } else if (!origLine && currLine) {
        diff.push(`+ ${currLine}`)
      } else {
        diff.push(`- ${origLine}`)
        diff.push(`+ ${currLine}`)
      }
    }
  }

  return diff.join('\n') || '没有变更'
})

const addedLines = computed(() => {
  return (diffPreview.value.match(/^\+/gm) || []).length
})

const removedLines = computed(() => {
  return (diffPreview.value.match(/^-/gm) || []).length
})

const committing = computed(() => committingInProgress.value)

async function handleCommit() {
  if (!form.value.message.trim()) {
    return
  }

  const success = await commitChanges(form.value.message)
  if (success) {
    visible.value = false
    form.value.message = ''
    selectedTemplate.value = ''
  }
}
</script>

<style scoped>
.diff-preview {
  width: 100%;
  background: #1e1e1e;
  border: 1px solid #333;
  border-radius: 4px;
  overflow: hidden;
}

.diff-stats {
  padding: 8px 12px;
  border-bottom: 1px solid #333;
  font-size: 13px;
}

.added {
  color: #89d185;
  margin-right: 12px;
}

.removed {
  color: #f48771;
}

.diff-content {
  padding: 12px;
  margin: 0;
  font-size: 12px;
  font-family: 'Consolas', 'Courier New', monospace;
  color: #d4d4d4;
  line-height: 1.5;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
