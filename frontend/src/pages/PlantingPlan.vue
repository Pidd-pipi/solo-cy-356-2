<template>
  <div class="page-card">
    <div style="display: flex; justify-content: space-between; align-items: center">
      <h3 class="page-title">种植计划</h3>
      <el-button type="primary" @click="openCreate">+ 制定种植计划</el-button>
    </div>

    <el-card shadow="never" style="margin-bottom: 16px">
      <template #header>🌿 当季作物推荐（按季节）</template>
      <el-radio-group v-model="season" @change="loadRecommendations">
        <el-radio-button v-for="(label, key) in SeasonText" :key="key" :value="key">{{ label }}</el-radio-button>
      </el-radio-group>
      <div style="margin-top: 12px">
        <el-tag v-for="c in recommendations" :key="c" style="margin: 4px" effect="plain">{{ c }}</el-tag>
        <span v-if="!recommendations.length" class="muted">暂无推荐</span>
      </div>
    </el-card>

    <DataTable :data="store.plans" :loading="store.loading" :total="store.total" :page-size="pagination.size.value" :current-page="pagination.page.value" @update:current-page="onPage">
      <el-table-column prop="crop_name" label="作物" width="110" />
      <el-table-column label="类型" width="90">
        <template #default="{ row }">{{ CropTypeText[row.crop_type] }}</template>
      </el-table-column>
      <el-table-column label="季节" width="80">
        <template #default="{ row }">{{ SeasonText[row.season] }}</template>
      </el-table-column>
      <el-table-column prop="plot_name" label="地块" min-width="120" />
      <el-table-column label="播种日" width="110">
        <template #default="{ row }">{{ formatDate(row.plant_date) }}</template>
      </el-table-column>
      <el-table-column label="预计收获" width="110">
        <template #default="{ row }">{{ formatDate(row.expected_harvest_date) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="110">
        <template #default="{ row }"><StatusBadge :value="row.status" :meta-map="PlanStatusMeta" /></template>
      </el-table-column>
      <el-table-column label="操作" width="130">
        <template #default="{ row }">
          <el-button v-if="PlanStatusNext[row.status]" type="primary" size="small" @click="transition(row)">{{ PlanStatusActions[row.status] }}</el-button>
          <span v-else class="muted">已完成</span>
        </template>
      </el-table-column>
    </DataTable>

    <el-dialog v-model="createVisible" title="制定种植计划" width="520px">
      <el-form :model="createForm" label-width="90px">
        <el-form-item label="地块">
          <el-select v-model="createForm.plot_id" placeholder="选择已认养地块">
            <el-option v-for="p in myAdoptedPlots" :key="p.id" :label="`${p.code} ${p.name}`" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="作物名称"><el-input v-model="createForm.crop_name" placeholder="如 番茄 / 菠菜" /></el-form-item>
        <el-form-item label="作物类型">
          <el-select v-model="createForm.crop_type">
            <el-option v-for="(t, k) in CropTypeText" :key="k" :label="t" :value="k" />
          </el-select>
        </el-form-item>
        <el-form-item label="季节">
          <el-select v-model="createForm.season">
            <el-option v-for="(t, k) in SeasonText" :key="k" :label="t" :value="k" />
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
import { usePlanStore } from '@/stores/plantingPlan'
import { type PlantingPlan } from '@/api/plantingPlan'
import { listPlots } from '@/api/plot'
import { usePagination } from '@/hooks/usePagination'
import DataTable from '@/components/DataTable.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { CropTypeText, SeasonText, PlanStatusMeta, PlanStatusNext, PlanStatusActions } from '@/constants'
import { formatDate } from '@/utils/format'

const store = usePlanStore()
const pagination = usePagination()
const season = ref<string>('spring')
const recommendations = ref<string[]>([])
const createVisible = ref(false)
const creating = ref(false)
const myAdoptedPlots = ref<Array<{ id: number; code: string; name: string }>>([])
const createForm = reactive({ plot_id: undefined as number | undefined, crop_name: '', crop_type: 'vegetable', season: 'spring', notes: '' })

async function fetch() {
  await store.fetchPlans({ page: pagination.page.value, page_size: pagination.size.value })
}

function onPage(page: number) {
  pagination.page.value = page
  fetch()
}

async function loadRecommendations() {
  try {
    const { getRecommendations } = await import('@/api/plantingPlan')
    const recs = await getRecommendations(season.value)
    recommendations.value = recs.map((r) => r.crops[0])
  } catch {
    recommendations.value = []
  }
}

async function openCreate() {
  createVisible.value = true
  try {
    const data = await listPlots({ page: 1, page_size: 100 })
    myAdoptedPlots.value = data.list.filter((p) => p.status === 'adopted').map((p) => ({ id: p.id, code: p.code, name: p.name }))
  } catch {
    myAdoptedPlots.value = []
  }
}

async function submitCreate() {
  if (!createForm.plot_id) {
    ElMessage.warning('请选择已认养的地块（先到“地块认养”页认养）')
    return
  }
  creating.value = true
  try {
    await store.create({ ...createForm, plot_id: createForm.plot_id })
    ElMessage.success('种植计划创建成功')
    createVisible.value = false
  } finally {
    creating.value = false
  }
}

async function transition(row: PlantingPlan) {
  const next = PlanStatusNext[row.status]
  await store.changeStatus(row.id, next)
  ElMessage.success(`状态已更新为 ${PlanStatusMeta[next]?.label}`)
}

onMounted(() => {
  fetch()
  loadRecommendations()
})
</script>
