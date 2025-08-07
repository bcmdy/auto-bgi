<template>
  <div class="other-page">
    <!-- 页面头部 -->
    <header class="page-header">
      <button class="btn" @click="$router.push('/')">返回首页</button>
      <h1>详细日志分析</h1>
    </header>

    <!-- 书签导航 -->
    <div v-if="analysisData.length > 0" class="bookmark-nav">
      <div class="bookmark-header">
        <span class="bookmark-title">📑 快速导航</span>
        <button class="bookmark-toggle" @click="toggleBookmark" :class="{ 'active': bookmarkVisible }">
          {{ bookmarkVisible ? '◀' : '▶' }}
        </button>
      </div>
      <transition name="slide-left">
        <div v-if="bookmarkVisible" class="bookmark-list">
          <div 
            v-for="(group, index) in analysisData" 
            :key="group.GroupName"
            class="bookmark-item"
            :class="{ 'active': currentActiveGroup === group.GroupName }"
            @click="scrollToGroup(group.GroupName)"
          >
            <span class="bookmark-number">{{ index + 1 }}</span>
            <span class="bookmark-name">{{ formatGroupName(group.GroupName) }}</span>
            <span class="bookmark-time">{{ group.Consuming }}</span>
          </div>
        </div>
      </transition>
    </div>

    <div class="container">
      <!-- 文件选择面板 -->
      <section class="panel file-selector-panel">
        <div class="file-selector-header">
          <h3>日志文件</h3>
          <select 
            v-model="selectedFile" 
            @change="loadAnalysisData" 
            class="file-select"
            :disabled="loading || logFiles.length === 0"
          >
            <option value="" disabled>请选择文件</option>
            <option v-for="file in logFiles" :key="file" :value="file">
              {{ formatFileName(file) }}
            </option>
          </select>
        </div>
      </section>

      <!-- 详细日志分析 -->
      <section v-if="analysisData.length > 0" class="panel analysis-panel">
        <div class="panel-title">
          <h2>📊 日志分析结果</h2>
          <div class="stats-badge">
            <span class="stats-count">{{ analysisData.length }}</span>
            <span class="stats-label">个配置组</span>
          </div>
        </div>
        
        <div class="analysis-result">
          <div 
            v-for="(group, index) in analysisData" 
            :key="group.GroupName"
            :id="`group-${group.GroupName}`"
            class="group-card"
            :style="{ '--delay': index * 0.1 + 's' }"
          >
            <!-- 卡片头部 - 始终可见 -->
            <div class="group-header">
              <div class="group-title">
                <div class="group-icon">🔧</div>
                <div class="group-main-info">
                  <h3 class="group-name">{{ group.GroupName }}</h3>
                  <div class="group-time-info">
                    <span class="time-badge start">{{ group.StartTime }}</span>
                    <span class="duration-arrow">→</span>
                    <span class="time-badge end">{{ group.EndTime }}</span>
                    <span class="duration-badge">{{ group.Consuming }}</span>
                  </div>
                </div>
              </div>
              <div class="group-actions">
                <button class="btn archive-btn-always" @click="archiveGroup(group)" title="归档此配置组">
                  📥 归档
                </button>
                <button class="toggle-btn" @click="toggleGroupDetails(group.GroupName)">
                  <span v-if="expandedGroups.includes(group.GroupName)" style="color: #ff6eb4;">📖 收起</span>
                  <span v-else style="color: #ff6eb4;">📋 详情</span>
                </button>
              </div>
            </div>

            <!-- 卡片内容 - 可折叠 -->
            <transition name="slide-down">
              <div v-if="expandedGroups.includes(group.GroupName)" class="group-content">
                                 <div class="error-section">
                   <h4 class="section-title">❗ 错误汇总</h4>
                   <div class="error-summary" v-html="formatMap(group.ErrorSummary)"></div>
                 </div>
                
                <!-- 子任务详情 -->
                <div v-if="group.LogAnalysis2Json && group.LogAnalysis2Json.length > 0" class="tasks-section">
                  <h4 class="section-title">📝 子任务详情</h4>
                  <div class="tasks-grid">
                    <div 
                      v-for="sub in group.LogAnalysis2Json" 
                      :key="sub.JsonName"
                      class="task-card"
                    >
                      <div class="task-header">
                        <span class="task-icon">⚙️</span>
                        <h5 class="task-name">{{ sub.JsonName }}</h5>
                      </div>
                      <div class="task-details">
                        <div class="task-time">
                          <span class="task-time-label">开始：</span>
                          <span class="task-start">{{ sub.StartTime }}</span>
                          <span class="task-time-label">结束：</span>
                          <span class="task-end">{{ sub.EndTime }}</span>
                          <span class="task-time-label">耗时：</span>
                          <span class="task-duration">{{ sub.Consuming }}</span>
                        </div>
                        <div class="task-income">
                          <strong>💰 收入：</strong>
                          <div class="income-content" v-html="formatMap(sub.Income)"></div>
                        </div>
                        <div class="task-errors">
                          <strong>⚠️ 错误：</strong>
                          <div class="error-content" v-html="formatMap(sub.Errors)"></div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                
                <div v-else class="no-tasks">
                  <div class="no-tasks-icon">📭</div>
                  <p>暂无子任务记录</p>
                </div>
              </div>
            </transition>
          </div>
        </div>
      </section>

      <!-- 加载状态 -->
      <section v-else-if="loading" class="panel">
        <p class="loading-text">正在加载数据...</p>
      </section>

          <!-- 无数据状态 -->
    <section v-else class="panel">
      <p class="no-data-text">暂无数据</p>
    </section>
    
  </div>

    <!-- 回到顶部按钮 -->
    <button 
      class="back-to-top-btn" 
      @click="scrollToTop"
      title="回到顶部"
    >
      <span class="back-to-top-icon">⬆️</span>
      <span class="back-to-top-text">顶部</span>
    </button>



  </div>
