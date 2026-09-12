import { get } from '@/utils/request'

export interface DashboardStats {
  users_by_role: Record<string, number>
  plots_by_status: Record<string, number>
  plans_by_status: Record<string, number>
  posts_by_type: Record<string, number>
  total_diaries: number
  total_harvests: number
}

export function getDashboardStats(): Promise<DashboardStats> {
  return get('/dashboard/stats')
}
