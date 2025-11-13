<template>
  <div class="obs-container">
    <!-- 背景动画 - 手机端关闭 -->
    <div class="sky-bg" v-if="$mq === 'desktop'">
      <div class="cloud cloud-1"></div>
      <div class="cloud cloud-2"></div>
      <div class="sakura sakura-1"></div>
      <div class="sakura sakura-2"></div>
      <div class="sparkle sparkle-1"></div>
      <div class="sparkle sparkle-2"></div>
    </div>

    <!-- 主布局：三列/单列响应式 -->
    <main class="main-layout">
      <!-- 左侧录制控制 -->
      <aside class="control-sidebar">
        <div class="card control-card">
          <div class="card-header">
            <h2 class="card-title">📹 录制控制</h2>
          </div>
          
          <div class="status-section">
            <div class="status-badge" :class="{ recording: isRecording }">
              <div class="status-icon">
                <span class="dot" :class="{ pulse: isRecording, loading: loadingStatus.gettingStatus }"></span>
              </div>
              <div class="status-content">
                <div class="status-main">
                  <span v-if="loadingStatus.gettingStatus">⏳ 连接中</span>
                  <span v-else>{{ isRecording ? '🔴 录制中' : '⭕ 待机' }}</span>
                </div>
                <div class="status-sub">
                  <span v-if="loadingStatus.gettingStatus">请稍等</span>
                  <span v-else-if="isRecording">正在录制</span>
                  <span v-else>点击开始</span>
                </div>
              </div>
            </div>
            
            <div class="status-badge replay-status" :class="{ recording: isReplayBufferActive }">
              <div class="status-icon">
                <span class="dot" :class="{ pulse: isReplayBufferActive, loading: loadingStatus.gettingReplayStatus }"></span>
              </div>
              <div class="status-content">
                <div class="status-main">
                  <span v-if="loadingStatus.gettingReplayStatus">⏳ 连接中</span>
                  <span v-else>{{ isReplayBufferActive ? '🟢 回放激活' : '⚪ 回放待机' }}</span>
                </div>
                <div class="status-sub">
                  <span v-if="loadingStatus.gettingReplayStatus">请稍等</span>
                  <span v-else-if="isReplayBufferActive">缓冲运行中</span>
                  <span v-else>启动后可用</span>
                </div>
              </div>
            </div>
          </div>

          <div class="control-buttons">
            <div class="section-title">📹 录制</div>
            <div class="btn-row">
              <button 
                class="btn primary" 
                @click="startRecording" 
                :disabled="isRecording || loadingStatus.starting"
              >
                <span class="btn-icon">
                  <MobileSpinner v-if="loadingStatus.starting" />
                  <span v-else>🎬</span>
                </span>
                <span class="btn-text">{{ loadingStatus.starting ? '启动中' : '开始' }}</span>
              </button>
              <button 
                class="btn secondary" 
                @click="stopRecording" 
                :disabled="!isRecording || loadingStatus.stopping"
              >
                <span class="btn-icon">
                  <MobileSpinner v-if="loadingStatus.stopping" />
                  <span v-else>⏹️</span>
                </span>
                <span class="btn-text">{{ loadingStatus.stopping ? '停止中' : '停止' }}</span>
              </button>
            </div>
            
            <div class="section-title">🔄 回放缓冲</div>
            <div class="btn-row">
              <button 
                class="btn primary" 
                @click="startReplayBuffer" 
                :disabled="isReplayBufferActive || loadingStatus.startingReplay"
              >
                <span class="btn-icon">
                  <MobileSpinner v-if="loadingStatus.startingReplay" />
                  <span v-else>▶️</span>
                </span>
                <span class="btn-text">{{ loadingStatus.startingReplay ? '启动中' : '启动' }}</span>
              </button>
              <button 
                class="btn secondary" 
                @click="stopReplayBuffer" 
                :disabled="!isReplayBufferActive || loadingStatus.stoppingReplay"
              >
                <span class="btn-icon">
                  <MobileSpinner v-if="loadingStatus.stoppingReplay" />
                  <span v-else>⏸️</span>
                </span>
                <span class="btn-text">{{ loadingStatus.stoppingReplay ? '停止中' : '停止' }}</span>
              </button>
            </div>
            <button 
              class="btn accent" 
              @click="saveReplayBuffer" 
              :disabled="!isReplayBufferActive || loadingStatus.savingReplay"
            >
              <span class="btn-icon">
                <MobileSpinner v-if="loadingStatus.savingReplay" />
                <span v-else>💾</span>
              </span>
              <span class="btn-text">{{ loadingStatus.savingReplay ? '保存中' : '保存回放' }}</span>
            </button>
            
            <div class="section-title">📂 文件</div>
            <button 
              class="btn ghost" 
              @click="fetchVideos" 
              :disabled="loadingStatus.fetchingVideos"
            >
              <span class="btn-icon">
                <MobileSpinner v-if="loadingStatus.fetchingVideos" />
                <span v-else>🔄</span>
              </span>
              <span class="btn-text">{{ loadingStatus.fetchingVideos ? '加载中' : '刷新列表' }}</span>
            </button>

                 <button 
              class="btn ghost" 
              @click="comeBack" 
            >
              <span class="btn-text">返回首页</span>
            </button>
          </div>
        </div>
      </aside>

      <!-- 中间播放器区域 -->
  <section class="player-section">
        <div v-if="currentVideo" class="player-container">
          <transition name="fade">
            <div class="card player-card">
              <div class="player-header">
                <h2 class="player-title">
                  <span class="icon">▶️</span>
                  {{ currentVideoName }}
                </h2>
                <button 
                  class="btn ghost small close-btn" 
                  @click="currentVideo = ''" 
                  :disabled="loadingStatus.loadingVideo"
                  title="关闭"
                >
                  <MobileSpinner v-if="loadingStatus.loadingVideo" />
                  <span v-else>✕</span>
                </button>
              </div>
              
              <div class="video-wrapper">
                <video
                  ref="videoRef"
                  class="main-video"
                  :src="`/api/abgiObs/PlayVideoStream?path=${encodeURIComponent(currentVideo)}`"
                  controls
                  autoplay
                  playsinline
                ></video>
              </div>
            </div>
          </transition>
        </div>
        
        <div v-else class="empty-player">
          <div class="empty-content">
            <div class="empty-icon">🎞️</div>
            <h3>选择视频播放</h3>
            <p class="mobile-hide">点击下方视频～</p>
            <p class="mobile-show">在下方列表选择视频</p>
          </div>
        </div>
      </section>

      <!-- 右侧视频列表 -->
  <aside class="video-sidebar">
        <div class="card list-card">
          <div class="card-header">
            <h2 class="card-title">
              <span v-if="loadingStatus.fetchingVideos">⏳</span>
              <span v-else>📂</span>
              {{ loadingStatus.fetchingVideos ? '加载中' : `文件 (${videos.length})` }}
            </h2>
            <button 
              class="btn ghost small mobile-close" 
              @click="fetchVideos" 
              :disabled="loadingStatus.fetchingVideos"
            >
              <MobileSpinner v-if="loadingStatus.fetchingVideos" />
              <span v-else>🔄</span>
            </button>
          </div>
          
          <div class="video-list">
            <transition-group name="list" tag="div">
              <div 
                v-for="video in videos" 
                :key="video.path" 
                class="video-item"
                @click="playVideo(video.name)"
              >
                <div class="video-thumbnail">
                  <div class="thumbnail-placeholder">
                    <span class="thumb-icon">🎬</span>
                  </div>
                </div>
                
                <div class="video-info">
                  <div class="video-title">{{ video.name }}</div>
                  <div class="video-meta mobile-hide">
                    <span class="size">{{ video.sizeMB.toFixed(1) }}MB</span>
                    <span class="duration">• {{ formatTime(video.sizeMB) }}</span>
                  </div>
                </div>
                
                <div class="video-actions">
                  <button 
                    class="btn primary small" 
                    @click.stop="playVideo(video.name)"
                    :disabled="loadingStatus.loadingVideo"
                    :title="'播放'"
                  >
                    <MobileSpinner v-if="loadingStatus.loadingVideo && currentVideo === video.name" />
                    <span v-else>▶️</span>
                  </button>
                  <button 
                    class="btn danger small" 
                    @click.stop="DeleteVideo(video.name)"
                    :disabled="loadingStatus.deletingVideo"
                    :title="'删除'"
                  >
                    <MobileSpinner v-if="loadingStatus.deletingVideo && deletingVideoName === video.name" />
                    <span v-else>🗑️</span>
                  </button>
                </div>
              </div>
            </transition-group>
            
            <div v-if="videos.length === 0 && !loadingStatus.fetchingVideos" class="empty-list">
              <div class="empty-icon">📭</div>
              <p>暂无文件</p>
              <p class="empty-sub mobile-show">开始录制～</p>
            </div>
            
            <!-- 手机端专用加载 -->
            <div v-if="loadingStatus.fetchingVideos" class="mobile-loading">
              <MobileSpinner class="list-spinner" />
              <p class="loading-text">加载文件...</p>
            </div>
          </div>
        </div>
      </aside>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import { apiMethods } from '@/utils/api'

