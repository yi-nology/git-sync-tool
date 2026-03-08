<template>
  <el-dialog
    v-model="visible"
    title="规则管理"
    width="700px"
    :close-on-click-modal="false"
  >
    <div class="rule-manager">
      <div class="toolbar">
        <el-input
          v-model="searchText"
          placeholder="搜索规则"
          clearable
          :prefix-icon="Search"
          style="width: 300px"
        />
        <el-select v-model="categoryFilter" placeholder="分类" clearable style="width: 150px">
          <el-option label="必需" value="required" />
          <el-option label="风格" value="style" />
          <el-option label="最佳实践" value="best-practice" />
          <el-option label="自定义" value="custom" />
        </el-select>
        <el-button type="primary" :icon="Plus" @click="showCreateDialog = true">
          创建规则
        </el-button>
      </div>

      <el-table :data="filteredRules" style="width: 100%" max-height="500">
        <el-table-column prop="name" label="规则名称" width="200" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="category" label="分类" width="120">
          <template #default="{ row }">
            <el-tag :type="getCategoryType(row.category)" size="small">
              {{ getCategoryLabel(row.category) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="severity" label="级别" width="100">
          <template #default="{ row }">
            <el-tag :type="getSeverityType(row.severity)" size="small">
              {{ getSeverityLabel(row.severity) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="启用" width="80">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              @change="handleToggle(row.id, row.enabled)"
            />
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 创建规则对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="创建自定义规则"
      width="600px"
      append-to-body
    >
      <el-form :model="createForm" label-width="100px" :rules="createRules" ref="createFormRef">
        <el-form-item label="规则 ID" prop="id">
          <el-input v-model="createForm.id" placeholder="例如: my-custom-rule" />
        </el-form-item>
        
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="createForm.name" placeholder="规则显示名称" />
        </el-form-item>
        
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="2"
            placeholder="规则描述"
          />
        </el-form-item>
        
        <el-form-item label="分类" prop="category">
          <el-select v-model="createForm.category" style="width: 100%">
            <el-option label="自定义" value="custom" />
            <el-option label="风格" value="style" />
            <el-option label="最佳实践" value="best-practice" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="严重级别" prop="severity">
          <el-select v-model="createForm.severity" style="width: 100%">
            <el-option label="错误" value="error" />
            <el-option label="警告" value="warning" />
            <el-option label="信息" value="info" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="匹配模式" prop="pattern">
          <el-input
            v-model="createForm.pattern"
            placeholder="正则表达式（例如: ^Name:\\s*\\S+）"
          />
          <div class="form-tip">使用正则表达式匹配规则，留空则使用内置规则</div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateRule" :loading="creating">
          创建
        </el-button>
      </template>
    </el-dialog>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Plus } from '@element-plus/icons-vue'
import { useSpecStore } from '@/stores/useSpecStore'
import { useSpecEditor } from '@/composables/useSpecEditor'
import { createLintRule } from '@/api/modules/spec'
import type { RuleCategory, LintRule } from '@/types/spec'

const visible = defineModel<boolean>()

const store = useSpecStore()
const { loadRules, toggleRule } = useSpecEditor()

const searchText = ref('')
const categoryFilter = ref<RuleCategory | ''>('')

watch(visible, async (val) => {
  if (val) {
    await loadRules()
  }
})

// 创建规则相关
const showCreateDialog = ref(false)
const creating = ref(false)
const createFormRef = ref()
const createForm = ref({
  id: '',
  name: '',
  description: '',
  category: 'custom' as RuleCategory,
  severity: 'warning' as 'error' | 'warning' | 'info',
  pattern: '',
})

const createRules = {
  id: [
    { required: true, message: '请输入规则 ID', trigger: 'blur' },
    { pattern: /^[a-z0-9-]+$/, message: 'ID 只能包含小写字母、数字和短横线', trigger: 'blur' },
  ],
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  description: [{ required: true, message: '请输入规则描述', trigger: 'blur' }],
  category: [{ required: true, message: '请选择分类', trigger: 'change' }],
  severity: [{ required: true, message: '请选择严重级别', trigger: 'change' }],
}

const filteredRules = computed(() => {
  let rules = store.rules

  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    rules = rules.filter(
      (r) =>
        r.name.toLowerCase().includes(search) ||
        r.description.toLowerCase().includes(search)
    )
  }

  if (categoryFilter.value) {
    rules = rules.filter((r) => r.category === categoryFilter.value)
  }

  return rules
})

function getCategoryType(category: RuleCategory) {
  const types: Record<RuleCategory, string> = {
    required: 'danger',
    style: 'warning',
    'best-practice': 'success',
    custom: 'info',
  }
  return types[category]
}

function getCategoryLabel(category: RuleCategory) {
  const labels: Record<RuleCategory, string> = {
    required: '必需',
    style: '风格',
    'best-practice': '最佳实践',
    custom: '自定义',
  }
  return labels[category]
}

function getSeverityType(severity: 'error' | 'warning' | 'info') {
  const types = {
    error: 'danger',
    warning: 'warning',
    info: 'info',
  }
  return types[severity]
}

function getSeverityLabel(severity: 'error' | 'warning' | 'info') {
  const labels = {
    error: '错误',
    warning: '警告',
    info: '信息',
  }
  return labels[severity]
}

async function handleToggle(id: string, enabled: boolean) {
  await toggleRule(id, enabled)
}

async function handleCreateRule() {
  if (!createFormRef.value) return
  
  try {
    await createFormRef.value.validate()
  } catch {
    return
  }

  try {
    creating.value = true
    
    const rule: LintRule = {
      id: createForm.value.id,
      name: createForm.value.name,
      description: createForm.value.description,
      category: createForm.value.category,
      severity: createForm.value.severity,
      enabled: true,
      pattern: createForm.value.pattern,
    }

    await createLintRule(rule)
    ElMessage.success('规则创建成功')
    
    // 重新加载规则列表
    await loadRules()
    
    // 重置表单
    createForm.value = {
      id: '',
      name: '',
      description: '',
      category: 'custom',
      severity: 'warning',
      pattern: '',
    }
    showCreateDialog.value = false
    
  } catch (error: any) {
    ElMessage.error(error.message || '创建规则失败')
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.rule-manager {
  min-height: 400px;
}

.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.form-tip {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}
</style>
