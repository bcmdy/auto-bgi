<template>
  <div class="auto-log-page">
    <!-- 动态背景 -->
    <div class="page-background" aria-hidden="true">
      <div class="gradient-orb orb-1"></div>
      <div class="gradient-orb orb-2"></div>
      <div class="gradient-orb orb-3"></div>
    </div>

    <!-- 搜索筛选卡片 -->
    <a-card class="modern-card search-section" :bordered="false">
      <div class="section-header">
        <div class="header-left">
          <div class="header-icon">
            <file-search-outlined />
          </div>
          <div class="header-title">
            <h2>日志查询</h2>
            <p>实时监控与分析</p>
          </div>
        </div>
        <a-tag color="#ff4d7d" class="header-badge">AutoTask</a-tag>
      </div>

      <!-- 筛选表单 -->
      <div class="filter-form">
        <div class="filter-row">
          <div class="filter-item">
            <label class="filter-label">日期</label>
            <a-date-picker
              v-model:value="selectedDate"
              value-format="YYYY-MM-DD"
              :disabled="loading"
              allow-clear
              placeholder="选择日期"
              @change="handleFetchData"
              class="custom-date-picker"
              :show-time="false"
              format="YYYY-MM-DD"
              :locale="datePickerLocale"
              :popup-style="{ width: '280px' }"
            >
              <template #suffixIcon>
                <calendar-outlined />
              </template>
            </a-date-picker>
          </div>

          <div class="filter-item">
            <label class="filter-label">关键词</label>
            <a-input
              v-model:value="localKeyword"
              allow-clear
              :disabled="loading || logs.length === 0"
              placeholder="搜索日志内容..."
              class="custom-input"
            >
              <template #prefix>
                <search-outlined />
              </template>
            </a-input>
          </div>

          <div class="filter-item">
            <label class="filter-label">级别</label>
            <a-select
              v-model:value="levelFilter"
              :disabled="loading || logs.length === 0"
              placeholder="全部"
              class="custom-select"
            >
              <a-select-option value="ALL">全部</a-select-option>
              <a-select-option value="ERROR">错误</a-select-option>
              <a-select-option value="WARN">警告</a-select-option>
              <a-select-option value="INFO">信息</a-select-option>
              <a-select-option value="DEBUG">调试</a-select-option>
            </a-select>
          </div>
        </div>

        <div class="action-row">
          <div class="left-actions">
            <a-button type="primary" class="action-btn primary" :loading="loading" @click="handleFetchData">
              <cloud-sync-outlined /> 查询
            </a-button>
            <a-button class="action-btn" @click="handleReset">
              <reload-outlined /> 重置
            </a-button>
          </div>

          <div class="right-actions">
            <a-tooltip title="自动刷新">
              <a-switch
                v-model:checked="autoRefresh"
                size="small"
                @change="handleAutoRefreshChange"
              />
            </a-tooltip>
            
            <a-dropdown :trigger="['click']" placement="bottomRight">
              <a-button class="action-btn icon-btn">
                <ellipsis-outlined />
              </a-button>
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1" @click="exportLogs" :disabled="!filteredLogs.length">
                    <download-outlined /> 导出CSV
                  </a-menu-item>
                  <a-menu-item key="2" @click="copyAllLogs" :disabled="!filteredLogs.length">
                    <copy-outlined /> 复制全部
                  </a-menu-item>
                  <a-menu-item key="3" @click="clearLogs" :disabled="!logs.length">
                    <delete-outlined /> 清空日志
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
          </div>
        </div>
      </div>
    </a-card>

    <!-- 结果展示卡片 -->
    <a-card class="modern-card result-section" :bordered="false">
      <!-- 统计面板 -->
      <transition name="fade-in">
        <div v-if="logSummary.total" class="stats-panel">
          <div class="stat-card total">
            <div class="stat-icon"><bars-outlined /></div>
            <div class="stat-content">
              <div class="stat-label">日志总数</div>
              <div class="stat-value">{{ logSummary.total }}</div>
            </div>
          </div>
          
          <div class="stat-card error" :class="{ 'has-error': logSummary.error > 0 }">
            <div class="stat-icon"><close-circle-outlined /></div>
            <div class="stat-content">
              <div class="stat-label">错误数</div>
              <div class="stat-value">{{ logSummary.error }}</div>
            </div>
          </div>

          <div class="stat-card success">
            <div class="stat-icon"><check-circle-outlined /></div>
            <div class="stat-content">
              <div class="stat-label">成功率</div>
              <div class="stat-value">{{ logSummary.successRate }}%</div>
            </div>
          </div>

          <div class="stat-card time">
            <div class="stat-icon"><clock-circle-outlined /></div>
            <div class="stat-content">
              <div class="stat-label">时间段</div>
              <div class="stat-value time-range">
                <span>{{ logSummary.earliestTime || '--' }}</span>
                <span class="separator">~</span>
                <span>{{ logSummary.latestTime || '--' }}</span>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <span class="result-info">
            <span v-if="localKeyword">筛选 <strong>{{ filteredLogs.length }}</strong> 条</span>
            <span v-else>共 <strong>{{ filteredLogs.length }}</strong> 条</span>
          </span>
          
          <a-radio-group v-model:value="levelFilter" button-style="solid" size="small" class="level-tabs">
            <a-radio-button value="ALL">全部</a-radio-button>
            <a-radio-button value="ERROR">错误({{ logSummary.error }})</a-radio-button>
            <a-radio-button value="WARN">警告({{ logSummary.warn }})</a-radio-button>
          </a-radio-group>
        </div>
        
        <div class="toolbar-right">
          <a-button type="text" size="small" class="icon-btn" @click="refresh" :loading="loading">
            <reload-outlined :spin="loading" />
          </a-button>
        </div>
      </div>

      <!-- PC端表格布局 -->
      <div class="table-wrapper" v-if="!showCardLayout">
        <a-table
          :data-source="filteredLogs"
          :columns="columns"
          :pagination="tablePagination"
          :scroll="tableScroll"
          size="middle"
          row-key="id"
          :row-class-name="getRowClassName"
          class="custom-table"
          :loading="loading"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.dataIndex === 'level'">
              <span class="level-badge" :class="getLevelConfig(record.level).class">
                <component :is="getLevelConfig(record.level).icon" />
                {{ record.level }}
              </span>
            </template>

            <template v-else-if="column.dataIndex === 'msg'">
              <div class="log-message">
                <a-typography-paragraph
                  class="log-text"
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
              <span class="time-badge">{{ record.time.split(' ')[1] || record.time }}</span>
            </template>
          </template>

          <template #emptyText>
            <div class="empty-placeholder">
              <img src="https://gw.alipayobjects.com/zos/antfincdn/ZHrcdLPrvN/empty.svg" alt="empty" />
              <p>{{ logs.length === 0 ? '暂无日志数据' : '没有找到匹配的结果' }}</p>
            </div>
          </template>
        </a-table>
      </div>

      <!-- 移动端卡片布局 -->
      <div class="mobile-layout" v-else>
        <div class="log-cards" v-if="mobilePagedLogs.length > 0">
          <div
            v-for="log in mobilePagedLogs"
            :key="log.id"
            class="log-card"
            :class="getRowClassName(log)"
          >
            <div class="card-header">
              <span class="card-time">
                <clock-circle-outlined />
                {{ log.time.split(' ')[1] || log.time }}
              </span>
              <span class="card-level" :class="getLevelConfig(log.level).class">
                <component :is="getLevelConfig(log.level).icon" />
                {{ log.level }}
              </span>
            </div>

            <div class="card-body">
              <a-typography-paragraph
                class="card-text"
                :ellipsis="{ rows: 3, expandable: true, symbol: '展开' }"
                :copyable="{ text: log.msg }"
                :content="log.msg"
              >
                <template v-if="localKeyword && log.msg">
                  <span v-html="highlightKeyword(log.msg)"></span>
                </template>
              </a-typography-paragraph>
            </div>

            <div class="card-footer">
              <span class="card-id">ID: {{ log.id }}</span>
            </div>
          </div>
        </div>

        <div class="empty-placeholder" v-else>
          <img src="https://gw.alipayobjects.com/zos/antfincdn/ZHrcdLPrvN/empty.svg" alt="empty" />
          <p>{{ logs.length === 0 ? '暂无日志数据' : '没有找到匹配的结果' }}</p>
        </div>

        <!-- 移动端分页 -->
        <div class="mobile-pagination" v-if="filteredLogs.length > 0">
          <a-pagination
            v-model:current="mobileCurrentPage"
            :page-size="mobilePageSize"
            :total="filteredLogs.length"
            size="small"
            show-less-items
            :show-total="total => `共 ${total} 条`"
            class="custom-pagination"
          />
        </div>
      </div>
    </a-card>
  </div>
