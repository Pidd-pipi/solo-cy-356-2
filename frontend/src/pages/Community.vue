<template>
  <div class="page-card">
    <div style="display: flex; justify-content: space-between; align-items: center">
      <h3 class="page-title">农友社区</h3>
      <el-button type="primary" @click="openCreate">+ 发布帖子</el-button>
    </div>

    <el-card shadow="never" style="margin-bottom: 16px">
      <template #header>📡 实时动态（WebSocket）</template>
      <div class="ws-status">
        <el-tag :type="wsConnected ? 'success' : 'info'" size="small">{{ wsConnected ? '已连接' : '未连接' }}</el-tag>
        <span v-if="wsMsg" class="muted" style="margin-left: 8px">{{ wsMsg }}</span>
      </div>
    </el-card>

    <el-radio-group v-model="typeFilter" style="margin-bottom: 16px" @change="fetch">
      <el-radio-button value="">全部</el-radio-button>
      <el-radio-button v-for="(label, key) in PostTypeText" :key="key" :value="key">{{ label }}</el-radio-button>
    </el-radio-group>

    <el-row :gutter="16">
      <el-col v-for="p in store.posts" :key="p.id" :span="12" style="margin-bottom: 16px">
        <el-card shadow="hover">
          <div style="display: flex; justify-content: space-between">
            <strong>{{ p.title }}</strong>
            <StatusBadge :value="p.post_type" :meta-map="postTypeMeta" />
          </div>
          <p class="muted">{{ p.nickname || p.username }} · {{ formatDateTime(p.created_at) }}</p>
          <p style="color: #606266">{{ p.content }}</p>
          <div style="display: flex; gap: 12px; align-items: center">
            <el-button size="small" @click="onLike(p)">👍 {{ p.like_count }}</el-button>
            <el-button size="small" @click="toggleComments(p)">💬 {{ p.comment_count }}</el-button>
            <el-button v-if="canRemove(p)" size="small" type="danger" plain @click="onRemove(p)">删除</el-button>
          </div>
          <div v-if="expandedId === p.id" style="margin-top: 12px">
            <div v-for="c in p.comments" :key="c.id" class="comment-item">
              <b>{{ c.username }}</b>：{{ c.content }}
            </div>
            <div style="display: flex; gap: 8px; margin-top: 8px">
              <el-input v-model="commentText[p.id]" placeholder="写评论..." size="small" />
              <el-button size="small" type="primary" @click="onComment(p)">发送</el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    <EmptyState v-if="!store.posts.length && !store.loading" description="暂无帖子，来发第一帖吧" />

    <el-dialog v-model="createVisible" title="发布帖子" width="520px">
      <el-form :model="createForm" label-width="90px">
        <el-form-item label="帖子类型">
          <el-select v-model="createForm.post_type">
            <el-option v-for="(t, k) in PostTypeText" :key="k" :label="t" :value="k" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题"><el-input v-model="createForm.title" /></el-form-item>
        <el-form-item label="内容"><el-input v-model="createForm.content" type="textarea" :rows="5" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">发布</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useCommunityStore } from '@/stores/community'
import { getPost, type CommunityPost } from '@/api/community'
import { useAuth } from '@/hooks/useAuth'
import StatusBadge from '@/components/StatusBadge.vue'
import EmptyState from '@/components/EmptyState.vue'
import { PostTypeText } from '@/constants'
import { formatDateTime } from '@/utils/format'

const store = useCommunityStore()
const { user, role } = useAuth()
const typeFilter = ref('')
const createVisible = ref(false)
const creating = ref(false)
const expandedId = ref<number | null>(null)
const commentText = reactive<Record<number, string>>({})
const createForm = reactive({ post_type: 'experience', title: '', content: '' })

const postTypeMeta = Object.fromEntries(Object.entries(PostTypeText).map(([k, v]) => [k, { label: v, type: 'primary' }]))

const wsConnected = ref(false)
const wsMsg = ref('')
let ws: WebSocket | null = null

async function fetch() {
  await store.fetchPosts({ page: 1, page_size: 20, post_type: typeFilter.value })
}

function canRemove(p: CommunityPost) {
  return role.value === 'admin' || p.user_id === user.value?.id
}

async function onLike(p: CommunityPost) {
  await store.like(p.id)
}

async function toggleComments(p: CommunityPost) {
  if (expandedId.value === p.id) {
    expandedId.value = null
    return
  }
  expandedId.value = p.id
  if (!p.comments || p.comments.length === 0) {
    const detail = await getPost(p.id)
    p.comments = detail.comments || []
  }
}

async function onComment(p: CommunityPost) {
  const text = commentText[p.id]
  if (!text) return
  await store.comment(p.id, text)
  commentText[p.id] = ''
  const detail = await getPost(p.id)
  p.comments = detail.comments || []
  p.comment_count = detail.comment_count
}

async function onRemove(p: CommunityPost) {
  await store.remove(p.id)
  ElMessage.success('帖子已删除')
}

function openCreate() {
  createVisible.value = true
}

async function submitCreate() {
  if (!createForm.title || !createForm.content) {
    ElMessage.warning('请填写标题与内容')
    return
  }
  creating.value = true
  try {
    await store.create({ ...createForm })
    ElMessage.success('发布成功')
    createVisible.value = false
  } finally {
    creating.value = false
  }
}

function connectWS() {
  try {
    const token = localStorage.getItem('cg_token')
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    ws = new WebSocket(`${proto}://${location.host}/api/v1/ws/community?token=${encodeURIComponent(token || '')}`)
    ws.onopen = () => {
      wsConnected.value = true
    }
    ws.onmessage = (ev) => {
      wsMsg.value = `收到实时消息：${ev.data}`
    }
    ws.onclose = () => {
      wsConnected.value = false
    }
    ws.onerror = () => {
      wsConnected.value = false
    }
  } catch {
    wsConnected.value = false
  }
}

onMounted(() => {
  fetch()
  connectWS()
})
onBeforeUnmount(() => {
  if (ws) {
    ws.close()
  }
})
</script>

<style scoped>
.comment-item { background: #f5f7fa; border-radius: 6px; padding: 6px 10px; margin-bottom: 6px; font-size: 13px; color: #606266; }
.ws-status { display: flex; align-items: center; }
</style>
