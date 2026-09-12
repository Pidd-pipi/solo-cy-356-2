<template>
  <div>
    <el-table :data="data" v-loading="loading" border stripe style="width: 100%">
      <slot />
      <template #empty>
        <EmptyState description="暂无数据" />
      </template>
    </el-table>
    <div style="display: flex; justify-content: flex-end; margin-top: 12px">
      <el-pagination
        background
        layout="total, prev, pager, next"
        :total="total"
        :page-size="pageSize"
        :current-page="currentPage"
        @current-change="onPageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import EmptyState from './EmptyState.vue'

defineProps<{
  data: any[]
  loading?: boolean
  total: number
  pageSize: number
  currentPage: number
}>()
const emit = defineEmits<{ (e: 'update:currentPage', page: number): void }>()

function onPageChange(page: number) {
  emit('update:currentPage', page)
}
</script>
