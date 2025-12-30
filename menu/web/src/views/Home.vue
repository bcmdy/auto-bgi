<template>
  <div class="home-container">
    <canvas ref="animeCanvas" class="anime-canvas"></canvas>

    <div class="side-carousel">
      <div class="carousel-wrapper">
        <transition-group name="fade">
          <img 
            v-if="currentImage" 
            :key="currentImage" 
            :src="currentImage" 
            class="carousel-img"
            alt="Character"
          />
        </transition-group>
      </div>
    </div>

    <div class="main-content">
      
      <header class="page-header">
        <div class="header-carousel">
            <div 
              v-for="(img, index) in headerCarouselImages" 
              :key="index"
              class="carousel-slide"
              :class="{ active: index === headerCurrentImageIndex }"
            >
              <img :src="img" alt="Header BG">
            </div>
        </div>
        <div class="header-content">
          <h1 class="header-title">✨ BGI 控制台 ✨</h1>
          <p class="header-subtitle">Auto-BGI Animation Dashboard</p>
        </div>
      </header>

      <div class="status-card glass-panel">
        <div class="card-header">
            <h2>🖥️ 运行状态监控</h2>
            <button class="refresh-btn" @click="indexSXBtn">🔄 刷新</button>
        </div>
        
        <div class="status-grid">
            <div class="status-item group-name">
                <span class="label">🧩 执行配置组:</span>
                <span class="value">{{ statusData.group }}</span>
                <div class="ExpectedToEnd">
                <pre >{{ statusData.ExpectedToEnd=="" ? '没有归档记录' : statusData.ExpectedToEnd }}</pre>
            </div>
            </div>
    
            <div class="status-item">
                <span class="label">📜 运行路线:</span>
                <span class="value">{{ statusData.line }}</span>
            </div>
            <div class="status-item">
                <span class="label">📜 运行脚本:</span>
                <span class="value">{{ statusData.scriptName }}</span>
            </div>
            <div class="status-item">
                <span class="label">🗺️ 进度:</span>
                <span class="value">{{ statusData.progress }}</span>
            </div>
            <div class="status-item">
                <span class="label">⚙️ 状态:</span>
                <span class="value">{{ statusData.running }}</span>
            </div>
            <div class="status-item full-width">
                <span class="label">✨ JS进度:</span>
                <span class="value">{{ statusData.jsProgress }}</span>
            </div>
        </div>

 
      </div>

      <div class="action-zone">

        <div class="button-group glass-panel">
            <h2 class="group-title">🔍 实时监测</h2>
          <div class="btn-grid">
            <button  @click="openScreenshot">查看桌面</button>
            <button  @click="sendImage">发送截图</button>
            <button  @click="router.push('/log')">实时日志</button>
          </div>
          
        </div>
        
        <div class="button-group glass-panel">
          <h2 class="group-title">📊 数据分析</h2>
          <div class="btn-grid">
            <button 
                v-for="(btn, index) in dataAnalysisButtons" 
                :key="index" 
                @click="router.push(btn.route)"
            >
                {{ btn.text }}
            </button>
          </div>
        </div>

        <div class="button-group glass-panel">
          <h2 class="group-title">🚀 自动化控制</h2>
          <div class="btn-grid">
             <button 
                v-for="(btn, index) in automationButtons" 
                :key="index" 
                @click="btn.action ? btn.action() : router.push(btn.route)"
            >
                {{ btn.text }}
            </button>
          </div>
        </div>

        <div class="button-group glass-panel">
          <h2 class="group-title">🧭 BGI 管理</h2>
          <div class="btn-grid">
            <button 
                v-for="(btn, index) in bgiButtons" 
                :key="index" 
                @click="btn.action ? btn.action() : router.push(btn.route)"
            >
                {{ btn.text }}
            </button>
          </div>
        </div>

      </div>
    </div>

    <a-modal
      v-model:open="oneLongModal.visible"
      title="🌸 选择启动的一条龙 🌸"
      :confirm-loading="oneLongModal.loading"
      @ok="handleOneLongOk"
      @cancel="handleOneLongCancel"
      ok-text="启动"
      cancel-text="取消"
      class="anime-modal"
    >
      <div style="padding: 20px 0;">
         <a-select
            v-model:value="oneLongModal.selectedValue"
            style="width: 100%"
            placeholder="请选择配置"
         >
            <a-select-option v-for="item in oneLongModal.options" :key="item" :value="item">
                {{ item }}
            </a-select-option>
         </a-select>
      </div>
    </a-modal>

    <a-modal
      v-model:open="screenshotModal.visible"
      title="🖥️ 桌面实时监控 (5s刷新)"
      :footer="null"
      :width="isMobile ? '98vw' : '80vw'"
      :afterClose="closeScreenshot"
      centered
      class="anime-modal"
    >
      <div class="screenshot-view" ref="screenshotContainer">
        <img 
            v-if="screenshotModal.url" 
            :src="screenshotModal.url" 
            :style="{ transform: `scale(${zoomScale})` }"
            @load="onScreenshotLoad"
            class="live-img"
        />
        <div v-else class="loading-placeholder">Waiting for signal...</div>
      </div>
      <div class="modal-tools">
         <button @click="refreshScreenshot">刷新</button>
         <button @click="zoomOut">缩小</button>
         <button @click="zoomIn">放大</button>
         <button @click="fitImage">适应</button>
         <button @click="closeScreenshot">关闭</button>
      </div>
    </a-modal>

    <a-modal
      v-model:open="uploadBgiModal.visible"
      title="📦 上传 BGI 更新包"
      :confirm-loading="uploadBgiModal.loading"
      @ok="handleUploadBgiOk"
      @cancel="handleUploadBgiCancel"
      ok-text="开始上传"
      cancel-text="取消"
      class="anime-modal"
    >
        <div class="upload-area">
            <input 
                type="file" 
                ref="bgiFileInput" 
                accept=".zip,.7z" 
                @change="handleBgiFileSelect"
                style="display: none"
            />
            <a-button size="large" @click="$refs.bgiFileInput.click()">
                📂 选择压缩包 (.zip / .7z)
            </a-button>
            <div v-if="uploadBgiModal.selectedFile" class="file-info">
                <p>已选: {{ uploadBgiModal.selectedFile.name }}</p>
                <p>大小: {{ (uploadBgiModal.selectedFile.size / 1024 / 1024).toFixed(2) }} MB</p>
            </div>
            <div v-if="uploadBgiModal.uploadProgress > 0" class="progress-bar">
                <div class="progress-fill" :style="{width: uploadBgiModal.uploadProgress + '%'}">
                    {{ uploadBgiModal.uploadProgress }}%
                </div>
            </div>
        </div>
    </a-modal>

  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, computed, watch, h } from 'vue'