</template>

<script setup>
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import locale from 'ant-design-vue/es/date-picker/locale/zh_CN'
import { computed, onMounted, onUnmounted, ref } from 'vue'
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
  FilterOutlined,
  SearchOutlined,
  CalendarOutlined,
  SettingOutlined,
  CopyOutlined,
  DeleteOutlined,
  DatabaseOutlined,
  EllipsisOutlined,
  SyncOutlined
} from '@ant-design/icons-vue'
// 请确保路径正确
import { apiMethods } from '@/utils/api'

// 设置 dayjs 为中文
dayjs.locale('zh-cn')

// --- State Definition ---
const datePickerLocale = ref(locale)
const localKeyword = ref('')
const selectedDate = ref(null)
const loading = ref(false)
const logs = ref([])
const levelFilter = ref('ALL')
const windowWidth = ref(window.innerWidth)
const mobileCurrentPage = ref(1)
const mobilePageSize = ref(10)
const autoRefresh = ref(false)
let refreshTimer = null 

// --- Table Config ---
const columns = [
  { title: '时间', dataIndex: 'time', key: 'time', width: 120, fixed: 'left', align: 'center' },
  { title: '级别', dataIndex: 'level', key: 'level', width: 110, align: 'center' },
  { title: '日志内容', dataIndex: 'msg', key: 'msg' }
]

const ellipsisConfig = { rows: 2, expandable: true, symbol: '展开' }
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
  showTotal: (total) => `共 ${total} 条`,
  size: 'default'
}))

// 响应式表格滚动高度
const tableScroll = computed(() => {
  const screenWidth = windowWidth.value
  let y = 'calc(100vh - 480px)' // 默认PC端高度

  if (screenWidth <= 992) {
    y = 'calc(100vh - 500px)' // 平板调整
  }

  if (screenWidth <= 768) {
    y = 'calc(100vh - 520px)' // 移动端调整
  }

  if (screenWidth <= 480) {
    y = 'calc(100vh - 500px)' // 小屏幕进一步调整
  }

  // 确保最小高度
  const minHeight = 200
  const calculatedHeight = Math.max(minHeight, parseInt(y) || minHeight)

  return {
    x: Math.min(800, screenWidth - 40), // 根据屏幕宽度调整水平滚动
    y: calculatedHeight
  }
})

// 是否显示卡片布局（移动端）
const showCardLayout = computed(() => {
  return windowWidth.value <= 768
})

// 移动端分页数据
const mobilePagedLogs = computed(() => {
  const start = (mobileCurrentPage.value - 1) * mobilePageSize.value
  const end = start + mobilePageSize.value
  return filteredLogs.value.slice(start, end)
})

// 快速筛选器
const quickFilters = computed(() => [
  {
    value: 'ALL',
    label: '全部',
    icon: BarsOutlined,
    color: 'default',
    count: logs.value.length
  },
  {
    value: 'ERROR',
    label: '错误',
    icon: CloseCircleOutlined,
    color: 'red',
    count: logSummary.value.error
  },
  {
    value: 'WARN',
    label: '警告',
    icon: WarningOutlined,
    color: 'orange',
    count: logSummary.value.warn
  },
  {
    value: 'INFO',
    label: '信息',
    icon: InfoCircleOutlined,
    color: 'blue',
    count: logs.value.length - logSummary.value.error - logSummary.value.warn
  }
])

// 应用快速筛选
const applyQuickFilter = (filterValue) => {
  levelFilter.value = filterValue
}

// 切换自动刷新
const toggleAutoRefresh = () => {
  autoRefresh.value = !autoRefresh.value
  handleAutoRefreshChange(autoRefresh.value)
}

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

// 复制到剪贴板
const copyToClipboard = async (text) => {
    try {
        await navigator.clipboard.writeText(text)
        message.success('已复制到剪贴板')
    } catch (err) {
        // 降级方案
        const textArea = document.createElement('textarea')
        textArea.value = text
        document.body.appendChild(textArea)
        textArea.select()
        try {
            document.execCommand('copy')
            message.success('已复制到剪贴板')
        } catch (err2) {
            message.error('复制失败')
        }
        document.body.removeChild(textArea)
    }
}

