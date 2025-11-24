<template>
  <div class="container" :class="{ 'hacker': isHackerTheme }">
    <!-- 背景动画canvas -->
    <canvas ref="animeStars" id="animeStars"></canvas>

    <!-- 页面头部 -->
    <header>
      <button class="homeBtn" @click="goHome">返回首页</button>
      <h1>实时日志查看</h1>
      <select id="logSelector" v-model="selectedLog" @change="onLogChange">
        <option v-for="file in logFiles" :key="file" :value="file">
          {{ file }}
        </option>
      </select>
      <button id="themeToggle" @click="toggleTheme">切换黑客风</button>
    </header>

    <!-- 主内容区域 -->
    <main>
      <div id="log" ref="logContainer">{{ logContent }}</div>
    </main>

    <!-- 轮播图背景 -->
    <div class="right-bg-swiper">
      <div class="swiper-wrapper" ref="swiperWrapper">
        <!-- 动态插入图片 -->
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { Swiper } from 'swiper/bundle'
import 'swiper/css/bundle'

const router = useRouter()

// 响应式数据
const animeStars = ref(null)
const logContainer = ref(null)
const swiperWrapper = ref(null)
const selectedLog = ref('')
const logFiles = ref([])
const logContent = ref('正在加载日志列表...')
const isHackerTheme = ref(false)

// 常量定义
const CONSTANTS = {
  STAR_COUNT: 100,
  SWIPER_CONFIG: {
    delay: 10000,  // 10秒切换一次
    speed: 1000
  },
  STATIC_IMAGES: ['bd.jpg', 'ff.png', 'ng.jpg', 'sh.jpg']
}

// WebSocket 相关
let ws = null
let canvas = null
let ctx = null
let width = 0
let height = 0
const stars = []
let mySwiper = null

// 路由跳转到首页
const goHome = () => {
  router.push('/')
}

// 切换主题
const toggleTheme = () => {
  isHackerTheme.value = !isHackerTheme.value
  
  // 重新设置星星效果以适应新主题
  setupStars()
  
  // 确保日志容器在主题切换后仍然可见
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.style.display = 'block'
      logContainer.value.style.visibility = 'visible'
      logContainer.value.style.opacity = '1'
    }
  })
}

// 日志选择变化
const onLogChange = () => {
  if (selectedLog.value) {
    connectWebSocket(selectedLog.value)
  }
}

// 设置背景动画
const setupStars = () => {
  if (!animeStars.value) return
  
  canvas = animeStars.value
  ctx = canvas.getContext('2d')
  width = window.innerWidth
  height = window.innerHeight
  canvas.width = width
  canvas.height = height
  
  stars.length = 0
  for (let i = 0; i < CONSTANTS.STAR_COUNT; i++) {
    const isHacker = isHackerTheme.value
    stars.push({
      x: Math.random() * width,
      y: Math.random() * height,
      size: Math.random() * (isHacker ? 4 : 3) + 1,
      speed: Math.random() * (isHacker ? 0.1 : 0.05) + 0.01,
      brightness: 0,
      maxBrightness: Math.random() * 0.8 + 0.2,
      increasing: true,
      color: isHacker 
        ? (Math.random() > 0.3 ? '#00ff41' : '#00ff88')
        : (Math.random() > 0.5 ? '#d9b3ff' : '#b385ff'),
      // 黑客主题下添加下降效果
      fallSpeed: isHacker ? Math.random() * 2 + 0.5 : 0
    })
  }
}