import { message, Modal, Select } from 'ant-design-vue'
import { useRouter } from 'vue-router'
// 假设这里有您的API工具，如果报错请保留您原来的引入方式
import { apiMethods } from '@/utils/api' 

const router = useRouter()

// --- 移动端检测 ---
const isMobile = ref(window.innerWidth <= 576)
const handleResizeForMobile = () => {
  isMobile.value = window.innerWidth <= 576
}
window.addEventListener('resize', handleResizeForMobile)

// --- 截图功能 ---
const screenshotModal = reactive({ visible: false, url: '' })
const screenshotContainer = ref(null)
const isZoomed = ref(false)
const zoomScale = ref(1)
const token = localStorage.getItem('aBgiToken')
let screenshotTimer = null
const SCREENSHOT_INTERVAL = 5000 

const refreshScreenshot = () => {
  const ts = Date.now()
  screenshotModal.url = `/api/aBgiJt?t=${ts}&tk=${token}`
}

const openScreenshot = () => {
  if (screenshotTimer) clearInterval(screenshotTimer)
  refreshScreenshot()
  screenshotModal.visible = true
  screenshotTimer = setInterval(() => {
    console.log('自动刷新截图...')
    refreshScreenshot()
  }, SCREENSHOT_INTERVAL)
}

const stopScreenshotTimer = () => {
  if (screenshotTimer) {
    clearInterval(screenshotTimer)
    screenshotTimer = null
  }
}

const closeScreenshot = () => {
  screenshotModal.visible = false
  stopScreenshotTimer()
}

const onScreenshotLoad = () => { fitImage() }
const zoomIn = () => { isZoomed.value = true; zoomScale.value = Math.min(zoomScale.value + 0.2, 6) }
const zoomOut = () => { if (!isZoomed.value) return; zoomScale.value = Math.max(zoomScale.value - 0.2, 0.2) }
const fitImage = () => { isZoomed.value = false; zoomScale.value = 1 }

// --- 认证与基础 ---
const handleLogout = () => {
  try {
    localStorage.removeItem('aBgiToken')
    router.push('/login')
  } catch (err) {
    console.error(err)
    router.push('/login')
  }
}