</template>

<script>
import api from '@/utils/api'

export default {
  name: 'Other',
  data() {
    return {
      logFiles: [],
      selectedFile: '',
      analysisData: [],
      loading: false,
      expandedGroups: [], // 记录展开的配置组
      bookmarkVisible: true, // 书签是否可见
      currentActiveGroup: '' // 当前活跃的配置组
    }
  },
  async mounted() {
    await this.loadLogFiles()
  },
  methods: {
    // 加载日志文件列表
    async loadLogFiles() {
      try {
        const response = await api.get('/api/logFiles')
        this.logFiles = response.files || []
        if (this.logFiles.length > 0) {
          this.selectedFile = this.logFiles[0]
          await this.loadAnalysisData()
        }
      } catch (error) {
        console.error('加载日志文件列表失败:', error)
        this.$message?.error('加载日志文件列表失败')
      }
    },

    // 加载分析数据
    async loadAnalysisData() {
      if (!this.selectedFile) return
      
      this.loading = true
      try {
        const response = await api.get(`/api/LogAnalysis2Page?file=${encodeURIComponent(this.selectedFile)}`)
        this.analysisData = response.data || []
        // 重置当前活跃组和展开状态
        this.currentActiveGroup = ''
        this.expandedGroups = []
      } catch (error) {
        console.error('加载分析数据失败:', error)
        this.$message?.error('加载分析数据失败')
      } finally {
        this.loading = false
      }
    },

    // 归档配置组
    async archiveGroup(group) {
      try {
        const archiveItem = {
          Title: group.GroupName,
          ExecuteTime: group.Consuming,
          StartTime: group.StartTime,
          EndTime: group.EndTime
        }
        
        const response = await api.post('/api/archive', archiveItem)
        this.$message?.success('归档成功: ' + response)
      } catch (error) {
        console.error('归档失败:', error)
        this.$message?.error('归档失败')
      }
    },

    // 格式化映射数据
    formatMap(mapData) {
      if (!mapData || Object.keys(mapData).length === 0) {
        return '(无记录)'
      }
      return Object.entries(mapData)
        .map(([k, v]) => `- ${k}：${v}`)
        .join('<br>')
    },

    // 格式化文件名显示
    formatFileName(fileName) {
      if (!fileName) return ''
      
      // 如果文件名太长，显示省略号
      if (fileName.length > 50) {
        return fileName.substring(0, 47) + '...'
      }
      return fileName
    },

    // 切换配置组详情展开/收起 - 手风琴效果
    toggleGroupDetails(groupName) {
      const index = this.expandedGroups.indexOf(groupName)
      if (index > -1) {
        // 如果当前组已展开，则收起
        this.expandedGroups.splice(index, 1)
      } else {
        // 如果当前组未展开，则收起所有其他组，只展开当前组
        this.expandedGroups = [groupName]
      }
    },

    // 切换书签显示/隐藏
    toggleBookmark() {
      this.bookmarkVisible = !this.bookmarkVisible
    },

    // 滚动到指定配置组
    scrollToGroup(groupName) {
      const element = document.getElementById(`group-${groupName}`)
      if (element) {
        element.scrollIntoView({ 
          behavior: 'smooth', 
          block: 'start',
          inline: 'nearest'
        })
        // 设置当前活跃组
        this.currentActiveGroup = groupName
        // 可选：自动展开该组的详情
        if (!this.expandedGroups.includes(groupName)) {
          this.expandedGroups = [groupName]
        }
      }
    },

    // 格式化配置组名称
    formatGroupName(groupName) {
      if (!groupName) return ''
      
      // 如果名称太长，显示省略号
      if (groupName.length > 20) {
        return groupName.substring(0, 17) + '...'
      }
      return groupName
    },

    // 测试点击
    testClick() {

      
      // 使用更简单有效的方法
      try {
        // 方法1: 滚动到页面顶部元素
        const pageHeader = document.querySelector('.page-header')
        if (pageHeader) {
          pageHeader.scrollIntoView({
            behavior: 'smooth',
            block: 'start'
          })
        }
        
        // 方法2: 直接设置滚动位置
        window.scrollTo({
          top: 0,
          behavior: 'smooth'
        })
        
        // 方法3: 备用方案
        document.documentElement.scrollTop = 0
        document.body.scrollTop = 0

      } catch (error) {
        console.error('滚动失败:', error)
      }
    },

    // 回到顶部
    scrollToTop() {
      console.log('回到顶部按钮被点击')
      try {
        // 方法1: 滚动到页面顶部元素
        const pageHeader = document.querySelector('.page-header')
        if (pageHeader) {
          pageHeader.scrollIntoView({
            behavior: 'smooth',
            block: 'start'
          })
        }
        
        // 方法2: 直接设置滚动位置
        window.scrollTo({
          top: 0,
          behavior: 'smooth'
        })
        
        // 方法3: 备用方案
        document.documentElement.scrollTop = 0
        document.body.scrollTop = 0

      } catch (error) {
        console.error('滚动到顶部失败:', error)
      }
    }
  }
}
</script>

