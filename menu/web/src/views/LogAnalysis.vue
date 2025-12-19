<template>
  <div class="container" role="application" aria-label="LogAnalysis">
    <!-- 背景装饰（不吃点击） -->
    <div class="bg-decoration" aria-hidden="true">
      <div
        class="floating-particle"
        v-for="i in 12"
        :key="i"
        :style="getParticleStyle(i)"
      ></div>
      <div
        class="floating-star"
        v-for="i in 6"
        :key="'star' + i"
        :style="getStarStyle(i)"
      >
        ★
      </div>
      <div class="bg-grid"></div>
      <div class="bg-glow bg-glow--pink"></div>
      <div class="bg-glow bg-glow--blue"></div>
    </div>

    <div class="card glass" :class="{ 'is-loading': loading }">
      <!-- 标题区域 -->
      <div class="header">
        <div class="title">
          <span class="icon" aria-hidden="true">🌸</span>
          <span class="main-title">采集数据</span>
          <span class="icon" aria-hidden="true">✨</span>
        </div>
        <div class="subtitle">「召唤日志之魂，解析掉落之语」</div>
        <div class="divider"></div>
      </div>

      <!-- 文件选择区域 -->
      <div class="select-container">
        <label class="select-label" for="logSelect">
          <span class="label-icon" aria-hidden="true">📁</span>
          选择日志文件：
        </label>

        <div class="select-wrapper">
          <select
            id="logSelect"
            v-model="selectedFile"
            @change="onFileChange"
            class="select-box"
            :disabled="loading && !selectedFile"
            aria-label="选择日志文件"
          >
            <option value="" disabled>请选择日志文件</option>
            <option v-for="file in logFiles" :key="file" :value="file">
              {{ file }}
            </option>
          </select>

          <span class="select-arrow" aria-hidden="true">▾</span>

          <!-- 右侧状态徽章 -->
          <span class="chip" :class="chipClass">
            <span class="chip-dot" aria-hidden="true"></span>
            <span class="chip-text">{{ chipText }}</span>
          </span>
        </div>

        <div class="hint">
          <span class="hint-dot" aria-hidden="true"></span>
          <span>
            排名按数量自动降序排列，前三名会点亮「传说光效」。
          </span>
        </div>
      </div>

      <!-- 内容区域 -->
      <div class="content-area">
        <!-- 加载状态 -->
        <div v-if="loading" class="loading-container" role="status" aria-live="polite">
          <div class="loading-orb">
            <div class="orb-ring"></div>
            <div class="orb-core"></div>
          </div>
          <div class="loading-text">
            <div class="loading-title">正在解析「{{ selectedFile || '未知卷轴' }}」…</div>
            <div class="loading-sub">请稍候，结界正在计算掉落概率（其实是在请求接口）</div>
          </div>
        </div>

        <!-- 错误状态 -->
        <div v-else-if="error" class="error-container" role="alert">
          <div class="error-top">
            <span class="error-icon" aria-hidden="true">⚠️</span>
            <div class="error-title">结界崩坏：数据加载失败</div>
          </div>
          <div class="error-text">{{ error }}</div>
          <div class="error-actions">
            <button class="btn btn-primary" @click="onFileChange">
              <span aria-hidden="true">🔁</span>
              重新召唤
            </button>
          </div>
        </div>

        <!-- 空状态 -->
        <div v-else-if="!selectedFile" class="empty-state">
          <div class="empty-illu" aria-hidden="true">
            <div class="halo"></div>
            <div class="book">📜</div>
          </div>
          <div class="empty-text">
            <div class="empty-title">尚未绑定日志卷轴</div>
            <div class="empty-sub">请先在上方选择一个文件开始分析</div>
          </div>
        </div>

        <!-- 数据列表 -->
        <div v-else class="data-list" role="table" aria-label="分析结果列表">
          <div class="list-header" role="row">
            <span class="rank-header" role="columnheader">排名</span>
            <span class="item-header" role="columnheader">材料名称</span>
            <span class="count-header" role="columnheader">收获数量</span>
          </div>

          <div class="list-body" role="rowgroup">
            <div
              v-for="([key, value], index) in sortedData"
              :key="key"
              class="data-item"
              :class="{
                'top-item': index < 3,
                'top-1': index === 0,
                'top-2': index === 1,
                'top-3': index === 2
              }"
              :style="{ animationDelay: `${Math.min(index, 15) * 0.06}s` }"
              role="row"
            >
              <div class="rank" role="cell">
                <span class="rank-number" :class="getRankClass(index)">
                  {{ index + 1 }}
                </span>
                <span v-if="index < 3" class="medal" aria-hidden="true">
                  {{ getMedal(index) }}
                </span>
              </div>

              <div class="item-name" role="cell" :title="key">
                <span class="name-glow" aria-hidden="true"></span>
                <span class="name-text">{{ key }}</span>
              </div>

              <div class="item-count" role="cell">
                <span class="count-value">{{ value }}</span>
                <span class="count-unit">个</span>
              </div>

              <!-- 右侧高光线 -->
              <div class="shine" aria-hidden="true"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- 悬浮返回桌面按钮 -->
      <router-link to="/" class="floating-back-btn" title="返回桌面" aria-label="返回桌面">
        <span class="floating-btn-icon" aria-hidden="true">🏠</span>
      </router-link>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { apiMethods } from '../utils/api.js'