// 自动刷新处理
const handleAutoRefreshChange = (checked) => {
  if (checked) {
    refreshTimer = setInterval(() => {
      if (!loading.value) {
        fetchLogs()
      }
    }, 5000) // 5秒自动刷新
    message.info('已开启自动刷新（5秒）')
  } else {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
    message.info('已关闭自动刷新')
  }
}

// 复制所有日志
const copyAllLogs = async () => {
  const allText = filteredLogs.value.map(log =>
    `[${log.time}] ${log.level}: ${log.msg}`
  ).join('\n')

  await copyToClipboard(allText)
}

// 清空日志
const clearLogs = () => {
  logs.value = []
  message.success('已清空日志')
}

// 窗口大小变化处理
const handleResize = () => {
  windowWidth.value = window.innerWidth
}

onMounted(() => {
  fetchLogs()
  window.addEventListener('resize', handleResize)
  
  // 修复日期选择器样式 - 终极版
  const fixDatePickerStyles = () => {
    const applyFix = () => {
      // 修复所有日历面板
      const panels = document.querySelectorAll('.ant-picker-dropdown')
      panels.forEach(panel => {
        panel.setAttribute('style', panel.getAttribute('style') + '; z-index: 99999 !important;')
      })
      
      // 修复容器
      const containers = document.querySelectorAll('.ant-picker-panel-container, .ant-picker-panel, .ant-picker-date-panel, .ant-picker-body')
      containers.forEach(container => {
        container.setAttribute('style', container.getAttribute('style') + '; width: 280px !important; min-width: 280px !important;')
      })
      
      // 修复表格
      const tables = document.querySelectorAll('.ant-picker-dropdown table')
      tables.forEach(table => {
        table.setAttribute('style', 'width: 100% !important; min-width: 100% !important; table-layout: fixed !important; border-collapse: separate !important; border-spacing: 0 !important;')
      })
      
      // 修复 thead 和 tbody
      const groups = document.querySelectorAll('.ant-picker-dropdown thead, .ant-picker-dropdown tbody')
      groups.forEach(group => {
        group.setAttribute('style', 'width: 100% !important; display: table-row-group !important;')
      })
      
      // 修复表行
      const rows = document.querySelectorAll('.ant-picker-dropdown tr')
      rows.forEach(row => {
        row.setAttribute('style', 'width: 100% !important; display: table-row !important;')
      })
      
      // 修复单元格 - 最关键
      const cells = document.querySelectorAll('.ant-picker-dropdown th, .ant-picker-dropdown td')
      cells.forEach(cell => {
        cell.setAttribute('style', 'width: 14.285714% !important; min-width: 14.285714% !important; max-width: 14.285714% !important; display: table-cell !important; border: none !important; padding: 4px !important; margin: 0 !important; text-align: center !important; box-sizing: border-box !important;')
      })
      
      console.log('日历修复已执行:', {
        panels: panels.length,
        tables: tables.length,
        rows: rows.length,
        cells: cells.length
      })
    }
    
    // 立即执行
    setTimeout(applyFix, 0)
    setTimeout(applyFix, 100)
    setTimeout(applyFix, 300)
    setTimeout(applyFix, 500)
    
    // 监听DOM变化
    const observer = new MutationObserver(() => {
      applyFix()
    })
    
    observer.observe(document.body, {
      childList: true,
      subtree: true,
      attributes: false
    })
    
    // 持续检查
    const interval = setInterval(applyFix, 200)
    
    setTimeout(() => {
      clearInterval(interval)
      console.log('日历修复监听结束')
    }, 10000)
  }
  
  fixDatePickerStyles()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})
</script>

<style scoped>
/* ==================== 全局覆盖 - 防止CSS污染 ==================== */
/* 强制覆盖全局样式对日期选择器的影响 */
:deep(.ant-picker-dropdown .ant-picker-panel-container) {
  width: 280px !important;
  min-width: 280px !important;
  max-width: 320px !important;
}

:deep(.ant-picker-dropdown .ant-picker-panel) {
  width: 280px !important;
}

:deep(.ant-picker-dropdown .ant-picker-date-panel) {
  width: 280px !important;
}

:deep(.ant-picker-dropdown .ant-picker-content table) {
  width: 100% !important;
  table-layout: fixed !important;
}

:deep(.ant-picker-dropdown .ant-picker-content thead th),
:deep(.ant-picker-dropdown .ant-picker-content tbody td) {
  width: 14.285714% !important;
  display: table-cell !important;
}

/* 额外的全局覆盖 - 直接针对body下的弹窗 */
:deep(body .ant-picker-dropdown) {
  z-index: 99999 !important;
  position: fixed !important;
}

:deep(body .ant-picker-dropdown .ant-picker-panel-container) {
  width: 280px !important;
  min-width: 280px !important;
  z-index: 99999 !important;
}

:deep(body .ant-picker-dropdown .ant-picker-panel) {
  width: 280px !important;
}

:deep(body .ant-picker-dropdown .ant-picker-date-panel) {
  width: 280px !important;
}

:deep(body .ant-picker-dropdown .ant-picker-body) {
  width: 280px !important;
}

:deep(body .ant-picker-dropdown table) {
  width: 100% !important;
  table-layout: fixed !important;
  border-collapse: separate !important;
  border-spacing: 0 !important;
}

:deep(body .ant-picker-dropdown thead),
:deep(body .ant-picker-dropdown tbody) {
  width: 100% !important;
}

:deep(body .ant-picker-dropdown tr) {
  width: 100% !important;
  display: table-row !important;
}

:deep(body .ant-picker-dropdown th),
:deep(body .ant-picker-dropdown td) {
  width: 14.285714% !important;
  max-width: 14.285714% !important;
  display: table-cell !important;
  border: 1px solid #f0f0f0 !important;
  margin: 0 !important;
  padding: 4px !important;
}

/* ==================== 基础定义 ==================== */
.auto-log-page {
  min-height: 100vh;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow-x: hidden;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

/* 动态背景 */
.page-background {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden;
  z-index: 0;
  pointer-events: none;
}

.gradient-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.3;
  animation: float 15s ease-in-out infinite;
}

.orb-1 {
  width: 400px;
  height: 400px;
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  top: -100px;
  left: -100px;
  animation-delay: 0s;
}

.orb-2 {
  width: 500px;
  height: 500px;
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  bottom: -150px;
  right: -150px;
  animation-delay: -5s;
}

.orb-3 {
  width: 300px;
  height: 300px;
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation-delay: -10s;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0) scale(1);
  }
  33% {
    transform: translateY(-30px) scale(1.1);
  }
  66% {
    transform: translateY(20px) scale(0.9);
  }
}