// 绘制星星动画
const drawStars = () => {
  if (!ctx) return
  
  ctx.clearRect(0, 0, width, height)
  const isHacker = isHackerTheme.value
  
  stars.forEach(star => {
    star.brightness += star.increasing ? star.speed : -star.speed
    if (star.brightness >= star.maxBrightness || star.brightness <= 0) {
      star.increasing = !star.increasing
    }
    
    // 黑客主题下的下降效果
    if (isHacker && star.fallSpeed) {
      star.y += star.fallSpeed
      if (star.y > height) {
        star.y = -star.size
        star.x = Math.random() * width
      }
    }
    
    ctx.beginPath()
    ctx.fillStyle = star.color
    ctx.globalAlpha = star.brightness
    
    if (isHacker) {
      // 黑客主题下绘制数字/字符效果
      const chars = ['0', '1', '/', '+', '-', '*', '=', '<', '>', '[', ']']
      const char = chars[Math.floor(Math.random() * chars.length)]
      ctx.font = `${star.size * 4}px Courier New`
      ctx.textAlign = 'center'
      ctx.fillText(char, star.x, star.y)
      
      // 添加光晕效果
      ctx.shadowColor = star.color
      ctx.shadowBlur = star.size * 2
      ctx.fillText(char, star.x, star.y)
      ctx.shadowBlur = 0
    } else {
      // 普通主题下绘制星星
      const spikes = 5
      const outerRadius = star.size
      const innerRadius = star.size / 2
      let rot = Math.PI / 2 * 3
      const step = Math.PI / spikes
      
      ctx.moveTo(star.x, star.y - outerRadius)
      for (let i = 0; i < spikes; i++) {
        ctx.lineTo(star.x + Math.cos(rot) * outerRadius, star.y + Math.sin(rot) * outerRadius)
        rot += step
        ctx.lineTo(star.x + Math.cos(rot) * innerRadius, star.y + Math.sin(rot) * innerRadius)
        rot += step
      }
      ctx.closePath()
      ctx.fill()
    }
    
    ctx.globalAlpha = 1
  })
  
  requestAnimationFrame(drawStars)
}

// WebSocket连接 - 优化版本
const connectWebSocket = (name) => {
  if (ws) {
    ws.close()
    ws = null
  }
  
  logContent.value = `正在连接 ${name} 日志服务...\n`

  try {
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
    const wsUrl = `${protocol}://${location.host}/ws/${encodeURIComponent(name)}`
    ws = new WebSocket(wsUrl)

    ws.onmessage = (e) => {
      if (!logContainer.value) return
      
      const atBottom = logContainer.value.scrollHeight - logContainer.value.scrollTop <= logContainer.value.clientHeight + 10
      logContent.value += e.data
      
      if (atBottom) {
        nextTick(() => {
          if (logContainer.value) {
            logContainer.value.scrollTop = logContainer.value.scrollHeight
          }
        })
      }
    }
    
    ws.onopen = () => {
      logContent.value += `[已连接到 ${name}]\n`
      console.log(`WebSocket 连接成功: ${name}`)
    }
    
    ws.onclose = (event) => {
      logContent.value += `\n[连接已关闭 - 代码: ${event.code}]`
      console.log(`WebSocket 连接关闭: ${name}, 代码: ${event.code}`)
    }
    
    ws.onerror = (error) => {
      logContent.value += "\n[连接出错]"
      console.error(`WebSocket 连接错误: ${name}`, error)
    }
  } catch (error) {
    logContent.value += `\n[创建WebSocket连接失败: ${error.message}]`
    console.error('WebSocket 创建失败:', error)
  }
}

// 加载日志文件列表
const loadLogFiles = async () => {
  try {
    const res = await fetch('/api/logFiles')
    const data = await res.json()
    if (data.files?.length) {
      logFiles.value = data.files
      selectedLog.value = data.files[0]
      connectWebSocket(data.files[0])
    } else {
      logContent.value = "未找到日志文件。"
    }
  } catch (err) {
    logContent.value = "加载日志列表失败。\n" + err
  }
}

