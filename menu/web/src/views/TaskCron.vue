<template>
  <div class="task-cron-page">
    <a-typography-title :level="2" class="page-title">
      定时任务管理
    </a-typography-title>

    <a-row :gutter="[16, 16]" class="content-row">
      <a-col :xs="24" :md="10" class="form-col">
        <a-card title="新增任务" bordered class="form-card">
          <a-alert
            type="info"
            show-icon
            :message="cronTip"
            class="cron-tip"
          />

          <a-form layout="vertical" @submit.prevent="handleSubmitTask">
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
                placeholder="例如：*/5 * * * *（每 5 分钟执行）"
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
              <a-button type="primary" :loading="formLoading" :disabled="submitDisabled" @click="handleSubmitTask">
                {{ isEditing ? '保存修改' : '添加任务' }}
              </a-button>
              <a-button class="ghost-button" @click="resetForm">
                {{ isEditing ? '取消编辑' : '重置' }}
              </a-button>

              <a-button class="back-button" @click="comeBack">
                返回首页
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

      <a-col :xs="24" :md="14" class="table-col">
        <a-card title="已配置任务" bordered class="table-card">
          <a-spin :spinning="tableLoading">
            <div v-if="taskCronList.length > 0" class="table-wrapper">
              <a-table
                :data-source="taskCronList"
                :columns="columns"
                :row-key="getRowKey"
                :pagination="false"
                size="middle"
                bordered
                :scroll="{ x: 860 }"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'action'">
                    <a-space>
                      <a-tooltip title="编辑任务">
                        <a-button
                          type="text"
                          size="small"
                          @click="startEdit(record)"
                        >
                          编辑
                        </a-button>
                      </a-tooltip>
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
                      <a-tooltip :title="record.paused ? '恢复任务' : '暂停任务'">
                        <a-popconfirm
                          :title="record.paused ? '确认恢复该任务？' : '确认暂停该任务？'"
                          ok-text="确定"
                          cancel-text="取消"
                          @confirm="togglePause(record)"
                        >
                          <a-button type="text" size="small">
                            {{ record.paused ? '恢复' : '暂停' }}
                          </a-button>
                        </a-popconfirm>
                      </a-tooltip>
                    </a-space>
                  </template>
                  <template v-else-if="column.key === 'next'">
                    <span>{{ record.paused ? '已暂停' : (record.next || '调度中...') }}</span>
                  </template>
                  <template v-else-if="column.key === 'status'">
                    <a-tag :color="record.paused ? 'orange' : 'green'">
                      {{ record.paused ? '已暂停' : '运行中' }}
                    </a-tag>
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

        <div class="visual-gallery">
          <img
            class="gallery-img"
            src="https://upload-bbs.miyoushe.com/upload/2025/09/30/162891450/40ba73d1c8ce78112681a7ed7137dad1_6952277989857444872.jpg?x-oss-process=image/resize,s_600/quality,q_80/auto-orient,0/interlace,1/format,jpg"
            alt=""
          />
          <img
            class="gallery-img"
            src="https://upload-bbs.miyoushe.com/upload/2022/11/28/17949827/b5cc5bf0ded8b38e961ab8b077f1b1e3_2998161025089190383.jpg"
            alt=""
          />
          <img
            class="gallery-img"
            src="https://upload-bbs.miyoushe.com/upload/2022/11/28/17949827/0c50bf57fdf196423eeeff3c23a70b78_3217896051162954880.jpg?x-oss-process=image//resize,s_600/quality,q_80/auto-orient,0/interlace,1/format,jpg"
            alt=""
          />
          <img
            class="gallery-img"
            src="https://upload-bbs.miyoushe.com/upload/2022/11/28/17949827/458f4678e4481f5e79bd25690adf4be1_6025800658609407775.jpg?x-oss-process=image//resize,s_600/quality,q_80/auto-orient,0/interlace,1/format,jpg"
            alt=""
          />
        </div>
      </a-col>
    </a-row>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, computed } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'

const cronTip = [
  'Cron 表达式由 6 个必填字段和 1 个可选字段组成，格式：秒 分 时 日 月 周 年（年份可省略）。',
  '各字段范围：秒/分 0-59；时 0-23；日 1-31（注意大小月）；月 1-12 或英文缩写；周 0-7（0=周日，6=周六，可用 ? 表示不指定）；年 1970-2099（可选）。',
  '在线生成器：https://cron.ciding.cc/'
].join('\\n')

const formState = reactive({
  id: 0,
  DBID: 0,
  name: '',
  spec: '',
  data: ''
})

const presetSpecs = [
  { label: '每天 4:05', spec: '0 5 4 * * *' },
  { label: '每周一 4:00', spec: '0 0 4 ? * MON' },
  { label: '每天 23:30', spec: '0 30 23 * * *' },
  { label: '除周一外每天 4:00', spec: '0 0 4 ? * TUE,WED,THU,FRI,SAT,SUN' }
]

const taskCronList = ref([])
const availableTaskNames = ref([])
const tableLoading = ref(false)
const formLoading = ref(false)
const dropdownLoading = ref(false)
const editingTaskId = ref(null)

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '任务名称', dataIndex: 'name', key: 'name', width: 150 },
  { title: 'Cron 表达式', dataIndex: 'spec', key: 'spec', width: 180 },
  { title: '下次执行时间', dataIndex: 'next', key: 'next', width: 200 },
  { title: '任务参数', dataIndex: 'data', key: 'data' },
  { title: '状态', dataIndex: 'status', key: 'status', width: 100 },
  { title: '操作', key: 'action', width: 220 }
]

