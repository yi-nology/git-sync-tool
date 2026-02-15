import request from '../request'

// 通知渠道管理相关API

export interface NotificationChannel {
  id: number
  name: string
  type: 'email' | 'dingtalk' | 'wechat' | 'webhook'
  config: string
  enabled: boolean
  notify_on_success: boolean
  notify_on_failure: boolean
  created_at: string
  updated_at: string
}

export interface CreateChannelParams {
  name: string
  type: string
  config: string
  enabled: boolean
  notify_on_success: boolean
  notify_on_failure: boolean
}

export interface UpdateChannelParams extends CreateChannelParams {
  id: number
}

// 列出通知渠道
export function listChannels(type?: string) {
  return request.get<unknown, { channels: NotificationChannel[] }>('/notification/channels', {
    params: type ? { type } : undefined
  }).then(res => res?.channels || [])
}

// 获取渠道详情
export function getChannel(id: number) {
  return request.get<unknown, { channel: NotificationChannel }>('/notification/channel', {
    params: { id }
  }).then(res => res?.channel)
}

// 创建渠道
export function createChannel(params: CreateChannelParams) {
  return request.post<unknown, { channel: NotificationChannel }>('/notification/channel/create', params)
}

// 更新渠道
export function updateChannel(params: UpdateChannelParams) {
  return request.post<unknown, { channel: NotificationChannel }>('/notification/channel/update', params)
}

// 删除渠道
export function deleteChannel(id: number) {
  return request.post('/notification/channel/delete', { id })
}

// 测试渠道
export function testChannel(id: number, message?: string) {
  return request.post<unknown, { success: boolean; error?: string }>('/notification/channel/test', {
    id,
    message
  })
}
