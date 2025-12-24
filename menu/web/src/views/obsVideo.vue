<template>
  <div class="obs-container anime-theme">
    <div class="sky-bg" v-if="$mq !== 'mobile'">
      <div class="cloud cloud-1"></div>
      <div class="cloud cloud-2"></div>
      <div class="sakura sakura-1"></div>
      <div class="sakura sakura-2"></div>
      <div class="sakura sakura-3"></div>
      <div class="sparkle sparkle-1"></div>
    </div>

    <main class="main-layout">
      <aside class="control-sidebar">
        <div class="card control-card glassy">
          <div class="card-header">
            <span class="card-title">📹 控制台</span>
            <div class="connection-status" :class="{ 'online': isObsConnected, 'offline': !isObsConnected }">
              <span class="status-dot"></span>
              {{ isObsConnected ? 'OBS已连接' : 'OBS未连接' }}
            </div>
          </div>
          
          <div class="status-section">
            <div class="status-badge" :class="{ 'recording': isRecording, 'disabled': !isObsConnected }">
              <div class="status-icon">
                <div class="dot" :class="{ 'loading': loadingStatus.gettingStatus }"></div>
              </div>
              <div class="status-content">
                <div class="status-main">{{ !isObsConnected ? '💔 断开' : (isRecording ? '🔴 录制中' : '⭕ 待机') }}</div>
                <div class="status-sub">{{ loadingStatus.starting ? '请稍等...' : (isRecording ? '正在录制' : '点击开始') }}</div>
              </div>
            </div>

            <div class="status-badge replay-status" :class="{ 'recording': isReplayBufferActive, 'disabled': !isObsConnected }">
              <div class="status-icon">
                <div class="dot" :class="{ 'loading': loadingStatus.gettingReplayStatus }"></div>
              </div>
              <div class="status-content">
                <div class="status-main">{{ !isObsConnected ? '💔 断开' : (isReplayBufferActive ? '🟢 回放激活' : '⚪ 回放待机') }}</div>
                <div class="status-sub">{{ isReplayBufferActive ? '缓冲运行中' : '启动后可用' }}</div>
              </div>
            </div>
          </div>

          <div class="control-buttons">
            <div v-if="!isObsConnected" class="offline-alert">
              ⚠️ 无法连接到 OBS 服务
              <button class="btn small refresh-btn" @click="checkConnection">重试</button>
            </div>

            <div class="section-title">📹 录制操作</div>
            <div class="btn-row">
              <button 
                class="btn primary large" 
                @click="startRecording" 
                :disabled="!isObsConnected || isRecording || loadingStatus.starting"
              >
                <span class="btn-icon" v-if="!loadingStatus.starting">🎬</span>
                <MobileSpinner v-else />
                <span class="btn-text">{{ loadingStatus.starting ? '启动中' : '开始录制' }}</span>
              </button>
              
              <button 
                class="btn secondary large" 
                @click="stopRecording" 
                :disabled="!isObsConnected || !isRecording || loadingStatus.stopping"
              >
                <span class="btn-icon" v-if="!loadingStatus.stopping">⏹️</span>
                <MobileSpinner v-else />
                <span class="btn-text">{{ loadingStatus.stopping ? '停止中' : '停止录制' }}</span>
              </button>
            </div>

            <div class="section-title">🔄 回放缓冲操作</div>
            <div class="btn-row">
              <button 
                class="btn accent" 
                @click="startReplayBuffer" 
                :disabled="!isObsConnected || isReplayBufferActive || loadingStatus.startingReplay"
              >
                <span class="btn-icon">▶️</span>
                <span class="btn-text">{{ loadingStatus.startingReplay ? '启动中' : '启动回放' }}</span>
              </button>
              
              <button 
                class="btn secondary" 
                @click="stopReplayBuffer" 
                :disabled="!isObsConnected || !isReplayBufferActive || loadingStatus.stoppingReplay"
              >
                <span class="btn-icon">⏸️</span>
                <span class="btn-text">{{ loadingStatus.stoppingReplay ? '停止中' : '停止回放' }}</span>
              </button>
            </div>

            <button 
              class="btn save-btn" 
              @click="saveReplayBuffer" 
              :disabled="!isObsConnected || !isReplayBufferActive || loadingStatus.savingReplay"
            >
              <MobileSpinner v-if="loadingStatus.savingReplay" />
              <span class="btn-icon" v-else>💾</span>
              <span class="btn-text">{{ loadingStatus.savingReplay ? '保存中...' : '保存回放片段' }}</span>
            </button>
            
            <div class="divider"></div>
            
            <div class="section-title">📂 系统</div>
            <div class="btn-row">
              <button 
                class="btn ghost" 
                @click="fetchVideos" 
                :disabled="loadingStatus.fetchingVideos"
              >
                <MobileSpinner v-if="loadingStatus.fetchingVideos" />
                <span class="btn-icon" v-else>🔄</span>
                <span class="btn-text">{{ loadingStatus.fetchingVideos ? '加载中' : '刷新列表' }}</span>
              </button>

              <button class="btn ghost" @click="comeBack">
                <span class="btn-icon">🏠</span>
                <span class="btn-text">返回首页</span>
              </button>
            </div>
          </div>
        </div>
      </aside>

      <section class="player-section" :class="{'active': currentVideo}">
        <div class="card player-card glassy">
          <transition name="fade" mode="out-in">
            <div v-if="currentVideo" class="player-container" key="player">
              <div class="player-header">
                <div class="player-title">
                  <span class="icon">▶️</span>
                  <span class="text">{{ currentVideoName }}</span>
                </div>
                <button class="btn small close-btn" @click="closePlayer">✕</button>
              </div>
              
              <div class="video-wrapper">
                <video 
                  ref="videoRef"
                  controls 
                  autoplay
                  class="main-video"
                  :src="getVideoStreamUrl(currentVideo)"
                  @ratechange="handleRateChange"
                >
                  您的浏览器不支持视频播放
                </video>

                <div class="speed-control-panel">
                  <div class="speed-label">🚀 播放倍速</div>
                  <div class="speed-grid">
                    <button class="speed-btn" :class="{ active: currentPlaybackRate === 1.0 }" @click="setPlaybackRate(1.0)">1.0x</button>
                    <button class="speed-btn" :class="{ active: currentPlaybackRate === 2.0 }" @click="setPlaybackRate(2.0)">2.0x</button>
                    <button class="speed-btn high-speed" :class="{ active: currentPlaybackRate === 5.0 }" @click="setPlaybackRate(5.0)">5.0x</button>
                    <button class="speed-btn high-speed" :class="{ active: currentPlaybackRate === 8.0 }" @click="setPlaybackRate(8.0)">8.0x</button>
                  </div>
                </div>

              </div>
            </div>

            <div v-else class="empty-player" key="empty">
              <div class="empty-content">
                <div class="empty-icon">🎞️</div>
                <h3>选择视频播放</h3>
                <p>点击右侧列表观看回放 ~</p>
                <div class="cute-decoration">｡◕‿◕｡</div>
              </div>
            </div>
          </transition>
        </div>
      </section>

      <aside class="video-sidebar">
        <div class="card list-card glassy">
          <div class="card-header">
            <div class="header-left">
              <span class="card-title">📂 文件列表</span>
              <span class="badge" v-if="videos.length">{{ videos.length }}</span>
            </div>
            <button class="btn small refresh-icon" @click="fetchVideos" :disabled="loadingStatus.fetchingVideos">
              <span :class="{'spin': loadingStatus.fetchingVideos}">🔄</span>
            </button>
          </div>

          <div class="video-list custom-scroll">
            <div v-if="loadingStatus.fetchingVideos" class="mobile-loading">
               <div class="list-spinner"></div>
               <p class="loading-text">加载文件中...</p>
            </div>

            <transition-group name="list" tag="div" v-else-if="videos.length > 0">
              <div 
                v-for="video in videos" 
                :key="video.name" 
                class="video-item"
                :class="{ 'active': currentVideo === video.name }"
                @click="playVideo(video.name)"
              >
                <div class="video-thumbnail">
                  <span class="thumb-icon">🎬</span>
                </div>
                <div class="video-info">
                  <div class="video-title">{{ video.name }}</div>
                  <div class="video-meta">
                    <span class="meta-tag size">{{ video.sizeMB.toFixed(1) }}MB</span>
                    <span class="meta-tag time">⏱ {{ formatTime(video.sizeMB) }}</span>
                  </div>
                </div>
                <div class="video-actions">
                  <button 
                    class="btn small delete-btn" 
                    @click.stop="DeleteVideo(video.name)"
                    :disabled="loadingStatus.deletingVideo && deletingVideoName === video.name"
                  >
                    🗑️
                  </button>
                </div>
              </div>
            </transition-group>

            <div v-else class="empty-list">
              <div class="empty-icon">📭</div>
              <h3>暂无文件</h3>
            </div>
          </div>
        </div>
      </aside>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive, computed } from 'vue'