export default {
  name: 'LogAnalysis',
  setup() {
    const logFiles = ref([])
    const selectedFile = ref('')
    const analysisData = ref({})
    const loading = ref(false)
    const error = ref('')

    // 计算排序后的数据（逻辑不改）
    const sortedData = computed(() => {
      return Object.entries(analysisData.value).sort((a, b) => b[1] - a[1])
    })

    // 状态chip（纯展示，不影响逻辑）
    const chipText = computed(() => {
      if (loading.value) return '解析中'
      if (error.value) return '异常'
      if (!selectedFile.value) return '待选择'
      return '就绪'
    })

    const chipClass = computed(() => {
      if (loading.value) return 'chip--loading'
      if (error.value) return 'chip--error'
      if (!selectedFile.value) return 'chip--idle'
      return 'chip--ok'
    })

    // 加载日志文件列表（逻辑不改）
    const loadLogFiles = async () => {
      try {
        const response = await apiMethods.getLogFiles()
        logFiles.value = response.files || []

        // 如果有文件，自动选择第一个（逻辑不改）
        if (logFiles.value.length > 0) {
          selectedFile.value = logFiles.value[0]
          await loadAnalysisData()
        }
      } catch (err) {
        error.value = '加载日志文件列表失败：' + err.message
      }
    }

    // 加载分析数据（逻辑不改）
    const loadAnalysisData = async () => {
      if (!selectedFile.value) {
        analysisData.value = {}
        return
      }

      loading.value = true
      error.value = ''

      try {
        const data = await apiMethods.getLogAnalysis(selectedFile.value)
        analysisData.value = data
      } catch (err) {
        error.value = '加载失败：' + err.message
        analysisData.value = {}
      } finally {
        loading.value = false
      }
    }

    // 文件选择变化处理（逻辑不改）
    const onFileChange = () => {
      loadAnalysisData()
    }

    // 获取排名样式类名（逻辑不改）
    const getRankClass = (index) => {
      if (index === 0) return 'first'
      if (index === 1) return 'second'
      if (index === 2) return 'third'
      return 'normal'
    }

    // 获取奖牌图标（逻辑不改）
    const getMedal = (index) => {
      const medals = ['🥇', '🥈', '🥉']
      return medals[index] || ''
    }

    // 生成粒子样式（逻辑不改）
    const getParticleStyle = (index) => {
      const positions = [
        { left: '10%', top: '20%', animationDelay: '0s' },
        { left: '85%', top: '15%', animationDelay: '2s' },
        { left: '20%', top: '80%', animationDelay: '4s' },
        { left: '80%', top: '70%', animationDelay: '1s' },
        { left: '50%', top: '10%', animationDelay: '3s' },
        { left: '15%', top: '50%', animationDelay: '5s' }
      ]
      return positions[index - 1] || { left: '50%', top: '50%' }
    }

    // 生成星星样式（逻辑不改）
    const getStarStyle = (index) => {
      const positions = [
        { left: '5%', top: '10%', animationDelay: '0s' },
        { left: '90%', top: '5%', animationDelay: '2s' },
        { left: '15%', top: '85%', animationDelay: '4s' },
        { left: '85%', top: '75%', animationDelay: '1s' },
        { left: '55%', top: '5%', animationDelay: '3s' },
        { left: '10%', top: '60%', animationDelay: '5s' }
      ]
      return positions[index - 1] || { left: '50%', top: '50%' }
    }

    // 组件挂载时加载数据（逻辑不改）
    onMounted(() => {
      loadLogFiles()
    })

    return {
      logFiles,
      selectedFile,
      analysisData,
      loading,
      error,
      sortedData,
      chipText,
      chipClass,
      onFileChange,
      getRankClass,
      getMedal,
      getParticleStyle,
      getStarStyle
    }
  }
}
</script>

