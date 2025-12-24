<template>
  <div class="collection-management">
    <a-card title="采集管理" class="collection-card">
      <template #extra>
        <a-space>
          <a-select 
            v-model:value="selectedAccount" 
            @change="onAccountChange" 
            style="width: 150px"
            placeholder="选择账户"
          >
            <a-select-option v-for="account in accountList" :key="account" :value="account">
              {{ account }}
            </a-select-option>
          </a-select>
          <!-- <a-date-picker 
            v-model:value="selectedDate" 
            format="YYYY-MM-DD"
            @change="onDateChange"
            placeholder="选择日期"
            style="width: 160px"
          /> -->
          <a-select 
            v-model:value="selectedStatus" 
            @change="onStatusChange" 
            style="width: 120px"
            placeholder="选择状态"
          >
            <a-select-option value="">全部状态</a-select-option>
            <a-select-option value="可采集">可采集</a-select-option>
            <a-select-option value="CD中">CD中</a-select-option>
          </a-select>
          <a-button @click="refreshData" :loading="loading">刷新</a-button>
        </a-space>
      </template>

      <a-collapse v-model:activeKey="activeKeys" accordion>
        <a-collapse-panel v-for="(items, material) in filteredData" :key="material" :header="material">
          <template #extra>
            <a-tag color="blue">{{ items.length }} 个采集点</a-tag>
          </template>
          
          <!-- 桌面端表格显示 -->
          <a-table 
            :columns="columns" 
            :data-source="items" 
            :pagination="false"
            size="small"
            :row-key="(record) => record.FileName"
            class="desktop-table"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'FileName'">
                <div class="file-name">{{ record.FileName }}</div>
              </template>
              
              <template v-if="column.key === 'CdTime'">
                <div class="cd-time">{{ record.CdTime }}</div>
              </template>
              
              <template v-if="column.key === 'Status'">
                <a-tag :color="getStatusColor(record.Status)">
                  {{ record.Status }}
                </a-tag>
              </template>
              
              <template v-if="column.key === 'action'">
                <a-button 
                  type="link" 
                  size="small"
                  @click="showHistory(record)"
                >
                  查看历史
                </a-button>
              </template>
            </template>
          </a-table>
          
          <!-- 移动端卡片显示 -->
          <div class="mobile-cards">
            <div 
              v-for="(record, index) in items" 
              :key="index"
              class="collection-item-card"
            >
              <div class="card-header">
                <div class="card-title">{{ record.FileName }}</div>
                <a-tag :color="getStatusColor(record.Status)" class="card-status">
                  {{ record.Status }}
                </a-tag>
              </div>
              
              <div class="card-body">
                <div class="card-info-row">
                  <span class="info-label">📅 CD时间：</span>
                  <span class="info-value">{{ record.CdTime }}</span>
                </div>
              </div>
              
              <div class="card-footer">
                <a-button 
                  type="primary" 
                  size="small"
                  @click="showHistory(record)"
                  block
                >
                  📊 查看历史
                </a-button>
              </div>
            </div>
          </div>
        </a-collapse-panel>
      </a-collapse>
    </a-card>

    <!-- 拾取记录 -->
    <a-card title="拾取记录" class="collection-card" style="margin-top: 20px">
      <template #extra>
        <a-button @click="refreshPickupData" :loading="pickupLoading">刷新</a-button>
      </template>

      <div v-if="pickupData.length === 0" class="empty-state">
        <div class="empty-icon">📅</div>
        <div class="empty-text">暂无拾取记录</div>
      </div>

      <a-timeline v-else mode="left" class="pickup-timeline">
        <a-timeline-item 
          v-for="(record, index) in pickupData" 
          :key="index"
          :color="index === 0 ? 'green' : 'blue'"
        >
          <template #dot>
            <span class="timeline-dot">{{ index === 0 ? '🌟' : '📅' }}</span>
          </template>
          
          <div class="pickup-record">
            <div class="pickup-header">
              <span class="pickup-date">
                <span class="date-icon">📆</span>
                {{ record.Date }}
              </span>
              <a-tag :color="index === 0 ? 'green' : 'blue'">
                {{ index === 0 ? '最近' : formatDateDiff(record.Date) }}
              </a-tag>
            </div>
            
            <a-divider style="margin: 12px 0" />
            
            <div class="pickup-stats">
              <div class="stat-item">
                <span class="stat-label">采集种类：</span>
                <span class="stat-value">{{ Object.keys(record.Item).length }} 种</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">总采集量：</span>
                <span class="stat-value">{{ calculateDailyTotal(record.Item) }} 个</span>
              </div>
            </div>
            
            <div class="pickup-items">
              <a-tag 
                v-for="(count, itemName) in sortedItems(record.Item)" 
                :key="itemName"
                :color="getItemTagColor(itemName)"
                class="pickup-item-tag"
              >
                <span class="item-name">{{ itemName }}</span>
                <span class="item-count">× {{ count }}</span>
              </a-tag>
            </div>
          </div>
        </a-timeline-item>
      </a-timeline>
    </a-card>

    <!-- 历史记录弹窗 -->
    <a-modal 
      v-model:open="historyVisible" 
      :width="900"
      :footer="null"
      class="history-modal"
    >
      <template #title>
        <div class="modal-title">
          <span class="title-icon">📊</span>
          <span class="title-text">采集历史记录</span>
        </div>
      </template>
      
      <div v-if="currentHistory" class="history-content">
        <div class="file-info-header">
          <div class="file-info-item">
            <span class="info-label">文件名：</span>
            <span class="info-value">{{ currentHistory.FileName }}</span>
          </div>
          <div class="file-info-item">
            <span class="info-label">状态：</span>
            <a-tag :color="getStatusColor(currentHistory.Status)">{{ currentHistory.Status }}</a-tag>
          </div>
          <div class="file-info-item">
            <span class="info-label">CD时间：</span>
            <span class="info-value">{{ currentHistory.CdTime }}</span>
          </div>
        </div>
        
        <a-divider style="margin: 16px 0" />
        
        <div class="stats-summary">
          <div class="stat-card">
            <div class="stat-label">历史记录</div>
            <div class="stat-value">{{ currentHistory.History.length }} 次</div>
          </div>
          <div class="stat-card">
            <div class="stat-label">平均耗时</div>
            <div class="stat-value">{{ calculateAvgTime(currentHistory.History) }} 秒</div>
          </div>
          <div class="stat-card">
            <div class="stat-label">总采集量</div>
            <div class="stat-value">{{ calculateTotalItems(currentHistory.History) }} 个</div>
          </div>
        </div>
        
        <a-divider style="margin: 16px 0">详细记录</a-divider>
        
        <div class="history-list">
          <div 
            v-for="(record, index) in currentHistory.History" 
            :key="index" 
            class="history-item"
          >
            <div class="history-item-header">
              <span class="history-index">第 {{ index + 1 }} 次</span>
              <span class="history-duration">
                <span class="duration-icon">⏱️</span>
                {{ record.DurationSec }} 秒
              </span>
            </div>
            <div class="history-item-body">
              <div class="items-label">采集物品：</div>
              <div class="items-tags">
                <a-tag 
                  v-for="(count, itemName) in record.Item" 
                  :key="itemName" 
                  :color="getItemTagColor(itemName)"
                  class="item-tag"
                >
                  <span class="item-name">{{ itemName }}</span>
                  <span class="item-count">× {{ count }}</span>
                </a-tag>
              </div>
            </div>
          </div>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'
