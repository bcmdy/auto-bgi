<template>
  <div class="hotkey-page">
    <header class="page-header">
      <button class="nav-btn" @click="goHome">🏠 返回首页</button>
      <div class="title-wrap">
        <h1 class="page-title">BGI快捷键</h1>
      </div>
      <button class="nav-btn" :disabled="loading" @click="loadHotKeys">
        {{ loading ? '加载中...' : '🔄 刷新' }}
      </button>
    </header>

    <div class="content">
      <div class="remote">
        <div class="remote-screen">
          <div class="screen-row">
            <span class="screen-label">状态</span>
            <span class="screen-value">{{ screenText }}</span>
          </div>
          <div class="screen-row">
            <span class="screen-label">按键数</span>
            <span class="screen-value">{{ allItems.length }}</span>
          </div>
        </div>

        <div class="remote-body">
          <div class="top-row">
            <button
              class="btn-circle power-btn"
              :class="{ active: pressingKey === mainButtons.power.key }"
              :title="mainButtons.power.key"
              :disabled="!mainButtons.power.key || pressing"
              @click="press(mainButtons.power.key)"
            >
              <span class="icon">⏻</span>
              <span class="txt">
                <span class="name">{{ mainButtons.power.name }}</span>
                <span class="key">{{ mainButtons.power.key || '未配置' }}</span>
              </span>
            </button>
          </div>

          <button
            class="btn-pill primary"
            :class="{ active: pressingKey === mainButtons.oneLong.key }"
            :title="mainButtons.oneLong.key"
            :disabled="!mainButtons.oneLong.key || pressing"
            @click="press(mainButtons.oneLong.key)"
          >
            <span class="pill-name">{{ mainButtons.oneLong.name }}</span>
            <span class="pill-key">{{ mainButtons.oneLong.key || '未配置' }}</span>
          </button>

          <div class="dpad">
            <div class="dpad-grid">
              <div></div>
              <button
                class="btn-circle"
                :class="{ active: pressingKey === mainButtons.logToggle.key }"
                :title="mainButtons.logToggle.key"
                :disabled="!mainButtons.logToggle.key || pressing"
                @click="press(mainButtons.logToggle.key)"
              >
                <span class="icon">🪟</span>
                <span class="small-name">{{ mainButtons.logToggle.name }}</span>
              </button>
              <div></div>

              <button
                class="btn-circle"
                :class="{ active: pressingKey === mainButtons.pauseTask.key }"
                :title="mainButtons.pauseTask.key"
                :disabled="!mainButtons.pauseTask.key || pressing"
                @click="press(mainButtons.pauseTask.key)"
              >
                <span class="icon">⏸</span>
                <span class="small-name">{{ mainButtons.pauseTask.name }}</span>
              </button>

              <button
                class="btn-circle center danger"
                :class="{ active: pressingKey === mainButtons.stopTask.key }"
                :title="mainButtons.stopTask.key"
                :disabled="!mainButtons.stopTask.key || pressing"
                @click="press(mainButtons.stopTask.key)"
              >
                <span class="icon">⏹</span>
                <span class="small-name">{{ mainButtons.stopTask.name }}</span>
              </button>

              <button
                class="btn-circle"
                :class="{ active: pressingKey === mainButtons.screenshot.key }"
                :title="mainButtons.screenshot.key"
                :disabled="!mainButtons.screenshot.key || pressing"
                @click="press(mainButtons.screenshot.key)"
              >
                <span class="icon">📸</span>
                <span class="small-name">{{ mainButtons.screenshot.name }}</span>
              </button>

              <div></div>
              <button
                class="btn-circle"
                :class="{ active: pressingKey === mainButtons.extra.key }"
                :title="mainButtons.extra.key"
                :disabled="!mainButtons.extra.key || pressing"
                @click="press(mainButtons.extra.key)"
              >
                <span class="icon">⭐</span>
                <span class="small-name">{{ mainButtons.extra.name }}</span>
              </button>
              <div></div>
            </div>
          </div>
        </div>
      </div>


    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'

const router = useRouter()

const loading = ref(false)
const pressingKey = ref('')
const hotKeyMap = ref({})
const lastError = ref('')

const normalizeAction = (s) => String(s || '').replace(/\s+/g, '').replace(/／/g, '/')

const allItems = computed(() => {
  const entries = Object.entries(hotKeyMap.value || {})
  return entries
    .map(([action, key]) => ({ action, key }))
    .sort((a, b) => a.action.localeCompare(b.action, 'zh-Hans-CN'))
})

const allItemsWithNorm = computed(() => {
  return allItems.value.map((it) => ({
    ...it,
    norm: normalizeAction(it.action)
  }))
})

const findByNorm = (norm) => {
  return allItemsWithNorm.value.find((it) => it.norm === norm) || null
}