const isRecording = ref(false)
const isReplayBufferActive = ref(false)
const videos = ref([])
const currentVideo = ref('')
const currentVideoName = ref('')
const videoRef = ref(null)
const deletingVideoName = ref('')

// 加载状态管理
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

// 手机端专用旋转组件
const MobileSpinner = {
  template: `<div class="mobile-spinner"></div>`
}

async function getRecordingStatus() {
  loadingStatus.gettingStatus = true
  try {
    const res = await apiMethods.IsRecording()
    isRecording.value = res.msg?.outputActive === true
  } catch (err) {
    console.error(err)
  } finally {
    loadingStatus.gettingStatus = false
  }
}

async function getReplayBufferStatus() {
  loadingStatus.gettingReplayStatus = true
  try {
    const res = await apiMethods.GetReplayBufferStatus()
    isReplayBufferActive.value = res.msg?.outputActive === true
  } catch (err) {
    console.error(err)
  } finally {
    loadingStatus.gettingReplayStatus = false
  }
}

async function startRecording() {
  loadingStatus.starting = true
  try {
    const res = await apiMethods.StartRecording()
    console.log("============",res.msg)
    if (res.status === 'success') {
      isRecording.value = true
      // 延迟后再查询状态，确保后端已更新
      setTimeout(() => {
        getRecordingStatus()
      }, 500)
      fetchVideos()
    }else {
      alert('❌'+res.msg)
    }
  } catch (err) {
    console.error(err)
  } finally {
    loadingStatus.starting = false
  }
}

