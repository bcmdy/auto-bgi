<template>
  <div class="container" :class="{ hacker: isHackerTheme }">
    <!-- 背景动画canvas -->
    <canvas ref="animeStars" id="animeStars"></canvas>

    <!-- 页面头部 -->
    <header class="topbar">
      <button class="homeBtn" @click="goHome">返回首页</button>

      <div class="titleWrap">
        <h1 class="title">实时日志查看</h1>
        <p class="subtitle">LIVE · STREAM · LOG</p>
      </div>

      <div class="controls">
        <label class="srOnly" for="logSelector">选择日志文件</label>
        <select id="logSelector" v-model="selectedLog" @change="onLogChange">
          <option v-for="file in logFiles" :key="file" :value="file">
            {{ file }}
          </option>
        </select>

        <button id="themeToggle" @click="toggleTheme">
          {{ isHackerTheme ? '切换中二风' : '切换黑客风' }}
        </button>
      </div>
    </header>

    <!-- 主内容区域 -->
    <main class="main">
      <div class="logShell">
        <div id="log" ref="logContainer">{{ logContent }}</div>
      </div>
    </main>

    <!-- 轮播图背景（桌面端显示） -->
    <div class="right-bg-swiper" aria-hidden="true">
      <div class="swiper-wrapper" ref="swiperWrapper">
        <!-- 动态插入图片 -->
      </div>
      <!-- 装饰层 -->
      <div class="swiperFrame" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
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
    delay: 10000, // 10秒切换一次
    speed: 1000
  },
  STATIC_IMAGES: ['bd.jpg', 'ff.png', 'ng.jpg', 'sh.jpg'],
  IMG_CACHE_KEY: 'chuunibyou_bg_cache_v1',
  IMG_CACHE_TTL: 1000 * 60 * 60 * 24 * 7 // 7天
}

// WebSocket 相关
let ws = null
let canvas = null
let ctx = null
let width = 0
let height = 0
const stars = []
let mySwiper = null
let rafId = 0

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
        ? Math.random() > 0.3
          ? '#00ff41'
          : '#00ff88'
        : Math.random() > 0.5
          ? '#ff5fd7'
          : '#8b5cff',
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
      const chars = ['0', '1', '/', '+', '-', '*', '=', '<', '>', '[', ']', '{', '}', '#']
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
      let rot = (Math.PI / 2) * 3
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

  rafId = requestAnimationFrame(drawStars)
}