const mainButtons = computed(() => {
  const power = findByNorm(normalizeAction('启动停止BetterGI')) || { action: '启动停止BetterGI', key: '' }
  const oneLong = findByNorm(normalizeAction('启动停止一条龙')) || { action: '启动停止一条龙', key: '' }
  const logToggle = findByNorm(normalizeAction('日志与状态窗口展示开关')) || { action: '日志与状态窗口展示开关', key: '' }
  const stopTask = findByNorm(normalizeAction('停止当前脚本/独立任务')) || findByNorm(normalizeAction('停止当前脚本/ 独立任务')) || { action: '停止当前脚本/ 独立任务', key: '' }
  const pauseTask = findByNorm(normalizeAction('暂停当前脚本/独立任务')) || findByNorm(normalizeAction('暂停当前脚本/ 独立任务')) || { action: '暂停当前脚本/ 独立任务', key: '' }
  const screenshot = findByNorm(normalizeAction('游戏截图')) || { action: '游戏截图', key: '' }

  const used = new Set([power.action, oneLong.action, logToggle.action, stopTask.action, pauseTask.action, screenshot.action])
  const extraCandidate = allItems.value.find((it) => !used.has(it.action)) || { action: '更多动作', key: '' }

  return {
    power: { name: power.action, key: power.key },
    oneLong: { name: oneLong.action, key: oneLong.key },
    logToggle: { name: logToggle.action, key: logToggle.key },
    stopTask: { name: stopTask.action, key: stopTask.key },
    pauseTask: { name: pauseTask.action, key: pauseTask.key },
    screenshot: { name: screenshot.action, key: screenshot.key },
    extra: { name: extraCandidate.action, key: extraCandidate.key }
  }
})

const pressing = computed(() => !!pressingKey.value)

const screenText = computed(() => {
  if (pressingKey.value) return `正在按下：${pressingKey.value}`
  if (loading.value) return '加载中...'
  if (lastError.value) return lastError.value
  if (allItems.value.length === 0) return '未获取到快捷键配置'
  return '就绪'
})

const goHome = () => router.push('/')

const loadHotKeys = async () => {
  loading.value = true
  lastError.value = ''
  try {
    const res = await apiMethods.hotKeyQuery()
    const data = res?.data || {}
    hotKeyMap.value = data && typeof data === 'object' ? data : {}
  } catch (e) {
    lastError.value = '加载失败'
    hotKeyMap.value = {}
    message.error('快捷键配置加载失败')
  } finally {
    loading.value = false
  }
}

const press = async (key) => {
  const k = String(key || '').trim()
  if (!k) return
  pressingKey.value = k
  try {
    const res = await apiMethods.pressHotKey(k)
    if (res?.status && res.status !== 'success') {
      message.warning(res?.message || '请求已发送')
      return
    }
    message.success(`已按下：${k}`)
  } catch (e) {
    message.error(`按键失败：${k}`)
  } finally {
    pressingKey.value = ''
  }
}

onMounted(() => {
  loadHotKeys()
})
</script>

<style scoped>
.hotkey-page {
  min-height: 100vh;
  padding: 20px;
  background: radial-gradient(circle at 20% 10%, rgba(255, 190, 230, 0.5), transparent 45%),
    radial-gradient(circle at 80% 30%, rgba(190, 220, 255, 0.55), transparent 50%),
    linear-gradient(180deg, #fff7fb 0%, #fff 35%, #f7fbff 100%);
}

.page-header {
  display: grid;
  grid-template-columns: 160px 1fr 160px;
  gap: 12px;
  align-items: center;
  margin-bottom: 18px;
}

.nav-btn {
  border: 0;
  border-radius: 14px;
  padding: 10px 14px;
  font-size: 14px;
  color: #5b2b46;
  background: rgba(255, 255, 255, 0.85);
  box-shadow: 0 10px 24px rgba(255, 102, 163, 0.15);
  cursor: pointer;
}

.nav-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.title-wrap {
  text-align: center;
}

.page-title {
  margin: 0;
  font-size: 26px;
  color: #d63384;
  letter-spacing: 1px;
}

.page-subtitle {
  margin: 6px 0 0;
  font-size: 13px;
  color: rgba(120, 55, 85, 0.75);
}

.content { display: flex; justify-content: center; align-items: flex-start; gap: 18px; }

.remote {
  width: 100%;
  max-width: 420px;
  margin: 0 auto;
  border-radius: 40px;
  padding: 18px 16px 22px;
  background: linear-gradient(160deg, rgba(255, 255, 255, 0.86), rgba(250, 248, 255, 0.86));
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.7);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    0 22px 46px rgba(0, 0, 0, 0.08),
    0 8px 18px rgba(214, 51, 132, 0.12);
}

.remote-screen {
  border-radius: 16px;
  padding: 12px;
  background: linear-gradient(135deg, rgba(28, 28, 38, 0.95), rgba(25, 25, 35, 0.9));
  color: #e8e8ef;
  margin-bottom: 16px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.04);
}

.screen-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  line-height: 1.4;
}

.screen-label {
  opacity: 0.75;
}

.screen-value {
  font-weight: 600;
}

.remote-body {
  display: grid;
  grid-template-rows: auto auto 1fr;
  gap: 14px;
}

.top-row {
  display: flex;
  justify-content: center;
}

