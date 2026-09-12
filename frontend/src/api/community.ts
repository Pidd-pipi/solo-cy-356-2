import { del, get, post, put } from '@/utils/request'

export interface CommunityComment {
  id: number
  post_id: number
  user_id: number
  username: string
  content: string
  created_at: string
}

export interface CommunityPost {
  id: number
  user_id: number
  username: string
  nickname: string
  title: string
  content: string
  post_type: string
  status: string
  like_count: number
  comment_count: number
  created_at: string
  comments?: CommunityComment[]
}

export function listPosts(params?: Record<string, any>): Promise<{ list: CommunityPost[]; total: number; page: number; page_size: number }> {
  return get('/community-posts', { params })
}

export function getPost(id: number): Promise<CommunityPost> {
  return get(`/community-posts/${id}`)
}

export function createPost(payload: { title: string; content: string; post_type: string }): Promise<CommunityPost> {
  return post('/community-posts', payload)
}

export function updatePost(id: number, payload: Partial<{ title: string; content: string; post_type: string }>): Promise<CommunityPost> {
  return put(`/community-posts/${id}`, payload)
}

export function removePost(id: number): Promise<{ removed: boolean }> {
  return del(`/community-posts/${id}`)
}

export function likePost(id: number): Promise<{ liked: boolean }> {
  return post(`/community-posts/${id}/like`)
}

export function commentPost(id: number, content: string): Promise<CommunityComment> {
  return post(`/community-posts/${id}/comments`, { content })
}
