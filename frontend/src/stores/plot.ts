import { defineStore } from 'pinia'
import { adoptPlot, listPlots, type Plot } from '@/api/plot'

interface PlotState {
  plots: Plot[]
  total: number
  loading: boolean
}

export const usePlotStore = defineStore('plot', {
  state: (): PlotState => ({ plots: [], total: 0, loading: false }),
  actions: {
    async fetchPlots(params?: Record<string, any>) {
      this.loading = true
      try {
        const data = await listPlots(params)
        this.plots = data.list
        this.total = data.total
      } finally {
        this.loading = false
      }
    },
    async adopt(id: number) {
      await adoptPlot(id)
      await this.fetchPlots()
    }
  }
})
