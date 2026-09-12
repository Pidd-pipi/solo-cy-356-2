import { get, post } from '@/utils/request'

export interface UserInfo {
  id: number
  username: string
  nickname: string
  email: string
  phone: string
  role: string
  status: string
  created_at: string
}

export interface LoginResult {
  token: string
  user: UserInfo
  expires_in: number
}

export function login(username: string, password: string): Promise<LoginResult> {
  return post('/auth/login', { username, password })
}

export function register(payload: { username: string; password: string; nickname?: string; email?: string; phone?: string }): Promise<UserInfo> {
  return post('/auth/register', payload)
}

export function fetchMe(): Promise<UserInfo> {
  return get('/me')
}
