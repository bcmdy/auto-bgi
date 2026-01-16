<template>
  <div class="task-cron-page">
    <div class="bg-circle circle-1"></div>
    <div class="bg-circle circle-2"></div>

    <div class="main-container">
      <a-typography-title :level="2" class="page-title">
        <span>✨ 定时任务管理 ✨</span>
      </a-typography-title>

      <a-row :gutter="[24, 24]" class="content-row">
        <a-col :xs="24" :lg="8" :xl="6" class="form-col">
          <div class="sticky-wrapper">
            <a-card :bordered="false" class="pink-card form-card hover-effect">
              <template #title>
                <div class="card-header-title">
                  <plus-circle-outlined />
                  <span>{{ isEditing ? '编辑任务' : '新增任务' }}</span>
                </div>
              </template>
              
              <a-alert
                type="info"
                show-icon
                class="cron-tip"
              >
                <template #message>
                  <span style="font-size: 12px; color: #666;">
                    Cron: 秒 分 时 日 月 周 (年)<br/>
                    示例: */5 * * * * (每5分)
                  </span>
                </template>
              </a-alert>

              <a-form layout="vertical" class="custom-form">
                <a-form-item label="任务类型">
                  <a-select
                    v-model:value="formState.name"
                    placeholder="选择后端注册任务"
                    :loading="dropdownLoading"
                    allow-clear
                    class="pink-input"
                    :disabled="isEditing"
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
                    placeholder="例如：0 0 12 * * ?"
                    allow-clear
                    class="pink-input"
                  >
                    <template #suffix>
                      <a-tooltip title="点击查看在线生成器">
                        <question-circle-outlined style="color: #ff85c0; cursor: pointer" @click="openCronTool"/>
                      </a-tooltip>
                    </template>
                  </a-input>
                  <div class="preset-tags">
                    <a-tag 
                      v-for="item in presetSpecs" 
                      :key="item.label" 
                      color="pink" 
                      @click="applyPreset(item.spec)"
                    >
                      {{ item.label }}
                    </a-tag>
                  </div>
                </a-form-item>

                <a-form-item label="任务参数 (可选)">
                  <a-textarea
                    v-model:value="formState.data"
                    :auto-size="{ minRows: 3, maxRows: 5 }"
                    placeholder="配置组名称等..."
                    allow-clear
                    class="pink-input"
                  />
                </a-form-item>

                <a-form-item label="备注 (可选)">
                  <a-textarea
                    v-model:value="formState.mark"
                    :auto-size="{ minRows: 2, maxRows: 4 }"
                    placeholder="任务备注信息..."
                    allow-clear
                    class="pink-input"
                  />
                </a-form-item>

                <div class="form-actions">
                  <a-button 
                    type="primary" 
                    class="pink-btn submit-btn" 
                    :loading="formLoading"
                    @click="handleSubmitTask"
                    :disabled="submitDisabled"
                  >
                    {{ isEditing ? '保存修改' : '立即添加' }}
                  </a-button>
                  
                  <a-button 
                    v-if="isEditing" 
                    class="pink-btn ghost-btn" 
                    @click="resetForm"
                  >
                    取消
                  </a-button>

                  <a-button class="pink-btn text-btn" type="text" @click="comeBack">
                    返回首页
                  </a-button>
                </div>
              </a-form>
            </a-card>
          </div>
        </a-col>

        <a-col :xs="24" :lg="16" :xl="18" class="list-col">
          <a-spin :spinning="tableLoading">
            <template v-if="taskCronList.length > 0">
              <a-list
                :grid="{ gutter: 16, xs: 1, sm: 1, md: 2, lg: 3, xl: 3, xxl: 3 }"
                :data-source="taskCronList"
              >
                <template #renderItem="{ item }">
                  <a-list-item class="full-height-item">
                    <a-card :bordered="false" class="task-item-card hover-effect">
                      
                      <div class="task-header">
                        <div class="task-name-wrapper">
                          <span class="task-name" :title="item.name">{{ item.name }}</span>
                        </div>
                        <a-tag :color="item.paused ? 'orange' : '#87d068'" class="status-tag">
                          {{ item.paused ? '已暂停' : '运行中' }}
                        </a-tag>
                      </div>

                      <div class="task-body">
                        <div class="info-block">
                          <div class="info-label">
                            <clock-circle-outlined /> Cron 表达式
                          </div>
                          <div class="cron-box" :title="item.spec">
                            {{ item.spec }}
                          </div>
                        </div>

                        <div class="info-row">
                          <field-time-outlined class="icon" />
                          <span class="sub-text">下次: {{ item.paused ? '-' : (item.next || '计算中...') }}</span>
                        </div>

                        <div class="info-row data-row" v-if="item.data">
                          <code-outlined class="icon" />
                          <span class="sub-text text-truncate" :title="item.data">{{ item.data }}</span>
                        </div>

                        <div class="info-row mark-row" v-if="item.mark">
                          <edit-outlined class="icon" />
                          <span class="sub-text" :title="item.mark">备注: {{ item.mark }}</span>
                        </div>
                      </div>

                      <div class="task-actions">
                        <a-tooltip title="立即运行一次">
                          <a-button 
                            shape="circle" size="small" 
                            class="action-btn run-btn"
                            @click="AtOnceRunTask(item.name, item.data)"
                          >
                            <caret-right-outlined />
                          </a-button>
                        </a-tooltip>

                        <a-tooltip title="编辑配置">
                          <a-button 
                            shape="circle" size="small" 
                            class="action-btn edit-btn"
                            @click="startEdit(item)"
                          >
                            <edit-outlined />
                          </a-button>
                        </a-tooltip>

                         <a-popconfirm
                                    :title="item.paused ? '恢复这个任务？' : '暂停这个任务？'"
                                    ok-text="确定"
                                    cancel-text="取消"
                                    @confirm="togglePause(item)"
                                  >
                                    <a-tooltip :title="item.paused ? '恢复' : '暂停'">
                                      <a-button 
                                        shape="circle" 
                                        size="small" 
                                        class="action-btn"
                                        :class="item.paused ? 'resume-btn' : 'pause-btn'"
                                      >
                                        <play-circle-outlined v-if="item.paused" />
                                        <pause-circle-outlined v-else />
                                        </a-button>
                                    </a-tooltip>
                                  </a-popconfirm>

                        <a-button 
                          shape="circle" size="small" 
                          class="action-btn delete-btn"
                          @click="confirmRemove(item)"
                        >
                          <delete-outlined />
                        </a-button>
                      </div>
                    </a-card>
                  </a-list-item>
                </template>
              </a-list>
            </template>
            
       
          </a-spin>
        </a-col>
      </a-row>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, computed } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { 
  PlusCircleOutlined, 
  QuestionCircleOutlined, 
  ClockCircleOutlined, 
  FieldTimeOutlined,
  CodeOutlined,
  CaretRightOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined
} from '@ant-design/icons-vue'
import { apiMethods } from '@/utils/api'