// 预加载图片
const preloadImages = () => {
  return Promise.all(CONSTANTS.STATIC_IMAGES.map(imgSrc => {
    return new Promise((resolve, reject) => {
      const img = new Image()
      img.onload = () => resolve(imgSrc)
      img.onerror = () => {
        console.warn(`图片预加载失败: ${imgSrc}`)
        resolve(null) // 不阻塞其他图片的加载
      }
      img.src = `/static/image/${imgSrc}`
    })
  }))
}

// 获取轮播图片 - 直接使用静态目录中的图片
const getImages = async () => {
  try {
    if (!swiperWrapper.value) {
      console.error("找不到轮播容器")
      return
    }
    
    console.log('开始加载轮播图片...')
    
    // 预加载图片
    const loadedImages = await preloadImages()
    const validImages = loadedImages.filter(img => img !== null)
    
    console.log('有效图片数量:', validImages.length, validImages)
    
    if (validImages.length === 0) {
      console.warn("没有可用的图片")
      return
    }
    
    // 确保至少有2张图片才能轮播
    if (validImages.length < 2) {
      console.warn("图片数量不足，无法轮播")
      return
    }
    
    swiperWrapper.value.innerHTML = ''
    
    // 等待所有图片完全加载并获取尺寸信息
    const imagePromises = validImages.map((imgSrc, i) => {
      return new Promise((resolve) => {
        const slide = document.createElement('div')
        slide.classList.add('swiper-slide')
        const img = document.createElement('img')
        img.src = `/static/image/${imgSrc}`
        img.alt = `轮播图${i + 1}`
        img.onload = () => {
          console.log(`图片 ${imgSrc} 加载完成，原始尺寸: ${img.naturalWidth}x${img.naturalHeight}`)
          
          // 根据图片比例调整显示
          const aspectRatio = img.naturalWidth / img.naturalHeight
          console.log(`图片宽高比: ${aspectRatio.toFixed(2)}`)
          
          // 根据图片比例设置样式以确保完整显示
          img.style.width = 'auto'
          img.style.height = 'auto'
          img.style.objectFit = 'contain'
          
          if (aspectRatio > 1.2) {
            // 横向图片：限制宽度
            img.style.maxWidth = '100%'
            img.style.maxHeight = '90vh'
          } else if (aspectRatio < 0.8) {
            // 纵向图片：限制高度
            img.style.maxWidth = '100%'
            img.style.maxHeight = '94vh'
          } else {
            // 近似正方形图片
            img.style.maxWidth = '100%'
            img.style.maxHeight = '92vh'
          }
          
          resolve()
        }
        img.onerror = () => {
          console.error(`图片 ${imgSrc} 加载失败`)
          resolve()
        }
        slide.appendChild(img)
        swiperWrapper.value.appendChild(slide)
      })
    })
    
    await Promise.all(imagePromises)
    console.log('所有图片加载完成')
    
    // 在下一个tick中初始化Swiper，确保DOM更新完成
    await nextTick()
    
    // 销毁现有的Swiper实例
    if (mySwiper) {
      mySwiper.destroy(true, true)
      mySwiper = null
    }
    
    console.log('开始初始化Swiper...')
    
    mySwiper = new Swiper('.right-bg-swiper', {
      // 基本配置
      slidesPerView: 1,
      spaceBetween: 0,
      
      // 循环配置
      loop: true,
      
      // 自动播放配置
      autoplay: {
        delay: CONSTANTS.SWIPER_CONFIG.delay,
        disableOnInteraction: false,
        pauseOnMouseEnter: false
      },
      
      // 切换效果
      effect: 'fade',
      fadeEffect: {
        crossFade: true
      },
      
      // 切换速度
      speed: CONSTANTS.SWIPER_CONFIG.speed,
      
      // 禁用触摸
      allowTouchMove: false,
      
      // 事件回调
      on: {
        init: function() {
          const swiper = this
          console.log('=== Swiper初始化完成 ===')
          console.log('Swiper实例:', swiper)
          console.log('幻灯片数量:', swiper.slides.length)
          console.log('当前索引:', swiper.activeIndex)
          console.log('真实索引:', swiper.realIndex)
          console.log('自动播放配置:', swiper.autoplay)
          console.log('循环配置:', swiper.params.loop)
          
          // 检查DOM结构
          const slides = document.querySelectorAll('.right-bg-swiper .swiper-slide')
          console.log('DOM中的幻灯片:', slides.length)
          slides.forEach((slide, index) => {
            console.log(`幻灯片 ${index}:`, slide)
          })
        },
        slideChange: function() {
          const swiper = this
          console.log('=== 幻灯片切换 ===')
          console.log('切换到索引:', swiper.activeIndex)
          console.log('真实索引:', swiper.realIndex)
          console.log('总幻灯片数:', swiper.slides.length)
        },
        autoplayStart: function() {
          console.log('=== 自动播放开始 ===')
          console.log('延迟时间:', this.autoplay.delay, 'ms')
        },
        autoplayStop: function() {
          console.log('=== 自动播放停止 ===')
        }
      }
    })
    
    console.log('Swiper创建完成:', mySwiper)
    
    // 使用setTimeout确保实例完全创建后再操作
    setTimeout(() => {
      if (mySwiper && mySwiper.autoplay) {
        console.log('启动自动播放，当前状态:', mySwiper.autoplay.running)
        mySwiper.autoplay.start()
        console.log('手动启动自动播放完成')
      }
      
      // 输出调试信息
      if (mySwiper) {
        console.log('最终Swiper状态:')
        console.log('- 幻灯片数量:', mySwiper.slides?.length)
        console.log('- 当前索引:', mySwiper.activeIndex)
        console.log('- 自动播放运行:', mySwiper.autoplay?.running)
        console.log('- 循环模式:', mySwiper.params.loop)
      }
    }, 100)
    
    console.log('Swiper实例:', mySwiper)
  } catch (err) {
    console.error("轮播图加载失败：", err)
  }
}

