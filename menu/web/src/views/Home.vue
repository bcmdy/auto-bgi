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
      <div class="status-card ">
        <h2 class="tooltip-wrapper">执行配置组：🧩<span>{{ statusData.group }}
          <pre  class="tooltip-content">{{ statusData.ExpectedToEnd }}</pre>
        </span>
        
        
        </h2>
        <!-- <div class="tooltip-wrapper">
          详细
          <pre  class="tooltip-content">{{ statusData.ExpectedToEnd }}</pre>
        </div> -->
        <p><span>📜</span> 运行路线：<span>{{ statusData.line }}</span></p>
        <p><span>📜</span> 运行脚本：<span>{{ statusData.scriptName }}</span></p>
        <p><span>🗺️</span> 配置组进度：<span>{{ statusData.progress }}</span></p>
        <p><span>🖥️</span> 软件状态：<span>{{ statusData.running }}</span></p>
        <p><span>✨</span><span>{{ statusData.jsProgress }}</span></p>
        <p><span class="indexSX" @click="indexSXBtn">刷新</span></p>
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

  <!-- 一条龙选择模态框 -->
  <a-modal
    v-model:open="oneLongModal.visible"
    title="选择启动的一条龙"
    :confirm-loading="oneLongModal.loading"
    @ok="handleOneLongOk"
    @cancel="handleOneLongCancel"
    ok-text="启动"
    cancel-text="取消"
  >
    <div style="padding: 20px 0;">
      <a-select
        v-model:value="oneLongModal.selectedValue"
        placeholder="请选择一条龙"
        style="width: 100%"
        :loading="oneLongModal.loading"
      >
        <a-select-option
          v-for="item in oneLongModal.options"
          :key="item"
          :value="item"
        >
          {{ item }}
        </a-select-option>
      </a-select>
    </div>
  </a-modal>

  <!-- 查看桌面图片模态框 -->
  <a-modal
    v-model:open="screenshotModal.visible"
    title="桌面图片(3秒自动刷新一次)"
    :footer="null"
    :width="isMobile ? '96vw' : '90vw'"
    :afterClose="closeScreenshot"
    centered
  >
    <div ref="screenshotContainer" style="text-align: center; overflow: auto; max-height: 80vh;">
      <img
        ref="screenshotImage"
        :src="screenshotModal.url"
        alt="desktop-screenshot"
        @load="onScreenshotLoad"
        :style="{
          maxWidth: isZoomed ? 'none' : '100%',
          maxHeight: isZoomed ? 'none' : '80vh',
          transform: isZoomed ? `scale(${zoomScale})` : 'none',
          transformOrigin: 'top left',
          display: 'inline-block'
        }"
      />
    </div>

    <div class="screenshot-toolbar">
      <a-button :size="isMobile ? 'small' : 'middle'" :block="isMobile" @click="refreshScreenshot">刷新</a-button>
      <a-button :size="isMobile ? 'small' : 'middle'" :block="isMobile" @click="zoomOut" :disabled="!isZoomed && zoomScale <= 1">缩小</a-button>
      <a-button :size="isMobile ? 'small' : 'middle'" :block="isMobile" @click="zoomIn">放大</a-button>
      <a-button :size="isMobile ? 'small' : 'middle'" :block="isMobile" @click="fitImage">适应</a-button>
      <a-button :size="isMobile ? 'small' : 'middle'" :block="isMobile" @click="resetZoom">1:1</a-button>
      <a-button :size="isMobile ? 'small' : 'middle'" :block="isMobile" type="primary" @click="closeScreenshot">关闭</a-button>
    </div>
  </a-modal>

</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, computed, watch,h } from 'vue'
import { message, Modal ,Select} from 'ant-design-vue'
import { useRouter } from 'vue-router'
import  { apiMethods } from '@/utils/api'

const router = useRouter()

// 简易移动端检测与监听
const isMobile = ref(window.innerWidth <= 576)
const handleResizeForMobile = () => {
  isMobile.value = window.innerWidth <= 576
}
window.addEventListener('resize', handleResizeForMobile)

// 查看桌面图片 - 状态
const screenshotModal = reactive({
  visible: false,
  url: ''
})

const screenshotContainer = ref(null)
const screenshotImage = ref(null)
const isZoomed = ref(false)
const zoomScale = ref(1)

const refreshScreenshot = () => {
  const ts = Date.now()
  screenshotModal.url = `/api/aBgiJt?t=${ts}`
}

// 定义定时器相关变量
let screenshotTimer = null
const SCREENSHOT_INTERVAL = 3000 // 5秒刷新一次

