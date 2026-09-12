<template>
  <div class="page-card">
    <h3 class="page-title">平台概览</h3>
    <el-row :gutter="16">
      <el-col :span="6"><div class="stat-card"><div class="stat-num">{{ usersByRoleCount }}</div><div class="stat-label">注册用户</div></div></el-col>
      <el-col :span="6"><div class="stat-card"><div class="stat-num">{{ plotsByStatusCount }}</div><div class="stat-label">菜园地块</div></div></el-col>
      <el-col :span="6"><div class="stat-card"><div class="stat-num">{{ plansByStatusCount }}</div><div class="stat-label">种植计划</div></div></el-col>
      <el-col :span="6"><div class="stat-card"><div class="stat-num">{{ stats?.total_diaries ?? 0 }}</div><div class="stat-label">种植日记</div></div></el-col>
    </el-row>

    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>地块状态分布</template>
          <div v-for="(label, key) in PlotStatusMeta" :key="key" class="bar-row">
            <span class="bar-label">{{ label.label }}</span>
            <div class="bar-track"><div class="bar-fill" :style="{ width: percent(plotsByStatus[key]) + '%' }" /></div>
            <span class="bar-value">{{ plotsByStatus[key] || 0 }}</span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>种植计划状态</template>
          <div v-for="(label, key) in PlanStatusMeta" :key="key" class="bar-row">
            <span class="bar-label">{{ label.label }}</span>
            <div class="bar-track"><div class="bar-fill" :style="{ width: percent(plansByStatus[key]) + '%' }" /></div>
            <span class="bar-value">{{ plansByStatus[key] || 0 }}</span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>农友社区动态</template>
          <div v-for="(label, key) in PostTypeText" :key="key" class="bar-row">
            <span class="bar-label">{{ label }}</span>
            <div class="bar-track"><div class="bar-fill" :style="{ width: percent(postsByType[key]) + '%' }" /></div>
            <span class="bar-value">{{ postsByType[key] || 0 }}</span>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="12">
        <el-card shadow="never">
          <template #header>🌾 采摘提醒（近 7 天成熟）</template>
          <EmptyState v-if="!reminders.length" description="暂无临近成熟的作物" />
          <el-table v-else :data="reminders" size="small">
            <el-table-column prop="crop_name" label="作物" />
            <el-table-column prop="plot_name" label="地块" />
            <el-table-column label="预计成熟日">
              <template #default="{ row }">{{ formatDate(row.expected_harvest_date) }}</template>
            </el-table-column>
            <el-table-column label="状态">
              <template #default="{ row }"><StatusBadge :value="row.status" :meta-map="PlanStatusMeta" /></template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never">
          <template #header>🌿 当季作物推荐</template>
          <el-tag v-for="crop in recommendedCrops" :key="crop" style="margin: 4px" effect="plain">{{ crop }}</el-tag>
          <EmptyState v-if="!recommendedCrops.length" description="暂无推荐" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getDashboardStats, type DashboardStats } from '@/api/dashboard'
import { getRecommendations, getHarvestReminders, type PlantingPlan } from '@/api/plantingPlan'
import { PlotStatusMeta, PlanStatusMeta, PostTypeText } from '@/constants'
import StatusBadge from '@/components/StatusBadge.vue'
import EmptyState from '@/components/EmptyState.vue'
import { formatDate } from '@/utils/format'

const stats = ref<DashboardStats | null>(null)
const reminders = ref<PlantingPlan[]>([])
const recommendedCrops = ref<string[]>([])

const usersByRoleCount = computed(() => Object.values(stats.value?.users_by_role || {}).reduce((a, b) => a + b, 0))
const plotsByStatusCount = computed(() => Object.values(stats.value?.plots_by_status || {}).reduce((a, b) => a + b, 0))
const plansByStatusCount = computed(() => Object.values(stats.value?.plans_by_status || {}).reduce((a, b) => a + b, 0))
const plotsByStatus = computed(() => stats.value?.plots_by_status || {})
const plansByStatus = computed(() => stats.value?.plans_by_status || {})
const postsByType = computed(() => stats.value?.posts_by_type || {})

function percent(v?: number) {
  const max = 10
  return Math.min(100, ((v || 0) / max) * 100)
}

const seasonMap: Record<string, string> = {
  '1': 'winter', '2': 'winter', '3': 'spring', '4': 'spring', '5': 'spring',
  '6': 'summer', '7': 'summer', '8': 'summer', '9': 'autumn', '10': 'autumn', '11': 'autumn', '12': 'winter'
}

onMounted(async () => {
  try {
    stats.value = await getDashboardStats()
  } catch {
    stats.value = null
  }
  try {
    reminders.value = await getHarvestReminders()
  } catch {
    reminders.value = []
  }
  const month = String(new Date().getMonth() + 1)
  try {
    const recs = await getRecommendations(seasonMap[month] || 'spring')
    recommendedCrops.value = recs.map((r) => r.crops[0]).slice(0, 8)
  } catch {
    recommendedCrops.value = []
  }
})
</script>
