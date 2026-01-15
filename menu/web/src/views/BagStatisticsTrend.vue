<template>
  <div class="material-trend-page">
    <!-- 背景装饰层 -->
    <div class="bg-layer" aria-hidden="true">
      <span class="orb orb-1"></span>
      <span class="orb orb-2"></span>
      <span class="orb orb-3"></span>
      <span class="star s1"></span>
      <span class="star s2"></span>
      <span class="star s3"></span>
      <span class="star s4"></span>
      <span class="star s5"></span>
      <span class="grid-glow"></span>
    </div>

    <a-page-header
      class="trend-header"
      title="材料变化图"
      sub-title="根据背包统计记录查看数量走势"
      @back="$router.back()"
    />

    <a-card class="controls-card glass" bordered>
      <div class="controls-row">
        <div class="select-wrapper">
          <span class="control-label">选择材料：</span>
          <a-select
            v-model:value="selectedMaterial"
            placeholder="请选择材料"
            style="min-width: 220px"
            :loading="loading"
            :options="materialOptions.map(material => ({ label: material, value: material }))"
            class="neon-select"
          />
        </div>

        <a-space size="middle" class="btn-group">
          <a-button type="primary" @click="refresh" :loading="loading" class="btn-neon btn-primary">
            重新获取
          </a-button>
          <a-button @click="goBackStatistics" class="btn-neon btn-ghost">
            返回材料表格
          </a-button>
        </a-space>
      </div>

      <div class="hint-line">
        <span class="badge">✦</span>
        <span class="hint-text">提示：选择材料后会自动刷新趋势图；数据点颜色表示相邻变化方向。</span>
      </div>
    </a-card>

    <a-row :gutter="16" class="summary-row">
      <a-col :xs="24" :sm="8">
        <a-card class="stat-card glass stat-neon" bordered>
          <div class="stat-label">最新数量</div>
          <div class="stat-value glow">{{ latestValueDisplay }}</div>
          <div class="stat-hint">{{ latestDateDisplay }}</div>
        </a-card>
      </a-col>

      <a-col :xs="24" :sm="8">
        <a-card class="stat-card glass stat-neon" bordered>
          <div class="stat-label">起始数量</div>
          <div class="stat-value glow">{{ firstValueDisplay }}</div>
          <div class="stat-hint">{{ firstDateDisplay }}</div>
        </a-card>
      </a-col>

      <a-col :xs="24" :sm="8">
        <a-card class="stat-card glass stat-neon" bordered>
          <div class="stat-label">区间变化</div>
          <div
            class="stat-value glow"
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

    <a-card class="chart-card glass" bordered>
      <template #title>
        <div class="chart-title">
          <span class="chart-title__icon">⟡</span>
          <span>趋势图</span>
          <span class="chart-title__sub">「记录的轨迹，会在此显形」</span>
        </div>
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

  // ✅ 仅换皮肤：不改变你的数据/逻辑
  chartInstance.setOption({
    backgroundColor: 'transparent',
    textStyle: {
      color: 'rgba(255,255,255,0.88)',
      fontFamily: 'ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial'
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'line' },
      backgroundColor: 'rgba(10,12,20,0.92)',
      borderColor: 'rgba(255, 77, 79, 0.35)',
      borderWidth: 1,
      extraCssText: 'backdrop-filter: blur(10px); border-radius: 12px; box-shadow: 0 12px 30px rgba(0,0,0,.45);',
      formatter(params) {
        const scatterParam = params.find(item => item.seriesType === 'scatter')
        const fallback = params[0]
        const dataIndex = scatterParam?.dataIndex ?? fallback?.dataIndex ?? 0
        const point = filteredSeries.value[dataIndex]
        if (!point) return '暂无数据'
        return `
          <div style="font-weight:700; margin-bottom:6px; color: rgba(255,255,255,.92);">
            ${point.dateLabel}
          </div>
          <div style="display:flex; gap:10px; align-items:center;">
            <span style="width:8px;height:8px;border-radius:999px;background:${pointColors[dataIndex] || flatColor};display:inline-block;box-shadow:0 0 12px ${pointColors[dataIndex] || flatColor};"></span>
            <span style="color: rgba(255,255,255,.86);">数量：</span>
            <span style="font-weight:800; color: rgba(255,255,255,.96);">${numberFormatter.format(point.value)}</span>
          </div>
        `
      }
    },
    legend: {
      top: 6,
      right: 10,
      textStyle: { color: 'rgba(255,255,255,0.78)' },
      itemWidth: 12,
      itemHeight: 6
    },
    grid: { left: 54, right: 22, bottom: 46, top: 56 },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: axisLabels,
      axisLabel: {
        color: 'rgba(255,255,255,0.75)',
        fontSize: 12,
        formatter(value) {
          return value.replace(' ', '\n')
        }
      },
      axisLine: { lineStyle: { color: 'rgba(255,255,255,0.18)' } },
      axisTick: { show: false },
      splitLine: { show: false }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        color: 'rgba(255,255,255,0.72)',
        formatter(value) {
          return numberFormatter.format(value)
        }
      },
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: {
        lineStyle: {
          type: 'dashed',
          color: 'rgba(255,255,255,0.14)'
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
        symbolSize: 9,
        itemStyle: {
          color(params) {
            return pointColors[params.dataIndex] || flatColor
          },
          borderColor: 'rgba(255,255,255,0.92)',
          borderWidth: 2,
          shadowBlur: 18,
          shadowColor: 'rgba(0,0,0,0.35)'
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
/* ====== 基础变量（高对比二次元中二）====== */
.material-trend-page {
  --bg0: #070814;
  --bg1: #0b0d22;
  --neon-pink: #ff2d55;
  --neon-red: #ff4d4f;
  --neon-cyan: #22d3ee;
  --neon-violet: #a78bfa;
  --text: rgba(255, 255, 255, 0.9);
  --muted: rgba(255, 255, 255, 0.68);
  --card: rgba(15, 18, 35, 0.58);
  --stroke: rgba(255, 255, 255, 0.14);
  --shadow: 0 18px 60px rgba(0, 0, 0, 0.55);
  --radius: 18px;

  position: relative;
  padding: 22px;
  min-height: 100vh;
  color: var(--text);
  background:
    radial-gradient(1200px 700px at 20% 10%, rgba(255,45,85,0.22), transparent 55%),
    radial-gradient(1100px 720px at 85% 0%, rgba(34,211,238,0.18), transparent 55%),
    radial-gradient(900px 520px at 60% 95%, rgba(167,139,250,0.14), transparent 58%),
    linear-gradient(180deg, var(--bg0), var(--bg1));
  overflow: hidden;
}

/* 背景装饰层 */
.bg-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}

.grid-glow {
  position: absolute;
  inset: -20%;
  background-image:
    linear-gradient(rgba(255,255,255,0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255,255,255,0.06) 1px, transparent 1px);
  background-size: 44px 44px;
  transform: rotate(-8deg);
  opacity: 0.28;
  filter: drop-shadow(0 0 18px rgba(34,211,238,0.18));
}

.orb {
  position: absolute;
  width: 480px;
  height: 480px;
  border-radius: 999px;
  filter: blur(26px);
  opacity: 0.85;
  animation: floaty 8s ease-in-out infinite;
}
.orb-1 {
  left: -140px;
  top: -120px;
  background: radial-gradient(circle at 30% 30%, rgba(255,45,85,0.55), transparent 62%);
}
.orb-2 {
  right: -160px;
  top: -80px;
  width: 520px;
  height: 520px;
  background: radial-gradient(circle at 40% 40%, rgba(34,211,238,0.45), transparent 62%);
  animation-duration: 10s;
}
.orb-3 {
  left: 30%;
  bottom: -220px;
  width: 560px;
  height: 560px;
  background: radial-gradient(circle at 40% 40%, rgba(167,139,250,0.35), transparent 65%);
  animation-duration: 12s;
}

.star {
  position: absolute;
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: rgba(255,255,255,0.75);
  box-shadow: 0 0 14px rgba(255,255,255,0.55);
  opacity: 0.7;
  animation: twinkle 2.6s ease-in-out infinite;
}
.s1 { left: 8%; top: 18%; }
.s2 { left: 22%; top: 42%; width: 4px; height: 4px; opacity: 0.6; animation-duration: 3.2s; }
.s3 { left: 64%; top: 22%; width: 5px; height: 5px; opacity: 0.65; animation-duration: 2.9s; }
.s4 { left: 78%; top: 58%; width: 4px; height: 4px; opacity: 0.55; animation-duration: 3.4s; }
.s5 { left: 52%; top: 78%; width: 5px; height: 5px; opacity: 0.6; animation-duration: 3.1s; }

@keyframes floaty {
  0%, 100% { transform: translate3d(0, 0, 0); }
  50% { transform: translate3d(0, 14px, 0); }
}
@keyframes twinkle {
  0%, 100% { transform: scale(1); opacity: 0.55; }
  50% { transform: scale(1.25); opacity: 0.95; }
}

/* ===== Header ===== */
.trend-header {
  border-radius: var(--radius);
  margin-bottom: 16px;
  background: linear-gradient(135deg, rgba(255, 45, 85, 0.22), rgba(34, 211, 238, 0.16));
  border: 1px solid rgba(255,255,255,0.12);
  box-shadow: var(--shadow);
  backdrop-filter: blur(10px);
}

/* Ant PageHeader 内文字增强（scoped 下用 :deep） */
.trend-header :deep(.ant-page-header-heading-title) {
  color: rgba(255,255,255,0.94);
  text-shadow: 0 0 16px rgba(255,45,85,0.25);
}
.trend-header :deep(.ant-page-header-heading-sub-title) {
  color: rgba(255,255,255,0.72);
}

/* ===== Glass Card ===== */
.glass {
  background: var(--card) !important;
  border: 1px solid rgba(255,255,255,0.14) !important;
  border-radius: var(--radius) !important;
  box-shadow: var(--shadow);
  backdrop-filter: blur(12px);
}

.controls-card {
  margin-bottom: 16px;
  position: relative;
  overflow: hidden;
}
.controls-card::before {
  content: "";
  position: absolute;
  inset: -2px;
  background: linear-gradient(90deg, rgba(255,45,85,0.22), rgba(34,211,238,0.18), rgba(167,139,250,0.18));
  filter: blur(18px);
  opacity: 0.55;
  pointer-events: none;
}

.controls-card :deep(.ant-card-body) {
  position: relative;
  z-index: 1;
}

/* 控制区布局 */
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
  gap: 10px;
  flex-wrap: wrap;
}

.control-label {
  font-weight: 800;
  letter-spacing: 0.5px;
  color: rgba(255,255,255,0.9);
  text-shadow: 0 0 12px rgba(34,211,238,0.22);
}

.hint-line {
  margin-top: 12px;
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 10px 12px;
  border-radius: 14px;
  border: 1px dashed rgba(255,255,255,0.14);
  background: rgba(0,0,0,0.18);
}
.badge {
  width: 26px;
  height: 26px;
  display: grid;
  place-items: center;
  border-radius: 999px;
  background: rgba(255,45,85,0.22);
  border: 1px solid rgba(255,45,85,0.35);
  box-shadow: 0 0 18px rgba(255,45,85,0.22);
}
.hint-text {
  color: rgba(255,255,255,0.72);
  font-size: 13px;
}

/* ===== 按钮统一霓虹 ===== */
.btn-group {
  flex-wrap: wrap;
}

.btn-neon {
  height: 36px;
  border-radius: 12px !important;
  font-weight: 800 !important;
  letter-spacing: 0.2px;
  transition: transform 0.18s ease, box-shadow 0.18s ease, filter 0.18s ease;
}

.btn-neon:hover {
  transform: translateY(-1px);
  filter: brightness(1.05);
}

.btn-primary {
  border: 1px solid rgba(255,45,85,0.45) !important;
  background: linear-gradient(135deg, rgba(255,45,85,0.95), rgba(255,77,79,0.85)) !important;
  box-shadow: 0 10px 26px rgba(255,45,85,0.25), 0 0 26px rgba(255,45,85,0.18);
}

.btn-ghost {
  color: rgba(255,255,255,0.9) !important;
  border: 1px solid rgba(34,211,238,0.35) !important;
  background: rgba(0,0,0,0.20) !important;
  box-shadow: 0 10px 22px rgba(0,0,0,0.35), 0 0 18px rgba(34,211,238,0.14);
}

/* ===== Select 霓虹皮肤（Ant 组件深度选择） ===== */
.neon-select :deep(.ant-select-selector) {
  background: rgba(0,0,0,0.25) !important;
  border: 1px solid rgba(255,255,255,0.18) !important;
  border-radius: 12px !important;
  color: rgba(255,255,255,0.92) !important;
  box-shadow: 0 0 0 1px rgba(34,211,238,0.08) inset;
}
.neon-select :deep(.ant-select-selection-placeholder) {
  color: rgba(255,255,255,0.55) !important;
}
.neon-select :deep(.ant-select-arrow) {
  color: rgba(255,255,255,0.72) !important;
}
.neon-select :deep(.ant-select-focused .ant-select-selector) {
  border-color: rgba(34,211,238,0.55) !important;
  box-shadow: 0 0 0 4px rgba(34,211,238,0.14) !important;
}

/* ===== Summary 卡片 ===== */
.summary-row {
  margin-bottom: 16px;
}
.stat-card {
  text-align: center;
  position: relative;
  overflow: hidden;
  transform: translateZ(0);
}
.stat-neon::after {
  content: "";
  position: absolute;
  inset: -2px;
  background: linear-gradient(120deg, rgba(255,45,85,0.22), rgba(34,211,238,0.16), rgba(167,139,250,0.16));
  opacity: 0.55;
  filter: blur(18px);
  pointer-events: none;
}
.stat-card :deep(.ant-card-body) {
  position: relative;
  z-index: 1;
}

.stat-label {
  font-size: 13px;
  font-weight: 800;
  color: rgba(255,255,255,0.78);
  margin-bottom: 8px;
  letter-spacing: 0.4px;
}

.stat-value {
  font-size: 30px;
  font-weight: 900;
  color: rgba(255,255,255,0.94);
}
.stat-value.glow {
  text-shadow: 0 0 18px rgba(255,45,85,0.18);
}

.stat-value.up {
  color: #ff6b8a;
  text-shadow: 0 0 18px rgba(255,45,85,0.36);
}
.stat-value.down {
  color: #6ee7b7;
  text-shadow: 0 0 18px rgba(82,196,26,0.32);
}

.stat-hint {
  margin-top: 6px;
  font-size: 12px;
  color: rgba(255,255,255,0.62);
}

/* ===== Chart Card ===== */
.chart-card {
  overflow: hidden;
}
.chart-card :deep(.ant-card-head) {
  border-bottom: 1px solid rgba(255,255,255,0.12);
}
.chart-card :deep(.ant-card-head-title) {
  color: rgba(255,255,255,0.92);
  font-weight: 900;
}
.chart-title {
  display: flex;
  align-items: baseline;
  gap: 10px;
}
.chart-title__icon {
  display: inline-grid;
  place-items: center;
  width: 22px;
  height: 22px;
  border-radius: 999px;
  background: rgba(34,211,238,0.16);
  border: 1px solid rgba(34,211,238,0.28);
  box-shadow: 0 0 18px rgba(34,211,238,0.16);
}
.chart-title__sub {
  color: rgba(255,255,255,0.58);
  font-weight: 700;
  font-size: 12px;
}

.chart-wrapper {
  position: relative;
  min-height: 320px;
}
.chart-container {
  width: 100%;
  height: 520px;
}
.chart-empty {
  position: absolute;
  inset: 0;
  display: flex;
  justify-content: center;
  align-items: center;
}
.chart-empty :deep(.ant-empty-description) {
  color: rgba(255,255,255,0.62);
}

/* ===== 移动端适配 ===== */
@media (max-width: 768px) {
  .material-trend-page {
    padding: 14px;
  }

  .controls-row {
    flex-direction: column;
    align-items: stretch;
  }

  .select-wrapper {
    width: 100%;
    justify-content: space-between;
  }

  /* 让 select 在手机端更好点 */
  .neon-select {
    width: 100%;
  }
  .neon-select :deep(.ant-select-selector) {
    min-height: 40px;
    align-items: center;
  }

  .btn-group :deep(.ant-space-item) {
    width: 100%;
  }
  .btn-neon {
    width: 100%;
    height: 40px;
  }

  .chart-container {
    height: 380px;
  }

  .stat-value {
    font-size: 24px;
  }
}
</style>
