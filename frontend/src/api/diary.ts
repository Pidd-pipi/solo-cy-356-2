import { del, get, post, put } from '@/utils/request'

export interface DiaryComment {
  id: number
  diary_id: number
  user_id: number
  username: string
  content: string
  created_at: string
}

export interface DiaryEntry {
  id: number
  plan_id: number
  plan_code: string
  user_id: number
  username: string
  nickname: string
  action_type: string
  title: string
  content: string
  image_url: string
  like_count: number
  created_at: string
  comments?: DiaryComment[]
}

export function listDiaries(params?: Record<string, any>): Promise<{ list: DiaryEntry[]; total: number; page: number; page_size: number }> {
  return get('/diaries', { params })
}

export function getDiary(id: number): Promise<DiaryEntry> {
  return get(`/diaries/${id}`)
}

export function createDiary(payload: { plan_id: number; action_type: string; title: string; content: string; image_url?: string }): Promise<DiaryEntry> {
  return post('/diaries', payload)
}

export function updateDiary(id: number, payload: Partial<{ action_type: string; title: string; content: string; image_url: string }>): Promise<DiaryEntry> {
  return put(`/diaries/${id}`, payload)
}

export function deleteDiary(id: number): Promise<{ deleted: boolean }> {
  return del(`/diaries/${id}`)
}

export function likeDiary(id: number): Promise<{ liked: boolean }> {
  return post(`/diaries/${id}/like`)
}

export function commentDiary(id: number, content: string): Promise<DiaryComment> {
  return post(`/diaries/${id}/comments`, { content })
}
