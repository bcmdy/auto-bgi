<template>
  <div class="app-container" :class="{ 'theme-hacker': isHackerTheme }">
    <canvas ref="animeStars" id="bg-canvas"></canvas>

    <header class="navbar">
      <div class="nav-center">
        <h1 class="page-title">BGI实时日志</h1>
        <div class="status-dot" :class="{ active: ws && ws.readyState === 1 }"></div>
      </div>

      <div class="nav-right">
        <div class="select-wrapper">
          <select v-model="selectedLog" @change="onLogChange" class="glass-select">
            <option value="" disabled>选择日志源...</option>
            <option v-for="file in logFiles" :key="file" :value="file">{{ file }}</option>
          </select>
        </div>
        <button class="glass-btn theme-btn" @click="toggleTheme">
          {{ isHackerTheme ? '魔法风' : '黑客风' }}
        </button>
      </div>
    </header>

    <div class="content-body">
      <section class="panel-section log-section">
        <div class="panel-header">
          <span class="panel-label">实时日志</span>
          <div class="window-controls">
            <span class="dot red"></span>
            <span class="dot yellow"></span>
            <span class="dot green"></span>
          </div>
        </div>
        <div class="log-viewport" ref="logContainer">
          <pre class="log-text">{{ logContent }}</pre>
        </div>
      </section>

      <aside class="panel-section swiper-section">
        <div class="swiper-container right-bg-swiper">
          <div class="swiper-wrapper" ref="swiperWrapper">
          </div>
          <div class="decorative-frame"></div>
        </div>
      </aside>
    </div>

    <button class="glass-btn floating-home-btn" @click="goHome" title="返回首页">
      <span class="icon">⌂</span>
    </button>

  </div>
</template>


<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { Swiper } from 'swiper/bundle'
import 'swiper/css/bundle'

const router = useRouter()

// --- 响应式数据 ---
const animeStars = ref(null)
const logContainer = ref(null)
const swiperWrapper = ref(null)
const selectedLog = ref('')
const logFiles = ref([])
const logContent = ref('>> System Initializing...\n>> Waiting for log selection...')
const isHackerTheme = ref(false)

// --- 常量配置 ---
const CONSTANTS = {
  STAR_COUNT: 80,
  SWIPER_CONFIG: {
    delay: 8000,
    speed: 1200
  },
  STATIC_IMAGES: ['bd.jpg', 'ff.png', 'ng.jpg', 'sh.jpg'], // 确保这些图片在 public/img/ 下
  IMG_CACHE_KEY: 'sys_bg_cache_v1',
  IMG_CACHE_TTL: 604800000 // 7天
}

// --- WebSocket 变量 ---
let ws = null
let canvas = null
let ctx = null
let width = 0
let height = 0
const stars = []
let mySwiper = null
let rafId = 0
// 自动滚动控制
let autoScroll = true
let scrollRaf = 0

// --- 核心逻辑 ---

const goHome = () => router.push('/')

const toggleTheme = () => {
  isHackerTheme.value = !isHackerTheme.value
  setupStars()
  // 强制滚动到底部
  nextTick(scrollToBottom)
}

const onLogChange = () => {
  if (selectedLog.value) connectWebSocket(selectedLog.value)
}

const scrollToBottom = () => {
  const el = logContainer.value
  if (!el) return
  // 使用 requestAnimationFrame 批处理多次消息导致的重复滚动
  if (scrollRaf) cancelAnimationFrame(scrollRaf)
  scrollRaf = requestAnimationFrame(() => {
    el.scrollTop = el.scrollHeight
    scrollRaf = 0
  })
}

// 监听用户手动滚动，更新 autoScroll 标志
const onLogScroll = () => {
  const el = logContainer.value
  if (!el) return
  // 如果用户接近底部则认为允许自动滚动（阈值20px）
  autoScroll = (el.scrollTop + el.clientHeight >= el.scrollHeight - 20)
}

// --- 动画效果 (保留原有逻辑，微调参数) ---
const setupStars = () => {
  if (!animeStars.value) return
  canvas = animeStars.value
  ctx = canvas.getContext('2d')
  width = window.innerWidth
  height = window.innerHeight
  canvas.width = width
  canvas.height = height

  stars.length = 0
  const isHacker = isHackerTheme.value

  for (let i = 0; i < CONSTANTS.STAR_COUNT; i++) {
    stars.push({
      x: Math.random() * width,
      y: Math.random() * height,
      size: Math.random() * (isHacker ? 3 : 2) + 1,
      speed: Math.random() * 0.05 + 0.02,
      brightness: Math.random(),
      maxBrightness: Math.random() * 0.5 + 0.5,
      increasing: Math.random() > 0.5,
      // 颜色配置：粉色魔法 vs 绿色代码
      color: isHacker 
        ? (Math.random() > 0.8 ? '#0f0' : '#00ff41') 
        : (Math.random() > 0.5 ? '#ff7eb3' : '#7afcff')
    })
  }
}

