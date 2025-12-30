<template>
  <div class="sakura-container" role="application" aria-label="LogAnalysis">
    <div class="bg-layer" aria-hidden="true">
      <div class="sakura-petal" v-for="i in 12" :key="'p'+i" :style="getParticleStyle(i)"></div>
      <div class="magic-circle"></div>
    </div>

    <div class="main-card animate-pop-in" :class="{ 'is-loading': loading }">
      <div class="header-section">
        <div class="badge-decoration">LOG ANALYZER</div>
        <h1 class="main-title">
          <span class="icon-l">🌸</span>
          魔法日志解析仪
          <span class="icon-r">✨</span>
        </h1>
        <p class="subtitle">~ 收集掉落数据的奇妙旅程 ~</p>
      </div>

      <div class="control-panel">
        <div class="input-group">
          <label class="input-label" for="logSelect">
            <span class="label-icon">📜</span> 选择契约书 (日志)
          </label>
          
          <div class="select-wrapper">
            <select
              id="logSelect"
              v-model="selectedFile"
              @change="onFileChange"
              class="magic-select"
              :disabled="loading && !selectedFile"
            >
              <option value="" disabled>请选择日志文件...</option>
              <option v-for="file in logFiles" :key="file" :value="file">
                {{ file }}
              </option>
            </select>
            <span class="select-arrow">▼</span>
          </div>
        </div>

        <div class="status-bar">
          <div class="status-pill" :class="chipClass">
            <span class="status-dot"></span>
            {{ chipText }}
          </div>
        </div>
      </div>

      <div class="display-area">
        <div v-if="loading" class="state-box loading-state">
          <div class="spinner-heart"></div>
          <p>正在以此方世界的魔力解析中...</p>
        </div>

        <div v-else-if="error" class="state-box error-state">
          <div class="error-icon">⚡</div>
          <h3>解析法阵失效</h3>
          <p>{{ error }}</p>
          <button class="retry-btn" @click="onFileChange">重新咏唱</button>
        </div>

        <div v-else-if="!selectedFile" class="state-box empty-state">
          <div class="empty-icon">📫</div>
          <p>请先在上方选择一份日志契约</p>
        </div>

        <div v-else class="data-grid">
          <div 
            v-for="([key, value], index) in sortedData" 
            :key="key"
            class="item-card"
            :class="{
              'rank-1': index === 0,
              'rank-2': index === 1,
              'rank-3': index === 2
            }"
            :style="{ animationDelay: `${index * 0.05}s` }"
          >
            <div class="rank-badge">NO.{{ index + 1 }}</div>
            
            <div class="card-content">
              <div class="info-col name-col">
                <span class="item-icon" v-if="index < 3">{{ getMedal(index) }}</span>
                <span class="item-name" :title="key">{{ key }}</span>
              </div>

              <div class="divider-line">
                <span class="line-deco"></span>
              </div>

              <div class="info-col count-col">
                <span class="count-label">获得</span>
                <span class="count-num">{{ value }}</span>
                <span class="count-unit">个</span>
              </div>

              <div class="action-col">
                <button 
                  class="focus-btn" 
                  @click="addToFocus(key)"
                  :disabled="addingItem === key"
                  title="添加到关注列表"
                >
                  {{ addingItem === key ? '⏳' : '⭐' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="card-footer">
        <router-link to="/" class="home-btn" title="返回主城">
          🏠
        </router-link>
      </div>
    </div>
  </div>
</template>

<script>
// 逻辑部分完全保持原样，只字未改
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
    const addingItem = ref('')

    const sortedData = computed(() => {
      return Object.entries(analysisData.value).sort((a, b) => b[1] - a[1])
    })

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

    const loadLogFiles = async () => {
      try {
        const response = await apiMethods.getLogFiles()
        logFiles.value = response.files || []

        if (logFiles.value.length > 0) {
          selectedFile.value = logFiles.value[0]
          await loadAnalysisData()
        }
      } catch (err) {
        error.value = '加载日志文件列表失败：' + err.message
      }
    }

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

    const onFileChange = () => {
      loadAnalysisData()
    }

    const addToFocus = async (materialName) => {
      if (!materialName) return
      
      addingItem.value = materialName
      try {
        await apiMethods.addBagStatistics(materialName)
        // 可以添加成功提示
        alert(`材料 "${materialName}" 已添加到关注列表`)
      } catch (err) {
        alert(`添加失败：${err.message}`)
      } finally {
        addingItem.value = ''
      }
    }

    const getRankClass = (index) => {
      if (index === 0) return 'first'
      if (index === 1) return 'second'
      if (index === 2) return 'third'
      return 'normal'
    }

    const getMedal = (index) => {
      const medals = ['👑', '💎', '🌟'] // 稍微换了一下图标更符合风格
      return medals[index] || ''
    }

    const getParticleStyle = (index) => {
      // 保留原有逻辑，CSS中会重新利用这些值
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

    // 虽然模板里没用到getStarStyle了，但为了不报错保留它
    const getStarStyle = (index) => ({})

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
      getStarStyle,
      addToFocus,
      addingItem
    }
  }
}
</script>

