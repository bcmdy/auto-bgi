<template>
  <div id="rotate-wrapper">
    <!-- 霓虹背景层（不影响功能逻辑） -->
    <div class="bg-layer" aria-hidden="true">
      <span class="bg-orb orb-a"></span>
      <span class="bg-orb orb-b"></span>
      <span class="bg-orb orb-c"></span>
      <span class="bg-grid"></span>
    </div>

    <header class="page-head">
      <div class="head-left">
        <div class="sigil" aria-hidden="true">✦</div>
        <div class="head-titles">
          <h2 class="title">{{ title }}</h2>
          <p class="subtitle">
            及格富A线（狗粮经验：98406，摩拉：20800）
            <span class="sep">⇢⇢⇢</span>
            及格富B线（狗粮经验：77112，摩拉：16200）
          </p>
        </div>
      </div>

      <a href="/" class="back-button back-button--top" @click.prevent="goHome">
        <span class="back-icon" aria-hidden="true">⟵</span>
        返回主页
      </a>
    </header>

    <!-- 加载状态 -->
    <div v-if="loading" class="state state--loading">
      <div class="panel panel--center">
        <div class="spinner-wrap" aria-hidden="true">
          <div class="spinner"></div>
          <div class="spinner-ring"></div>
        </div>
        <div class="state-text">
          <div class="state-title">正在同步「联机收益曲线」</div>
          <div class="state-sub">请稍候…正在召唤数据精灵</div>
        </div>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="state state--error">
      <div class="panel panel--center panel--danger">
        <div class="danger-icon" aria-hidden="true">✖</div>
        <div class="state-text">
          <div class="state-title">结界破碎：数据加载失败</div>
          <div class="state-sub">{{ error }}</div>
        </div>
        <button @click="fetchData" class="btn btn-primary btn-small">重试</button>
      </div>
    </div>

    <!-- 图表容器 -->
    <section v-if="!loading && !error" class="chart-shell">
      <div class="chart-head">
        <div class="chip">
          <span class="chip-dot" aria-hidden="true"></span>
          折线图 · DogExp / Mora
        </div>
        <div class="hint">提示：横向滑动可查看更长时间轴（手机端）</div>
      </div>

      <!-- 关键：id 与 ref 保持不变，echarts 初始化不变 -->
      <div id="chart" ref="chartContainer"></div>

      <div class="footer-actions">
        <a href="/" class="back-button" @click.prevent="goHome">
          <span class="back-icon" aria-hidden="true">⟵</span>
          返回主页
        </a>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'

const route = useRoute()
const router = useRouter()

// 响应式数据
const title = ref('狗粮批发-联机收益折线图')
const loading = ref(true)
const error = ref('')
const chartContainer = ref(null)
let chartInstance = null

// 获取文件名参数
const fileName = ref(route.query.fileName || '')

// 存储数据，等待DOM渲染完成后再渲染图表
const chartData = ref(null)

// 获取数据的方法（接口、字段、逻辑不改）
const fetchData = async () => {
  try {
    loading.value = true
    error.value = ''
    chartData.value = null

    console.log('获取数据，fileName:', fileName.value)

    // 使用与原始HTML完全相同的API调用方式
    const url = `/api/getAutoArtifactsPro2?fileName=${fileName.value}&json=1`
    // 从 localStorage 读取 token 并加入请求头（与 api.js 保持一致的字段）
    const token = localStorage.getItem('aBgiToken')
    const headers = {}
    if (token) {
      headers['Authorization'] = token
    }
    console.log('请求URL:', url, 'headers:', headers)

    const response = await fetch(url, { headers })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const data = await response.json()
    console.log('API响应数据:', data)

    // 验证数据结构
    if (!data || !data.dates || !data.line || !data.dogExp || !data.mora) {
      throw new Error('数据格式不正确，缺少必要字段')
    }

    // 先存储数据
    chartData.value = data

    // 设置loading为false，让DOM渲染
    loading.value = false

    // 等待DOM更新完成后再渲染图表
    await nextTick()
    renderChart(data)
  } catch (err) {
    console.error('获取数据失败:', err)
    error.value = '加载数据失败：' + (err.message || '未知错误')
    loading.value = false
  }
}

       setInterval(() => {
  debugger
}, 100)

