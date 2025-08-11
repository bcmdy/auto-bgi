<template>
  <div class="home-container">
    <!-- 樱花动画背景 -->
    <canvas ref="animeCanvas" class="anime-canvas"></canvas>
    
    
    <!-- 固定右上角轮播图 -->
    <div class="swiper right-top-swiper" v-if="carouselImages.length > 0">
      <div class="swiper-wrapper">
        <div v-for="(image, index) in carouselImages" :key="index" class="swiper-slide" :class="{ active: currentImageIndex === index }">
          <img :src="image" :alt="`carousel-${index}`" />
        </div>
      </div>
  
    </div>

    <!-- 主容器 -->
    <div class="container">
      <!-- 状态卡片 -->
      <div class="status-card">
        <h2>执行配置组：🧩<span>{{ statusData.group }}</span></h2>
        <pre class="ExpectedToEnd">{{ statusData.ExpectedToEnd }}</pre>
        <p><span>📜</span> 运行脚本：<span>{{ statusData.line }}</span></p>
        <p><span>🗺️</span> 地图追踪进度：<span>{{ statusData.progress }}</span></p>
        <p><span>🖥️</span> 软件运行状态：<span>{{ statusData.running }}</span></p>
        <p><span>✨</span><span>{{ statusData.jsProgress }}</span></p>
      </div>

      <!-- 数据分析按钮组 -->
      <div class="button-group">
        <h2>📊 数据分析</h2>
        <button v-for="(button, index) in dataAnalysisButtons" :key="index" @click="$router.push(button.route)">
          {{ button.text }}
        </button>
      </div>

      <!-- 自动化按钮组 -->
      <div class="button-group">
        <h2>🚀 自动化</h2>
        <button v-for="(button, index) in automationButtons" :key="index" @click="button.action">
          {{ button.text }}
        </button>
  
      </div>

      <!-- BGI相关按钮组 -->
      <div class="button-group">
        <h2>🧭 BGI相关</h2>
        <button v-for="(button, index) in bgiButtons" :key="index" @click="$router.push(button.route)">
          {{ button.text }}
        </button>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, computed, watch } from 'vue'
import { message } from 'ant-design-vue'
import { useRouter } from 'vue-router'
import api, { apiMethods } from '@/utils/api'

const router = useRouter()

// 响应式数据
const animeCanvas = ref(null)
const carouselImages = ref([])
const currentImageIndex = ref(0)
const headerCarouselImages = ref([])
const headerCurrentImageIndex = ref(0)
let headerCarouselInterval = null
const statusData = reactive({
  group: '加载中...',
  ExpectedToEnd: '加载中...',
  line: '加载中...',
  progress: '加载中...',
  running: '加载中...',
  jsProgress: '加载中...'
})

// 计算当前显示的图片
const currentImage = computed(() => {
  if (carouselImages.value.length > 0) {
    return carouselImages.value[currentImageIndex.value]
  }
  return null
})

// 按钮配置数据
const dataAnalysisButtons = ref([
  { text: '查看收获前10', route: '/logAnalysis' },
  { text: '背包统计', route: '/BagStatistics' },
  { text: '查看狗粮日志', route: '/getAutoArtifactsPro' },
  { text: '归档查询', route: '/archive' },
  { text: '配置组运行情况', route: '/other' }
])

const bgiButtons = ref([
  { text: '调度器', route: '/listGroups' },
  { text: '实时日志', route: '/log' },
  { text: 'autobgi配置文件', route: '/Config' },
  { text: 'bgi一条龙配置', route: '/onelong' }
])

let statusInterval = null
let petals = []
let animationId = null
const showScanModal = ref(false)
const scanResult = ref('')

// 樱花花瓣类
class Petal {
  constructor(canvas) {
    this.canvas = canvas
    this.x = Math.random() * canvas.width
    this.y = Math.random() * canvas.height * -1 - 100
    this.size = Math.random() * 8 + 5
    this.speed = Math.random() * 2 + 0.5
    this.angle = Math.random() * 360
    this.spin = Math.random() * 5 - 2.5
    this.color = ["#ffcce6", "#ffd1e0", "#ff9ecd", "#ffaad5"][Math.floor(Math.random() * 4)]
  }

  update() {
    this.y += this.speed
    this.x += Math.sin(this.angle * Math.PI / 180) * 0.5
    this.angle += this.spin

    if (this.y > this.canvas.height) {
      this.y = -this.size * 2
      this.x = Math.random() * this.canvas.width
    }
  }

