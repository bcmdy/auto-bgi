<template>
  <div class="kawaii-page">
    <div class="main-container">
      <a-page-header
        title="一条龙计划"
        sub-title="按日期 / 材料条件启动配置组"
        @back="goHome"
      />

      <a-card bordered class="glass">
        <div style="display: flex; gap: 12px; flex-wrap: wrap; margin-bottom: 12px;">
          <a-button type="primary" @click="openCreate">新增</a-button>
          <a-button :loading="loading" @click="refresh">刷新</a-button>
        </div>

        <a-card v-if="formVisible" bordered style="margin-bottom: 14px;">
          <a-row :gutter="12">
            <a-col :xs="24" :sm="12">
              <div style="margin-bottom: 6px; font-weight: 700; color: #ff66a3;">类型</div>
              <a-select v-model:value="form.type" style="width: 100%">
                <a-select-option value="日期">日期</a-select-option>
                <a-select-option value="材料">材料</a-select-option>
              </a-select>
            </a-col>

            <a-col v-if="form.type === '日期'" :xs="24" :sm="12">
              <div style="margin-bottom: 6px; font-weight: 700; color: #ff66a3;">日期</div>
              <a-select v-model:value="form.weekday" style="width: 100%">
                <a-select-option v-for="d in weekdays" :key="d" :value="d">{{ d }}</a-select-option>
              </a-select>
            </a-col>

            <a-col v-else :xs="24" :sm="12">
              <div style="margin-bottom: 6px; font-weight: 700; color: #ff66a3;">材料名称</div>
              <a-input v-model:value="form.materialName" placeholder="例如：清心" />
            </a-col>

            <a-col v-if="form.type === '材料'" :xs="24" :sm="12">
              <div style="margin-bottom: 6px; font-weight: 700; color: #ff66a3;">数量（小于则启动）</div>
              <a-input-number v-model:value="form.materialCount" :min="0" style="width: 100%" />
            </a-col>

            <a-col :xs="24" :sm="12">
              <div style="margin-bottom: 6px; font-weight: 700; color: #ff66a3;">配置组名称</div>
              <a-select v-model:value="form.groupName" style="width: 100%" :loading="loadingGroups">
                <a-select-option v-for="g in groups" :key="g" :value="g">{{ g }}</a-select-option>
              </a-select>
            </a-col>
          </a-row>

          <div style="display: flex; justify-content: flex-end; gap: 10px; margin-top: 12px;">
            <a-button :disabled="saving" @click="closeForm">取消</a-button>
            <a-button type="primary" :loading="saving" @click="save">保存</a-button>
          </div>
        </a-card>

        <a-table
          :data-source="plans"
          :columns="columns"
          :loading="loading"
          row-key="id"
          :pagination="false"
          size="middle"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'actions'">
              <a-space>
                <a-button size="small" @click="openEdit(record)">编辑</a-button>
                <a-button size="small" danger @click="remove(record)">删除</a-button>
                <!-- <a-button size="small" type="primary" @click="run(record)">执行</a-button> -->
              </a-space>
            </template>
          </template>
        </a-table>
      </a-card>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { Modal, message } from 'ant-design-vue'
import { useRouter } from 'vue-router'
import { apiMethods } from '@/utils/api'

const router = useRouter()
const loading = ref(false)
const saving = ref(false)
const loadingGroups = ref(false)

const plans = ref([])
const groups = ref([])
const weekdays = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']

const formVisible = ref(false)
const editingId = ref(null)
const form = reactive({
  type: '日期',
  weekday: '周一',
  materialName: '清心',
  materialCount: 500,
  groupName: ''
})

const columns = computed(() => [
  { title: '类型', dataIndex: 'type', key: 'type', width: 90 },
  { title: '值', dataIndex: 'value', key: 'value' },
  { title: '配置组', dataIndex: 'groupName', key: 'groupName' },
  { title: '操作', key: 'actions', width: 220 }
])

const goHome = () => router.push('/')

