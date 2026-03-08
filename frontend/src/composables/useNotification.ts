import { ElMessage, ElNotification } from 'element-plus'

/**
 * 统一的通知管理 composable
 * 提供统一的消息提示、错误处理
 */
export function useNotification() {
  /**
   * 成功消息
   */
  const showSuccess = (message: string, duration = 3000) => {
    ElMessage.success({
      message,
      duration,
      showClose: true,
    })
  }

  /**
   * 错误消息
   */
  const showError = (message: string, error?: Error | unknown, duration = 5000) => {
    ElMessage.error({
      message,
      duration,
      showClose: true,
    })

    // 记录错误到控制台
    if (error) {
      console.error('[Error]', message, error)
    }
  }

  /**
   * 警告消息
   */
  const showWarning = (message: string, duration = 4000) => {
    ElMessage.warning({
      message,
      duration,
      showClose: true,
    })
  }

  /**
   * 信息消息
   */
  const showInfo = (message: string, duration = 3000) => {
    ElMessage.info({
      message,
      duration,
      showClose: true,
    })
  }

  /**
   * 成功通知（带图标）
   */
  const notifySuccess = (title: string, message: string) => {
    ElNotification.success({
      title,
      message,
      duration: 4500,
    })
  }

  /**
   * 错误通知（带图标）
   */
  const notifyError = (title: string, message: string, error?: Error | unknown) => {
    ElNotification.error({
      title,
      message,
      duration: 6000,
    })

    if (error) {
      console.error('[Error]', title, message, error)
    }
  }

  /**
   * 警告通知（带图标）
   */
  const notifyWarning = (title: string, message: string) => {
    ElNotification.warning({
      title,
      message,
      duration: 5000,
    })
  }

  /**
   * 信息通知（带图标）
   */
  const notifyInfo = (title: string, message: string) => {
    ElNotification.info({
      title,
      message,
      duration: 4000,
    })
  }

  return {
    // 简单消息
    showSuccess,
    showError,
    showWarning,
    showInfo,

    // 带图标的通知
    notifySuccess,
    notifyError,
    notifyWarning,
    notifyInfo,
  }
}