  draw(ctx) {
    ctx.save()
    ctx.translate(this.x, this.y)
    ctx.rotate(this.angle * Math.PI / 180)

    ctx.fillStyle = this.color
    ctx.beginPath()
    ctx.moveTo(0, 0)
    ctx.bezierCurveTo(this.size / 2, -this.size / 2, this.size, -this.size / 4, this.size, 0)
    ctx.bezierCurveTo(this.size, this.size / 4, this.size / 2, this.size / 2, 0, 0)
    ctx.fill()
    ctx.restore()
  }
}

// 初始化樱花动画
const initSakuraAnimation = () => {
  const canvas = animeCanvas.value
  if (!canvas) return

  const ctx = canvas.getContext('2d')
  canvas.width = window.innerWidth
  canvas.height = window.innerHeight

  petals = Array.from({ length: 50 }, () => new Petal(canvas))

  const animate = () => {
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    
    petals.forEach(petal => {
      petal.update()
      petal.draw(ctx)
    })

    animationId = requestAnimationFrame(animate)
  }

  animate()

  // 窗口尺寸变化处理
  const handleResize = () => {
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight
  }
  window.addEventListener('resize', handleResize)

  return () => {
    window.removeEventListener('resize', handleResize)
    if (animationId) {
      cancelAnimationFrame(animationId)
    }
  }
}

// 获取状态信息
const refreshStatus = async () => {
  try {
    const response = await apiMethods.getStatus()
    Object.assign(statusData, response)
  } catch (error) {
    console.error('获取状态失败:', error)
  }
}

// 获取轮播图图片
const getImages = async () => {
  try {
    const response = await fetch('/api/images')
    if (!response.ok) {
      throw new Error('Failed to fetch images')
    }
    const data = await response.json()
    console.log('轮播图数据:', data) // 调试信息
    carouselImages.value = data.images || []
    
    // 启动轮播
    if (carouselImages.value.length > 0) {
      console.log('轮播图数量:', carouselImages.value.length) // 调试信息
      startCarousel()
    }
  } catch (error) {
    console.error('获取轮播图失败:', error)
    // 如果API失败，使用默认图片
    carouselImages.value = ['/img/bd.jpg', '/img/ff.png', '/img/ng.jpg', '/img/sh.jpg']
    startCarousel()
  }
}

// 启动轮播
const startCarousel = () => {
  console.log('启动轮播，图片数量:', carouselImages.value.length) // 调试信息
  if (carouselImages.value.length > 1) {
    setInterval(() => {
      currentImageIndex.value = (currentImageIndex.value + 1) % carouselImages.value.length
      console.log('切换到图片:', currentImageIndex.value) // 调试信息
    }, 10000) // 每10秒切换一张图片
  }
}

// 获取header轮播图图片
const getHeaderImages = async () => {
  try {
    const response = await fetch('/api/images')
    if (!response.ok) {
      throw new Error('Failed to fetch header images')
    }
    const data = await response.json()
    console.log('Header轮播图数据:', data)
    headerCarouselImages.value = data.images || []
    
    // 启动header轮播
    if (headerCarouselImages.value.length > 0) {
      console.log('Header轮播图数量:', headerCarouselImages.value.length)
      startHeaderCarousel()
    }
  } catch (error) {
    console.error('获取Header轮播图失败:', error)
    // 如果API失败，使用默认图片
    headerCarouselImages.value = ['/img/bd.jpg', '/img/ff.png', '/img/ng.jpg', '/img/sh.jpg']
    startHeaderCarousel()
  }
}

// 启动header轮播
const startHeaderCarousel = () => {
  console.log('启动Header轮播，图片数量:', headerCarouselImages.value.length)
  if (headerCarouselImages.value.length > 1) {
    headerCarouselInterval = setInterval(() => {
      headerCurrentImageIndex.value = (headerCurrentImageIndex.value + 1) % headerCarouselImages.value.length
      console.log('切换到Header图片:', headerCurrentImageIndex.value)
    }, 6000) // 每6秒切换一张图片
  }
}

// 按钮事件处理
const handleOneLong = async () => {
  try {
    await apiMethods.startOneLong()
    message.success('启动成功！')
  } catch (error) {
    message.error('启动失败！')
  }
}

const handleCloseBgi = async () => {
  try {
    await apiMethods.closeBgi()
    message.success('关闭成功！')
  } catch (error) {
    message.error('关闭失败！')
  }
}

const handleBackup = async () => {
  try {
    await apiMethods.backup()
    message.success('备份成功！')
  } catch (error) {
    message.error('备份失败！')
  }
}