<style scoped>
/* ======================
   霓虹二次元中二主题变量
====================== */
.container {
  --bg1: #120a24;
  --bg2: #1b0f33;
  --pink: #ff69c9;
  --pink2: #ffb3e6;
  --violet: #b389ff;
  --cyan: #58e6ff;
  --text: rgba(255, 255, 255, 0.92);
  --muted: rgba(255, 255, 255, 0.72);
  --card: rgba(20, 10, 40, 0.58);
  --card2: rgba(255, 255, 255, 0.06);
  --stroke: rgba(255, 105, 201, 0.42);
  --stroke2: rgba(88, 230, 255, 0.24);
  --shadow: 0 24px 64px rgba(0, 0, 0, 0.42);
  --shadow2: 0 10px 26px rgba(255, 105, 201, 0.12);
  --radius: 28px;

  min-height: 100vh;
  background:
    radial-gradient(1200px 700px at 20% 10%, rgba(255, 105, 201, 0.16), transparent 55%),
    radial-gradient(900px 540px at 80% 20%, rgba(88, 230, 255, 0.14), transparent 55%),
    linear-gradient(160deg, var(--bg1), var(--bg2));
  font-family: "Segoe UI", "Microsoft YaHei", system-ui, -apple-system, sans-serif;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 22px;
  position: relative;
  overflow-x: hidden;
  color: var(--text);
}

/* 背景装饰层 */
.bg-decoration {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}

.bg-grid {
  position: absolute;
  inset: -20%;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.06) 1px, transparent 1px);
  background-size: 56px 56px;
  transform: rotate(-8deg);
  opacity: 0.18;
  filter: blur(0.2px);
}

.bg-glow {
  position: absolute;
  width: 520px;
  height: 520px;
  border-radius: 50%;
  filter: blur(42px);
  opacity: 0.38;
}
.bg-glow--pink {
  left: -120px;
  top: 10%;
  background: radial-gradient(circle, rgba(255, 105, 201, 0.8), transparent 65%);
}
.bg-glow--blue {
  right: -160px;
  bottom: 6%;
  background: radial-gradient(circle, rgba(88, 230, 255, 0.7), transparent 65%);
}

.floating-particle {
  position: absolute;
  width: 12px;
  height: 12px;
  background: radial-gradient(circle, rgba(255, 105, 201, 0.9) 40%, rgba(255, 255, 255, 0) 70%);
  border-radius: 50%;
  animation: float 8s ease-in-out infinite;
  opacity: 0.38;
  filter: blur(1.2px);
}

.floating-star {
  position: absolute;
  font-size: 18px;
  color: rgba(255, 105, 201, 0.9);
  opacity: 0.18;
  animation: starFloat 7s infinite alternate;
  text-shadow: 0 0 18px rgba(255, 105, 201, 0.35);
}

@keyframes float {
  0%, 100% { transform: translateY(0px) rotate(0deg); }
  50% { transform: translateY(-26px) rotate(160deg); }
}
@keyframes starFloat {
  0% { transform: scale(1) rotate(-8deg); }
  100% { transform: scale(1.25) rotate(18deg); }
}

/* ======================
   玻璃卡片主体
====================== */
.card.glass {
  width: min(920px, 96vw);
  position: relative;
  z-index: 10;

  background: linear-gradient(135deg, rgba(255, 255, 255, 0.09), rgba(255, 255, 255, 0.04));
  border-radius: var(--radius);
  border: 1px solid rgba(255, 105, 201, 0.28);
  box-shadow: var(--shadow), var(--shadow2);
  backdrop-filter: blur(18px) saturate(1.15);
  overflow: hidden;
  padding: 34px 30px 26px;

  animation: slideIn 0.7s cubic-bezier(.4,0,.2,1);
}

