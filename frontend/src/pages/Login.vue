<template>
  <div class="auth-wrap">
    <el-card class="auth-card">
      <h2 class="auth-title">🌱 城市共享菜园管理平台</h2>
      <p class="auth-sub">认养一块地，做个都市农夫</p>
      <el-form :model="form" label-width="0" @keyup.enter="onSubmit">
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名（如 admin / farmer / citizen）" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="密码（如 admin123）" size="large" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" style="width: 100%" :loading="loading" @click="onSubmit">登 录</el-button>
        </el-form-item>
      </el-form>
      <div class="auth-footer">
        还没有账号？<router-link to="/register">立即注册</router-link>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

async function onSubmit() {
  if (!form.username || !form.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    await authStore.login(form.username, form.password)
    ElMessage.success('登录成功')
    router.push((route.query.redirect as string) || '/')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-wrap { height: 100%; display: flex; align-items: center; justify-content: center; background: linear-gradient(135deg, #a8e063, #56ab2f); }
.auth-card { width: 420px; padding: 8px 12px; border-radius: 12px; }
.auth-title { text-align: center; margin: 0 0 4px; }
.auth-sub { text-align: center; color: #909399; margin: 0 0 24px; }
.auth-footer { text-align: center; color: #909399; font-size: 13px; }
</style>