// 窗口大小变化处理
const handleResize = () => {
  setupStars()
}

// 组件挂载
onMounted(() => {
  setupStars()
  drawStars()
  loadLogFiles()
  getImages()
  window.addEventListener('resize', handleResize)
  
  // 确保组件可见
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.style.display = 'block'
      logContainer.value.style.visibility = 'visible'
      logContainer.value.style.opacity = '1'
    }
  })
})

// 组件卸载
onUnmounted(() => {
  if (ws) ws.close()
  window.removeEventListener('resize', handleResize)
  if (mySwiper) {
    mySwiper.destroy(true, true)
  }
})
</script>

<style scoped>
* { 
  box-sizing: border-box; 
  margin: 0; 
  padding: 0; 
}

.container {
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  font-family: "Comic Sans MS", "Mochiy Pop One", "Segoe UI", sans-serif;
  transition: all 0.3s ease;
  position: relative;
  background-color: #f5f0ff;
  background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100"><path d="M30,50 Q50,30 70,50 Q50,70 30,50 Z" fill="%23e0d6ff" opacity="0.4"/></svg>');
  color: #9966cc;
}

.container.hacker {
  background-color: #000;
  background-image: 
    linear-gradient(rgba(0, 255, 0, 0.03) 50%, transparent 50%),
    radial-gradient(circle at 25% 25%, #00ff41 1px, transparent 1px),
    radial-gradient(circle at 75% 75%, #00ff88 1px, transparent 1px);
  background-size: 100% 2px, 50px 50px, 30px 30px;
  color: #00ff41;
  font-family: 'Courier New', 'Consolas', monospace;
  animation: scanlines 2s linear infinite;
}

@keyframes scanlines {
  0% { background-position: 0 0, 0 0, 0 0; }
  100% { background-position: 0 100%, 0 0, 0 0; }
}

canvas#animeStars {
  position: fixed;
  top: 0; 
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
  background-color: transparent;
  transition: all 0.3s ease;
}

.container.hacker canvas#animeStars {
  background: 
    radial-gradient(ellipse at center, rgba(0, 20, 0, 0.1) 0%, rgba(0, 0, 0, 0.8) 100%),
    repeating-linear-gradient(
      90deg,
      transparent,
      transparent 2px,
      rgba(0, 255, 65, 0.01) 2px,
      rgba(0, 255, 65, 0.01) 4px
    );
  filter: contrast(1.2);
}

