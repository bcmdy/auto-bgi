<template>
  <div class="elegant-bg">
    <div class="glass-card animate-fade-in">
      <div class="card-header">
        <div class="sponsor-tag">
          <span class="heart pulse">❤</span>
          <span>感谢 <strong class="name">思姐</strong> 赞助</span>
        </div>
        <div class="version-chip">v{{ currentVersion }} ➔ v{{ latestVersion }}</div>
      </div>

      <div class="title-group">
        <h2 class="main-title">系统版本管理</h2>
        <p class="sub-title">保持程序与 BetterGI 为最新版以获得最佳体验（别怂，更新就完事了）</p>
      </div>

      <div class="module-block">
        <div class="module-header">
          <div class="mod-title">
            <span class="mod-icon pink-glow">🌸</span>
            <span class="mod-text">ABGI 在线更新</span>
          </div>
          <button class="circle-btn" @click="refresh" :disabled="checking" title="刷新状态">
            <span class="icon-spin" :class="{ 'is-spinning': checking }">↻</span>
          </button>
        </div>

        <div class="version-display">
          <div class="v-card">
            <span class="v-label">当前版本</span>
            <span class="v-val" :title="currentVersion">{{ currentVersion }}</span>
          </div>
          <div class="v-arrow">
            <div class="arrow-line"></div>
            <span class="arrow-head">➔</span>
          </div>
          <div class="v-card highlight">
            <span class="v-label">最新版本</span>
            <span class="v-val" :title="latestVersion">{{ latestVersion }}</span>
          </div>
        </div>

        <div class="action-row">
          <button class="btn-primary" @click="doUpdate" :disabled="!isDifferent || loading">
            <span v-if="loading" class="btn-loader"></span>
            <span>{{ loading ? '更新中...' : (isDifferent ? '立即更新 (Update)' : '已是最新版') }}</span>
          </button>
          <div class="warning-text animate-slide-down" v-if="isDifferent">
            <span class="warn-icon">⚠️</span> 更新成功后将自动跳转至登录页
          </div>
        </div>
      </div>

      <div class="elegant-divider"></div>

      <div class="module-block">
        <div class="module-header">
          <div class="mod-title">
            <span class="mod-icon blue-glow">🎀</span>
            <span class="mod-text">BGI 在线更新</span>
          </div>
          <div class="mod-actions">
            <button class="circle-btn" @click="openBgiHistory" :disabled="downloading" title="历史版本">🕰️</button>
            <button class="circle-btn" @click="openDisclaimer" title="免责声明">?</button>
            <button class="circle-btn" @click="refreshBgiVersions" :disabled="downloading" title="刷新状态">
              <span class="icon-spin" :class="{ 'is-spinning': downloading }">↻</span>
            </button>
          </div>
        </div>

        <div class="version-display">
          <div class="v-card">
            <span class="v-label">当前版本</span>
            <span class="v-val" :title="bgiCurrentVersion">{{ bgiCurrentVersion }}</span>
          </div>
          <div class="v-arrow">
            <div class="arrow-line"></div>
            <span class="arrow-head">➔</span>
          </div>
          <div class="v-card highlight">
            <span class="v-label">最新版本</span>
            <span class="v-val" :title="bgiLatestVersion">{{ bgiLatestVersion }}</span>
          </div>
        </div>

        <div class="action-row">
          <button class="btn-secondary" @click="downloadByUrl" :disabled="!bgiCanUpdate || downloading" title="在线下载并替换当前 BGI">
            <div class="btn-bg-progress" v-if="downloading" :style="{ width: downloadPercent + '%' }"></div>
            <span class="btn-content" v-if="downloading">下载中 {{ downloadPercent }}%</span>
            <span class="btn-content" v-else>{{ bgiCanUpdate ? '在线更新 (Download)' : '无需更新' }}</span>
          </button>

          <div v-if="downloading" class="dl-progress-wrapper animate-slide-down">
            <div class="dl-track">
              <div class="dl-bar stripes" :style="{ width: downloadPercent + '%' }"></div>
            </div>
            <div class="dl-info">正在从服务器获取分片数据...</div>
          </div>
        </div>
      </div>

      <div v-if="note" class="error-alert animate-pop-in">
        <span class="err-icon">!</span> {{ note }}
      </div>

      <div v-if="showBgiHistory" class="modal-backdrop" @click.self="closeBgiHistory">
        <div class="modal-panel animate-pop-in">
          <div class="modal-header">
            <h3>🕰️ BGI 历史版本</h3>
            <button class="close-btn" @click="closeBgiHistory">✕</button>
          </div>
          <div class="modal-body">
            <div v-if="bgiHistoryLoading" class="loading-state">
              <span class="icon-spin is-spinning">↻</span> 加载中...
            </div>
            <div v-else-if="bgiHistoryVersions.length === 0" class="empty-state">
              暂无历史版本
            </div>
            <div v-else class="radio-list">
              <label v-for="ver in bgiHistoryVersions" :key="ver" class="radio-item" :class="{'is-active': bgiSelectedVersion === ver}">
                <input type="radio" name="bgiHistoryVer" :value="ver" v-model="bgiSelectedVersion" />
                <span class="ver-text">v{{ ver }}</span>
                <span class="check-circle"></span>
              </label>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn-outline" @click="closeBgiHistory">关闭</button>
            <button class="btn-primary mini" :disabled="!bgiSelectedVersion || bgiHistoryLoading || bgiRollingBack" @click="doBgiRollback">
              <span v-if="bgiRollingBack" class="btn-loader"></span>
              <span>{{ bgiRollingBack ? '回滚中...' : '确认回滚' }}</span>
            </button>
          </div>
        </div>
      </div>

      <div v-if="showDisclaimer" class="modal-backdrop" @click.self="closeDisclaimer">
        <div class="modal-panel animate-pop-in">
          <div class="modal-header">
            <h3>📝 免责声明</h3>
            <button class="close-btn" @click="closeDisclaimer">✕</button>
          </div>
          <div class="modal-body text-content">
            <div class="notice-box">
              <span class="notice-icon">💡</span>
              <p>请确保 BGI 的文件夹是默认名字：<strong>BetterGI</strong></p>
            </div>
            <p class="desc-text">本工具提供“BGI 在线更新”功能仅为方便用户获取更新，开发者和供应方不对因更新引起的任何直接或间接损失承担责任。</p>
            <p class="desc-text">很多反馈把本体删除的问题是正常的，程序模拟的就是手动更新操作：<strong>备份 ➔ 下载 ➔ 删除原包 ➔ 解压新包 ➔ 还原配置</strong>。</p>
            <p class="path-text">📂 备份路径 (ABGI根目录下)：<br>User 备份在 <code>/users</code>，Log 备份在 <code>/backups</code></p>
          </div>
          <div class="modal-footer center">
            <button class="btn-primary full" @click="closeDisclaimer">我已知晓</button>
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
const showBgiHistory = ref(false)
const bgiHistoryLoading = ref(false)
const bgiRollingBack = ref(false)
const bgiHistoryVersions = ref([])
const bgiSelectedVersion = ref('')

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