// WebSocket连接 - 优化版本（保持接口不变）
const connectWebSocket = name => {
  if (ws) {
    ws.close()
    ws = null
  }

  logContent.value = `正在连接 ${name} 日志服务...\n`

  try {
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
    const wsUrl = `${protocol}://${location.host}/ws/${encodeURIComponent(name)}`
    ws = new WebSocket(wsUrl)

    ws.onmessage = e => {
      if (!logContainer.value) return

      const atBottom =
        logContainer.value.scrollHeight - logContainer.value.scrollTop <=
        logContainer.value.clientHeight + 10
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

    ws.onclose = event => {
      logContent.value += `\n[连接已关闭 - 代码: ${event.code}]`
      console.log(`WebSocket 连接关闭: ${name}, 代码: ${event.code}`)
    }

    ws.onerror = error => {
      logContent.value += '\n[连接出错]'
      console.error(`WebSocket 连接错误: ${name}`, error)
    }
  } catch (error) {
    logContent.value += `\n[创建WebSocket连接失败: ${error.message}]`
    console.error('WebSocket 创建失败:', error)
  }
}

// 加载日志文件列表（保持接口不变）
const loadLogFiles = async () => {
  try {
    const token = localStorage.getItem('aBgiToken')
    console.log('获取到的 token:', token)

    const headers = {}
    if (token) {
      headers['Authorization'] = `${token}`
    }

    const res = await fetch('/api/logFiles', { headers })

    if (!res.ok) {
      if (res.status === 401) {
        logContent.value = '未授权 (401)，请先登录并确保 token 可用。'
      } else {
        let txt = ''
        try {
          txt = await res.text()
        } catch (e) {
          txt = ''
        }
        logContent.value = `加载日志列表失败：${res.status} ${res.statusText}\n${txt}`
      }
      return
    }

    const data = await res.json()
    if (data.files?.length) {
      logFiles.value = data.files
      selectedLog.value = data.files[0]
      connectWebSocket(data.files[0])
    } else {
      logContent.value = '未找到日志文件。'
    }
  } catch (err) {
    logContent.value = '加载日志列表失败。\n' + (err && err.message ? err.message : String(err))
  }
}

// 轻量缓存：记住“哪些静态图曾经加载成功”，减少无意义重试（不改后端）
const getCachedImages = () => {
  try {
    const raw = localStorage.getItem(CONSTANTS.IMG_CACHE_KEY)
    if (!raw) return null
    const data = JSON.parse(raw)
    if (!data || !Array.isArray(data.list) || !data.ts) return null
    if (Date.now() - data.ts > CONSTANTS.IMG_CACHE_TTL) return null
    return data.list
  } catch {
    return null
  }
}

const setCachedImages = list => {
  try {
    localStorage.setItem(
      CONSTANTS.IMG_CACHE_KEY,
      JSON.stringify({ ts: Date.now(), list })
    )
  } catch {
    // ignore
  }
}

// 预加载图片
const preloadImages = list => {
  return Promise.all(
    list.map(imgSrc => {
      return new Promise(resolve => {
        const img = new Image()
        img.onload = () => resolve(imgSrc)
        img.onerror = () => {
          console.warn(`图片预加载失败: ${imgSrc}`)
          resolve(null)
        }
        img.src = `/img/${imgSrc}`
      })
    })
  )
}

// 获取轮播图片 - 直接使用静态目录中的图片
const getImages = async () => {
  try {
    if (!swiperWrapper.value) {
      console.error('找不到轮播容器')
      return
    }

    // 移动端不需要轮播：直接不初始化（避免浪费性能）
    if (window.matchMedia && window.matchMedia('(max-width: 768px)').matches) {
      return
    }

    console.log('开始加载轮播图片...')

    // 优先用缓存（成功列表），没有再用静态常量
    const cached = getCachedImages()
    const baseList = cached && cached.length ? cached : CONSTANTS.STATIC_IMAGES

    const loadedImages = await preloadImages(baseList)
    const validImages = loadedImages.filter(img => img !== null)

    // 写回缓存（只缓存成功项）
    if (validImages.length) setCachedImages(validImages)

    console.log('有效图片数量:', validImages.length, validImages)

    if (validImages.length < 2) {
      console.warn('图片数量不足，无法轮播')
      return
    }

    swiperWrapper.value.innerHTML = ''

    const imagePromises = validImages.map((imgSrc, i) => {
      return new Promise(resolve => {
        const slide = document.createElement('div')
        slide.classList.add('swiper-slide')
        const img = document.createElement('img')
        img.src = `/img/${imgSrc}`
        img.alt = `轮播图${i + 1}`

        img.onload = () => {
          const aspectRatio = img.naturalWidth / img.naturalHeight
          img.style.width = 'auto'
          img.style.height = 'auto'
          img.style.objectFit = 'contain'

          if (aspectRatio > 1.2) {
            img.style.maxWidth = '100%'
            img.style.maxHeight = '90vh'
          } else if (aspectRatio < 0.8) {
            img.style.maxWidth = '100%'
            img.style.maxHeight = '94vh'
          } else {
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
    await nextTick()

    if (mySwiper) {
      mySwiper.destroy(true, true)
      mySwiper = null
    }

    mySwiper = new Swiper('.right-bg-swiper', {
      slidesPerView: 1,
      spaceBetween: 0,
      loop: true,
      autoplay: {
        delay: CONSTANTS.SWIPER_CONFIG.delay,
        disableOnInteraction: false,
        pauseOnMouseEnter: false
      },
      effect: 'fade',
      fadeEffect: { crossFade: true },
      speed: CONSTANTS.SWIPER_CONFIG.speed,
      allowTouchMove: false
    })

    setTimeout(() => {
      if (mySwiper && mySwiper.autoplay) {
        mySwiper.autoplay.start()
      }
    }, 100)
  } catch (err) {
    console.error('轮播图加载失败：', err)
  }
}

// 窗口大小变化处理
const handleResize = () => {
  setupStars()
  // 桌面/移动切换时：销毁/重建轮播（不改接口）
  if (window.matchMedia && window.matchMedia('(max-width: 768px)').matches) {
    if (mySwiper) {
      mySwiper.destroy(true, true)
      mySwiper = null
    }
  } else {
    // 重新拉起
    getImages()
  }
}

// 组件挂载
onMounted(() => {
  setupStars()
  drawStars()
  loadLogFiles()
  getImages()
  window.addEventListener('resize', handleResize)

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
  if (rafId) cancelAnimationFrame(rafId)
})
</script>

<style scoped>
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

/* 无障碍：屏幕阅读器专用 */
.srOnly {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  border: 0;
}

/* ======= 主题变量（中二二次元 · 霓虹魔法阵） ======= */
.container {
  --bg1: #120a2a;
  --bg2: #2a0f49;
  --glass: rgba(255, 255, 255, 0.08);
  --glass2: rgba(255, 255, 255, 0.12);
  --line: rgba(255, 255, 255, 0.16);

  --txt: rgba(255, 255, 255, 0.92);
  --muted: rgba(255, 255, 255, 0.62);

  --p1: #ff4fd8;
  --p2: #8b5cff;
  --p3: #46f5ff;
  --p4: #ffe66d;

  height: 100vh;
  width: 100vw;
  overflow: hidden;
  position: relative;
  color: var(--txt);
  font-family: "Mochiy Pop One", "Comic Sans MS", system-ui, -apple-system, "Segoe UI", sans-serif;

  background: radial-gradient(1200px 600px at 20% 10%, rgba(255, 79, 216, 0.18), transparent 60%),
    radial-gradient(900px 500px at 80% 20%, rgba(139, 92, 255, 0.18), transparent 60%),
    radial-gradient(900px 600px at 50% 90%, rgba(70, 245, 255, 0.10), transparent 60%),
    linear-gradient(145deg, var(--bg1), var(--bg2));
  transition: all 0.25s ease;
}

/* 魔法阵点阵 */
.container::before {
  content: "";
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 1;
  background-image:
    radial-gradient(circle at 20% 30%, rgba(255, 79, 216, 0.12) 1px, transparent 1px),
    radial-gradient(circle at 70% 60%, rgba(139, 92, 255, 0.12) 1px, transparent 1px),
    radial-gradient(circle at 40% 80%, rgba(70, 245, 255, 0.10) 1px, transparent 1px);
  background-size: 54px 54px, 72px 72px, 90px 90px;
  filter: blur(0.2px);
  opacity: 0.75;
}

/* ======= 黑客主题（暗黑终端 · 赛博绿） ======= */
.container.hacker {
  --bg1: #000000;
  --bg2: #041106;
  --glass: rgba(0, 255, 65, 0.06);
  --glass2: rgba(0, 255, 65, 0.10);
  --line: rgba(0, 255, 65, 0.22);

  --txt: #00ff41;
  --muted: rgba(0, 255, 136, 0.70);

  --p1: #00ff41;
  --p2: #00ff88;
  --p3: #00ffaa;
  --p4: #b8ffbf;

  font-family: "Courier New", Consolas, monospace;
  animation: scanlines 2s linear infinite;
}

@keyframes scanlines {
  0% { background-position: 0 0, 0 0, 0 0, 0 0; }
  100% { background-position: 0 100%, 0 0, 0 0, 0 0; }
}

/* 背景动画层 */
canvas#animeStars {
  position: fixed;
  inset: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
  background: transparent;
}

/* ======= 顶栏 ======= */
.topbar {
  position: relative;
  z-index: 5;
  padding: 14px 16px;
  margin: 10px 12px 0;
  border-radius: 18px;

  background: linear-gradient(135deg, rgba(255, 255, 255, 0.10), rgba(255, 255, 255, 0.06));
  border: 1px solid var(--line);
  box-shadow:
    0 10px 30px rgba(0, 0, 0, 0.25),
    0 0 0 1px rgba(255, 255, 255, 0.04) inset;
  backdrop-filter: blur(10px);
}

/* 顶栏霓虹边 */
.topbar::before {
  content: "";
  position: absolute;
  inset: -1px;
  border-radius: 18px;
  padding: 1px;
  background: linear-gradient(90deg, rgba(255,79,216,0.55), rgba(139,92,255,0.55), rgba(70,245,255,0.45));
  -webkit-mask: linear-gradient(#000 0 0) content-box, linear-gradient(#000 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  pointer-events: none;
  opacity: 0.85;
}

.container.hacker .topbar::before {
  background: linear-gradient(90deg, transparent, rgba(0,255,65,0.9), rgba(0,255,136,0.9), rgba(0,255,65,0.9), transparent);
  opacity: 0.95;
}

.topbar {
  display: grid;
  grid-template-columns: 140px 1fr auto;
  gap: 10px;
  align-items: center;
}

.titleWrap {
  text-align: center;
  min-width: 0;
}

.title {
  font-size: 1.25rem;
  letter-spacing: 1px;
  text-shadow:
    0 0 18px rgba(255, 79, 216, 0.20),
    0 0 18px rgba(139, 92, 255, 0.18);
}

.subtitle {
  margin-top: 2px;
  font-size: 0.72rem;
  color: var(--muted);
  letter-spacing: 2px;
  opacity: 0.9;
}

.controls {
  display: flex;
  align-items: center;
  gap: 10px;
  justify-content: flex-end;
}

/* ======= 按钮/选择框（统一中二风） ======= */
select,
button {
  -webkit-tap-highlight-color: transparent;
}

select {
  padding: 10px 12px;
  font-size: 0.95rem;
  color: var(--txt);
  background: var(--glass);
  border: 1px solid var(--line);
  border-radius: 14px;
  outline: none;
  box-shadow: 0 10px 22px rgba(0, 0, 0, 0.18);
  backdrop-filter: blur(10px);
  max-width: min(48vw, 360px);
}

.container.hacker select {
  border-radius: 8px;
  text-transform: uppercase;
  letter-spacing: 1px;
  box-shadow: 0 0 18px rgba(0, 255, 65, 0.25);
}

button {
  padding: 10px 12px;
  border-radius: 14px;
  border: 1px solid var(--line);
  cursor: pointer;
  color: var(--txt);
  background: linear-gradient(135deg, rgba(255, 79, 216, 0.16), rgba(139, 92, 255, 0.12));
  box-shadow:
    0 10px 22px rgba(0, 0, 0, 0.20),
    0 0 0 1px rgba(255, 255, 255, 0.04) inset;
  transition: transform 0.12s ease, box-shadow 0.12s ease, background 0.12s ease;
  backdrop-filter: blur(10px);
}

button:hover {
  transform: translateY(-1px);
  box-shadow:
    0 14px 28px rgba(0, 0, 0, 0.24),
    0 0 18px rgba(255, 79, 216, 0.18);
}

button:active {
  transform: translateY(0);
}

.container.hacker button {
  background: linear-gradient(145deg, rgba(0, 0, 0, 0.88), rgba(0, 40, 0, 0.55));
  border-radius: 8px;
  box-shadow: 0 0 18px rgba(0, 255, 65, 0.25);
  text-transform: uppercase;
  letter-spacing: 1px;
  font-weight: 700;
}

.container.hacker button:hover {
  box-shadow: 0 0 26px rgba(0, 255, 65, 0.42);
  color: var(--p2);
}

.homeBtn {
  justify-self: start;
}

#themeToggle {
  justify-self: end;
}

/* ======= 主体区域（避开右侧轮播） ======= */
.main {
  position: relative;
  z-index: 4;
  height: calc(100vh - 92px);
  padding: 12px;
}

/* 桌面端：为右侧轮播预留空间 */
@media (min-width: 769px) {
  .main {
    padding-right: clamp(12px, 2vw, 24px);
    margin-right: clamp(0px, 0vw, 0px);
  }
  /* 让日志区域不要被右侧轮播盖住 */
  .logShell {
    width: calc(100% - clamp(360px, 25vw, 940px) + 18px);
  }
}

/* 移动端：日志全宽 */
@media (max-width: 768px) {
  .logShell {
    width: 100%;
  }
}

.logShell {
  height: 100%;
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.10), rgba(255, 255, 255, 0.06));
  border: 1px solid var(--line);
  box-shadow:
    0 16px 40px rgba(0, 0, 0, 0.22),
    0 0 0 1px rgba(255, 255, 255, 0.04) inset;
  backdrop-filter: blur(12px);
  overflow: hidden;
  position: relative;
}

/* 角落“封印符文”装饰 */
.logShell::before {
  content: "";
  position: absolute;
  inset: -40px;
  background:
    radial-gradient(circle at 20% 30%, rgba(255,79,216,0.18), transparent 45%),
    radial-gradient(circle at 80% 20%, rgba(139,92,255,0.16), transparent 45%),
    radial-gradient(circle at 60% 85%, rgba(70,245,255,0.12), transparent 45%);
  filter: blur(2px);
  pointer-events: none;
}

.container.hacker .logShell::before {
  background:
    radial-gradient(circle at 20% 30%, rgba(0,255,65,0.14), transparent 45%),
    radial-gradient(circle at 80% 20%, rgba(0,255,136,0.12), transparent 45%);
}

/* 日志文本区 */
#log {
  position: relative;
  z-index: 2;
  height: 100%;
  padding: 14px 14px 18px;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-y: auto;
  font-size: 0.95rem;
  line-height: 1.55;
  color: var(--txt);
  text-shadow: 0 0 10px rgba(0, 0, 0, 0.18);
}

/* 更好看的滚动条 */
#log::-webkit-scrollbar {
  width: 10px;
}
#log::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.06);
  border-radius: 10px;
}
#log::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg, rgba(255,79,216,0.55), rgba(139,92,255,0.45));
  border-radius: 10px;
  box-shadow: 0 0 12px rgba(255,79,216,0.18);
}
#log::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(180deg, rgba(255,79,216,0.75), rgba(139,92,255,0.65));
}

