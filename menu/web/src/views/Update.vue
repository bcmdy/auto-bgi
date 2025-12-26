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
          <span class="text">ABGI 核心组件</span>
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
          <span class="text">BGI 远程组件</span>
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
            <span v-if="downloading">下载中...</span>
            <span v-else>{{ bgiCanUpdate ? '在线更新 (Download)' : '无需更新' }}</span>
          </button>
        </div>
      </div>

      <div v-if="note" class="error-note">
        {{ note }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'

// --- State Definitions ---
const currentVersion = ref('加载中...')
const latestVersion = ref('加载中...')
const loading = ref(false)
const checking = ref(false)
const note = ref('')
const downloading = ref(false)

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

// 刷新 ABGI
const refresh = async () => {
  checking.value = true
  note.value = ''
  try {
    const cur = await apiMethods.aBgiGetCurrentVersion()
    currentVersion.value = cur?.version ?? cur?.data?.version ?? (typeof cur === 'string' ? cur : JSON.stringify(cur))

    const last = await apiMethods.aBgiGetLastVersion()
    latestVersion.value = last?.version ?? last?.data?.version ?? (typeof last === 'string' ? last : JSON.stringify(last))
  } catch (err) {
    console.error(err)
    message.error('获取版本信息失败')
    note.value = err?.message || String(err)
  } finally {
    checking.value = false
  }
}

// 执行 ABGI 更新
const doUpdate = async () => {
  if (!isDifferent.value) return
  loading.value = true
  note.value = ''
  try {
     await apiMethods.aBgiUpdate()
      setTimeout(() => {
        window.location.href = '/' // 跳转到登录页
      }, 3500)
      
  } catch (err) {
    console.error(err)
    message.error('更新失败：' + (err?.message || String(err)))
    note.value = err?.message || String(err)
  } finally {
    loading.value = false
  }
}

// 刷新 BGI 版本
const refreshBgiVersions = async () => {
  bgiCurrentVersion.value = '加载中...'
  bgiLatestVersion.value = '加载中...'
  bgiCanUpdate.value = false
  try {
    const res = await apiMethods.aBgiGetVersions()
    if (res && typeof res === 'object') {
      console.debug('aBgiGetVersions response:', res)
      bgiCurrentVersion.value = res.currentVersion ?? res.current ?? bgiCurrentVersion.value
      bgiLatestVersion.value = res.lastVersion ?? res.latest ?? bgiLatestVersion.value

      if (Object.prototype.hasOwnProperty.call(res, 'canUpdate')) {
        if (typeof res.canUpdate === 'boolean') bgiCanUpdate.value = res.canUpdate
        else if (typeof res.canUpdate === 'string') bgiCanUpdate.value = res.canUpdate === 'true'
        else bgiCanUpdate.value = Boolean(res.canUpdate)
      } else {
        bgiCanUpdate.value = normalize(bgiCurrentVersion.value) !== normalize(bgiLatestVersion.value) && bgiLatestVersion.value !== ''
      }
    } else {
      bgiCurrentVersion.value = '未知'
      bgiLatestVersion.value = ''
      bgiCanUpdate.value = false
    }
  } catch (err) {
    console.warn('刷新 BGI 版本失败', err)
    bgiCurrentVersion.value = '获取失败'
    bgiLatestVersion.value = ''
    bgiCanUpdate.value = false
  }
}

// 执行 BGI 更新 (BGI 不需要跳转)
const downloadByUrl = async () => {
  downloading.value = true
  try {
    const res = await apiMethods.downloadBgi()
    message.success((res && (res.message || res.msg)) || '下载更新请求已发送')
    await refreshBgiVersions()
  } catch (err) {
    console.error(err)
    message.error('通过 URL 更新失败：' + (err?.message || String(err)))
    note.value = err?.message || String(err)
  } finally {
    downloading.value = false
  }
}

onMounted(() => {
  refresh()
  refreshBgiVersions()
})
</script>

<style scoped>
/* 全局容器与背景 
  二次元粉色系渐变背景
*/
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

/* 卡片主体 */
.update-card {
  position: relative;
  width: 100%;
  max-width: 600px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  border-radius: 24px;
  padding: 30px;
  box-shadow: 0 10px 40px rgba(255, 182, 193, 0.3), 
              0 0 0 1px rgba(255, 255, 255, 0.6) inset;
  overflow: hidden;
  margin-top: 60px; /* 给上方留点空间 */
}

/* 顶部粉色装饰条 */
.card-deco {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 6px;
  background: linear-gradient(90deg, #ff9eb5, #ff69b4);
}

/* 赞助商标签 - 像个可爱的贴纸 */
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
  box-shadow: 0 2px 8px rgba(255, 105, 180, 0.1);
  margin-bottom: 20px;
}
.sponsor-badge .heart {
  color: #ff1493;
  margin-right: 6px;
  animation: beat 1.5s infinite;
}
.sponsor-badge .name {
  color: #ff1493;
  margin: 0 4px;
  font-weight: 800;
}

@keyframes beat {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.2); }
}