const submitDisabled = computed(() => {
  return !formState.name || !formState.spec.trim()
})

const isEditing = computed(() => editingTaskId.value !== null)

const fetchTaskList = async () => {
  tableLoading.value = true
  try {
    const list = await apiMethods.getTaskCronList()
    taskCronList.value = Array.isArray(list)
      ? list.map(normalizeTaskCron)
      : []
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

const comeBack = () => {
  window.location.href = '/'
}

const handleSubmitTask = async () => {
  if (submitDisabled.value) {
    message.warning('请选择任务并填写 Cron 表达式')
    return
  }

  const editing = isEditing.value
  formLoading.value = true
  try {
    const payload = {
      id: formState.id,
      DBID: formState.DBID,
      name: formState.name,
      spec: formState.spec.trim(),
      data: formState.data?.trim() || ''
    }
    let res
    if (editing) {
      res = await apiMethods.updateTaskCron({
        id: editingTaskId.value,
        ...payload
      })
    } else {
      res = await apiMethods.addTaskCron(payload)
    }
    const msg = typeof res === 'string'
      ? res
      : editing
        ? '任务已更新'
        : '任务已添加'
    message.success(msg)
    resetForm()
    fetchTaskList()
  } catch (error) {
    message.error(editing ? '修改任务失败' : '添加任务失败')
  } finally {
    formLoading.value = false
  }
}

const resetForm = () => {
  editingTaskId.value = null
  formState.id = 0
  formState.DBID = 0
  formState.name = ''
  formState.spec = ''
  formState.data = ''
}

const startEdit = (record) => {
  editingTaskId.value = record.id
  formState.id = record.id
  formState.DBID = record.DBID
  formState.name = record.name
  formState.spec = record.spec
  formState.data = record.data || ''
}

const applyPreset = (spec) => {
  formState.spec = spec
}

const confirmRemove = (record) => {
  Modal.confirm({
    title: `确认删除任务「${record.name}」？`,
    content: '删除后需要重新创建才能恢复。',
    okText: '确定',
    cancelText: '再想想',
    okButtonProps: { danger: true },
    onOk: () => removeTask(record.id, record.DBID || record.dbid)
  })
}

const removeTask = async (id, dbid) => {
  try {
    const res = await apiMethods.removeTaskCron(id, dbid)
    const msg = typeof res === 'string' ? res : '任务已删除'
    message.success(msg)
    fetchTaskList()
  } catch (error) {
    message.error('删除任务失败')
  }
}

const togglePause = async (record) => {
  const paused = Boolean(record?.paused)
  const dbid = record?.DBID ?? record?.dbid
  if (!dbid) {
    message.error('未找到任务 dbid，无法执行操作')
    return
  }
  try {
    let res
    if (paused) {
      res = await apiMethods.resumeTaskCron(dbid)
    } else {
      res = await apiMethods.pauseTaskCron(dbid)
    }
    const defaultMsg = paused ? '任务已恢复' : '任务已暂停'
    const msg = typeof res === 'object' && res !== null && res.msg
      ? res.msg
      : typeof res === 'string'
        ? res
        : defaultMsg
    message.success(msg)
    if (paused && res?.id && editingTaskId.value === record.id) {
      editingTaskId.value = res.id
    } else if (!paused && editingTaskId.value === record.id) {
      resetForm()
    }
    await fetchTaskList()
  } catch (error) {
    message.error(paused ? '恢复任务失败' : '暂停任务失败')
  }
}

const normalizeTaskCron = (item) => {
  const id = Number(item?.id) || 0
  const dbid = item?.DBID ?? item?.dbid
  const statusNum = Number(item?.status)
  const paused = !(statusNum === 1 && id > 0)
  return {
    ...item,
    id,
    DBID: dbid,
    dbid,
    status: statusNum,
    paused
  }
}

const getRowKey = (record) => record?.id || record?.DBID || record?.dbid || record?.name

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

.content-row {
  align-items: stretch;
}

.form-col,
.table-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-card {
  flex: 1;
}

.table-card {
  flex: 1;
}

.cron-tip {
  margin-bottom: 16px;
  white-space: pre-line;
}

.form-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.form-actions .ant-btn {
  min-width: 110px;
}

.ghost-button {
  margin-left: 8px;
}

.back-button {
  margin-left: 8px;
  background-color: aquamarine;
  border-color: aquamarine;
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

.table-wrapper {
  overflow-x: auto;
}

.table-wrapper :deep(.ant-table) {
  min-width: 820px;
}

.visual-gallery {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.gallery-img {
  width: 100%;
  height: 190px;
  border-radius: 12px;
  object-fit: cover;
  box-shadow: 0 6px 12px rgba(0, 0, 0, 0.08);
}

@media (max-width: 768px) {
  .task-cron-page {
    padding: 16px;
  }

  .form-actions {
    flex-direction: column;
    align-items: stretch;
  }

  .form-actions .ant-btn {
    width: 100%;
  }

  .table-wrapper :deep(.ant-table) {
    min-width: 700px;
  }

  .gallery-img {
    height: 160px;
  }
}

@media (min-width: 992px) {
  .visual-gallery {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .gallery-img {
    height: 220px;
  }
}
</style>
