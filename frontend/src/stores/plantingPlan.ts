import { defineStore } from 'pinia'
import { createPlan, listPlans, updatePlanStatus, type PlantingPlan } from '@/api/plantingPlan'

interface PlanState {
  plans: PlantingPlan[]
  total: number
  loading: boolean
}

export const usePlanStore = defineStore('plantingPlan', {
  state: (): PlanState => ({ plans: [], total: 0, loading: false }),
  actions: {
    async fetchPlans(params?: Record<string, any>) {
      this.loading = true
      try {
        const data = await listPlans(params)
        this.plans = data.list
        this.total = data.total
      } finally {
        this.loading = false
      }
    },
    async create(payload: Parameters<typeof createPlan>[0]) {
      await createPlan(payload)
      await this.fetchPlans()
    },
    async changeStatus(id: number, status: string) {
      await updatePlanStatus(id, status)
      await this.fetchPlans()
    }
  }
})
