<template>
  <div class="auto-log-page">
    <div class="page-background" aria-hidden="true"></div>

    <a-card class="glass-card search-card" :bordered="false">
      <div class="search-header">
        <div class="card-title">
          <file-search-outlined class="title-icon" />
          <span>日志查询</span>
          <a-tag color="purple" class="title-tag">自动任务</a-tag>
        </div>
        
        <a-form layout="inline" class="search-form" @submit.prevent>
          <a-form-item label="日期范围">
            <a-date-picker
              v-model:value="selectedDate"
              value-format="YYYY-MM-DD"
              :disabled="loading"
              allow-clear
              placeholder="默认今天"
              @change="handleFetchData"
            />
          </a-form-item>
          
          <a-form-item label="本地筛选">
            <a-input
              v-model:value="localKeyword"
              allow-clear
              :disabled="loading || logs.length === 0"
              placeholder="在结果中过滤内容/ID..."
              style="width: 260px"
            >
              <template #prefix>
                <filter-outlined style="color: rgba(0,0,0,0.25)"/>
              </template>
            </a-input>
          </a-form-item>

          <a-form-item>
            <a-space>
              <a-button type="primary" :loading="loading" @click="handleFetchData">
                <template #icon><cloud-sync-outlined /></template>
                获取日志
              </a-button>
              
              <a-button @click="handleReset">重置</a-button>
              
              <a-tooltip title="导出当前筛选结果为CSV">
                 <a-button :disabled="!filteredLogs.length" @click="exportLogs">
                    <download-outlined />
                 </a-button>
              </a-tooltip>
            </a-space>
          </a-form-item>
        </a-form>
      </div>
    </a-card>

    <a-card class="glass-card result-card" :bordered="false" :body-style="{ padding: '16px 24px', height: '100%', display: 'flex', flexDirection: 'column' }">
      
      <transition name="fade">
        <div v-if="logSummary.total" class="summary-dashboard">
          <div class="stat-item">
            <div class="stat-icon info-bg"><bars-outlined /></div>
            <div class="stat-info">
              <div class="label">日志总数</div>
              <div class="value">{{ logSummary.total }}</div>
            </div>
          </div>
          
          <div class="stat-item" :class="{ 'has-error': logSummary.error > 0 }">
             <div class="stat-icon error-bg"><close-circle-outlined /></div>
             <div class="stat-info">
               <div class="label">错误数</div>
               <div class="value">{{ logSummary.error }}</div>
             </div>
          </div>

          <div class="stat-item">
            <div class="stat-icon success-bg"><check-circle-outlined /></div>
            <div class="stat-info">
              <div class="label">成功率</div>
              <div class="value">{{ logSummary.successRate }}%</div>
            </div>
          </div>

          <div class="stat-item time-range">
             <div class="stat-icon time-bg"><clock-circle-outlined /></div>
             <div class="stat-info">
               <div class="label">执行时段</div>
               <div class="value-sm">{{ logSummary.earliestTime || '--:--' }}</div>
               <div class="value-sm sub">至 {{ logSummary.latestTime || '--:--' }}</div>
             </div>
          </div>
        </div>
      </transition>

      <div class="table-toolbar">
        <div class="left-tools">
            <span class="result-count">
              <span v-if="localKeyword">已筛选 <b>{{ filteredLogs.length }}</b> 条</span>
              <span v-else>共 {{ filteredLogs.length }} 条结果</span>
            </span>
            <a-divider type="vertical" />
            
            <a-radio-group v-model:value="levelFilter" button-style="solid" size="small">
              <a-radio-button value="ALL">全部</a-radio-button>
              <a-radio-button value="ERROR" class="btn-error">仅错误 ({{ logSummary.error }})</a-radio-button>
              <a-radio-button value="WARN" class="btn-warn">仅警告 ({{ logSummary.warn }})</a-radio-button>
            </a-radio-group>
        </div>
        <div class="right-tools">
            <a-button type="text" size="small" @click="refresh" :loading="loading">
                <template #icon><reload-outlined /></template> 刷新数据
            </a-button>
        </div>
      </div>

      <div class="table-container">
        <a-table
          :data-source="filteredLogs"
          :columns="columns"
          :pagination="tablePagination"
          :scroll="{ x: 800, y: 'calc(100vh - 420px)' }"
          size="middle"
          row-key="id"
          :row-class-name="getRowClassName"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.dataIndex === 'level'">
              <a-tag :color="getLevelConfig(record.level).color" class="level-tag">
                <component :is="getLevelConfig(record.level).icon" />
                {{ record.level }}
              </a-tag>
            </template>

            <template v-else-if="column.dataIndex === 'msg'">
              <div class="log-content-wrapper">
                <a-typography-paragraph
                  class="log-text monospace"
                  :ellipsis="ellipsisConfig"
                  :copyable="{ text: record.msg }"
                  :content="record.msg"
                >
                    <template v-if="localKeyword && record.msg">
                      <span v-html="highlightKeyword(record.msg)"></span>
                   </template>
                </a-typography-paragraph>
              </div>
            </template>
            
            <template v-else-if="column.dataIndex === 'time'">
               <span class="time-text">{{ record.time }}</span>
            </template>
          </template>
          
          <template #emptyText>
             <a-empty :description="logs.length === 0 ? '暂无日志数据' : '未找到匹配的筛选结果'" />
          </template>
        </a-table>
      </div>
    </a-card>
  </div>