.container.hacker #log {
  text-shadow: 0 0 4px rgba(0, 255, 65, 0.45);
  font-size: 0.92rem;
}
.container.hacker #log::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.35);
}
.container.hacker #log::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg, rgba(0,255,65,0.9), rgba(0,255,136,0.75));
  box-shadow: 0 0 12px rgba(0,255,65,0.35);
}

/* ======= 右侧轮播（桌面端） ======= */
.right-bg-swiper {
  position: fixed;
  top: 92px;
  bottom: 14px;
  right: 14px;
  width: clamp(320px, 30vw, 400px);
  z-index: 3;
  border-radius: 18px;
  overflow: hidden;

  background: rgba(0, 0, 0, 0.14);
  border: 1px solid var(--line);
  box-shadow:
    0 18px 50px rgba(0, 0, 0, 0.30),
    0 0 0 1px rgba(255, 255, 255, 0.03) inset;
  backdrop-filter: blur(10px);
}

/* 轮播框装饰（魔法封印边） */
.swiperFrame {
  position: absolute;
  inset: 10px;
  border-radius: 14px;
  pointer-events: none;
  border: 1px solid rgba(255, 255, 255, 0.14);
  box-shadow:
    0 0 0 1px rgba(255, 79, 216, 0.10) inset,
    0 0 0 1px rgba(139, 92, 255, 0.10);
  opacity: 0.9;
}

