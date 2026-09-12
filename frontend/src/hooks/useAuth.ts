import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'

// 共享认证 hook：暴露 token/用户/角色判断
export function useAuth() {
  const store = useAuthStore()
  const isLoggedIn = computed(() => store.isLoggedIn)
  const user = computed(() => store.user)
  const role = computed(() => store.user?.role ?? '')
  const isAdmin = computed(() => role.value === 'admin')
  const isFarmer = computed(() => role.value === 'farmer')
  return { isLoggedIn, user, role, isAdmin, isFarmer, logout: store.logout, fetchMe: store.fetchMe }
}
