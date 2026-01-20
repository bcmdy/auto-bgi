<template>
  <div class="anime-bg">
    <div class="update-card">
      <div class="card-deco"></div>

      <div class="sponsor-badge">
        <span class="heart">❤</span> 感谢 <span class="name">思姐</span> 赞助
      </div>

      <h2 class="title">系统版本管理</h2>

      <div class="section-block">
        <div class="section-header">
          <span class="icon">🌸</span>
          <span class="text">ABGI 在线更新</span>
          <button class="refresh-btn" @click="refresh" :disabled="checking" title="刷新状态">
            <span :class="{ 'spin': checking }">↻</span>
          </button>
        </div>

        <div class="version-grid">
          <div class="v-item">
            <span class="label">当前版本</span>
            <span class="value">{{ currentVersion }}</span>
          </div>
          <div class="v-arrow">➜</div>
          <div class="v-item">
            <span class="label">最新版本</span>
            <span class="value highlight">{{ latestVersion }}</span>
          </div>
        </div>

        <div class="action-area">
          <button 
            class="anime-btn primary" 
            @click="doUpdate" 
            :disabled="!isDifferent || loading"
          >
            <span v-if="loading">更新中...</span>
            <span v-else>{{ isDifferent ? '立即更新 (Update)' : '已是最新版' }}</span>
          </button>
          <p class="tip-text" v-if="isDifferent">⚠️ 更新成功后将自动跳转至登录页</p>
        </div>
      </div>

      <div class="divider"></div>

      <div class="section-block">
        <div class="section-header">
          <span class="icon">🎀</span>
          <span class="text">茶包BGI 在线更新</span>
          <button class="help-btn" @click="openDisclaimer" title="免责声明">?</button>
          <button class="refresh-btn" @click="refreshBgiVersions" :disabled="downloading" title="刷新状态">
            <span>↻</span>
          </button>
        </div>

        <div class="version-grid">
          <div class="v-item">
            <span class="label">当前版本</span>
            <span class="value">{{ bgiCurrentVersion }}</span>
          </div>
          <div class="v-arrow">➜</div>
          <div class="v-item">
            <span class="label">最新版本</span>
            <span class="value highlight">{{ bgiLatestVersion }}</span>
          </div>
        </div>

        <div class="action-area">
          <button 
            class="anime-btn secondary" 
            @click="downloadByUrl" 
            :disabled="!bgiCanUpdate || downloading"
          >
            <span v-if="downloading">下载中 {{ downloadPercent }}%</span>
            <span v-else>{{ bgiCanUpdate ? '在线更新 (Download)' : '无需更新' }}</span>
          </button>

          <div v-if="downloading" class="progress-wrapper">
            <div class="progress-container">
              <div class="progress-bar" :style="{ width: downloadPercent + '%' }"></div>
            </div>
            <div class="progress-info">正在从服务器获取分片数据...</div>
          </div>
        </div>
      </div>

      <div v-if="note" class="error-note">
        {{ note }}
      </div>
      <!-- Disclaimer modal -->
      <div v-if="showDisclaimer" class="modal-overlay" @click.self="closeDisclaimer">
        <div class="disclaimer-modal" role="dialog" aria-modal="true">
          <div class="modal-header">
            <h3>免责声明</h3>
            <button class="modal-close" @click="closeDisclaimer" aria-label="关闭">✕</button>
          </div>
          <div class="modal-body">
            <p>请确保bgi的文件夹是默认名字：BetterGI</p>
            <p>本工具提供“茶包BGI 在线更新”功能仅为方便用户获取更新，开发者和供应方不对因更新引起的任何直接或间接损失承担责任。</p>
          </div>
          <div class="modal-footer">
            <button class="anime-btn secondary" @click="closeDisclaimer">我已知晓</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { message } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'

// --- State Definitions ---
const currentVersion = ref('加载中...')
const latestVersion = ref('加载中...')
const loading = ref(false)
const checking = ref(false)
const note = ref('')
const downloading = ref(false)
const downloadPercent = ref(0) // 新增：下载进度百分比

// BGI State
const bgiCurrentVersion = ref('加载中...')
const bgiLatestVersion = ref('加载中...')
const bgiCanUpdate = ref(false)

// --- Helpers ---
const normalize = (v) => (v == null ? '' : String(v).trim())

