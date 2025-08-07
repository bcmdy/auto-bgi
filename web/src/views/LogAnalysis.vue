<template>
  <div class="container">
    <div class="card">
      <!-- 标题区域 -->
      <div class="header">
        <div class="title">
          <span class="icon">📊</span>
          收获前10的材料
          <span class="icon">✨</span>
        </div>
        <div class="subtitle">数据分析统计</div>
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
      <div class="floating-particle" v-for="i in 6" :key="i" :style="getParticleStyle(i)"></div>
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
      getParticleStyle
    }
  }
}
</script>

<style scoped>
/* 全局容器 */
.container {
  min-height: 100vh;
  background: linear-gradient(135deg, #fff0f7 0%, #ffe6f2 50%, #fff0f7 100%);
  background-attachment: fixed;
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
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 0;
}

.floating-particle {
  position: absolute;
  width: 8px;
  height: 8px;
  background: radial-gradient(circle, #ff99cc, #ffccee);
  border-radius: 50%;
  animation: float 6s ease-in-out infinite;
  opacity: 0.6;
}

@keyframes float {
  0%, 100% { transform: translateY(0px) rotate(0deg); }
  50% { transform: translateY(-20px) rotate(180deg); }
}

/* 主卡片 */
.card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: 3px solid transparent;
  background-image: linear-gradient(rgba(255, 255, 255, 0.95), rgba(255, 255, 255, 0.95)),
                    linear-gradient(45deg, #ff99cc, #ffccee, #ff99cc);
  background-origin: border-box;
  background-clip: content-box, border-box;
  border-radius: 30px;
  padding: 40px;
  width: 95vw;
  max-width: 800px;
  box-shadow: 0 20px 60px rgba(255, 102, 153, 0.3),
              0 0 0 1px rgba(255, 153, 204, 0.1);
  position: relative;
  z-index: 10;
  overflow: hidden;
  animation: slideIn 0.8s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.card::before {
  content: "";
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(45deg, #ff99cc, #ffccee, #ff99cc);
  border-radius: 30px;
  z-index: -1;
  opacity: 0.3;
  animation: borderGlow 3s linear infinite;
}

@keyframes borderGlow {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 0.6; }
}

/* 头部区域 */
.header {
  text-align: center;
  margin-bottom: 30px;
}

.title {
  font-size: 32px;
  font-weight: bold;
  color: #e91e63;
  text-shadow: 0 0 15px rgba(233, 30, 99, 0.3);
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.title .icon {
  font-size: 28px;
  animation: bounce 2s infinite;
}

@keyframes bounce {
  0%, 20%, 50%, 80%, 100% { transform: translateY(0); }
  40% { transform: translateY(-10px); }
  60% { transform: translateY(-5px); }
}

.subtitle {
  font-size: 16px;
  color: #ad1457;
  opacity: 0.8;
}

/* 选择器区域 */
.select-container {
  margin-bottom: 30px;
  text-align: left;
}

.select-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  color: #c2185b;
  font-weight: 600;
  margin-bottom: 12px;
}

.label-icon {
  font-size: 18px;
}

.select-wrapper {
  position: relative;
  display: inline-block;
  width: 100%;
  max-width: 300px;
}

.select-box {
  width: 100%;
  font-size: 16px;
  padding: 14px 40px 14px 16px;
  border-radius: 15px;
  border: 2px solid #f8bbd9;
  background: linear-gradient(135deg, #fff 0%, #ffeef5 100%);
  color: #ad1457;
  cursor: pointer;
  transition: all 0.3s ease;
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
}

.select-box:focus {
  outline: none;
  border-color: #e91e63;
  box-shadow: 0 0 0 3px rgba(233, 30, 99, 0.1);
  transform: translateY(-2px);
}

.select-arrow {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #e91e63;
  pointer-events: none;
  transition: transform 0.3s ease;
}

.select-wrapper:hover .select-arrow {
  transform: translateY(-50%) scale(1.2);
}

/* 内容区域 */
.content-area {
  margin-bottom: 30px;
  min-height: 200px;
}

/* 加载状态 */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  padding: 40px;
}

.loading-spinner {
  width: 50px;
  height: 50px;
  border: 4px solid #f8bbd9;
  border-top: 4px solid #e91e63;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.loading-text {
  font-size: 18px;
  color: #c2185b;
  font-weight: 500;
}

/* 错误状态 */
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
  padding: 30px;
  background: linear-gradient(135deg, #ffebee 0%, #fce4ec 100%);
  border-radius: 20px;
  border: 2px solid #f8bbd9;
}

.error-icon {
  font-size: 40px;
}

.error-text {
  color: #c2185b;
  font-size: 16px;
  text-align: center;
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
  padding: 40px;
  color: #c2185b;
}

.empty-icon {
  font-size: 50px;
  opacity: 0.7;
}

.empty-text {
  font-size: 18px;
  opacity: 0.8;
}

/* 数据列表 */
.data-list {
  background: rgba(248, 187, 217, 0.1);
  border-radius: 20px;
  padding: 20px;
  box-shadow: inset 0 2px 10px rgba(233, 30, 99, 0.05);
}

.list-header {
  display: grid;
  grid-template-columns: 80px 1fr 120px;
  gap: 20px;
  padding: 15px 20px;
  font-weight: bold;
  color: #ad1457;
  border-bottom: 2px solid #f8bbd9;
  margin-bottom: 15px;
}

.rank-header, .item-header, .count-header {
  text-align: center;
}

.item-header {
  text-align: left;
}

/* 数据项 */
.data-item {
  display: grid;
  grid-template-columns: 80px 1fr 120px;
  gap: 20px;
  align-items: center;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.9) 0%, rgba(254, 240, 248, 0.9) 100%);
  border: 2px solid transparent;
  border-radius: 18px;
  padding: 18px 20px;
  margin-bottom: 12px;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
  animation: fadeInUp 0.6s ease-out both;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.data-item:hover {
  transform: translateY(-5px) scale(1.02);
  box-shadow: 0 15px 35px rgba(233, 30, 99, 0.2);
  border-color: #f8bbd9;
}

.data-item::before {
  content: "";
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(233, 30, 99, 0.1), transparent);
  transition: left 0.6s ease;
}

.data-item:hover::before {
  left: 100%;
}

/* 前三名特殊样式 */
.top-item {
  background: linear-gradient(135deg, 
    rgba(255, 235, 59, 0.15) 0%, 
    rgba(255, 255, 255, 0.95) 50%, 
    rgba(255, 193, 7, 0.15) 100%);
  border-color: #ffd54f;
}

/* 排名区域 */
.rank {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.rank-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 35px;
  height: 35px;
  border-radius: 50%;
  font-weight: bold;
  font-size: 16px;
}

.rank-number.first {
  background: linear-gradient(135deg, #ffd700, #ffed4a);
  color: #b8860b;
  box-shadow: 0 4px 15px rgba(255, 215, 0, 0.4);
}

.rank-number.second {
  background: linear-gradient(135deg, #c0c0c0, #e5e5e5);
  color: #666;
  box-shadow: 0 4px 15px rgba(192, 192, 192, 0.4);
}

.rank-number.third {
  background: linear-gradient(135deg, #cd7f32, #daa520);
  color: #8b4513;
  box-shadow: 0 4px 15px rgba(205, 127, 50, 0.4);
}

.rank-number.normal {
  background: linear-gradient(135deg, #f8bbd9, #fce4ec);
  color: #ad1457;
}

.medal {
  font-size: 20px;
  animation: rotate 3s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 物品名称 */
.item-name {
  font-size: 18px;
  font-weight: 600;
  color: #4a148c;
  text-align: left;
}

/* 数量区域 */
.item-count {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.count-value {
  font-size: 20px;
  font-weight: bold;
  color: #e91e63;
}

.count-unit {
  font-size: 14px;
  color: #ad1457;
  opacity: 0.8;
}

/* 底部区域 */
.footer {
  text-align: center;
  padding-top: 20px;
  border-top: 2px solid rgba(248, 187, 217, 0.3);
}

.back-button {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 15px 30px;
  font-size: 18px;
  font-weight: 600;
  background: linear-gradient(135deg, #fff 0%, #ffeef5 100%);
  color: #e91e63;
  border: 3px solid #f8bbd9;
  border-radius: 25px;
  text-decoration: none;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.back-button::before {
  content: "";
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #e91e63, #f06292);
  transition: left 0.4s ease;
  z-index: -1;
}

.back-button:hover {
  color: #fff;
  transform: translateY(-3px) scale(1.05);
  box-shadow: 0 15px 35px rgba(233, 30, 99, 0.3);
}

.back-button:hover::before {
  left: 0;
}

.button-icon {
  font-size: 20px;
  transition: transform 0.3s ease;
}

.back-button:hover .button-icon {
  transform: scale(1.2);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .card {
    padding: 25px;
    margin: 10px;
  }
  
  .title {
    font-size: 26px;
  }
  
  .list-header,
  .data-item {
    grid-template-columns: 60px 1fr 100px;
    gap: 15px;
  }
  
  .item-name {
    font-size: 16px;
  }
  
  .count-value {
    font-size: 18px;
  }
}

@media (max-width: 480px) {
  .list-header,
  .data-item {
    grid-template-columns: 1fr;
    gap: 10px;
    text-align: center;
  }
  
  .rank {
    justify-content: center;
  }
  
  .item-name {
    text-align: center;
  }
}
</style>