.card.glass::before {
  content: "";
  position: absolute;
  inset: -2px;
  background: conic-gradient(
    from 180deg,
    rgba(255,105,201,0.55),
    rgba(88,230,255,0.35),
    rgba(179,137,255,0.38),
    rgba(255,105,201,0.55)
  );
  filter: blur(24px);
  opacity: 0.22;
  z-index: 0;
}

.card.glass::after {
  content: "";
  position: absolute;
  inset: 0;
  background:
    radial-gradient(1200px 220px at 50% 0%, rgba(255,255,255,0.14), transparent 60%),
    linear-gradient(180deg, rgba(255,255,255,0.06), transparent 36%);
  z-index: 0;
}

.card.glass > * {
  position: relative;
  z-index: 1;
}

@keyframes slideIn {
  from { opacity: 0; transform: translateY(26px) scale(0.98); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

/* ======================
   Header
====================== */
.header {
  text-align: center;
  margin-bottom: 18px;
}

.title {
  font-size: 34px;
  font-weight: 900;
  letter-spacing: 1px;
  color: rgba(255, 255, 255, 0.96);
  text-shadow:
    0 0 22px rgba(255, 105, 201, 0.25),
    0 0 32px rgba(88, 230, 255, 0.18);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.title .icon {
  font-size: 30px;
  animation: bounce 2s infinite;
  filter: drop-shadow(0 6px 18px rgba(255, 105, 201, 0.18));
}

@keyframes bounce {
  0%, 20%, 50%, 80%, 100% { transform: translateY(0); }
  40% { transform: translateY(-10px); }
  60% { transform: translateY(-5px); }
}

.subtitle {
  margin-top: 6px;
  font-size: 14px;
  color: var(--muted);
  opacity: 0.92;
}

.divider {
  width: 96px;
  height: 10px;
  margin: 16px auto 0;
  border-radius: 999px;
  background: linear-gradient(90deg, rgba(255,105,201,0.85), rgba(88,230,255,0.75));
  box-shadow:
    0 8px 22px rgba(255, 105, 201, 0.14),
    0 8px 22px rgba(88, 230, 255, 0.10);
}

/* ======================
   Select + hint + chip
====================== */
.select-container {
  margin-top: 8px;
  margin-bottom: 18px;
}

.select-label {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  color: rgba(255,255,255,0.86);
  font-weight: 700;
  margin-bottom: 12px;
}

.label-icon {
  font-size: 18px;
  filter: drop-shadow(0 6px 16px rgba(255, 105, 201, 0.18));
}

.select-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  max-width: 720px;
}

.select-box {
  flex: 1;
  font-size: 15px;
  padding: 14px 44px 14px 16px;
  border-radius: 16px;
  border: 1px solid rgba(255, 105, 201, 0.35);
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.10), rgba(255, 255, 255, 0.05));
  color: rgba(255,255,255,0.92);
  cursor: pointer;
  transition: transform 0.25s ease, box-shadow 0.25s ease, border-color 0.25s ease;
  appearance: none;
  box-shadow: 0 10px 24px rgba(0,0,0,0.22);
  outline: none;
}

.select-box option {
  color: #111;
}

.select-box:focus {
  border-color: rgba(88, 230, 255, 0.58);
  box-shadow:
    0 0 0 3px rgba(88,230,255,0.14),
    0 16px 36px rgba(255,105,201,0.10);
  transform: translateY(-1px);
}

.select-arrow {
  position: absolute;
  right: 104px;
  top: 50%;
  transform: translateY(-50%);
  color: rgba(255,255,255,0.78);
  pointer-events: none;
  transition: transform 0.25s ease;
  text-shadow: 0 0 18px rgba(88, 230, 255, 0.18);
}

.select-wrapper:hover .select-arrow {
  transform: translateY(-50%) scale(1.12);
}