<style scoped>
:root {
  --primary-color: #ff6eb4;
  --background-light: #fff6fb;
  --text-color: #ff6eb4;
  --border-color: #ffc0da;
  --hover-color: rgba(255, 192, 218, 0.3);
}

.other-page {
  min-height: 100vh;
  background-color: var(--background-light);
  color: var(--text-color);
  background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100"><circle cx="20" cy="20" r="5" fill="%23ffd6eb" opacity="0.6"/><circle cx="70" cy="70" r="7" fill="%23ffc0da" opacity="0.5"/></svg>');
  padding-bottom: 50px;
}

.page-header {
  background-color: rgba(255, 255, 255, 0.8);
  padding: 30px 0 10px;
  text-align: center;
  box-shadow: 0 0 20px var(--primary-color);
  border-radius: 0 0 30px 30px;
  position: sticky;
  top: 0;
  z-index: 10;
}

.page-header h1 {
  color: var(--primary-color);
  font-size: 2rem;
  text-shadow: 0 0 10px var(--primary-color);
  margin-top: 15px;
}

.btn {
  background-color: #fff;
  color: var(--primary-color);
  border: 2px solid var(--primary-color);
  border-radius: 50px;
  padding: 8px 16px;
  font-size: 1rem;
  cursor: pointer;
  box-shadow: 0 0 10px var(--primary-color);
  transition: all 0.3s ease;
  font-weight: bold;
  margin-top: 10px;
}

.btn:hover {
  background-color: var(--primary-color);
  color: #fff;
  box-shadow: 0 0 20px var(--primary-color);
  transform: scale(1.05);
}

.container {
  max-width: 1200px;
  margin: 30px auto;
  padding: 0 20px;
}

.panel {
  background: rgba(255, 255, 255, 0.8);
  box-shadow: 0 0 15px #ffcce6;
  border-radius: 20px;
  padding: 20px 25px;
  margin-bottom: 30px;
}

.panel h2 {
  color: var(--primary-color);
  font-size: 1.6rem;
  margin-bottom: 15px;
  border-bottom: 2px solid var(--primary-color);
  padding-bottom: 5px;
  display: inline-block;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 15px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}

