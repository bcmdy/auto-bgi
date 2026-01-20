<template>
  <div class="collection-management">
    <!-- 顶部工具栏 -->
    <div class="toolbar-container">
      <div class="toolbar-left">
        <div class="page-title">
          <span class="title-icon">🌸</span>
          <span class="title-text">采集管理</span>
        </div>
      </div>
      <div class="toolbar-right">
        <a-space :size="12" wrap>
          <a-select 
            v-model:value="selectedAccount" 
            @change="onAccountChange" 
            class="account-select"
            placeholder="选择账户"
            :loading="accountLoading"
          >
            <template #suffixIcon>
              <span>👤</span>
            </template>
            <a-select-option v-for="account in accountList" :key="account" :value="account">
              {{ account }}
            </a-select-option>
          </a-select>
          
          
          <a-button 
            @click="refreshData" 
            :loading="loading"
            type="primary"
            class="refresh-btn"
          >
            <template #icon><span v-if="!loading">🔄</span></template>
            {{ loading ? '刷新数据' : '刷新' }}
          </a-button>
        </a-space>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="content-container">
      <!-- 统计卡片 -->
      <div v-if="statisticsData" class="stats-cards">
        <div
          class="stat-card stat-card-total"
          :class="{ 'stat-card-active': selectedStatus === '' }"
          @click="onCardClick('')"
        >
          <div class="stat-icon">📁</div>
          <div class="stat-info">
            <div class="stat-label">采集路径</div>
            <div class="stat-value">{{ statisticsData.totalFiles }}</div>
            <div class="stat-desc">条路径文件</div>
          </div>
          <div v-if="selectedStatus === ''" class="stat-card-indicator">✓</div>
        </div>
        <div
          class="stat-card stat-card-available"
          :class="{ 'stat-card-active': selectedStatus === '可采集' }"
          @click="onCardClick('可采集')"
        >
          <div class="stat-icon">✅</div>
          <div class="stat-info">
            <div class="stat-label">可采集</div>
            <div class="stat-value">{{ statisticsData.availableCount }}</div>
            <div class="stat-desc">{{ ((statisticsData.availableCount / statisticsData.totalFiles) * 100).toFixed(1) }}%</div>
          </div>
          <div v-if="selectedStatus === '可采集'" class="stat-card-indicator">✓</div>
        </div>
        <div
          class="stat-card stat-card-cooling"
          :class="{ 'stat-card-active': selectedStatus === '冷却中' }"
          @click="onCardClick('冷却中')"
        >
          <div class="stat-icon">⏳</div>
          <div class="stat-info">
            <div class="stat-label">冷却中</div>
            <div class="stat-value">{{ statisticsData.coolingCount }}</div>
            <div class="stat-desc">{{ ((statisticsData.coolingCount / statisticsData.totalFiles) * 100).toFixed(1) }}%</div>
          </div>
          <div v-if="selectedStatus === '冷却中'" class="stat-card-indicator">✓</div>
        </div>
        <div
          class="stat-card stat-card-material"
          :class="{ 'stat-card-active': selectedStatus === 'material' }"
          @click="onCardClick('material')"
        >
          <div class="stat-icon">🎯</div>
          <div class="stat-info">
            <div class="stat-label">材料种类</div>
            <div class="stat-value">{{ statisticsData.materialTypes }}</div>
            <div class="stat-desc">种材料</div>
          </div>
          <div v-if="selectedStatus === 'material'" class="stat-card-indicator">✓</div>
        </div>
      </div>

      <!-- 标签页切换 -->
      <a-card class="tabs-container" style="margin-top: 20px">
        <a-tabs v-model:activeKey="activeTab" type="card" size="large">
          <!-- 拾取记录标签页 -->
          <a-tab-pane key="pickup" tab="📅 拾取记录">
            <template #tab>
              <span class="tab-title">
                <span class="tab-icon">📅</span>
                <span>拾取记录</span>
              </span>
            </template>
            
            <div class="tab-toolbar">
              <a-button @click="refreshPickupData" :loading="pickupLoading" type="primary">
                <template #icon><span v-if="!pickupLoading">🔄</span></template>
                {{ pickupLoading ? '刷新中...' : '刷新数据' }}
              </a-button>
            </div>

            <div v-if="pickupData.length === 0" class="empty-state">
              <div class="empty-icon">📅</div>
              <div class="empty-text">暂无拾取记录</div>
            </div>

            <a-timeline v-if="pickupData.length > 0" mode="left" class="pickup-timeline">
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
          </a-tab-pane>

          <!-- CD记录标签页 -->
          <a-tab-pane key="cd" tab="⏰ CD记录">
            <template #tab>
              <span class="tab-title">
                <span class="tab-icon">⏰</span>
                <span>CD记录</span>
              </span>
            </template>
            
            <div class="tree-container">
        <a-tree
          v-if="treeData.length > 0"
          :tree-data="treeData"
          :show-line="{ showLeafIcon: false }"
          :show-icon="true"
          :default-expand-all="false"
          :selectable="false"
          class="collection-tree"
        >
          <template #title="{ title, key, dataRef, ...nodeData }">
            <div class="tree-node-wrapper" :class="`node-type-${nodeData.is_dir ? 'folder' : 'file'}`">
              <div class="node-main">
                <div class="node-header">
                  <span class="node-title">{{ title }}</span>
                  
                  <!-- 文件夹节点：显示子节点数量 -->
                  <template v-if="nodeData.is_dir">
                    <!-- <span class="children-count-badge">
                      📁 {{ nodeData.children ? nodeData.children.length : 0 }} 项
                    </span> -->
                    <span 
                      class="children-count-badge" 
                      style="margin-left: 8px; background: linear-gradient(135deg, #1890ff 0%, #36cfc9 100%);"
                    >
                      📊 {{ (nodeData.nodeStats || (dataRef && dataRef.nodeStats) || {available: 0, total: 0}).available }}/{{ (nodeData.nodeStats || (dataRef && dataRef.nodeStats) || {available: 0, total: 0}).total }}
                    </span>
                  </template>
                  
                  <!-- 文件节点：显示冷却倒计时 -->
                  <template v-else-if="nodeData.record && nodeData.record.FileName">
                    <span v-if="nodeData.countdown !== undefined" class="countdown-badge" :class="getCountdownClass(nodeData.countdown)">
                      {{ formatCountdown(nodeData.countdown) }}
                    </span>
                  </template>
                </div>
                
                <!-- 文件节点：显示record信息 -->
                <div v-if="!nodeData.is_dir && nodeData.record" class="file-info-row" style="display: flex; margin-top: 12px; padding-top: 12px; border-top: 1px solid #f0f0f0;">
                  <div class="file-info-left">

                      <!-- 显示 CD 时间 -->
                      <div v-if="nodeData.record.CdTime" class="cd-time-info">
                        <span class="cd-time-label">⏰ CD时间：</span>
                        <span class="cd-time-value">
                          {{ nodeData.record.CdTime }}
                        </span>
                      </div>
                      <!-- 显示最近一次采集历史 -->
                      <div v-if="nodeData.record?.History && Array.isArray(nodeData.record.History) && nodeData.record.History.length > 0" class="latest-collect">
                        <span class="latest-label">📌 最近采集（{{ nodeData.record.History.length }}次记录）：</span>
                        <a-tag 
                          v-for="(count, itemName) in nodeData.record.History[0].Item" 
                          :key="itemName"
                          :color="getItemTagColor(itemName)"
                          size="small"
                          class="item-mini-tag"
                        >
                          {{ itemName }} ×{{ count }}
                        </a-tag>
                      </div>
                      <div v-else class="latest-collect">
                        <span class="latest-label">📌 采集历史：</span>
                        <span class="no-history-hint">暂无历史记录</span>
                      </div>
                    </div>
                    <div class="file-info-right">
                      <!-- 显示状态 -->
                      <a-tag 
                        v-if="nodeData.record.Status"
                        :color="getStatusColor(nodeData.record.Status)" 
                        class="status-tag"
                      >
                        <span v-if="nodeData.record.Status === '可采集'">✅</span>
                        <span v-else-if="nodeData.record.Status === '冷却中'">⏳</span>
                        <span v-else>❓</span>
                        {{ nodeData.record.Status }}
                      </a-tag>
                      <a-button 
                        v-if="nodeData.record.History && nodeData.record.History.length > 0"
                        type="primary"
                        size="small"
                        @click="showHistory(nodeData.record)"
                        class="history-btn"
                      >
                        📊 查看完整历史
                      </a-button>
                    </div>
                  </div>
              </div>
            </div>
          </template>
          
          <template #icon="{ dataRef, ...nodeData }">
            <span class="tree-node-icon" :class="`icon-${nodeData.is_dir ? 'folder' : 'file'}`">
              <template v-if="nodeData.is_dir">📁</template>
              <template v-else-if="!nodeData.is_dir && nodeData.record && nodeData.record.FileName">
                <span 
                  v-if="nodeData.record.Status === '可采集'" 
                  class="file-icon-available"
                >✅</span>
                <span 
                  v-else-if="nodeData.record.Status === '冷却中'" 
                  class="file-icon-cd"
                >⏳</span>
                <span v-else>❓</span>
              </template>
              <template v-else>📄</template>
            </span>
          </template>
        </a-tree>
        
        <!-- 空状态 -->
        <div v-else class="empty-state">
          <div class="empty-icon">📭</div>
          <div class="empty-text">暂无采集数据</div>
          <div class="empty-hint">请选择账户后刷新数据</div>
        </div>
            </div>
          </a-tab-pane>
        </a-tabs>
      </a-card>
    </div>

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
            <div class="stat-value">{{ currentHistory.History && Array.isArray(currentHistory.History) ? currentHistory.History.length : 0 }} 次</div>
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
        
        <div v-if="currentHistory.History && Array.isArray(currentHistory.History) && currentHistory.History.length > 0" class="history-list">
          <div 
            v-for="(record, index) in currentHistory.History" 
            :key="index" 
            class="history-item"
          >
            <div class="history-item-header">
              <span class="history-index">第 {{ index + 1 }} 次</span>
              <span class="history-duration">
                <span class="duration-icon">⏱️</span>
                {{ record.DurationSec || 0 }} 秒
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
        <div v-else class="empty-state">
          <div class="empty-icon">📋</div>
          <div class="empty-text">暂无采集历史记录</div>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, onUnmounted } from 'vue'