/* chip */
.chip {
  flex: none;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 999px;
  border: 1px solid rgba(255,255,255,0.14);
  background: rgba(255,255,255,0.06);
  backdrop-filter: blur(10px);
  color: rgba(255,255,255,0.86);
  box-shadow: 0 10px 22px rgba(0,0,0,0.18);
  user-select: none;
  white-space: nowrap;
}
.chip-dot {
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: rgba(255,255,255,0.75);
  box-shadow: 0 0 18px rgba(255,255,255,0.18);
}
.chip--ok { border-color: rgba(88,230,255,0.28); }
.chip--ok .chip-dot { background: rgba(88,230,255,0.95); box-shadow: 0 0 18px rgba(88,230,255,0.30); }

.chip--idle { border-color: rgba(255,105,201,0.22); }
.chip--idle .chip-dot { background: rgba(255,105,201,0.9); box-shadow: 0 0 18px rgba(255,105,201,0.28); }

.chip--error { border-color: rgba(255, 99, 99, 0.35); }
.chip--error .chip-dot { background: rgba(255, 99, 99, 0.92); box-shadow: 0 0 18px rgba(255, 99, 99, 0.28); }

.chip--loading { border-color: rgba(179,137,255,0.35); }
.chip--loading .chip-dot {
  background: rgba(179,137,255,0.95);
  box-shadow: 0 0 18px rgba(179,137,255,0.28);
  animation: pulse 1.1s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 0.9; }
  50% { transform: scale(1.25); opacity: 1; }
}

/* hint */
.hint {
  margin-top: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
  color: rgba(255,255,255,0.72);
  font-size: 13px;
}
.hint-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(255,105,201,0.8);
  box-shadow: 0 0 16px rgba(255,105,201,0.24);
}

/* ======================
   Content area
====================== */
.content-area {
  min-height: 220px;
}

/* Loading */
.loading-container {
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 28px;
  border-radius: 18px;
  border: 1px solid rgba(255,255,255,0.12);
  background: rgba(255,255,255,0.05);
  box-shadow: 0 14px 30px rgba(0,0,0,0.22);
}

.loading-orb {
  width: 56px;
  height: 56px;
  position: relative;
  flex: none;
}
.orb-ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  border: 3px solid rgba(88,230,255,0.32);
  border-top-color: rgba(255,105,201,0.75);
  animation: spin 1.0s linear infinite;
  filter: drop-shadow(0 10px 22px rgba(255,105,201,0.12));
}
.orb-core {
  position: absolute;
  inset: 10px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(255,105,201,0.55), rgba(88,230,255,0.22), transparent 70%);
  box-shadow: inset 0 0 18px rgba(255,255,255,0.10);
}
@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
.loading-title {
  font-weight: 800;
  letter-spacing: 0.5px;
}
.loading-sub {
  margin-top: 6px;
  color: rgba(255,255,255,0.70);
  font-size: 13px;
}

/* Error */
.error-container {
  padding: 22px;
  border-radius: 18px;
  border: 1px solid rgba(255, 99, 99, 0.25);
  background:
    linear-gradient(135deg, rgba(255, 99, 99, 0.08), rgba(255, 255, 255, 0.03));
  box-shadow: 0 14px 30px rgba(0,0,0,0.22);
}
.error-top {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}
.error-icon {
  font-size: 22px;
}
.error-title {
  font-weight: 900;
  letter-spacing: 0.4px;
}
.error-text {
  color: rgba(255,255,255,0.76);
  line-height: 1.6;
}
.error-actions {
  margin-top: 14px;
  display: flex;
  gap: 10px;
}

/* Buttons */
.btn {
  appearance: none;
  border: none;
  cursor: pointer;
  border-radius: 14px;
  padding: 12px 14px;
  font-weight: 800;
  letter-spacing: 0.4px;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}
.btn:active { transform: translateY(1px) scale(0.99); }
.btn-primary {
  color: rgba(255,255,255,0.92);
  background: linear-gradient(135deg, rgba(255,105,201,0.85), rgba(88,230,255,0.62));
  box-shadow: 0 14px 28px rgba(255,105,201,0.12), 0 14px 28px rgba(88,230,255,0.10);
}
.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 18px 34px rgba(255,105,201,0.14), 0 18px 34px rgba(88,230,255,0.12);
}

