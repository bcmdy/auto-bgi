<template>
  <div class="calculate-container">
    <a-card title="任务计算器" class="calculate-card">
      <a-form :model="taskForm" layout="vertical" @finish="calculateTasks">
        <!-- 基础参数 -->
        <a-card title="基础参数" size="small" style="margin-bottom: 16px">
          <a-row :gutter="16">
            <a-col :span="8">
              <a-form-item label="当前原神等级" name="arLevel">
                <a-input-number 
                  v-model:value="taskForm.arLevel" 
                  :min="1" 
                  :max="60"
                  placeholder="冒险等级"
                  style="width: 100%"
                />
              </a-form-item>
            </a-col>
            <a-col :span="8">
              <a-form-item label="可用体力" name="resin">
                <a-input-number 
                  v-model:value="taskForm.resin" 
                  :min="0" 
                  :max="200"
                  placeholder="当前体力"
                  style="width: 100%"
                />
              </a-form-item>
            </a-col>
            <a-col :span="8">
              <a-form-item label="每日预算时间(分钟)" name="dailyTime">
                <a-input-number 
                  v-model:value="taskForm.dailyTime" 
                  :min="30" 
                  :max="480"
                  placeholder="游戏时间"
                  style="width: 100%"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </a-card>

        <!-- 任务优先级 -->
        <a-card title="任务优先级设置" size="small" style="margin-bottom: 16px">
          <a-table 
            :columns="priorityColumns" 
            :data-source="taskPriorities"
            :pagination="false"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'priority'">
                <a-select v-model:value="record.priority" style="width: 100%">
                  <a-select-option :value="1">高</a-select-option>
                  <a-select-option :value="2">中</a-select-option>
                  <a-select-option :value="3">低</a-select-option>
                </a-select>
              </template>
              <template v-if="column.key === 'enabled'">
                <a-switch v-model:checked="record.enabled" />
              </template>
              <template v-if="column.key === 'time'">
                <a-input-number 
                  v-model:value="record.time" 
                  :min="1" 
                  :max="60"
                  size="small"
                  style="width: 100%"
                />
              </template>
            </template>
          </a-table>
        </a-card>

        <!-- 计算按钮 -->
        <a-form-item>
          <a-button type="primary" html-type="submit" :loading="calculating" size="large">
            计算最优任务列表
          </a-button>
        </a-form-item>
      </a-form>

      <!-- 计算结果 -->
      <a-divider v-if="calculationResult" />
      
      <a-card v-if="calculationResult" title="计算结果" size="small">
        <!-- 推荐任务列表 -->
        <h4>推荐任务顺序：</h4>
        <a-list 
          :data-source="calculationResult.recommendedTasks" 
          size="small"
          :split="false"
        >
          <template #renderItem="{ item, index }">
            <a-list-item>
              <a-list-item-meta>
                <template #title>
                  <a-tag :color="getPriorityColor(item.priority)">{{ index + 1 }}</a-tag>
                  {{ item.name }}
                </template>
                <template #description>
                  预计时间: {{ item.time }}分钟 | 体力消耗: {{ item.resinCost }}
                </template>
              </a-list-item-meta>
            </a-list-item>
          </template>
        </a-list>

        <!-- 统计信息 -->
        <a-divider />
        <a-row :gutter="16">
          <a-col :span="6">
            <a-statistic 
              title="总用时" 
              :value="calculationResult.totalTime" 
              suffix="分钟"
              :value-style="{ color: '#ff5599' }"
            />
          </a-col>
          <a-col :span="6">
            <a-statistic 
              title="体力消耗" 
              :value="calculationResult.totalResin" 
              :value-style="{ color: '#ff5599' }"
            />
          </a-col>
          <a-col :span="6">
            <a-statistic 
              title="剩余体力" 
              :value="calculationResult.remainingResin" 
              :value-style="{ color: '#ff5599' }"
            />
          </a-col>
          <a-col :span="6">
            <a-statistic 
              title="效率评分" 
              :value="calculationResult.efficiency" 
              suffix="%"
              :value-style="{ color: '#ff5599' }"
            />
          </a-col>
        </a-row>

        <!-- 导出功能 -->
        <a-divider />
        <a-space>
          <a-button @click="exportResult">导出结果</a-button>
          <a-button @click="saveAsTemplate">保存为模板</a-button>
        </a-space>
      </a-card>
    </a-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { message } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'

const calculating = ref(false)
const calculationResult = ref(null)

const taskForm = reactive({
  arLevel: 45,
  resin: 160,
  dailyTime: 120
})

const priorityColumns = [
  {
    title: '任务类型',
    dataIndex: 'name',
    key: 'name'
  },
  {
    title: '启用',
    key: 'enabled',
    width: 80,
    align: 'center'
  },
  {
    title: '优先级',
    key: 'priority',
    width: 100
  },
  {
    title: '预计时间(分钟)',
    key: 'time',
    width: 140
  },
  {
    title: '体力消耗',
    dataIndex: 'resinCost',
    key: 'resinCost',
    width: 100
  }
]