import { apiMethods } from '@/utils/api'

function getToken() {
  return localStorage.getItem('aBgiToken') || ''
}

function getVideoStreamUrl(path) {
  const token = getToken()
  const url = `/api/abgiObs/PlayVideoStream?path=${encodeURIComponent(path)}${token ? `&tk=${encodeURIComponent(token)}` : ''}`
  return url
}

// 简单的响应式判断
const windowWidth = ref(window.innerWidth)
const $mq = computed(() => windowWidth.value > 768 ? 'desktop' : 'mobile')
window.addEventListener('resize', () => { windowWidth.value = window.innerWidth })

// --- 状态变量 ---
const isObsConnected = ref(false) // 新增：连接状态
const isRecording = ref(false)
const isReplayBufferActive = ref(false)
const videos = ref([])
const currentVideo = ref('')
const currentVideoName = ref('')
const videoRef = ref(null)
const deletingVideoName = ref('')
const currentPlaybackRate = ref(1.0)

const loadingStatus = reactive({
  gettingStatus: false,
  gettingReplayStatus: false,
  starting: false,
  stopping: false,
  startingReplay: false,
  stoppingReplay: false,
  savingReplay: false,
  fetchingVideos: false,
  loadingVideo: false,
  deletingVideo: false
})

const MobileSpinner = {
  template: `<div class="mobile-spinner"></div>`
}

