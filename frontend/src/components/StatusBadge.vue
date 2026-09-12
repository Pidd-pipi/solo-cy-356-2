<template>
  <el-tag :type="meta?.type || 'info'" size="small" effect="light">{{ meta?.label || value }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ value: string; metaMap?: Record<string, { label: string; type: string }> }>()
const meta = computed(() => (props.metaMap ? props.metaMap[props.value] : undefined))
</script>