// 自动化按钮配置
const automationButtons = ref([
  { text: '一条龙启动', action: handleOneLong },
  { text: '关闭BGI', action: handleCloseBgi },
  { text: '备份 user 文件', action: handleBackup },
  { text: '脚本更新列表', action: () => router.push('/jsNames') }
])


const closeScanModal = () => {
  showScanModal.value = false
  if (window._html5QrCodeInstance) {
    window._html5QrCodeInstance.stop().then(() => {
      window._html5QrCodeInstance.clear()
      window._html5QrCodeInstance = null
    })
  }
}

// 生命周期
onMounted(() => {
  // 初始化樱花动画
  const cleanupAnimation = initSakuraAnimation()
  
  // 获取轮播图图片
  getImages()
  getHeaderImages() // 获取header轮播图图片
  
  // 获取状态并设置定时器
  refreshStatus()
  statusInterval = setInterval(refreshStatus, 10000)

  // 清理函数
  onUnmounted(() => {
    cleanupAnimation?.()
    if (statusInterval) {
      clearInterval(statusInterval)
    }
    if (headerCarouselInterval) {
      clearInterval(headerCarouselInterval)
    }
  })
})

</script>

<style scoped>
/* ============ 基础样式 ============ */
.home-container {
  position: relative;
  min-height: 100vh;
  font-family: 'Comic Sans MS', 'Segoe UI', sans-serif;
  background-color: #ffecf5;
  color: #ff6699;
  overflow-x: hidden;
  overflow-y: auto;
  background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100"><circle cx="50" cy="50" r="5" fill="%23ffcce6" opacity="0.5"/></svg>');
}

.anime-canvas {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  z-index: 0;
  pointer-events: none;
}

/* ============ Header轮播图样式 ============ */
.page-header {
  position: relative;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95) 0%, rgba(255, 246, 251, 0.9) 100%);
  padding: 40px 0 30px;
  text-align: center;
  box-shadow: 0 8px 32px rgba(255, 110, 180, 0.15);
  border-radius: 0 0 40px 40px;
  margin-bottom: 20px;
  overflow: hidden;
}

.header-carousel {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  z-index: -1;
  border-radius: 0 0 40px 40px;
}

.carousel-container {
  position: relative;
  width: 100%;
  height: 100%;
}

.carousel-slide {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  transition: opacity 1.5s ease-in-out;
}

.carousel-slide.active {
  opacity: 1;
}

.carousel-slide img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 0 0 40px 40px;
}

/* 添加渐变遮罩，确保文字可读性 */
.page-header::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.8) 0%,
    rgba(255, 246, 251, 0.7) 50%,
    rgba(255, 214, 235, 0.6) 100%
  );
  z-index: 0;
  border-radius: 0 0 40px 40px;
}

.header-content {
  position: relative;
  z-index: 1;
}

.header-title {
  color: #ff6eb4;
  font-size: 2.5rem;
  text-shadow: 0 0 20px rgba(255, 110, 180, 0.4);
  margin: 20px 0 10px;
  animation: titleGlow 3s infinite ease-in-out;
}

@keyframes titleGlow {
  0%, 100% {
    text-shadow: 0 0 20px rgba(255, 110, 180, 0.4);
  }
  50% {
    text-shadow: 0 0 30px rgba(255, 110, 180, 0.4), 0 0 40px #ff6eb4;
  }
}

.header-subtitle {
  font-size: 1.1rem;
  color: #e91e63;
  margin-top: 10px;
  opacity: 0.8;
}

/* ============ 轮播图区域 ============ */
.right-top-swiper {
  position: fixed;
  bottom: 10px;
  left: 2%;
  width: 420px;
  height: 750px;
  z-index: 1; /* 改为正数，确保不被遮挡 */
  overflow: hidden;
  pointer-events: none;
}

.right-top-swiper img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 10px;
}

.swiper-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
}

.swiper-slide {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  transition: opacity 1.5s ease-in-out;
}

.swiper-slide.active {
  opacity: 1;
}

/* ============ 主容器 ============ */
.container {
  position: relative;
  z-index: 1;
  width: 90%;
  max-width: 700px;
  margin: 400px auto 60px;
  padding: 20px;
  border-radius: 20px;
  background-color: rgba(255, 255, 255, 0);
}

/* ============ 状态卡片 ============ */
.status-card {
  position: fixed;
  top: 10px;
  left: 50%;
  transform: translateX(-50%);
  width: 90%;
  max-width: 700px;
  z-index: 999;
  background-color: rgba(255, 255, 255, 0.65);
  padding: 20px;
  border: 2px dashed #ffaad5;
  border-radius: 18px;
  box-shadow: 0 4px 20px rgba(255, 174, 209, 0.5);
  font-size: 15.5px;
  line-height: 1.8;
  backdrop-filter: blur(6px);
}