import dayjs from 'dayjs'

const loading = ref(false)
const selectedDate = ref(null)
const selectedStatus = ref('')
const selectedAccount = ref('') // 当前选中的账户
const accountList = ref([]) // 账户列表
const activeKeys = ref([])
const collectionData = ref({})
const historyVisible = ref(false)
const currentHistory = ref(null)
const pickupLoading = ref(false) // 采集历史加载状态
const pickupData = ref([]) // 采集历史数据

const columns = [
  {
    title: '文件名',
    key: 'FileName',
    dataIndex: 'FileName',
    width: '40%'
  },
  {
    title: 'CD时间',
    key: 'CdTime',
    dataIndex: 'CdTime',
    width: '25%'
  },
  {
    title: '状态',
    key: 'Status',
    dataIndex: 'Status',
    width: '20%',
    align: 'center'
  },
  {
    title: '操作',
    key: 'action',
    width: '15%',
    align: 'center'
  }
]

const historyColumns = [
  {
    title: '次数',
    key: 'index',
    width: 80
  },
  {
    title: '采集物品',
    key: 'Item',
    dataIndex: 'Item'
  },
  {
    title: '耗时',
    key: 'DurationSec',
    dataIndex: 'DurationSec',
    width: 100,
    align: 'center'
  }
]