<style scoped>
/* =========================================
   全局主题变量 - 魔法少女粉色系
   ========================================= */
.sakura-container {
  --primary-pink: #ff9ecd;
  --dark-pink: #ff7eb9;
  --light-pink: #ffe6f2;
  --bg-gradient-start: #fff0f5;
  --bg-gradient-end: #ffe4e1;
  --text-main: #6b4c5e;
  --text-sub: #9e8592;
  --white: #ffffff;
  --border-radius: 20px;
  --card-shadow: 0 8px 32px rgba(255, 158, 205, 0.25);
  
  min-height: 100vh;
  background: linear-gradient(135deg, var(--bg-gradient-start) 0%, var(--bg-gradient-end) 100%);
  font-family: "Varela Round", "Microsoft YaHei", "Rounded Mplus 1c", sans-serif;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
  position: relative;
  overflow-x: hidden;
  color: var(--text-main);
}

/* =========================================
   背景装饰 - 樱花与魔法阵
   ========================================= */
.bg-layer {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  overflow: hidden;
}

.magic-circle {
  position: absolute;
  width: 600px;
  height: 600px;
  border: 2px dashed rgba(255, 255, 255, 0.6);
  border-radius: 50%;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation: spin 60s linear infinite;
  opacity: 0.5;
}

.sakura-petal {
  position: absolute;
  width: 16px;
  height: 16px;
  background: #ffb7d5;
  border-radius: 16px 0 16px 0;
  opacity: 0.6;
  animation: float 10s ease-in-out infinite;
  box-shadow: 0 0 10px rgba(255, 183, 213, 0.5);
}

@keyframes spin { 100% { transform: translate(-50%, -50%) rotate(360deg); } }
@keyframes float {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-30px) rotate(45deg); }
}

/* =========================================
   主卡片容器
   ========================================= */
.main-card {
  position: relative;
  z-index: 10;
  width: 100%;
  max-width: 800px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(20px);
  border-radius: 30px;
  border: 4px solid var(--white);
  box-shadow: var(--card-shadow), inset 0 0 0 2px rgba(255, 158, 205, 0.3);
  padding: 40px;
  display: flex;
  flex-direction: column;
}

