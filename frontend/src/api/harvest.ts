import { del, get, post, put } from '@/utils/request'

export interface HarvestRecord {
  id: number
  plan_id: number
  plan_code: string
  crop_name: string
  user_id: number
  username: string
  harvest_date: string
  weight_kg: number
  quality: string
  notes: string
  created_at: string
}

export interface CreateHarvestPayload {
  plan_id: number
  crop_name: string
  harvest_date: string
  weight_kg: number
  quality: string
  notes?: string
}

export interface AnnualStats {
  year: number
  total_weight_kg: number
  harvest_count: number
  by_crop_type: Record<string, number>
  by_quality: Record<string, number>
}

export function listHarvests(params?: Record<string, any>): Promise<{ list: HarvestRecord[]; total: number; page: number; page_size: number }> {
  return get('/harvests', { params })
}

export function createHarvest(payload: CreateHarvestPayload): Promise<HarvestRecord> {
  return post('/harvests', payload)
}

export function updateHarvest(id: number, payload: Partial<CreateHarvestPayload>): Promise<HarvestRecord> {
  return put(`/harvests/${id}`, payload)
}

export function deleteHarvest(id: number): Promise<{ deleted: boolean }> {
  return del(`/harvests/${id}`)
}

export function getAnnualStats(year: number): Promise<AnnualStats> {
  return get('/stats/annual', { params: { year } })
}