.title {
  font-size: 24px;
  color: #4a4a4a;
  margin-bottom: 30px;
  font-weight: 700;
  text-align: center;
}

/* 模块块级样式 */
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
.section-header .icon {
  font-size: 20px;
  margin-right: 8px;
}
.section-header .text {
  font-size: 16px;
  font-weight: 700;
  color: #555;
  flex: 1;
}
.refresh-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 18px;
  color: #aaa;
  transition: color 0.3s;
  padding: 4px;
}
.refresh-btn:hover {
  color: #ff69b4;
}
.spin {
  display: inline-block;
  animation: spin 1s linear infinite;
}
@keyframes spin { 100% { transform: rotate(360deg); } }

/* 版本对比网格 */
.version-grid {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  background: #fff;
  padding: 15px;
  border-radius: 12px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.02);
}
.v-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}
.v-item .label {
  font-size: 12px;
  color: #999;
  margin-bottom: 4px;
}
.v-item .value {
  font-size: 15px;
  font-weight: 600;
  color: #333;
  text-align: center;
  word-break: break-all;
}
.v-item .value.highlight {
  color: #ff69b4;
}
.v-arrow {
  color: #ddd;
  font-size: 18px;
  padding: 0 10px;
}

/* 操作区域 */
.action-area {
  text-align: center;
}
.tip-text {
  font-size: 12px;
  color: #ff8da1;
  margin-top: 8px;
  margin-bottom: 0;
}

/* 按钮基础样式 */
.anime-btn {
  width: 100%;
  padding: 12px;
  border-radius: 50px; /* 胶囊按钮 */
  border: none;
  font-weight: 700;
  font-size: 15px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}
.anime-btn:active {
  transform: scale(0.98);
}

/* 主要按钮 (实心粉色) */
.anime-btn.primary {
  background: linear-gradient(90deg, #ff9eb5, #ff69b4);
  color: white;
  box-shadow: 0 4px 15px rgba(255, 105, 180, 0.4);
}
.anime-btn.primary:hover:not(:disabled) {
  background: linear-gradient(90deg, #ffaec0, #ff7ac1);
  box-shadow: 0 6px 20px rgba(255, 105, 180, 0.6);
}

/* 次要按钮 (空心/浅色) */
.anime-btn.secondary {
  background: #fff0f5;
  color: #ff69b4;
  border: 1px solid #ffb6c1;
}
.anime-btn.secondary:hover:not(:disabled) {
  background: #ffe4e1;
}

/* 禁用状态 - 灰色，不可点击 */
.anime-btn:disabled {
  background: #e0e0e0;
  color: #999;
  border: 1px solid #ccc;
  cursor: not-allowed;
  box-shadow: none;
}

/* 分割线 */
.divider {
  height: 1px;
  background-image: linear-gradient(to right, #ccc 0%, #ccc 50%, transparent 50%);
  background-size: 8px 1px;
  background-repeat: repeat-x;
  opacity: 0.3;
  margin: 25px 0;
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

/* 移动端适配 */
@media (max-width: 480px) {
  .anime-bg {
    padding: 10px;
  }
  .update-card {
    margin-top: 20px;
    padding: 20px 15px;
  }
  .title {
    font-size: 20px;
  }
  .version-grid {
    flex-direction: column;
    gap: 10px;
    text-align: center;
  }
  .v-arrow {
    transform: rotate(90deg);
  }
}
</style>