// --- 逻辑部分保持原样 ---

const formState = reactive({
  id: 0,
  entry_id: 0,
  name: '',
  spec: '',
  data: '',
  mark: '',
  status: 1
})

const presetSpecs = [
  { label: '每天 4:05', spec: '0 5 4 * * *' },
  { label: '每周一 4:00', spec: '0 0 4 ? * MON' },
  { label: '每天 23:30', spec: '0 30 23 * * *' },
  { label: '每5分钟', spec: '0 */5 * * * ?' }
]

const taskCronList = ref([])
const availableTaskNames = ref([])
const tableLoading = ref(false)
const formLoading = ref(false)
const dropdownLoading = ref(false)
const editingTaskId = ref(null)

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

const openCronTool = () => {
    window.open('https://cron.ciding.cc/', '_blank')
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
      entry_id: formState.entry_id,
      name: formState.name,
      spec: formState.spec.trim(),
      data: formState.data?.trim() || '',
      mark: formState.mark?.trim() || '',
      status: formState.status
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
  formState.entry_id = 0
  formState.name = ''
  formState.spec = ''
  formState.data = ''
  formState.mark = ''
  formState.status = 0
}

const startEdit = (record) => {
  editingTaskId.value = record.id
  formState.id = record.id
  formState.entry_id = record.entry_id
  formState.name = record.name
  formState.spec = record.spec
  formState.data = record.data || ''
  formState.mark = record.mark || ''
  formState.status = record.status || 0
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const applyPreset = (spec) => {
  formState.spec = spec
}

const confirmRemove = (record) => {
  Modal.confirm({
    title: `确认删除任务「${record.name}」？`,
    content: '删除后需要重新创建才能恢复。',
    okText: '狠狠删除',
    cancelText: '再想想',
    okButtonProps: { danger: true },
    centered: true,
    onOk: () => removeTask(record.id, record.entry_id)
  })
}


const AtOnceRunTask = (type, data) => {
    Modal.confirm({
        title: '确认立即运行？',
        content: '是否确认立即运行当前的任务？',
        okText: '确定',
        cancelText: '取消',
        centered: true, // 居中显示
        onOk: async () => {
            try { 
                  const res = await apiMethods.AtOnceRunTaskCron(type, data)
                  const msg = typeof res === 'string' ? res : '任务指令已发送'
                  message.success(msg)
                  fetchTaskList()
            } catch(e) { 
                message.error('执行任务失败',JSON.stringify(e)) 
            }
        }
    })
}

// const AtOnceRunTask = async (type, data) => {
//   try {
//     const res = await apiMethods.AtOnceRunTaskCron(type, data)
//     const msg = typeof res === 'string' ? res : '任务指令已发送'
//     message.success(msg)
//     fetchTaskList()
//   } catch (error) {
//     message.error('执行任务失败')
//   }
// }

const removeTask = async (id, entry_id) => {
  try {
    const res = await apiMethods.removeTaskCron(id, entry_id)
    const msg = typeof res === 'string' ? res : '任务已删除'
    message.success(msg)
    fetchTaskList()
  } catch (error) {
    message.error('删除任务失败')
  }
}

const togglePause = async (record) => {
  const paused = Boolean(record?.paused)
  const id = record?.id
  if (!id) {
    message.error('未找到任务 id，无法执行操作')
    return
  }
  try {
    let res
    if (paused) {
      res = await apiMethods.resumeTaskCron(id)
    } else {
      res = await apiMethods.pauseTaskCron(id)
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
  const entry_id = item?.entry_id
  const statusNum = Number(item?.status)
  const paused = !(statusNum === 1 && id > 0)
  return {
    ...item,
    id,
    entry_id,
    status: statusNum,
    paused
  }
}

onMounted(() => {
  fetchTaskList()
  fetchTaskNameOptions()
})
</script>

<style scoped>
/* 全局布局变量 */
:deep(*) {
  box-sizing: border-box;
}

.task-cron-page {
  position: relative;
  padding: 24px;
  background: #fff0f6; /* 浅粉色背景 */
  min-height: 100vh;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  overflow-x: hidden;
}

/* 装饰性背景球 */
.bg-circle {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  z-index: 0;
  opacity: 0.5;
}
.circle-1 {
  top: -60px;
  left: -60px;
  width: 300px;
  height: 300px;
  background: #ffadd2;
}
.circle-2 {
  bottom: 50px;
  right: -50px;
  width: 400px;
  height: 400px;
  background: #bae7ff;
}

.main-container {
  position: relative;
  z-index: 1;
  max-width: 1600px; /* 加宽容器适配三列 */
  margin: 0 auto;
}

.page-title {
  color: #eb2f96 !important;
  text-align: center;
  margin-bottom: 32px;
  font-weight: 700;
  text-shadow: 2px 2px 0px #fff;
  letter-spacing: 2px;
}

/* 卡片通用样式 */
.pink-card {
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.95);
  box-shadow: 0 8px 20px rgba(235, 47, 150, 0.08);
  border: 1px solid rgba(255, 240, 246, 0.8);
  transition: all 0.3s ease;
}

.hover-effect:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(235, 47, 150, 0.15);
}

.card-header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #ff4d4f;
  font-size: 16px;
  font-weight: 600;
}