.container.hacker .swiperFrame {
  border-color: rgba(0, 255, 65, 0.20);
  box-shadow: 0 0 0 1px rgba(0, 255, 65, 0.18) inset;
}

.right-bg-swiper .swiper-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
}

.right-bg-swiper .swiper-slide {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 14px;
}

.right-bg-swiper img {
  width: 100%;
  height: auto;
  max-height: 95vh;
  object-fit: contain;
  display: block;
  border-radius: 14px;
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.28);
}

/* 移动端隐藏轮播 */
@media (max-width: 768px) {
  .right-bg-swiper {
    display: none;
  }

  .topbar {
    margin: 8px 8px 0;
    border-radius: 16px;
    grid-template-columns: 1fr;
    gap: 10px;
    text-align: left;
  }

  .titleWrap {
    text-align: left;
  }

  .controls {
    justify-content: space-between;
    gap: 8px;
    flex-wrap: wrap;
  }

  select {
    max-width: 100%;
    flex: 1 1 240px;
  }

  #themeToggle {
    flex: 0 0 auto;
  }

  .main {
    height: calc(100vh - 128px);
    padding: 10px 8px 12px;
  }

  .logShell {
    border-radius: 16px;
  }
}

/* 超宽屏微调 */
@media (min-width: 1600px) {
  .topbar {
    margin: 14px 18px 0;
  }
  .right-bg-swiper {
    right: 18px;
    bottom: 18px;
  }
}

/* 选中文本 */
::selection {
  background: rgba(255, 79, 216, 0.32);
  color: #fff;
}
.container.hacker ::selection {
  background: rgba(0, 255, 65, 0.28);
  color: rgba(0, 255, 136, 1);
}
</style>