// 计算平均时间
const calculateAvgTime = (history) => {
  if (!history || history.length === 0) return 0
  const total = history.reduce((sum, record) => sum + record.DurationSec, 0)
  return Math.round(total / history.length)
}

// 计算总采集量
const calculateTotalItems = (history) => {
  if (!history || history.length === 0) return 0
  let total = 0
  history.forEach(record => {
    Object.values(record.Item).forEach(count => {
      total += count
    })
  })
  return total
}

// 物品标签颜色
const getItemTagColor = (itemName) => {
  // 根据物品名称返回不同颜色
  const colorMap = {
    '夏槲果': 'green',
    '宿影花': 'purple',
    '薄荷': 'cyan',
    '树莓': 'red',
    '甘甜花': 'pink',
    '青蛙': 'blue',
    '月萤虫': 'orange',
    '海蓝蟹': 'geekblue',
    '薄红蟹': 'volcano',
    '白灵果': 'lime',
    '鸟蛋': 'gold',
    '蜜桃': 'magenta',
    '蝴蝶': 'blue',
    '蜜蟹': 'orange'
  }
  return colorMap[itemName] || 'default'
}

// 计算每日总采集量
const calculateDailyTotal = (items) => {
  return Object.values(items).reduce((sum, count) => sum + count, 0)
}

// 按数量排序物品
const sortedItems = (items) => {
  return Object.fromEntries(
    Object.entries(items).sort((a, b) => b[1] - a[1])
  )
}

// 计算日期差值
const formatDateDiff = (dateStr) => {
  const date = dayjs(dateStr)
  const today = dayjs()
  const diff = today.diff(date, 'day')
  
  if (diff === 0) return '今天'
  if (diff === 1) return '昨天'
  if (diff === 2) return '前天'
  return `${diff} 天前`
}

// 过滤数据
const filteredData = computed(() => {
  if (!collectionData.value) return {}
  
  const result = {}
  
  Object.keys(collectionData.value).forEach(material => {
    const items = collectionData.value[material]
    
    const filtered = items.filter(item => {
      // 日期过滤
      if (selectedDate.value) {
        const selectedDateStr = dayjs(selectedDate.value).format('YYYY-MM-DD')
        const itemDateStr = item.CdTime.split(' ')[0]
        if (itemDateStr !== selectedDateStr) {
          return false
        }
      }
      
      // 状态过滤
      if (selectedStatus.value && item.Status !== selectedStatus.value) {
        return false
      }
      
      return true
    })
    
    if (filtered.length > 0) {
      result[material] = filtered
    }
  })
  
  return result
})

const getStatusColor = (status) => {
  switch (status) {
    case '可采集': return 'green'
    case 'CD中': return 'orange'
    default: return 'default'
  }
}

const showHistory = (record) => {
  currentHistory.value = record
  historyVisible.value = true
}