/* ==================== 现代卡片 ==================== */
.modern-card {
  position: relative !important;
  z-index: 1 !important;
  background: rgba(255, 255, 255, 0.95) !important;
  backdrop-filter: blur(20px) !important;
  -webkit-backdrop-filter: blur(20px) !important;
  border: 1px solid rgba(255, 255, 255, 0.3) !important;
  border-radius: 24px !important;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15) !important;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;
  margin-bottom: 24px !important;
  overflow: visible !important;
}

.modern-card:hover {
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.2) !important;
  transform: translateY(-2px) !important;
}

.modern-card :deep(.ant-card-body) {
  padding: 24px !important;
  overflow: visible !important;
}

/* ==================== 搜索区域 ==================== */
.section-header {
  display: flex !important;
  justify-content: space-between !important;
  align-items: center !important;
  margin-bottom: 24px !important;
  padding-bottom: 20px !important;
  border-bottom: 2px solid rgba(102, 126, 234, 0.1) !important;
}

.header-left {
  display: flex !important;
  align-items: center !important;
  gap: 16px !important;
}

.header-icon {
  width: 56px !important;
  height: 56px !important;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  border-radius: 16px !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  color: white !important;
  font-size: 28px !important;
  box-shadow: 0 8px 20px rgba(102, 126, 234, 0.3) !important;
}

.header-title h2 {
  margin: 0 !important;
  font-size: 24px !important;
  font-weight: 700 !important;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  -webkit-background-clip: text !important;
  -webkit-text-fill-color: transparent !important;
  background-clip: text !important;
}

.header-title p {
  margin: 4px 0 0 0 !important;
  font-size: 14px !important;
  color: #64748b !important;
}

.header-badge {
  padding: 6px 16px !important;
  border-radius: 20px !important;
  font-weight: 600 !important;
  font-size: 13px !important;
  border: none !important;
  background: linear-gradient(135deg, #ff4d7d 0%, #ff6b9d 100%) !important;
  color: white !important;
  box-shadow: 0 4px 12px rgba(255, 77, 125, 0.3) !important;
}

/* 筛选表单 */
.filter-form {
  margin-top: 20px !important;
}

.filter-row {
  display: grid !important;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)) !important;
  gap: 16px !important;
  margin-bottom: 20px !important;
}

.filter-item {
  display: flex !important;
  flex-direction: column !important;
  gap: 8px !important;
}

.filter-label {
  font-size: 13px !important;
  font-weight: 600 !important;
  color: #475569 !important;
  margin: 0 !important;
}

/* 自定义输入框 - 隔离全局样式 */
.custom-input,
.custom-select,
.custom-date-picker {
  width: 100% !important;
}

.custom-input :deep(.ant-input),
.custom-select :deep(.ant-select-selector),
.custom-date-picker :deep(.ant-picker) {
  height: 38px !important;
  border-radius: 10px !important;
  border: 2px solid #e2e8f0 !important;
  background: white !important;
  transition: all 0.3s ease !important;
  font-size: 13px !important;
  padding: 6px 10px !important;
  box-shadow: none !important;
}

.custom-input :deep(.ant-input):hover,
.custom-select :deep(.ant-select-selector):hover,
.custom-date-picker :deep(.ant-picker):hover {
  border-color: #667eea !important;
}

.custom-input :deep(.ant-input):focus,
.custom-select :deep(.ant-select-selector):focus,
.custom-date-picker :deep(.ant-picker-focused) {
  border-color: #667eea !important;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1) !important;
}

.custom-input :deep(.ant-input-prefix) {
  color: #94a3b8 !important;
  margin-right: 8px !important;
}

/* 日期选择器特殊处理 */
.custom-date-picker :deep(.ant-picker-input) {
  display: flex !important;
  align-items: center !important;
  width: 100% !important;
}

.custom-date-picker :deep(.ant-picker-input > input) {
  flex: 1 !important;
  border: none !important;
  background: transparent !important;
  padding: 0 !important;
  font-size: 13px !important;
}

.custom-date-picker :deep(.ant-picker-suffix) {
  color: #94a3b8 !important;
  flex-shrink: 0 !important;
}

/* 日期选择器弹窗 - 强制完整宽度 */
.custom-date-picker :deep(.ant-picker-dropdown) {
  z-index: 99999 !important;
}

.custom-date-picker :deep(.ant-picker-panel-container) {
  border-radius: 12px !important;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12) !important;
  border: none !important;
  min-width: 280px !important;
  max-width: 320px !important;
  width: 280px !important;
  box-sizing: border-box !important;
  overflow: visible !important;
}

/* 日历面板整体 - 强制宽度 */
.custom-date-picker :deep(.ant-picker-panel) {
  font-size: 13px !important;
  width: 280px !important;
  min-width: 280px !important;
  max-width: 280px !important;
  box-sizing: border-box !important;
  overflow: visible !important;
}

.custom-date-picker :deep(.ant-picker-date-panel) {
  width: 280px !important;
  min-width: 280px !important;
  max-width: 280px !important;
  box-sizing: border-box !important;
  overflow: visible !important;
}

.custom-date-picker :deep(.ant-picker-header) {
  padding: 6px 8px !important;
  width: 100% !important;
  box-sizing: border-box !important;
}

.custom-date-picker :deep(.ant-picker-header-view) {
  font-size: 13px !important;
  font-weight: 600 !important;
}

.custom-date-picker :deep(.ant-picker-body) {
  padding: 6px !important;
  width: 100% !important;
  box-sizing: border-box !important;
}

/* 强制表格完整显示 */
.custom-date-picker :deep(.ant-picker-content) {
  width: 100% !important;
  min-width: 100% !important;
  table-layout: fixed !important;
  box-sizing: border-box !important;
}

.custom-date-picker :deep(.ant-picker-content table) {
  width: 100% !important;
  min-width: 100% !important;
  table-layout: fixed !important;
  border-collapse: collapse !important;
}

.custom-date-picker :deep(.ant-picker-content thead) {
  width: 100% !important;
}

.custom-date-picker :deep(.ant-picker-content tbody) {
  width: 100% !important;
}

.custom-date-picker :deep(.ant-picker-content tr) {
  width: 100% !important;
  display: table-row !important;
}

.custom-date-picker :deep(.ant-picker-content thead th) {
  font-size: 12px !important;
  padding: 4px 0 !important;
  width: 14.285714% !important;
  max-width: 14.285714% !important;
  min-width: 36px !important;
  text-align: center !important;
  display: table-cell !important;
}

