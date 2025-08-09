<template>
  <div id="rotate-wrapper">
    <h2>{{ title }}</h2>
    <h4>及格富A线（狗粮经验：98406，摩拉：20800）=======及格富B线（狗粮经验：77112，摩拉：16200）</h4>
    
    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner">
        <div class="spinner"></div>
        <p>正在加载数据...</p>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="error-container">
      <div class="error-box">
        <span class="error-icon">❌</span>
        <h3>数据加载失败</h3>
        <p>{{ error }}</p>
        <button @click="fetchData" class="btn btn-small">重试</button>
      </div>
    </div>

    <!-- 图表容器 -->
    <div v-if="!loading && !error" id="chart" ref="chartContainer"></div>
    
    <a href="/" class="back-button" @click.prevent="goHome">返回主页</a>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'

const route = useRoute()
const router = useRouter()

// 响应式数据
const title = ref('狗粮批发收益折线图')
const loading = ref(true)
const error = ref('')
const chartContainer = ref(null)
let chartInstance = null

// 获取文件名参数
const fileName = ref(route.query.fileName || '')

// 存储数据，等待DOM渲染完成后再渲染图表
const chartData = ref(null)

// 获取数据的方法
const fetchData = async () => {
  try {
    loading.value = true
    error.value = ''
    chartData.value = null
    
    console.log('获取数据，fileName:', fileName.value)
    
    // 使用与原始HTML完全相同的API调用方式
    const url = `/api/getAutoArtifactsPro2?fileName=${fileName.value}&json=1`
    console.log('请求URL:', url)
    
    const response = await fetch(url)
    
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

// 渲染图表
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
    backgroundColor: 'rgba(255, 255, 255, 0.6)',
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
      backgroundColor: 'rgba(255, 255, 255, 0.9)',
      borderColor: '#ffb7d5',
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
      top: 30,
      textStyle: {
        color: '#ff6699'
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
      bottom: '5%',
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
          color: 'rgba(255, 183, 213, 0.3)'
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
        symbolSize: 15,
        label: {
          show: true,
          position: 'top',
          offset: [20, 0],
          color: '#ff6699',
          fontSize: 15
        },
        lineStyle: {
          color: '#ff6699'
        },
        itemStyle: {
          color: '#ff6699'
        },
        data: dogExpWithLine
      },
      {
        name: '摩拉',
        type: 'line',
        symbolSize: 15,
        label: {
          show: true,
          position: 'top',
          offset: [20, 0],
          color: '#b28dff',
          fontSize: 15
        },
        textStyle: {
          fontSize: 30
        },
        lineStyle: {
          color: '#b28dff'
        },
        itemStyle: {
          color: '#b28dff'
        },
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
:root {
  --bg-color: #fff0f7;
  --text-color: #ff6699;
  --panel-bg: rgba(255, 255, 255, 0.8);
  --border-color: #ffb7d5;
  --accent-color: #b28dff;
}

* {
  box-sizing: border-box;
}

#rotate-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
  margin: 0;
  font-family: "Comic Sans MS", "Segoe UI", sans-serif;
  background-color: var(--bg-color);
  color: var(--text-color);
  background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100"><circle cx="30" cy="30" r="3" fill="%23ffb7d5" opacity="0.5"/><circle cx="70" cy="70" r="4" fill="%23b28dff" opacity="0.3"/></svg>');
  min-height: 100vh;
}

h2 {
  text-align: center;
  margin-bottom: 20px;
  font-size: 2rem;
  text-shadow: 0 0 5px var(--border-color);
  position: relative;
  display: inline-block;
  width: 100%;
}

h2::before, h2::after {
  content: "♡";
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  font-size: 1.5rem;
  color: var(--text-color);
  opacity: 0.7;
}

h2::before {
  left: 25%;
}

h2::after {
  right: 25%;
}

h4 {
  text-align: center;
  color: #b28dff;
  margin-bottom: 20px;
}

#chart {
  width: 100%;
  max-width: 1200px;
  height: 70vh;
  margin: 0 auto 20px;
  background: var(--panel-bg);
  border: 2px solid var(--border-color);
  border-radius: 20px;
  padding: 15px;
  box-shadow: 0 0 20px rgba(255, 183, 213, 0.5);
  position: relative;
  overflow: hidden;
}

#chart::before {
  content: "";
  position: absolute;
  top: -10px;
  left: -10px;
  right: -10px;
  bottom: -10px;
  background: linear-gradient(45deg, transparent, rgba(255, 183, 213, 0.1), transparent);
  z-index: -1;
  animation: shimmer 3s linear infinite;
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

.back-button {
  display: inline-block;
  margin-top: 20px;
  padding: 10px 25px;
  background-color: #fff;
  color: var(--text-color);
  border: 2px solid var(--border-color);
  border-radius: 50px;
  text-decoration: none;
  font-weight: bold;
  box-shadow: 0 0 10px rgba(255, 183, 213, 0.5);
  transition: all 0.3s ease;
  cursor: pointer;
}

.back-button:hover {
  background-color: var(--text-color);
  color: white;
  transform: translateY(-3px);
  box-shadow: 0 5px 15px rgba(255, 183, 213, 0.7);
}

/* 加载状态样式 */
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 400px;
  margin: 20px 0;
}

.loading-spinner {
  text-align: center;
  color: var(--text-color);
}

.spinner {
  border: 4px solid rgba(255, 183, 213, 0.3);
  border-left: 4px solid var(--text-color);
  border-radius: 50%;
  width: 40px;
  height: 40px;
  animation: spin 1s linear infinite;
  margin: 0 auto 15px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

/* 错误状态样式 */
.error-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 400px;
  margin: 20px 0;
}

.error-box {
  text-align: center;
  padding: 30px;
  background: var(--panel-bg);
  border: 2px solid var(--border-color);
  border-radius: 20px;
  color: var(--text-color);
}

.error-icon {
  font-size: 3rem;
  display: block;
  margin-bottom: 15px;
}

.btn {
  padding: 8px 16px;
  border: 2px solid var(--border-color);
  border-radius: 20px;
  background: white;
  color: var(--text-color);
  cursor: pointer;
  font-weight: bold;
  transition: all 0.3s ease;
  text-decoration: none;
  display: inline-block;
}

.btn:hover {
  background: var(--text-color);
  color: white;
  transform: translateY(-2px);
}

.btn-small {
  padding: 6px 12px;
  font-size: 0.9rem;
}
</style>