// 获取所有账户列表
const fetchAccountList = async () => {
  try {
    const response = await apiMethods.getAllUserFiles()
    accountList.value = response || []
    
    // 默认选中第一个账户
    if (accountList.value.length > 0 && !selectedAccount.value) {
      selectedAccount.value = accountList.value[0]
    }
  } catch (error) {
    message.error('获取账户列表失败: ' + error.message)
  }
}

const refreshData = async () => {
  if (!selectedAccount.value) {
    message.warning('请先选择账户')
    return
  }
  
  loading.value = true
  try {
    const response = await apiMethods.getCollectionManagement(selectedAccount.value)
    collectionData.value = response || {}
    
    // 默认展开第一个材料分类
    if (Object.keys(collectionData.value).length > 0) {
      activeKeys.value = [Object.keys(collectionData.value)[0]]
    }
  } catch (error) {
    message.error('获取采集数据失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

// 获取拾取记录
const refreshPickupData = async () => {
  if (!selectedAccount.value) {
    message.warning('请先选择账户')
    return
  }
  
  pickupLoading.value = true
  try {
    const response = await apiMethods.getPickupHistory(selectedAccount.value)
    pickupData.value = response || []
  } catch (error) {
    message.error('获取采集历史失败: ' + error.message)
  } finally {
    pickupLoading.value = false
  }
}

const onDateChange = () => {
  // 日期变化时自动刷新过滤
}

const onStatusChange = () => {
  // 状态变化时自动刷新过滤
}

const onAccountChange = () => {
  // 账户变化时刷新数据
  refreshData()
  refreshPickupData()
}

onMounted(async () => {
  await fetchAccountList()
  if (selectedAccount.value) {
    await refreshData()
    await refreshPickupData()
  }
})
</script>

<style scoped>
.collection-management {
  padding: 20px;
}

.collection-card {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.file-name {
  font-size: 14px;
  color: #333;
}

.cd-time {
  font-size: 13px;
  color: #666;
}

/* 折叠面板样式优化 */
:deep(.ant-collapse) {
  background: #fff;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
}

:deep(.ant-collapse-header) {
  font-weight: 600;
  font-size: 15px;
  color: #1890ff;
  padding: 12px 16px !important;
}

:deep(.ant-collapse-content-box) {
  padding: 16px;
}

/* 表格样式优化 */
:deep(.ant-table) {
  font-size: 13px;
}

:deep(.ant-table-thead > tr > th) {
  background-color: #fafafa;
  font-weight: 600;
}

/* 移动端卡片样式 */
.mobile-cards {
  display: none; /* 桌面端隐藏 */
}

.collection-item-card {
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  transition: all 0.3s ease;
}

.collection-item-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.collection-item-card:last-child {
  margin-bottom: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 10px;
  gap: 8px;
}

.card-title {
  flex: 1;
  font-size: 13px;
  font-weight: 600;
  color: #333;
  line-height: 1.4;
  word-break: break-all;
}

.card-status {
  flex-shrink: 0;
}

.card-body {
  margin-bottom: 10px;
}

.card-info-row {
  display: flex;
  align-items: center;
  margin-bottom: 6px;
  font-size: 12px;
}

.card-info-row:last-child {
  margin-bottom: 0;
}

.info-label {
  color: #666;
  font-weight: 500;
  margin-right: 4px;
}

.info-value {
  color: #333;
  font-weight: 400;
}

.card-footer {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #f0f0f0;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .collection-management {
    padding: 10px;
  }
  
  :deep(.ant-card-head) {
    padding: 12px;
  }
  
  :deep(.ant-card-body) {
    padding: 12px;
  }
  
  :deep(.ant-space) {
    flex-wrap: wrap;
    gap: 8px !important;
  }
  
  :deep(.ant-select) {
    width: 100% !important;
  }
  
  :deep(.ant-btn) {
    width: 100%;
  }
  
  :deep(.ant-collapse-header) {
    font-size: 14px;
    padding: 10px 12px !important;
  }
  
  /* 隐藏表格，显示卡片 */
  .desktop-table {
    display: none !important;
  }
  
  .mobile-cards {
    display: block !important;
  }
  
  .collection-item-card {
    margin-bottom: 10px;
  }
  
  .card-title {
    font-size: 12px;
  }
  
  .card-info-row {
    font-size: 11px;
  }
  
  :deep(.ant-tag) {
    font-size: 11px;
    padding: 2px 6px;
  }
}

@media (max-width: 480px) {
  .collection-management {
    padding: 8px;
  }
  
  :deep(.ant-card-head-title) {
    font-size: 16px;
  }
  
  :deep(.ant-card-head) {
    padding: 10px;
  }
  
  :deep(.ant-card-body) {
    padding: 10px;
  }
  
  :deep(.ant-date-picker),
  :deep(.ant-select) {
    width: 100% !important;
    margin-bottom: 8px;
  }
  
  :deep(.ant-space) {
    width: 100%;
    gap: 8px !important;
  }
  
  :deep(.ant-space-item) {
    width: 100%;
  }
  
  :deep(.ant-btn) {
    width: 100%;
  }
  
  :deep(.ant-collapse-header) {
    font-size: 13px;
    padding: 8px 10px !important;
  }
  
  /* 小屏手机卡片优化 */
  .desktop-table {
    display: none !important;
  }
  
  .mobile-cards {
    display: block !important;
  }
  
  .collection-item-card {
    padding: 10px;
    margin-bottom: 8px;
  }
  
  .card-title {
    font-size: 12px;
  }
  
  .card-info-row {
    font-size: 11px;
  }
  
  .info-label {
    font-size: 11px;
  }
  
  .info-value {
    font-size: 11px;
  }
  
  :deep(.ant-tag) {
    font-size: 10px;
    padding: 1px 4px;
  }
  
  .card-footer :deep(.ant-btn) {
    font-size: 12px;
    height: 32px;
  }
}

/* 历史记录模态框样式 */
.history-modal :deep(.ant-modal-header) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-bottom: none;
}

.modal-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #fff;
  font-size: 18px;
  font-weight: 600;
}

.title-icon {
  font-size: 20px;
}

.history-content {
  max-height: 70vh;
  overflow-y: auto;
}

/* 文件信息头部 */
.file-info-header {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  padding: 16px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  border-radius: 8px;
  margin-bottom: 16px;
}

.file-info-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-label {
  font-weight: 600;
  color: #666;
  font-size: 14px;
}

.info-value {
  color: #333;
  font-size: 14px;
  font-weight: 500;
}

/* 统计摘要 */
.stats-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.stat-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  padding: 20px;
  text-align: center;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
  transition: transform 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.4);
}

