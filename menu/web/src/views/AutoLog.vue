<template>
  <div class="auto-log-page anime-theme">
    <div class="page-background" aria-hidden="true">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="blob blob-3"></div>
    </div>

    <a-card class="glass-card search-card" :bordered="false">
      <div class="search-header">
        <div class="card-title">
          <div class="icon-box">
            <file-search-outlined />
          </div>
          <span class="title-text">日志查询</span>
          <a-tag color="#ffadd2" class="title-tag">AutoTask</a-tag>
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
              class="anime-input"
            />
          </a-form-item>
          
          <a-form-item label="本地筛选">
            <a-input
              v-model:value="localKeyword"
              allow-clear
              :disabled="loading || logs.length === 0"
              placeholder="过滤内容/ID..."
              class="anime-input local-search-input"
            >
              <template #prefix>
                <filter-outlined style="color: #ff85c0"/>
              </template>
            </a-input>
          </a-form-item>

          <a-form-item class="action-buttons">
            <a-space wrap>
              <a-button type="primary" class="anime-btn primary-btn" :loading="loading" @click="handleFetchData">
                <template #icon><cloud-sync-outlined /></template>
                获取
              </a-button>
              
              <a-button class="anime-btn default-btn" @click="handleReset">重置</a-button>
              
              <a-tooltip title="导出CSV">
                 <a-button class="anime-btn icon-only-btn" :disabled="!filteredLogs.length" @click="exportLogs">
                    <download-outlined />
                 </a-button>
              </a-tooltip>
            </a-space>
          </a-form-item>
        </a-form>
      </div>
    </a-card>

    <a-card class="glass-card result-card" :bordered="false" :body-style="{ padding: '16px', height: '100%', display: 'flex', flexDirection: 'column' }">
      
      <transition name="fade-slide">
        <div v-if="logSummary.total" class="summary-dashboard">
          <div class="stat-item total">
            <div class="stat-icon"><bars-outlined /></div>
            <div class="stat-info">
              <div class="label">日志总数</div>
              <div class="value">{{ logSummary.total }}</div>
            </div>
          </div>
          
          <div class="stat-item error" :class="{ 'is-active': logSummary.error > 0 }">
             <div class="stat-icon"><close-circle-outlined /></div>
             <div class="stat-info">
               <div class="label">错误异常</div>
               <div class="value">{{ logSummary.error }}</div>
             </div>
          </div>

          <div class="stat-item success">
            <div class="stat-icon"><check-circle-outlined /></div>
            <div class="stat-info">
              <div class="label">成功率</div>
              <div class="value">{{ logSummary.successRate }}%</div>
            </div>
          </div>

          <div class="stat-item time">
             <div class="stat-icon"><clock-circle-outlined /></div>
             <div class="stat-info">
               <div class="label">执行时段</div>
               <div class="value-group">
                 <span class="main-time">{{ logSummary.earliestTime || '--:--' }}</span>
                 <span class="separator">~</span>
                 <span class="sub-time">{{ logSummary.latestTime || '--:--' }}</span>
               </div>
             </div>
          </div>
        </div>
      </transition>

      <div class="table-toolbar">
        <div class="left-tools">
            <span class="result-count">
              <span v-if="localKeyword">🔍 已筛选 <b>{{ filteredLogs.length }}</b> 条</span>
              <span v-else>🌸 共 {{ filteredLogs.length }} 条结果</span>
            </span>
            <a-divider type="vertical" />
            
            <a-radio-group v-model:value="levelFilter" button-style="solid" size="small" class="anime-radio-group">
              <a-radio-button value="ALL">全部</a-radio-button>
              <a-radio-button value="ERROR" class="btn-error">仅错误 ({{ logSummary.error }})</a-radio-button>
              <a-radio-button value="WARN" class="btn-warn">仅警告 ({{ logSummary.warn }})</a-radio-button>
            </a-radio-group>
        </div>
        <div class="right-tools">
            <a-button type="text" size="small" class="refresh-btn" @click="refresh" :loading="loading">
                <template #icon><reload-outlined /></template> 刷新
            </a-button>
        </div>
      </div>

      <div class="table-container">
        <a-table
          :data-source="filteredLogs"
          :columns="columns"
          :pagination="tablePagination"
          :scroll="{ x: 800, y: 'calc(100vh - 460px)' }"
          size="middle"
          row-key="id"
          :row-class-name="getRowClassName"
          class="anime-table"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.dataIndex === 'level'">
              <span class="anime-tag" :class="getLevelConfig(record.level).class">
                <component :is="getLevelConfig(record.level).icon" />
                {{ record.level }}
              </span>
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
               <span class="time-pill">{{ record.time.split(' ')[1] || record.time }}</span>
            </template>
          </template>
          
          <template #emptyText>
             <div class="empty-state">
                <img src="https://gw.alipayobjects.com/zos/antfincdn/ZHrcdLPrvN/empty.svg" alt="empty" style="height: 100px; opacity: 0.6;" />
                <p>{{ logs.length === 0 ? '暂无日志数据' : '没有找到匹配的结果哦~' }}</p>
             </div>
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
const localKeyword = ref('')   
const selectedDate = ref(null) 
const loading = ref(false)
const logs = ref([])           
const levelFilter = ref('ALL') 