// --- 动画与轮播 ---
const animeCanvas = ref(null)
const carouselImages = ref([])
const currentImageIndex = ref(0)
const headerCarouselImages = ref([])
const headerCurrentImageIndex = ref(0)
let headerCarouselInterval = null
let statusInterval = null
let petals = []
let animationId = null

const currentImage = computed(() => {
  if (carouselImages.value.length > 0) {
    return carouselImages.value[currentImageIndex.value]
  }
  return null // 或者默认图片
})

// --- 状态数据 ---
const statusData = reactive({
  group: '加载中...',
  ExpectedToEnd: '...',
  line: '...',
  progress: '...',
  running: '...',
  jsProgress: '...',
  scriptName: '...'
})

// --- 按钮配置 (保持不变) ---
const dataAnalysisButtons = ref([
  { text: '查看狗粮日志', route: '/getAutoArtifactsPro' },
  { text: '屑荧进村', route: '/logAnalysis' },
  { text: '归档查询', route: '/archive' },
  { text: '旅行者轧记', route: '/BagStatistics' },
  { text: '配置组运行情况', route: '/other' },
  { text: 'CD管理自动采集', route: '/CDAwareAutoGather' },
  { text: '采集管理', route: '/CollectionManagement' },
  { text: 'ABGI日志查询', route: '/autoLog' },
  { text: 'ABGI定时任务', route: '/TaskCron' }
])

// --- BGI上传逻辑 ---
const uploadBgiModal = reactive({ visible: false, loading: false, selectedFile: null, uploadProgress: 0 })
const bgiFileInput = ref(null)

const handleUploadBgiClick = () => {
  uploadBgiModal.selectedFile = null
  uploadBgiModal.uploadProgress = 0
  uploadBgiModal.visible = true
}

// setInterval(() => {
//   debugger
// }, 100)


const handleBgiFileSelect = (event) => {
  const file = event.target.files?.[0]
  if (!file) return
  if (!file.name.endsWith('.zip') && !file.name.endsWith('.7z')) {
    message.error('只能选择 .zip 或 .7z！')
    return
  }
  if (file.size > 500 * 1024 * 1024) {
    message.error('文件过大！')
    return
  }
  uploadBgiModal.selectedFile = file
}

const handleUploadBgiOk = async () => {
  if (!uploadBgiModal.selectedFile) return message.warning('请先选择文件')
  uploadBgiModal.loading = true
  
  const formData = new FormData()
  formData.append('file', uploadBgiModal.selectedFile)
  const xhr = new XMLHttpRequest()
  
  xhr.upload.addEventListener('progress', (e) => {
    if (e.lengthComputable) {
      uploadBgiModal.uploadProgress = Math.round((e.loaded / e.total) * 100)
    }
  })
  
  xhr.addEventListener('load', () => {
    uploadBgiModal.loading = false
    if (xhr.status === 200) {
      message.success('更新成功，请重启')
      uploadBgiModal.visible = false
    } else {
      message.error('更新失败')
    }
  })
  
  xhr.addEventListener('error', () => { uploadBgiModal.loading = false; message.error('网络错误') })
  
  try {
    const token = localStorage.getItem('aBgiToken')
    xhr.open('POST', '/api/UpdateBgi/Upload')
    if (token) xhr.setRequestHeader('Authorization', token)
    xhr.send(formData)
  } catch (e) {
    uploadBgiModal.loading = false
  }
}

const handleUploadBgiCancel = () => { uploadBgiModal.visible = false }

// --- 其他功能按钮逻辑 ---
const mysSignIn = () => {
    Modal.confirm({
        title: '确认签到？', content: '是否要米游社签到？', okText: '确定', cancelText: '取消',
        onOk: async () => {
            try {
                const res = await apiMethods.mysSignIn()
                Modal.info({ title: '结果', content: res.message || '发送成功' })
            } catch(e) { message.error('失败') }
        }
    })
}

const handleCloseBgi = () => {
    Modal.confirm({
        title: '确认关闭？', content: '是否关闭【BGI和原神】？',
        onOk: async () => {
            try { await apiMethods.closeBgi(); message.success('已发送关闭指令') } 
            catch(e) { message.error('失败') }
        }
    })
}

