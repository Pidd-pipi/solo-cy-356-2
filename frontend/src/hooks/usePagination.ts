import { computed, ref } from 'vue'

// 共享分页 hook
export function usePagination(pageSize = 10) {
  const page = ref(1)
  const size = ref(pageSize)
  const total = ref(0)
  const params = computed(() => ({ page: page.value, page_size: size.value }))
  function reset() {
    page.value = 1
  }
  return { page, size, total, params, reset }
}