const drawStars = () => {
  if (!ctx) return
  ctx.clearRect(0, 0, width, height)
  const isHacker = isHackerTheme.value

  stars.forEach(star => {
    // 闪烁逻辑
    star.brightness += star.increasing ? star.speed : -star.speed
    if (star.brightness >= star.maxBrightness) star.increasing = false
    if (star.brightness <= 0.1) star.increasing = true

    // 移动逻辑 (Hacker模式下坠，Magic模式漂浮)
    if (isHacker) {
      star.y += star.speed * 20
      if (star.y > height) star.y = 0
    } else {
      star.y -= star.speed * 5
      if (star.y < 0) star.y = height
    }

    ctx.globalAlpha = star.brightness
    ctx.fillStyle = star.color
    
    if (isHacker) {
      // 绘制字符
      ctx.font = '12px monospace'
      ctx.fillText(Math.random() > 0.5 ? '1' : '0', star.x, star.y)
    } else {
      // 绘制圆点
      ctx.beginPath()
      ctx.arc(star.x, star.y, star.size, 0, Math.PI * 2)
      ctx.fill()
    }
  })
  ctx.globalAlpha = 1
  rafId = requestAnimationFrame(drawStars)
}

// --- WebSocket 逻辑 ---
const connectWebSocket = name => {
  if (ws) {
    ws.close()
    ws = null
  }
  logContent.value = `>> 连接日志: ${name}...\n`

  try {
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
    const wsUrl = `${protocol}://${location.host}/ws/${encodeURIComponent(name)}`
    ws = new WebSocket(wsUrl)

    ws.onmessage = e => {
      if (!logContainer.value) return
      // 根据用户最近一次手动滚动决定是否自动滚动
      logContent.value += e.data
      if (autoScroll) nextTick(scrollToBottom)
    }

    ws.onopen = () => {
      logContent.value += `>> [SUCCESS] 连接成功: ${name}\n----------------------------------------\n`
    }
    ws.onclose = e => {
      logContent.value += `\n>> [CLOSED] 连接已终止 (Code: ${e.code})`
    }
    ws.onerror = () => {
      logContent.value += `\n>> [ERROR] 连接失败.`
    }
  } catch (err) {
    logContent.value += `\n>> [FATAL] 连接失败: ${err.message}`
  }
}

// --- 数据加载 (保持原有) ---
const loadLogFiles = async () => {
  try {
    const token = localStorage.getItem('aBgiToken')
    const headers = token ? { 'Authorization': token } : {}
    const res = await fetch('/api/logFiles', { headers })
    
    if (!res.ok) throw new Error(`Status ${res.status}`)
    const data = await res.json()
    
    if (data.files?.length) {
      logFiles.value = data.files
      selectedLog.value = data.files[0]
      connectWebSocket(data.files[0])
    } else {
      logContent.value = '>> No log files available.'
    }
  } catch (err) {
    logContent.value = `>> Failed to load file list: ${err.message}`
  }
}

// --- 轮播图逻辑 (优化适配) ---
const getImages = async () => {
  // 移动端直接跳过
  if (window.innerWidth <= 768) return
  if (!swiperWrapper.value) return

  // 简单的缓存/加载逻辑
  const cached = localStorage.getItem(CONSTANTS.IMG_CACHE_KEY)
  const list = cached ? JSON.parse(cached).list : CONSTANTS.STATIC_IMAGES
  
  // 渲染DOM
  swiperWrapper.value.innerHTML = ''
  list.forEach(src => {
    const div = document.createElement('div')
    div.className = 'swiper-slide'
    const img = document.createElement('img')
    img.src = `/img/${src}`
    // 使用 Object-fit: cover 保证填满1/3区域
    img.style.width = '100%'
    img.style.height = '100%'
    img.style.objectFit = 'cover'
    img.style.borderRadius = '12px'
    div.appendChild(img)
    swiperWrapper.value.appendChild(div)
  })

  await nextTick()

  // 初始化 Swiper
  if (mySwiper) mySwiper.destroy()
  mySwiper = new Swiper('.right-bg-swiper', {
    effect: 'fade',
    fadeEffect: { crossFade: true },
    loop: true,
    speed: CONSTANTS.SWIPER_CONFIG.speed,
    autoplay: {
      delay: CONSTANTS.SWIPER_CONFIG.delay,
      disableOnInteraction: false
    },
    allowTouchMove: false // 禁止手动拖拽
  })
}