const handleBackup = () => {
    Modal.confirm({
        title: '确认备份？',
        content: '是否确认备份当前的 USER 文件？',
        okText: '确定',
        cancelText: '取消',
        centered: true, // 居中显示
        onOk: async () => {
            try { 
                await apiMethods.backup()
                message.success('备份成功') 
            } catch(e) { 
                message.error('备份失败') 
            }
        }
    })
}

const sendImage = () => {
    Modal.confirm({
        title: '发送截图', content: '确认发送当前截图？',
        onOk: async () => {
             try { const res = await apiMethods.sendImage(); Modal.info({content: res.data || '成功'}) }
             catch(e) { message.error('失败') }
        }
    })
}

const indexSXBtn = () => {
    apiMethods.indexSX()
    refreshStatus()
    message.success('刷新指令已发送')
}

// --- 按钮定义 ---
const automationButtons = ref([
  { text: '一条龙启动', action: () => { oneLongModal.visible = true; handleOneLongLoad() } },
  { text: '关闭BGI和原神', action: handleCloseBgi },
  { text: '调度器启动', route: '/listGroups' },
  { text: '备份 USER 文件', action: handleBackup },
  { text: '脚本屋', route: '/jsNames' },
  { text: '地图追踪', route: '/Pathing' },
  { text: '联机管理', route: '/Online' },
  { text: 'BGI一条龙配置', route: '/BgiConfig' },
])

const bgiButtons = ref([
  { text: '录屏管理', route: '/obsVideo' },
  { text: '仓库管理', route: '/GitLog' },
  { text: '手动更新BGI', action: handleUploadBgiClick },
  { text: '米游社签到', action: mysSignIn },
  { text: '检查更新', action: () => router.push('/Update') },
  { text: 'ABGI设置', route: '/Config' },
  { text: '退出登录', action: handleLogout }
])

// --- 一条龙逻辑 ---
const oneLongModal = reactive({ visible: false, loading: false, options: [], selectedValue: '' })
const handleOneLongLoad = async () => {
    try {
        oneLongModal.loading = true
        const res = await apiMethods.getOneLongAllName()
        oneLongModal.options = res.data || []
        if(oneLongModal.options.length) oneLongModal.selectedValue = oneLongModal.options[0]
    } catch(e) { message.error('加载列表失败') } finally { oneLongModal.loading = false }
}
const handleOneLongOk = async () => {
    if(!oneLongModal.selectedValue) return
    try {
        oneLongModal.loading = true
        await apiMethods.startOneLong(oneLongModal.selectedValue)
        message.success(`启动 ${oneLongModal.selectedValue}`)
        oneLongModal.visible = false
    } catch(e) { message.error('启动失败') } finally { oneLongModal.loading = false }
}
const handleOneLongCancel = () => { oneLongModal.visible = false }

// --- 樱花动画类 ---
class Petal {
  constructor(canvas) {
    this.canvas = canvas
    this.x = Math.random() * canvas.width
    this.y = Math.random() * canvas.height * -1 - 100
    this.size = Math.random() * 8 + 5
    this.speed = Math.random() * 2 + 0.5
    this.angle = Math.random() * 360
    this.spin = Math.random() * 5 - 2.5
    this.color = ["#ffcce6", "#ffd1e0", "#ff9ecd"][Math.floor(Math.random() * 3)]
  }
  update() {
    this.y += this.speed
    this.x += Math.sin(this.angle * Math.PI / 180) * 0.5
    this.angle += this.spin
    if (this.y > this.canvas.height) {
      this.y = -20
      this.x = Math.random() * this.canvas.width
    }
  }
  draw(ctx) {
    ctx.save()
    ctx.translate(this.x, this.y)
    ctx.rotate(this.angle * Math.PI / 180)
    ctx.fillStyle = this.color
    ctx.beginPath()
    ctx.arc(0, 0, this.size/2, 0, Math.PI*2)
    ctx.fill()
    ctx.restore()
  }
}

// --- 初始化与生命周期 ---
const initSakuraAnimation = () => {
    const canvas = animeCanvas.value
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    const resize = () => { canvas.width = window.innerWidth; canvas.height = window.innerHeight }
    resize()
    window.addEventListener('resize', resize)
    petals = Array.from({ length: 40 }, () => new Petal(canvas))
    const animate = () => {
        ctx.clearRect(0, 0, canvas.width, canvas.height)
        petals.forEach(p => { p.update(); p.draw(ctx) })
        animationId = requestAnimationFrame(animate)
    }
    animate()
    return () => { window.removeEventListener('resize', resize); cancelAnimationFrame(animationId) }
}

