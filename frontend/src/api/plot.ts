import { get, post } from '@/utils/request'
import type { UserInfo } from './auth'

export interface Plot {
  id: number
  name: string
  code: string
  area: number
  soil_type: string
  sunlight: string
  latitude: number
  longitude: number
  status: string
  adopter_id: number | null
  adopter: UserInfo | null
  description: string
  created_at: string
}

export interface PlotPayload {
  name: string
  code: string
  area: number
  soil_type: string
  sunlight: string
  latitude: number
  longitude: number
  description?: string
}

export function listPlots(params?: Record<string, any>): Promise<{ list: Plot[]; total: number; page: number; page_size: number }> {
  return get('/plots', { params })
}

export function getPlot(id: number): Promise<Plot> {
  return get(`/plots/${id}`)
}

export function createPlot(payload: PlotPayload): Promise<Plot> {
  return post('/plots', payload)
}

export function adoptPlot(id: number): Promise<Plot> {
  return post(`/plots/${id}/adopt`)
}

export function releasePlot(id: number): Promise<Plot> {
  return post(`/plots/${id}/release`)
}