/* Empty */
.empty-state {
  padding: 28px;
  border-radius: 18px;
  border: 1px solid rgba(255,255,255,0.12);
  background: rgba(255,255,255,0.05);
  box-shadow: 0 14px 30px rgba(0,0,0,0.22);
  display: grid;
  grid-template-columns: 92px 1fr;
  gap: 18px;
  align-items: center;
}
.empty-illu {
  position: relative;
  width: 92px;
  height: 92px;
  display: grid;
  place-items: center;
}
.halo {
  position: absolute;
  inset: 6px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(255,105,201,0.35), rgba(88,230,255,0.16), transparent 70%);
  filter: blur(0.2px);
  animation: halo 2.8s ease-in-out infinite;
}
@keyframes halo {
  0%, 100% { transform: scale(1); opacity: 0.85; }
  50% { transform: scale(1.08); opacity: 1; }
}
.book {
  position: relative;
  font-size: 34px;
  filter: drop-shadow(0 12px 22px rgba(255,105,201,0.18));
}
.empty-title {
  font-weight: 900;
  letter-spacing: 0.4px;
  font-size: 16px;
}
.empty-sub {
  margin-top: 6px;
  color: rgba(255,255,255,0.72);
  font-size: 13px;
}

/* ======================
   Data list
====================== */
.data-list {
  border-radius: 18px;
  padding: 16px;
  border: 1px solid rgba(255,255,255,0.12);
  background: rgba(255,255,255,0.05);
  box-shadow: 0 14px 30px rgba(0,0,0,0.22);
}

.list-header {
  display: grid;
  grid-template-columns: 96px 1fr 120px;
  gap: 14px;
  padding: 12px 14px;
  border-radius: 14px;
  background: rgba(255,255,255,0.06);
  border: 1px dashed rgba(255, 105, 201, 0.26);
  color: rgba(255,255,255,0.80);
  font-weight: 900;
  letter-spacing: 0.5px;
}

.rank-header, .count-header { text-align: right; }
.item-header { text-align: left; }

.list-body {
  margin-top: 12px;
}

.data-item {
  display: grid;
  grid-template-columns: 96px 1fr 120px;
  gap: 14px;
  align-items: center;

  border-radius: 16px;
  padding: 14px 14px;
  margin-bottom: 10px;

  background: linear-gradient(135deg, rgba(255,255,255,0.09), rgba(255,255,255,0.04));
  border: 1px solid rgba(255,255,255,0.10);
  box-shadow: 0 10px 24px rgba(0,0,0,0.20);
  position: relative;
  overflow: hidden;

  animation: fadeInUp 0.5s ease-out both;
  transform-origin: center;
  transition: transform 0.22s ease, box-shadow 0.22s ease, border-color 0.22s ease;
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}

.data-item:hover {
  transform: translateY(-3px) scale(1.01);
  border-color: rgba(88, 230, 255, 0.24);
  box-shadow: 0 18px 40px rgba(0,0,0,0.28);
}

/* 右侧流光 */
.shine {
  position: absolute;
  top: -40%;
  right: -40%;
  width: 220px;
  height: 220px;
  background: radial-gradient(circle, rgba(88,230,255,0.20), transparent 60%);
  transform: rotate(25deg);
  opacity: 0.0;
  transition: opacity 0.28s ease;
}
.data-item:hover .shine { opacity: 1; }

/* top items */
.top-item {
  border-color: rgba(255,105,201,0.24);
  background: linear-gradient(135deg, rgba(255, 105, 201, 0.12), rgba(255,255,255,0.05));
}
.top-1 {
  box-shadow: 0 18px 44px rgba(255,105,201,0.16), 0 18px 44px rgba(88,230,255,0.10);
}
.top-2 {
  box-shadow: 0 18px 44px rgba(179,137,255,0.14);
}
.top-3 {
  box-shadow: 0 18px 44px rgba(255, 205, 120, 0.10);
}

