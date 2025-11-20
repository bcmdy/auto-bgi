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
            message="Cron表达式由6个必填字段和1个可选字段组成，格式为：秒 分 时 日 月 周 年（年份可省略）
            各字段说明：秒、分：0-59,   
            支持字符 * , - /小时：0-23，    
            规则同上日期：1-31（需注意月份天数）
               月份：1-12 或月份缩写
            星期：1-7（1=周日，7=周六），可用?表示不指定
            年份：1970-2099（可选）。在线生成器：https://cron.ciding.cc/"
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
              <a-button type="primary" :loading="formLoading" :disabled="submitDisabled" @click="handleSubmitTask">
                {{ isEditing ? '保存修改' : '添加任务' }}
              </a-button>
              <a-button style="margin-left: 8px" @click="resetForm">
                {{ isEditing ? '取消编辑' : '重置' }}
              </a-button>

              <a-button style="margin-left: 10px;background-color: aquamarine;" @click="comeBack">
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

      <a-col :xs="24" :md="14">
        <a-card title="已配置任务" bordered>
          <a-spin :spinning="tableLoading">
            <div v-if="taskCronList.length > 0">
              <a-table
                :data-source="taskCronList"
                :columns="columns"
                :row-key="getRowKey"
                :pagination="false"
                size="middle"
                bordered
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
<div style="width: 115vh;  position:absolute; display: flex; flex-wrap: wrap; overflow: auto;">
  <img style="width: 45%; height: 345px; border-radius: 10px; margin: 5px; object-fit: cover;" 
       src="https://upload-bbs.miyoushe.com/upload/2025/09/30/162891450/40ba73d1c8ce78112681a7ed7137dad1_6952277989857444872.jpg?x-oss-process=image/resize,s_600/quality,q_80/auto-orient,0/interlace,1/format,jpg" alt="">
  <img style="width: 45%; height: 345px; border-radius: 10px; margin: 5px; object-fit: cover;" 
       src="https://upload-bbs.miyoushe.com/upload/2022/11/28/17949827/b5cc5bf0ded8b38e961ab8b077f1b1e3_2998161025089190383.jpg" alt="">
  <img style="width: 45%; height: 345px; border-radius: 10px; margin: 5px; object-fit: cover;" 
       src="https://upload-bbs.miyoushe.com/upload/2022/11/28/17949827/0c50bf57fdf196423eeeff3c23a70b78_3217896051162954880.jpg?x-oss-process=image//resize,s_600/quality,q_80/auto-orient,0/interlace,1/format,jpg" alt="">
  <img style="width: 45%; height: 345px; border-radius: 10px; margin: 5px; object-fit: cover;" 
       src="https://upload-bbs.miyoushe.com/upload/2022/11/28/17949827/458f4678e4481f5e79bd25690adf4be1_6025800658609407775.jpg?x-oss-process=image//resize,s_600/quality,q_80/auto-orient,0/interlace,1/format,jpg" alt="">
</div>
      </a-col>
    </a-row>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, computed } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'

const formState = reactive({
  id: 0,
  DBID: 0,
  name: '',
  spec: '',
  data: ''
})

const presetSpecs = [
  { label: '每天4点5分', spec: '0 5 4 * * *' },
  { label: '每周一四点运行', spec: '0 0 4 ? * MON' },
  { label: '每天23点30分', spec: '0 30 23 * * *' },
    { label: '除周一外其他天四点运行', spec: '0 0 4 ? * TUE,WED,THU,FRI,SAT,SUN' }
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
  { title: '操作', key: 'action', width: 200 }
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

async function comeBack() {
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
      id:formState.id,
      DBID:formState.DBID,
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
    title: `确认删除任务「${record.name}」?`,
    content: '删除后需要重新创建才能恢复。',
    okText: '确定',
    cancelText: '再想想',
    okButtonProps: { danger: true },
    onOk: () => removeTask(record.id, record.dbid)
  })
}

const removeTask = async (id,dbid) => {
  try {
    const res = await apiMethods.removeTaskCron(id,dbid)
    const msg = typeof res === "string" ? res : "任务已删除"
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