async function stopRecording() {
  loadingStatus.stopping = true
  try {
    const res = await apiMethods.StopRecording()
    if (res.status === 'success') {
      isRecording.value = false
      // 延迟后再查询状态，确保后端已更新
      setTimeout(() => {
        getRecordingStatus()
      }, 1000)
      fetchVideos()
    }
  } catch (err) {
    console.error(err)
  } finally {
    loadingStatus.stopping = false
  }
}

async function startReplayBuffer() {
  loadingStatus.startingReplay = true
  try {
    const res = await apiMethods.StartReplayBuffer()
    console.log("============", res.msg)
    if (res.status === 'success') {
      isReplayBufferActive.value = true
      // 延迟后再查询状态，确保后端已更新
      setTimeout(() => {
        getReplayBufferStatus()
      }, 500)
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
  loadingStatus.stoppingReplay = true
  try {
    const res = await apiMethods.StopReplayBuffer()
    if (res.status === 'success') {
      isReplayBufferActive.value = false
      // 延迟后再查询状态，确保后端已更新
      setTimeout(() => {
        getReplayBufferStatus()
      }, 500)
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
  loadingStatus.savingReplay = true
  try {
    const res = await apiMethods.SaveReplayBuffer()
    if (res.status === 'success') {
      alert('✨ 回放已保存！')
      fetchVideos()
    }else {
      alert('❌'+res.msg)
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
      videos.value = res.msg==null?[]:res.msg
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
    if (videoRef.value) {
      videoRef.value.load()
      await videoRef.value.play()
    }
  } catch (err) {
    console.error(err)
  } finally {
    loadingStatus.loadingVideo = false
  }
}

async function DeleteVideo(name) {
  if (!confirm(`删除 "${name}"？`)) return
  deletingVideoName.value = name
  loadingStatus.deletingVideo = true
  try {
    const res = await apiMethods.DeleteVideo(name)
    if (res.status === 'success') {
      fetchVideos()
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
  getRecordingStatus()
  getReplayBufferStatus()
  fetchVideos()
})
</script>

<style scoped>
:root {
  --pink-50: #fdf2f8;
  --pink-100: #fce7f3;
  --pink-200: #fbcfe8;
  --pink-300: #f9a8d4;
  --pink-500: #ec4899;
  --text-accent: #60a5fa; /* 淡蓝色，用于主要文字 */
  --blue-200: #dbeafe;
  --purple-400: #c084fc;
  --gray-50: #f9fafb;
  --gray-600: #4b5563;
  --white: #ffffff;
  --shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

.obs-container {
  min-height: 100vh;
  background: linear-gradient(135deg, var(--pink-50), var(--blue-200));
  font-family: 'Noto Sans SC', 'Microsoft YaHei', sans-serif;
  position: relative;
  overflow-x: hidden;
}

.sky-bg {
  position: fixed;
  inset: 0;
  z-index: -1;
  pointer-events: none;
}

.cloud {
  position: absolute;
  background: rgba(255, 255, 255, 0.7);
  border-radius: 50px;
  filter: blur(8px);
  opacity: 0.8;
}
.cloud-1 { width: 200px; height: 60px; top: 20%; left: -100px; animation: float 25s ease-in-out infinite; }
.cloud-2 { width: 150px; height: 50px; top: 60%; right: -80px; animation: float 30s ease-in-out infinite reverse; }

.sakura {
  position: absolute;
  width: 6px; height: 6px;
  background: var(--pink-300);
  border-radius: 50%;
  animation: fall 15s linear infinite;
}
.sakura-1 { top: -10px; left: 20%; }
.sakura-2 { top: -10px; left: 80%; animation-delay: -5s; }

.sparkle {
  position: absolute;
  width: 2px; height: 2px;
  background: var(--purple-400);
  border-radius: 50%;
  animation: twinkle 3s ease-in-out infinite;
}
.sparkle-1 { top: 30%; left: 10%; }
.sparkle-2 { top: 70%; right: 20%; animation-delay: 1.5s; }

.main-layout {
  display: grid;
  grid-template-columns: 280px 1fr 340px;
  gap: 24px;
  max-width: 1700px;
  margin: 0 auto;
  padding: 0 20px 40px;
  min-height: calc(100vh - 140px);
}

.card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.8);
  box-shadow: var(--shadow);
  overflow: hidden;
  transition: all 0.3s ease;
}
.card:hover { transform: translateY(-4px); box-shadow: var(--shadow-lg); }

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px 16px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
}
.card-title {
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--text-accent);
}

.control-sidebar { grid-column: 1; }
.control-card { height: fit-content; }

.status-section {
  padding: 0 24px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.status-badge {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: var(--pink-100);
  border-radius: 16px;
  border: 1px solid var(--pink-200);
  color: var(--gray-600);
}
.status-badge.recording {
  background: #fee2e2;
  border-color: var(--pink-500);
}
.status-badge.replay-status.recording {
  background: #d1fae5;
  border-color: #10b981;
}
.status-icon .dot {
  width: 12px; height: 12px;
  border-radius: 50%;
  background: #a3a3a3;
}
.status-icon .dot.loading {
  background: var(--purple-400);
  animation: spin 1s linear infinite;
}
.status-badge.recording .dot {
  background: var(--pink-500);
  animation: pulse 1.5s ease-in-out infinite;
}
.status-content .status-main {
  font-size: 1.1rem;
  font-weight: 700;
}
.status-sub {
  font-size: 0.85rem;
  color: #4a4060;
}

.control-buttons {
  padding: 0 24px 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.section-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--gray-600);
  margin-top: 8px;
  padding-left: 4px;
  opacity: 0.8;
}
.section-title:first-child {
  margin-top: 0;
}
.btn-row { 
  display: flex; 
  gap: 12px; 
}
.btn-row .btn {
  flex: 1;
}
.btn.large { padding: 14px 20px; font-size: 1rem; }
.btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border: none;
  border-radius: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #0d59e7;
  color: #2d1b3a;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
.btn.primary {
  background: var(--pink-500);
  color: rgb(14, 226, 241);
}
.btn.secondary {
  background: #d1d5db;
  color: var(--gray-600);
}
.btn.accent {
  background: var(--purple-400);
  color: rgb(14, 226, 233);
}
.btn.ghost {
  background: transparent;
  color: #4a4060;
  border: 1px solid #d1d5db;
}
.btn.danger {
  background: #fca5a5;
  color: white;
}
.btn:hover:not(:disabled) { transform: translateY(-2px); }
.btn:active:not(:disabled) { transform: translateY(0); }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }

.player-section { 
  grid-column: 2; 
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 600px;
}

.player-container {
  width: 100%;
  max-width: 1000px;
}
.player-card {
  padding: 0;
}
.player-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
}
.player-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.3rem;
  font-weight: 700;
  color: var(--text-accent);
}
.player-title .icon { font-size: 1.4rem; }

