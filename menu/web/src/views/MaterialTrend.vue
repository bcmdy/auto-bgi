<template>
  <div class="material-trend-page">
    <a-page-header
      class="trend-header"
      title="材料变化图"
      sub-title="根据背包统计记录查看数量走势"
      @back="$router.back()"
    />

    <a-card class="controls-card" bordered>
      <div class="controls-row">
        <div class="select-wrapper">
          <span class="control-label">选择材料：</span>
          <a-select
            v-model:value="selectedMaterial"
            placeholder="请选择材料"
            style="min-width: 220px"
            :loading="loading"
            :options="materialOptions.map(material => ({ label: material, value: material }))"
          />
        </div>
        <a-space size="middle">
          <a-button type="primary" @click="refresh" :loading="loading">
            重新获取
          </a-button>
          <a-button @click="goBackStatistics">
            返回材料表格
          </a-button>
        </a-space>
      </div>
    </a-card>

    <a-row :gutter="16" class="summary-row">
      <a-col :xs="24" :sm="8">
        <a-card class="stat-card" bordered>
          <div class="stat-label">最新数量</div>
          <div class="stat-value">{{ latestValueDisplay }}</div>
          <div class="stat-hint">{{ latestDateDisplay }}</div>
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="8">
        <a-card class="stat-card" bordered>
          <div class="stat-label">起始数量</div>
          <div class="stat-value">{{ firstValueDisplay }}</div>
          <div class="stat-hint">{{ firstDateDisplay }}</div>
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="8">
        <a-card class="stat-card" bordered>
          <div class="stat-label">区间变化</div>
          <div
            class="stat-value"
            :class="{
              up: deltaValue > 0,
              down: deltaValue < 0
            }"
          >
            {{ deltaPrefix }}{{ numberFormatter.format(Math.abs(deltaValue)) }}
          </div>
          <div class="stat-hint">
            {{ deltaValue === 0 ? '数量稳定' : `变化率 ${deltaRate}` }}
          </div>
        </a-card>
      </a-col>
    </a-row>

    <a-card class="chart-card" bordered>
      <template #title>
        <span>趋势图</span>
      </template>

      <a-spin :spinning="loading">
        <div class="chart-wrapper">
          <div ref="chartRef" class="chart-container" />
          <a-empty
            v-if="!loading && !filteredSeries.length"
            class="chart-empty"
            description="暂无该材料的统计数据"
          />
        </div>
      </a-spin>
    </a-card>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import * as echarts from 'echarts'
import { useRouter } from 'vue-router'
import { apiMethods } from '@/utils/api'

const loading = ref(false)
const statistics = ref([])
const selectedMaterial = ref('')
const chartRef = ref(null)
let chartInstance = null
const router = useRouter()

const numberFormatter = new Intl.NumberFormat('zh-CN')
const riseColor = '#ff4d4f'
const fallColor = '#52c41a'
const flatColor = '#ff4d4f'

const materialOptions = computed(() => {
  const set = new Set()
  statistics.value.forEach(item => {
    if (item.material) {
      set.add(item.material)
    }
  })
  return Array.from(set)
})

const filteredSeries = computed(() => {
  if (!selectedMaterial.value) {
    return []
  }
  return statistics.value.filter(item => item.material === selectedMaterial.value)
})

const latestPoint = computed(() => filteredSeries.value.at(-1))
const firstPoint = computed(() => filteredSeries.value[0])

const latestValueDisplay = computed(() =>
  latestPoint.value ? numberFormatter.format(latestPoint.value.value) : '--'
)
const firstValueDisplay = computed(() =>
  firstPoint.value ? numberFormatter.format(firstPoint.value.value) : '--'
)
const latestDateDisplay = computed(() => latestPoint.value?.dateLabel ?? '暂无数据')
const firstDateDisplay = computed(() => firstPoint.value?.dateLabel ?? '暂无数据')

const deltaValue = computed(() => {
  if (!latestPoint.value || !firstPoint.value) {
    return 0
  }
  return latestPoint.value.value - firstPoint.value.value
})

const deltaPrefix = computed(() => {
  if (deltaValue.value > 0) return '+'
  if (deltaValue.value < 0) return '-'
  return ''
})

const deltaRate = computed(() => {
  if (!firstPoint.value || firstPoint.value.value === 0) {
    return '--'
  }
  const rate = (deltaValue.value / firstPoint.value.value) * 100
  return `${rate > 0 ? '+' : ''}${rate.toFixed(2)}%`
})

const normalizeRecord = (record = {}) => {
  const rawNumber = record.Num?.toString().replace(/[^\d.-]/g, '') ?? '0'
  const dateString = record.Data?.trim()
  const parsedDate = dateString ? new Date(dateString.replace(/-/g, '/')) : null
  return {
    material: record.Cl?.trim() || '未知材料',
    value: Number(rawNumber) || 0,
    date: parsedDate?.getTime() || 0,
    dateLabel: dateString || ''
  }
}

const fetchStatistics = async () => {
  loading.value = true
  try {
    const response = await apiMethods.getBagStatistics()
    if (!Array.isArray(response)) {
      statistics.value = []
      message.warning('未获取到材料统计数据')
      return
    }
    const normalized = response
      .map(normalizeRecord)
      .filter(item => item.date)
      .sort((a, b) => a.date - b.date)
    statistics.value = normalized
    if (!selectedMaterial.value && normalized.length) {
      selectedMaterial.value = normalized[0].material
    } else if (!materialOptions.value.includes(selectedMaterial.value)) {
      selectedMaterial.value = normalized[0]?.material || ''
    }
  } catch (error) {
    message.error('获取材料统计失败')
  } finally {
    loading.value = false
  }
}