const taskPriorities = reactive([
  {
    key: '1',
    name: '日常委托',
    enabled: true,
    priority: 1,
    time: 15,
    resinCost: 0
  },
  {
    key: '2',
    name: '天赋本',
    enabled: true,
    priority: 1,
    time: 5,
    resinCost: 20
  },
  {
    key: '3',
    name: '武器本',
    enabled: true,
    priority: 2,
    time: 5,
    resinCost: 20
  },
  {
    key: '4',
    name: '圣遗物',
    enabled: true,
    priority: 2,
    time: 5,
    resinCost: 20
  },
  {
    key: '5',
    name: '世界BOSS',
    enabled: true,
    priority: 1,
    time: 10,
    resinCost: 40
  },
  {
    key: '6',
    name: '周本BOSS',
    enabled: true,
    priority: 1,
    time: 15,
    resinCost: 30
  },
  {
    key: '7',
    name: '采集材料',
    enabled: false,
    priority: 3,
    time: 30,
    resinCost: 0
  }
])

const getPriorityColor = (priority) => {
  switch (priority) {
    case 1: return 'red'
    case 2: return 'orange'
    case 3: return 'blue'
    default: return 'default'
  }
}

const calculateTasks = async () => {
  calculating.value = true
  
  try {
    // 模拟计算逻辑
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    const enabledTasks = taskPriorities.filter(task => task.enabled)
    const sortedTasks = enabledTasks.sort((a, b) => a.priority - b.priority)
    
    let totalTime = 0
    let totalResin = 0
    const recommendedTasks = []
    
    for (const task of sortedTasks) {
      if (totalTime + task.time <= taskForm.dailyTime && totalResin + task.resinCost <= taskForm.resin) {
        recommendedTasks.push(task)
        totalTime += task.time
        totalResin += task.resinCost
      }
    }
    
    calculationResult.value = {
      recommendedTasks,
      totalTime,
      totalResin,
      remainingResin: taskForm.resin - totalResin,
      efficiency: Math.round((totalTime / taskForm.dailyTime) * 100)
    }
    
    message.success('计算完成')
  } catch (error) {
    message.error('计算失败: ' + error.message)
  } finally {
    calculating.value = false
  }
}

const exportResult = () => {
  if (!calculationResult.value) return
  
  const data = JSON.stringify(calculationResult.value, null, 2)
  const blob = new Blob([data], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'task-calculation-result.json'
  a.click()
  URL.revokeObjectURL(url)
  
  message.success('结果已导出')
}

const saveAsTemplate = () => {
  const template = {
    taskForm,
    taskPriorities: taskPriorities.map(t => ({ ...t }))
  }
  
  localStorage.setItem('task-template', JSON.stringify(template))
  message.success('模板已保存')
}
</script>

<style scoped>
.calculate-container {
  padding: 20px;
  background: #ffeaf4;
  min-height: 100vh;
}

.calculate-card {
  border-radius: 16px !important;
  box-shadow: 0 4px 12px rgba(255, 105, 180, 0.1) !important;
}

:deep(.ant-card-head-title) {
  color: #ff5599;
  font-size: 18px;
  font-weight: bold;
}

:deep(.ant-table-thead > tr > th) {
  background: #fff0f7;
  color: #ff5599;
  font-weight: bold;
}

:deep(.ant-btn-primary) {
  background: #ff99cc;
  border-color: #ff99cc;
}

:deep(.ant-btn-primary:hover) {
  background: #ff6699 !important;
  border-color: #ff6699 !important;
}

:deep(.ant-statistic-title) {
  color: #ff5599;
}

:deep(.ant-switch-checked) {
  background-color: #ff99cc;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .calculate-page {
    padding: 15px;
  }
  
  .page-title {
    font-size: 2rem;
  }
  
  .page-subtitle {
    font-size: 1rem;
  }
  
  /* Ant Design 组件移动端优化 */
  :deep(.ant-card) {
    margin-bottom: 12px;
  }
  
  :deep(.ant-card-head-title) {
    font-size: 1rem;
  }
  
  :deep(.ant-table) {
    font-size: 0.9rem;
  }
  
  :deep(.ant-table-thead > tr > th) {
    padding: 8px 4px;
    font-size: 0.85rem;
  }
  
  :deep(.ant-table-tbody > tr > td) {
    padding: 6px 4px;
    font-size: 0.85rem;
  }
  
  :deep(.ant-table-cell) {
    word-break: break-word;
    white-space: normal;
  }
  
  /* 隐藏不必要的列 */
  :deep(.ant-table-thead > tr > th:nth-child(4),
         .ant-table-tbody > tr > td:nth-child(4)) {
    display: none;
  }
  
  /* 调整列宽 */
  :deep(.ant-table-thead > tr > th:first-child,
         .ant-table-tbody > tr > td:first-child) {
    width: 80px;
    min-width: 80px;
  }
  
  :deep(.ant-table-thead > tr > th:last-child,
         .ant-table-tbody > tr > td:last-child) {
    width: 60px;
    min-width: 60px;
  }
  
  /* 表单组件优化 */
  :deep(.ant-form-item) {
    margin-bottom: 12px;
  }
  
  :deep(.ant-form-item-label) {
    font-size: 0.9rem;
  }
  
  :deep(.ant-input-number) {
    width: 100%;
  }
  
  :deep(.ant-select) {
    width: 100%;
  }
  
  /* 按钮优化 */
  :deep(.ant-btn) {
    width: 100%;
    height: 44px;
    font-size: 1rem;
  }
  
  /* 列表优化 */
  :deep(.ant-list-item) {
    padding: 8px 0;
  }
  
  :deep(.ant-list-item-meta-title) {
    font-size: 0.9rem;
  }
  
  :deep(.ant-list-item-meta-description) {
    font-size: 0.8rem;
  }
  
  /* 统计信息优化 */
  :deep(.ant-statistic-title) {
    font-size: 0.8rem;
  }
  
  :deep(.ant-statistic-content) {
    font-size: 1rem;
  }
  
  /* 标签优化 */
  :deep(.ant-tag) {
    font-size: 0.75rem;
    padding: 2px 6px;
  }
}

