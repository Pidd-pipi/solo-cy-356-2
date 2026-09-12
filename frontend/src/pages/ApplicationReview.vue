<template>
  <div class="page-card">
    <div style="display: flex; justify-content: space-between; align-items: center">
      <h3 class="page-title">认养申请审核（管理员）</h3>
      <el-select v-model="statusFilter" placeholder="全部状态" clearable style="width: 160px" @change="onFilter">
        <el-option v-for="(m, s) in ApplicationStatusMeta" :key="s" :label="m.label" :value="s" />
      </el-select>
    </div>

    <DataTable :data="store.adminList" :loading="store.loading" :total="store.adminTotal" :page-size="pagination.size.value" :current-page="pagination.page.value" @update:current-page="onPage">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column label="申请人" width="130">
        <template #default="{ row }">{{ row.user?.nickname || row.user?.username }}</template>
      </el-table-column>
      <el-table-column label="地块" min-width="150">
        <template #default="{ row }">{{ row.plot?.code }} {{ row.plot?.name }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }"><StatusBadge :value="row.status" :meta-map="ApplicationStatusMeta" /></template>
      </el-table-column>
      <el-table-column prop="reason" label="申请理由" min-width="160" show-overflow-tooltip />
      <el-table-column label="申请时间" width="160">
        <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="审核备注" min-width="130">
        <template #default="{ row }">{{ row.review_note || '-' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="170" fixed="right">
        <template #default="{ row }">
          <template v-if="row.status === 'pending'">
            <el-button type="success" size="small" @click="approve(row)">通过</el-button>
            <el-button type="danger" size="small" @click="openReject(row)">驳回</el-button>
          </template>
          <span v-else>-</span>
        </template>
      </el-table-column>
    </DataTable>

    <el-dialog v-model="rejectVisible" title="驳回认养申请" width="480px">
      <el-form label-width="90px">
        <el-form-item label="申请">
          <span>{{ rejectTarget?.user?.nickname || rejectTarget?.user?.username }} → {{ rejectTarget?.plot?.code }} {{ rejectTarget?.plot?.name }}</span>
        </el-form-item>
        <el-form-item label="驳回原因">
          <el-input v-model="rejectNote" type="textarea" :rows="3" maxlength="512" placeholder="将展示给申请人（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="reviewing" @click="submitReject">确认驳回</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useApplicationStore } from '@/stores/application'
import type { AdoptionApplication } from '@/api/application'
import { usePagination } from '@/hooks/usePagination'
import DataTable from '@/components/DataTable.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { ApplicationStatusMeta } from '@/constants'
import { formatDateTime } from '@/utils/format'

const store = useApplicationStore()
const pagination = usePagination()

const statusFilter = ref('')
const reviewing = ref(false)
const rejectVisible = ref(false)
const rejectTarget = ref<AdoptionApplication | null>(null)
const rejectNote = ref('')

async function fetch() {
  await store.fetchAll({ page: pagination.page.value, page_size: pagination.size.value, status: statusFilter.value || undefined })
}

function onPage(page: number) {
  pagination.page.value = page
  fetch()
}

function onFilter() {
  pagination.reset()
  fetch()
}

async function approve(row: AdoptionApplication) {
  try {
    await ElMessageBox.confirm(
      `确认通过 ${row.user?.nickname || row.user?.username} 对地块 ${row.plot?.code} 的认养申请吗？通过后地块将立即置为已认养。`,
      '审核通过确认',
      { type: 'success' }
    )
  } catch {
    return
  }
  reviewing.value = true
  try {
    await store.review(row.id, 'approve', '')
    ElMessage.success('已通过申请，地块认养状态已同步')
    await fetch()
  } finally {
    reviewing.value = false
  }
}

function openReject(row: AdoptionApplication) {
  rejectTarget.value = row
  rejectNote.value = ''
  rejectVisible.value = true
}

async function submitReject() {
  if (!rejectTarget.value) return
  reviewing.value = true
  try {
    await store.review(rejectTarget.value.id, 'reject', rejectNote.value)
    ElMessage.success('已驳回，候补申请已按申请时间自动转正')
    rejectVisible.value = false
    await fetch()
  } finally {
    reviewing.value = false
  }
}

onMounted(fetch)
</script>
