<template>
  <div class="harvest-container">
    <a-card title="收获统计" class="harvest-card">
      <template #extra>
        <a-space>
          <a-select v-model:value="selectedPeriod" @change="onPeriodChange" style="width: 120px">
            <a-select-option value="today">今日</a-select-option>
            <a-select-option value="week">本周</a-select-option>
            <a-select-option value="month">本月</a-select-option>
          </a-select>
          <a-button @click="refreshData" :loading="loading">刷新</a-button>
        </a-space>
      </template>

      <a-table 
        :columns="columns" 
        :data-source="harvestData" 
        :loading="loading"
        :pagination="false"
        size="middle"
      >
        <template #bodyCell="{ column, record, index }">
          <template v-if="column.key === 'rank'">
            <a-tag :color="getRankColor(index + 1)">
              第{{ index + 1 }}名
            </a-tag>
          </template>
          <template v-if="column.key === 'count'">
            <span class="count-text">{{ record.count }}</span>
          </template>
          <template v-if="column.key === 'icon'">
            <div class="item-icon">
              <img v-if="record.icon" :src="record.icon" :alt="record.name" />
              <span v-else>📦</span>
            </div>
          </template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'

const loading = ref(false)
const selectedPeriod = ref('today')
const harvestData = ref([])

const columns = [
  {
    title: '排名',
    key: 'rank',
    width: 100,
    align: 'center'
  },
  {
    title: '图标',
    key: 'icon',
    width: 60,
    align: 'center'
  },
  {
    title: '物品名称',
    dataIndex: 'name',
    key: 'name'
  },
  {
    title: '收获数量',
    key: 'count',
    width: 120,
    align: 'right',
    sorter: (a, b) => a.count - b.count
  },
  {
    title: '类型',
    dataIndex: 'type',
    key: 'type',
    width: 100
  }
]

const getRankColor = (rank) => {
  switch (rank) {
    case 1: return 'gold'
    case 2: return 'orange'
    case 3: return 'red'
    case 4:
    case 5: return 'purple'
    default: return 'blue'
  }
}

const refreshData = async () => {
  loading.value = true
  try {
    const response = await apiMethods.getHarvest()
    harvestData.value = response.data || []
  } catch (error) {
    message.error('获取收获数据失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const onPeriodChange = (period) => {
  selectedPeriod.value = period
  refreshData()
}

onMounted(() => {
  refreshData()
})
</script>

<style scoped>
.harvest-page {
  background: linear-gradient(135deg, #fff6fb 0%, #ffe6f2 50%, #ffd6eb 100%);
  min-height: 100vh;
  padding: 20px;
  font-family: "Comic Sans MS", "Segoe UI", sans-serif;
}

.header {
  text-align: center;
  margin-bottom: 30px;
  color: #ff6eb4;
}

.title {
  font-size: 2.5rem;
  margin-bottom: 10px;
  text-shadow: 0 0 20px rgba(255, 110, 180, 0.3);
}

.subtitle {
  font-size: 1.1rem;
  opacity: 0.8;
}

.controls {
  margin-bottom: 20px;
  display: flex;
  justify-content: center;
  gap: 15px;
}

.count-text {
  font-weight: bold;
  color: #ff6eb4;
}

.item-icon {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 40px;
  height: 40px;
}

.item-icon img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.item-icon span {
  font-size: 1.5rem;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .harvest-page {
    padding: 15px;
  }
  
  .title {
    font-size: 2rem;
  }
  
  .subtitle {
    font-size: 1rem;
  }
  
  .controls {
    flex-direction: column;
    align-items: center;
    gap: 10px;
  }
  
  /* Ant Design 表格移动端优化 */
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
  :deep(.ant-table-thead > tr > th:nth-child(2),
         .ant-table-tbody > tr > td:nth-child(2)) {
    display: none;
  }
  
  /* 调整列宽 */
  :deep(.ant-table-thead > tr > th:first-child,
         .ant-table-tbody > tr > td:first-child) {
    width: 60px;
    min-width: 60px;
  }
  
  :deep(.ant-table-thead > tr > th:last-child,
         .ant-table-tbody > tr > td:last-child) {
    width: 80px;
    min-width: 80px;
  }
  
  /* 排名标签优化 */
  :deep(.ant-tag) {
    font-size: 0.75rem;
    padding: 2px 6px;
  }
  
  /* 数量文本优化 */
  .count-text {
    font-size: 0.85rem;
  }
  
  /* 图标优化 */
  .item-icon {
    width: 30px;
    height: 30px;
  }
  
  .item-icon span {
    font-size: 1.2rem;
  }
}

@media (max-width: 480px) {
  .harvest-page {
    padding: 10px;
  }
  
  .title {
    font-size: 1.8rem;
  }
  
  .subtitle {
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
  
  /* 隐藏类型列 */
  :deep(.ant-table-thead > tr > th:nth-child(5),
         .ant-table-tbody > tr > td:nth-child(5)) {
    display: none;
  }
  
  /* 进一步调整列宽 */
  :deep(.ant-table-thead > tr > th:first-child,
         .ant-table-tbody > tr > td:first-child) {
    width: 50px;
    min-width: 50px;
  }
  
  :deep(.ant-table-thead > tr > th:last-child,
         .ant-table-tbody > tr > td:last-child) {
    width: 60px;
    min-width: 60px;
  }
  
  /* 排名标签进一步优化 */
  :deep(.ant-tag) {
    font-size: 0.7rem;
    padding: 1px 4px;
  }
  
  /* 数量文本进一步优化 */
  .count-text {
    font-size: 0.8rem;
  }
  
  /* 图标进一步优化 */
  .item-icon {
    width: 25px;
    height: 25px;
  }
  
  .item-icon span {
    font-size: 1rem;
  }
}

/* 横屏模式优化 */
@media (max-width: 768px) and (orientation: landscape) {
  .harvest-page {
    padding: 10px;
  }
  
  .title {
    font-size: 1.8rem;
  }
  
  .subtitle {
    font-size: 0.9rem;
  }
  
  /* 横屏时显示更多列 */
  :deep(.ant-table-thead > tr > th:nth-child(2),
         .ant-table-tbody > tr > td:nth-child(2)) {
    display: table-cell;
  }
  
  :deep(.ant-table-thead > tr > th:nth-child(5),
         .ant-table-tbody > tr > td:nth-child(5)) {
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
}
</style>