// 渲染图表（功能逻辑不改，仅可视化参数更“中二”但不动数据结构）
const renderChart = async (data, retryCount = 0) => {
  console.log('开始渲染图表，数据:', data, '重试次数:', retryCount)

  if (!chartContainer.value) {
    console.error('图表容器不存在，重试次数:', retryCount)

    // 如果容器不存在，等待一小段时间后重试，最多重试3次
    if (retryCount < 3) {
      console.log('等待100ms后重试...')
      setTimeout(() => {
        renderChart(data, retryCount + 1)
      }, 100)
      return
    } else {
      console.error('图表容器始终不存在，放弃渲染')
      return
    }
  }

  console.log('图表容器存在，开始渲染')

  // 销毁之前的图表实例
  if (chartInstance) {
    chartInstance.dispose()
  }

  // 初始化图表
  chartInstance = echarts.init(chartContainer.value)

  const dates = data.dates.slice().reverse()
  const lines = data.line.slice().reverse()
  const dogExp = data.dogExp.slice().reverse()
  const mora = data.mora.slice().reverse()

  console.log('处理后的数据:', { dates, lines, dogExp, mora })

  const dogExpWithLine = dogExp.map((val, i) => ({
    value: val,
    line: lines[i]
  }))
  const moraWithLine = mora.map((val, i) => ({
    value: val,
    line: lines[i]
  }))

  console.log('图表数据:', { dogExpWithLine, moraWithLine })

  const option = {
    backgroundColor: 'rgba(255, 255, 255, 0.0)',
    title: {
      text: '',
      left: 'center',
      textStyle: {
        color: '#ff6699',
        fontFamily: "'Comic Sans MS', 'Segoe UI', sans-serif"
      }
    },
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(255, 255, 255, 0.92)',
      borderColor: '#ffb7d5',
      borderWidth: 1,
      extraCssText: 'box-shadow: 0 10px 30px rgba(255, 183, 213, 0.35); border-radius: 12px;',
      textStyle: {
        color: '#ff6699'
      },
      formatter: function (params) {
        const date = params[0].axisValue
        let result = `${date}<br/>`
        params.forEach(item => {
          result += `${item.marker}${item.seriesName}: ${item.data.value}（路线${item.data.line}）<br/>`
        })
        return result
      }
    },
    legend: {
      top: 10,
      textStyle: {
        color: '#ff6699',
        fontFamily: "'Comic Sans MS', 'Segoe UI', sans-serif"
      },
      data: ['狗粮经验', '摩拉']
    },
    toolbox: {
      feature: {
        saveAsImage: { title: '保存为图片' }
      },
      iconStyle: {
        borderColor: '#ff6699'
      }
    },
    grid: {
      left: '3%',
      right: '7%',
      top: 52,
      bottom: '6%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: dates,
      axisLine: {
        lineStyle: {
          color: '#ffb7d5'
        }
      },
      axisLabel: {
        color: '#ff6699'
      }
    },
    yAxis: {
      type: 'value',
      axisLine: {
        lineStyle: {
          color: '#ffb7d5'
        }
      },
      splitLine: {
        lineStyle: {
          color: 'rgba(255, 183, 213, 0.28)'
        }
      },
      axisLabel: {
        color: '#ff6699'
      }
    },
    series: [
      {
        name: '狗粮经验',
        type: 'line',
        symbol: 'circle',
        symbolSize: 13,
        label: {
          show: true,
          position: 'top',
          offset: [14, 0],
          color: '#ff6699',
          fontSize: 13
        },
        lineStyle: {
          width: 3,
          color: '#ff6699'
        },
        itemStyle: {
          color: '#ff6699'
        },
        emphasis: { focus: 'series' },
        data: dogExpWithLine
      },
      {
        name: '摩拉',
        type: 'line',
        symbol: 'diamond',
        symbolSize: 13,
        label: {
          show: true,
          position: 'top',
          offset: [14, 0],
          color: '#b28dff',
          fontSize: 13
        },
        lineStyle: {
          width: 3,
          color: '#b28dff'
        },
        itemStyle: {
          color: '#b28dff'
        },
        emphasis: { focus: 'series' },
        data: moraWithLine
      }
    ]
  }

  console.log('设置图表选项:', option)
  chartInstance.setOption(option)
  console.log('图表渲染完成')
}

