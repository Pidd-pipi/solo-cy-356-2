<template>
  <el-empty :description="description">
    <template v-if="$slots.default" #default>
      <slot />
    </template>
  </el-empty>
</template>

<script setup lang="ts">
defineProps<{ description?: string }>()
</script>
