import { get, post } from '@/utils/request'

export interface PlantingPlan {
  id: number
  plot_id: number
  plot_code: string
  plot_name: string
  user_id: number
  username: string
  crop_name: string
  crop_type: string
  season: string
  status: string
  plant_date: string | null
  expected_harvest_date: string | null
  notes: string
  created_at: string
}

export interface CreatePlanPayload {
  plot_id: number
  crop_name: string
  crop_type: string
  season: string
  plant_date?: string
  notes?: string
}

export function listPlans(params?: Record<string, any>): Promise<{ list: PlantingPlan[]; total: number; page: number; page_size: number }> {
  return get('/planting-plans', { params })
}

export function getPlan(id: number): Promise<PlantingPlan> {
  return get(`/planting-plans/${id}`)
}

export function createPlan(payload: CreatePlanPayload): Promise<PlantingPlan> {
  return post('/planting-plans', payload)
}

export function updatePlanStatus(id: number, status: string): Promise<PlantingPlan> {
  return post(`/planting-plans/${id}/status`, { status })
}

export function getRecommendations(season: string): Promise<Array<{ season: string; crops: string[]; harvest_in_days: number }>> {
  return get('/crops/recommendations', { params: { season } })
}

export function getHarvestReminders(): Promise<PlantingPlan[]> {
  return get('/reminders/harvest')
}
