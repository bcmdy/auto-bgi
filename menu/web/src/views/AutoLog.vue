<template>
  <div class="auto-log-page">
    <div class="page-ornament" aria-hidden="true"></div>

    <a-card class="search-card" :bordered="false">
      <template #title>
        <div class="card-title">
          <span>日志查询</span>
          <a-tag color="pink" class="title-tag">自动任务</a-tag>
        </div>
      </template>

      <a-form layout="inline" class="search-form" @submit.prevent>
        <a-form-item label="日期">
          <a-date-picker
            v-model:value="selectedDate"
            value-format="YYYY-MM-DD"
            format="YYYY-MM-DD"
            :disabled="loading"
            allow-clear
            placeholder="默认今天"
            @change="handleDateChange"
          />
        </a-form-item>
        <a-form-item label="关键词">
          <a-input
            v-model:value="keyword"
            allow-clear
            :disabled="loading"
            placeholder="请输入日期（默认今天）或关键字"
            @pressEnter="handleSearch"
          />
        </a-form-item>
        <a-form-item class="search-actions">
          <a-space :size="12" wrap>
            <a-button type="primary" :loading="loading" @click="handleSearch">查询</a-button>
            <a-button :disabled="loading || !keyword" @click="handleReset">重置</a-button>
            <a-button :disabled="loading" @click="refresh">刷新</a-button>
          </a-space>
        </a-form-item>
      </a-form>

      <a-alert
        show-icon
        type="info"
        class="helper-alert"
        message="支持日期或关键词模糊检索，输入留空将展示最近的自动任务日志。"
      />
    </a-card>

    <a-card class="result-card" :bordered="false">
      <template #title>
        <div class="card-title">
          <span>查询结果</span>
          <a-badge :count="filteredLogs.length" class="badge" />
        </div>
      </template>

      <div v-if="logSummary.total" class="summary-grid">
        <div class="summary-card total">
          <p class="label">日志总数</p>
          <p class="value">{{ logSummary.total }}</p>
          <span class="desc">含 {{ logSummary.error }} 条错误 / {{ logSummary.warn }} 条警告</span>
        </div>
        <div class="summary-card success">
          <p class="label">成功率</p>
          <p class="value">{{ logSummary.successRate }}%</p>
          <span class="desc">以 ERROR 数量估算</span>
        </div>
        <div class="summary-card timeline">
          <p class="label">时间范围</p>
          <p class="value">{{ logSummary.earliestTime || '-' }}</p>
          <span class="desc">最新：{{ logSummary.latestTime || '-' }}</span>
        </div>
      </div>

      <div v-if="logSummary.total" class="filter-bar">
        <a-radio-group
          v-model:value="levelFilter"
          :options="levelFilterOptions"
          option-type="button"
          button-style="solid"
        />
        <a-tag v-if="levelFilter !== 'ALL'" color="blue" class="filter-tag">
          当前筛选：{{ levelFilterLabel }}
        </a-tag>
      </div>

      <a-spin :spinning="loading">
        <a-empty v-if="!loading && filteredLogs.length === 0" description="暂无符合条件的日志" />
        <div v-else class="table-wrapper">
          <a-table
            :data-source="filteredLogs"
            :columns="columns"
            :pagination="tablePagination"
            :scroll="tableScroll"
            size="middle"
            row-key="id"
            :row-class-name="rowClassName"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.dataIndex === 'level'">
                <a-tag :color="getLevelColor(record.level)" class="level-tag">
                  {{ record.level || '未知' }}
                </a-tag>
              </template>
              <template v-else-if="column.dataIndex === 'msg'">
                <a-typography-paragraph
                  class="log-message"
                  :ellipsis="ellipsisConfig"
                  :copyable="copyConfig(record.msg)"
                >
                  {{ record.msg }}
                </a-typography-paragraph>
              </template>
              <template v-else>
                <span>{{ record[column.dataIndex] || '-' }}</span>
              </template>
            </template>
          </a-table>
        </div>
      </a-spin>
    </a-card>
  </div>
</template>

<script setup>
import dayjs from 'dayjs'
import { computed, onMounted, ref } from 'vue'
import { message } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'

const keyword = ref('')
const selectedDate = ref(null)
const loading = ref(false)
const logs = ref([])
const levelFilter = ref('ALL')

const levelFilterOptions = [
  { label: '全部', value: 'ALL' },
  { label: '仅错误', value: 'ERROR' },
  { label: '仅警告', value: 'WARN' },
  { label: '仅通知', value: 'INFO' }
]

const levelFilterLabel = computed(() => {
  const current = levelFilterOptions.find((item) => item.value === levelFilter.value)
  return current?.label || '全部'
})

const columns = [
  { title: '时间', dataIndex: 'time', key: 'time', width: 200 },
  { title: '级别', dataIndex: 'level', key: 'level', width: 120 },
  { title: '内容', dataIndex: 'msg', key: 'msg' }
]