</template>

<script setup>
import dayjs from 'dayjs'
import { computed, onMounted, ref } from 'vue'
import { message } from 'ant-design-vue'
import { 
  FileSearchOutlined, 
  ReloadOutlined, 
  DownloadOutlined,
  BarsOutlined,
  CloseCircleOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  InfoCircleOutlined,
  WarningOutlined,
  CloudSyncOutlined,
  FilterOutlined
} from '@ant-design/icons-vue'
// 请确保路径正确
import { apiMethods } from '@/utils/api'

// --- State Definition ---
const localKeyword = ref('')   // 本地过滤用的关键词
const selectedDate = ref(null) // 接口传参用的日期
const loading = ref(false)
const logs = ref([])           // 存储接口返回的全量原始数据
const levelFilter = ref('ALL') // 级别过滤器

// --- Table Config ---
const columns = [
  { title: '时间', dataIndex: 'time', key: 'time', width: 170, fixed: 'left' },
  { title: '级别', dataIndex: 'level', key: 'level', width: 110, align: 'center' },
  { title: '日志内容', dataIndex: 'msg', key: 'msg' }
]

const ellipsisConfig = { rows: 2, expandable: true, symbol: '展开' }

// --- Computed Properties ---

// 核心：双重过滤逻辑（先过滤级别，再过滤关键词）
const filteredLogs = computed(() => {
  let result = logs.value

  // 1. 级别过滤
  if (levelFilter.value !== 'ALL') {
    const target = levelFilter.value.toUpperCase()
    result = result.filter((log) => (log.level || '').toUpperCase().includes(target))
  }

  // 2. 本地关键词过滤
  if (localKeyword.value) {
    const k = localKeyword.value.toLowerCase().trim()
    result = result.filter((log) => {
      const msgMatch = (log.msg || '').toLowerCase().includes(k)
      const timeMatch = (log.time || '').includes(k)
      const idMatch = (log.id || '').toLowerCase().includes(k)
      return msgMatch || timeMatch || idMatch
    })
  }

  return result
})

const tablePagination = computed(() => ({
  pageSize: 50,
  showSizeChanger: true,
  pageSizeOptions: ['20', '50', '100', '200'],
  showTotal: (total) => `共 ${total} 条`
}))

// 统计面板的数据（基于全量 logs 计算，反映整体任务情况）
const logSummary = computed(() => {
  if (!logs.value.length) return { total: 0, error: 0, warn: 0, successRate: '0.0', earliestTime: '', latestTime: '' }
  
  let error = 0, warn = 0
  const validTimes = []

  logs.value.forEach(log => {
    const lvl = (log.level || '').toUpperCase()
    if (lvl.includes('ERR')) error++
    else if (lvl.includes('WARN')) warn++
    
    if (log.time && dayjs(log.time).isValid()) validTimes.push(dayjs(log.time))
  })

  // 计算时间范围
  let earliest = '-', latest = '-'
  if (validTimes.length) {
      validTimes.sort((a, b) => a.valueOf() - b.valueOf())
      earliest = validTimes[0].format('HH:mm:ss')
      latest = validTimes[validTimes.length - 1].format('HH:mm:ss')
  }

  return {
    total: logs.value.length,
    error,
    warn,
    successRate: (((logs.value.length - error) / logs.value.length) * 100).toFixed(1),
    earliestTime: earliest,
    latestTime: latest
  }
})