// --- Table Config ---
const columns = [
  { title: '时间', dataIndex: 'time', key: 'time', width: 120, fixed: 'left', align: 'center' },
  { title: '级别', dataIndex: 'level', key: 'level', width: 110, align: 'center' },
  { title: '日志内容', dataIndex: 'msg', key: 'msg' }
]

const ellipsisConfig = { rows: 2, expandable: true, symbol: '展开' }

       setInterval(() => {
  debugger
}, 100)
// --- Computed Properties ---
const filteredLogs = computed(() => {
  let result = logs.value

  if (levelFilter.value !== 'ALL') {
    const target = levelFilter.value.toUpperCase()
    result = result.filter((log) => (log.level || '').toUpperCase().includes(target))
  }

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
const normalizeLogs = (raw) => {
  if (!raw) return []
  let items = raw
  
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

const fetchLogs = async () => {
  loading.value = true
  try {
    const res = await apiMethods.queryAutoLogs(selectedDate.value)
    if (res?.status === 'success') {
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

const highlightKeyword = (text) => {
    if (!localKeyword.value) return text
    const k = localKeyword.value
    // 简单的正则替换，生产环境建议转义正则字符
    const reg = new RegExp(`(${k})`, 'gi') 
    return text.replace(reg, '<span style="background-color: #ffe58f; color: #d46b08; font-weight: bold; padding: 0 2px; border-radius: 2px;">$1</span>')
}

// 修改：返回 class 名称而非 color，便于 CSS 控制
const getLevelConfig = (level = '') => {
  const upper = level.toUpperCase()
  if (upper.includes('ERR')) return { class: 'tag-error', icon: CloseCircleOutlined }
  if (upper.includes('WARN')) return { class: 'tag-warn', icon: WarningOutlined }
  if (upper.includes('DEBUG')) return { class: 'tag-debug', icon: InfoCircleOutlined }
  return { class: 'tag-info', icon: InfoCircleOutlined }
}

const getRowClassName = (record) => {
  const upper = (record?.level || '').toUpperCase()
  if (upper.includes('ERR')) return 'row-error'
  if (upper.includes('WARN')) return 'row-warn'
  return ''
}

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
/* --- 二次元粉色主题定义 --- */
.anime-theme {
  --primary-pink: #ffadd2;
  --deep-pink: #eb2f96;
  --soft-bg: #fff0f6;
  --glass-bg: rgba(255, 255, 255, 0.75);
  --glass-border: rgba(255, 255, 255, 0.8);
  --text-main: #5c3a58;
  --radius-lg: 20px;
  --radius-md: 12px;
  --shadow-soft: 0 8px 32px 0 rgba(235, 47, 150, 0.1);
}

.auto-log-page {
  position: relative;
  min-height: 100vh;
  padding: 20px;
  background-color: #fff0f6; /* 兜底色 */
  font-family: 'Nunito', 'PingFang SC', 'Microsoft YaHei', sans-serif; /* 选用圆润字体 */
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 16px;
  color: var(--text-main);
}

/* 动态背景球 */
.page-background {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  z-index: 0;
  overflow: hidden;
  background: linear-gradient(180deg, #fff0f6 0%, #fff7e6 100%);
}

.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(50px);
  opacity: 0.6;
  animation: float 10s infinite ease-in-out;
}
.blob-1 { width: 300px; height: 300px; background: #ffadd2; top: -50px; left: -50px; animation-delay: 0s; }
.blob-2 { width: 400px; height: 400px; background: #e6f7ff; bottom: -100px; right: -50px; animation-delay: -3s; }
.blob-3 { width: 250px; height: 250px; background: #d3adf7; top: 40%; left: 30%; opacity: 0.4; animation-delay: -5s; }

@keyframes float {
  0%, 100% { transform: translateY(0) scale(1); }
  50% { transform: translateY(20px) scale(1.05); }
}

/* 卡片玻璃拟态 */
.glass-card {
  position: relative;
  z-index: 1;
  background: var(--glass-bg);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--glass-border);
  box-shadow: var(--shadow-soft);
  border-radius: var(--radius-lg);
  transition: all 0.3s ease;
}

/* 顶部搜索栏 */
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
  gap: 10px;
}
.icon-box {
  width: 36px; height: 36px;
  background: #fff;
  border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  color: var(--deep-pink);
  font-size: 18px;
  box-shadow: 0 2px 8px rgba(235, 47, 150, 0.15);
}
.title-text {
  font-size: 18px;
  font-weight: 700;
  color: #780650;
}
.title-tag {
  border: none;
  color: #c41d7f;
  background: #ffd6e7;
  border-radius: 12px;
  padding: 0 10px;
}

/* 搜索表单美化 */
.search-form :deep(.ant-form-item-label > label) { color: #886278; }
.anime-input :deep(.ant-input), 
.anime-input :deep(.ant-picker-input > input) {
  border-radius: 20px;
  background: rgba(255,255,255,0.6);
  border-color: #ffd6e7;
  transition: all 0.3s;
}
.anime-input:hover :deep(.ant-input),
.anime-input:focus :deep(.ant-input) {
  border-color: var(--deep-pink);
  box-shadow: 0 0 0 2px rgba(235, 47, 150, 0.1);
}

.local-search-input { width: 220px; }

/* 按钮美化 */
.anime-btn { border-radius: 20px; border: none; font-weight: 600; box-shadow: 0 2px 6px rgba(0,0,0,0.05); }
.primary-btn { 
  background: linear-gradient(135deg, #ffadd2 0%, #eb2f96 100%); 
  color: white; 
}
.primary-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(235, 47, 150, 0.3);
  background: linear-gradient(135deg, #ffadd2 0%, #eb2f96 100%); 
  opacity: 0.9;
}
.default-btn { background: #fff; color: #eb2f96; border: 1px solid #ffadd2; }
.icon-only-btn { padding: 4px 10px; border: 1px solid #ffd6e7; color: #eb2f96; background: #fff; }

/* 统计区域 */
.summary-dashboard {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.stat-item {
  background: rgba(255,255,255,0.6);
  border-radius: 16px;
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  border: 1px solid rgba(255,255,255,0.5);
  transition: transform 0.2s;
}
.stat-item:hover { transform: translateY(-3px); background: #fff; }
.stat-icon {
  width: 42px; height: 42px;
  border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  font-size: 20px;
}

/* 颜色变体 */
.stat-item.total .stat-icon { background: #e6f7ff; color: #1890ff; }
.stat-item.error .stat-icon { background: #fff1f0; color: #ff4d4f; }
.stat-item.success .stat-icon { background: #f6ffed; color: #52c41a; }
.stat-item.time .stat-icon { background: #f9f0ff; color: #722ed1; }

.stat-item.error.is-active { background: #fff1f0; border-color: #ffa39e; }

.stat-info .label { font-size: 12px; color: #999; margin-bottom: 2px; }
.stat-info .value { font-size: 22px; font-weight: 800; color: #333; line-height: 1.1; font-family: 'Arial Rounded MT Bold', sans-serif; }

.value-group { display: flex; flex-direction: column; line-height: 1.2; }
.main-time { font-size: 16px; font-weight: 700; color: #722ed1; }
.sub-time { font-size: 12px; color: #b37feb; }
.separator { display: none; }

/* 筛选器 Tab */
.anime-radio-group :deep(.ant-radio-button-wrapper) {
  border: 1px solid #ffd6e7;
  color: #eb2f96;
  background: transparent;
}
.anime-radio-group :deep(.ant-radio-button-wrapper:first-child) { border-radius: 16px 0 0 16px; }
.anime-radio-group :deep(.ant-radio-button-wrapper:last-child) { border-radius: 0 16px 16px 0; }
.anime-radio-group :deep(.ant-radio-button-wrapper-checked) {
  background: #eb2f96 !important;
  color: #fff !important;
  border-color: #eb2f96 !important;
  box-shadow: none;
}
.anime-radio-group :deep(.btn-error.ant-radio-button-wrapper-checked) { background: #ff4d4f !important; border-color: #ff4d4f !important; }
.anime-radio-group :deep(.btn-warn.ant-radio-button-wrapper-checked) { background: #faad14 !important; border-color: #faad14 !important; }

/* 表格美化 */
.table-toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.result-count { font-size: 13px; color: #eb2f96; }
.refresh-btn { color: #999; }
.refresh-btn:hover { color: #eb2f96; background: #fff0f6; border-radius: 12px; }

.table-container { 
    flex: 1; 
    overflow: hidden; 
    border-radius: 16px; 
    background: rgba(255,255,255,0.4); 
    border: 1px solid #fff;
}

/* Ant Table 深度定制 */
:deep(.ant-table-thead > tr > th) {
  background: rgba(255, 240, 246, 0.8) !important;
  color: #780650;
  font-weight: 600;
  border-bottom: 1px solid #ffd6e7;
}
:deep(.ant-table-tbody > tr > td) {
  border-bottom: 1px solid #fff0f6;
  transition: background 0.3s;
}
:deep(.ant-table-tbody > tr:hover > td) {
  background: rgba(255, 240, 246, 0.5) !important;
}

/* 状态胶囊 */
.anime-tag {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 2px 8px; border-radius: 10px;
  font-size: 12px; font-weight: 600;
  border: 1px solid transparent;
}
.tag-info { background: #e6f7ff; color: #096dd9; border-color: #91d5ff; }
.tag-debug { background: #f0f0f0; color: #595959; border-color: #d9d9d9; }
.tag-warn { background: #fffbe6; color: #d48806; border-color: #ffe58f; }
.tag-error { background: #fff2f0; color: #cf1322; border-color: #ffccc7; }

.time-pill {
    background: #f9f0ff; color: #722ed1; padding: 2px 6px; border-radius: 6px; font-family: monospace; font-size: 12px;
}

/* 错误/警告行高亮 */
:deep(.row-error) { background-color: #fff2f0 !important; }
:deep(.row-warn) { background-color: #fffbe6 !important; }

.log-text { font-size: 13px; color: #444; }
.monospace { font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace; }

/* 动画 */
.fade-slide-enter-active, .fade-slide-leave-active { transition: all 0.5s ease; }
.fade-slide-enter-from, .fade-slide-leave-to { opacity: 0; transform: translateY(-10px); }

/* --- 移动端适配 --- */
@media (max-width: 768px) {
  .auto-log-page { padding: 10px; gap: 10px; }
  
  /* 搜索区调整 */
  .search-header { flex-direction: column; align-items: stretch; gap: 12px; }
  .card-title { justify-content: center; margin-bottom: 4px; }
  .search-form { width: 100%; display: flex; flex-direction: column; gap: 8px; }
  .search-form :deep(.ant-form-item) { margin-right: 0; margin-bottom: 0; width: 100%; }
  
  .anime-input, .local-search-input { width: 100% !important; }
  
  /* 按钮组 */
  .action-buttons { margin-top: 4px; }
  .action-buttons :deep(.ant-form-item-control-input-content) { display: flex; justify-content: stretch; }
  .action-buttons .ant-space { width: 100%; justify-content: space-between; }
  .primary-btn { flex: 1; }
  
  /* 统计区变紧凑 */
  .summary-dashboard { grid-template-columns: 1fr 1fr; gap: 10px; }
  .stat-item { padding: 12px; flex-direction: column; align-items: flex-start; gap: 8px; }
  .stat-icon { width: 32px; height: 32px; font-size: 16px; align-self: flex-start; }
  .stat-info .value { font-size: 18px; }
  
  /* 工具栏调整 */
  .table-toolbar { flex-direction: column; align-items: flex-start; gap: 10px; }
  .left-tools { width: 100%; display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
  .right-tools { position: absolute; top: 16px; right: 16px; } /* 移动端将刷新放回右上角 */
  
  /* 表格调整 */
  .result-card :deep(.ant-card-body) { padding: 12px; }
  /* 在手机上隐藏固定列阴影以减少视觉杂乱 */
  :deep(.ant-table-ping-left .ant-table-cell-fix-left) { box-shadow: none !important; border-right: 1px solid #eee; }
  
  .anime-radio-group { width: 100%; display: flex; }
  .anime-radio-group :deep(.ant-radio-button-wrapper) { flex: 1; text-align: center; padding: 0; font-size: 12px; }
}
</style>