const openScreenshot = () => {
    // 先停止可能存在的定时器
  if (screenshotTimer) {
    clearInterval(screenshotTimer)
    screenshotTimer = null
  }
  refreshScreenshot()
  screenshotModal.visible = true

    // 启动定时器
  screenshotTimer = setInterval(() => {
    console.log('定时刷新截图...')
    refreshScreenshot()
  }, SCREENSHOT_INTERVAL)
}

const closeScreenshot = () => {
  screenshotModal.visible = false
    // 停止定时器
  stopScreenshotTimer()
}

// 单独定义停止定时器的函数
const stopScreenshotTimer = () => {
  if (screenshotTimer) {
    clearInterval(screenshotTimer)
    screenshotTimer = null
    console.log('截图定时器已停止')
  }
}

const onScreenshotLoad = () => {
  fitImage()
}

const zoomIn = () => {
  isZoomed.value = true
  zoomScale.value = Math.min(zoomScale.value + 0.2, 6)
}

const zoomOut = () => {
  if (!isZoomed.value) return
  zoomScale.value = Math.max(zoomScale.value - 0.2, 0.2)
}

const resetZoom = () => {
  isZoomed.value = true
  zoomScale.value = 1
}

const fitImage = () => {
  // Reset first
  isZoomed.value = false
  zoomScale.value = 1
}


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
  jsProgress: '加载中...',
  scriptName: '加载中...'
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
  { text: '查看收获超过10', route: '/logAnalysis' },
  { text: '背包统计', route: '/BagStatistics' },
  { text: '查看狗粮日志', route: '/getAutoArtifactsPro' },
  { text: '归档查询', route: '/archive' },
  { text: '配置组运行情况', route: '/other' },
  { text: 'CD管理自动采集', route: '/CDAwareAutoGather' },
  { text: '日志查询', route: '/autoLog' },
  { text: '定时任务', route: '/TaskCron' }

])