.animate-pop-in {
  animation: popIn 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes popIn {
  from { transform: scale(0.9) translateY(30px); opacity: 0; }
  to { transform: scale(1) translateY(0); opacity: 1; }
}

/* =========================================
   Header 区域
   ========================================= */
.header-section {
  text-align: center;
  margin-bottom: 30px;
}

.badge-decoration {
  display: inline-block;
  background: var(--dark-pink);
  color: white;
  font-size: 10px;
  padding: 4px 12px;
  border-radius: 20px;
  letter-spacing: 2px;
  margin-bottom: 10px;
  font-weight: bold;
}

.main-title {
  font-size: 28px;
  color: var(--text-main);
  margin: 0;
  text-shadow: 2px 2px 0px var(--light-pink);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.subtitle {
  font-size: 14px;
  color: var(--text-sub);
  margin-top: 8px;
}

/* =========================================
   控制面板 (Select)
   ========================================= */
.control-panel {
  background: var(--light-pink);
  padding: 20px;
  border-radius: 20px;
  border: 2px dashed rgba(255, 158, 205, 0.5);
  margin-bottom: 25px;
}

.input-group {
  margin-bottom: 15px;
}

.input-label {
  display: block;
  font-size: 14px;
  font-weight: bold;
  margin-bottom: 8px;
  color: var(--text-main);
}

.select-wrapper {
  position: relative;
}

.magic-select {
  width: 100%;
  padding: 12px 16px;
  font-size: 14px;
  border-radius: 12px;
  border: 2px solid white;
  background: white;
  color: var(--text-main);
  outline: none;
  appearance: none;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 4px 10px rgba(255, 158, 205, 0.1);
}

.magic-select:focus {
  border-color: var(--primary-pink);
  box-shadow: 0 0 0 4px rgba(255, 158, 205, 0.2);
}

.select-arrow {
  position: absolute;
  right: 15px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--dark-pink);
  pointer-events: none;
  font-size: 12px;
}

/* 状态条 */
.status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: bold;
  background: white;
  border: 1px solid rgba(0,0,0,0.05);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.chip--ok .status-dot { background: #4cd964; box-shadow: 0 0 8px #4cd964; }
.chip--loading .status-dot { background: #ffcc00; animation: pulse 1s infinite; }
.chip--error .status-dot { background: #ff3b30; }
.chip--idle .status-dot { background: #ccc; }

.status-hint {
  font-size: 12px;
  color: var(--text-sub);
  font-style: italic;
}

/* =========================================
   数据展示区域 (卡片列表)
   ========================================= */
.display-area {
  min-height: 200px;
}

.data-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 核心：列表项卡片样式 */
.item-card {
  background: white;
  border: 2px solid var(--primary-pink);
  border-radius: 16px;
  padding: 0;
  position: relative;
  transition: all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  display: flex;
  align-items: stretch;
  overflow: hidden;
  animation: slideUp 0.5s backwards;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.item-card:hover {
  transform: translateY(-4px) scale(1.01);
  box-shadow: 0 10px 20px rgba(255, 158, 205, 0.3);
}

.rank-badge {
  background: var(--primary-pink);
  color: white;
  font-size: 12px;
  font-weight: bold;
  writing-mode: vertical-lr; /* 竖排文字 */
  padding: 10px 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  letter-spacing: 2px;
}

.card-content {
  flex: 1;
  display: flex;
  align-items: center;
  padding: 12px 16px;
  gap: 10px;
}

/* 前三名特殊样式 */
.rank-1 { border-color: #ffd700; background: linear-gradient(to right, #fffdf0, #fff); }
.rank-1 .rank-badge { background: #ffd700; color: #b38b00; }

.rank-2 { border-color: #c0c0c0; }
.rank-2 .rank-badge { background: #c0c0c0; }

.rank-3 { border-color: #cd7f32; }
.rank-3 .rank-badge { background: #cd7f32; }

/* 布局列 */
.info-col {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.name-col {
  flex: 1;
  align-items: flex-start;
  min-width: 0;
}

.item-name {
  font-weight: bold;
  font-size: 15px;
  color: var(--text-main);
  word-break: break-all;
  line-height: 1.4;
}

.item-icon {
  font-size: 12px;
  margin-bottom: 2px;
}

/* 关键：中间的横杠/分割线 */
.divider-line {
  flex: 0 0 40px; /* 宽度 */
  display: flex;
  align-items: center;
  justify-content: center;
}

.line-deco {
  width: 100%;
  height: 2px;
  background-image: linear-gradient(to right, transparent 0%, var(--primary-pink) 30%, var(--primary-pink) 70%, transparent 100%);
  opacity: 0.5;
}

.count-col {
  align-items: flex-end;
  min-width: 60px;
}

.count-label {
  font-size: 10px;
  color: var(--text-sub);
  margin-bottom: 2px;
}

.count-num {
  font-size: 20px;
  font-weight: 900;
  color: var(--dark-pink);
  font-family: "Arial Rounded MT Bold", sans-serif;
  line-height: 1;
}

.count-unit {
  font-size: 10px;
  color: var(--text-sub);
}

.action-col {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-left: 10px;
}

.focus-btn {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: 2px solid var(--primary-pink);
  background: white;
  color: var(--primary-pink);
  font-size: 16px;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
}

.focus-btn:hover:not(:disabled) {
  background: var(--primary-pink);
  transform: scale(1.1) rotate(15deg);
}

.focus-btn:active:not(:disabled) {
  transform: scale(0.95);
}

.focus-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* =========================================
   其他状态 (Loading, Error, Empty)
   ========================================= */
.state-box {
  text-align: center;
  padding: 40px 0;
  color: var(--text-sub);
  background: white;
  border-radius: 16px;
  border: 2px dashed rgba(0,0,0,0.1);
}

.spinner-heart {
  width: 40px;
  height: 40px;
  background: var(--primary-pink);
  margin: 0 auto 20px;
  transform: rotate(45deg);
  animation: heartBeat 1.2s infinite;
}
.spinner-heart::before,
.spinner-heart::after {
  content: "";
  position: absolute;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--primary-pink);
}
.spinner-heart::before { left: -20px; }
.spinner-heart::after { top: -20px; }

@keyframes heartBeat {
  0% { transform: rotate(45deg) scale(0.8); }
  5% { transform: rotate(45deg) scale(0.9); }
  10% { transform: rotate(45deg) scale(0.8); }
  15% { transform: rotate(45deg) scale(1); }
  50% { transform: rotate(45deg) scale(0.8); }
  100% { transform: rotate(45deg) scale(0.8); }
}

.empty-icon, .error-icon {
  font-size: 40px;
  margin-bottom: 10px;
}

.retry-btn {
  margin-top: 15px;
  background: var(--primary-pink);
  color: white;
  border: none;
  padding: 8px 20px;
  border-radius: 20px;
  cursor: pointer;
  font-weight: bold;
  transition: transform 0.2s;
}
.retry-btn:hover { transform: scale(1.05); }

/* =========================================
   底部
   ========================================= */
.card-footer {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

.home-btn {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: white;
  border: 2px solid var(--primary-pink);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  text-decoration: none;
  transition: all 0.3s;
  box-shadow: 0 4px 10px rgba(0,0,0,0.1);
}

.home-btn:hover {
  background: var(--primary-pink);
  transform: rotate(360deg);
}

/* =========================================
   响应式适配 (手机端)
   ========================================= */
@media (max-width: 600px) {
  .sakura-container {
    padding: 10px;
  }
  
  .main-card {
    padding: 20px 15px;
    border-width: 2px;
  }
  
  .main-title { font-size: 22px; }
  .control-panel { padding: 15px; }
  
  /* 手机端列表卡片调整 */
  .item-card {
    border-radius: 12px;
  }
  
  .rank-badge {
    padding: 0 4px;
    font-size: 10px;
  }
  
  .card-content {
    padding: 10px 12px;
    gap: 5px;
  }
  
  .item-name { font-size: 14px; }
  .count-num { font-size: 18px; }
  
  .divider-line {
    flex: 0 0 20px;
  }
}
</style>