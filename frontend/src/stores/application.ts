import { defineStore } from 'pinia'
import {
  applyForPlot,
  listApplications,
  listMyApplications,
  reviewApplication,
  withdrawApplication,
  type AdoptionApplication
} from '@/api/application'

interface ApplicationState {
  mine: AdoptionApplication[]
  mineTotal: number
  adminList: AdoptionApplication[]
  adminTotal: number
  loading: boolean
}

export const useApplicationStore = defineStore('application', {
  state: (): ApplicationState => ({ mine: [], mineTotal: 0, adminList: [], adminTotal: 0, loading: false }),
  getters: {
    // 待审核中的申请（同一居民同一时间仅允许一份；候补不限制）
    pendingApplication: (s) => s.mine.find((a) => a.status === 'pending'),
    waitlistedApplications: (s) => s.mine.filter((a) => a.status === 'waitlisted')
  },
  actions: {
    async fetchMine(params?: Record<string, any>) {
      this.loading = true
      try {
        const data = await listMyApplications(params)
        this.mine = data.list
        this.mineTotal = data.total
      } finally {
        this.loading = false
      }
    },
    async fetchAll(params?: Record<string, any>) {
      this.loading = true
      try {
        const data = await listApplications(params)
        this.adminList = data.list
        this.adminTotal = data.total
      } finally {
        this.loading = false
      }
    },
    async apply(plotId: number, reason: string) {
      const app = await applyForPlot({ plot_id: plotId, reason })
      return app
    },
    async withdraw(id: number) {
      await withdrawApplication(id)
    },
    async review(id: number, action: 'approve' | 'reject', note: string) {
      await reviewApplication(id, { action, note })
    }
  }
})