.stat-label {
  color: rgba(255, 255, 255, 0.9);
  font-size: 13px;
  margin-bottom: 8px;
}

.stat-value {
  color: #fff;
  font-size: 24px;
  font-weight: bold;
}

/* 历史记录列表 */
.history-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.history-item {
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 16px;
  transition: all 0.3s ease;
}

.history-item:hover {
  border-color: #1890ff;
  box-shadow: 0 2px 12px rgba(24, 144, 255, 0.15);
  transform: translateX(4px);
}

.history-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 10px;
  border-bottom: 2px solid #f0f0f0;
}

.history-index {
  font-size: 15px;
  font-weight: 600;
  color: #1890ff;
  background: linear-gradient(135deg, #e6f7ff 0%, #bae7ff 100%);
  padding: 4px 12px;
  border-radius: 6px;
}

.history-duration {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: #52c41a;
  font-weight: 600;
  background: #f6ffed;
  padding: 4px 12px;
  border-radius: 6px;
}

.duration-icon {
  font-size: 16px;
}

.history-item-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.items-label {
  font-size: 13px;
  color: #666;
  font-weight: 500;
}

.items-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.item-tag {
  font-size: 13px;
  padding: 4px 12px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
  border: none;
}

.item-name {
  font-weight: 600;
}

.item-count {
  font-weight: bold;
  opacity: 0.9;
}

/* 移动端优化 */
@media (max-width: 768px) {
  .history-modal :deep(.ant-modal) {
    max-width: 95vw;
    margin: 10px auto;
  }
  
  .file-info-header {
    flex-direction: column;
    gap: 12px;
  }
  
  .stats-summary {
    grid-template-columns: 1fr;
  }
  
  .stat-card {
    padding: 16px;
  }
  
  .stat-value {
    font-size: 20px;
  }
  
  .history-item {
    padding: 12px;
  }
  
  .history-item-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .items-tags {
    gap: 6px;
  }
  
  .item-tag {
    font-size: 12px;
    padding: 3px 10px;
  }
}

/* 拾取记录样式 */
.empty-state {
  text-align: center;
  padding: 60px 20px;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-text {
  font-size: 16px;
  color: #999;
}

.pickup-timeline {
  margin-top: 20px;
}

.timeline-dot {
  font-size: 20px;
  display: inline-block;
  transform: scale(1.2);
}

.pickup-record {
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 12px;
  padding: 16px;
  margin-left: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  transition: all 0.3s ease;
}

.pickup-record:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
}

.pickup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.pickup-date {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.date-icon {
  font-size: 18px;
}

.pickup-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 16px;
  padding: 12px;
  background: linear-gradient(135deg, #f5f7fa 0%, #e8eaf0 100%);
  border-radius: 8px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat-label {
  font-size: 13px;
  color: #666;
  font-weight: 500;
}

.stat-value {
  font-size: 15px;
  color: #1890ff;
  font-weight: 600;
}

.pickup-items {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.pickup-item-tag {
  font-size: 13px;
  padding: 5px 12px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
  border: none;
  cursor: default;
  transition: all 0.2s ease;
}

.pickup-item-tag:hover {
  transform: scale(1.05);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.pickup-item-tag .item-name {
  font-weight: 600;
}

.pickup-item-tag .item-count {
  font-weight: bold;
  opacity: 0.9;
}

/* 采集历史移动端优化 */
@media (max-width: 768px) {
  .pickup-record {
    margin-left: 10px;
    padding: 12px;
  }
  
  .pickup-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .pickup-date {
    font-size: 14px;
  }
  
  .pickup-stats {
    grid-template-columns: 1fr;
    gap: 8px;
    padding: 10px;
  }
  
  .stat-item {
    font-size: 12px;
  }
  
  .stat-label {
    font-size: 12px;
  }
  
  .stat-value {
    font-size: 14px;
  }
  
  .pickup-items {
    gap: 6px;
  }
  
  .pickup-item-tag {
    font-size: 12px;
    padding: 4px 10px;
  }
}

@media (max-width: 480px) {
  .empty-state {
    padding: 40px 10px;
  }
  
  .empty-icon {
    font-size: 48px;
  }
  
  .empty-text {
    font-size: 14px;
  }
  
  .pickup-record {
    margin-left: 5px;
    padding: 10px;
  }
  
  .pickup-date {
    font-size: 13px;
  }
  
  .date-icon {
    font-size: 16px;
  }
  
  .pickup-stats {
    padding: 8px;
  }
  
  .stat-item {
    font-size: 11px;
  }
  
  .stat-label {
    font-size: 11px;
  }
  
  .stat-value {
    font-size: 13px;
  }
  
  .pickup-items {
    gap: 4px;
  }
  
  .pickup-item-tag {
    font-size: 11px;
    padding: 3px 8px;
  }
  
  .timeline-dot {
    font-size: 16px;
  }
}
</style>