// --- Methods ---

// 数据清洗与格式化
const normalizeLogs = (raw) => {
  if (!raw) return []
  let items = raw
  
  // 兼容字符串形式的 JSON 响应
  if (typeof raw === 'string') {
    try {
      items = JSON.parse(raw)
    } catch {
      items = raw.split(/\r?\n/).filter(Boolean).map(l => ({ msg: l }))
    }
  }
  if (!Array.isArray(items)) items = [items]

  return items.map((item, index) => {
    let parsed = item
    if (typeof item === 'string') {
      try { parsed = JSON.parse(item) } catch { parsed = { msg: item } }
    }
    
    return {
      id: `${parsed.id || index}`,
      time: parsed.time || parsed.timestamp || parsed.date || dayjs().format('YYYY-MM-DD HH:mm:ss'),
      level: (parsed.level || parsed.Level || 'INFO').toUpperCase(),
      msg: typeof parsed.msg === 'object' ? JSON.stringify(parsed.msg) : (parsed.msg || parsed.message || JSON.stringify(parsed))
    }
  })
}

// 核心请求方法：只传日期
const fetchLogs = async () => {
  loading.value = true
  try {
    // 接口调用：仅传入 selectedDate
    const res = await apiMethods.queryAutoLogs(selectedDate.value)
    if (res?.status === 'success') {
       // 获取最新数据，localKeyword 会自动通过 computed 重新计算
      logs.value = normalizeLogs(res.msg).reverse()
    } else {
      logs.value = []
      message.warning(res?.msg || '未查询到数据')
    }
  } catch (err) {
    logs.value = []
    message.error('日志获取失败: ' + err.message)
  } finally {
    loading.value = false
  }
}

const handleFetchData = () => {
  fetchLogs()
}

const handleReset = () => {
  localKeyword.value = ''
  selectedDate.value = null
  levelFilter.value = 'ALL'
  fetchLogs()
}

const refresh = () => fetchLogs()

// 简单的关键词高亮函数
const highlightKeyword = (text) => {
    if (!localKeyword.value) return text
    const k = localKeyword.value
    // 简单的正则替换，注意：生产环境如果包含特殊字符需要转义
    const reg = new RegExp(`(${k})`, 'gi') 
    return text.replace(reg, '<span style="background-color: #ffd591; font-weight: bold;">$1</span>')
}

// 辅助：获取级别对应的颜色和图标
const getLevelConfig = (level = '') => {
  const upper = level.toUpperCase()
  if (upper.includes('ERR')) return { color: 'error', icon: CloseCircleOutlined }
  if (upper.includes('WARN')) return { color: 'warning', icon: WarningOutlined }
  if (upper.includes('DEBUG')) return { color: 'default', icon: InfoCircleOutlined }
  return { color: 'processing', icon: InfoCircleOutlined }
}

// 辅助：根据级别设置行样式
const getRowClassName = (record) => {
  const upper = (record?.level || '').toUpperCase()
  if (upper.includes('ERR')) return 'row-error-light'
  if (upper.includes('WARN')) return 'row-warn-light'
  return ''
}