/* rank */
.rank {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

.rank-number {
  width: 42px;
  height: 42px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 1000;
  letter-spacing: 0.2px;
  background: rgba(255,255,255,0.08);
  border: 1px solid rgba(255,255,255,0.12);
  box-shadow: inset 0 0 18px rgba(255,255,255,0.06);
}

.rank-number.first {
  background: linear-gradient(135deg, rgba(255, 231, 148, 0.28), rgba(255, 255, 255, 0.10));
  border-color: rgba(255, 231, 148, 0.35);
}
.rank-number.second {
  background: linear-gradient(135deg, rgba(220, 220, 220, 0.22), rgba(255, 255, 255, 0.10));
  border-color: rgba(220, 220, 220, 0.30);
}
.rank-number.third {
  background: linear-gradient(135deg, rgba(255,105,201,0.22), rgba(255, 255, 255, 0.10));
  border-color: rgba(255,105,201,0.30);
}
.rank-number.normal {
  background: rgba(255,255,255,0.07);
  border-color: rgba(255,255,255,0.10);
}

.medal {
  font-size: 20px;
  animation: rotate 3s linear infinite;
  filter: drop-shadow(0 10px 18px rgba(255, 105, 201, 0.14));
}
@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* item name */
.item-name {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0; /* 允许省略号 */
}
.name-glow {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: rgba(255,105,201,0.85);
  box-shadow: 0 0 18px rgba(255,105,201,0.25);
  flex: none;
}
.name-text {
  font-size: 16px;
  font-weight: 900;
  letter-spacing: 0.4px;
  color: rgba(255,255,255,0.92);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* count */
.item-count {
  display: flex;
  align-items: baseline;
  justify-content: flex-end;
  gap: 6px;
}
.count-value {
  font-size: 22px;
  font-weight: 1000;
  color: rgba(88,230,255,0.92);
  text-shadow: 0 0 18px rgba(88,230,255,0.18);
}
.count-unit {
  font-size: 13px;
  color: rgba(255,255,255,0.68);
}

/* ======================
   Floating Home Button
====================== */
.floating-back-btn {
  position: fixed !important;
  right: 22px;
  bottom: 22px;
  z-index: 9999;

  width: 56px;
  height: 56px;
  border-radius: 50%;
  display: grid;
  place-items: center;

  background: linear-gradient(135deg, rgba(255,105,201,0.85), rgba(88,230,255,0.62));
  border: 1px solid rgba(255,255,255,0.18);
  box-shadow: 0 18px 38px rgba(0,0,0,0.34);
  backdrop-filter: blur(10px);
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}
.floating-back-btn:hover {
  transform: translateY(-2px) scale(1.06);
  box-shadow: 0 24px 44px rgba(0,0,0,0.40);
}
.floating-btn-icon {
  font-size: 26px;
  filter: drop-shadow(0 10px 18px rgba(0,0,0,0.22));
}

/* ======================
   Responsive
====================== */
@media (max-width: 768px) {
  .container { padding: 14px; }
  .card.glass {
    width: 100%;
    padding: 18px 14px 16px;
    border-radius: 20px;
  }

  .title { font-size: 24px; gap: 8px; }
  .title .icon { font-size: 22px; }

  .select-wrapper { max-width: 100%; }
  .select-arrow { right: 104px; }

  .data-list { padding: 12px; }
  .list-header,
  .data-item {
    grid-template-columns: 62px 1fr 92px;
    gap: 10px;
    padding: 12px 12px;
  }
  .rank-number { width: 34px; height: 34px; }
  .name-text { font-size: 14px; }
  .count-value { font-size: 18px; }
}

@media (max-width: 480px) {
  .card.glass { padding: 14px 10px 12px; border-radius: 18px; }

  .subtitle { font-size: 12px; }

  .select-wrapper { gap: 8px; }
  .select-box { padding: 12px 42px 12px 12px; font-size: 14px; }
  .chip { padding: 9px 10px; }
  .select-arrow { right: 96px; }

  .loading-container {
    flex-direction: column;
    align-items: flex-start;
  }

  .list-header,
  .data-item {
    grid-template-columns: 56px 1fr 78px;
    gap: 8px;
    padding: 10px 10px;
    border-radius: 14px;
  }

  .rank { gap: 8px; }
  .medal { font-size: 18px; }
  .count-value { font-size: 16px; }
  .count-unit { font-size: 12px; }

  .floating-back-btn { right: 14px; bottom: 14px; width: 52px; height: 52px; }
}

/* 低动态偏好：减少动画 */
@media (prefers-reduced-motion: reduce) {
  * { animation: none !important; transition: none !important; }
}
</style>