// --- 倍速控制 ---
function setPlaybackRate(rate) {
  if (videoRef.value) {
    videoRef.value.playbackRate = rate
    currentPlaybackRate.value = rate
  }
}

function handleRateChange(e) {
  if(e.target) {
    currentPlaybackRate.value = e.target.playbackRate
  }
}

function closePlayer() {
  if(videoRef.value) videoRef.value.pause();
  currentVideo.value = '';
  currentVideoName.value = '';
}

// --- API 方法 (包含连接检测逻辑) ---

// 用于手动点击重试连接
function checkConnection() {
  getRecordingStatus()
  getReplayBufferStatus()
  fetchVideos()
}

async function getRecordingStatus() {
  loadingStatus.gettingStatus = true
  try {
    const res = await apiMethods.IsRecording()
    // 如果 API 成功返回，说明连接正常
    isObsConnected.value = true 
    isRecording.value = res.msg?.outputActive === true
  } catch (err) {
    console.error("OBS Connection Error:", err)
    // 如果报错，判定为断开连接
    isObsConnected.value = false 
    isRecording.value = false
  } finally {
    loadingStatus.gettingStatus = false
  }
}

async function getReplayBufferStatus() {
  loadingStatus.gettingReplayStatus = true
  try {
    const res = await apiMethods.GetReplayBufferStatus()
    isObsConnected.value = true // 更新连接状态
    isReplayBufferActive.value = res.msg?.outputActive === true
  } catch (err) {
    console.error(err)
    isObsConnected.value = false
  } finally {
    loadingStatus.gettingReplayStatus = false
  }
}

async function startRecording() {
  if (!isObsConnected.value) return; // 双重保险
  loadingStatus.starting = true
  try {
    const res = await apiMethods.StartRecording()
    if (res.status === 'success') {
      isRecording.value = true
      setTimeout(() => getRecordingStatus(), 500)
      fetchVideos()
    } else {
      alert('❌ ' + res.msg)
    }
  } catch (err) {
    console.error(err)
    alert('请求失败，请检查网络')
  } finally {
    loadingStatus.starting = false
  }
}

async function stopRecording() {
  if (!isObsConnected.value) return;
  loadingStatus.stopping = true
  try {
    const res = await apiMethods.StopRecording()
    if (res.status === 'success') {
      isRecording.value = false
      setTimeout(() => getRecordingStatus(), 1000)
      fetchVideos()
    }
  } catch (err) {
    console.error(err)
  } finally {
    loadingStatus.stopping = false
  }
}

async function startReplayBuffer() {
  if (!isObsConnected.value) return;
  loadingStatus.startingReplay = true
  try {
    const res = await apiMethods.StartReplayBuffer()
    if (res.status === 'success') {
      isReplayBufferActive.value = true
      setTimeout(() => getReplayBufferStatus(), 500)
    } else {
      alert('❌ ' + res.msg)
    }
  } catch (err) {
    console.error(err)
  } finally {
    loadingStatus.startingReplay = false
  }
}