const filteredLogs = computed(() => {
  if (levelFilter.value === 'ALL') return logs.value
  const target = levelFilter.value.toUpperCase()
  return logs.value.filter((log) => (log.level || '').toUpperCase().includes(target))
})

const tablePagination = computed(() => ({
  pageSize: 20,
  showSizeChanger: false,
  total: filteredLogs.value.length,
  showTotal: (total) => `共 ${total} 条`
}))

const tableScroll = computed(() => ({
  x: 720,
  y:
    typeof window !== 'undefined' && window.innerHeight
      ? Math.max(window.innerHeight - 360, 320)
      : 360
}))

const logSummary = computed(() => {
  if (!logs.value.length) {
    return {
      total: 0,
      error: 0,
      warn: 0,
      info: 0,
      latestTime: '',
      earliestTime: '',
      successRate: '0.0'
    }
  }

  const summary = {
    total: logs.value.length,
    error: 0,
    warn: 0,
    info: 0,
    latestTime: '',
    earliestTime: '',
    successRate: '0.0'
  }

  const validTimes = []

  logs.value.forEach((log) => {
    const level = (log.level || '').toUpperCase()
    if (level.includes('ERROR') || level.includes('ERR')) {
      summary.error += 1
    } else if (level.includes('WARN')) {
      summary.warn += 1
    } else {
      summary.info += 1
    }

    if (log.time) {
      const parsed = dayjs(log.time)
      if (parsed.isValid()) {
        validTimes.push(parsed)
      }
    }
  })

  if (validTimes.length) {
    validTimes.sort((a, b) => a.valueOf() - b.valueOf())
    summary.earliestTime = validTimes[0].format('YYYY-MM-DD HH:mm:ss')
    summary.latestTime = validTimes[validTimes.length - 1].format('YYYY-MM-DD HH:mm:ss')
  }

  summary.successRate = (((summary.total - summary.error) / summary.total) * 100).toFixed(1)

  return summary
})

const ellipsisConfig = {
  rows: 2,
  expandable: true,
  symbol: '展开'
}

const copyConfig = (text) => (text ? { text, tooltips: ['复制', '复制成功'] } : false)

const getLevelColor = (level = '') => {
  const upper = level.toUpperCase()
  if (upper.includes('ERROR') || upper.includes('ERR')) return 'red'
  if (upper.includes('WARN')) return 'orange'
  if (upper.includes('DEBUG')) return 'purple'
  return 'blue'
}

const rowClassName = (record) => {
  const upper = (record?.level || '').toUpperCase()
  if (upper.includes('ERROR') || upper.includes('ERR')) return 'row-error'
  if (upper.includes('WARN')) return 'row-warn'
  if (upper.includes('INFO')) return 'row-info'
  return ''
}

const handleDateChange = (value) => {
  selectedDate.value = value
  keyword.value = value || ''
}

const normalizeLogs = (raw) => {
  if (!raw) return []

  if (Array.isArray(raw)) {
    return raw
  }

  if (typeof raw === 'string') {
    const trimmed = raw.trim()
    if (!trimmed) return []

    try {
      const parsed = JSON.parse(trimmed)
      if (Array.isArray(parsed)) return parsed
      if (parsed && typeof parsed === 'object') return [parsed]
    } catch (e) {
      // ignore parse error, fallback to line split
    }

    return trimmed
      .split(/\r?\n/)
      .map((line) => line.trim())
      .filter(Boolean)
      .map((line) => {
        try {
          const parsedLine = JSON.parse(line)
          return parsedLine
        } catch (err) {
          return { msg: line }
        }
      })
  }

  if (typeof raw === 'object') {
    return [raw]
  }

  return []
}

const enhanceLogs = (items) =>
  items.map((item, index) => {
    let parsed = item

    if (typeof item === 'string') {
      try {
        parsed = JSON.parse(item)
      } catch (err) {
        parsed = { msg: item }
      }
    }

    if (parsed && typeof parsed === 'object') {
      return {
        id: `${parsed.time || 'log'}-${index}`,
        time: parsed.time || parsed.timestamp || '',
        level: (parsed.level || parsed.Level || '').toString().toUpperCase(),
        msg:
          parsed.msg ||
          parsed.message ||
          parsed.content ||
          (typeof parsed === 'object' ? JSON.stringify(parsed) : String(parsed))
      }
    }

    return {
      id: `log-${index}`,
      time: '',
      level: '',
      msg: String(item)
    }
  })