// --- Computed ---
const isDifferent = computed(() => {
  return normalize(currentVersion.value) !== normalize(latestVersion.value) && latestVersion.value !== ''
})

// --- Methods ---

const refresh = async () => {
  checking.value = true
  note.value = ''
  try {
    const cur = await apiMethods.aBgiGetCurrentVersion()
    currentVersion.value = cur?.version ?? cur?.data?.version ?? (typeof cur === 'string' ? cur : JSON.stringify(cur))

    const last = await apiMethods.aBgiGetLastVersion()
    latestVersion.value = last?.version ?? last?.data?.version ?? (typeof last === 'string' ? last : JSON.stringify(last))
  } catch (err) {
    message.error('获取版本信息失败')
    note.value = err?.message || String(err)
  } finally {
    checking.value = false
  }
}

const doUpdate = async () => {
  if (!isDifferent.value) return
  loading.value = true
  note.value = ''
  try {
    await apiMethods.aBgiUpdate()
    setTimeout(() => {
      window.location.href = '/'
    }, 3500)
  } catch (err) {
    if (err.status === 888) {
      message.info('更新已启动，等待系统重启中，请稍后...')
      return
    }
    message.error((err?.message || String(err)))
  } finally {
    loading.value = false
  }
}

const refreshBgiVersions = async () => {
  bgiCurrentVersion.value = '加载中...'
  bgiLatestVersion.value = '加载中...'
  bgiCanUpdate.value = false
  try {
    const res = await apiMethods.aBgiGetVersions()
    if (res) {
      bgiCurrentVersion.value = res.currentVersion ?? res.current ?? bgiCurrentVersion.value
      bgiLatestVersion.value = res.lastVersion ?? res.latest ?? bgiLatestVersion.value
      // bgiCanUpdate.value = !!res.canUpdate
      if (normalize(bgiCurrentVersion.value) !== normalize(bgiLatestVersion.value) && bgiLatestVersion.value !== '') {
        bgiCanUpdate.value = true
      } else {
        bgiCanUpdate.value = false
      }
    }
  } catch (err) {
    console.warn('刷新 BGI 版本失败', err)
  }
}

let bgiTimerId = null

const startBgiPolling = () => {
  if (bgiTimerId) {
    clearInterval(bgiTimerId)
  }
  bgiTimerId = setInterval(async () => {
    try {
      const status = await apiMethods.getBgiDownloadStatus()
      if (!status) return

      if (typeof status.percent !== 'undefined') {
        downloadPercent.value = parseFloat(status.percent) || 0
      }

      if (status.status === 'done') {
        downloading.value = false
        downloadPercent.value = 100
        clearInterval(bgiTimerId)
        bgiTimerId = null
        await refreshBgiVersions()
        message.success('更新包下载完成！')
      }

      if (status.status === 'error') {
        downloading.value = false
        clearInterval(bgiTimerId)
        bgiTimerId = null
        note.value = status.error || '下载失败'
        message.error(note.value)
      }
    } catch (err) {
      console.warn('轮询 BGI 下载状态失败', err)
    }
  }, 1000)
}

const downloadByUrl = () => {
  if (downloading.value) return
  downloading.value = true
  downloadPercent.value = 0
  note.value = ''
  apiMethods.downloadBgi()
    .then(() => {
      startBgiPolling()
    })
    .catch((err) => {
      downloading.value = false
      message.error(err?.message || '启动下载失败')
    })
}

const resumeBgiDownloadIfNeeded = async () => {
  try {
    const status = await apiMethods.getBgiDownloadStatus()
    if (!status) return
    if (status.status === 'downloading') {
      downloading.value = true
      if (typeof status.percent !== 'undefined') {
        downloadPercent.value = parseFloat(status.percent) || 0
      }
      startBgiPolling()
    }
  } catch (err) {
    console.warn('检查 BGI 下载状态失败', err)
  }
}

onMounted(() => {
  refresh()
  refreshBgiVersions()
  resumeBgiDownloadIfNeeded()
})

onUnmounted(() => {
  if (bgiTimerId) {
    clearInterval(bgiTimerId)
    bgiTimerId = null
  }
})

// --- Disclaimer modal state & handlers ---
const showDisclaimer = ref(false)

const openDisclaimer = () => {
  showDisclaimer.value = true
}

