import request from '../request'
import type { AuditLogDTO } from '@/types/stats'

export function getAuditLogs(params?: { page?: number; page_size?: number }) {
  return request.get<unknown, { items: AuditLogDTO[]; total: number; page: number; size: number }>('/audit/logs', {
    params,
  })
}

export function getAuditLog(id: number) {
  return request.get<unknown, AuditLogDTO>('/audit/log', { params: { id } })
}
