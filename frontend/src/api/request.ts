import axios, { type AxiosInstance, type AxiosResponse, type InternalAxiosRequestConfig, type AxiosError } from 'axios'
import type { ApiResponse } from '@/types/common'
import { useNotification } from '@/composables/useNotification'

// 创建 axios 实例
const service: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求取消控制器
const cancelTokens = new Map<string, AbortController>()

// 生成请求标识
const generateRequestKey = (config: InternalAxiosRequestConfig): string => {
  const { method, url, params, data } = config
  return `${method}:${url}:${JSON.stringify(params || {})}:${JSON.stringify(data || {})}`
}

// 请求拦截器
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // 取消之前的相同请求
    const requestKey = generateRequestKey(config)
    const existingController = cancelTokens.get(requestKey)
    if (existingController) {
      existingController.abort()
    }

    // 创建新的取消控制器
    const controller = new AbortController()
    config.signal = controller.signal
    cancelTokens.set(requestKey, controller)

    // 可以在这里添加认证信息等
    // const token = localStorage.getItem('token')
    // if (token) {
    //   config.headers.Authorization = `Bearer ${token}`
    // }

    return config
  },
  (error: AxiosError) => {
    console.error('Request error:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
service.interceptors.response.use(
  (response: AxiosResponse) => {
    const { showError } = useNotification()
    
    // 清理取消令牌
    const requestKey = generateRequestKey(response.config)
    cancelTokens.delete(requestKey)

    // 处理非 JSON 响应
    const contentType = response.headers['content-type']
    if (contentType && !contentType.includes('application/json')) {
      return response.data
    }

    // 处理 API 响应
    const res = response.data as ApiResponse
    if (res.code !== 0) {
      showError(res.msg || '请求失败')
      return Promise.reject(new Error(res.msg || '请求失败'))
    }

    return response.data ? res.data : response
  },
  (error: AxiosError) => {
    const { showError } = useNotification()
    
    // 清理取消令牌
    if (error.config) {
      const requestKey = generateRequestKey(error.config)
      cancelTokens.delete(requestKey)
    }

    // 处理取消请求
    if (axios.isCancel(error)) {
      console.log('Request cancelled:', error.message)
      return Promise.reject(error)
    }

    // 处理网络错误
    let message = '网络错误'
    if (error.response) {
      // 服务器返回错误
      const status = error.response.status
      switch (status) {
        case 400:
          message = '请求参数错误'
          break
        case 401:
          message = '未授权，请重新登录'
          // 可以在这里处理登录跳转
          break
        case 403:
          message = '拒绝访问'
          break
        case 404:
          message = '请求的资源不存在'
          break
        case 500:
          message = '服务器内部错误'
          break
        default:
          message = (error.response.data as any)?.message || error.message || '请求失败'
      }
    } else if (error.request) {
      // 请求已发出但没有收到响应
      message = '服务器无响应'
    } else {
      // 请求配置错误
      message = error.message || '请求失败'
    }

    showError(message)
    console.error('Response error:', error)
    return Promise.reject(error)
  }
)

// 导出请求方法
export default service

// 导出工具方法
export const api = {
  get: <T = any>(url: string, params?: any) => service.get<T>(url, { params }),
  post: <T = any>(url: string, data?: any) => service.post<T>(url, data),
  put: <T = any>(url: string, data?: any) => service.put<T>(url, data),
  delete: <T = any>(url: string, params?: any) => service.delete<T>(url, { params }),
  patch: <T = any>(url: string, data?: any) => service.patch<T>(url, data),
}

// 取消所有请求
export const cancelAllRequests = () => {
  cancelTokens.forEach((controller) => {
    controller.abort()
  })
  cancelTokens.clear()
}

