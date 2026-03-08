import { ref, reactive, type Ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { useNotification } from './useNotification'

export interface FormOptions<T> {
  initialValues: T
  rules?: FormRules
  validateBeforeSubmit?: boolean
}

export interface UseFormReturn {
  form: any
  loading: Ref<boolean>
  formRef: Ref<FormInstance | undefined>
  rules: FormRules
  validate: () => Promise<boolean>
  reset: () => void
  submit: (submitFn: (form: any) => Promise<void>) => Promise<void>
  setFieldValue: <K extends string>(field: K, value: any) => void
  getFieldValue: <K extends string>(field: K) => any
}

/**
 * 通用表单处理 composable
 * @param options 表单配置选项
 * @returns 表单处理相关方法和状态
 */
export function useForm<T extends Record<string, any>>(
  options: FormOptions<T>
): UseFormReturn {
  const { showError, showSuccess } = useNotification()
  
  const formRef = ref<FormInstance>()
  const loading = ref(false)
  const form = reactive<T>({ ...options.initialValues })
  const rules = options.rules || {}

  /**
   * 验证表单
   * @returns 是否验证通过
   */
  const validate = async (): Promise<boolean> => {
    if (!formRef.value) return false
    
    try {
      await formRef.value.validate()
      return true
    } catch (error) {
      return false
    }
  }

  /**
   * 重置表单
   */
  const reset = (): void => {
    if (formRef.value) {
      formRef.value.resetFields()
    }
    Object.assign(form, options.initialValues)
  }

  /**
   * 提交表单
   * @param submitFn 提交函数
   */
  const submit = async (submitFn: (form: any) => Promise<void>): Promise<void> => {
    // 验证表单
    if (options.validateBeforeSubmit !== false) {
      const isValid = await validate()
      if (!isValid) {
        showError('表单验证失败，请检查输入')
        return
      }
    }

    loading.value = true
    try {
      await submitFn(form)
      showSuccess('操作成功')
    } catch (error: any) {
      showError('操作失败', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  /**
   * 设置字段值
   * @param field 字段名
   * @param value 字段值
   */
  const setFieldValue = <K extends string>(field: K, value: any): void => {
    form[field] = value
  }

  /**
   * 获取字段值
   * @param field 字段名
   * @returns 字段值
   */
  const getFieldValue = <K extends string>(field: K): any => {
    return form[field]
  }

  return {
    form,
    loading,
    formRef,
    rules,
    validate,
    reset,
    submit,
    setFieldValue,
    getFieldValue
  }
}