const closeDisclaimer = () => {
  showDisclaimer.value = false
}

const _onKeydown = (e) => {
  if (e && e.key === 'Escape') {
    showDisclaimer.value = false
  }
}

onMounted(() => {
  window.addEventListener('keydown', _onKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', _onKeydown)
})
</script>

<style scoped>
/* 原有样式保持... */
.anime-bg {
  min-height: 100vh;
  width: 100%;
  background: linear-gradient(135deg, #fff0f5 0%, #ffe4e1 100%);
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 40px 16px;
  box-sizing: border-box;
}

.update-card {
  position: relative;
  width: 100%;
  max-width: 600px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  border-radius: 24px;
  padding: 30px;
  box-shadow: 0 10px 40px rgba(255, 182, 193, 0.3), 0 0 0 1px rgba(255, 255, 255, 0.6) inset;
  overflow: hidden;
  margin-top: 60px;
}

.card-deco {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 6px;
  background: linear-gradient(90deg, #ff9eb5, #ff69b4);
}

.sponsor-badge {
  display: inline-flex;
  align-items: center;
  background: #fff;
  border: 1px solid #ffd1dc;
  color: #ff69b4;
  padding: 6px 16px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 20px;
}

.title {
  font-size: 24px;
  color: #4a4a4a;
  margin-bottom: 30px;
  font-weight: 700;
  text-align: center;
}

.section-block {
  background: rgba(255, 255, 255, 0.5);
  border-radius: 16px;
  padding: 20px;
  border: 1px solid rgba(255, 255, 255, 0.8);
}

.section-header {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
}

.section-header .text {
  font-size: 16px;
  font-weight: 700;
  color: #555;
  flex: 1;
}

.version-grid {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  background: #fff;
  padding: 15px;
  border-radius: 12px;
}

.v-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}

.v-item .value.highlight {
  color: #ff69b4;
}

.anime-btn {
  width: 100%;
  padding: 12px;
  border-radius: 50px;
  border: none;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.3s ease;
}

.anime-btn.primary {
  background: linear-gradient(90deg, #ff9eb5, #ff69b4);
  color: white;
}

.anime-btn.secondary {
  background: #fff0f5;
  color: #ff69b4;
  border: 1px solid #ffb6c1;
}

.divider {
  height: 1px;
  background-image: linear-gradient(to right, #ccc 0%, #ccc 50%, transparent 50%);
  background-size: 8px 1px;
  background-repeat: repeat-x;
  margin: 25px 0;
  opacity: 0.3;
}

/* --- 新增：进度条样式 --- */
.progress-wrapper {
  margin-top: 15px;
}

.progress-container {
  width: 100%;
  height: 10px;
  background: #f0f0f0;
  border-radius: 5px;
  overflow: hidden;
  border: 1px solid #ffe4e1;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #ffb6c1, #ff69b4);
  transition: width 0.4s ease;
  box-shadow: 0 0 10px rgba(255, 105, 180, 0.3);
}

.progress-info {
  font-size: 11px;
  color: #ff9eb5;
  margin-top: 6px;
  text-align: center;
  font-style: italic;
}

.error-note {
  margin-top: 15px;
  padding: 10px;
  background: #fff1f0;
  border: 1px solid #ffa39e;
  border-radius: 8px;
  color: #cf1322;
  font-size: 12px;
  text-align: center;
}

/* Help button near section title */
.help-btn {
  background: transparent;
  border: 1px solid rgba(0,0,0,0.06);
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-right: 8px;
  cursor: pointer;
  font-weight: 700;
  color: #ff69b4;
}

/* Modal styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  padding: 20px;
}

.disclaimer-modal {
  width: 100%;
  max-width: 520px;
  background: #fff;
  border-radius: 12px;
  padding: 18px;
  box-shadow: 0 10px 40px rgba(0,0,0,0.25);
  animation: modalIn 200ms ease;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  color: #333;
}

.modal-close {
  background: transparent;
  border: none;
  font-size: 18px;
  cursor: pointer;
  color: #999;
}

.modal-body {
  margin-top: 12px;
  color: #444;
  font-size: 14px;
  line-height: 1.6;
}

.modal-footer {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

@keyframes modalIn {
  from { transform: translateY(-6px); opacity: 0 }
  to { transform: translateY(0); opacity: 1 }
}
</style>