import { message } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'
import dayjs from 'dayjs'

const loading = ref(false)
const accountLoading = ref(false)
const selectedStatus = ref('')
const selectedAccount = ref('') // 当前选中的账户
const accountList = ref([]) // 账户列表
const rawTreeData = ref(null) // 原始树形数据
const historyVisible = ref(false)
const currentHistory = ref(null)
const pickupLoading = ref(false) // 采集历史加载状态
const pickupData = ref([]) // 采集历史数据
const activeTab = ref('cd') // 当前激活的标签页，默认显示CD记录
const currentTime = ref(dayjs()) // 当前时间，用于倒计时
let countdownTimer = null // 倒计时定时器




// 倒计时格式化
const formatCountdown = (countdown) => {
  if (countdown <= 0) return '已可采集'
  
  const days = Math.floor(countdown / (24 * 60 * 60))
  const hours = Math.floor((countdown % (24 * 60 * 60)) / (60 * 60))
  const minutes = Math.floor((countdown % (60 * 60)) / 60)
  const seconds = countdown % 60
  
  if (days > 0) return `${days}天${hours}小时`
  if (hours > 0) return `${hours}小时${minutes}分钟`
  if (minutes > 0) return `${minutes}分${seconds}秒`
  return `${seconds}秒`
}

// 获取倒计时样式类
const getCountdownClass = (countdown) => {
  if (countdown <= 0) return 'countdown-available'
  if (countdown < 60 * 60) return 'countdown-soon' // 1小时内
  if (countdown < 24 * 60 * 60) return 'countdown-today' // 24小时内
  return 'countdown-long'
}

