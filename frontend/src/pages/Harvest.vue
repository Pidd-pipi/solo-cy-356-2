<template>
  <div class="page-card">
    <div style="display: flex; justify-content: space-between; align-items: center">
      <h3 class="page-title">收成记录与年度统计</h3>
      <el-button type="primary" @click="openCreate">+ 记录收成</el-button>
    </div>

    <el-row :gutter="16" style="margin-bottom: 16px">
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>年度总收成（{{ year }} 年）</template>
          <div class="stat-num">{{ stats?.total_weight_kg?.toFixed(2) ?? '-' }} kg</div>
          <div class="stat-label">共 {{ stats?.harvest_count ?? 0 }} 次采摘</div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>按作物类型</template>
          <div v-for="(v, k) in stats?.by_crop_type || {}" :key="k" class="bar-row">
            <span class="bar-label">{{ CropTypeText[k] || k }}</span>
            <div class="bar-track"><div class="bar-fill" :style="{ width: pct(v, stats?.total_weight_kg) + '%' }" /></div>
            <span class="bar-value">{{ v?.toFixed(2) }} kg</span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>按品质</template>
          <div v-for="(v, k) in stats?.by_quality || {}" :key="k" class="bar-row">
            <span class="bar-label">{{ HarvestQualityText[k] || k }}</span>
            <div class="bar-track"><div class="bar-fill" :style="{ width: pct(v, stats?.harvest_count) + '%' }" /></div>
            <span class="bar-value">{{ v }} 次</span>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <DataTable :data="records" :loading="loading" :total="total" :page-size="pagination.size.value" :current-page="pagination.page.value" @update:current-page="onPage">
      <el-table-column prop="crop_name" label="作物" width="110" />
      <el-table-column label="采摘日期" width="120">
        <template #default="{ row }">{{ formatDate(row.harvest_date) }}</template>
      </el-table-column>
      <el-table-column label="重量" width="100">
        <template #default="{ row }">{{ formatWeight(row.weight_kg) }}</template>
      </el-table-column>
      <el-table-column label="品质" width="90">
        <template #default="{ row }"><StatusBadge :value="row.quality" :meta-map="qualityMeta" /></template>
      </el-table-column>
      <el-table-column prop="notes" label="备注" min-width="160" />
    </DataTable>

    <el-dialog v-model="createVisible" title="记录收成" width="520px">
      <el-form :model="createForm" label-width="90px">
        <el-form-item label="种植计划">
          <el-select v-model="createForm.plan_id" placeholder="选择生长中/采收中的计划">
            <el-option v-for="p in harvestablePlans" :key="p.id" :label="`${p.crop_name}（${PlanStatusMeta[p.status]?.label}）`" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="作物名称"><el-input v-model="createForm.crop_name" /></el-form-item>
        <el-form-item label="采摘日期"><el-date-picker v-model="createForm.harvest_date" type="date" value-format="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="重量(kg)"><el-input-number v-model="createForm.weight_kg" :min="0.1" :precision="2" /></el-form-item>
        <el-form-item label="品质">
          <el-select v-model="createForm.quality">
            <el-option v-for="(t, k) in HarvestQualityText" :key="k" :label="t" :value="k" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="createForm.notes" type="textarea" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { createHarvest, getAnnualStats, listHarvests, type AnnualStats, type HarvestRecord } from '@/api/harvest'
import { listPlans } from '@/api/plantingPlan'
import { usePagination } from '@/hooks/usePagination'
import DataTable from '@/components/DataTable.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { CropTypeText, HarvestQualityText, PlanStatusMeta } from '@/constants'
import { formatDate, formatWeight } from '@/utils/format'

const records = ref<HarvestRecord[]>([])
const total = ref(0)
const loading = ref(false)
const stats = ref<AnnualStats | null>(null)
const year = new Date().getFullYear()
const pagination = usePagination()
const createVisible = ref(false)
const creating = ref(false)
const harvestablePlans = ref<Array<{ id: number; crop_name: string; status: string }>>([])
const createForm = reactive({ plan_id: undefined as number | undefined, crop_name: '', harvest_date: '', weight_kg: 1, quality: 'good', notes: '' })

const qualityMeta = Object.fromEntries(Object.entries(HarvestQualityText).map(([k, v]) => [k, { label: v, type: 'primary' }]))

function pct(v?: number, total?: number) {
  if (!total) return 0
  return Math.min(100, ((v || 0) / total) * 100)
}

async function fetch() {
  loading.value = true
  try {
    const data = await listHarvests({ page: pagination.page.value, page_size: pagination.size.value })
    records.value = data.list
    total.value = data.total
  } finally {
    loading.value = false
  }
  try {
    stats.value = await getAnnualStats(year)
  } catch {
    stats.value = null
  }
}

function onPage(page: number) {
  pagination.page.value = page
  fetch()
}

async function openCreate() {
  createVisible.value = true
  try {
    const data = await listPlans({ page: 1, page_size: 100 })
    harvestablePlans.value = data.list.filter((p) => p.status === 'growing' || p.status === 'harvesting').map((p) => ({ id: p.id, crop_name: p.crop_name, status: p.status }))
  } catch {
    harvestablePlans.value = []
  }
}

async function submitCreate() {
  if (!createForm.plan_id || !createForm.harvest_date) {
    ElMessage.warning('请选择种植计划与采摘日期')
    return
  }
  creating.value = true
  try {
    await createHarvest({ ...createForm, plan_id: createForm.plan_id })
    ElMessage.success('收成已记录')
    createVisible.value = false
    await fetch()
  } finally {
    creating.value = false
  }
}

onMounted(fetch)
</script>