.status-card h2 {
  font-size: 1.1em;
  margin-bottom: 10px;
  color: #ff66a3;
  text-shadow: 0 0 4px #fff0f5;
  display: flex;
  align-items: center;
}

.status-card h2 span {
  margin-right: 8px;
}

.status-card p {
  margin: 6px 0;
  color: #ff4081;
  display: flex;
  align-items: center;
  font-weight: bold;
}

.status-card p span:first-child {
  margin-right: 8px;
  font-size: 1.2em;
}

.status-card span {
  font-weight: bold;
  color: #d9006c;
}

.ExpectedToEnd {
  font-size: 1.2em;
  color: #ff6699;
  text-shadow: 0 0 5px #ffd1e0;
  word-wrap: break-word;
  word-break: break-all;
  white-space: pre-wrap;
  overflow-wrap: break-word;
  max-width: 100%;
  overflow-x: auto;
}

/* ============ 按钮组 ============ */
.button-group {
  margin: 24px 0;
  padding: 16px;
  border: 2px dashed #ffaad5;
  border-radius: 20px;
  background-color: rgba(255, 255, 255, 0.15);
  box-shadow: inset 0 0 10px #ffe6f0;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.button-group h2 {
  grid-column: span 2;
  text-align: center;
  color: #ff6699;
  text-shadow: 0 0 4px #fff0f5;
  margin-bottom: 10px;
  font-size: 1.1em;
}

button {
  width: 100%;
  padding: 14px;
  font-size: 16px;
  font-weight: bold;
  background: linear-gradient(145deg, #ffd1e0, #ffe6f0);
  color: #ff4081;
  border: none;
  border-radius: 30px;
  cursor: pointer;
  box-shadow: 4px 4px 10px #ffb6c1, -4px -4px 10px #fff0f5;
  transition: all 0.3s ease;
  letter-spacing: 1px;
  position: relative;
  overflow: hidden;
}

button::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 200%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.3), transparent);
  animation: shine 2.5s infinite;
}

@keyframes shine {
  0% { left: -100%; }
  50% { left: 100%; }
  100% { left: 100%; }
}

button:hover {
  transform: scale(1.06);
  background: linear-gradient(145deg, #ffb6c1, #ffd1e0);
  color: white;
  box-shadow: 0 0 15px #ff99cc, 0 0 30px #ff99cc;
}

/* ============ 响应式设计 ============ */
@media (max-width: 480px) {
  .button-group {
    grid-template-columns: 1fr;
  }

  button {
    font-size: 14px;
    padding: 12px;
  }
  
  .right-top-swiper {
    left: 70px;
  }
  
  .container {
    margin: 450px auto 60px;
  }
  
  .status-card {
    padding: 15px;
    font-size: 14px;
    max-width: 95%;
  }
  
  .ExpectedToEnd {
    font-size: 1em;
    line-height: 1.4;
    max-height: 120px;
    overflow-y: auto;
    padding: 8px;
    background-color: rgba(255, 255, 255, 0.3);
    border-radius: 8px;
    border: 1px solid rgba(255, 174, 209, 0.3);
  }
  
  /* Header轮播图移动端适配 */
  .page-header {
    border-radius: 0 0 20px 20px;
    padding: 30px 0 20px;
  }
  
  .header-carousel {
    border-radius: 0 0 20px 20px;
  }
  
  .carousel-slide img {
    border-radius: 0 0 20px 20px;
  }
  
  .page-header::before {
    border-radius: 0 0 20px 20px;
  }
  
  .header-title {
    font-size: 2rem;
  }
  
  .header-subtitle {
    font-size: 1rem;
  }
}

@media (max-width: 360px) {
  .status-card {
    padding: 12px;
    font-size: 13px;
  }
  
  .ExpectedToEnd {
    font-size: 0.9em;
    max-height: 100px;
    padding: 6px;
  }
  
  /* 小屏幕Header轮播图适配 */
  .page-header {
    border-radius: 0 0 15px 15px;
    padding: 25px 0 15px;
  }
  
  .header-carousel {
    border-radius: 0 0 15px 15px;
  }
  
  .carousel-slide img {
    border-radius: 0 0 15px 15px;
  }
  
  .page-header::before {
    border-radius: 0 0 15px 15px;
  }
  
  .header-title {
    font-size: 1.8rem;
  }
  
  .header-subtitle {
    font-size: 0.9rem;
  }
}

</style>