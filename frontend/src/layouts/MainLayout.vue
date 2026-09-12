<template>
  <el-container class="layout">
    <el-aside width="220px" class="aside">
      <div class="logo">🌱 城市共享菜园</div>
      <el-menu :default-active="activeMenu" router background-color="#001529" text-color="#a6adb4" active-text-color="#fff">
        <el-menu-item index="/"><el-icon><Odometer /></el-icon><span>平台概览</span></el-menu-item>
        <el-menu-item index="/plots"><el-icon><MapLocation /></el-icon><span>地块认养</span></el-menu-item>
        <el-menu-item index="/plans"><el-icon><Calendar /></el-icon><span>种植计划</span></el-menu-item>
        <el-menu-item index="/diaries"><el-icon><Notebook /></el-icon><span>种植日记</span></el-menu-item>
        <el-menu-item index="/harvests"><el-icon><Basket /></el-icon><span>收成记录</span></el-menu-item>
        <el-menu-item index="/community"><el-icon><ChatDotRound /></el-icon><span>农友社区</span></el-menu-item>
        <el-menu-item v-if="isAdmin" index="/audit"><el-icon><Document /></el-icon><span>审计日志</span></el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="header-title">城市共享菜园管理平台</div>
        <div class="header-user">
          <el-tag size="small" :type="role === 'admin' ? 'danger' : role === 'farmer' ? 'warning' : 'success'">{{ RoleText[role] || role }}</el-tag>
          <el-dropdown @command="onCommand">
            <span class="username">{{ user?.nickname || user?.username }} <el-icon><ArrowDown /></el-icon></span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from '@/hooks/useAuth'
import { RoleText } from '@/constants'

const route = useRoute()
const router = useRouter()
const { user, role, isAdmin, logout, fetchMe } = useAuth()

const activeMenu = computed(() => (route.path === '/' ? '/' : route.path))

onMounted(() => {
  fetchMe().catch(() => undefined)
})

function onCommand(cmd: string) {
  if (cmd === 'logout') {
    logout()
    router.push('/login')
  }
}
</script>

<style scoped>
.layout { height: 100%; }
.aside { background: #001529; }
.logo { height: 60px; line-height: 60px; color: #fff; font-size: 18px; font-weight: 700; text-align: center; background: #002140; }
.header { background: #fff; display: flex; align-items: center; justify-content: space-between; box-shadow: 0 1px 4px rgba(0,0,0,0.08); }
.header-title { font-size: 16px; font-weight: 600; color: #303133; }
.header-user { display: flex; align-items: center; gap: 12px; }
.username { cursor: pointer; color: #303133; display: flex; align-items: center; gap: 4px; }
.main { padding: 0; background: #f5f7fa; }
:deep(.el-menu) { border-right: none; }
</style>