const fetchLogs = async () => {
  loading.value = true
  try {
    const response = await apiMethods.queryAutoLogs(keyword.value)
    if (response?.status === 'success') {
      const rawLogs = normalizeLogs(response.msg)
      logs.value = enhanceLogs(rawLogs)
    } else {
      logs.value = []
      const errorMsg = response?.msg || '查询日志失败，请稍后重试'
      message.error(errorMsg)
    }
  } catch (error) {
    logs.value = []
    message.error(error?.message || '查询日志失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  fetchLogs()
}

const handleReset = () => {
  keyword.value = ''
  selectedDate.value = null
  levelFilter.value = 'ALL'
  fetchLogs()
}

const refresh = () => {
  fetchLogs()
}

onMounted(() => {
  fetchLogs()
})
</script>

<style scoped>
.auto-log-page {
  position: relative;
  min-height: 100vh;
  padding: 32px;
  background: linear-gradient(135deg, #ffe1f1 0%, #fff5fb 100%);
  box-sizing: border-box;
  overflow: hidden;
}

.page-ornament {
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 20% 20%, rgba(255, 192, 203, 0.6), transparent 60%),
    radial-gradient(circle at 80% 0%, rgba(147, 197, 253, 0.4), transparent 50%),
    radial-gradient(circle at 50% 80%, rgba(255, 255, 255, 0.6), transparent 60%);
  filter: blur(60px);
  opacity: 0.5;
  pointer-events: none;
}

.auto-log-page > :not(.page-ornament) {
  position: relative;
  z-index: 1;
}

.search-card,
.result-card {
  max-width: 1200px;
  margin: 0 auto 24px;
  background-color: rgba(255, 255, 255, 0.88);
  box-shadow: 0 20px 50px rgba(255, 153, 204, 0.2);
  border-radius: 20px;
  backdrop-filter: blur(16px);
}

.card-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-weight: 600;
  font-size: 18px;
}

.title-tag {
  border-radius: 12px;
}

.search-form {
  row-gap: 16px;
  width: 100%;
}

.search-actions {
  margin-left: auto;
}

.helper-alert {
  margin-top: 12px;
  border-radius: 12px;
  background: rgba(24, 144, 255, 0.08);
  border: none;
}

.badge {
  margin-left: auto;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.summary-card {
  padding: 16px 20px;
  border-radius: 16px;
  background: linear-gradient(135deg, #fff 0%, #ffe6f2 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.4);
}

.summary-card.total {
  border: 1px solid rgba(255, 122, 198, 0.3);
}

.summary-card.success {
  background: linear-gradient(135deg, #e6fffb 0%, #f0fffe 100%);
  border: 1px solid rgba(47, 200, 186, 0.3);
}

.summary-card.timeline {
  background: linear-gradient(135deg, #edf4ff 0%, #f7fbff 100%);
  border: 1px solid rgba(124, 170, 255, 0.3);
}

.summary-card .label {
  margin: 0;
  color: #7a7a7a;
  font-size: 13px;
}

.summary-card .value {
  margin: 8px 0 4px;
  font-size: 28px;
  font-weight: 600;
  color: #333;
}

.summary-card .desc {
  color: #9b9b9b;
  font-size: 13px;
}

.filter-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
}

.filter-tag {
  border-radius: 999px;
}

.table-wrapper {
  width: 100%;
}

.level-tag {
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.log-message {
  margin-bottom: 0;
  word-break: break-all;
  font-size: 13px;
  line-height: 1.6;
}

:deep(.ant-radio-button-wrapper) {
  border-radius: 999px !important;
  border: none !important;
  background: rgba(0, 0, 0, 0.04);
  margin-right: 4px;
}

:deep(.ant-radio-button-wrapper-checked:not(.ant-radio-button-wrapper-disabled)) {
  background: #ff8ac0;
  color: #fff;
  box-shadow: 0 8px 20px rgba(255, 138, 192, 0.3);
}

:deep(.row-error .ant-table-cell) {
  background: rgba(255, 99, 132, 0.1) !important;
}

:deep(.row-warn .ant-table-cell) {
  background: rgba(250, 173, 20, 0.12) !important;
}

:deep(.row-info .ant-table-cell) {
  background: rgba(24, 144, 255, 0.08) !important;
}

:deep(.ant-table-tbody > tr.row-error:hover > td),
:deep(.ant-table-tbody > tr.row-warn:hover > td),
:deep(.ant-table-tbody > tr.row-info:hover > td) {
  background: rgba(255, 255, 255, 0.6) !important;
}

@media (max-width: 768px) {
  .auto-log-page {
    padding: 16px;
  }

  .search-card,
  .result-card {
    padding: 12px;
  }

  .search-actions {
    width: 100%;
    margin-left: 0;
  }

  .search-actions :deep(.ant-space) {
    width: 100%;
  }

  .search-actions :deep(.ant-btn) {
    flex: 1;
  }

  .table-wrapper {
    margin: 0 -12px;
    padding-bottom: 12px;
    overflow-x: auto;
  }

  .table-wrapper :deep(.ant-table) {
    min-width: 680px;
  }
}
</style>