.custom-date-picker :deep(.ant-picker-content tbody td) {
  width: 14.285714% !important;
  max-width: 14.285714% !important;
  min-width: 36px !important;
  text-align: center !important;
  display: table-cell !important;
}

.custom-date-picker :deep(.ant-picker-cell) {
  padding: 1px !important;
  display: table-cell !important;
}

.custom-date-picker :deep(.ant-picker-cell-inner) {
  width: 32px !important;
  height: 32px !important;
  line-height: 32px !important;
  font-size: 13px !important;
  border-radius: 6px !important;
  margin: 0 auto !important;
  display: block !important;
  text-align: center !important;
}

.custom-date-picker :deep(.ant-picker-footer) {
  padding: 6px 8px !important;
  border-top: 1px solid #f0f0f0 !important;
  width: 100% !important;
  box-sizing: border-box !important;
}

.custom-date-picker :deep(.ant-picker-today-btn) {
  font-size: 12px !important;
  color: #667eea !important;
}

/* 操作按钮行 */
.action-row {
  display: flex !important;
  justify-content: space-between !important;
  align-items: center !important;
  flex-wrap: wrap !important;
  gap: 12px !important;
  padding-top: 16px !important;
  border-top: 2px solid #f1f5f9 !important;
}

.left-actions,
.right-actions {
  display: flex !important;
  align-items: center !important;
  gap: 12px !important;
  flex-wrap: wrap !important;
}

.action-btn {
  height: 40px !important;
  padding: 0 20px !important;
  border-radius: 12px !important;
  font-weight: 600 !important;
  font-size: 14px !important;
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
  gap: 8px !important;
  border: none !important;
  transition: all 0.3s ease !important;
  box-shadow: none !important;
}

.action-btn.primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  color: white !important;
}

.action-btn.primary:hover {
  transform: translateY(-2px) !important;
  box-shadow: 0 8px 20px rgba(102, 126, 234, 0.3) !important;
}

.action-btn:not(.primary) {
  background: white !important;
  border: 2px solid #e2e8f0 !important;
  color: #64748b !important;
}

.action-btn:not(.primary):hover {
  border-color: #667eea !important;
  color: #667eea !important;
}

.action-btn.icon-btn {
  width: 40px !important;
  padding: 0 !important;
}

/* 弹出菜单 */
:deep(.ant-dropdown-menu) {
  border-radius: 12px !important;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15) !important;
  border: none !important;
  padding: 8px !important;
}

:deep(.ant-dropdown-menu-item) {
  border-radius: 8px !important;
  padding: 10px 14px !important;
  margin: 2px 0 !important;
  transition: all 0.2s !important;
}

:deep(.ant-dropdown-menu-item:hover) {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%) !important;
  color: #667eea !important;
}

/* 开关 */
:deep(.ant-switch) {
  background: #cbd5e1 !important;
}

:deep(.ant-switch-checked) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
}

/* ==================== 统计面板 ==================== */
.stats-panel {
  display: grid !important;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)) !important;
  gap: 16px !important;
  margin-bottom: 24px !important;
  animation: fadeIn 0.5s ease !important;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.stat-card {
  background: white !important;
  border-radius: 16px !important;
  padding: 20px !important;
  display: flex !important;
  align-items: center !important;
  gap: 16px !important;
  border: 2px solid transparent !important;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;
  position: relative !important;
  overflow: hidden !important;
}

.stat-card::before {
  content: '' !important;
  position: absolute !important;
  top: 0 !important;
  left: 0 !important;
  right: 0 !important;
  height: 4px !important;
  opacity: 0 !important;
  transition: opacity 0.3s !important;
}

.stat-card:hover {
  transform: translateY(-4px) !important;
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.1) !important;
}

.stat-card:hover::before {
  opacity: 1 !important;
}

.stat-card.total::before {
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%) !important;
}

