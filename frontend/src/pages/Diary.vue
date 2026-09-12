<template>
  <div class="page-card">
    <div style="display: flex; justify-content: space-between; align-items: center">
      <h3 class="page-title">种植日记</h3>
      <el-button type="primary" @click="openCreate">+ 写日记</el-button>
    </div>

    <el-row :gutter="16">
      <el-col v-for="d in store.diaries" :key="d.id" :span="12" style="margin-bottom: 16px">
        <el-card shadow="hover">
          <div style="display: flex; justify-content: space-between">
            <strong>{{ d.title }}</strong>
            <StatusBadge :value="d.action_type" :meta-map="actionMeta" />
          </div>
          <p class="muted">{{ d.nickname || d.username }} · {{ formatDateTime(d.created_at) }} · 计划：{{ d.plan_code }}</p>
          <p style="color: #606266">{{ d.content }}</p>
          <div style="display: flex; gap: 12px; align-items: center">
            <el-button size="small"  @click="onLike(d)">👍 {{ d.like_count }}</el-button>
            <el-button size="small" @click="toggleComments(d)">💬 {{ d.comments?.length || 0 }}</el-button>
          </div>
          <div v-if="expandedId === d.id" style="margin-top: 12px">
            <div v-for="c in d.comments" :key="c.id" class="comment-item">
              <b>{{ c.username }}</b>：{{ c.content }}
            </div>
            <div style="display: flex; gap: 8px; margin-top: 8px">
              <el-input v-model="commentText[d.id]" placeholder="写评论..." size="small" />
              <el-button size="small" type="primary" @click="onComment(d)">发送</el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    <EmptyState v-if="!store.diaries.length && !store.loading" description="还没有种植日记，记录下第一次播种吧" />

    <el-dialog v-model="createVisible" title="发布种植日记" width="520px">
      <el-form :model="createForm" label-width="90px">
        <el-form-item label="关联计划">
          <el-select v-model="createForm.plan_id" placeholder="选择种植计划">
            <el-option v-for="p in myPlans" :key="p.id" :label="`${p.crop_name}（${p.plot_name}）`" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="动作类型">
          <el-select v-model="createForm.action_type">
            <el-option v-for="(t, k) in DiaryActionText" :key="k" :label="t" :value="k" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题"><el-input v-model="createForm.title" /></el-form-item>
        <el-form-item label="内容"><el-input v-model="createForm.content" type="textarea" :rows="4" /></el-form-item>
        <el-form-item label="图片URL"><el-input v-model="createForm.image_url" placeholder="选填" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">发布</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useDiaryStore } from '@/stores/diary'
import { getDiary, type DiaryEntry } from '@/api/diary'
import { listPlans } from '@/api/plantingPlan'
import StatusBadge from '@/components/StatusBadge.vue'
import EmptyState from '@/components/EmptyState.vue'
import { DiaryActionText } from '@/constants'
import { formatDateTime } from '@/utils/format'

const store = useDiaryStore()
const createVisible = ref(false)
const creating = ref(false)
const expandedId = ref<number | null>(null)
const commentText = reactive<Record<number, string>>({})
const myPlans = ref<Array<{ id: number; crop_name: string; plot_name: string }>>([])
const createForm = reactive({ plan_id: undefined as number | undefined, action_type: 'sowing', title: '', content: '', image_url: '' })

const actionMeta = Object.fromEntries(Object.entries(DiaryActionText).map(([k, v]) => [k, { label: v, type: 'primary' }]))

async function fetch() {
  await store.fetchDiaries({ page: 1, page_size: 20 })
}

async function onLike(d: DiaryEntry) {
  await store.like(d.id)
}

async function toggleComments(d: DiaryEntry) {
  if (expandedId.value === d.id) {
    expandedId.value = null
    return
  }
  expandedId.value = d.id
  if (!d.comments || d.comments.length === 0) {
    const detail = await getDiary(d.id)
    d.comments = detail.comments || []
  }
}

async function onComment(d: DiaryEntry) {
  const text = commentText[d.id]
  if (!text) return
  await store.comment(d.id, text)
  commentText[d.id] = ''
  const detail = await getDiary(d.id)
  d.comments = detail.comments || []
}

async function openCreate() {
  createVisible.value = true
  try {
    const data = await listPlans({ page: 1, page_size: 100 })
    myPlans.value = data.list.map((p) => ({ id: p.id, crop_name: p.crop_name, plot_name: p.plot_name }))
  } catch {
    myPlans.value = []
  }
}

async function submitCreate() {
  if (!createForm.plan_id || !createForm.title || !createForm.content) {
    ElMessage.warning('请填写计划、标题与内容')
    return
  }
  creating.value = true
  try {
    await store.create({ ...createForm, plan_id: createForm.plan_id })
    ElMessage.success('日记已发布')
    createVisible.value = false
  } finally {
    creating.value = false
  }
}

onMounted(fetch)
</script>

<style scoped>
.comment-item { background: #f5f7fa; border-radius: 6px; padding: 6px 10px; margin-bottom: 6px; font-size: 13px; color: #606266; }
</style>