@media (max-width: 480px) {
  .calculate-page {
    padding: 10px;
  }
  
  .page-title {
    font-size: 1.8rem;
  }
  
  .page-subtitle {
    font-size: 0.9rem;
  }
  
  /* 进一步优化表格 */
  :deep(.ant-table) {
    font-size: 0.8rem;
  }
  
  :deep(.ant-table-thead > tr > th) {
    padding: 6px 2px;
    font-size: 0.8rem;
  }
  
  :deep(.ant-table-tbody > tr > td) {
    padding: 4px 2px;
    font-size: 0.8rem;
  }
  
  /* 隐藏更多列 */
  :deep(.ant-table-thead > tr > th:nth-child(3),
         .ant-table-tbody > tr > td:nth-child(3)) {
    display: none;
  }
  
  /* 进一步调整列宽 */
  :deep(.ant-table-thead > tr > th:first-child,
         .ant-table-tbody > tr > td:first-child) {
    width: 60px;
    min-width: 60px;
  }
  
  :deep(.ant-table-thead > tr > th:last-child,
         .ant-table-tbody > tr > td:last-child) {
    width: 50px;
    min-width: 50px;
  }
  
  /* 表单组件进一步优化 */
  :deep(.ant-form-item-label) {
    font-size: 0.85rem;
  }
  
  /* 按钮进一步优化 */
  :deep(.ant-btn) {
    height: 40px;
    font-size: 0.9rem;
  }
  
  /* 列表进一步优化 */
  :deep(.ant-list-item-meta-title) {
    font-size: 0.85rem;
  }
  
  :deep(.ant-list-item-meta-description) {
    font-size: 0.75rem;
  }
  
  /* 统计信息进一步优化 */
  :deep(.ant-statistic-title) {
    font-size: 0.75rem;
  }
  
  :deep(.ant-statistic-content) {
    font-size: 0.9rem;
  }
  
  /* 标签进一步优化 */
  :deep(.ant-tag) {
    font-size: 0.7rem;
    padding: 1px 4px;
  }
}

/* 横屏模式优化 */
@media (max-width: 768px) and (orientation: landscape) {
  .calculate-page {
    padding: 10px;
  }
  
  .page-title {
    font-size: 1.8rem;
  }
  
  .page-subtitle {
    font-size: 0.9rem;
  }
  
  /* 横屏时显示更多列 */
  :deep(.ant-table-thead > tr > th:nth-child(3),
         .ant-table-tbody > tr > td:nth-child(3)) {
    display: table-cell;
  }
  
  :deep(.ant-table-thead > tr > th:nth-child(4),
         .ant-table-tbody > tr > td:nth-child(4)) {
    display: table-cell;
  }
}

/* 触摸优化 */
@media (pointer: coarse) {
  :deep(.ant-table-thead > tr > th) {
    min-height: 44px;
  }
  
  :deep(.ant-table-tbody > tr > td) {
    min-height: 44px;
  }
  
  :deep(.ant-tag) {
    min-height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  :deep(.ant-switch) {
    min-width: 44px;
    min-height: 24px;
  }
  
  :deep(.ant-select-selector) {
    min-height: 44px;
  }
  
  :deep(.ant-input-number-input) {
    min-height: 44px;
  }
}
</style>