header {
  background-color: rgba(255, 240, 255, 0.8);
  padding: 1rem;
  text-align: center;
  border-bottom: 1px solid #d9b3ff;
  position: relative;
  z-index: 2;
  border-radius: 0 0 20px 20px;
  box-shadow: 0 4px 10px rgba(153, 102, 204, 0.2);
  transition: all 0.3s ease;
}

.container.hacker header {
  background: linear-gradient(135deg, rgba(0, 0, 0, 0.9) 0%, rgba(0, 20, 0, 0.8) 100%);
  border-bottom: 2px solid #00ff41;
  border-image: linear-gradient(90deg, transparent, #00ff41, #00ff88, #00ff41, transparent) 1;
  box-shadow: 
    0 4px 20px rgba(0, 255, 65, 0.3),
    inset 0 1px 0 rgba(0, 255, 65, 0.1);
  position: relative;
}

.container.hacker header::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, #00ff41, transparent);
  animation: glow-line 3s ease-in-out infinite alternate;
}

@keyframes glow-line {
  0% { opacity: 0.5; transform: scaleX(0.8); }
  100% { opacity: 1; transform: scaleX(1); }
}

h1 {
  margin: 0 0 0.5rem;
  font-size: 1.5rem;
  color: #9966cc;
  text-shadow: 1px 1px 3px #d9b3ff;
  transition: all 0.3s ease;
}

.container.hacker h1 {
  color: #00ff41;
  text-shadow: 
    0 0 5px #00ff41,
    0 0 10px #00ff41,
    0 0 15px #00ff41,
    0 0 20px #00ff88;
  font-weight: bold;
  letter-spacing: 2px;
  text-transform: uppercase;
  animation: text-flicker 2s linear infinite;
  position: relative;
}

.container.hacker h1::after {
  content: '_';
  animation: cursor-blink 1s infinite;
}

@keyframes text-flicker {
  0%, 19%, 21%, 23%, 25%, 54%, 56%, 100% {
    text-shadow: 
      0 0 5px #00ff41,
      0 0 10px #00ff41,
      0 0 15px #00ff41,
      0 0 20px #00ff88;
  }
  20%, 24%, 55% {
    text-shadow: none;
  }
}

@keyframes cursor-blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

select {
  padding: 0.4rem;
  font-size: 1rem;
  background-color: #f0e6ff;
  color: #9966cc;
  border: 1px solid #b385ff;
  border-radius: 20px;
  box-shadow: 0 0 5px #d9b3ff;
  transition: all 0.3s ease;
}

.container.hacker select {
  background: linear-gradient(145deg, rgba(0, 0, 0, 0.9), rgba(0, 20, 0, 0.7));
  color: #00ff41;
  border: 1px solid #00ff41;
  border-radius: 4px;
  box-shadow: 
    0 0 10px rgba(0, 255, 65, 0.3),
    inset 0 1px 0 rgba(0, 255, 65, 0.1);
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
  letter-spacing: 1px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.container.hacker select:hover {
  box-shadow: 
    0 0 20px rgba(0, 255, 65, 0.5),
    inset 0 1px 0 rgba(0, 255, 65, 0.2);
  border-color: #00ff88;
}

.container.hacker select:focus {
  outline: none;
  box-shadow: 
    0 0 25px rgba(0, 255, 65, 0.7),
    inset 0 1px 0 rgba(0, 255, 65, 0.3);
}

#themeToggle {
  position: absolute;
  top: 1rem;
  right: 1rem;
  z-index: 3;
  padding: 0.3rem 0.8rem;
  border-radius: 20px;
  border: 1px solid #b385ff;
  background-color: #f0e6ff;
  color: #9966cc;
  cursor: pointer;
  box-shadow: 0 0 5px #d9b3ff;
  transition: all 0.3s ease;
}