const initChart = () => {
  if (chartRef.value) {
    chartInstance = echarts.init(chartRef.value)
  }
}

const updateChart = () => {
  if (!chartInstance || !filteredSeries.value.length) {
    chartInstance?.clear()
    return
  }

  const axisLabels = filteredSeries.value.map(item => item.dateLabel)
  const valueSeries = filteredSeries.value.map(item => item.value)
  const risingSeries = new Array(valueSeries.length).fill(null)
  const fallingSeries = new Array(valueSeries.length).fill(null)
  const flatSeries = new Array(valueSeries.length).fill(null)

  for (let i = 1; i < valueSeries.length; i += 1) {
    const prev = valueSeries[i - 1]
    const curr = valueSeries[i]
    let target
    if (curr > prev) {
      target = risingSeries
    } else if (curr < prev) {
      target = fallingSeries
    } else {
      target = flatSeries
    }
    target[i - 1] = prev
    target[i] = curr
  }

  const pointColors = valueSeries.map((value, index, arr) => {
    if (index === 0) {
      const next = arr[1]
      if (next == null) return flatColor
      if (next > value) return riseColor
      if (next < value) return fallColor
      return flatColor
    }

    const prev = arr[index - 1]
    if (value > prev) return riseColor
    if (value < prev) return fallColor
    return flatColor
  })

  const buildLineSeries = (name, data, color, z = 1) => ({
    name,
    type: 'line',
    data,
    smooth: true,
    symbol: 'none',
    connectNulls: false,
    lineStyle: {
      color,
      width: 3
    },
    emphasis: {
      focus: 'series'
    },
    z
  })

  chartInstance.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'line'
      },
      formatter(params) {
        const scatterParam = params.find(item => item.seriesType === 'scatter')
        const fallback = params[0]
        const dataIndex = scatterParam?.dataIndex ?? fallback?.dataIndex ?? 0
        const point = filteredSeries.value[dataIndex]
        if (!point) return '暂无数据'
        return `${point.dateLabel}<br/>数量：${numberFormatter.format(point.value)}`
      }
    },
    grid: {
      left: 50,
      right: 20,
      bottom: 40,
      top: 40
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: axisLabels,
      axisLabel: {
        formatter(value) {
          return value.replace(' ', '\n')
        }
      }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        formatter(value) {
          return numberFormatter.format(value)
        }
      },
      splitLine: {
        lineStyle: {
          type: 'dashed'
        }
      }
    },
    series: [
      buildLineSeries('下降', fallingSeries, fallColor, 2),
      buildLineSeries('上升', risingSeries, riseColor, 3),
      buildLineSeries('持平', flatSeries, flatColor, 1),
      {
        name: '数据点',
        type: 'scatter',
        data: valueSeries,
        symbol: 'circle',
        symbolSize: 8,
        itemStyle: {
          color(params) {
            return pointColors[params.dataIndex] || flatColor
          },
          borderColor: '#fff',
          borderWidth: 2
        },
        z: 4
      }
    ]
  })
}

const refresh = () => {
  fetchStatistics()
}

const goBackStatistics = () => {
  router.push('/BagStatistics')
}

const handleResize = () => {
  chartInstance?.resize()
}

watch(filteredSeries, () => {
  nextTick(() => updateChart())
})

onMounted(async () => {
  initChart()
  await fetchStatistics()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chartInstance?.dispose()
  chartInstance = null
})
</script>

<style scoped>
.material-trend-page {
  padding: 24px;
  min-height: 100vh;
  background: #fff0f6;
}

.trend-header {
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(255, 92, 141, 0.15), rgba(255, 204, 230, 0.3));
  margin-bottom: 16px;
}

.controls-card {
  margin-bottom: 16px;
}

.controls-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.select-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.control-label {
  font-weight: 600;
  color: #ff5c8d;
}

.summary-row {
  margin-bottom: 16px;
}

.stat-card {
  text-align: center;
  background: linear-gradient(160deg, rgba(255, 92, 141, 0.1), rgba(255, 255, 255, 0.9));
  border: none;
  border-radius: 16px;
}

.stat-label {
  font-size: 14px;
  color: #ff5c8d;
  margin-bottom: 6px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #333;
}

.stat-value.up {
  color: #ff5c8d;
}

.stat-value.down {
  color: #1f8ef1;
}

.stat-hint {
  margin-top: 4px;
  font-size: 12px;
  color: #888;
}

.chart-card {
  border: none;
  border-radius: 16px;
}

.chart-wrapper {
  position: relative;
  min-height: 320px;
}

.chart-container {
  width: 100%;
  height: 480px;
}

.chart-empty {
  position: absolute;
  inset: 0;
  display: flex;
  justify-content: center;
  align-items: center;
}

@media (max-width: 768px) {
  .material-trend-page {
    padding: 16px;
  }

  .controls-row {
    flex-direction: column;
    align-items: flex-start;
  }

  .chart-container {
    height: 360px;
  }

  .stat-value {
    font-size: 22px;
  }
}
</style>