async function stopReplayBuffer() {
  if (!isObsConnected.value) return;
  loadingStatus.stoppingReplay = true
  try {
    const res = await apiMethods.StopReplayBuffer()
    if (res.status === 'success') {
      isReplayBufferActive.value = false
      setTimeout(() => getReplayBufferStatus(), 500)
    } else {
      alert('❌ ' + res.msg)
    }
  } catch (err) {
    console.error(err)
  } finally {
    loadingStatus.stoppingReplay = false
  }
}

async function saveReplayBuffer() {
  if (!isObsConnected.value) return;
  loadingStatus.savingReplay = true
  try {
    const res = await apiMethods.SaveReplayBuffer()
    if (res.status === 'success') {
      alert('✨ 回放已保存！')
      fetchVideos()
    } else {
      alert('❌ ' + res.msg)
    }
  } catch (err) {
    console.error(err)
  } finally {
    loadingStatus.savingReplay = false
  }
}

async function fetchVideos() {
  loadingStatus.fetchingVideos = true
  try {
    const res = await apiMethods.GetVideoInfo()
    if (res.status === 'success') {
      videos.value = res.msg || [] 
    }
  } catch (err) {
    console.error(err)
  } finally {
    loadingStatus.fetchingVideos = false
  }
}

async function comeBack() {
  window.location.href = '/'
}

async function playVideo(name) {
  loadingStatus.loadingVideo = true
  try {
    currentVideo.value = name
    const videoItem = videos.value.find(v => v.name === name)
    currentVideoName.value = videoItem ? videoItem.name : ''
    
    // 手机端滚动
    if ($mq.value === 'mobile') {
      setTimeout(() => {
        const player = document.querySelector('.player-section')
        if(player) player.scrollIntoView({ behavior: 'smooth' })
      }, 100)
    }

    setPlaybackRate(1.0) // 默认1倍速

    if (videoRef.value) {
      videoRef.value.load()
      try {
        await videoRef.value.play()
      } catch(e) { console.log('Autoplay blocked', e) }
    }
  } catch (err) {
    console.error(err)
  } finally {
    loadingStatus.loadingVideo = false
  }
}

async function DeleteVideo(name) {
  if (!confirm(`确认要删除 "${name}" 吗？(>_<)`)) return
  deletingVideoName.value = name
  loadingStatus.deletingVideo = true
  try {
    const res = await apiMethods.DeleteVideo(name)
    if (res.status === 'success') {
      fetchVideos()
      if (currentVideo.value === name) closePlayer()
    }
  } catch (err) {
    console.error(err)
  } finally {
    loadingStatus.deletingVideo = false
    deletingVideoName.value = ''
  }
}

function formatTime(sizeMB) {
  const seconds = Math.round(sizeMB * 60)
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

onMounted(() => {
  // 初始化时检测连接
  getRecordingStatus()
  getReplayBufferStatus()
  fetchVideos()
})
</script>

<style scoped>
@import '../assets/css2.css';

:root {
  --pink-light: #fff0f5;
  --pink-bg: #ffe4e1;
  --pink-primary: #ffb7c5;
  --pink-deep: #ff69b4;
  --purple-accent: #b19cd9;
  --text-main: #5d4d68;
  --text-light: #8b7d96;
  --glass: rgba(255, 255, 255, 0.85);
  --shadow-soft: 0 8px 32px 0 rgba(255, 183, 197, 0.3);
  --radius-lg: 24px;
  --radius-md: 16px;
}

/* 全局容器 */
.obs-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #fff0f5 0%, #e6e6fa 50%, #e0ffff 100%);
  font-family: 'M PLUS Rounded 1c', 'Microsoft YaHei', sans-serif;
  color: var(--text-main);
  position: relative;
  overflow-x: hidden;
}