// 导出 CSV
const exportLogs = () => {
    const header = 'Time,Level,Message\n'
    const content = filteredLogs.value.map(l => {
        const safeMsg = `"${(l.msg || '').replace(/"/g, '""')}"`
        return `${l.time},${l.level},${safeMsg}`
    }).join('\n')
    
    const blob = new Blob([header + content], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = `logs_${dayjs().format('MMDD_HHmm')}.csv`
    link.click()
}

onMounted(() => {
  fetchLogs()
})
</script>

<style scoped>
/* 全局变量 */
:root {
  --primary-color: #1890ff;
}

.auto-log-page {
  position: relative;
  min-height: 100vh;
  padding: 24px;
  background-color: #f0f2f5;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 动感背景 */
.page-background {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: 
    radial-gradient(at 0% 0%, hsla(253,16%,7%,1) 0, transparent 50%), 
    radial-gradient(at 50% 0%, hsla(225,39%,30%,1) 0, transparent 50%), 
    radial-gradient(at 100% 0%, hsla(339,49%,30%,1) 0, transparent 50%);
  background-size: cover;
  opacity: 0.05;
  z-index: 0;
  pointer-events: none;
}

.glass-card {
  position: relative;
  z-index: 1;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
  border-radius: 12px;
  transition: all 0.3s ease;
}

.search-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: #1f1f1f;
}

.title-icon { font-size: 20px; color: #1890ff; }

/* 结果区域布局 */
.result-card { flex: 1; display: flex; flex-direction: column; }

/* 统计仪表盘 */
.summary-dashboard {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.stat-item {
  background: #fff;
  border-radius: 10px;
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 16px;
  border: 1px solid #f0f0f0;
  transition: transform 0.2s;
}

.stat-item:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0,0,0,0.05);
}

.stat-item.has-error {
    border-color: #ffccc7;
    background: #fff1f0;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}
.info-bg { background: #e6f7ff; color: #1890ff; }
.error-bg { background: #fff2f0; color: #ff4d4f; }
.success-bg { background: #f6ffed; color: #52c41a; }
.time-bg { background: #f9f0ff; color: #722ed1; }

.stat-info { flex: 1; }
.stat-info .label { font-size: 12px; color: #8c8c8c; margin-bottom: 4px; }
.stat-info .value { font-size: 24px; font-weight: 700; color: #262626; line-height: 1; }
.stat-info .value-sm { font-size: 16px; font-weight: 600; color: #262626; }
.stat-info .sub { font-size: 12px; color: #8c8c8c; font-weight: normal; }

/* 工具栏 */
.table-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.left-tools { display: flex; align-items: center; gap: 12px; }
.result-count { font-size: 13px; color: #666; }

/* 表格样式 */
.table-container { flex: 1; overflow: hidden; }
.log-text { margin: 0; font-size: 13px; color: #333; }
.monospace { font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace; }
.time-text { color: #666; font-family: monospace; }

/* 状态行指示条 */
:deep(.row-error-light) td:first-child { box-shadow: inset 4px 0 0 #ff4d4f; }
:deep(.row-warn-light) td:first-child { box-shadow: inset 4px 0 0 #faad14; }

:deep(.row-error-light) { background-color: #fff1f0; }
:deep(.row-warn-light) { background-color: #fffbe6; }

/* AntDV 样式覆盖 */
:deep(.ant-card-body) { padding: 20px 24px; }
:deep(.ant-table-wrapper), :deep(.ant-spin-nested-loading), :deep(.ant-spin-container) { height: 100%; }
:deep(.ant-table), :deep(.ant-table-container) { height: 100%; }
:deep(.ant-table-body) { height: calc(100% - 40px) !important; }

/* 按钮筛选器颜色 */
:deep(.btn-error.ant-radio-button-wrapper-checked:not(.ant-radio-button-wrapper-disabled)) {
    background: #ff4d4f; border-color: #ff4d4f; box-shadow: -1px 0 0 0 #ff4d4f;
}
:deep(.btn-warn.ant-radio-button-wrapper-checked:not(.ant-radio-button-wrapper-disabled)) {
    background: #faad14; border-color: #faad14; box-shadow: -1px 0 0 0 #faad14;
}

/* 响应式 */
@media (max-width: 768px) {
  .auto-log-page { padding: 12px; }
  .search-header { flex-direction: column; align-items: stretch; }
  .search-form { width: 100%; }
  .search-form :deep(.ant-form-item) { margin-right: 0; width: 100%; margin-bottom: 12px; }
  .search-form :deep(.ant-input), .search-form :deep(.ant-picker) { width: 100% !important; }
  .summary-dashboard { grid-template-columns: 1fr 1fr; }
}
</style>