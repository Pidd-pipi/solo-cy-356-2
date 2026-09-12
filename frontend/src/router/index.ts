import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('@/pages/Login.vue'), meta: { public: true } },
    { path: '/register', name: 'register', component: () => import('@/pages/Register.vue'), meta: { public: true } },
    {
      path: '/',
      component: MainLayout,
      children: [
        { path: '', name: 'dashboard', component: () => import('@/pages/Dashboard.vue') },
        { path: 'plots', name: 'plots', component: () => import('@/pages/PlotMap.vue') },
        { path: 'plans', name: 'plans', component: () => import('@/pages/PlantingPlan.vue') },
        { path: 'diaries', name: 'diaries', component: () => import('@/pages/Diary.vue') },
        { path: 'harvests', name: 'harvests', component: () => import('@/pages/Harvest.vue') },
        { path: 'community', name: 'community', component: () => import('@/pages/Community.vue') },
        { path: 'audit', name: 'audit', component: () => import('@/pages/Audit.vue'), meta: { admin: true } }
      ]
    }
  ]
})

// 路由守卫：未登录跳转登录页；已登录访问登录/注册页跳回首页
router.beforeEach((to) => {
  const token = localStorage.getItem('cg_token')
  if (!to.meta.public && !token) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.meta.public && token) {
    return { path: '/' }
  }
  return true
})

export default router