const openBgiHistory = async () => {
  if (downloading.value) return
  showBgiHistory.value = true
  bgiHistoryLoading.value = true
  bgiSelectedVersion.value = ''
  bgiHistoryVersions.value = []
  try {
    const res = await apiMethods.queryHistoryVersion('BetterGI.7z')
    if (!res || res.success !== true) {
      throw new Error(res?.message || '获取历史版本失败')
    }
    bgiHistoryVersions.value = Array.isArray(res.versions) ? res.versions : []
    if (bgiHistoryVersions.value.length > 0) {
      bgiSelectedVersion.value = bgiHistoryVersions.value[0]
    }
  } catch (err) {
    message.error(err?.message || String(err))
  } finally {
    bgiHistoryLoading.value = false
  }
}

const closeBgiHistory = () => {
  showBgiHistory.value = false
  setTimeout(() => {
    bgiHistoryVersions.value = []
    bgiSelectedVersion.value = ''
    bgiHistoryLoading.value = false
    bgiRollingBack.value = false
  }, 200)
}

const doBgiRollback = async () => {
  if (!bgiSelectedVersion.value) return
  bgiRollingBack.value = true
  try {
    const res = await apiMethods.bgiRollbackHistoryVersion({
      version: bgiSelectedVersion.value,
      jsName: 'BetterGI.7z'
    })
    if (!res || res.success !== true) {
      throw new Error(res?.message || '回滚失败')
    }
    message.success(res.message || '更新并还原成功')
    closeBgiHistory()
    await refreshBgiVersions()
  } catch (err) {
    message.error(err?.message || String(err))
  } finally {
    bgiRollingBack.value = false
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
/* ================= 全局 & 背景 ================= */
.elegant-bg {
  min-height: 100vh;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 50px 20px;
  box-sizing: border-box;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  background: linear-gradient(135deg, #fdfbfb 0%, #ebedee 100%);
  background-size: 400% 400%;
  animation: bgPan 15s ease infinite;
}

/* ================= 核心卡片 ================= */
.glass-card {
  width: 100%;
  max-width: 760px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-radius: 24px;
  padding: 36px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.05), 0 1px 3px rgba(0, 0, 0, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.8);
}

/* ================= 头部区域 ================= */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.sponsor-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: linear-gradient(120deg, #fff0f5, #ffe4e1);
  color: #d81b60;
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
  box-shadow: 0 2px 10px rgba(216, 27, 96, 0.1);
}
.sponsor-tag .name { color: #c2185b; font-weight: 800; }
.version-chip {
  font-size: 13px;
  color: #64748b;
  background: #f1f5f9;
  padding: 6px 12px;
  border-radius: 12px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.title-group { text-align: center; margin-bottom: 32px; }
.main-title {
  font-size: 26px;
  color: #1e293b;
  margin: 0 0 8px 0;
  font-weight: 800;
  letter-spacing: 1px;
}
.sub-title { color: #64748b; font-size: 14px; margin: 0; }

/* ================= 模块区块 ================= */
.module-block {
  background: #ffffff;
  border-radius: 18px;
  padding: 24px;
  border: 1px solid #e2e8f0;
  box-shadow: 0 4px 20px rgba(0,0,0,0.02);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}
.module-block:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 30px rgba(0,0,0,0.04);
}

.module-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.mod-title { display: flex; align-items: center; gap: 12px; }
.mod-icon {
  width: 38px;
  height: 38px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  font-size: 18px;
}
.pink-glow { background: #fff1f2; color: #f43f5e; }
.blue-glow { background: #f0f9ff; color: #0ea5e9; }
.mod-text { font-size: 17px; font-weight: 800; color: #334155; }
.mod-actions { display: flex; gap: 8px; }

.circle-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: #f8fafc;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s ease;
}
.circle-btn:hover:not(:disabled) {
  background: #e2e8f0;
  color: #0f172a;
  transform: scale(1.05);
}
.circle-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* ================= 版本展示对比 ================= */
.version-display {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #f8fafc;
  padding: 16px 20px;
  border-radius: 16px;
  margin-bottom: 20px;
  gap: 12px;
  flex-wrap: wrap;
}
.v-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  min-width: 100px;
}
.v-label { font-size: 12px; color: #94a3b8; font-weight: 600; text-transform: uppercase; }
.v-val {
  font-size: 18px;
  font-weight: 800;
  color: #334155;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-all;
  text-align: center;
  padding: 2px;
  line-height: 1.4;
  position: relative;
  cursor: help;
}
.v-val[title]:hover {
  background-color: rgba(244, 63, 94, 0.05);
  border-radius: 4px;
  padding: 4px;
}
.v-card.highlight .v-val {
  color: #f43f5e;
  font-size: 20px;
}

.v-arrow {
  display: flex;
  align-items: center;
  color: #cbd5e1;
  padding: 0 8px;
  flex-shrink: 0;
}
.arrow-line { width: 20px; height: 2px; background: #cbd5e1; border-radius: 2px; }
.arrow-head { margin-left: -3px; font-size: 16px; }

/* ================= 按钮 & 进度条 ================= */
.action-row { display: flex; flex-direction: column; gap: 12px; }

.btn-primary, .btn-secondary, .btn-outline {
  position: relative;
  width: 100%;
  padding: 14px 20px;
  border-radius: 14px;
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  border: none;
}

.btn-primary {
  background: linear-gradient(135deg, #f43f5e, #fb7185);
  color: white;
  box-shadow: 0 8px 20px rgba(244, 63, 94, 0.25);
}
.btn-primary:hover:not(:disabled) { box-shadow: 0 10px 25px rgba(244, 63, 94, 0.35); transform: translateY(-1px); }
.btn-primary:active:not(:disabled) { transform: translateY(1px); box-shadow: 0 4px 10px rgba(244, 63, 94, 0.2); }
.btn-primary:disabled { background: #cbd5e1; box-shadow: none; color: #f1f5f9; cursor: not-allowed; }

.btn-secondary {
  background: #ffffff;
  color: #f43f5e;
  border: 2px solid #ffe4e6;
}
.btn-secondary:hover:not(:disabled) { border-color: #fda4af; background: #fff1f2; }
.btn-secondary:disabled { border-color: #e2e8f0; color: #94a3b8; cursor: not-allowed; }

.btn-bg-progress {
  position: absolute;
  top: 0; left: 0; height: 100%;
  background: #fff1f2;
  transition: width 0.3s ease;
  z-index: 0;
}
.btn-content { position: relative; z-index: 1; }

.warning-text { font-size: 13px; color: #d97706; text-align: center; font-weight: 600; }
.warn-icon { margin-right: 4px; }

.dl-progress-wrapper { background: #f8fafc; padding: 16px; border-radius: 12px; }
.dl-track { width: 100%; height: 8px; background: #e2e8f0; border-radius: 4px; overflow: hidden; margin-bottom: 8px; }
.dl-bar {
  height: 100%;
  background: linear-gradient(90deg, #38bdf8, #818cf8);
  border-radius: 4px;
  transition: width 0.4s ease-out;
}
.dl-bar.stripes {
  background-image: linear-gradient(45deg, rgba(255,255,255,0.15) 25%, transparent 25%, transparent 50%, rgba(255,255,255,0.15) 50%, rgba(255,255,255,0.15) 75%, transparent 75%, transparent);
  background-size: 1rem 1rem;
  animation: progressStripes 1s linear infinite;
}
.dl-info { font-size: 12px; color: #64748b; text-align: center; }

.elegant-divider {
  height: 1px;
  background: linear-gradient(90deg, transparent, #e2e8f0, transparent);
  margin: 24px 0;
}

.error-alert {
  margin-top: 20px;
  padding: 14px;
  background: #fef2f2;
  border-left: 4px solid #ef4444;
  border-radius: 8px;
  color: #b91c1c;
  font-size: 14px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 10px;
}
.err-icon {
  background: #ef4444; color: white; width: 20px; height: 20px;
  display: inline-flex; align-items: center; justify-content: center;
  border-radius: 50%; font-size: 12px;
}

/* ================= 弹窗样式 ================= */
.modal-backdrop {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(15, 23, 42, 0.4);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  display: flex; align-items: center; justify-content: center;
  z-index: 9999;
  padding: 20px;
}
.modal-panel {
  background: #ffffff;
  width: 100%;
  max-width: 480px;
  border-radius: 20px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.modal-header {
  padding: 20px 24px;
  border-bottom: 1px solid #f1f5f9;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #f8fafc;
}
.modal-header h3 { margin: 0; font-size: 18px; color: #1e293b; font-weight: 800; }
.close-btn { background: none; border: none; font-size: 20px; color: #94a3b8; cursor: pointer; transition: color 0.2s; }
.close-btn:hover { color: #f43f5e; }

.modal-body { padding: 24px; max-height: 60vh; overflow-y: auto; }
.modal-body.text-content p { margin: 0 0 12px 0; line-height: 1.6; color: #475569; font-size: 14px; }
.notice-box {
  background: #fffbeb; padding: 12px 16px; border-radius: 12px;
  display: flex; gap: 10px; margin-bottom: 16px; align-items: flex-start;
}
.notice-box p { margin: 0 !important; color: #b45309; }
.path-text { background: #f1f5f9; padding: 12px; border-radius: 8px; font-family: monospace; font-size: 13px !important; }

.radio-list { display: flex; flex-direction: column; gap: 10px; }
.radio-item {
  display: flex; align-items: center; justify-content: space-between;
  padding: 14px 16px; border: 2px solid #e2e8f0; border-radius: 12px;
  cursor: pointer; transition: all 0.2s;
}
.radio-item:hover { border-color: #cbd5e1; background: #f8fafc; }
.radio-item.is-active { border-color: #f43f5e; background: #fff1f2; }
.radio-item input { display: none; }
.ver-text { font-weight: 700; color: #334155; font-size: 15px; }
.check-circle {
  width: 20px; height: 20px; border-radius: 50%; border: 2px solid #cbd5e1;
  position: relative; transition: all 0.2s;
}
.radio-item.is-active .check-circle {
  border-color: #f43f5e; background: #f43f5e;
}
.radio-item.is-active .check-circle::after {
  content: ""; position: absolute; top: 4px; left: 4px; width: 8px; height: 8px;
  background: white; border-radius: 50%;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid #f1f5f9;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background: #f8fafc;
}
.modal-footer.center { justify-content: center; }
.btn-primary.mini { width: auto; padding: 10px 24px; font-size: 14px; }
.btn-primary.full { width: 100%; }
.btn-outline {
  width: auto; padding: 10px 24px; border-radius: 14px; background: white;
  border: 1px solid #cbd5e1; color: #475569; font-weight: 700; cursor: pointer; transition: 0.2s;
}
.btn-outline:hover { background: #f1f5f9; color: #0f172a; }

/* ================= 动画与微交互 ================= */
.icon-spin { display: inline-block; }
.icon-spin.is-spinning { animation: spin 1s linear infinite; }
.pulse { animation: pulse 1.5s infinite; }

.btn-loader {
  width: 16px; height: 16px; border: 2px solid rgba(255,255,255,0.4);
  border-top-color: white; border-radius: 50%; animation: spin 0.8s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }
@keyframes pulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.15); }
  100% { transform: scale(1); }
}
@keyframes bgPan {
  0% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
  100% { background-position: 0% 50%; }
}
@keyframes progressStripes {
  from { background-position: 1rem 0; }
  to { background-position: 0 0; }
}

.animate-fade-in { animation: fadeIn 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards; }
.animate-pop-in { animation: popIn 0.4s cubic-bezier(0.16, 1, 0.3, 1) forwards; }
.animate-slide-down { animation: slideDown 0.3s ease-out forwards; }

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes popIn {
  from { opacity: 0; transform: scale(0.95) translateY(10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}
@keyframes slideDown {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

/* 响应式调整 */
@media (max-width: 600px) {
  .elegant-bg { padding: 20px 10px; }
  .glass-card { padding: 20px; }
  .version-display {
    padding: 12px 10px;
    flex-direction: column;
    gap: 10px;
  }
  .v-card {
    width: 100%;
    min-width: auto;
  }
  .v-val {
    font-size: 14px;
    max-height: 2.8em;
    line-height: 1.4;
    word-break: break-word;
    white-space: normal;
  }
  .v-card.highlight .v-val {
    font-size: 16px;
  }
  .v-arrow {
    transform: rotate(90deg);
    padding: 4px 0;
  }
  .arrow-line { width: 15px; }
  .mod-text { font-size: 15px; }
  .action-row { gap: 8px; }
}
</style>