const normalizeGroupOptions = (raw) => {
  if (Array.isArray(raw)) return raw
  if (raw && Array.isArray(raw.items)) return raw.items
  if (raw && raw.data && Array.isArray(raw.data)) return raw.data
  return []
}

const buildValue = () => {
  if (form.type === '日期') return form.weekday || ''
  const name = (form.materialName || '').trim()
  const count = Number(form.materialCount || 0)
  return `${name}-${count}`
}

const parseValue = (type, value) => {
  const v = (value || '').toString()
  if (type === '日期') return { weekday: v || '周一', materialName: '清心', materialCount: 500 }
  const parts = v.split('-')
  if (parts.length < 2) return { weekday: '周一', materialName: v || '清心', materialCount: 500 }
  const materialName = parts.slice(0, -1).join('-') || '清心'
  const materialCount = Number(parts[parts.length - 1]) || 0
  return { weekday: '周一', materialName, materialCount }
}

const loadGroups = async () => {
  loadingGroups.value = true
  try {
    const res = await apiMethods.getListGroups()
    groups.value = normalizeGroupOptions(res)
    if (!form.groupName) {
      form.groupName = groups.value.includes('全自动') ? '全自动' : (groups.value[0] || '')
    }
  } catch (e) {
    message.error('加载配置组失败')
    groups.value = []
  } finally {
    loadingGroups.value = false
  }
}

const loadPlans = async () => {
  loading.value = true
  try {
    const res = await apiMethods.getOneLongPlanList()
    plans.value = res?.data || []
  } catch (e) {
    message.error('加载计划失败')
    plans.value = []
  } finally {
    loading.value = false
  }
}

const refresh = async () => {
  await Promise.all([loadGroups(), loadPlans()])
}

const openCreate = () => {
  editingId.value = null
  form.type = '日期'
  form.weekday = '周一'
  form.materialName = '清心'
  form.materialCount = 500
  form.groupName = groups.value.includes('全自动') ? '全自动' : (groups.value[0] || '')
  formVisible.value = true
}

const openEdit = (record) => {
  editingId.value = record.id
  form.type = record.type || '日期'
  const extra = parseValue(form.type, record.value)
  form.weekday = extra.weekday
  form.materialName = extra.materialName
  form.materialCount = extra.materialCount
  form.groupName = record.groupName || ''
  formVisible.value = true
}

const closeForm = () => {
  formVisible.value = false
  editingId.value = null
}

const save = async () => {
  if (saving.value) return
  saving.value = true
  try {
    const payload = {
      id: editingId.value || 0,
      type: form.type,
      value: buildValue(),
      groupName: form.groupName
    }
    if (editingId.value) {
      await apiMethods.updateOneLongPlan(payload)
      message.success('已保存')
    } else {
      await apiMethods.addOneLongPlan(payload)
      message.success('已新增')
    }
    closeForm()
    await loadPlans()
  } catch (e) {
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}

const remove = (record) => {
  Modal.confirm({
    title: '确认删除？',
    content: `删除【${record.type || ''} - ${record.value || ''} -> ${record.groupName || ''}】后不可恢复`,
    okText: '删除',
    cancelText: '取消',
    okType: 'danger',
    centered: true,
    onOk: async () => {
      try {
        await apiMethods.deleteOneLongPlan(record.id)
        message.success('已删除')
        await loadPlans()
      } catch (e) {
        message.error('删除失败')
      }
    }
  })
}

const run = async (record) => {
  try {
    const res = await apiMethods.runOneLongPlan(record.id)
    const data = res?.data || {}
    const startedGroups = data.startedGroups || []
    if (Array.isArray(startedGroups) && startedGroups.length) {
      message.success(`已启动：${startedGroups.join('、')}`)
    } else {
      message.info(data.msg || '未启动任何配置组')
    }
  } catch (e) {
    message.error('执行失败')
  }
}

onMounted(async () => {
  await refresh()
})
</script>