const refreshStatus = async () => {
    try {
        const res = await apiMethods.getStatus()
        Object.assign(statusData, res)
    } catch(e) { console.error(e) }
}

const getImages = async () => {
    try {
        const res = await fetch('/api/images')
        const data = await res.json()
        carouselImages.value = data.images || []
        if(carouselImages.value.length) startCarousel()
    } catch(e) {
        carouselImages.value = ['/img/bd.jpg', '/img/ff.png'] // Fallback
        startCarousel()
    }
}
const startCarousel = () => {
    setInterval(() => {
        currentImageIndex.value = (currentImageIndex.value + 1) % carouselImages.value.length
    }, 10000)
}

const getHeaderImages = async () => {
    try {
        const res = await fetch('/api/images') // Or separate API
        const data = await res.json()
        headerCarouselImages.value = data.images || []
        if(headerCarouselImages.value.length) startHeaderCarousel()
    } catch(e) {
        headerCarouselImages.value = ['/img/bd.jpg'] 
        startHeaderCarousel()
    }
}
const startHeaderCarousel = () => {
    headerCarouselInterval = setInterval(() => {
        headerCurrentImageIndex.value = (headerCurrentImageIndex.value + 1) % headerCarouselImages.value.length
    }, 6000)
}

onMounted(() => {
    const cleanup = initSakuraAnimation()
    getImages()
    getHeaderImages()
    refreshStatus()
    statusInterval = setInterval(refreshStatus, 3000)
    
    onUnmounted(() => {
        cleanup && cleanup()
        if (statusInterval) clearInterval(statusInterval)
        if (headerCarouselInterval) clearInterval(headerCarouselInterval)
    })
})
</script>