.btn-circle {
  appearance: none;
  border: 0;
  width: 88px;
  height: 88px;
  border-radius: 50%;
  background: radial-gradient(circle at 30% 25%, #fff0f5 0%, #ffd6e7 30%, #ffc2dd 60%, #ffb3d6 100%);
  box-shadow:
    0 10px 24px rgba(255, 102, 163, 0.18),
    inset 0 2px 0 rgba(255, 255, 255, 0.8);
  cursor: pointer;
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #6b1f46;
  transition: transform 0.08s ease, box-shadow 0.12s ease, filter 0.12s ease;
}

.btn-circle .icon {
  font-size: 22px;
  line-height: 1;
}

.btn-circle .small-name {
  font-size: 11px;
  margin-top: 4px;
  opacity: 0.9;
  text-align: center;
  padding: 0 6px;
}

.btn-circle .txt {
  display: grid;
  place-items: center;
}
.btn-circle .name {
  font-weight: 700;
  font-size: 13px;
  margin-top: 4px;
}
.btn-circle .key {
  font-size: 11px;
  opacity: 0.8;
}

.btn-circle:active:not(:disabled),
.btn-circle.active {
  transform: translateY(1px) scale(0.98);
  box-shadow:
    0 6px 14px rgba(255, 102, 163, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
  filter: brightness(0.98);
}

.power-btn {
  background: radial-gradient(circle at 35% 25%, #ffe9ef 0%, #ffc2d4 55%, #ffb3d0 75%, #ff9fc3 100%);
  color: #6b1f46;
}

.btn-pill {
  appearance: none;
  border: 0;
  padding: 12px 16px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: linear-gradient(135deg, rgba(238, 240, 255, 0.96), rgba(255, 255, 255, 0.9));
  color: #2c1230;
  box-shadow: 0 12px 26px rgba(99, 102, 241, 0.16);
  border: 1px solid rgba(99, 102, 241, 0.2);
  cursor: pointer;
  transition: transform 0.08s ease, box-shadow 0.12s ease, filter 0.12s ease;
}

.btn-pill .pill-name {
  font-weight: 800;
}
.btn-pill .pill-key {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  font-size: 12px;
  opacity: 0.75;
}

.btn-pill:active:not(:disabled),
.btn-pill.active {
  transform: translateY(1px) scale(0.99);
  box-shadow: 0 8px 18px rgba(99, 102, 241, 0.14);
  filter: brightness(0.98);
}

.btn-pill:disabled,
.btn-circle:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.dpad {
  padding: 10px 0 2px;
}
.dpad-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  grid-template-rows: 1fr 1fr 1fr;
  gap: 12px;
  align-items: center;
  justify-items: center;
}
.dpad-grid .center {
  width: 110px;
  height: 110px;
  background: radial-gradient(circle at 30% 25%, #ffe9ef 0%, #ffd5df 55%, #ffc8d9 75%, #ffb3d0 100%);
  box-shadow:
    0 14px 28px rgba(255, 102, 163, 0.18),
    inset 0 2px 0 rgba(255, 255, 255, 0.85);
}

.danger {
  border: 1px solid rgba(255, 99, 132, 0.25);
  background: linear-gradient(135deg, rgba(255, 235, 240, 0.96), rgba(255, 255, 255, 0.9));
  color: #7a1331;
}

.mini-btn {
  width: 64px;
  border: 0;
  border-radius: 12px;
  padding: 6px 10px;
  background: linear-gradient(135deg, #ff99cc 0%, #ff66a3 100%);
  color: #fff;
  font-weight: 700;
  cursor: pointer;
}

.list-panel {
  border-radius: 24px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.65);
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

.panel-title {
  font-size: 16px;
  font-weight: 800;
  color: #d63384;
  margin-bottom: 12px;
}

.panel-loading,
.panel-empty {
  padding: 16px 0;
  color: rgba(44, 18, 48, 0.65);
}

.panel-table {
  border-radius: 16px;
  overflow: hidden;
  border: 1px solid rgba(255, 102, 163, 0.15);
}

.table-row {
  display: grid;
  grid-template-columns: 1fr 180px 90px;
  gap: 0;
  align-items: center;
  background: rgba(255, 255, 255, 0.85);
}

.table-row.header {
  background: rgba(255, 102, 163, 0.12);
  font-weight: 800;
}

.col {
  padding: 10px 12px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
}

.table-row:last-child .col {
  border-bottom: 0;
}

.col.key {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
}

.mini-btn {
  width: 64px;
  border: 0;
  border-radius: 12px;
  padding: 6px 10px;
  background: linear-gradient(135deg, #ff99cc 0%, #ff66a3 100%);
  color: #fff;
  font-weight: 700;
  cursor: pointer;
}

.mini-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

@media (max-width: 980px) {
  .content { justify-content: center; }
  .page-header {
    grid-template-columns: 120px 1fr 120px;
  }
}

@media (max-width: 576px) {
  .hotkey-page {
    padding: 14px;
  }
  .page-header {
    grid-template-columns: 1fr;
  }
  .nav-btn {
    width: 100%;
  }
  .table-row { grid-template-columns: 1fr 120px 80px; }
}
</style>