/* === 背景动画 === */
.sky-bg { position: fixed; inset: 0; z-index: 0; pointer-events: none; }
.cloud { position: absolute; background: #fff; border-radius: 100px; opacity: 0.6; filter: blur(5px); }
.cloud-1 { width: 180px; height: 60px; top: 10%; left: -200px; animation: float 35s linear infinite; }
.cloud-2 { width: 120px; height: 40px; top: 25%; left: -150px; animation: float 28s linear infinite 5s; }
.sakura { position: absolute; background: #ffb7c5; border-radius: 100% 0 100% 0; animation: fall linear infinite; }
.sakura-1 { width: 12px; height: 12px; left: 20%; animation-duration: 10s; }
.sakura-2 { width: 8px; height: 8px; left: 70%; animation-duration: 14s; animation-delay: 2s; }
.sakura-3 { width: 15px; height: 15px; left: 40%; animation-duration: 12s; animation-delay: 5s; }

/* === 主布局 === */
.main-layout {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: 300px 1fr 320px;
  gap: 24px;
  max-width: 1600px;
  margin: 0 auto;
  padding: 24px;
  height: 100vh;
  box-sizing: border-box;
}

/* === 卡片 === */
.card {
  background: var(--glass);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 2px solid #fff;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-soft);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.card-header {
  padding: 16px 20px;
  border-bottom: 2px dashed var(--pink-bg);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(255,255,255,0.4);
}
.card-title { font-weight: 700; font-size: 1.1rem; color: var(--pink-deep); }

/* === 左侧：状态与连接 === */
.connection-status {
  font-size: 0.75rem;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 12px;
  font-weight: bold;
}
.connection-status.online { background: #d1fae5; color: #047857; }
.connection-status.offline { background: #fee2e2; color: #b91c1c; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; background: currentColor; }

.status-section { padding: 20px; display: flex; flex-direction: column; gap: 12px; }
.status-badge {
  display: flex; align-items: center; gap: 15px; padding: 12px 16px;
  background: #fff; border-radius: var(--radius-md); border: 2px solid var(--pink-bg);
  transition: all 0.3s;
}
/* 禁用状态的 Badge */
.status-badge.disabled { opacity: 0.6; background: #f3f4f6; border-color: #e5e7eb; filter: grayscale(1); }

.status-badge.recording { background: #fff0f5; border-color: var(--pink-deep); transform: scale(1.02); }
.status-icon .dot { width: 14px; height: 14px; background: #ddd; border-radius: 50%; }
.status-badge.recording .dot { background: #ff4757; animation: pulse 1.5s infinite; }

/* === 按钮区域 === */
.control-buttons { padding: 0 20px 20px; display: flex; flex-direction: column; gap: 10px; }
.offline-alert {
  background: #fee2e2; color: #991b1b; padding: 10px; border-radius: 12px;
  font-size: 0.85rem; display: flex; align-items: center; justify-content: space-between;
}
.section-title { font-size: 0.85rem; color: var(--pink-deep); font-weight: bold; margin-top: 10px; }

.btn-row { display: flex; gap: 10px; }
.btn-row .btn { flex: 1; }

.btn {
  border: none; cursor: pointer; border-radius: 50px;
  display: flex; align-items: center; justify-content: center;
  gap: 6px; font-family: inherit; font-weight: 700;
  transition: all 0.2s;
}
/* 禁用状态 */
.btn:disabled { opacity: 0.5; cursor: not-allowed; background: #e5e7eb !important; color: #9ca3af !important; box-shadow: none !important; }

.btn.large { padding: 12px 20px; font-size: 1rem; }
.btn { padding: 10px 16px; font-size: 0.9rem; }
.btn.small { padding: 6px; min-width: 32px; min-height: 32px; border-radius: 50%; }

.btn.primary { background: linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%); color: #fff; box-shadow: 0 4px 15px rgba(255, 154, 158, 0.4); }
.btn.secondary { background: #f1f2f6; color: var(--text-light); }
.btn.accent { background: linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%); color: white; }
.btn.save-btn { background: var(--pink-primary); color: white; width: 100%; }
.btn.ghost { background: transparent; border: 1px dashed var(--pink-primary); color: var(--text-main); }

/* === 播放器区域 === */
.player-section { display: flex; flex-direction: column; height: 100%; }
.player-card { flex: 1; background: rgba(255, 255, 255, 0.95); }
.player-header { border-bottom: 2px solid var(--pink-light); }
.player-title { font-size: 1.2rem; color: var(--pink-deep); display: flex; align-items: center; gap: 8px; }

.video-wrapper { padding: 20px; display: flex; flex-direction: column; gap: 16px; align-items: center; justify-content: center; flex: 1; }
.main-video { width: 100%; max-height: 55vh; border-radius: var(--radius-md); background: #000; box-shadow: 0 10px 30px rgba(0,0,0,0.1); }

/* === 新版倍速控制面板 === */
.speed-control-panel {
  width: 100%;
  background: #fff;
  border-radius: 20px;
  padding: 12px 20px;
  border: 2px solid var(--pink-light);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.speed-label { font-weight: bold; color: var(--pink-deep); font-size: 0.9rem; white-space: nowrap; }
.speed-grid { display: flex; gap: 8px; flex-wrap: wrap; justify-content: flex-end; flex: 1; }

.speed-btn {
  border: 1px solid #eee;
  background: rgb(235, 240, 238);
  color: #666;
  padding: 6px 14px;
  border-radius: 12px;
  font-size: 0.9rem;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.2s;
}
.speed-btn:hover { background: var(--pink-light); color: var(--pink-deep); }
.speed-btn.active {
  background: var(--pink-deep);
  color: rgb(129, 235, 203);
  border-color: var(--pink-deep);
  transform: scale(1.05);
  box-shadow: 0 4px 10px rgba(255, 105, 180, 0.3);
}
.speed-btn.high-speed {
  border-color: var(--purple-accent);
  color: var(--purple-accent);
}
.speed-btn.high-speed.active {
  background: var(--purple-accent);
  color: rgb(189, 221, 115);
}

.empty-player { height: 100%; display: flex; align-items: center; justify-content: center; text-align: center; color: var(--pink-primary); }
.empty-icon { font-size: 4rem; margin-bottom: 20px; opacity: 0.5; }

/* === 视频列表 === */
.video-sidebar { height: 100%; overflow: hidden; }
.list-card { height: 100%; }
.video-list { flex: 1; overflow-y: auto; padding: 10px; }
.video-item {
  display: flex; align-items: center; gap: 12px; padding: 12px; margin-bottom: 8px;
  background: rgba(255,255,255,0.6); border-radius: 12px; cursor: pointer; transition: all 0.2s;
  border: 1px solid transparent;
}
.video-item:hover { background: #fff; transform: translateX(4px); border-color: var(--pink-primary); }
.video-item.active { background: var(--pink-light); border-color: var(--pink-deep); }
.video-thumbnail { width: 50px; height: 50px; background: var(--pink-bg); border-radius: 10px; display: flex; align-items: center; justify-content: center; }
.video-info { flex: 1; min-width: 0; }
.video-title { font-weight: bold; font-size: 0.9rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin-bottom: 4px; }
.video-meta { display: flex; gap: 8px; font-size: 0.75rem; color: var(--text-light); }
.meta-tag { background: rgba(255,255,255,0.8); padding: 2px 6px; border-radius: 4px; }
.delete-btn { background: #ffebee; color: #ff5252; }

/* === 动画 === */
@keyframes float { 0% { transform: translateX(0); } 100% { transform: translateX(100vw); } }
@keyframes fall { 0% { top: -10%; transform: translateX(0) rotate(0deg); } 100% { top: 110%; transform: translateX(20px) rotate(360deg); } }
@keyframes spin { 100% { transform: rotate(360deg); } }
@keyframes pulse { 0%, 100% { transform: scale(1); opacity: 1; } 50% { transform: scale(1.2); opacity: 0.8; } }

/* === 手机端响应式 === */
@media (max-width: 768px) {
  .sky-bg { display: none; }
  .main-layout { display: flex; flex-direction: column; padding: 12px; height: auto; gap: 16px; }
  .player-section { order: 1; margin-bottom: 10px; }
  .control-sidebar { order: 2; }
  .video-sidebar { order: 3; }
  .player-section:not(.active) { display: none; }
  .player-section.active { display: block; height: auto; }
  .card { border-radius: 16px; }
  .video-list { max-height: 400px; }
  
  /* 手机端倍速面板优化 */
  .speed-control-panel { flex-direction: column; align-items: flex-start; gap: 8px; }
  .speed-grid { width: 100%; justify-content: space-between; }
  .speed-btn { flex: 1; text-align: center; padding: 8px 0; }
  
  .mobile-spinner { width: 18px; height: 18px; border: 2px solid transparent; border-top: 2px solid currentColor; border-radius: 50%; animation: spin 0.8s linear infinite; }
  .list-spinner { width: 30px; height: 30px; border: 3px solid var(--pink-primary); border-top-color: transparent; border-radius: 50%; margin: 20px auto; animation: spin 0.8s linear infinite; }
}
</style>