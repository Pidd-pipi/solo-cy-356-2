<template>
  <div class="page-card">
    <h3 class="page-title">操作审计日志（管理员）</h3>
    <DataTable :data="logs" :loading="loading" :total="total" :page-size="pagination.size.value" :current-page="pagination.page.value" @update:current-page="onPage">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="username" label="操作人" width="110" />
      <el-table-column prop="role" label="角色" width="90">
        <template #default="{ row }">{{ RoleText[row.role] || row.role }}</template>
      </el-table-column>
      <el-table-column prop="action" label="动作" min-width="180" />
      <el-table-column prop="resource_type" label="资源类型" width="130" />
      <el-table-column prop="detail" label="详情" min-width="180" />
      <el-table-column prop="ip" label="IP" width="130" />
      <el-table-column label="时间" width="160">
        <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
      </el-table-column>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listAuditLogs, type AuditLog } from '@/api/audit'
import { usePagination } from '@/hooks/usePagination'
import DataTable from '@/components/DataTable.vue'
import { RoleText } from '@/constants'
import { formatDateTime } from '@/utils/format'

const logs = ref<AuditLog[]>([])
const total = ref(0)
const loading = ref(false)
const pagination = usePagination()

async function fetch() {
  loading.value = true
  try {
    const data = await listAuditLogs({ page: pagination.page.value, page_size: pagination.size.value })
    logs.value = data.list
    total.value = data.total
  } finally {
    loading.value = false
  }
}

function onPage(page: number) {
  pagination.page.value = page
  fetch()
}

onMounted(fetch)
</script>
