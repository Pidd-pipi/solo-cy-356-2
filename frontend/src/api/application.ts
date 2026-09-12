import { get, post } from '@/utils/request'
import type { Plot } from './plot'
import type { UserInfo } from './auth'

// 认养申请（与后端 AdoptionApplication 对应）
export interface AdoptionApplication {
  id: number
  plot_id: number
  plot: Plot | null
  user_id: number
  user: UserInfo | null
  status: string
  reason: string
  review_note: string
  reviewed_by: number | null
  reviewed_at: string
  created_at: string
}

export interface ApplicationPage {
  list: AdoptionApplication[]
  total: number
  page: number
  page_size: number
}

export function applyForPlot(payload: { plot_id: number; reason?: string }): Promise<AdoptionApplication> {
  return post('/applications', payload)
}

export function listMyApplications(params?: Record<string, any>): Promise<ApplicationPage> {
  return get('/applications/mine', { params })
}

export function withdrawApplication(id: number): Promise<AdoptionApplication> {
  return post(`/applications/${id}/withdraw`)
}

export function listApplications(params?: Record<string, any>): Promise<ApplicationPage> {
  return get('/applications', { params })
}

export function reviewApplication(id: number, payload: { action: 'approve' | 'reject'; note?: string }): Promise<AdoptionApplication> {
  return post(`/applications/${id}/review`, payload)
}