.panel-header h2 {
  margin: 0;
}

/* 文件选择器面板样式 - 简化版 */
.file-selector-panel {
  padding: 15px 20px;
  margin-bottom: 20px;
}

.file-selector-header {
  display: flex;
  align-items: center;
  gap: 15px;
  flex-wrap: wrap;
}

.file-selector-header h3 {
  color: var(--primary-color);
  font-size: 1.2rem;
  margin: 0;
  font-weight: bold;
  white-space: nowrap;
}

.file-select {
  padding: 8px 12px;
  border: 2px solid var(--primary-color);
  border-radius: 8px;
  color: var(--primary-color);
  background-color: #fff;
  font-size: 0.95rem;
  cursor: pointer;
  min-width: 200px;
  transition: all 0.2s ease;
}

.file-select:focus {
  outline: none;
  border-color: #ff4d9a;
  box-shadow: 0 0 0 2px rgba(255, 110, 180, 0.2);
}

.file-select:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  background-color: #f8f8f8;
}

/* 书签导航样式 */
.bookmark-nav {
  position: fixed;
  top: 50%;
  right: 0;
  transform: translateY(-50%);
  z-index: 100;
  max-height: 70vh;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(255, 240, 250, 0.9));
  border: 2px solid var(--border-color);
  border-right: none;
  border-radius: 20px 0 0 20px;
  box-shadow: -5px 0 20px rgba(255, 110, 180, 0.2);
  overflow: hidden;
}

.bookmark-header {
  background: linear-gradient(45deg, var(--primary-color), #ff9ecf);
  color: white;
  padding: 12px 15px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: bold;
  font-size: 0.9rem;
}

.bookmark-title {
  font-size: 0.85rem;
  color: #000;
}

.bookmark-toggle {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 4px 8px;
  border-radius: 15px;
  cursor: pointer;
  font-size: 0.8rem;
  transition: all 0.3s ease;
  font-weight: bold;
}

.bookmark-toggle:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: scale(1.05);
}

.bookmark-toggle.active {
  background: rgba(255, 255, 255, 0.9);
  color: var(--primary-color);
}

.bookmark-list {
  max-height: 60vh;
  overflow-y: auto;
  padding: 8px 0;
}

.bookmark-list::-webkit-scrollbar {
  width: 4px;
}

.bookmark-list::-webkit-scrollbar-track {
  background: rgba(255, 110, 180, 0.1);
}

.bookmark-list::-webkit-scrollbar-thumb {
  background: var(--primary-color);
  border-radius: 2px;
}

.bookmark-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  border-bottom: 1px solid rgba(255, 110, 180, 0.1);
  gap: 8px;
  min-width: 200px;
}

.bookmark-item:hover {
  background: rgba(255, 110, 180, 0.1);
  transform: translateX(-3px);
  border-left: 3px solid var(--primary-color);
  padding-left: 9px;
}

.bookmark-item.active {
  background: rgba(255, 110, 180, 0.15);
  border-left: 3px solid var(--primary-color);
  padding-left: 9px;
  font-weight: bold;
}

.bookmark-number {
  background: linear-gradient(45deg, var(--primary-color), #ff9ecf);
  color: rgb(235, 13, 135);
  font-size: 0.7rem;
  font-weight: bold;
  padding: 2px 6px;
  border-radius: 10px;
  min-width: 16px;
  text-align: center;
  box-shadow: 0 2px 4px rgba(255, 110, 180, 0.3);
  flex-shrink: 0;
}

.bookmark-name {
  flex: 1;
  font-size: 0.8rem;
  color: #333;
  font-weight: 500;
  line-height: 1.2;
}

.bookmark-time {
  font-size: 0.7rem;
  color: var(--primary-color);
  background: rgba(255, 110, 180, 0.1);
  padding: 2px 6px;
  border-radius: 8px;
  font-weight: bold;
  flex-shrink: 0;
}

/* 书签过渡动画 */
.slide-left-enter-active, .slide-left-leave-active {
  transition: all 0.3s ease;
}

.slide-left-enter-from, .slide-left-leave-to {
  opacity: 0;
  transform: translateX(100%);
}

.slide-left-enter-to, .slide-left-leave-from {
  opacity: 1;
  transform: translateX(0);
}

/* 分析面板美化样式 */
.analysis-panel {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(255, 245, 252, 0.9));
  border: 2px solid var(--border-color);
  position: relative;
  overflow: hidden;
}