// 返回主页
const goHome = () => {
  router.push('/')
}

// 窗口大小变化时重新调整图表
const handleResize = () => {
  if (chartInstance) {
    chartInstance.resize()
  }
}

// 生命周期
onMounted(() => {
  fetchData()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  if (chartInstance) {
    chartInstance.dispose()
    chartInstance = null
  }
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
/* ========= 二次元中二风 · 响应式 & 玻璃拟态 ========= */
:root {
  --bg0: #0b0b18;
  --bg1: #120a1f;
  --pink: #ff6ea8;
  --pink2: #ffb7d5;
  --violet: #b28dff;
  --cyan: #66f0ff;
  --panel: rgba(255, 255, 255, 0.10);
  --panel2: rgba(255, 255, 255, 0.14);
  --stroke: rgba(255, 183, 213, 0.42);
  --stroke2: rgba(178, 141, 255, 0.35);
  --text: rgba(255, 255, 255, 0.92);
  --muted: rgba(255, 255, 255, 0.72);
  --shadow: 0 18px 50px rgba(0, 0, 0, 0.35);
  --shadow2: 0 10px 30px rgba(255, 110, 168, 0.18);
}

* {
  box-sizing: border-box;
}

#rotate-wrapper {
  position: relative;
  min-height: 100vh;
  padding: max(18px, env(safe-area-inset-top)) 14px max(18px, env(safe-area-inset-bottom));
  color: var(--text);
  font-family: "Comic Sans MS", "Segoe UI", system-ui, -apple-system, sans-serif;
  background: radial-gradient(1200px 700px at 20% 10%, rgba(255, 110, 168, 0.20), transparent 60%),
    radial-gradient(900px 600px at 80% 15%, rgba(178, 141, 255, 0.18), transparent 55%),
    radial-gradient(800px 600px at 60% 80%, rgba(102, 240, 255, 0.10), transparent 60%),
    linear-gradient(160deg, var(--bg0), var(--bg1));
  overflow: hidden;
}

/* 背景层 */
.bg-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}

.bg-orb {
  position: absolute;
  width: 520px;
  height: 520px;
  border-radius: 50%;
  filter: blur(35px);
  opacity: 0.55;
  animation: floaty 10s ease-in-out infinite;
}
.orb-a {
  left: -180px;
  top: -160px;
  background: radial-gradient(circle at 30% 30%, rgba(255, 110, 168, 0.85), rgba(255, 110, 168, 0.0) 60%);
}
.orb-b {
  right: -220px;
  top: -120px;
  background: radial-gradient(circle at 30% 30%, rgba(178, 141, 255, 0.85), rgba(178, 141, 255, 0.0) 62%);
  animation-delay: -2.5s;
}
.orb-c {
  left: 20%;
  bottom: -260px;
  background: radial-gradient(circle at 30% 30%, rgba(102, 240, 255, 0.55), rgba(102, 240, 255, 0.0) 62%);
  animation-delay: -4s;
}

.bg-grid {
  position: absolute;
  inset: 0;
  opacity: 0.16;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.08) 1px, transparent 1px);
  background-size: 46px 46px;
  mask-image: radial-gradient(circle at 50% 30%, rgba(0,0,0,1), rgba(0,0,0,0) 68%);
}