/* 左侧表单样式 */
.sticky-wrapper {
  position: sticky;
  top: 24px;
}

.custom-form :deep(.ant-form-item-label > label) {
  color: #555;
  font-weight: 600;
}

.pink-input :deep(.ant-input),
.pink-input :deep(.ant-select-selector) {
  border-radius: 8px;
  border-color: #ffd6e7;
  background: #fff;
}

.pink-input :deep(.ant-input):focus,
.pink-input :deep(.ant-select-selector-focused) {
  border-color: #ffadd2;
  box-shadow: 0 0 0 2px rgba(255, 173, 210, 0.2);
}

.preset-tags {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.preset-tags :deep(.ant-tag) {
  cursor: pointer;
  border-radius: 6px;
  border: none;
  background: #fff0f6;
  color: #eb2f96;
  padding: 2px 10px;
  transition: all 0.2s;
}
.preset-tags :deep(.ant-tag):hover {
  background: #eb2f96;
  color: white;
}

/* 按钮样式 */
.pink-btn {
  border-radius: 20px;
  height: 38px;
  font-weight: 600;
}
.submit-btn {
  flex: 1;
  background: linear-gradient(135deg, #ff85c0 0%, #ff4d4f 100%);
  border: none;
  box-shadow: 0 4px 10px rgba(255, 77, 79, 0.3);
}
.submit-btn:hover {
  background: linear-gradient(135deg, #ffadd2 0%, #ff7875 100%);
  transform: scale(1.02);
}
.ghost-btn {
  border: 1px solid #ffadd2;
  color: #eb2f96;
}
.text-btn {
  color: #999;
}

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
  flex-wrap: wrap;
}

/* 右侧列表布局关键样式 */
.full-height-item {
  height: 100%;
}

.task-item-card {
  border-radius: 16px;
  background: rgb(182, 224, 221);
  border: 1px solid #fff0f6;
  /* 关键：使用 Flex 布局让卡片内容垂直分布，实现底部对齐 */
  display: flex;
  flex-direction: column;
  height: 100%; 
}

/* 覆盖 Ant Design Card body 默认 padding，并设为 flex */
.task-item-card :deep(.ant-card-body) {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 20px;
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.task-name-wrapper {
  flex: 1;
  margin-right: 12px;
}

.task-name {
  font-size: 16px;
  font-weight: 700;
  color: #333;
  word-break: break-all;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.status-tag {
  border-radius: 6px;
  border: none;
  padding: 0 8px;
  font-size: 12px;
  height: 22px;
  line-height: 22px;
}

/* 任务主体内容：占据剩余空间，将底部按钮推下去 */
.task-body {
  flex: 1;
  margin-bottom: 16px;
}

.info-block {
  background: #fafafa;
  border-radius: 8px;
  padding: 10px;
  margin-bottom: 12px;
  border: 1px solid #f0f0f0;
}

.info-label {
  font-size: 12px;
  color: #999;
  margin-bottom: 4px;
  display: flex;
  align-items: center;
  gap: 4px;
}

/* Cron 表达式美化与适配 */
.cron-box {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
  color: #eb2f96;
  font-weight: 600;
  font-size: 14px;
  /* 关键：允许在任意字符间换行 */
  word-break: break-all; 
  white-space: pre-wrap;
  line-height: 1.5;
}

.info-row {
  display: flex;
  align-items: center;
  margin-bottom: 6px;
  color: #160808;
  font-size: 13px;
}
.info-row:last-child {
  margin-bottom: 0;
}

.info-row .icon {
  margin-right: 8px;
  color: #ff85c0;
}

.sub-text {
  color: #0c0505;
  white-space: normal; /* 允许文本自动换行 */
  word-wrap: break-word; /* 在单词长且无法容纳时强制换行 */
  overflow: visible; /* 保证文字不被隐藏 */
}

.text-truncate {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
  display: inline-block;
  vertical-align: middle;
    white-space: normal; /* 允许自动换行 */
  overflow: visible; /* 让溢出的内容显示出来 */
}

/* 卡片底部按钮 */
.task-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px dashed #f0f0f0;
  margin-top: auto; /* 辅助 flex 布局 */
}

.action-btn {
  border: none;
  background: #f1e8ec;
  color: #eb2f96;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}
.action-btn:hover {
  background: #eb2f96;
  color: white;
  transform: scale(1.1);
}

.run-btn:hover { background: #52c41a; }
.pause-btn:hover { background: #faad14; }
.resume-btn:hover { background: #52c41a; }
.delete-btn:hover { background: #ff4d4f; }

.empty-state {
  text-align: center;
  padding: 60px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
}

/* 移动端适配 */
@media (max-width: 992px) {
  .list-col {
    margin-top: 24px;
  }
}

@media (max-width: 768px) {
  .task-cron-page {
    padding: 16px;
  }
  .form-actions {
    flex-direction: column;
  }
  .submit-btn {
    width: 100%;
  }
  .sticky-wrapper {
    position: static;
  }
}
</style>