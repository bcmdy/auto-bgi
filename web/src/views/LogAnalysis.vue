<template>
  <div class="container">
    <div class="card glass">
      <!-- 标题区域 -->
      <div class="header">
        <div class="title">
          <span class="icon">🌸</span>
          <span class="main-title">收获前10的材料</span>
          <span class="icon">✨</span>
        </div>
        <!-- <div class="subtitle">二次元风格数据分析统计</div> -->
        <div class="divider"></div>
      </div>

      <!-- 文件选择区域 -->
      <div class="select-container">
        <label class="select-label">
          <span class="label-icon">📁</span>
          选择日志文件：
        </label>
        <div class="select-wrapper">
          <select v-model="selectedFile" @change="onFileChange" class="select-box">
            <option value="" disabled>请选择日志文件</option>
            <option v-for="file in logFiles" :key="file" :value="file">
              {{ file }}
            </option>
          </select>
          <span class="select-arrow">▼</span>
        </div>
      </div>

      <!-- 数据显示区域 -->
      <div class="content-area">
        <!-- 加载状态 -->
        <div v-if="loading" class="loading-container">
          <div class="loading-spinner"></div>
          <div class="loading-text">正在分析数据中...</div>
        </div>
        
        <!-- 错误状态 -->
        <div v-else-if="error" class="error-container">
          <span class="error-icon">⚠️</span>
          <div class="error-text">{{ error }}</div>
        </div>
        
        <!-- 空状态 -->
        <div v-else-if="!selectedFile" class="empty-state">
          <span class="empty-icon">📋</span>
          <div class="empty-text">请先选择日志文件开始分析</div>
        </div>
        
        <!-- 数据列表 -->
        <div v-else class="data-list">
          <div class="list-header">
            <span class="rank-header">排名</span>
            <span class="item-header">材料名称</span>
            <span class="count-header">收获数量</span>
          </div>
          <div
            v-for="([key, value], index) in sortedData"
            :key="key"
            class="data-item"
            :class="{ 'top-item': index < 3 }"
            :style="{ animationDelay: `${index * 0.1}s` }"
          >
            <div class="rank">
              <span class="rank-number" :class="getRankClass(index)">
                {{ index + 1 }}
              </span>
              <span v-if="index < 3" class="medal">{{ getMedal(index) }}</span>
            </div>
            <div class="item-name">{{ key }}</div>
            <div class="item-count">
              <span class="count-value">{{ value }}</span>
              <span class="count-unit">个</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 底部按钮 -->
      <div class="footer">
        <router-link to="/" class="back-button">
          <span class="button-icon">🏠</span>
          返回主页
        </router-link>
      </div>
    </div>
    
    <!-- 背景装饰 -->
    <div class="bg-decoration">
      <div class="floating-particle" v-for="i in 12" :key="i" :style="getParticleStyle(i)"></div>
      <div class="floating-star" v-for="i in 6" :key="'star'+i" :style="getStarStyle(i)">★</div>
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

    // 计算排序后的数据
    const sortedData = computed(() => {
      return Object.entries(analysisData.value)
        .sort((a, b) => b[1] - a[1])
    })

    // 加载日志文件列表
    const loadLogFiles = async () => {
      try {
        const response = await apiMethods.getLogFiles()
        logFiles.value = response.files || []
        
        // 如果有文件，自动选择第一个
        if (logFiles.value.length > 0) {
          selectedFile.value = logFiles.value[0]
          await loadAnalysisData()
        }
      } catch (err) {
        error.value = '加载日志文件列表失败：' + err.message
      }
    }

    // 加载分析数据
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

    // 文件选择变化处理
    const onFileChange = () => {
      loadAnalysisData()
    }

    // 获取排名样式类名
    const getRankClass = (index) => {
      if (index === 0) return 'first'
      if (index === 1) return 'second'
      if (index === 2) return 'third'
      return 'normal'
    }

    // 获取奖牌图标
    const getMedal = (index) => {
      const medals = ['🥇', '🥈', '🥉']
      return medals[index] || ''
    }

    // 生成粒子样式
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

    // 生成星星样式
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

    // 组件挂载时加载数据
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
/* 全局容器 */
.container {
  min-height: 100vh;
  background: linear-gradient(120deg, #fbeaf3 0%, #f7e6f2 50%, #e6e6fa 100%);
  font-family: "Comic Sans MS", "Segoe UI", "Microsoft YaHei", sans-serif;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
  position: relative;
  overflow-x: hidden;
}

/* 背景装饰 */
.bg-decoration {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  pointer-events: none;
  z-index: 0;
}

.floating-particle {
  position: absolute;
  width: 12px;
  height: 12px;
  background: radial-gradient(circle, #f7b6d9 60%, #fbeaf3 100%);
  border-radius: 50%;
  animation: float 8s ease-in-out infinite;
  opacity: 0.35;
  filter: blur(2px);
}

.floating-star {
  position: absolute;
  font-size: 18px;
  color: #e48abf;
  opacity: 0.22;
  animation: starFloat 7s infinite alternate;
  pointer-events: none;
}

@keyframes float {
  0%, 100% { transform: translateY(0px) rotate(0deg); }
  50% { transform: translateY(-24px) rotate(180deg); }
}

@keyframes starFloat {
  0% { transform: scale(1) rotate(0deg);}
  100% { transform: scale(1.2) rotate(20deg);}
}

/* 卡片玻璃风格 */
.card.glass {
  background: rgba(255,255,255,0.85);
  border-radius: 32px;
  box-shadow: 0 16px 48px 0 rgba(228,138,191,0.18), 0 0 0 2px #f7b6d9;
  backdrop-filter: blur(18px) saturate(1.2);
  border: 2px solid #f7b6d9;
  padding: 44px 36px 32px 36px;
  width: 95vw;
  max-width: 820px;
  position: relative;
  z-index: 10;
  overflow: hidden;
  animation: slideIn 0.8s cubic-bezier(.4,0,.2,1);
}

@keyframes slideIn {
  from { opacity: 0; transform: translateY(40px);}
  to { opacity: 1; transform: translateY(0);}
}

/* 卡通分割线 */
.divider {
  width: 80px;
  height: 8px;
  margin: 18px auto 0 auto;
  background: linear-gradient(90deg, #f7b6d9 0%, #e48abf 100%);
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(228,138,191,0.12);
}

/* 头部区域 */
.header {
  text-align: center;
  margin-bottom: 24px;
}

.title {
  font-size: 34px;
  font-weight: bold;
  color: #e48abf;
  text-shadow: 0 0 18px rgba(228, 138, 191, 0.18);
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.title .icon {
  font-size: 32px;
  animation: bounce 2s infinite;
}

.main-title {
  letter-spacing: 2px;
}

@keyframes bounce {
  0%, 20%, 50%, 80%, 100% { transform: translateY(0);}
  40% { transform: translateY(-12px);}
  60% { transform: translateY(-6px);}
}

.subtitle {
  font-size: 17px;
  color: #b97fae;
  opacity: 0.8;
  margin-top: 2px;
}

/* 选择器区域 */
.select-container {
  margin-bottom: 28px;
  text-align: left;
}

.select-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  color: #c18bb7;
  font-weight: 600;
  margin-bottom: 12px;
}

.label-icon {
  font-size: 19px;
}

.select-wrapper {
  position: relative;
  display: inline-block;
  width: 100%;
  max-width: 340px;
}

.select-box {
  width: 100%;
  font-size: 16px;
  padding: 14px 40px 14px 18px;
  border-radius: 16px;
  border: 2px solid #f7b6d9;
  background: linear-gradient(135deg, #fff 0%, #fbeaf3 100%);
  color: #b97fae;
  cursor: pointer;
  transition: all 0.3s ease;
  appearance: none;
  box-shadow: 0 2px 12px rgba(228, 138, 191, 0.09);
}

.select-box:focus {
  outline: none;
  border-color: #e48abf;
  box-shadow: 0 0 0 3px rgba(228, 138, 191, 0.12);
  transform: translateY(-2px);
}

.select-arrow {
  position: absolute;
  right: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: #e48abf;
  pointer-events: none;
  transition: transform 0.3s ease;
}

.select-wrapper:hover .select-arrow {
  transform: translateY(-50%) scale(1.2);
}

/* 内容区域 */
.content-area {
  margin-bottom: 28px;
  min-height: 200px;
}

/* 加载状态 */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 22px;
  padding: 44px;
}

.loading-spinner {
  width: 54px;
  height: 54px;
  border: 5px solid #f7b6d9;
  border-top: 5px solid #e48abf;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  box-shadow: 0 2px 12px rgba(228, 138, 191, 0.09);
}

@keyframes spin {
  0% { transform: rotate(0deg);}
  100% { transform: rotate(360deg);}
}

.loading-text {
  font-size: 19px;
  color: #b97fae;
  font-weight: 500;
  letter-spacing: 1px;
}

/* 错误状态 */
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 32px;
  background: linear-gradient(135deg, #fbeaf3 0%, #f7e6f2 100%);
  border-radius: 18px;
  border: 2px solid #f7b6d9;
  box-shadow: 0 2px 12px rgba(228, 138, 191, 0.09);
}

.error-icon {
  font-size: 44px;
}

.error-text {
  color: #b97fae;
  font-size: 17px;
  text-align: center;
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 44px;
  color: #b97fae;
}

.empty-icon {
  font-size: 54px;
  opacity: 0.7;
}

.empty-text {
  font-size: 19px;
  opacity: 0.8;
}

/* 数据列表 */
.data-list {
  background: rgba(247, 182, 217, 0.10);
  border-radius: 18px;
  padding: 22px;
  box-shadow: 0 2px 12px rgba(228, 138, 191, 0.07) inset;
  border: 1px dashed #e48abf;
}

.list-header {
  display: grid;
  grid-template-columns: 90px 1fr 120px;
  gap: 22px;
  padding: 16px 22px;
  font-weight: bold;
  color: #b97fae;
  border-bottom: 2px dashed #f7b6d9;
  margin-bottom: 18px;
  background: rgba(247, 182, 217, 0.07);
  border-radius: 12px 12px 0 0;
}
.rank-header, .count-header {
  text-align: right;
}
.item-header {
  text-align: left;
}

/* 数据项 */
.data-item {
  display: grid;
  grid-template-columns: 90px 1fr 120px;
  gap: 22px;
  align-items: center;
  background: linear-gradient(135deg, rgba(255,255,255,0.96) 0%, rgba(251,234,243,0.96) 100%);
  border: 2px solid transparent;
  border-radius: 16px;
  padding: 18px 22px;
  margin-bottom: 14px;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
  animation: fadeInUp 0.6s ease-out both;
  box-shadow: 0 2px 12px rgba(228, 138, 191, 0.09);
}
.rank {
  justify-content: flex-end;
}
.item-count {
  justify-content: flex-end;
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(24px);}
  to { opacity: 1; transform: translateY(0);}
}

.data-item:hover {
  transform: translateY(-5px) scale(1.03);
  box-shadow: 0 16px 32px rgba(228, 138, 191, 0.18);
  border-color: #f7b6d9;
}

.data-item::before {
  content: "";
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(228, 138, 191, 0.10), transparent);
  transition: left 0.6s ease;
}

.data-item:hover::before {
  left: 100%;
}

/* 前三名特殊样式 */
.top-item {
  background: linear-gradient(135deg, 
    rgba(255, 235, 59, 0.13) 0%, 
    rgba(255, 255, 255, 0.96) 50%, 
    rgba(255, 193, 7, 0.13) 100%);
  border-color: #ffe6b3;
  box-shadow: 0 4px 24px rgba(255, 230, 179, 0.18);
}

/* 排名区域 */
.rank {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.rank-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  font-weight: bold;
  font-size: 18px;
  box-shadow: 0 2px 8px rgba(228, 138, 191, 0.09);
}

.rank-number.first {
  background: linear-gradient(135deg, #ffe6b3, #fff7d6);
  color: #b8860b;
  box-shadow: 0 2px 12px rgba(255, 230, 179, 0.18);
}

.rank-number.second {
  background: linear-gradient(135deg, #e5e5e5, #f7f7f7);
  color: #666;
  box-shadow: 0 2px 12px rgba(192, 192, 192, 0.18);
}

.rank-number.third {
  background: linear-gradient(135deg, #f7b6d9, #fbeaf3);
  color: #ad1457;
  box-shadow: 0 2px 12px rgba(247, 182, 217, 0.18);
}

.rank-number.normal {
  background: linear-gradient(135deg, #fbeaf3, #f7e6f2);
  color: #b97fae;
}

.medal {
  font-size: 22px;
  animation: rotate 3s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg);}
  to { transform: rotate(360deg); }
}

/* 物品名称 */
.item-name {
  font-size: 19px;
  font-weight: 600;
  color: #7c5e9c;
  text-align: left;
  letter-spacing: 1px;
}

/* 数量区域 */
.item-count {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.count-value {
  font-size: 22px;
  font-weight: bold;
  color: #e48abf;
}

.count-unit {
  font-size: 15px;
  color: #b97fae;
  opacity: 0.7;
}

/* 底部区域 */
.footer {
  text-align: center;
  padding-top: 22px;
  border-top: 2px solid rgba(247, 182, 217, 0.18);
}

.back-button {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 16px 36px;
  font-size: 19px;
  font-weight: 600;
  background: linear-gradient(135deg, #fff 0%, #fbeaf3 100%);
  color: #e48abf;
  border: 2px solid #f7b6d9;
  border-radius: 24px;
  text-decoration: none;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
  box-shadow: 0 2px 12px rgba(228, 138, 191, 0.09);
}

.back-button::before {
  content: "";
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #e48abf, #f7b6d9);
  transition: left 0.4s ease;
  z-index: -1;
}

.back-button:hover {
  color: #fff;
  transform: translateY(-2px) scale(1.06);
  box-shadow: 0 8px 24px rgba(228, 138, 191, 0.18);
}

.back-button:hover::before {
  left: 0;
}

.button-icon {
  font-size: 22px;
  transition: transform 0.3s ease;
}

.back-button:hover .button-icon {
  transform: scale(1.18);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .card.glass {
    padding: 12px 6px;
    margin: 2px;
    width: 99vw;
    max-width: 99vw;
    border-radius: 16px;
  }
  .title {
    font-size: 22px;
    gap: 6px;
  }
  .main-title {
    letter-spacing: 1px;
  }
  .list-header,
  .data-item {
    grid-template-columns: 44px 1fr 54px;
    gap: 6px;
    padding: 10px 8px;
    font-size: 16px;
  }
  .list-header .rank-header,
  .list-header .count-header {
    text-align: right;
  }
  .data-item .rank {
    justify-content: flex-end;
  }
  .data-item .item-count {
    justify-content: flex-end;
  }
  .data-list {
    padding: 10px;
    border-radius: 10px;
    margin-bottom: 0;
  }
  .item-name {
    font-size: 15px;
    font-weight: 500;
  }
  .count-value {
    font-size: 16px;
  }
  .rank-number {
    width: 30px;
    height: 30px;
    font-size: 16px;
  }
  .medal {
    font-size: 17px;
  }
  .data-item {
    padding: 10px 8px;
    margin-bottom: 8px;
    border-radius: 10px;
  }
  .loading-container,
  .error-container,
  .empty-state {
    padding: 22px;
    font-size: 16px;
  }
  .footer {
    padding-top: 12px;
  }
  .back-button {
    padding: 10px 20px;
    font-size: 16px;
    border-radius: 14px;
    gap: 8px;
  }
  .button-icon {
    font-size: 16px;
  }
}

@media (max-width: 480px) {
  .card.glass {
    padding: 6px 2px;
    margin: 0;
    width: 100vw;
    max-width: 100vw;
    border-radius: 8px;
  }
  .list-header,
  .data-item {
    grid-template-columns: 1fr 1fr 1fr;
    gap: 4px;
    text-align: left;
    padding: 8px 4px;
    font-size: 14px;
  }
  .list-header .rank-header,
  .list-header .count-header {
    text-align: right;
  }
  .data-item .rank {
    justify-content: flex-end;
  }
  .data-item .item-count {
    justify-content: flex-end;
  }
  .data-list {
    padding: 4px;
    border-radius: 6px;
  }
  .item-name {
    font-size: 13px;
  }
  .count-value {
    font-size: 14px;
  }
  .rank-number {
    width: 22px;
    height: 22px;
    font-size: 13px;
  }
  .medal {
    font-size: 13px;
  }
  .data-item {
    padding: 8px 4px;
    margin-bottom: 4px;
    border-radius: 6px;
  }
  .loading-container,
  .error-container,
  .empty-state {
    padding: 12px;
    font-size: 14px;
  }
  .footer {
    padding-top: 6px;
  }
  .back-button {
    padding: 6px 12px;
    font-size: 13px;
    border-radius: 8px;
    gap: 4px;
  }
  .button-icon {
    font-size: 13px;
  }
}
</style>
