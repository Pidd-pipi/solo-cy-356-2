import { defineStore } from 'pinia'
import { commentPost, createPost, likePost, listPosts, removePost, type CommunityPost } from '@/api/community'

interface CommunityState {
  posts: CommunityPost[]
  total: number
  loading: boolean
}

export const useCommunityStore = defineStore('community', {
  state: (): CommunityState => ({ posts: [], total: 0, loading: false }),
  actions: {
    async fetchPosts(params?: Record<string, any>) {
      this.loading = true
      try {
        const data = await listPosts(params)
        this.posts = data.list
        this.total = data.total
      } finally {
        this.loading = false
      }
    },
    async create(payload: Parameters<typeof createPost>[0]) {
      await createPost(payload)
      await this.fetchPosts()
    },
    async like(id: number) {
      await likePost(id)
      await this.fetchPosts()
    },
    async comment(id: number, content: string) {
      await commentPost(id, content)
      await this.fetchPosts()
    },
    async remove(id: number) {
      await removePost(id)
      await this.fetchPosts()
    }
  }
})