// --- 生命周期 ---
const handleResize = () => {
  setupStars()
  // 简单的去抖动处理
  if (window.innerWidth <= 768 && mySwiper) {
    mySwiper.destroy()
    mySwiper = null
  } else if (window.innerWidth > 768 && !mySwiper) {
    getImages()
  }
}

onMounted(() => {
  setupStars()
  drawStars()
  loadLogFiles()
  getImages()
  window.addEventListener('resize', handleResize)
  // 在挂载后注册日志容器的滚动监听器，用于检测用户是否手动滚动
  nextTick(() => {
    if (logContainer.value) logContainer.value.addEventListener('scroll', onLogScroll)
  })
})

onUnmounted(() => {
  if (ws) ws.close()
  if (mySwiper) mySwiper.destroy()
  if (rafId) cancelAnimationFrame(rafId)
  window.removeEventListener('resize', handleResize)
  if (logContainer.value) logContainer.value.removeEventListener('scroll', onLogScroll)
})
</script>

<style scoped>
/* 核心修复：去除浏览器默认边距，防止灰边和滚动条 */
html, body {
  margin: 0;
  padding: 0;
  width: 100%;
  height: 100%;
  overflow: hidden; /* 禁止整个页面的滚动条 */
  background-color: #1a1b2e; /* 补上底色，防止露馅 */
}

/* 确保 Vue 根节点也占满 */
#app {
  width: 100%;
  height: 100%;
}
</style>

<style scoped>
@import '../assets/css2.css';

/* --- 全局变量 --- */
.app-container {
  /* Magic Theme (默认) */
  --bg-core: #1a1b2e;
  --bg-gradient: linear-gradient(135deg, #1a1b2e 0%, #2d1b4e 100%);
  --glass-bg: rgba(255, 255, 255, 0.05);
  --glass-border: rgba(255, 255, 255, 0.1);
  --primary: #ff7eb3;
  --text-main: #ffffff;
  --text-dim: rgba(255, 255, 255, 0.6);
  --font-ui: 'Nunito', sans-serif;
  --font-code: 'JetBrains Mono', monospace;
  --shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.3);
  
  /* 布局修复 */
  position: relative;
  width: 100%;  /* 改用 100% 避免 vw 计算误差 */
  height: 100%; /* 改用 100% */
  background: var(--bg-gradient);
  color: var(--text-main);
  font-family: var(--font-ui);
  display: flex;
  flex-direction: column;
}

/* Hacker Theme */
.theme-hacker {
  --bg-core: #000000;
  --bg-gradient: #000000;
  --glass-bg: rgba(0, 50, 0, 0.3);
  --glass-border: #00ff41;
  --primary: #00ff41;
  --text-main: #00ff41;
  --text-dim: #008F11;
  --font-ui: 'Courier New', monospace;
  --font-code: 'Courier New', monospace;
  --shadow: 0 0 15px rgba(0, 255, 65, 0.2);
}

#bg-canvas {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 0;
  pointer-events: none;
}

/* --- 顶部导航栏 --- */
.navbar {
  position: relative;
  z-index: 10;
  height: 64px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  background: rgba(20, 20, 30, 0.4);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--glass-border);
  flex-shrink: 0;
}

.nav-right{
    display: flex; /* 使用flexbox让内容在一行排列 */
  align-items: center; /* 垂直居中对齐 */
  gap: 10px; /* 可以调整两个元素之间的间距 */
}

