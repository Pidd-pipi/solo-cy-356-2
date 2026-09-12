<template>
  <div class="auth-wrap">
    <el-card class="auth-card">
      <h2 class="auth-title">注册农友账号</h2>
      <el-form :model="form" label-width="0" @keyup.enter="onSubmit">
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名（3-32 位）" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="密码（至少 6 位）" size="large" show-password />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.nickname" placeholder="昵称" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.email" placeholder="邮箱（选填）" size="large" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" style="width: 100%" :loading="loading" @click="onSubmit">注 册</el-button>
        </el-form-item>
      </el-form>
      <div class="auth-footer">
        已有账号？<router-link to="/login">去登录</router-link>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const form = reactive({ username: '', password: '', nickname: '', email: '', phone: '' })

async function onSubmit() {
  if (!form.username || form.password.length < 6) {
    ElMessage.warning('用户名必填，密码至少 6 位')
    return
  }
  loading.value = true
  try {
    await authStore.register({ ...form })
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-wrap { height: 100%; display: flex; align-items: center; justify-content: center; background: linear-gradient(135deg, #a8e063, #56ab2f); }
.auth-card { width: 420px; padding: 8px 12px; border-radius: 12px; }
.auth-title { text-align: center; margin: 0 0 24px; }
.auth-footer { text-align: center; color: #909399; font-size: 13px; }
</style>
