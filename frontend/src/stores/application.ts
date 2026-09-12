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
    // 进行中的申请（待审核/候补中），同一用户同一时间仅允许一份
    activeApplication: (s) => s.mine.find((a) => a.status === 'pending' || a.status === 'waitlisted')
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