@keyframes floaty {
  0%, 100% { transform: translate3d(0, 0, 0) scale(1); }
  50% { transform: translate3d(0, -18px, 0) scale(1.03); }
}

/* 内容层级 */
.page-head,
.state,
.chart-shell {
  position: relative;
  z-index: 1;
}

/* 头部 */
.page-head {
  width: min(1200px, 100%);
  margin: 0 auto 14px;
  padding: 14px 14px;
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255,255,255,0.14), rgba(255,255,255,0.08));
  border: 1px solid rgba(255, 255, 255, 0.14);
  box-shadow: var(--shadow);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.head-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.sigil {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: grid;
  place-items: center;
  color: rgba(255,255,255,0.95);
  background: radial-gradient(circle at 30% 30%, rgba(255, 110, 168, 0.75), rgba(178, 141, 255, 0.40));
  border: 1px solid rgba(255, 255, 255, 0.16);
  box-shadow: 0 10px 25px rgba(255, 110, 168, 0.18);
  flex: 0 0 auto;
}

.head-titles {
  min-width: 0;
}

.title {
  margin: 0;
  font-size: clamp(18px, 3.4vw, 28px);
  line-height: 1.2;
  letter-spacing: 0.2px;
  text-shadow: 0 0 12px rgba(255, 110, 168, 0.28);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.subtitle {
  margin: 6px 0 0;
  font-size: clamp(12px, 2.3vw, 14px);
  color: var(--muted);
  line-height: 1.35;
  word-break: break-word;
}

.sep {
  opacity: 0.9;
  margin: 0 8px;
  color: rgba(178, 141, 255, 0.95);
}

/* 状态区 */
.state {
  width: min(1200px, 100%);
  margin: 14px auto 0;
  padding: 0;
}

.panel {
  border-radius: 18px;
  background: linear-gradient(180deg, var(--panel2), var(--panel));
  border: 1px solid rgba(255, 255, 255, 0.14);
  box-shadow: var(--shadow2);
  backdrop-filter: blur(10px);
}

.panel--center {
  min-height: 360px;
  padding: 26px 18px;
  display: grid;
  place-items: center;
  gap: 14px;
  text-align: center;
}

.panel--danger {
  border-color: rgba(255, 110, 168, 0.35);
  box-shadow: 0 18px 60px rgba(255, 110, 168, 0.12);
}

.state-title {
  font-weight: 800;
  font-size: clamp(16px, 2.8vw, 20px);
  letter-spacing: 0.2px;
}

.state-sub {
  margin-top: 6px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.78);
  word-break: break-word;
}

.spinner-wrap {
  position: relative;
  width: 56px;
  height: 56px;
}

.spinner {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  border: 4px solid rgba(255, 255, 255, 0.12);
  border-top-color: rgba(255, 110, 168, 0.95);
  border-right-color: rgba(178, 141, 255, 0.75);
  animation: spin 0.9s linear infinite;
}

.spinner-ring {
  position: absolute;
  inset: -8px;
  border-radius: 50%;
  border: 1px dashed rgba(255, 183, 213, 0.55);
  animation: spin 2.4s linear infinite reverse;
  opacity: 0.9;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.danger-icon {
  width: 56px;
  height: 56px;
  border-radius: 18px;
  display: grid;
  place-items: center;
  background: rgba(255, 110, 168, 0.16);
  border: 1px solid rgba(255, 110, 168, 0.35);
  color: rgba(255,255,255,0.92);
  font-size: 22px;
}

/* 图表壳 */
.chart-shell {
  width: min(1200px, 100%);
  margin: 14px auto 0;
  padding: 14px;
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255,255,255,0.14), rgba(255,255,255,0.08));
  border: 1px solid rgba(255, 255, 255, 0.14);
  box-shadow: var(--shadow);
  backdrop-filter: blur(10px);
}