.analysis-panel::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, var(--primary-color), #ff9ecf, #ffc0da, var(--primary-color));
  background-size: 200% 100%;
  animation: shimmer 3s linear infinite;
}

@keyframes shimmer {
  0% { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}

@keyframes pulse {
  0% {
    box-shadow: 0 4px 20px rgba(255, 110, 180, 0.4);
  }
  50% {
    box-shadow: 0 4px 20px rgba(255, 110, 180, 0.6), 0 0 0 10px rgba(255, 110, 180, 0.1);
  }
  100% {
    box-shadow: 0 4px 20px rgba(255, 110, 180, 0.4);
  }
}

.panel-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 25px;
  padding-bottom: 15px;
  border-bottom: 2px solid var(--border-color);
}

.panel-title h2 {
  margin: 0;
  font-size: 1.8rem;
  color: var(--primary-color);
  text-shadow: 0 2px 4px rgba(255, 110, 180, 0.3);
}

.stats-badge {
  background: linear-gradient(45deg, var(--primary-color), #ff9ecf);
  color: white;
  padding: 8px 16px;
  border-radius: 25px;
  display: flex;
  align-items: center;
  gap: 5px;
  font-weight: bold;
  box-shadow: 0 4px 12px rgba(255, 110, 180, 0.3);
}

.stats-count {
  font-size: 1.2rem;
  color: #ff6eb4;
}

.stats-label {
  font-size: 0.9rem;
  color: #ff6eb4;
}

/* 配置组卡片样式 - 增强层次感 */
.group-card {
  background: linear-gradient(135deg, #ffffff, #fefcff);
  border: 1px solid #ff6eb4;
  border-radius: 24px;
  margin-bottom: 30px;
  overflow: hidden;
  box-shadow: 
    0 12px 40px rgba(255, 110, 180, 0.12),
    0 4px 16px rgba(255, 110, 180, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  animation: slideInUp 0.5s ease-out var(--delay, 0s) both;
  position: relative;
}

@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.group-card:hover {
  transform: translateY(-8px);
  box-shadow: 
    0 20px 60px rgba(255, 110, 180, 0.2),
    0 8px 24px rgba(255, 110, 180, 0.15),
    inset 0 1px 0 rgba(255, 255, 255, 0.9);
  border-color: var(--primary-color);
}

.group-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 6px;
  background: linear-gradient(90deg, var(--primary-color), #ff9ecf, #ffc0da, var(--primary-color));
  background-size: 200% 100%;
  animation: shimmer 4s linear infinite;
  border-radius: 24px 24px 0 0;
}

.group-header {
  padding: 20px;
  background: linear-gradient(135deg, rgba(255, 240, 250, 0.8), rgba(255, 255, 255, 0.9));
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.group-title {
  display: flex;
  align-items: center;
  gap: 15px;
  flex: 1;
}

.group-icon {
  font-size: 2rem;
  padding: 10px;
  background: linear-gradient(45deg, var(--primary-color), #ff9ecf);
  border-radius: 15px;
  box-shadow: 0 4px 12px rgba(255, 110, 180, 0.3);
}

.group-main-info {
  flex: 1;
}

.group-name {
  margin: 0 0 8px 0;
  font-size: 1.4rem;
  color: var(--primary-color);
  font-weight: bold;
}

.group-time-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.time-badge {
  background: rgba(255, 110, 180, 0.1);
  color: var(--primary-color);
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 0.85rem;
  border: 1px solid var(--border-color);
  font-family: 'Courier New', monospace;
}

.time-badge.start {
  background: rgba(76, 175, 80, 0.1);
  color: #2e7d32;
  border-color: #4caf50;
}

.time-badge.end {
  background: rgba(244, 67, 54, 0.1);
  color: #c62828;
  border-color: #f44336;
}

.duration-arrow {
  color: var(--primary-color);
  font-weight: bold;
  font-size: 1.2rem;
}

.duration-badge {
  background: linear-gradient(45deg, var(--primary-color), #ff9ecf);
  color: #ff6eb4;
  padding: 4px 12px;
  border-radius: 15px;
  font-size: 0.85rem;
  font-weight: bold;
  box-shadow: 0 2px 6px rgba(255, 110, 180, 0.3);
}

.group-actions {
  display: flex;
  gap: 10px;
  align-items: center;
}

.archive-btn-always {
  background: linear-gradient(45deg, #4caf50, #66bb6a);
  color: white;
  border: none;
  padding: 10px 16px;
  font-size: 0.9rem;
  font-weight: bold;
  box-shadow: 0 4px 12px rgba(76, 175, 80, 0.3);
  transition: all 0.3s ease;
}

.archive-btn-always:hover {
  background: linear-gradient(45deg, #66bb6a, #4caf50);
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(76, 175, 80, 0.4);
}

.toggle-btn {
  background: linear-gradient(45deg, var(--primary-color), #ff9ecf);
  color: white;
  border: none;
  padding: 10px 16px;
  border-radius: 50px;
  font-size: 0.9rem;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 4px 12px rgba(255, 110, 180, 0.3);
}

.toggle-btn:hover {
  background: linear-gradient(45deg, #ff9ecf, var(--primary-color));
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(255, 110, 180, 0.4);
}

/* 卡片内容样式 - 增强层次感 */
.group-content {
  padding: 30px;
  background: linear-gradient(135deg, rgba(252, 250, 255, 0.9), rgba(255, 255, 255, 0.8));
  border-top: 1px solid rgba(255, 192, 218, 0.3);
  position: relative;
}

.group-content::before {
  content: '';
  position: absolute;
  top: 0;
  left: 30px;
  right: 30px;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(255, 110, 180, 0.2), transparent);
}

.section-title {
  color: var(--primary-color);
  font-size: 1.2rem;
  margin: 0 0 15px 0;
  padding-bottom: 8px;
  border-bottom: 2px solid var(--border-color);
  display: flex;
  align-items: center;
  gap: 8px;
}

.error-section {
  margin-bottom: 30px;
  padding: 20px;
  background: rgba(255, 240, 246, 0.4);
  border-radius: 16px;
  border: 1px solid rgba(255, 204, 230, 0.3);
  position: relative;
}

.error-section::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, rgba(244, 67, 54, 0.3), rgba(255, 110, 180, 0.2), rgba(244, 67, 54, 0.3));
  border-radius: 16px 16px 0 0;
}

.error-summary {
  background: linear-gradient(135deg, #fefafc, #fff5f9);
  border: 2px solid rgba(255, 204, 230, 0.4);
  border-radius: 14px;
  padding: 18px;
  word-break: break-word;
  font-size: 0.9rem;
  color: #666;
  line-height: 1.6;
  box-shadow: 
    0 6px 20px rgba(255, 110, 180, 0.08),
    0 2px 8px rgba(255, 110, 180, 0.05),
    inset 0 1px 0 rgba(255, 255, 255, 0.6);
  margin-top: 15px;
}

.error-summary br {
  display: block;
  margin: 8px 0;
}

.tasks-section {
  margin-bottom: 25px;
  padding: 20px;
  background: rgba(255, 245, 252, 0.3);
  border-radius: 16px;
  border: 1px solid rgba(255, 192, 218, 0.2);
  position: relative;
}

.tasks-section::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, rgba(255, 110, 180, 0.3), rgba(255, 158, 207, 0.2), rgba(255, 110, 180, 0.3));
  border-radius: 16px 16px 0 0;
}

.tasks-grid {
  display: grid;
  gap: 20px;
  grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
  margin-top: 15px;
}

.task-card {
  background: linear-gradient(135deg, #fcfaff, #ffffff);
  border: 2px solid rgba(255, 192, 218, 0.4);
  border-radius: 16px;
  padding: 18px;
  margin-left: 20px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 
    0 6px 20px rgba(255, 110, 180, 0.08),
    0 2px 8px rgba(255, 110, 180, 0.05),
    inset 0 1px 0 rgba(255, 255, 255, 0.6);
  position: relative;
  transform: translateX(10px);
}

.task-card:hover {
  transform: translateY(-4px) translateX(5px);
  box-shadow: 
    0 12px 32px rgba(255, 110, 180, 0.15),
    0 4px 16px rgba(255, 110, 180, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
  border-color: rgba(255, 110, 180, 0.6);
}

.task-card::before {
  content: '';
  position: absolute;
  left: -20px;
  top: 50%;
  transform: translateY(-50%);
  width: 4px;
  height: 40%;
  background: linear-gradient(180deg, var(--primary-color), #ff9ecf);
  border-radius: 2px;
  opacity: 0.6;
}

.task-card::after {
  content: '';
  position: absolute;
  left: -16px;
  top: 50%;
  transform: translateY(-50%);
  width: 8px;
  height: 2px;
  background: linear-gradient(90deg, var(--primary-color), transparent);
  border-radius: 1px;
}

.task-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}

.task-icon {
  font-size: 1.2rem;
}

.task-name {
  margin: 0;
  color: var(--primary-color);
  font-size: 1.1rem;
  font-weight: bold;
}

.task-details > div {
  margin-bottom: 10px;
}

.task-time {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 15px;
  flex-wrap: wrap;
}

.task-time-label {
  color: var(--primary-color);
  font-weight: bold;
  font-size: 0.85rem;
}

.task-start {
  background: rgba(76, 175, 80, 0.1);
  color: #2e7d32;
  padding: 3px 8px;
  border-radius: 8px;
  font-size: 0.8rem;
  font-family: 'Courier New', monospace;
  border: 1px solid rgba(76, 175, 80, 0.3);
}

.task-end {
  background: rgba(244, 67, 54, 0.1);
  color: #c62828;
  padding: 3px 8px;
  border-radius: 8px;
  font-size: 0.8rem;
  font-family: 'Courier New', monospace;
  border: 1px solid rgba(244, 67, 54, 0.3);
}

.task-duration {
  background: rgba(255, 110, 180, 0.1);
  color: var(--primary-color);
  padding: 3px 8px;
  border-radius: 8px;
  font-size: 0.8rem;
  font-weight: bold;
  border: 1px solid rgba(255, 110, 180, 0.3);
}

.income-content,
.error-content {
  background: rgba(255, 255, 255, 0.8);
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  margin-top: 5px;
  font-size: 0.85rem;
  color: #666;
  line-height: 1.5;
  word-break: break-word;
}

.income-content br,
.error-content br {
  display: block;
  margin: 4px 0;
}

.no-tasks {
  text-align: center;
  padding: 30px;
  color: #999;
}

.no-tasks-icon {
  font-size: 3rem;
  margin-bottom: 10px;
}

/* 过渡动画 */
.slide-down-enter-active, .slide-down-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.slide-down-enter-from, .slide-down-leave-to {
  opacity: 0;
  max-height: 0;
  transform: translateY(-20px);
}

.slide-down-enter-to, .slide-down-leave-from {
  opacity: 1;
  max-height: 1000px;
  transform: translateY(0);
}

.loading-text, .no-data-text {
  text-align: center;
  color: var(--primary-color);
  font-size: 1.2rem;
  padding: 40px 20px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.9), rgba(255, 245, 252, 0.8));
  border-radius: 15px;
  border: 2px dashed var(--border-color);
}

/* 测试滚动内容样式 */
.test-scroll-panel {
  margin-top: 30px;
}

.test-scroll-panel h3 {
  color: var(--primary-color);
  font-size: 1.4rem;
  margin-bottom: 20px;
  text-align: center;
}

.test-item {
  padding: 15px;
  margin-bottom: 15px;
  background: rgba(255, 255, 255, 0.6);
  border-radius: 10px;
  border: 1px solid var(--border-color);
}

.test-item p {
  margin: 5px 0;
  color: #666;
  line-height: 1.5;
}

/* 回到顶部按钮样式 */
.back-to-top-btn {
  position: fixed;
  bottom: 30px;
  right: 30px;
  width: 60px;
  height: 60px;
  background: linear-gradient(45deg, var(--primary-color), #ff9ecf);
  color: white;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 20px rgba(255, 110, 180, 0.4);
  transition: all 0.3s ease;
  z-index: 999999;
  font-weight: bold;
  /* 恢复脉冲动画效果 */
  animation: pulse 2s infinite;
}

.back-to-top-btn:hover {
  background: linear-gradient(45deg, #ff9ecf, var(--primary-color));
  transform: translateY(-3px) scale(1.1);
  box-shadow: 0 6px 25px rgba(255, 110, 180, 0.5);
  animation: none; /* 悬停时停止脉冲动画 */
}

.back-to-top-btn:active {
  transform: translateY(-1px);
}

.back-to-top-icon {
  font-size: 1.2rem;
  margin-bottom: 2px;
}

.back-to-top-text {
  font-size: 0.7rem;
  line-height: 1;
}



@media (max-width: 600px) {
  .panel-header {
    flex-direction: column;
    align-items: flex-start;
  }

  /* 文件选择器响应式设计 */
  .file-selector-header {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }
  
  .file-selector-header h3 {
    font-size: 1.1rem;
  }
  
  .file-select {
    min-width: auto;
    width: 100%;
    font-size: 0.9rem;
  }

  /* 分析面板响应式设计 */
  .panel-title {
    flex-direction: column;
    align-items: stretch;
    gap: 15px;
  }

  .panel-title h2 {
    font-size: 1.5rem;
  }

  .stats-badge {
    align-self: center;
  }

  .group-header {
    flex-direction: column;
    align-items: stretch;
    gap: 15px;
  }

  .group-title {
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 10px;
  }

  .group-icon {
    font-size: 1.5rem;
    padding: 8px;
  }

  .group-name {
    font-size: 1.2rem;
  }

  .group-time-info {
    justify-content: center;
    flex-wrap: wrap;
  }

  .group-actions {
    flex-direction: column;
    gap: 8px;
  }

  .archive-btn-always,
  .toggle-btn {
    width: 100%;
    justify-content: center;
  }

  .group-content {
    padding: 15px;
    width: 100%;
    box-sizing: border-box;
  }

  .group-content::before {
    left: 15px;
    right: 15px;
  }

  .error-section {
    padding: 15px;
    margin-bottom: 20px;
    width: 100%;
    box-sizing: border-box;
  }

  .error-summary {
    padding: 15px;
    margin-top: 10px;
    width: 100%;
    box-sizing: border-box;
    font-size: 0.85rem;
  }

  .tasks-section {
    padding: 15px;
    margin-bottom: 20px;
  }

  .tasks-grid {
    grid-template-columns: 1fr;
    gap: 15px;
    margin-top: 10px;
    width: 100%;
  }

  .tasks-section {
    padding: 15px;
    margin-bottom: 20px;
    width: 100%;
    box-sizing: border-box;
  }

  .task-card {
    padding: 15px;
    margin-left: 0;
    transform: none;
    width: 100%;
    box-sizing: border-box;
  }

  .task-card::before {
    display: none;
  }

  .task-card::after {
    display: none;
  }

  .error-section {
    padding: 15px;
    margin-bottom: 20px;
  }

  .error-summary {
    padding: 15px;
    margin-top: 10px;
  }

  .task-time {
    flex-direction: column;
    align-items: flex-start;
    gap: 5px;
    margin-bottom: 10px;
  }

  .task-time-label {
    font-size: 0.8rem;
    min-width: 40px;
  }

  .task-start,
  .task-end,
  .task-duration {
    font-size: 0.75rem;
    padding: 2px 6px;
    word-break: break-all;
    max-width: 100%;
  }

  .task-name {
    font-size: 1rem;
    word-break: break-word;
    line-height: 1.3;
  }

  .task-details {
    margin-top: 10px;
  }

  .task-income,
  .task-errors {
    margin-bottom: 8px;
  }

  .income-content,
  .error-content {
    font-size: 0.8rem;
    padding: 6px 8px;
    word-break: break-word;
    max-width: 100%;
  }

  /* 书签导航响应式设计 */
  .bookmark-nav {
    position: relative;
    top: auto;
    right: auto;
    transform: none;
    max-height: none;
    margin-bottom: 20px;
    border-right: 2px solid var(--border-color);
    border-radius: 20px;
  }

  .bookmark-header {
    padding: 10px 12px;
    font-size: 0.8rem;
  }

  .bookmark-title {
    font-size: 0.8rem;
  }

  .bookmark-toggle {
    padding: 3px 6px;
    font-size: 0.75rem;
  }

  .bookmark-list {
    max-height: 200px;
  }

  .bookmark-item {
    padding: 6px 10px;
    min-width: auto;
  }

  .bookmark-name {
    font-size: 0.75rem;
  }

  .bookmark-time {
    font-size: 0.65rem;
  }

  .bookmark-number {
    font-size: 0.65rem;
    padding: 1px 4px;
    min-width: 14px;
  }

  /* 移动端回到顶部按钮适配 */
  .back-to-top-btn {
    bottom: 20px;
    right: 20px;
    width: 50px;
    height: 50px;
  }
  
  .back-to-top-icon {
    font-size: 1rem;
  }
  
  .back-to-top-text {
    font-size: 0.6rem;
  }
}
</style>