.stat-card.error::before {
  background: linear-gradient(90deg, #f093fb 0%, #f5576c 100%) !important;
}

.stat-card.success::before {
  background: linear-gradient(90deg, #4facfe 0%, #00f2fe 100%) !important;
}

.stat-card.time::before {
  background: linear-gradient(90deg, #fa709a 0%, #fee140 100%) !important;
}

.stat-card.has-error {
  border-color: rgba(245, 87, 108, 0.3) !important;
  background: linear-gradient(135deg, rgba(240, 147, 251, 0.05) 0%, rgba(245, 87, 108, 0.05) 100%) !important;
}

.stat-icon {
  width: 56px !important;
  height: 56px !important;
  border-radius: 14px !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  font-size: 26px !important;
  flex-shrink: 0 !important;
  transition: transform 0.3s !important;
}

.stat-card:hover .stat-icon {
  transform: scale(1.1) rotate(5deg) !important;
}

.stat-card.total .stat-icon {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%) !important;
  color: #667eea !important;
}

.stat-card.error .stat-icon {
  background: linear-gradient(135deg, rgba(240, 147, 251, 0.1) 0%, rgba(245, 87, 108, 0.1) 100%) !important;
  color: #f5576c !important;
}

.stat-card.success .stat-icon {
  background: linear-gradient(135deg, rgba(79, 172, 254, 0.1) 0%, rgba(0, 242, 254, 0.1) 100%) !important;
  color: #4facfe !important;
}

.stat-card.time .stat-icon {
  background: linear-gradient(135deg, rgba(250, 112, 154, 0.1) 0%, rgba(254, 225, 64, 0.1) 100%) !important;
  color: #fa709a !important;
}

.stat-content {
  flex: 1 !important;
  min-width: 0 !important;
}

.stat-label {
  font-size: 13px !important;
  color: #64748b !important;
  font-weight: 600 !important;
  margin-bottom: 6px !important;
  text-transform: uppercase !important;
  letter-spacing: 0.5px !important;
}

.stat-value {
  font-size: 28px !important;
  font-weight: 800 !important;
  color: #1e293b !important;
  line-height: 1 !important;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif !important;
}

.stat-value.time-range {
  font-size: 16px !important;
  display: flex !important;
  flex-direction: column !important;
  gap: 4px !important;
  font-family: 'Courier New', monospace !important;
}

.stat-value.time-range .separator {
  display: none !important;
}

/* ==================== 工具栏 ==================== */
.toolbar {
  display: flex !important;
  justify-content: space-between !important;
  align-items: center !important;
  padding: 16px !important;
  background: #f8fafc !important;
  border-radius: 12px !important;
  margin-bottom: 20px !important;
  flex-wrap: wrap !important;
  gap: 12px !important;
}

.toolbar-left {
  display: flex !important;
  align-items: center !important;
  gap: 16px !important;
  flex-wrap: wrap !important;
}

.toolbar-right {
  display: flex !important;
  align-items: center !important;
  gap: 8px !important;
}

.result-info {
  font-size: 14px !important;
  color: #475569 !important;
  font-weight: 600 !important;
}

.result-info strong {
  color: #667eea !important;
  font-size: 18px !important;
  margin: 0 4px !important;
}

.level-tabs :deep(.ant-radio-button-wrapper) {
  height: 34px !important;
  line-height: 32px !important;
  padding: 0 16px !important;
  border: 2px solid #e2e8f0 !important;
  background: white !important;
  color: #64748b !important;
  font-weight: 600 !important;
  font-size: 13px !important;
  transition: all 0.3s !important;
  border-radius: 0 !important;
}

.level-tabs :deep(.ant-radio-button-wrapper:first-child) {
  border-radius: 10px 0 0 10px !important;
}

.level-tabs :deep(.ant-radio-button-wrapper:last-child) {
  border-radius: 0 10px 10px 0 !important;
}

.level-tabs :deep(.ant-radio-button-wrapper):not(:first-child) {
  border-left: none !important;
}

.level-tabs :deep(.ant-radio-button-wrapper:hover) {
  color: #667eea !important;
}

.level-tabs :deep(.ant-radio-button-wrapper-checked) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  border-color: #667eea !important;
  color: white !important;
  z-index: 1 !important;
}

.icon-btn {
  width: 36px !important;
  height: 36px !important;
  padding: 0 !important;
  border-radius: 10px !important;
  border: none !important;
  background: white !important;
  color: #64748b !important;
  transition: all 0.3s !important;
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
}

.icon-btn:hover {
  background: #667eea !important;
  color: white !important;
  transform: rotate(180deg) !important;
}

/* ==================== 表格样式 ==================== */
.table-wrapper {
  background: white !important;
  border-radius: 16px !important;
  overflow: hidden !important;
  border: 2px solid #f1f5f9 !important;
}

.custom-table :deep(.ant-table) {
  background: transparent !important;
}

.custom-table :deep(.ant-table-thead > tr > th) {
  background: #f8fafc !important;
  color: #475569 !important;
  font-weight: 700 !important;
  font-size: 13px !important;
  text-transform: uppercase !important;
  letter-spacing: 0.5px !important;
  border: none !important;
  border-bottom: 2px solid #667eea !important;
  padding: 16px !important;
}

.custom-table :deep(.ant-table-tbody > tr) {
  transition: all 0.3s !important;
  border: none !important;
}

.custom-table :deep(.ant-table-tbody > tr:hover) {
  background: rgba(102, 126, 234, 0.08) !important;
}

.custom-table :deep(.ant-table-tbody > tr > td) {
  padding: 14px 16px !important;
  border: none !important;
  border-bottom: 1px solid #f0f0f0 !important;
}

.custom-table :deep(.ant-table-tbody > tr:last-child > td) {
  border-bottom: none !important;
}

/* 错误/警告行 */
.custom-table :deep(.row-error) {
  background: linear-gradient(90deg, rgba(245, 87, 108, 0.05) 0%, transparent 100%) !important;
  border-left: 3px solid #f5576c !important;
}

.custom-table :deep(.row-warn) {
  background: linear-gradient(90deg, rgba(254, 225, 64, 0.05) 0%, transparent 100%) !important;
  border-left: 3px solid #fee140 !important;
}

/* 级别徽章 */
.level-badge {
  display: inline-flex !important;
  align-items: center !important;
  gap: 6px !important;
  padding: 6px 12px !important;
  border-radius: 20px !important;
  font-size: 12px !important;
  font-weight: 700 !important;
  transition: all 0.3s !important;
}

.level-badge.tag-info {
  background: rgba(79, 172, 254, 0.1) !important;
  color: #4facfe !important;
  border: 2px solid rgba(79, 172, 254, 0.2) !important;
}

.level-badge.tag-debug {
  background: rgba(148, 163, 184, 0.1) !important;
  color: #94a3b8 !important;
  border: 2px solid rgba(148, 163, 184, 0.2) !important;
}

.level-badge.tag-warn {
  background: rgba(254, 225, 64, 0.1) !important;
  color: #d97706 !important;
  border: 2px solid rgba(254, 225, 64, 0.2) !important;
}

.level-badge.tag-error {
  background: rgba(245, 87, 108, 0.1) !important;
  color: #f5576c !important;
  border: 2px solid rgba(245, 87, 108, 0.2) !important;
}

/* 时间徽章 */
.time-badge {
  display: inline-block !important;
  padding: 6px 12px !important;
  background: rgba(250, 112, 154, 0.1) !important;
  color: #fa709a !important;
  border-radius: 20px !important;
  font-family: 'Courier New', monospace !important;
  font-size: 13px !important;
  font-weight: 600 !important;
  border: 2px solid rgba(250, 112, 154, 0.2) !important;
}

/* 日志消息 */
.log-message {
  font-size: 14px !important;
  line-height: 1.6 !important;
  color: #334155 !important;
  font-family: 'Consolas', 'Monaco', monospace !important;
}

.log-message :deep(.ant-typography) {
  margin-bottom: 0 !important;
  color: inherit !important;
}

/* 空状态 */
.empty-placeholder {
  padding: 60px 20px !important;
  text-align: center !important;
}

.empty-placeholder img {
  height: 120px !important;
  opacity: 0.6 !important;
  margin-bottom: 16px !important;
}

.empty-placeholder p {
  color: #94a3b8 !important;
  font-size: 15px !important;
  margin: 0 !important;
}

/* 分页 */
.custom-table :deep(.ant-pagination) {
  margin: 20px 0 !important;
}

.custom-table :deep(.ant-pagination-item) {
  border-radius: 8px !important;
  border: 2px solid #e2e8f0 !important;
}

.custom-table :deep(.ant-pagination-item:hover) {
  border-color: #667eea !important;
}

.custom-table :deep(.ant-pagination-item-active) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  border-color: #667eea !important;
}

.custom-table :deep(.ant-pagination-item-active a) {
  color: white !important;
}

/* ==================== 移动端卡片布局 ==================== */
.mobile-layout {
  width: 100% !important;
}

.log-cards {
  display: flex !important;
  flex-direction: column !important;
  gap: 16px !important;
  margin-bottom: 20px !important;
}

.log-card {
  background: white !important;
  border-radius: 16px !important;
  padding: 16px !important;
  border: 2px solid #f1f5f9 !important;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;
  position: relative !important;
}

