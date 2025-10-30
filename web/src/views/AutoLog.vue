<template>
  <div class="auto-log-page">
    <a-card class="search-card" :bordered="false">
      <template #title>
        <div class="card-title">
          <span>日志查询</span>
      
        </div>
      </template>

      <a-form layout="inline" @submit.prevent>
        <a-form-item label="日期">
          <a-input
            v-model:value="keyword"
            allow-clear
            placeholder="请输入日期（默认今天）"
            @pressEnter="handleSearch"
          />
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" :loading="loading" @click="handleSearch">
              查询
            </a-button>
            <a-button :disabled="loading || !keyword.value" @click="handleReset">
              重置
            </a-button>
            <a-button :disabled="loading" @click="refresh">
              刷新
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>

      <!-- <a-alert
        show-icon
        type="info"
        message="支持根据关键词模糊匹配日志内容，查询留空则展示最近记录。"
      /> -->
    </a-card>

    <a-card class="result-card" :bordered="false">
      <template #title>
        <div class="card-title">
          <span>查询结果</span>
          <a-badge :count="logs.length" class="badge" />
        </div>
      </template>
      <a-spin :spinning="loading">
        <a-empty v-if="!loading && logs.length === 0" description="暂无日志数据" />
        <a-table
          v-else
          :data-source="logs"
          :columns="columns"
          :pagination="tablePagination"
          :scroll="tableScroll"
          size="middle"
          row-key="id"
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
      </a-spin>
    </a-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { message } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'

const keyword = ref('')
const loading = ref(false)
const logs = ref([])

const columns = [
  { title: '时间', dataIndex: 'time', key: 'time', width: 200 },
  { title: '级别', dataIndex: 'level', key: 'level', width: 120 },
  { title: '内容', dataIndex: 'msg', key: 'msg' }
]

const tablePagination = computed(() => ({
  pageSize: 20,
  showSizeChanger: false,
  showTotal: (total) => `共 ${total} 条`
}))

const tableScroll = computed(() => ({
  x: '100%',
  y: window.innerHeight ? Math.max(window.innerHeight - 360, 320) : 360
}))

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
  min-height: 100vh;
  padding: 32px;
  background: linear-gradient(135deg, #ffe1f1 0%, #fff5fb 100%);
  box-sizing: border-box;
}

.search-card,
.result-card {
  max-width: 1200px;
  margin: 0 auto 24px;
  background-color: rgba(255, 255, 255, 0.85);
  box-shadow: 0 10px 30px rgba(255, 153, 204, 0.2);
  border-radius: 18px;
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

.badge {
  margin-left: auto;
}

.ant-form-item {
  margin-bottom: 16px;
}

.level-tag {
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.log-message {
  margin-bottom: 0;
  word-break: break-all;
}

@media (max-width: 768px) {
  .auto-log-page {
    padding: 16px;
  }

  .search-card,
  .result-card {
    padding: 12px;
  }

  .ant-form {
    flex-direction: column;
  }

  .ant-form-item {
    width: 100%;
    margin-right: 0 !important;
  }

  .ant-form-item .ant-input,
  .ant-form-item .ant-btn {
    width: 100%;
  }
}
</style>