.nav-center {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-title {
  font-size: 1.2rem;
  letter-spacing: 2px;
  font-weight: 700;
  text-shadow: 0 0 10px var(--primary);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #555;
  box-shadow: 0 0 0 2px rgba(255,255,255,0.1);
  transition: all 0.3s;
}

.status-dot.active {
  background: #00ff41;
  box-shadow: 0 0 8px #00ff41;
}

.glass-btn, .glass-select {
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  color: var(--text-main);
  padding: 8px 16px;
  border-radius: 8px;
  backdrop-filter: blur(4px);
  cursor: pointer;
  transition: all 0.3s ease;
  font-family: inherit;
  font-size: 0.7rem;
  outline: none;
}

.glass-btn:hover {
  background: rgba(255, 255, 255, 0.15);
  border-color: var(--primary);
  box-shadow: 0 0 12px var(--primary);
}

.glass-select {
  min-width: 200px;
}

.glass-select option {
  background: #2a1f3d;
  color: white;
}

/* --- 主体内容 --- */
.content-body {
  position: relative;
  z-index: 5;
  flex: 1;
  display: flex;
  padding: 20px;
  gap: 20px;
  overflow: hidden;
  box-sizing: border-box; /* 关键：防止padding撑大容器 */
}

/* 左侧日志区 */
.panel-section {
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: 16px;
  backdrop-filter: blur(12px);
  box-shadow: var(--shadow);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: all 0.3s ease;
}

.log-section {
  flex: 2; 
  position: relative;
  min-width: 0; /* 防止Flex子项溢出 */
}

/* 右侧轮播区 */
.swiper-section {
  flex: 1;
  padding: 0;
  background: transparent;
  border: none;
  box-shadow: none;
  backdrop-filter: none;
  min-width: 0; /* 防止Flex子项溢出 */
}

.panel-header {
  height: 40px;
  background: rgba(0, 0, 0, 0.2);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  border-bottom: 1px solid var(--glass-border);
}

.panel-label {
  font-size: 0.8rem;
  color: var(--text-dim);
  font-weight: 700;
  letter-spacing: 1px;
}

.window-controls {
  display: flex;
  gap: 6px;
}

.dot { width: 10px; height: 10px; border-radius: 50%; opacity: 0.7; }
.red { background: #ff5f56; }
.yellow { background: #ffbd2e; }
.green { background: #27c93f; }

.log-viewport {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  scroll-behavior: smooth;
}

.log-text {
  font-family: var(--font-code);
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
  color: var(--text-main);
}

/* 滚动条美化 */
.log-viewport::-webkit-scrollbar { width: 8px; }
.log-viewport::-webkit-scrollbar-track { background: rgba(0,0,0,0.1); }
.log-viewport::-webkit-scrollbar-thumb { background: rgba(255, 255, 255, 0.2); border-radius: 4px; }
.log-viewport::-webkit-scrollbar-thumb:hover { background: var(--primary); }

.swiper-container {
  width: 100%;
  height: 100%;
  border-radius: 16px;
  overflow: hidden;
  position: relative;
  box-shadow: var(--shadow);
  border: 1px solid var(--glass-border);
}

.swiper-slide img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* 装饰框 */
.decorative-frame {
  position: absolute;
  top: 10px; left: 10px; right: 10px; bottom: 10px;
  border: 1px solid rgba(255,255,255,0.15);
  border-radius: 12px;
  pointer-events: none;
  z-index: 10;
}

.theme-hacker .decorative-frame {
  border-color: var(--primary);
  box-shadow: inset 0 0 10px var(--primary);
}


/* --- 新增：右下角悬浮按钮样式 --- */
.floating-home-btn {
  position: fixed;
  bottom: 40px;
  right: 40px;
  z-index: 1000; /* 确保在最上层 */
  width: 56px;
  height: 56px;
  border-radius: 50%; /* 圆形 */
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
  font-size: 1.5rem; /* 图标放大 */
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275); /* 弹性动画 */
}

/* 悬浮时的效果增强 */
.floating-home-btn:hover {
  transform: translateY(-5px) scale(1.1);
  box-shadow: 0 0 20px var(--primary); /* 跟随主题色发光 */
  border-color: var(--primary);
}

/* 点击时的微缩反馈 */
.floating-home-btn:active {
  transform: scale(0.95);
}

/* 图标微调 */
.floating-home-btn .icon {
  margin-top: -4px; /* 视觉修正，让房子图标居中 */
}

/* --- 移动端适配 --- */
@media (max-width: 768px) {
  .content-body {
    flex-direction: column;
    padding: 10px;
    gap: 10px;
  }
.floating-home-btn {
    bottom: 20px;
    right: 20px;
    width: 48px;
    height: 48px;
  }
  .swiper-section {
    display: none !important; /* 强制隐藏 */
  }
  .page-title {
    display: none;
  }

  /* 移动端调整选择框尺寸（保留下方非空定义） */

  .log-section {
    flex: 1;
    width: 100%;
  }
  
  .navbar {
    padding: 0 12px;
  }
  
  .glass-select {
    min-width: 120px;
    font-size: 0.8rem;
  }
}
</style>