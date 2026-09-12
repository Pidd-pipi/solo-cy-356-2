<template>
  <slot :confirm="confirm" />
</template>

<script setup lang="ts">
import { ElMessageBox } from 'element-plus'

const props = defineProps<{ title?: string; message?: string; confirmText?: string }>()

async function confirm(): Promise<boolean> {
  try {
    await ElMessageBox.confirm(props.message || '确认执行该操作吗？', props.title || '操作确认', {
      confirmButtonText: props.confirmText || '确认',
      cancelButtonText: '取消',
      type: 'warning'
    })
    return true
  } catch {
    return false
  }
}
</script>