.container.hacker #themeToggle {
  background: linear-gradient(145deg, rgba(0, 0, 0, 0.9), rgba(0, 30, 0, 0.8));
  color: #00ff41;
  border: 1px solid #00ff41;
  border-radius: 4px;
  box-shadow: 
    0 0 15px rgba(0, 255, 65, 0.4),
    inset 0 1px 0 rgba(0, 255, 65, 0.1);
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
  letter-spacing: 1px;
  font-weight: bold;
  overflow: hidden;
  transition: all 0.3s ease;
  position: absolute;
}

.container.hacker #themeToggle:hover {
  background: linear-gradient(145deg, rgba(0, 30, 0, 0.9), rgba(0, 50, 0, 0.8));
  color: #00ff88;
  border-color: #00ff88;
  box-shadow: 
    0 0 25px rgba(0, 255, 65, 0.6),
    inset 0 1px 0 rgba(0, 255, 65, 0.2);
  transform: translateY(-1px);
}

main {
  position: relative;
  z-index: 2;
  padding: 1rem;
  height: calc(100vh - 120px);
  overflow: auto;
  margin: 0 1rem 1rem 1rem;
  background-color: rgba(255, 255, 255, 0.8);
  border-radius: 20px;
  box-shadow: 0 0 15px #d9b3ff;
  color: #8a4fff;
  font-size: 0.9rem;
}

.container.hacker main {
  background: linear-gradient(145deg, rgba(0, 0, 0, 0.95), rgba(0, 15, 0, 0.9));
  color: #00ff41;
  border: 1px solid #00ff41;
  border-radius: 8px;
  box-shadow: 
    0 0 30px rgba(0, 255, 65, 0.3),
    inset 0 1px 0 rgba(0, 255, 65, 0.1),
    inset 0 -1px 0 rgba(0, 255, 65, 0.05);
  font-family: 'Courier New', monospace;
  position: relative;
  overflow: auto;
}

.container.hacker main::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, #00ff41, #00ff88, #00ff41, transparent);
  animation: border-glow 2s ease-in-out infinite alternate;
}

.container.hacker main::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0.3rem;
  background-image: 
    repeating-linear-gradient(
      0deg,
      transparent,
      transparent 2px,
      rgba(0, 255, 65, 0.02) 2px,
      rgba(0, 255, 65, 0.02) 4px
    );
  pointer-events: none;
  z-index: 0;
}

@keyframes border-glow {
  0% { opacity: 0.5; }
  100% { opacity: 1; }
}

#log {
  white-space: pre-wrap;
  word-break: break-word;
  overflow-y: auto;
  height: 100%;
  position: relative;
  z-index: 10;
}

.container.hacker #log {
  color: #00ff41;
  text-shadow: 0 0 2px rgba(0, 255, 65, 0.5);
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
  line-height: 1.4;
  animation: typing-glow 3s ease-in-out infinite alternate;
}

.container.hacker #log::-webkit-scrollbar {
  width: 8px;
}

.container.hacker #log::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.3);
  border-radius: 4px;
}

.container.hacker #log::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg, #00ff41, #00ff88);
  border-radius: 4px;
  box-shadow: 0 0 6px rgba(0, 255, 65, 0.5);
}

.container.hacker #log::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(180deg, #00ff88, #00ffaa);
  box-shadow: 0 0 10px rgba(0, 255, 65, 0.7);
}

@keyframes typing-glow {
  0% { text-shadow: 0 0 2px rgba(0, 255, 65, 0.3); }
  100% { text-shadow: 0 0 4px rgba(0, 255, 65, 0.7); }
}