.video-wrapper {
  padding: 24px;
}
.main-video {
  width: 100%;
  height: auto;
  max-height: 600px;
  border-radius: 16px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
}

.empty-player {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 400px;
  text-align: center;
  color: #4a4060;
}
.empty-icon { font-size: 4rem; margin-bottom: 16px; }
.empty-content h3 { font-size: 1.5rem; margin: 0 0 8px 0; }
.empty-content p { margin: 0; font-size: 1rem; }

.video-sidebar { grid-column: 3; }
.list-card { height: fit-content; max-height: 80vh; }

.video-list {
  max-height: calc(80vh - 120px);
  overflow-y: auto;
  padding-right: 8px;
}
.video-item {
  display: flex;
  gap: 12px;
  padding: 16px 24px;
  cursor: pointer;
  transition: all 0.2s ease;
  border-radius: 12px;
  margin: 0 4px;
}
.video-item:hover { background: var(--pink-50); transform: translateX(4px); }
.video-thumbnail {
  flex-shrink: 0;
  width: 64px;
  height: 48px;
}
.thumbnail-placeholder {
  width: 100%;
  height: 100%;
  background: var(--pink-100);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.thumb-icon { font-size: 1.5rem; color: var(--pink-500); }

.video-info {
  flex: 1;
  min-width: 0;
}
.video-title {
  font-weight: 600;
  color: var(--text-accent);
  /* 在移动端允许换行以显示全名 */
  white-space: normal;
  overflow: visible;
  text-overflow: initial;
  margin-bottom: 4px;
}
.video-meta {
  display: flex;
  gap: 8px;
  font-size: 0.85rem;
  color: var(--gray-600);
}
.video-actions { 
  display: flex; 
  flex-direction: column; 
  gap: 6px; 
  align-items: flex-end;
  justify-content: center;
}
.btn.small { 
  padding: 8px;
  font-size: 1rem;
  min-width: 42px;
  min-height: 42px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.btn.close-btn {
  min-width: 36px;
  min-height: 36px;
  padding: 6px;
  font-size: 1.2rem;
  line-height: 1;
}

.empty-list {
  padding: 60px 24px;
  text-align: center;
  color: #4a4060;
}
.empty-list .empty-icon { font-size: 3rem; margin-bottom: 16px; }
.empty-sub { font-size: 0.9rem; margin-top: 4px; }

/* ========== 手机端优化 ========== */
.mobile-hide { display: none; }
.mobile-show { display: block; }

/* 桌面上显示文件大小，移动端隐藏以节省空间 */
@media (min-width: 1081px) {
  .video-meta { display: flex !important; }
}
@media (max-width: 1080px) {
  .video-meta { display: none !important; }
}
.btn-text { white-space: nowrap; }

.mobile-spinner {
  width: 20px !important;
  height: 20px !important;
  border: 2px solid transparent;
  border-top: 2px solid var(--pink-500);
  border-radius: 50%;
  animation: mobile-spin 0.8s linear infinite !important;
}

.list-spinner {
  width: 28px !important;
  height: 28px !important;
  border-width: 3px !important;
  display: block !important;
  margin: 20px auto !important;
}

.loading-text {
  text-align: center !important;
  font-size: 14px !important;
  color: #4a4060 !important;
  margin: 0 !important;
}

.mobile-loading {
  padding: 40px 16px !important;
  text-align: center !important;
}

/* ========== 响应式 ========== */
@media (max-width: 1080px) {
  .main-layout { 
    grid-template-columns: 1fr; 
    max-width: 800px; 
    gap: 16px;
  }
  /* 确保视频标题与 meta 在窄屏时完整显示 */
  .video-title { white-space: normal; word-break: break-word; }
  .control-sidebar, .video-sidebar { grid-column: 1; }
  .player-section { min-height: 400px; }
  .player-container { max-width: 800px; }
  .main-video { max-height: 450px; }
  .btn.large { padding: 12px 16px; }
  .btn.small { padding: 5px 10px; }
}

@media (max-width: 768px) {
  .main-layout { 
    padding: 0 12px 20px; 
    gap: 12px;
    display: flex;
    flex-direction: column;
  }
  
  /* 控制区域优化 - 保持显示但更紧凑 */
  .control-sidebar { 
    order: 1;
    width: 100%;
  }
  .control-card { 
    padding: 0;
  }
  .card-header { 
    padding: 14px 16px 12px; 
  }
  .card-title {
    font-size: 1.1rem;
  }
  .status-section {
    padding: 0 16px 14px;
    gap: 10px;
  }
  .status-badge {
    padding: 12px;
  }
  .status-main { 
    font-size: 0.95rem;
  }
  .status-sub { 
    font-size: 0.75rem;
  }
  
  /* 按钮区域优化 */
  .control-buttons { 
    padding: 0 16px 16px; 
    gap: 8px;
  }
  .section-title {
    font-size: 0.85rem;
    margin-top: 6px;
  }
  .btn-row { 
    gap: 10px; 
  }
  .btn.large { 
    padding: 13px 18px; 
    font-size: 0.95rem;
  }
  .btn-text { 
    font-size: 0.85rem;
  }
  .btn-icon {
    font-size: 1rem;
  }
  
  /* 视频列表区域 - 第二显示 */
  .video-sidebar { 
    order: 2;
    width: 100%;
  }
  .list-card { 
    max-height: none;
  }
  .video-list { 
    max-height: 400px;
    overflow-y: auto;
  }
  .video-item { 
    padding: 10px 16px;
    gap: 10px;
  }
  .video-thumbnail { 
    width: 52px;
    height: 39px;
  }
  .video-title { 
    font-size: 0.9rem;
    line-height: 1.4;
  }
  .video-actions { 
    gap: 4px;
  }
  .btn.small { 
    width: 38px;
    height: 38px;
    padding: 0;
    min-width: 38px;
    font-size: 1rem;
  }
  
  /* 播放器区域 - 最后显示（选择后展开） */
  .player-section { 
    order: 3;
    min-height: auto;
  }
  .player-container { 
    max-width: 100%; 
  }
  .main-video { 
    max-height: 350px;
  }
  .player-header {
    padding: 14px 16px;
  }
  .player-title {
    font-size: 1.1rem;
  }
  .video-wrapper {
    padding: 16px;
  }
  .empty-player {
    height: 280px;
  }
  .empty-icon {
    font-size: 3rem;
  }
  .empty-content h3 {
    font-size: 1.2rem;
  }
  
  /* 移动端专用样式 */
  .close-btn {
    min-width: 40px !important;
    min-height: 40px !important;
    padding: 8px !important;
    font-size: 1.3rem !important;
  }
  .mobile-hide { 
    display: none !important; 
  }
  .mobile-show { 
    display: block !important; 
  }
  
  /* 动画简化 */
  .list-enter-active, 
  .list-leave-active,
  .fade-enter-active, 
  .fade-leave-active { 
    transition: all 0.2s ease; 
  }
  .list-enter-from, 
  .list-leave-to { 
    opacity: 0; 
    transform: translateY(-8px); 
  }
}

/* ========== 动画 ========== */
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes mobile-spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

@keyframes loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

@keyframes float { 0%, 100% { transform: translateY(0px); } 50% { transform: translateY(-20px); } }
@keyframes fall { to { transform: translateY(100vh) rotate(360deg); } }
@keyframes twinkle { 0%, 100% { opacity: 0; transform: scale(0); } 50% { opacity: 1; transform: scale(1); } }
@keyframes pulse { 0%, 100% { transform: scale(1); } 50% { transform: scale(1.2); } }

.fade-enter-from, .fade-leave-to { opacity: 0; transform: translateY(20px); }

.video-list::-webkit-scrollbar { width: 6px; }
.video-list::-webkit-scrollbar-track { background: transparent; }
.video-list::-webkit-scrollbar-thumb { background: var(--pink-300); border-radius: 3px; }
</style>