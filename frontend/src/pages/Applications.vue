<template>
  <div class="page-card">
    <h3 class="page-title">我的认养申请</h3>

    <el-card shadow="never" style="margin-bottom: 16px">
      <template #header>发起认养申请</template>
      <el-alert
        v-if="store.pendingApplication"
        type="warning"
        :closable="false"
        style="margin-bottom: 12px"
        :title="`您已有一份待审核的申请（地块 ${store.pendingApplication.plot?.code}），同一居民仅允许一份待审核申请；仍可申请其他地块进入候补队列。`"
      />
      <el-alert
        v-else-if="store.waitlistedApplications.length"
        type="info"
        :closable="false"
        style="margin-bottom: 12px"
        :title="`您有 ${store.waitlistedApplications.length} 份候补中的申请，仍可继续申请其他空闲地块。`"
      />
      <el-form inline @submit.prevent>
        <el-form-item label="空闲地块">
          <el-select v-model="applyForm.plotId" placeholder="选择地块" style="width: 240px">
            <el-option v-for="p in availablePlots" :key="p.id" :label="`${p.code} ${p.name}（${formatArea(p.area)}）`" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="申请理由">
          <el-input v-model="applyForm.reason" placeholder="说说你的种植计划（可选）" style="width: 320px" maxlength="512" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="submitting" :disabled="!applyForm.plotId" @click="submitApply">
            提交申请
          </el-button>
        </el-form-item>
      </el-form>
      <div class="tip">同一居民在同一地块仅允许一份进行中的申请，同一时间仅允许一份待审核申请；地块已有待审申请时，新申请将自动进入候补队列，前序申请被驳回/撤回或地块释放后按申请时间依次转正。</div>
    </el-card>

    <el-card shadow="never">
      <template #header>申请进度</template>
      <DataTable :data="store.mine" :loading="store.loading" :total="store.mineTotal" :page-size="pagination.size.value" :current-page="pagination.page.value" @update:current-page="onPage">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="地块" min-width="160">
          <template #default="{ row }">{{ row.plot?.code }} {{ row.plot?.name }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }"><StatusBadge :value="row.status" :meta-map="ApplicationStatusMeta" /></template>
        </el-table-column>
        <el-table-column prop="reason" label="申请理由" min-width="160" show-overflow-tooltip />
        <el-table-column label="审核备注" min-width="140">
          <template #default="{ row }">{{ row.review_note || '-' }}</template>
        </el-table-column>
        <el-table-column label="申请时间" width="160">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="审核时间" width="160">
          <template #default="{ row }">{{ row.reviewed_at ? formatDateTime(row.reviewed_at) : '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button v-if="row.status === 'pending' || row.status === 'waitlisted'" type="warning" size="small" @click="withdraw(row)">撤回</el-button>
          </template>
        </el-table-column>
      </DataTable>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useApplicationStore } from '@/stores/application'
import { listPlots, type Plot } from '@/api/plot'
import type { AdoptionApplication } from '@/api/application'
import { usePagination } from '@/hooks/usePagination'
import DataTable from '@/components/DataTable.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { ApplicationStatusMeta } from '@/constants'
import { formatArea, formatDateTime } from '@/utils/format'

const store = useApplicationStore()
const pagination = usePagination()

const availablePlots = ref<Plot[]>([])
const submitting = ref(false)
const applyForm = reactive({ plotId: undefined as number | undefined, reason: '' })

async function fetch() {
  await store.fetchMine({ page: pagination.page.value, page_size: pagination.size.value })
}

async function fetchPlots() {
  const data = await listPlots({ status: 'available', page: 1, page_size: 100 })
  availablePlots.value = data.list
}

function onPage(page: number) {
  pagination.page.value = page
  fetch()
}

async function submitApply() {
  if (!applyForm.plotId) return
  submitting.value = true
  try {
    const app = await store.apply(applyForm.plotId, applyForm.reason)
    if (app.status === 'waitlisted') {
      ElMessage.info('该地块已有待审申请，您的申请已进入候补队列')
    } else {
      ElMessage.success('认养申请已提交，请等待管理员审核')
    }
    applyForm.plotId = undefined
    applyForm.reason = ''
    pagination.reset()
    await Promise.all([fetch(), fetchPlots()])
  } finally {
    submitting.value = false
  }
}

async function withdraw(row: AdoptionApplication) {
  try {
    await ElMessageBox.confirm(`确认撤回地块 ${row.plot?.code} 的认养申请吗？`, '撤回确认', { type: 'warning' })
  } catch {
    return
  }
  await store.withdraw(row.id)
  ElMessage.success('认养申请已撤回')
  await Promise.all([fetch(), fetchPlots()])
}

onMounted(() => {
  fetch()
  fetchPlots()
})
</script>

<style scoped>
.tip { color: #909399; font-size: 12px; margin-top: 4px; }
</style>