.right-bg-swiper {
  position: fixed;
  top: 2%;
  bottom: 2%;
  right: 1rem;
  width: 550px;
  height: auto;
  z-index: 2;
  border-radius: 10px;
  overflow: hidden;
  /* box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1); */
}

.right-bg-swiper .swiper-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  z-index: 1;
}

.right-bg-swiper .swiper-slide {
  position: relative;
  width: 100%;
  height: 100%;
  flex-shrink: 0;
  opacity: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 10px;
  box-sizing: border-box;
}

.right-bg-swiper img {
  width: 100%;
  height: auto;
  max-height: 95vh;
  object-fit: contain;
  display: block;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.container.hacker .right-bg-swiper {
  background: rgba(0, 0, 0, 0.3);
  box-shadow: 0 4px 20px rgba(0, 255, 65, 0.1);
}

.container.hacker .right-bg-swiper img {
  background: transparent;
}

.homeBtn {
  position: absolute;
  top: 1rem;
  left: 1rem;
  z-index: 3;
  padding: 0.3rem 0.8rem;
  border-radius: 20px;
  border: 1px solid #b385ff;
  background-color: #f0e6ff;
  color: #9966cc;
  cursor: pointer;
  box-shadow: 0 0 5px #d9b3ff;
  transition: all 0.3s ease;
}

.container.hacker .homeBtn {
  background: linear-gradient(145deg, rgba(0, 0, 0, 0.9), rgba(0, 30, 0, 0.8));
  color: #00ff41;
  border: 1px solid #00ff41;
  border-radius: 4px;
  box-shadow: 
    0 0 15px rgba(0, 255, 65, 0.4),
    inset 0 1px 0 rgba(0, 255, 65, 0.1);
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
  letter-spacing: 1px;
  font-weight: bold;
  overflow: hidden;
  transition: all 0.3s ease;
  position: absolute;
}

.container.hacker .homeBtn::before {
  content: '> ';
  animation: cursor-blink 1.5s infinite;
}

.container.hacker .homeBtn:hover {
  background: linear-gradient(145deg, rgba(0, 30, 0, 0.9), rgba(0, 50, 0, 0.8));
  color: #00ff88;
  border-color: #00ff88;
  box-shadow: 
    0 0 25px rgba(0, 255, 65, 0.6),
    inset 0 1px 0 rgba(0, 255, 65, 0.2);
  transform: translateY(-1px) scale(1.02);
}

/* 黑客主题全局增强效果 */
.container.hacker * {
  transition: all 0.3s ease !important;
}

.container.hacker *:hover {
  text-shadow: 0 0 3px rgba(0, 255, 65, 0.5) !important;
}

/* 添加黑客风格的选中文本效果 */
.container.hacker ::selection {
  background: rgba(0, 255, 65, 0.3);
  color: #00ff88;
  text-shadow: 0 0 5px rgba(0, 255, 65, 0.8);
}

/* 黑客主题下的loading效果 */
.container.hacker .loading {
  color: #00ff41;
  animation: loading-dots 1.5s infinite;
}

@keyframes loading-dots {
  0%, 20% { content: '.'; }
  40% { content: '..'; }
  60%, 100% { content: '...'; }
}

/* 屏幕适配 */
@media (max-width: 768px) {
  .right-bg-swiper {
    display: none;
  }
  
  .container.hacker header {
    border-radius: 0;
  }
  
  .container.hacker main {
    margin: 0.5rem;
    border-radius: 4px;
  }
}

@media (min-width: 1200px) {
  .right-bg-swiper {
    width: 600px;
    right: 2rem;
    top: 2%;
    bottom: 2%;
  }
}

@media (min-width: 1600px) {
  .right-bg-swiper {
    width: 650px;
    right: 3rem;
    top: 2%;
    bottom: 2%;
  }
}
</style>
