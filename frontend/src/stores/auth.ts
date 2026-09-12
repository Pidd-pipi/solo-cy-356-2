import { defineStore } from 'pinia'
import { login as apiLogin, register as apiRegister, fetchMe } from '@/api/auth'
import type { UserInfo } from '@/api/auth'

interface AuthState {
  token: string
  user: UserInfo | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: localStorage.getItem('cg_token') || '',
    user: JSON.parse(localStorage.getItem('cg_user') || 'null')
  }),
  getters: {
    isLoggedIn: (s) => !!s.token
  },
  actions: {
    async login(username: string, password: string) {
      const data = await apiLogin(username, password)
      this.token = data.token
      this.user = data.user
      localStorage.setItem('cg_token', data.token)
      localStorage.setItem('cg_user', JSON.stringify(data.user))
    },
    async register(payload: { username: string; password: string; nickname?: string; email?: string; phone?: string }) {
      await apiRegister(payload)
    },
    async fetchMe() {
      const me = await fetchMe()
      this.user = me
      localStorage.setItem('cg_user', JSON.stringify(me))
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('cg_token')
      localStorage.removeItem('cg_user')
    }
  }
})
