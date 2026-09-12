import { get } from '@/utils/request'

export interface AuditLog {
  id: number
  user_id: number
  username: string
  role: string
  action: string
  resource_type: string
  resource_id: string
  detail: string
  ip: string
  request_id: string
  created_at: string
}

export function listAuditLogs(params?: Record<string, any>): Promise<{ list: AuditLog[]; total: number; page: number; page_size: number }> {
  return get('/audit-logs', { params })
}