// 启动倒计时定时器
const startCountdownTimer = () => {
  if (countdownTimer) clearInterval(countdownTimer)
  countdownTimer = setInterval(() => {
    currentTime.value = dayjs()
  }, 1000) // 每秒1秒更新
}

// 停止倒计时定时器
const stopCountdownTimer = () => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}

// 计算平均时间
const calculateAvgTime = (history) => {
  if (!history || !Array.isArray(history) || history.length === 0) return 0
  const total = history.reduce((sum, record) => sum + (record.DurationSec || 0), 0)
  return Math.round(total / history.length)
}

// 计算总采集量
const calculateTotalItems = (history) => {
  if (!history || !Array.isArray(history) || history.length === 0) return 0
  let total = 0
  history.forEach(record => {
    if (record.Item && typeof record.Item === 'object') {
      Object.values(record.Item).forEach(count => {
        total += (count || 0)
      })
    }
  })
  return total
}

// 物品标签颜色
const getItemTagColor = (itemName) => {
  const colorMap = {
    '琉鳞石': 'red',
    '绯樱绣球': 'pink',
    '云岩裂叶': 'green',
    '薄荷': 'cyan',
    '甘甘花': 'lime',
    '铁块': 'default',
    '月莔虫': 'orange',
    '夏槲果': 'green',
    '宿影花': 'purple',
    '树莓': 'red',
    '青蛙': 'blue',
    '海蓝蟹': 'geekblue',
    '薄红蟹': 'volcano',
    '白灵果': 'lime',
    '鸟蛋': 'gold',
    '蜜桃': 'magenta',
    '蝶蝶': 'blue',
    '蜜蟹': 'orange'
  }
  return colorMap[itemName] || 'blue'
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



// 获取所有账户列表
const fetchAccountList = async () => {
  accountLoading.value = true
  try {
    const response = await apiMethods.getAllUserFiles()
    accountList.value = response || []
    
    // 默认选中第一个账户
    if (accountList.value.length > 0 && !selectedAccount.value) {
      selectedAccount.value = accountList.value[0]
    }
  } catch (error) {
    message.error('获取账户列表失败: ' + error.message)
  } finally {
    accountLoading.value = false
  }
}

// 计算统计数据
const statisticsData = computed(() => {
  if (!rawTreeData.value) return null

  let totalFiles = 0
  let availableCount = 0
  let coolingCount = 0
  const materialSet = new Set()

  const countFiles = (node) => {
    if (!node) return

    // 统计所有文件，不仅仅是.json文件
    if (!node.is_dir && node.name) {
      totalFiles++

      // 如果有有效的record数据，统计状态
      if (node.record && node.record.FileName) {
        if (node.record.Status === '可采集') {
          availableCount++
        } else if (node.record.Status === '冷却中') {
          coolingCount++
        }

        // 从历史记录中提取材料种类
        if (node.record.History && Array.isArray(node.record.History)) {
          node.record.History.forEach(history => {
            if (history.Item && typeof history.Item === 'object') {
              Object.keys(history.Item).forEach(itemName => {
                if (itemName && itemName.trim()) {
                  materialSet.add(itemName.trim())
                }
              })
            }
          })
        }
      }
    }

    // 递归处理子节点
    if (node.children && Array.isArray(node.children)) {
      node.children.forEach(child => countFiles(child))
    }
  }

  countFiles(rawTreeData.value)

  return {
    totalFiles,
    availableCount,
    coolingCount,
    materialTypes: materialSet.size
  }
})

  // 转换树形数据为 Ant Design Tree 组件所需格式
const convertToTreeData = (node, parentKey = '0') => {
  if (!node) return []
  
  // 如果是根目录 pathing，直接处理其子节点，不显示根节点本身
  if (node.name === 'pathing' && node.children && Array.isArray(node.children)) {
    const children = []
    node.children.forEach((child, index) => {
      const childNodes = convertToTreeData(child, `0-${index}`)
      children.push(...childNodes)
    })
    return children
  }
  
  const key = `${parentKey}-${node.name}`
  
  const treeNode = {
    title: node.name,
    key: key,
    // 直接将 node 的所有属性展开到树节点中
    ...node,
    // 保留 dataRef 以保证兼容性
    dataRef: { 
      ...node
    }
  }
  
  // 如果是文件节点且有有效的 record 数据，计算倒计时
  if (!node.is_dir && node.record && node.record.FileName) {
    if (node.record.CdTime) {
      const cdTime = dayjs(node.record.CdTime)
      const now = currentTime.value
      const countdown = cdTime.diff(now, 'second')
      treeNode.dataRef.countdown = countdown > 0 ? countdown : 0
    }
  }
  
  // 递归处理子节点
  if (node.children && Array.isArray(node.children) && node.children.length > 0) {
    const children = []
    node.children.forEach((child, index) => {
      const childNodes = convertToTreeData(child, `${key}-${index}`)
      children.push(...childNodes)
    })
    if (children.length > 0) {
      treeNode.children = children
    }
  }
  
  return [treeNode]
}

// 检查节点是否包含特定材料
const hasMaterial = (node, materialName) => {
  if (!node.dataRef?.record?.History || !Array.isArray(node.dataRef.record.History)) {
    return false
  }

  for (const history of node.dataRef.record.History) {
    if (history.Item && typeof history.Item === 'object') {
      if (Object.keys(history.Item).some(item => item.trim() === materialName)) {
        return true
      }
    }
  }
  return false
}

// 计算树形数据（带状态过滤）
const treeData = computed(() => {
  if (!rawTreeData.value) return []

  const convertedData = convertToTreeData(rawTreeData.value)

  // 如果没有状态过滤，直接返回
  if (!selectedStatus.value) {
    return convertedData
  }

  // 应用状态过滤
  const filterTree = (nodes) => {
    return nodes.map(node => {
      // 如果是文件节点，应用状态过滤
      if (!node.dataRef?.is_dir && node.dataRef?.record && node.dataRef.record.FileName) {
        if (selectedStatus.value === '可采集' || selectedStatus.value === '冷却中') {
          // 状态筛选
          if (node.dataRef.record.Status !== selectedStatus.value) {
            return null
          }
        } else if (selectedStatus.value === 'material') {
          // 材料种类筛选 - 显示所有有历史记录的文件
          if (!node.dataRef.record.History || !Array.isArray(node.dataRef.record.History) || node.dataRef.record.History.length === 0) {
            return null
          }
        }
      }

      // 递归过滤子节点
      if (node.children) {
        const filteredChildren = filterTree(node.children).filter(Boolean)
        if (filteredChildren.length > 0) {
          return { ...node, children: filteredChildren }
        } else if (!node.dataRef?.is_dir) {
          // 如果是文件节点且满足过滤条件，保留
          return node
        }
        return null
      }

      return node
    }).filter(Boolean)
  }

  return filterTree(convertedData)
})

const getStatusColor = (status) => {
  switch (status) {
    case '可采集': return 'green'
    case '冷却中': return 'orange'
    default: return 'default'
  }
}

const showHistory = (record) => {
  if (!record || !record.FileName) return
  // 确保 History 是数组
  if (!record.History || !Array.isArray(record.History)) {
    record.History = []
  }
  currentHistory.value = record
  historyVisible.value = true
}

const calculateNodeStats = (node) => {
  let total = 0
  let available = 0

  if (!node) return { total, available }

  if (!node.is_dir) {
    // It's a file
    total = 1
    if (node.record && node.record.Status === '可采集') {
      available = 1
    }
  } else if (node.children && Array.isArray(node.children)) {
    // It's a directory
    node.children.forEach(child => {
      const childStats = calculateNodeStats(child)
      total += childStats.total
      available += childStats.available
    })
  }

  // Attach stats to the node
  node.nodeStats = { total, available }
  
  return { total, available }
}

const refreshData = async () => {
  if (!selectedAccount.value) {
    message.warning('请先选择账户')
    return
  }
  
  loading.value = true
  try {
    const response = await apiMethods.getCollectionManagement(selectedAccount.value)
    
    // 如果返回的是树形结构
    if (response && typeof response === 'object') {
      calculateNodeStats(response)
      rawTreeData.value = response
    } else {
      rawTreeData.value = null
    }
    
    message.success('刷新成功')
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

const onCardClick = (status) => {
  if (selectedStatus.value === status) {
    // 如果点击的是当前已选中的卡片，则取消筛选
    selectedStatus.value = ''
  } else {
    selectedStatus.value = status
  }
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
  // 启动倒计时定时器
  startCountdownTimer()
})

onUnmounted(() => {
  // 组件销毁时停止定时器
  stopCountdownTimer()
})
</script>

<style scoped>
.collection-management {
  padding: 20px;
  background: #f5f7fa;
  min-height: 100vh;
}

/* 顶部工具栏 */
.toolbar-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 20px 24px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.toolbar-left .page-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.title-icon {
  font-size: 28px;
}

.title-text {
  font-size: 22px;
  font-weight: 700;
  color: #1890ff;
}

.toolbar-right {
  display: flex;
  align-items: center;
}

.account-select,
.status-select {
  min-width: 150px;
}

.refresh-btn {
  font-weight: 600;
}

/* 统计卡片 */
.stats-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
  margin-bottom: 20px;  
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 24px;
  background: #b5f1f1!important;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  border: 2px solid transparent;
  position: relative;
  overflow: hidden;
  cursor: pointer;
}

.stat-card::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 5px;
  background: linear-gradient(180deg, #e5f5cf 0%, #d4e065 100%);
}

.stat-card-total::before {
  background: linear-gradient(180deg, #e3f1b0 0%, #d3eb9b 100%);
}

.stat-card-available::before {
  background: linear-gradient(180deg, #52c41a 0%, #389e0d 100%);
}

.stat-card-cooling::before {
  background: linear-gradient(180deg, #fa8c16 0%, #d46b08 100%);
}

.stat-card-material::before {
  background: linear-gradient(180deg, #c3e997 0%, #9ee21f 100%);
}

.stat-card:hover {
  transform: translateY(-6px);
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.15);
}

.stat-card-total:hover {
  border-color: #a9e98b;
  box-shadow: 0 6px 24px rgba(221, 213, 137, 0.3);
}

.stat-card-available:hover {
  border-color: #52c41a;
  box-shadow: 0 6px 24px rgba(82, 196, 26, 0.3);
}

.stat-card-cooling:hover {
  border-color: #fa8c16;
  box-shadow: 0 6px 24px rgba(250, 140, 22, 0.3);
}

.stat-card-material:hover {
  border-color: #f3c911;
  box-shadow: 0 6px 24px rgba(114, 46, 209, 0.3);
}

/* 卡片激活状态 */
.stat-card-active {
  border-color: #1890ff !important;
  box-shadow: 0 6px 24px rgba(24, 144, 255, 0.3) !important;
  transform: translateY(-4px);
}

.stat-card-indicator {
  position: absolute;
  top: 12px;
  right: 12px;
  width: 24px;
  height: 24px;
  background: #1890ff;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 16px;
  box-shadow: 0 2px 8px rgba(24, 144, 255, 0.3);
}

.stat-icon {
  font-size: 48px;
  line-height: 1;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.1));
}

.stat-info {
  flex: 1;
}

.stat-label {
  font-size: 20px!important;
  color: #040702!important;
  margin-bottom: 6px;
  font-weight: 1000!important;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  color: #a8e266;
  line-height: 1;
  margin-bottom: 4px;
}

.stat-desc {
  font-size: 12px;
  color: #ec0404;
  font-weight: 500;
}

.stat-card-total .stat-value {
  color: #1890ff;
}

.stat-card-available .stat-value {
  color: #52c41a;
}

.stat-card-cooling .stat-value {
  color: #fa8c16;
}

.stat-card-material .stat-value {
  color: #722ed1;
}

/* 内容容器 */
.content-container {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

/* 树容器 */
.tree-container {
  padding: 0;
  border: none;
  border-radius: 0;
  overflow-x: auto;
}

.collection-tree {
  background: transparent;
  min-width: 100%; /* 确保树占满容器宽度 */
}

/* 树节点样式 */
.tree-node-wrapper {
  width: 100%;
  padding: 0;
  max-width: 100%; /* 防止节点超出容器 */
  overflow: hidden; /* 隐藏溢出内容 */
}

.node-main {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px 16px;
  background: #fafafa;
  border-radius: 8px;
  border: 1px solid #e8e8e8;
  transition: all 0.3s ease;
  word-wrap: break-word; /* 长文本自动换行 */
  overflow-wrap: break-word;
}

.tree-node-wrapper:hover .node-main {
  background: #f0f7ff;
  border-color: #1890ff;
  box-shadow: 0 2px 8px rgba(24, 144, 255, 0.15);
}

.node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.node-title {
  font-size: 15px;
  font-weight: 600;
  color: #333;
  flex: 1;
  word-break: break-all; /* 强制长文件名换行 */
  line-height: 1.4;
}

/* 倒计时徽章 */
.countdown-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
  animation: pulse 2s ease-in-out infinite;
}

.countdown-available {
  background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%);
  color: #fff;
  animation: none;
}

.countdown-soon {
  background: linear-gradient(135deg, #ff4d4f 0%, #ff7875 100%);
  color: #fff;
}

.countdown-today {
  background: linear-gradient(135deg, #fa8c16 0%, #ffa940 100%);
  color: #fff;
}

.countdown-long {
  background: linear-gradient(135deg, #1890ff 0%, #40a9ff 100%);
  color: #fff;
}

/* 子节点数量徽章 */
.children-count-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  background: linear-gradient(135deg, #722ed1 0%, #9254de 100%);
  color: #fff;
  white-space: nowrap;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.8;
    transform: scale(1.05);
  }
}

/* 文件夹节点样式 */
.node-type-folder .node-main {
  background: linear-gradient(135deg, #f9f0ff 0%, #fff0f6 100%);
  border-left: 4px solid #722ed1;
}

.node-type-folder .node-title {
  color: #722ed1;
  font-size: 16px;
  font-weight: 700;
}

.node-type-folder:hover .node-main {
  background: linear-gradient(135deg, #efdbff 0%, #ffd6e7 100%);
  border-left-color: #9254de;
}

/* 文件节点样式 */
.node-type-file .node-main {
  background: #fff;
  border-left: 4px solid #1890ff;
}

.node-type-file:hover .node-main {
  border-left-color: #40a9ff;
  background: #e6f7ff;
}

.file-info-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
  flex-wrap: wrap; /* 移动端自动换行 */
}

.file-info-left {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}

/* 文件名信息样式 */
.file-name-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-name-label {
  font-size: 13px;
  color: #666;
  font-weight: 600;
}

.file-name-value {
  font-size: 13px;
  color: #333;
  font-weight: 600;
  background: linear-gradient(135deg, #f0f2f5 0%, #e8eaed 100%);
  padding: 4px 10px;
  border-radius: 6px;
  max-width: 100%; /* 移动端限制宽度 */
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  word-break: break-all; /* 强制换行 */
}

.cd-time-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cd-time-label {
  font-size: 13px;
  color: #666;
  font-weight: 600;
}

.cd-time-value {
  font-size: 14px;
  color: #fff;
  font-weight: 700;
  background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
  padding: 4px 12px;
  border-radius: 6px;
  box-shadow: 0 2px 6px rgba(24, 144, 255, 0.3);
}

.latest-collect {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.latest-label {
  font-size: 12px;
  color: #999;
  font-weight: 500;
}

.no-history-hint {
  font-size: 12px;
  color: #bbb;
  font-style: italic;
}

.item-mini-tag {
  font-size: 11px;
  padding: 2px 6px;
  margin: 0;
}

.file-info-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.status-tag {
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 6px;
  font-size: 13px;
}

.history-btn {
  font-weight: 600;
  border-radius: 6px;
  box-shadow: 0 2px 6px rgba(24, 144, 255, 0.3);
}

.history-btn:hover {
  box-shadow: 0 4px 12px rgba(24, 144, 255, 0.4);
}

/* 材料节点样式 */
.node-type-material .node-main {
  background: linear-gradient(135deg, #f6ffed 0%, #fcffe6 100%);
  border-left: 4px solid #52c41a;
}

.node-type-material .node-title {
  color: #52c41a;
  font-size: 16px;
}

.node-type-material:hover .node-main {
  background: linear-gradient(135deg, #d9f7be 0%, #f4ffb8 100%);
}

.material-stats {
  display: flex;
  align-items: center;
  gap: 12px;
}

.material-tag {
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 6px;
  font-size: 13px;
}

.material-progress {
  flex: 1;
  max-width: 200px;
}

/* 路径组节点样式 */
.node-type-group .node-main {
  background: linear-gradient(135deg, #f9f0ff 0%, #fff0f6 100%);
  border-left: 4px solid #722ed1;
}

.node-type-group .node-title {
  color: #722ed1;
  font-size: 16px;
}

.node-type-group:hover .node-main {
  background: linear-gradient(135deg, #efdbff 0%, #ffd6e7 100%);
}

/* 树节点图标 */
.tree-node-icon {
  font-size: 20px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
}

.file-icon-available {
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.8; transform: scale(1.1); }
}

/* Tree 组件样式覆盖 */
:deep(.ant-tree-treenode) {
  padding: 6px 0;
}

:deep(.ant-tree-node-content-wrapper) {
  flex: 1;
  min-width: 0;
  padding: 0 !important;
  line-height: 1;
}

:deep(.ant-tree-node-content-wrapper:hover) {
  background: transparent !important;
}

:deep(.ant-tree-title) {
  width: 100%;
}

:deep(.ant-tree-switcher) {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  color: #1890ff;
}

:deep(.ant-tree-indent-unit) {
  width: 28px;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 80px 20px;
}

.empty-icon {
  font-size: 72px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-text {
  font-size: 18px;
  color: #666;
  font-weight: 600;
  margin-bottom: 8px;
}

.empty-hint {
  font-size: 14px;
  color: #999;
}

.collection-card {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* 标签页容器样式 */
.tabs-container {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border-radius: 12px;
  overflow: hidden;
}

.tabs-container :deep(.ant-card-body) {
  padding: 0;
}

.tabs-container :deep(.ant-tabs) {
  margin: 0;
}

.tabs-container :deep(.ant-tabs-nav) {
  margin: 0;
  padding: 16px 16px 0 16px;
  background: linear-gradient(135deg, #f5f7fa 0%, #e8eaf0 100%);
}

.tabs-container :deep(.ant-tabs-tab) {
  background: #fff;
  border: 2px solid #e8e8e8;
  border-radius: 8px 8px 0 0;
  margin-right: 8px;
  padding: 10px 20px;
  font-weight: 600;
  transition: all 0.3s ease;
}

.tabs-container :deep(.ant-tabs-tab:hover) {
  border-color: #1890ff;
  transform: translateY(-2px);
}

.tabs-container :deep(.ant-tabs-tab-active) {
  background: linear-gradient(135deg, #f0d0eb 0%, #e93de0 100%);
  border-color: #1890ff;
  transform: translateY(-2px);
}

.tabs-container :deep(.ant-tabs-tab-active .tab-title) {
  color: #fff;
}

.tab-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  color: #333;
  transition: color 0.3s ease;
}

.tab-icon {
  font-size: 18px;
}

.tabs-container :deep(.ant-tabs-content) {
  padding: 24px;
  min-height: 400px;
}

.tab-toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 2px solid #f0f0f0;
}



/* 移动端适配 */
@media (max-width: 768px) {
  .collection-management {
    padding: 8px;
  }
  
  /* 工具栏移动端适配 */
  .toolbar-container {
    flex-direction: column;
    gap: 12px;
    padding: 12px;
    margin-bottom: 12px;
  }
  
  .toolbar-left,
  .toolbar-right {
    width: 100%;
  }
  
  .page-title {
    justify-content: center;
  }
  
  .title-icon {
    font-size: 20px;
  }
  
  .title-text {
    font-size: 16px;
  }
  
  :deep(.ant-space) {
    width: 100%;
  }
  
  :deep(.ant-space-item) {
    width: 100%;
  }
  
  .account-select,
  .status-select,
  .refresh-btn {
    width: 100% !important;
  }
  
  /* 统计卡片移动端适配 */
  .stats-cards {
    grid-template-columns: 1fr;
    gap: 8px;
    margin-bottom: 12px;
  }
  
  .stat-card {
    padding: 12px;
  }
  
  .stat-icon {
    font-size: 28px;
  }
  
  .stat-value {
    font-size: 20px;
  }
  
  .stat-label {
    font-size: 13px !important;
  }
  
  /* 树容器移动端适配 - 关键优化 */
  .tree-container {
    padding: 0;
    overflow-x: visible; /* 移动端不使用横向滚动 */
  }
  
  /* 标签页移动端适配 */
  .tabs-container :deep(.ant-tabs-nav) {
    padding: 8px 8px 0 8px;
  }
  
  .tabs-container :deep(.ant-tabs-tab) {
    padding: 6px 10px;
    margin-right: 4px;
  }
  
  .tab-title {
    font-size: 12px;
    gap: 4px;
  }
  
  .tab-icon {
    font-size: 14px;
  }
  
  .tabs-container :deep(.ant-tabs-content) {
    padding: 12px;
    min-height: 250px;
  }
  
  .tab-toolbar {
    margin-bottom: 12px;
  }
  
  .collection-tree {
    width: 100%;
  }
  
  /* 树节点移动端适配 - 关键优化 */
  .node-main {
    padding: 8px 10px;
    max-width: 100%;
  }
  
  .node-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
  }
  
  .node-title {
    font-size: 13px;
    width: 100%;
    white-space: normal; /* 允许标题换行 */
  }
  
  /* 倒计时和子节点徽章移动端适配 */
  .countdown-badge,
  .children-count-badge {
    font-size: 10px;
    padding: 2px 6px;
    width: fit-content;
  }
  
  /* 文件信息行移动端适配 - 垂直布局 */
  .file-info-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
    width: 100%;
  }
  
  .file-info-left {
    width: 100%;
  }
  
  .file-name-info {
    flex-direction: column;
    align-items: flex-start;
    width: 100%;
  }
  
  .file-name-value {
    max-width: 100%;
    font-size: 11px;
    white-space: normal; /* 移动端允许文件名换行 */
    word-break: break-all;
  }
  
  .cd-time-info {
    width: 100%;
  }
  
  .cd-time-label {
    font-size: 11px;
  }
  
  .cd-time-value {
    font-size: 11px;
    padding: 2px 6px;
  }
  
  .latest-collect {
    width: 100%;
  }
  
  .file-info-right {
    width: 100%;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 8px;
  }
  
  .status-tag {
    font-size: 11px;
    padding: 2px 8px;
  }
  
  .history-btn {
    flex: 1;
    min-width: 120px;
  }
  
  /* Tree 组件深度样式覆盖 */
  :deep(.ant-tree-treenode) {
    width: 100% !important;
  }
  
  :deep(.ant-tree-node-content-wrapper) {
    width: 100% !important;
    min-width: 0 !important;
  }
  
  :deep(.ant-tree-title) {
    width: 100% !important;
  }
  
  :deep(.ant-tree-indent-unit) {
    width: 12px !important; /* 减少缩进以节省空间 */
  }
  
  /* 材料统计移动端适配 */
  .material-stats {
    flex-wrap: wrap;
  }
  
  .tree-node-icon {
    font-size: 16px;
    width: 24px;
    height: 24px;
  }
  
  /* 空状态移动端适配 */
  .empty-state {
    padding: 40px 16px;
  }
  
  .empty-icon {
    font-size: 48px;
  }
  
  .empty-text {
    font-size: 14px;
  }
  
  .empty-hint {
    font-size: 12px;
  }
}

@media (max-width: 480px) {
  .collection-management {
    padding: 6px;
  }
  
  .toolbar-container {
    padding: 10px;
    gap: 10px;
    margin-bottom: 10px;
  }
  
  .title-icon {
    font-size: 18px;
  }
  
  .title-text {
    font-size: 15px;
  }
  
  .stats-cards {
    gap: 6px;
  }
  
  .stat-card {
    padding: 10px 12px;
  }
  
  .stat-icon {
    font-size: 24px;
  }
  
  .stat-label {
    font-size: 11px !important;
  }
  
  .stat-value {
    font-size: 18px;
  }
  
  .tree-container {
    padding: 0;
  }
  
  .tree-node-wrapper {
    padding: 6px 8px;
  }
  
  .node-title {
    font-size: 12px;
  }
  
  .cd-time {
    font-size: 10px;
  }
  
  .btn-text {
    font-size: 10px;
  }
  
  .custom-tree-icon {
    font-size: 14px;
    width: 20px;
    height: 20px;
  }
  
  .empty-state {
    padding: 30px 12px;
  }
  
  .empty-icon {
    font-size: 40px;
  }
  
  .empty-text {
    font-size: 13px;
  }
  
  .empty-hint {
    font-size: 11px;
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
  border: 3px solid rgb(66, 164, 230);
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
    margin-left: 8px;
    padding: 10px;
  }
  
  .pickup-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
  }
  
  .pickup-date {
    font-size: 13px;
  }
  
  .pickup-stats {
    grid-template-columns: 1fr;
    gap: 6px;
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
    gap: 5px;
  }
  
  .pickup-item-tag {
    font-size: 11px;
    padding: 3px 8px;
  }
}

@media (max-width: 480px) {
  .empty-state {
    padding: 30px 8px;
  }
  
  .empty-icon {
    font-size: 40px;
  }
  
  .empty-text {
    font-size: 13px;
  }
  
  .pickup-record {
    margin-left: 4px;
    padding: 8px;
  }
  
  .pickup-date {
    font-size: 12px;
  }
  
  .date-icon {
    font-size: 14px;
  }
  
  .pickup-stats {
    padding: 6px;
  }
  
  .stat-item {
    font-size: 10px;
  }
  
  .stat-label {
    font-size: 10px;
  }
  
  .stat-value {
    font-size: 12px;
  }
  
  .pickup-items {
    gap: 3px;
  }
  
  .pickup-item-tag {
    font-size: 10px;
    padding: 2px 6px;
  }
  
  .timeline-dot {
    font-size: 14px;
  }
}
</style>