const bgiButtons = ref([
    { text: '仓库提交记录', route: '/GitLog' },
  { text: '调度器', route: '/listGroups' },
  { text: '实时日志', route: '/log' },
    {text: '实时屏幕',route: '/screen'} ,
  { text: '录屏管理', route: '/obsVideo' },
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

// 一条龙选择相关的响应式数据
const oneLongModal = reactive({
  visible: false,
  loading: false,
  options: [],
  selectedValue: ''
})

// 按钮事件处理
const handleOneLong = async () => {
  try {
    oneLongModal.loading = true
    // 先获取一条龙列表
    const response = await apiMethods.getOneLongAllName()
    const oneLongList = response.data || []
    
    if (oneLongList.length === 0) {
      message.warning('没有可用的一条龙配置')
      return
    }

    oneLongModal.options = oneLongList
    oneLongModal.selectedValue = oneLongList[0]
    oneLongModal.visible = true
  } catch (error) {
    console.error('获取一条龙列表失败:', error)
    message.error('获取一条龙列表失败！')
  } finally {
    oneLongModal.loading = false
  }
}



// 处理一条龙启动确认
const handleOneLongOk = async () => {
  if (!oneLongModal.selectedValue) {
    message.warning('请选择一条龙')
    return
  }
  
  try {
    oneLongModal.loading = true
    await apiMethods.startOneLong(oneLongModal.selectedValue)
    message.success(`已启动【${oneLongModal.selectedValue}】！`)
    oneLongModal.visible = false
    oneLongModal.selectedValue = ''
  } catch (e) {
    message.error('启动失败！')
  } finally {
    oneLongModal.loading = false
  }
}

// 取消一条龙选择
const handleOneLongCancel = () => {
  oneLongModal.visible = false
  oneLongModal.selectedValue = ''
}




const handleCloseBgi = () => {
  Modal.confirm({
    title: '确认关闭？',
    content: '是否要关闭【BGI和原神】？',
    okText: '确定',
    cancelText: '取消',
    onOk: async () => {
      try {
        await apiMethods.closeBgi()
        message.success('关闭成功！')
      } catch (error) {
        message.error('关闭失败！')
      }
    }
  })
}

const handleBackup = async () => {
  try {
    await apiMethods.backup()
    message.success('备份成功！')
  } catch (error) {
    message.error('备份失败！')
  }
}

// 发送桌面截图
const sendImage = () => {
  Modal.confirm({
    title: '确认发送？',
    content: '是否要发送当前桌面截图？',
    okText: '确定',
    cancelText: '取消',
    onOk: async () => {
      try {
        const response = await apiMethods.sendImage()
        scanResult.value = response.data || '发送成功！'
        Modal.info({
          title: '发送结果',
          content: scanResult.value,
          okText: '关闭'
        })
      } catch (error) {
        message.error('发送失败！')
      }
    }
  })

}

//刷新
const indexSXBtn = () => {
  apiMethods.indexSX()
  
  refreshStatus()
  message.success('刷新成功！')
}

//更新updateABgi
const updateABgi = async () => {
  const response = await apiMethods.updateABgi()
  console.log("===============",response)
  if(response.code!=200){
      message.success('更新成功！请重启软件！')
    return
  }
 message.error('更新失败！请稍后再试！')

}

const mysSignIn = () => {
  Modal.confirm({
    title: '确认签到？',
    content: '是否要米游社签到？',
    okText: '确定',
    cancelText: '取消',
    onOk: async () => {
      try {
        const response = await apiMethods.mysSignIn()
        scanResult.value = response.message || '发送成功！'
        Modal.info({
          title: '发送结果',
          content: scanResult.value,
          okText: '关闭'
        })
      } catch (error) {
        message.error('发送失败！')
      }
    }
  })

}




// 自动化按钮配置
const automationButtons = ref([
  { text: '一条龙启动', action: handleOneLong },
  { text: '关闭BGI和原神', action: handleCloseBgi },
  { text: '备份 user 文件', action: handleBackup },
  { text: '脚本更新列表', action: () => router.push('/jsNames') },
  { text: '地图追踪', action: () => router.push('/Pathing') },
  { text: '发送桌面截图', action: sendImage },
  { text: '查看桌面图片', action: openScreenshot },
  { text: '米游社手动签到', action: mysSignIn },
  { text: '联机管理', action: () => router.push('/Online') }

])




// 生命周期
onMounted(() => {
  // 初始化樱花动画
  const cleanupAnimation = initSakuraAnimation()
  
  // 获取轮播图图片
  getImages()
  getHeaderImages() // 获取header轮播图图片
  
  // 获取状态并设置定时器
  refreshStatus()
  statusInterval = setInterval(refreshStatus, 3000)

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

/* 粉色卡通风确认框样式 */
.ant-modal-confirm {
  border-radius: 20px !important;
  overflow: hidden;
  animation: modalPop 0.3s ease-out;
}

.ant-modal-confirm .ant-modal-content {
  background: linear-gradient(135deg, #ffe6f0, #fff0f5) !important;
  border: 2px dashed #ffaad5;
  border-radius: 20px;
  box-shadow: 0 0 20px rgba(255, 174, 209, 0.5);
}

/* 工具栏响应式布局 */
.screenshot-toolbar {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
  margin-top: 12px;
}

/* Desktop: keep buttons inline, auto width */
.screenshot-toolbar :deep(.ant-btn) {
  flex: 0 0 auto;
  width: auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  line-height: 1.2;
  min-height: 36px;
}

@media (max-width: 576px) {
  .screenshot-toolbar {
    justify-content: stretch;
  }
  .screenshot-toolbar :deep(.ant-btn) {
    flex: 1 1 calc(50% - 8px);
    min-height: 32px;
  }
}

.ant-modal-confirm .ant-modal-confirm-title {
  color: #ff66a3 !important;
  font-weight: bold;
  text-shadow: 0 0 5px #fff0f5;
}

.ant-modal-confirm .ant-modal-confirm-content {
  color: #d9006c !important;
  font-size: 15px;
  font-weight: bold;
}

.indexSX{
  border: 3px solid #ff99cc;
  border-radius: 12px;
  cursor: pointer;
  padding: 2px 8px;
  user-select: none;
}
.ant-modal-confirm .ant-btn-primary {
  background: linear-gradient(145deg, #ff99cc, #ff66a3) !important;
  border: none !important;
  border-radius: 30px;
  box-shadow: 0 0 12px #ff99cc;
}

.ant-modal-confirm .ant-btn-primary:hover {
  background: linear-gradient(145deg, #ff66a3, #ff3385) !important;
  box-shadow: 0 0 18px #ff66a3;
}

.ant-modal-confirm .ant-btn-default {
  border-radius: 30px;
  background: #fff0f5 !important;
  border: 1px solid #ffb6c1 !important;
  color: #ff4081 !important;
}

.ant-modal-confirm .ant-btn-default:hover {
  background: #ffe6f0 !important;
  color: #d9006c !important;
}

@keyframes modalPop {
  0% { transform: scale(0.8); opacity: 0; }
  100% { transform: scale(1); opacity: 1; }
}


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

.tooltip-wrapper {
  position: relative;
  display: inline-block;
  cursor: pointer;
}

.tooltip-content {
  display: none;
  position: absolute;
  top: 120%;
  left: 50%;
  transform: translateX(-50%);
  background: #333;
  color: #fff;
  padding: 8px 12px;
  border-radius: 6px;
  white-space: nowrap;
}

.tooltip-wrapper:hover .tooltip-content {
  display: block;
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
    right: 10px;
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