.log-card::before {
  content: '' !important;
  position: absolute !important;
  top: 0 !important;
  left: 0 !important;
  right: 0 !important;
  height: 4px !important;
  border-radius: 16px 16px 0 0 !important;
  opacity: 0 !important;
  transition: opacity 0.3s !important;
}

.log-card.row-error::before {
  background: linear-gradient(90deg, #f5576c 0%, #ff6b9d 100%) !important;
  opacity: 1 !important;
}

.log-card.row-warn::before {
  background: linear-gradient(90deg, #fee140 0%, #ffc53d 100%) !important;
  opacity: 1 !important;
}

.log-card:active {
  transform: scale(0.98) !important;
}

.card-header {
  display: flex !important;
  justify-content: space-between !important;
  align-items: center !important;
  margin-bottom: 12px !important;
  padding-bottom: 12px !important;
  border-bottom: 2px solid #f1f5f9 !important;
}

.card-time {
  display: inline-flex !important;
  align-items: center !important;
  gap: 6px !important;
  font-size: 13px !important;
  color: #fa709a !important;
  font-weight: 600 !important;
  font-family: 'Courier New', monospace !important;
  padding: 4px 10px !important;
  background: rgba(250, 112, 154, 0.1) !important;
  border-radius: 12px !important;
}

.card-level {
  display: inline-flex !important;
  align-items: center !important;
  gap: 6px !important;
  padding: 4px 10px !important;
  border-radius: 12px !important;
  font-size: 11px !important;
  font-weight: 700 !important;
  text-transform: uppercase !important;
}

.card-level.tag-info {
  background: rgba(79, 172, 254, 0.1) !important;
  color: #4facfe !important;
  border: 2px solid rgba(79, 172, 254, 0.2) !important;
}

.card-level.tag-debug {
  background: rgba(148, 163, 184, 0.1) !important;
  color: #94a3b8 !important;
  border: 2px solid rgba(148, 163, 184, 0.2) !important;
}

.card-level.tag-warn {
  background: rgba(254, 225, 64, 0.1) !important;
  color: #d97706 !important;
  border: 2px solid rgba(254, 225, 64, 0.2) !important;
}

.card-level.tag-error {
  background: rgba(245, 87, 108, 0.1) !important;
  color: #f5576c !important;
  border: 2px solid rgba(245, 87, 108, 0.2) !important;
}

.card-body {
  margin-bottom: 12px !important;
}

.card-text {
  font-size: 14px !important;
  line-height: 1.6 !important;
  color: #334155 !important;
  margin: 0 !important;
  font-family: 'Consolas', 'Monaco', monospace !important;
}

.card-text :deep(.ant-typography) {
  margin-bottom: 0 !important;
  color: inherit !important;
}

.card-footer {
  display: flex !important;
  justify-content: space-between !important;
  align-items: center !important;
  padding-top: 8px !important;
  border-top: 1px solid #f1f5f9 !important;
}

.card-id {
  font-size: 11px !important;
  color: #94a3b8 !important;
  font-family: 'Courier New', monospace !important;
}

/* 移动端分页 */
.mobile-pagination {
  display: flex !important;
  justify-content: center !important;
  padding: 20px 0 !important;
}

.custom-pagination :deep(.ant-pagination-item) {
  border-radius: 8px !important;
  border: 2px solid #e2e8f0 !important;
  min-width: 36px !important;
  height: 36px !important;
  line-height: 32px !important;
}

.custom-pagination :deep(.ant-pagination-item-active) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  border-color: #667eea !important;
}

.custom-pagination :deep(.ant-pagination-item-active a) {
  color: white !important;
}

.custom-pagination :deep(.ant-pagination-prev),
.custom-pagination :deep(.ant-pagination-next) {
  border-radius: 8px !important;
  border: 2px solid #5193e9 !important;
  min-width: 36px !important;
  height: 36px !important;
  line-height: 32px !important;
}

/* ==================== 动画效果 ==================== */
@keyframes fade-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.fade-in-enter-active {
  animation: fade-in 0.4s ease !important;
}

/* ==================== 高亮显示 ==================== */
:deep(.highlight) {
  background: linear-gradient(135deg, rgba(255, 77, 125, 0.2) 0%, rgba(255, 107, 157, 0.2) 100%) !important;
  color: #f5576c !important;
  font-weight: 700 !important;
  padding: 2px 4px !important;
  border-radius: 4px !important;
  border: 1px solid rgba(245, 87, 108, 0.3) !important;
}

/* ==================== 响应式适配 ==================== */

/* 大屏幕 */
@media (min-width: 1400px) {
  .auto-log-page {
    padding: 32px !important;
  }

  .modern-card :deep(.ant-card-body) {
    padding: 32px !important;
  }

  .filter-row {
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)) !important;
    gap: 20px !important;
  }

  .stats-panel {
    grid-template-columns: repeat(4, 1fr) !important;
    gap: 20px !important;
  }
}

/* 平板 */
@media (max-width: 992px) {
  .auto-log-page {
    padding: 16px !important;
  }

  .modern-card {
    border-radius: 20px !important;
    margin-bottom: 16px !important;
  }

  .modern-card :deep(.ant-card-body) {
    padding: 20px !important;
  }

  .section-header {
    flex-direction: column !important;
    align-items: flex-start !important;
    gap: 12px !important;
  }

  .header-badge {
    align-self: flex-start !important;
  }

  .filter-row {
    grid-template-columns: 1fr !important;
    gap: 12px !important;
  }

  .action-row {
    flex-direction: column !important;
    align-items: stretch !important;
  }

  .left-actions,
  .right-actions {
    width: 100% !important;
    justify-content: center !important;
  }

  .stats-panel {
    grid-template-columns: repeat(2, 1fr) !important;
    gap: 12px !important;
  }

  .stat-value.time-range {
    font-size: 13px !important;
  }

  .toolbar {
    flex-direction: column !important;
    align-items: stretch !important;
  }

  .toolbar-left,
  .toolbar-right {
    justify-content: center !important;
  }

  .level-tabs {
    width: 100% !important;
  }

  .level-tabs :deep(.ant-radio-button-wrapper) {
    flex: 1 !important;
    text-align: center !important;
  }
}

