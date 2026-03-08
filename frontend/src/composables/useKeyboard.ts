import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'

/**
 * 全局快捷键支持
 */
export function useKeyboard() {
  const router = useRouter()

  const handleKeydown = (event: KeyboardEvent) => {
    // Cmd/Ctrl + K: 全局搜索（未来实现）
    if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
      event.preventDefault()
      console.log('[Keyboard] Global search triggered')
      // TODO: 打开全局搜索弹窗
    }

    // Cmd/Ctrl + N: 新建（当前页面相关的新建操作）
    if ((event.metaKey || event.ctrlKey) && event.key === 'n') {
      event.preventDefault()
      console.log('[Keyboard] New item triggered')
      // 根据当前路由决定新建什么
      if (window.location.pathname.includes('/repos')) {
        router.push('/repos/register')
      }
    }

    // Cmd/Ctrl + R: 刷新当前页面
    if ((event.metaKey || event.ctrlKey) && event.key === 'r') {
      // 阻止浏览器默认刷新
      // event.preventDefault()
      // 让浏览器默认行为生效
    }

    // Esc: 关闭弹窗（由 Element Plus 自动处理）
  }

  onMounted(() => {
    window.addEventListener('keydown', handleKeydown)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown)
  })

  return {
    // 可以返回一些手动触发的快捷键方法
  }
}
