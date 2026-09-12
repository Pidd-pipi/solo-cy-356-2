import { defineStore } from 'pinia'
import { commentDiary, createDiary, listDiaries, likeDiary, type DiaryEntry } from '@/api/diary'

interface DiaryState {
  diaries: DiaryEntry[]
  total: number
  loading: boolean
}

export const useDiaryStore = defineStore('diary', {
  state: (): DiaryState => ({ diaries: [], total: 0, loading: false }),
  actions: {
    async fetchDiaries(params?: Record<string, any>) {
      this.loading = true
      try {
        const data = await listDiaries(params)
        this.diaries = data.list
        this.total = data.total
      } finally {
        this.loading = false
      }
    },
    async create(payload: Parameters<typeof createDiary>[0]) {
      await createDiary(payload)
      await this.fetchDiaries()
    },
    async like(id: number) {
      await likeDiary(id)
      await this.fetchDiaries()
    },
    async comment(id: number, content: string) {
      await commentDiary(id, content)
      await this.fetchDiaries()
    }
  }
})
