<template>
  <div class="app">
    <header class="header">
      <h1 @click="$router.go(-1)">屏幕共享----<span>返回</span></h1>

    
      <div class="status">
        <span :class="['status-indicator', connectionStatus]"></span>
        {{ statusText }}
      </div>
    </header>

    
    <main class="main">
      <div class="video-container">
        <img 
          ref="videoElement"
          :src="currentFrame" 
          alt="屏幕画面"
          class="video-frame"
        />
        <div v-if="!isConnected" class="loading">
          <div class="spinner"></div>
          <p>正在连接...</p>
        </div>
      </div>
      
      <div class="info">
        <div class="stats">
          <div class="stat">
            <span class="label">帧率:</span>
            <span class="value">{{ fps }} FPS</span>
          </div>
          <div class="stat">
            <span class="label">延迟:</span>
            <span class="value">N/A</span>
          </div>
          <div class="stat">
            <span class="label">分辨率:</span>
            <!-- <span class="value">{{ resolution }}</span> -->
            <span class="value">1920x1080</span>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script>
import  { apiMethods } from '@/utils/api'
export default {
  name: 'App',
  data() {
    return {
      ws: null,
      isConnected: false,
      currentFrame: '',
      fps: 0,
      resolution: '1280x720',
      frameCount: 0,
      lastFpsTime: Date.now(),

    }
  },
  computed: {
    connectionStatus() {
      return this.isConnected ? 'connected' : 'disconnected'
    },
    statusText() {
      return this.isConnected ? '已连接' : '未连接'
    }
  },
  mounted() {
    this.connectWebSocket()

  },
  beforeUnmount() {
    if (this.ws) this.ws.close()
    if (this.statusTimer) clearInterval(this.statusTimer)
  },
  methods: {
    connectWebSocket() {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/api/abgiScreen/ws`
      
      this.ws = new WebSocket(wsUrl)
      this.ws.binaryType = "blob" // 接收二进制流

      this.ws.onopen = () => {
        console.log('WebSocket连接已建立')
        this.isConnected = true
      }

      this.ws.onmessage = (event) => {
        const blob = event.data
        const url = URL.createObjectURL(blob)
        this.currentFrame = url
        setTimeout(() => URL.revokeObjectURL(url), 1000)

        // FPS统计
        const now = Date.now()
        this.frameCount++
        if (now - this.lastFpsTime >= 1000) {
          this.fps = Math.round((this.frameCount * 1000) / (now - this.lastFpsTime))
          this.frameCount = 0
          this.lastFpsTime = now
        }
      }

      this.ws.onclose = () => {
        console.log('WebSocket连接已关闭')
        this.isConnected = false
        setTimeout(() => this.connectWebSocket(), 3000)
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket错误:', error)
        this.isConnected = false
      }
    },
    // 获取状态信息
    async fetchStatus() {
      try {
        const response = await apiMethods.getStatus()
        this.statusData = response.data || {}
      } catch (error) {
        console.error('获取状态信息失败:', error)
      }
    }
  }
  
}
</script>

<style scoped>
/* 样式保持原来的，你不用改动 */
.app { min-height: 100vh; background: #1a1a1a; color: white; }
.header { display: flex; justify-content: space-between; align-items: center; padding: 1rem 2rem; background: #2a2a2a; border-bottom: 1px solid #333; }
.header h1 { margin: 0; font-size: 1.5rem; font-weight: 600; }
.status { display: flex; align-items: center; gap: 0.5rem; font-size: 0.9rem; }
.status-indicator { width: 8px; height: 8px; border-radius: 50%; background: #666; }
.status-indicator.connected { background: #4ade80; box-shadow: 0 0 8px rgba(74, 222, 128, 0.5); }
.status-indicator.disconnected { background: #ef4444; box-shadow: 0 0 8px rgba(239, 68, 68, 0.5); }
.main { padding: 2rem; display: flex; flex-direction: column; align-items: center; gap: 2rem; }
.video-container { position: relative; background: #000; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5); }
.video-frame { display: block; max-width: 100%; height: auto; image-rendering: -webkit-optimize-contrast; image-rendering: crisp-edges; }
.loading { position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); text-align: center; color: #999; }
.spinner { width: 40px; height: 40px; border: 3px solid #333; border-top: 3px solid #4ade80; border-radius: 50%; animation: spin 1s linear infinite; margin: 0 auto 1rem; }
@keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
.info { width: 100%; max-width: 800px; }
.stats { display: flex; justify-content: center; gap: 2rem; flex-wrap: wrap; }
.stat { display: flex; flex-direction: column; align-items: center; gap: 0.25rem; padding: 1rem; background: #2a2a2a; border-radius: 8px; min-width: 100px; }
.stat .label { font-size: 0.8rem; color: #999; }
.stat .value { font-size: 1.2rem; font-weight: 600; color: #4ade80; }
@media (max-width: 768px) { .header { padding: 1rem; } .main { padding: 1rem; } .stats { gap: 1rem; } .stat { min-width: 80px; padding: 0.75rem; } }
/* 状态卡片样式 */
.status-card {
  position: fixed;
  background: #2a2a2a;
  padding: 1rem 1.5rem;
  border-radius: 8px;
  min-width: 220px;
  width: 220px;
  color: white;
  flex-shrink: 0;
  margin-top: 100px;
}

.status-card h2 {
  font-size: 1rem;
  margin-bottom: 0.5rem;
}

.status-card pre {
  background: #1a1a1a;
  padding: 0.5rem;
  border-radius: 4px;
  overflow-x: auto;
  margin-bottom: 0.5rem;
}

.status-card p {
  margin: 0.25rem 0;
  font-size: 0.9rem;
}

.status-card .indexSX {
  cursor: pointer;
  color: #4ade80;
  font-weight: bold;
}
</style>