<style scoped>
/* ==== 全局布局与背景 ==== */
.home-container {
  position: relative;
  min-height: 100vh;
  background-color: #ffecf5;
  font-family: 'Comic Sans MS', 'Microsoft YaHei', sans-serif;
  overflow-x: hidden;
  /* 波点背景 */
  background-image: radial-gradient(#ffcce6 2px, transparent 2px);
  background-size: 20px 20px;
}

.anime-canvas {
  position: fixed;
  top: 0; left: 0;
  width: 100%; height: 100%;
  z-index: 5;
  pointer-events: none;
}

/* ==== 轮播图 (左下角) ==== */
.side-carousel {
  position: fixed;
  bottom: 0;
  left: 0;
  width: 45vw;
  max-width: 500px;
  height: auto;
  z-index: 0; /* 最底层 */
  pointer-events: none; /* 点击穿透 */
}

.carousel-wrapper {
  position: relative;
  width: 100%;
  padding-bottom: 120%; /* Aspect Ratio placeholder */
}

.ExpectedToEnd{
  background: rgb(252, 207, 230);
  position: absolute; 
  opacity: 0;
  display: none;
  transition: all .2s ease;
  border-radius: 5px;
}

.group-name:hover .ExpectedToEnd{
  opacity: 1;
  display: block;
  visibility: visible;
}

.carousel-img {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: auto;
  max-height: 80vh;
  object-fit: contain;
  object-position: bottom left;
  mask-image: linear-gradient(to top, black 70%, transparent 100%);
  -webkit-mask-image: linear-gradient(to top, black 70%, transparent 100%);
}

.fade-enter-active, .fade-leave-active { transition: opacity 1s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

/* ==== 主内容区域 ==== */
.main-content {
  position: relative;
  z-index: 10; /* 保证在轮播图之上 */
  width: 92%;
  max-width: 650px;
  margin: 0 auto;
  padding-bottom: 50px;
}

/* 玻璃拟态面板通用样式 */
.glass-panel {
  background: rgba(255, 255, 255, 0.65);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 2px solid rgba(255, 255, 255, 0.8);
  box-shadow: 0 8px 32px rgba(255, 105, 180, 0.15);
  border-radius: 24px;
  padding: 20px;
  margin-bottom: 24px;
}

/* Header 样式 */
.page-header {
  position: relative;
  height: 180px;
  border-radius: 0 0 30px 30px;
  overflow: hidden;
  margin-bottom: 25px;
  box-shadow: 0 5px 15px rgba(255,105,180,0.3);
}

.header-carousel {
  position: absolute;
  top: 0; left: 0; width: 100%; height: 100%;
}

.carousel-slide {
  position: absolute;
  width: 100%; height: 100%;
  opacity: 0; transition: opacity 1s;
}
.carousel-slide.active { opacity: 1; }
.carousel-slide img { 
  width: 100%; 
  height: 250%; 
  object-fit: cover; 
}

.header-content {
  position: relative;
  z-index: 2;
  text-align: center;
  padding-top: 50px;
  text-shadow: 0 2px 4px rgba(255,255,255,0.8);
}

.header-title {
  font-size: 2.2rem;
  color: #ff3385;
  margin: 0;
  font-weight: 800;
}
.header-subtitle {
  color: #ff66a3;
  font-size: 1rem;
  background: rgba(255,255,255,0.7);
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
}

/* 状态卡片 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  border-bottom: 2px dashed #ffb6c1;
  padding-bottom: 10px;
}
.card-header h2 { margin: 0; color: #ff3385; font-size: 1.2rem; }
.refresh-btn {
  background: #ffecf5; color: #ff3385;
  border: 1px solid #ff99cc;
  padding: 4px 12px; border-radius: 15px;
  cursor: pointer;
  width: auto;
}

.status-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}
.status-item {
  background: rgba(255,255,255,0.5);
  padding: 8px;
  border-radius: 12px;
  font-size: 14px;
}
.full-width { grid-column: span 2; }
.label { color: #ff80ab; font-weight: bold; margin-right: 5px; }
.value { color: #d81b60; font-weight: bold; word-break: break-all; }
.value.highlight { font-size: 1.1em; color: #c2185b; }

.screenshot-toolbar {
  margin-top: 15px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}
.tool-label { color: #ff3385; font-weight: bold; width: 100%; margin-bottom: 5px; }

/* 按钮组样式 */
.group-title {
  color: #ff6699;
  text-align: center;
  margin-bottom: 15px;
  font-size: 1.1rem;
  text-shadow: 1px 1px 0 #fff;
}

.btn-grid {
  display: grid;
  grid-template-columns: 1fr 1fr; /* 强制两列 */
  gap: 12px;
}

button {
  width: 100%;
  padding: 12px 5px;
  font-size: 14px;
  font-weight: bold;
  color: #fff;
  border: none;
  border-radius: 18px;
  cursor: pointer;
  background: linear-gradient(135deg, #ff99cc 0%, #ff66a3 100%);
  box-shadow: 0 4px 10px rgba(255, 105, 180, 0.3);
  transition: all 0.2s;
  position: relative;
  overflow: hidden;
  text-shadow: 0 1px 1px rgba(0,0,0,0.1);
}

button:active { transform: scale(0.95); }
button::after {
  content: '';
  position: absolute;
  top: 0; left: -100%;
  width: 100%; height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.4), transparent);
  transition: 0.5s;
}
button:hover::after { left: 100%; }

/* 模态框美化 */
.anime-modal :deep(.ant-modal-content) {
  border-radius: 20px;
  border: 3px solid #ffcce6;
  background: #fff0f5;
}
.anime-modal :deep(.ant-modal-header) {
  background: transparent;
  border-bottom: 2px dashed #ffb6c1;
}
.anime-modal :deep(.ant-modal-title) {
  color: #ff3385;
  text-align: center;
}

/* 截图查看器 */
.screenshot-view {
  background: #000;
  border-radius: 10px;
  min-height: 200px;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
  margin-bottom: 10px;
}
.live-img { max-width: 100%; transition: transform 0.3s ease; }
.modal-tools { 
  display: flex; 
  gap: 14px; 
  justify-content: center; 
  align-items: center;
  flex-wrap: nowrap; 
    align-items: center; /* 垂直居中 */
}
.modal-tools button {
  padding: 8px 14px;
  font-size: 20px;

}



.loading-placeholder { color: #ff66a3; }

/* ==== 移动端适配特别处理 ==== */
@media (max-width: 576px) {
  .side-carousel {
    width: 120vw; /* 移动端轮播图变大一点作为背景 */
    opacity: 0.8;
  }
  
  .main-content {
    width: 95%;
  }

  .status-grid {
    font-size: 12px;
  }
  
    .modal-tools button {
    font-size: 15px;
    padding: 6px 10px;
  }
  /* 确保按钮在移动端清晰且不拥挤 */
  button {
    font-size: 13px;
    padding: 10px 2px;
  }

  .glass-panel {
    /* 移动端增强模糊，确保背景图片不干扰文字 */
    backdrop-filter: blur(15px);
    background: rgba(255, 255, 255, 0.75);
  }
}
</style>