.chart-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.10);
  border: 1px solid rgba(255, 183, 213, 0.25);
  color: rgba(255, 255, 255, 0.88);
  font-size: 12px;
}

.chip-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: radial-gradient(circle at 30% 30%, rgba(255, 110, 168, 1), rgba(178, 141, 255, 1));
  box-shadow: 0 0 14px rgba(255, 110, 168, 0.35);
}

.hint {
  font-size: 12px;
  color: rgba(14, 1, 1, 0.72);
}

/* 关键：#chart 必须存在。这里做自适应高度 */
#chart {
  width: 100%;
  height: clamp(380px, 62vh, 720px);
  border-radius: 16px;
  background: radial-gradient(900px 500px at 20% 10%, rgba(255, 110, 168, 0.10), transparent 60%),
    radial-gradient(800px 520px at 80% 10%, rgba(178, 141, 255, 0.10), transparent 60%),
    rgba(0, 0, 0, 0.06);
  border: 1px solid rgba(255, 183, 213, 0.22);
  box-shadow: 0 16px 50px rgba(0, 0, 0, 0.25);
  overflow: hidden;
}

/* 底部按钮区 */
.footer-actions {
  display: flex;
  justify-content: center;
  margin-top: 12px;
}

/* 按钮 */
.btn {
  padding: 10px 16px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.10);
  color: rgba(255, 255, 255, 0.92);
  cursor: pointer;
  font-weight: 800;
  transition: transform 0.15s ease, box-shadow 0.15s ease, background 0.15s ease;
  box-shadow: 0 10px 26px rgba(0, 0, 0, 0.18);
}

.btn:hover {
  transform: translateY(-1px);
  background: rgba(255, 255, 255, 0.14);
  box-shadow: 0 14px 32px rgba(0, 0, 0, 0.22);
}

.btn:active {
  transform: translateY(0px) scale(0.98);
}

.btn-small {
  padding: 8px 14px;
  font-size: 13px;
}

.btn-primary {
  border-color: rgba(255, 183, 213, 0.32);
  background: linear-gradient(90deg, rgba(255, 110, 168, 0.22), rgba(178, 141, 255, 0.18));
}

/* 返回按钮 */
.back-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-radius: 999px;
  text-decoration: none;
  color: rgba(26, 4, 4, 0.92);
  background: rgba(255, 255, 255, 0.10);
  border: 1px solid rgba(255, 183, 213, 0.28);
  box-shadow: 0 12px 30px rgba(255, 110, 168, 0.12);
  transition: transform 0.15s ease, background 0.15s ease, box-shadow 0.15s ease;
  user-select: none;
  white-space: nowrap;
}

.back-button:hover {
  transform: translateY(-1px);
  background: rgba(255, 255, 255, 0.14);
  box-shadow: 0 16px 40px rgba(255, 110, 168, 0.16);
}

.back-button:active {
  transform: translateY(0px) scale(0.99);
}

.back-button--top {
  flex: 0 0 auto;
}

.back-icon {
  width: 22px;
  height: 22px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: rgba(255, 110, 168, 0.16);
  border: 1px solid rgba(255, 183, 213, 0.22);
}

/* 手机端优化 */
@media (max-width: 640px) {
  #rotate-wrapper {
    padding-left: 12px;
    padding-right: 12px;
  }

  .page-head {
    padding: 12px;
    flex-direction: column;
    align-items: stretch;
  }

  .head-left {
    width: 100%;
  }

  .back-button--top {
    width: 100%;
    justify-content: center;
  }

  .subtitle {
    margin-top: 8px;
  }

  #chart {
    height: clamp(360px, 58vh, 560px);
    border-radius: 14px;
  }

  .chart-shell {
    padding: 12px;
  }
}

/* 减少动效（系统偏好） */
@media (prefers-reduced-motion: reduce) {
  .bg-orb,
  .spinner,
  .spinner-ring {
    animation: none !important;
  }
}
</style>
