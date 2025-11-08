<template>
  <div class="task-cron-page">
    <a-typography-title :level="2" class="page-title">
      ⏰ 定时任务管理
    </a-typography-title>

    <a-row :gutter="24">
      <a-col :xs="24" :md="10">
        <a-card title="新增任务" bordered class="form-card">
          <a-alert
            type="info"
            show-icon
            message="Cron 表达式需包含秒（WithSeconds）。例如：0 0 * * * * 表示每小时整点。"
            class="cron-tip"
          />

          <a-form layout="vertical" @submit.prevent="handleAddTask">
            <a-form-item label="任务名称">
              <a-select
                v-model:value="formState.name"
                placeholder="请选择后端已注册的任务"
                :loading="dropdownLoading"
                allow-clear
              >
                <a-select-option
                  v-for="taskName in availableTaskNames"
                  :key="taskName"
                  :value="taskName"
                >
                  {{ taskName }}
                </a-select-option>
              </a-select>
            </a-form-item>

            <a-form-item label="Cron 表达式">
              <a-input
                v-model:value="formState.spec"
                placeholder="例如：0 */5 * * * *（每 5 分钟执行）"
                allow-clear
              />
            </a-form-item>

            <a-form-item label="任务参数（可选）">
              <a-textarea
                v-model:value="formState.data"
                :auto-size="{ minRows: 3, maxRows: 6 }"
                placeholder="一条龙名字或者配置组名字，多个配置组用空格分隔"
                allow-clear
              />
            </a-form-item>

            <div class="form-actions">
              <a-button type="primary" :loading="formLoading" :disabled="submitDisabled" @click="handleAddTask">
                添加任务
              </a-button>
              <a-button style="margin-left: 8px" @click="resetForm">
                重置
              </a-button>
            </div>
          </a-form>

          <div class="quick-presets">
            <h4>常用 Cron 示例</h4>
            <div class="preset-group">
              <a-tag
                v-for="item in presetSpecs"
                :key="item.spec"
                color="pink"
                @click="applyPreset(item.spec)"
              >
                {{ item.label }}
              </a-tag>
            </div>
          </div>
        </a-card>
      </a-col>

      <a-col :xs="24" :md="14">
        <a-card title="已配置任务" bordered>
          <a-spin :spinning="tableLoading">
            <div v-if="taskCronList.length > 0">
              <a-table
                :data-source="taskCronList"
                :columns="columns"
                row-key="id"
                :pagination="false"
                size="middle"
                bordered
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'action'">
                    <a-space>
                      <a-tooltip title="删除任务">
                        <a-button
                          type="text"
                          danger
                          size="small"
                          @click="confirmRemove(record)"
                        >
                          删除
                        </a-button>
                      </a-tooltip>
                    </a-space>
                  </template>
                  <template v-else-if="column.key === 'next'">
                    <span>{{ record.next || '调度中...' }}</span>
                  </template>
                  <template v-else>
                    <span>{{ record[column.dataIndex] }}</span>
                  </template>
                </template>
              </a-table>
            </div>

            <a-empty v-else-if="!tableLoading" description="暂无任务" />
          </a-spin>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, computed } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'

const formState = reactive({
  name: '',
  spec: '',
  data: ''
})

const presetSpecs = [
  { label: '每天4点5分', spec: '0 5 4 * * *' },
  { label: '每个月的1号', spec: '0 0 0 1 * *' },
  { label: '每小时整点', spec: '0 0 * * * *' },
  { label: '每天 3 点', spec: '0 0 3 * * *' }
]

const taskCronList = ref([])
const availableTaskNames = ref([])
const tableLoading = ref(false)
const formLoading = ref(false)
const dropdownLoading = ref(false)

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '任务名称', dataIndex: 'name', key: 'name', width: 150 },
  { title: 'Cron 表达式', dataIndex: 'spec', key: 'spec', width: 180 },
  { title: '下次执行时间', dataIndex: 'next', key: 'next', width: 200 },
  { title: '任务参数', dataIndex: 'data', key: 'data' },
  { title: '操作', key: 'action', width: 100 }
]

const submitDisabled = computed(() => {
  return !formState.name || !formState.spec.trim()
})

const fetchTaskList = async () => {
  tableLoading.value = true
  try {
    const list = await apiMethods.getTaskCronList()
    taskCronList.value = Array.isArray(list) ? list : []
  } catch (error) {
    message.error('获取任务列表失败')
  } finally {
    tableLoading.value = false
  }
}

const fetchTaskNameOptions = async () => {
  dropdownLoading.value = true
  try {
    const names = await apiMethods.getAvailableTaskCronNames()
    availableTaskNames.value = Array.isArray(names) ? names : []
  } catch (error) {
    message.error('获取可用任务名称失败')
  } finally {
    dropdownLoading.value = false
  }
}

const handleAddTask = async () => {
  if (submitDisabled.value) {
    message.warning('请选择任务并填写 Cron 表达式')
    return
  }

  formLoading.value = true
  try {
    const payload = {
      name: formState.name,
      spec: formState.spec.trim(),
      data: formState.data?.trim() || ''
    }
    const res = await apiMethods.addTaskCron(payload)
    const msg = typeof res === 'string' ? res : '任务已添加'
    message.success(msg)
    resetForm()
    fetchTaskList()
  } catch (error) {
    message.error('添加任务失败')
  } finally {
    formLoading.value = false
  }
}

const resetForm = () => {
  formState.name = ''
  formState.spec = ''
  formState.data = ''
}

const applyPreset = (spec) => {
  formState.spec = spec
}

const confirmRemove = (record) => {
  Modal.confirm({
    title: `确认删除任务「${record.name}」?`,
    content: '删除后需要重新创建才能恢复。',
    okText: '确定',
    cancelText: '再想想',
    okButtonProps: { danger: true },
    onOk: () => removeTask(record.id)
  })
}

const removeTask = async (id) => {
  try {
    const res = await apiMethods.removeTaskCron(id)
    const msg = typeof res === 'string' ? res : '任务已删除'
    message.success(msg)
    fetchTaskList()
  } catch (error) {
    message.error('删除任务失败')
  }
}

onMounted(() => {
  fetchTaskList()
  fetchTaskNameOptions()
})
</script>

<style scoped>
.task-cron-page {
  padding: 24px;
  background: #fff0f6;
  min-height: 100vh;
}

.page-title {
  color: #ff5c8d !important;
  text-align: center;
  margin-bottom: 24px;
}

.form-card {
  margin-bottom: 24px;
}

.cron-tip {
  margin-bottom: 16px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
}

.quick-presets {
  margin-top: 24px;
}

.quick-presets h4 {
  margin-bottom: 8px;
  color: #ff6699;
}

.preset-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.preset-group :deep(.ant-tag) {
  cursor: pointer;
}

@media (max-width: 768px) {
  .form-actions {
    flex-direction: column;
    gap: 8px;
  }

  .form-actions .ant-btn {
    width: 100%;
  }
}
</style>