/* 手机 */
@media (max-width: 768px) {
  .auto-log-page {
    padding: 12px !important;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  }

  .orb-1,
  .orb-2,
  .orb-3 {
    opacity: 0.2 !important;
  }

  .modern-card {
    border-radius: 16px !important;
    margin-bottom: 12px !important;
  }

  .modern-card :deep(.ant-card-body) {
    padding: 16px !important;
  }

  .section-header {
    padding-bottom: 16px !important;
  }

  .header-icon {
    width: 48px !important;
    height: 48px !important;
    font-size: 24px !important;
  }

  .header-title h2 {
    font-size: 20px !important;
  }

  .header-title p {
    font-size: 13px !important;
  }

  .filter-label {
    font-size: 12px !important;
  }

  .custom-input :deep(.ant-input),
  .custom-select :deep(.ant-select-selector),
  .custom-date-picker :deep(.ant-picker) {
    height: 40px !important;
    font-size: 14px !important;
  }

  .action-btn {
    height: 42px !important;
    width: 100% !important;
    justify-content: center !important;
  }

  .stats-panel {
    grid-template-columns: 1fr !important;
    gap: 12px !important;
  }

  .stat-card {
    padding: 16px !important;
  }

  .stat-icon {
    width: 48px !important;
    height: 48px !important;
    font-size: 22px !important;
  }

  .stat-value {
    font-size: 24px !important;
  }

  .stat-value.time-range {
    font-size: 12px !important;
  }

  .toolbar {
    padding: 12px !important;
  }

  .result-info {
    font-size: 13px !important;
  }

  .result-info strong {
    font-size: 16px !important;
  }

  .level-tabs :deep(.ant-radio-button-wrapper) {
    padding: 0 12px !important;
    font-size: 12px !important;
  }
}

/* 小屏手机 */
@media (max-width: 480px) {
  .auto-log-page {
    padding: 8px !important;
  }

  .modern-card :deep(.ant-card-body) {
    padding: 12px !important;
  }

  .header-left {
    gap: 12px !important;
  }

  .header-icon {
    width: 40px !important;
    height: 40px !important;
    font-size: 20px !important;
    border-radius: 12px !important;
  }

  .header-title h2 {
    font-size: 18px !important;
  }

  .header-title p {
    font-size: 12px !important;
  }

  .action-btn {
    font-size: 13px !important;
    padding: 0 16px !important;
  }

  .stat-label {
    font-size: 11px !important;
  }

  .stat-value {
    font-size: 20px !important;
  }

  .log-card {
    padding: 12px !important;
  }

  .card-text {
    font-size: 13px !important;
  }
}

/* 小屏手机 - 日期选择器额外优化 */
@media (max-width: 480px) {
  .custom-date-picker :deep(.ant-picker-panel-container) {
    margin: 0 !important;
  }

  .custom-date-picker :deep(.ant-picker-date-panel) {
    max-width: 280px !important;
  }

  .custom-date-picker :deep(.ant-picker-cell-inner) {
    width: 32px !important;
    height: 32px !important;
    line-height: 32px !important;
    font-size: 13px !important;
  }

  .custom-date-picker :deep(.ant-picker-content thead th) {
    font-size: 11px !important;
    padding: 2px !important;
  }

  .custom-date-picker :deep(.ant-picker-month-panel .ant-picker-cell-inner),
  .custom-date-picker :deep(.ant-picker-year-panel .ant-picker-cell-inner) {
    width: 50px !important;
    height: 36px !important;
    line-height: 36px !important;
    font-size: 12px !important;
  }

  .custom-date-picker :deep(.ant-picker-header-view) {
    font-size: 13px !important;
  }

  .custom-date-picker :deep(.ant-picker-body) {
    padding: 4px !important;
  }
}

/* 移动端日期选择器弹窗适配 */
@media (max-width: 768px) {
  .custom-date-picker :deep(.ant-picker-dropdown) {
    max-width: calc(100vw - 20px) !important;
    left: 10px !important;
    right: 10px !important;
  }

  .custom-date-picker :deep(.ant-picker-panel-container) {
    max-width: 100% !important;
    min-width: auto !important;
    overflow: hidden !important;
  }

  .custom-date-picker :deep(.ant-picker-date-panel) {
    width: 100% !important;
    max-width: 320px !important;
    margin: 0 auto !important;
  }

  .custom-date-picker :deep(.ant-picker-month-panel),
  .custom-date-picker :deep(.ant-picker-year-panel) {
    max-width: 100% !important;
    overflow: hidden !important;
  }

  /* 日历内容自适应 */
  .custom-date-picker :deep(.ant-picker-body) {
    padding: 8px !important;
  }

  .custom-date-picker :deep(.ant-picker-content) {
    width: 100% !important;
    overflow-x: hidden !important;
  }

  .custom-date-picker :deep(.ant-picker-content table) {
    width: 100% !important;
  }

  /* 日历单元格 */
  .custom-date-picker :deep(.ant-picker-cell) {
    padding: 2px !important;
  }

  .custom-date-picker :deep(.ant-picker-cell-inner) {
    width: 36px !important;
    height: 36px !important;
    line-height: 36px !important;
    font-size: 14px !important;
  }

  /* 星期标题行 */
  .custom-date-picker :deep(.ant-picker-content thead th) {
    font-size: 12px !important;
    padding: 4px !important;
  }

  /* 月份选择器自适应 */
  .custom-date-picker :deep(.ant-picker-month-panel .ant-picker-cell-inner) {
    width: 60px !important;
    font-size: 13px !important;
  }

  /* 年份选择器自适应 */
  .custom-date-picker :deep(.ant-picker-year-panel .ant-picker-cell-inner) {
    width: 60px !important;
    font-size: 13px !important;
  }

  /* 头部按钮区域 */
  .custom-date-picker :deep(.ant-picker-header) {
    padding: 8px !important;
  }

  .custom-date-picker :deep(.ant-picker-header-view) {
    font-size: 14px !important;
  }

  /* 底部按钮区域 */
  .custom-date-picker :deep(.ant-picker-footer) {
    padding: 8px !important;
  }

  .custom-date-picker :deep(.ant-picker-footer .ant-picker-today-btn) {
    font-size: 13px !important;
  }
}

/* 触摸设备优化 */
@media (hover: none) and (pointer: coarse) {
  .modern-card:hover {
    transform: none !important;
  }

  .stat-card:hover {
    transform: none !important;
  }

  .stat-card:hover .stat-icon {
    transform: none !important;
  }

  .action-btn.primary:hover,
  .action-btn:not(.primary):hover {
    transform: none !important;
  }

  .log-card:active {
    transform: scale(0.98) !important;
  }

  .action-btn:active {
    transform: scale(0.97) !important;
  }

  /* 改善触摸目标大小 */
  .action-btn,
  .icon-btn,
  .custom-pagination :deep(.ant-pagination-item) {
    min-height: 44px !important;
    min-width: 44px !important;
  }
}

/* 减少动画模式 */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
</style>