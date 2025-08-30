<template>
  <div class="other-page">
    <!-- 页面头部 -->
    <header class="page-header enhanced-header">
      <div class="header-bg"></div>
      <div class="header-content">
        <div class="header-logo">
          <span class="logo-icon">🧩</span>
        </div>
        <div class="header-title-group">
          <h1>详细日志分析</h1>
        </div>
        <button class="btn header-btn" @click="$router.push('/')">返回首页</button>

      </div>
      <div class="header-divider"></div>
    </header>

    <!-- 书签导航 -->
    <div v-if="analysisData.length > 0" class="bookmark-nav">
      <div class="bookmark-header" @click="toggleBookmark">
        <span class="bookmark-title">📑 快速导航</span>
        <!-- <button class="bookmark-toggle"  :class="{ 'active': bookmarkVisible }">
          {{ bookmarkVisible ? '◀' : '▶' }}
        </button> -->
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
              <div class="group-title" @click="toggleGroupDetails(group.GroupName)">
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
                <button class="btn ai-analysis-btn-small" @click="performAIAnalysisForGroup(group.GroupName)" title="AI分析此配置组" :disabled="aiAnalysisLoading">
                  <span v-if="aiAnalysisLoading && currentAnalyzingGroup === group.GroupName" class="ai-loading-text">🔍 分析中...</span>
                  <span v-else>🔍 AI分析</span>
                </button>
                <button class="btn archive-btn-always" @click="archiveGroup(group)" title="归档此配置组">
                  📥 归档
                </button>
                <!-- <button class="toggle-btn" @click="toggleGroupDetails(group.GroupName)">
                  <span v-if="expandedGroups.includes(group.GroupName)" style="color: #ff6eb4;">📖 收起</span>
                  <span v-else style="color: #ff6eb4;">📋 详情</span>
                </button> -->
              </div>
            </div>

            <!-- 卡片内容 - 可折叠 -->
            <transition name="slide-down">
                <div v-if="expandedGroups.includes(group.GroupName)" class="group-content">
                                 <div class="error-section">
                   <h4 class="section-title">❗ 错误汇总</h4>
                   <div class="error-summary" v-html="formatMap(group.ErrorSummary)"></div>
                </div>

                <!-- 收入汇总 -->
                <div  class="group-content" >
                  <h4 class="section-title" style="cursor: pointer;" @click="lookIncome">💰 查询收入汇总</h4>
                  <div class="error-summary income" v-html="formatMap(group.SumIncome)"></div>
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
                           <strong>⚠️ 相关坐标：</strong>
                          <div class="error-content" v-html="formatMap(sub.ErrorsMark)"></div>
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

    <!-- AI分析结果弹窗 -->
    <div v-if="showAIModal" class="ai-modal-overlay" @click="closeAIModal">
      <div class="ai-modal" @click.stop>
                 <div class="ai-modal-header">
           <h3>
             🔍 错误分析报告 - {{ currentAnalyzingGroup || selectedGroup }}
             <span v-if="aiStreaming" class="streaming-status">(正在生成...)</span>
             <span v-else-if="aiAnalysisLoading" class="loading-status">(分析中...)</span>
           </h3>
                      <div class="ai-modal-actions">
             <button class="btn download-btn" @click="downloadAIAnalysisResult" title="下载报告">
               📥 下载
             </button>
             <button class="btn copy-btn" @click="copyAIAnalysisResult" title="复制内容">
               📋 复制
             </button>
             <button class="btn close-btn" @click="closeAIModal" title="关闭">
               ✕
             </button>
           </div>
        </div>
                 <div class="ai-modal-content">
           <div v-if="aiAnalysisLoading && !aiAnalysisResult" class="ai-loading">
             <div class="ai-loading-spinner"></div>
             <p>正在分析错误信息，请稍候...</p>
           </div>
           <div v-else class="ai-result">
             <div v-html="formatMarkdown(aiAnalysisResult)"></div>
             <div v-if="aiStreaming" class="streaming-indicator">
               <span class="typing-dots">正在生成</span>
             </div>
           </div>
         </div>
      </div>
    </div>

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
      selectedGroup: '', // 当前选中的配置组
      analysisData: [],
      loading: false,
      expandedGroups: [], // 记录展开的配置组
      bookmarkVisible: false, // 书签是否可见，默认折叠
      currentActiveGroup: '', // 当前活跃的配置组
      aiAnalysisLoading: false, // AI分析加载状态
      aiAnalysisResult: '', // AI分析结果
      showAIModal: false, // 是否显示AI分析结果弹窗
      aiStreaming: false, // 是否正在流式输出
      currentAnalyzingGroup: '' // 当前正在分析的配置组
    }
  },
  async mounted() {
    await this.loadLogFiles()
    // 初始化防抖滚动函数
    this.debouncedScrollToBottom = this.debounce(this.scrollToBottom, 100)
  },
  watch: {
    // 监听 selectedFile 变化，自动加载分析数据
    selectedFile(newVal, oldVal) {
      if (newVal && newVal !== oldVal) {
        this.loadAnalysisData()
      }
    }
  },
  methods: {
    // 加载日志文件列表
    async loadLogFiles() {
      try {
        const response = await api.get('/api/logFiles')
        this.logFiles = response.files || []
        if (this.logFiles.length > 0) {
          this.selectedFile = this.logFiles[0] // 默认选择最新的文件
          // 不再这里调用 loadAnalysisData，交由 watch 处理
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
      // 点击导航时自动展开导航
      this.bookmarkVisible = true
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
    },
    // 查询收入汇总
    lookIncome(){
      const incomeElements = document.querySelectorAll('.income');
      incomeElements.forEach(el => {
        if (el.style.display === 'none') {
          el.style.display = 'block';
        } else {
          el.style.display = 'none';
        }
      });
    },
    
    // AI分析功能 - 分析指定配置组
    async performAIAnalysisForGroup(groupName) {
      if (this.aiAnalysisLoading) return
      
      this.aiAnalysisLoading = true
      this.currentAnalyzingGroup = groupName
      this.selectedGroup = groupName // 设置当前选中的配置组
      this.aiAnalysisResult = ''
      this.showAIModal = true
      this.aiStreaming = false
      
      try {
        // 获取指定配置组的数据
        const currentGroupData = this.analysisData.find(group => group.GroupName === groupName)
        
        if (!currentGroupData) {
          this.$message?.error('未找到配置组的数据')
          return
        }
        
        // 构建发送给AI的消息
        const analysisDataStr = JSON.stringify(currentGroupData, null, 2)
        const messages = [
          {
            role: 'user',
            content: `请分析以下配置组"${groupName}"的日志数据中的错误信息，生成一份详细的错误分析报表。包括：

1. **错误统计概览**
   - 总错误数量
   - 错误类型分布
2. **错误详情分析**
   - 错误发生的子任务
   - 各类错误的具体描述
   - 错误相关的坐标信息（坐标信息是错误发生的坐标，不是任务的坐标）
请生成一份结构化的Markdown格式错误分析报表，表格形式，不要回答其他东西。

配置组数据：
${analysisDataStr}`
          }
        ]
        
        // 调用流式API
        const controller = new AbortController()
        const timeoutId = setTimeout(() => controller.abort(), 120000) // 2分钟超时
        
        const response = await fetch('/api/chatStream', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ messages }),
          signal: controller.signal
        })
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }
        
        const reader = response.body.getReader()
        const decoder = new TextDecoder()
        let result = ''
        this.aiStreaming = true // 开始流式输出
        
        while (true) {
          const { done, value } = await reader.read()
          if (done) break
          
          const chunk = decoder.decode(value)
          const lines = chunk.split('\n')
          
          for (const line of lines) {
            if (line.startsWith('data: ')) {
              const data = line.slice(6)
              if (data === '[DONE]') {
                break
              }
              try {
                const parsed = JSON.parse(data)
                // 处理你的SSE数据格式
                if (parsed.content && parsed.role !== 'error') {
                  result += parsed.content
                  this.aiAnalysisResult = result
                  // 使用防抖来优化滚动性能
                  this.debouncedScrollToBottom()
                }
              } catch (e) {
                // 忽略解析错误
              }
            } else if (line.startsWith('event: error')) {
              // 处理错误事件
              console.error('SSE错误事件:', line)
              throw new Error('AI分析超时或发生错误')
                //关闭弹窗
                closeAIModal()
            }
          }
        }
        
        this.aiStreaming = false // 结束流式输出
        
        clearTimeout(timeoutId) // 清理超时定时器
        
        this.$message?.success('错误分析完成！')
        
        // 如果结果为空，显示提示
        if (!result.trim()) {
          this.$message?.warning('错误分析结果为空，请重试')
        }
        
      } catch (error) {
        console.error('AI分析失败:', error)
        
        // 根据错误类型显示不同的错误信息
        if (error.name === 'AbortError') {
          this.$message?.error('错误分析超时，请重试')
        } else if (error.message.includes('context deadline exceeded')) {
          this.$message?.error('错误分析超时，请稍后重试')
        } else if (error.message.includes('Failed to fetch')) {
          this.$message?.error('网络连接失败，请检查网络')
        } else {
          this.$message?.error('错误分析失败: ' + error.message)
        }
      } finally {
        this.aiAnalysisLoading = false
        this.currentAnalyzingGroup = ''
      }
    },
    

     
     // 关闭AI分析弹窗
     closeAIModal() {
       this.showAIModal = false
       this.aiAnalysisResult = ''
       this.selectedGroup = '' // 重置选中的配置组
     },
     
     // 复制AI分析结果
     copyAIAnalysisResult() {
       if (this.aiAnalysisResult) {
         navigator.clipboard.writeText(this.aiAnalysisResult).then(() => {
           this.$message?.success('已复制到剪贴板')
         }).catch(() => {
           this.$message?.error('复制失败')
         })
       }
     },
     
           // 下载AI分析结果为HTML
      downloadAIAnalysisResult() {
        if (!this.aiAnalysisResult) {
          this.$message?.error('没有可下载的内容')
          return
        }

        try {
          this.$message?.info('正在生成错误分析报告...')
          
          // 创建完整的HTML文档，包含优化的样式
          const htmlContent = `
            <!DOCTYPE html>
            <html lang="zh-CN">
            <head>
              <meta charset="UTF-8">
              <meta name="viewport" content="width=device-width, initial-scale=1.0">
              <title>错误分析报告 - ${this.currentAnalyzingGroup || this.selectedGroup}</title>
              <style>
                @import url('https://fonts.googleapis.com/css2?family=Noto+Sans+SC:wght@400;500;600;700&display=swap');
                
                * {
                  box-sizing: border-box;
                }
                
                body {
                  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans', Helvetica, Arial, sans-serif;
                  margin: 0;
                  padding: 20px;
                  background: #ffffff;
                  color: #1f2328;
                  line-height: 1.6;
                  font-size: 16px;
                }
                
                .ai-result {
                  background: #ffffff;
                  padding: 30px;
                  border-radius: 6px;
                  border: 1px solid #d0d7de;
                  max-width: 1200px;
                  margin: 0 auto;
                }
                
                h1, h2, h3, h4, h5, h6 {
                  margin: 1em 0 0.6em;
                  font-weight: 600;
                  color: #24292f;
                  border-bottom: 1px solid #d0d7de;
                  padding-bottom: 0.3em;
                }
                
                h1 { 
                  font-size: 2em; 
                }
                
                h2 { 
                  font-size: 1.5em; 
                }
                
                h3 { 
                  font-size: 1.25em; 
                }
                
                h4 { 
                  font-size: 1em; 
                }
                
                p { 
                  margin: 0.6em 0; 
                  line-height: 1.6;
                  color: #1f2328;
                }
                
                ul, ol { 
                  padding-left: 2em; 
                  margin: 0.6em 0;
                }
                
                li {
                  margin: 0.25em 0;
                  line-height: 1.6;
                  color: #1f2328;
                }
                
                code {
                  background: #f6f8fa;
                  color: #1f2328;
                  padding: 0.2em 0.4em;
                  border-radius: 6px;
                  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Consolas, 'Liberation Mono', Menlo, monospace;
                  font-size: 85%;
                  border: 1px solid rgba(175, 184, 193, 0.2);
                }
                
                pre {
                  background: #f6f8fa;
                  border: 1px solid #d0d7de;
                  border-radius: 6px;
                  padding: 16px;
                  overflow: auto;
                  margin: 1em 0;
                }
                
                pre code {
                  background: transparent;
                  padding: 0;
                  border-radius: 0;
                  font-size: 85%;
                  line-height: 1.45;
                  color: #1f2328;
                  border: none;
                }
                
                blockquote {
                  margin: 1em 0;
                  padding: 20px 25px;
                  color: #6b7280;
                  border-left: 5px solid #ff6eb4;
                  background: linear-gradient(135deg, rgba(255, 110, 180, 0.05), rgba(255, 158, 207, 0.05));
                  border-radius: 0 15px 15px 0;
                  position: relative;
                  font-style: italic;
                }
                
                blockquote::before {
                  content: '💭';
                  position: absolute;
                  top: 10px;
                  right: 15px;
                  font-size: 1.5em;
                  opacity: 0.4;
                }
                
                table {
                  border-collapse: collapse;
                  width: 100%;
                  margin: 1em 0;
                  font-size: 85%;
                  line-height: 1.6;
                }
                
                table th, table td {
                  border: 1px solid #d0d7de;
                  padding: 6px 13px;
                  text-align: left;
                }
                
                table th {
                  background: #f6f8fa;
                  color: #1f2328;
                  font-weight: 600;
                }
                
                table tr:nth-child(even) {
                  background: #f6f8fa;
                }
                
                table tr:nth-child(odd) {
                  background: #ffffff;
                }
                
                a { 
                  color: #0969da; 
                  text-decoration: none;
                }
                
                a:hover {
                  text-decoration: underline;
                }
                
                strong {
                  font-weight: 600;
                  color: #1f2328;
                }
                
                em {
                  font-style: italic;
                  color: #1f2328;
                }
                
                hr {
                  height: 0.25em;
                  padding: 0;
                  margin: 24px 0;
                  background: #d0d7de;
                  border: 0;
                }
                
                @media print {
                  body {
                    background: white !important;
                    font-size: 14px;
                  }
                  
                  .ai-result {
                    border: 1px solid #ddd !important;
                    padding: 20px;
                  }
                  
                  h1, h2, h3, h4, h5, h6 {
                    color: #333 !important;
                  }
                  
                  p, li {
                    color: #333 !important;
                  }
                }
                
                @media (max-width: 768px) {
                  body {
                    padding: 10px;
                    font-size: 14px;
                  }
                  
                  .ai-result {
                    padding: 20px;
                    border-radius: 6px;
                  }
                  
                  h1 { font-size: 1.8em; }
                  h2 { font-size: 1.5em; }
                  h3 { font-size: 1.3em; }
                  h4 { font-size: 1.1em; }
                  
                  ul, ol {
                    padding-left: 1.5em;
                  }
                  
                  li {
                    margin: 0.2em 0;
                    font-size: 14px;
                  }
                  
                  table {
                    font-size: 0.8em;
                  }
                  
                  table th, table td {
                    padding: 4px 8px;
                  }
                }
              </style>
            </head>
            <body>
              <div class="ai-result">
                <h1>错误分析报告 - ${this.currentAnalyzingGroup || this.selectedGroup}</h1>
                ${this.formatMarkdown(this.aiAnalysisResult)}
              </div>
            </body>
            </html>
          `
          
          // 检测是否为移动设备
          const isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
          
          if (isMobile) {
            // 移动端下载策略
            this.downloadForMobile(htmlContent)
          } else {
            // 桌面端下载策略
            this.downloadForDesktop(htmlContent)
          }
          
        } catch (error) {
          console.error('生成错误分析报告失败:', error)
          this.$message?.error('生成错误分析报告失败，请重试')
        }
      },

      // 移动端下载方法
      downloadForMobile(htmlContent) {
        try {
          const blob = new Blob([htmlContent], { type: 'text/html;charset=utf-8' })
          const url = URL.createObjectURL(blob)
          
          // 移动端使用新窗口打开，用户可以手动保存
          const newWindow = window.open(url, '_blank')
          if (newWindow) {
            this.$message?.success('报告已在新窗口打开，请手动保存')
          } else {
            // 如果弹窗被阻止，提供备用方案
            this.$message?.info('请复制以下链接到浏览器地址栏打开：')
            console.log('下载链接:', url)
            setTimeout(() => URL.revokeObjectURL(url), 30000) // 30秒后清理
          }
        } catch (error) {
          console.error('移动端下载失败:', error)
          this.$message?.error('移动端下载失败，请使用桌面浏览器')
        }
      },

      // 桌面端下载方法
      downloadForDesktop(htmlContent) {
        try {
          const blob = new Blob([htmlContent], { type: 'text/html;charset=utf-8' })
          const url = URL.createObjectURL(blob)
          const a = document.createElement('a')
          a.href = url
          a.download = `错误分析报告_${this.currentAnalyzingGroup || this.selectedGroup}_${this.selectedFile}_${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.html`
          a.style.display = 'none'
          document.body.appendChild(a)
          a.click()
          document.body.removeChild(a)
          URL.revokeObjectURL(url)
          
          this.$message?.success('错误分析报告已下载')
        } catch (error) {
          console.error('桌面端下载失败:', error)
          this.$message?.error('下载失败，请重试')
        }
      },
     
           // 格式化Markdown内容
      formatMarkdown(markdown) {
        if (!markdown) return ''
        
        let result = markdown
        
        // 处理表格 - 更准确的表格解析
        const tableRegex = /(\|.*\|[\r\n]+)+/g
        result = result.replace(tableRegex, (tableMatch) => {
          console.log('Found table:', tableMatch)
          const lines = tableMatch.trim().split('\n').filter(line => line.trim())
          let tableHtml = '<table>'
          
          lines.forEach((line, index) => {
            const cells = line.split('|').slice(1, -1).map(cell => cell.trim())
            if (cells.length > 0) {
              const tag = index === 0 ? 'th' : 'td' // 第一行作为表头
              const rowHtml = cells.map(cell => `<${tag}>${cell}</${tag}>`).join('')
              tableHtml += `<tr>${rowHtml}</tr>`
            }
          })
          
          tableHtml += '</table>'
          console.log('Generated table HTML:', tableHtml)
          return tableHtml
        })
          
          // 处理标题
          .replace(/^### (.*$)/gim, '<h3>$1</h3>')
          .replace(/^## (.*$)/gim, '<h2>$1</h2>')
          .replace(/^# (.*$)/gim, '<h1>$1</h1>')
          
          // 处理粗体和斜体
          .replace(/\*\*(.*?)\*\*/gim, '<strong>$1</strong>')
          .replace(/\*(.*?)\*/gim, '<em>$1</em>')
          
          // 处理代码
          .replace(/`(.*?)`/gim, '<code>$1</code>')
          
          // 处理列表项
          .replace(/^- (.*$)/gim, '<li>$1</li>')
          .replace(/^\d+\. (.*$)/gim, '<li>$1</li>')
          
          // 处理换行
          .replace(/\n/gim, '<br>')
        
        // 将连续的li标签包装在ul中
        result = result.replace(/(<li>.*?<\/li>)+/g, '<ul>$&</ul>')
        
        return result
      },
      
      // 防抖滚动到底部
      debouncedScrollToBottom: null,
      
      // 滚动到底部
      scrollToBottom() {
        this.$nextTick(() => {
          const modalContent = document.querySelector('.ai-modal-content')
          if (modalContent) {
            modalContent.scrollTop = modalContent.scrollHeight
          }
        })
      },
      
      // 防抖函数
      debounce(func, wait) {
        let timeout
        return function executedFunction(...args) {
          const later = () => {
            clearTimeout(timeout)
            func(...args)
          }
          clearTimeout(timeout)
          timeout = setTimeout(later, wait)
        }
      }
  }
}
</script>

<style>
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

.page-header.enhanced-header {
  position: relative;
  background: linear-gradient(90deg, #fff6fb 60%, #ff9ecf 100%);
  box-shadow: 0 8px 32px rgba(255, 110, 180, 0.15), 0 2px 8px rgba(255, 110, 180, 0.08);
  border-radius: 0 0 36px 36px;
  padding: 0;
  margin-bottom: 10px;
  overflow: hidden;
  z-index: 10;
}

.header-bg {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: radial-gradient(circle at 20% 40%, #e9a0d1 0%, #ecccde 60%, transparent 100%);
  opacity: 0.7;
  z-index: 0;
}

.header-content {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 32px;
  padding: 36px 0 18px 0;
  z-index: 1;
}

.header-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 10px;
}

.logo-icon {
  font-size: 2.8rem;
  background: linear-gradient(45deg, #ff6eb4, #ff9ecf);
  border-radius: 18px;
  box-shadow: 0 2px 12px rgba(255, 110, 180, 0.18);
  padding: 10px 16px;
  color: #fff;
  border: 2px solid #ffc0da;
}

.header-title-group {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
}

.page-header.enhanced-header h1 {
  color: #ff6eb4;
  font-size: 2.2rem;
  font-weight: bold;
  margin: 0;
  text-shadow: 0 2px 12px #ffc0da;
  letter-spacing: 2px;
}



.header-btn {
  margin-top: 0;
  margin-left: auto;
  font-size: 1rem;
  padding: 10px 22px;
  border-radius: 30px;
  box-shadow: 0 2px 8px #ffc0da;
  background: linear-gradient(45deg, #fff, #ffe3f3);
  color: #ff6eb4;
  border: 2px solid #ff6eb4;
  font-weight: bold;
  transition: all 0.3s;
}

.header-btn:hover {
  background: linear-gradient(45deg, #ff6eb4, #ff9ecf);
  color: #fff;
  box-shadow: 0 4px 16px #ff9ecf;
  transform: scale(1.07);
}



.ai-analysis-btn-small {
  margin-right: 8px;
  font-size: 0.8rem;
  padding: 6px 12px;
  border-radius: 20px;
  box-shadow: 0 2px 6px #ffc0da;
  background: linear-gradient(45deg, #4CAF50, #66BB6A);
  color: white;
  border: 1px solid #4CAF50;
  font-weight: 500;
  transition: all 0.3s;
}

.ai-analysis-btn-small:hover:not(:disabled) {
  background: linear-gradient(45deg, #66BB6A, #4CAF50);
  color: #fff;
  box-shadow: 0 3px 12px #66BB6A;
  transform: scale(1.05);
}

.ai-analysis-btn-small:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

.ai-loading-text {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.ai-loading-text::after {
  content: '';
  width: 12px;
  height: 12px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top: 2px solid white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.header-divider {
  width: 80%;
  height: 4px;
  margin: 0 auto 0 auto;
  background: linear-gradient(90deg, #ff6eb4 0%, #ff9ecf 100%);
  border-radius: 2px;
  box-shadow: 0 2px 8px #ffc0da;
  opacity: 0.25;
  margin-bottom: 2px;
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
  cursor: pointer;
}

.bookmark-title {
  font-size: 0.85rem;
  color: #000;
}

/* .bookmark-toggle {
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
} */

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
  color: #14050c;
}

.stats-label {
  font-size: 0.9rem;
  color: #f4ecf0;
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
  cursor: pointer;
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
  color: #f7eef2;
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

.income{
  display: none;
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

/* AI分析弹窗样式 */
.ai-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  padding: 20px;
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
}

.ai-modal {
  background: linear-gradient(135deg, #ffffff, #fefcff);
  border-radius: 24px;
  box-shadow: 0 20px 60px rgba(255, 110, 180, 0.3);
  max-width: 90vw;
  max-height: 90vh;
  width: 800px;
  display: flex;
  flex-direction: column;
  border: 2px solid var(--border-color);
  overflow: hidden;
}

.ai-modal-header {
  background: linear-gradient(45deg, var(--primary-color), #ff9ecf);
  color: white;
  padding: 20px 25px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-color);
}

.ai-modal-header h3 {
  margin: 0;
  font-size: 1.4rem;
  font-weight: bold;
}

.ai-modal-actions {
  display: flex;
  gap: 10px;
}

.download-btn {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.3);
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.3s;
}

.download-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-1px);
}

.copy-btn {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.3);
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.3s;
}

.copy-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-1px);
}

.close-btn {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.3);
  padding: 8px 12px;
  border-radius: 50%;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.3s;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: scale(1.1);
}

.ai-modal-content {
  flex: 1;
  padding: 25px;
  overflow-y: auto;
  max-height: calc(90vh - 100px);
}

.ai-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: var(--primary-color);
}

.ai-loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid rgba(255, 110, 180, 0.2);
  border-top: 4px solid var(--primary-color);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 20px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

/* GitHub风格 Markdown 样式 */
.ai-result {
  color: #1f2328;
  line-height: 1.6;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans', Helvetica, Arial, sans-serif;
  background: #ffffff;
  padding: 25px;
  border-radius: 6px;
  border: 1px solid #d0d7de;
  font-size: 16px;
}



.ai-result h1, .ai-result h2, .ai-result h3,
.ai-result h4, .ai-result h5, .ai-result h6 {
  margin: 1em 0 0.6em;
  font-weight: 600;
  color: #24292f;
  border-bottom: 1px solid #d0d7de;
  padding-bottom: 0.3em;
}



.ai-result h1 { 
  font-size: 2em; 
}

.ai-result h2 { 
  font-size: 1.5em; 
}

.ai-result h3 { 
  font-size: 1.25em; 
}

.ai-result h4 { 
  font-size: 1em; 
}

.ai-result p { 
  margin: 0.6em 0; 
  line-height: 1.6;
  color: #1f2328;
}

.ai-result ul, .ai-result ol { 
  padding-left: 2em; 
  margin: 0.6em 0;
}

.ai-result li {
  margin: 0.25em 0;
  line-height: 1.6;
  color: #1f2328;
}

.ai-result code {
  background: #f6f8fa;
  color: #1f2328;
  padding: 0.2em 0.4em;
  border-radius: 6px;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 85%;
  border: 1px solid rgba(175, 184, 193, 0.2);
}

.ai-result pre {
  background: #f6f8fa;
  border: 1px solid #d0d7de;
  border-radius: 6px;
  padding: 16px;
  overflow: auto;
  margin: 1em 0;
}

.ai-result pre code {
  background: transparent;
  padding: 0;
  border-radius: 0;
  font-size: 85%;
  line-height: 1.45;
  color: #1f2328;
  border: none;
}

.ai-result blockquote {
  margin: 1em 0;
  padding: 0 1em;
  color: #656d76;
  border-left: 0.25em solid #d0d7de;
  background: transparent;
}

.ai-result table {
  border-collapse: collapse;
  width: 100%;
  margin: 1em 0;
  font-size: 85%;
  line-height: 1.6;
}

.ai-result table th,
.ai-result table td {
  border: 1px solid #d0d7de;
  padding: 6px 13px;
  text-align: left;
}

.ai-result table th {
  background: #f6f8fa;
  color: #1f2328;
  font-weight: 600;
}

.ai-result table tr:nth-child(even) {
  background: #f6f8fa;
}

.ai-result table tr:nth-child(odd) {
  background: #ffffff;
}

.ai-result a { 
  color: #0969da; 
  text-decoration: none;
}

.ai-result a:hover {
  text-decoration: underline;
}

.ai-result strong {
  font-weight: 600;
  color: #1f2328;
}

.ai-result em {
  font-style: italic;
  color: #1f2328;
}

.ai-result hr {
  height: 0.25em;
  padding: 0;
  margin: 24px 0;
  background: #d0d7de;
  border: 0;
}

/* 流式输出指示器样式 */
.streaming-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 15px;
  padding: 10px 15px;
  background: rgba(255, 110, 180, 0.1);
  border-radius: 20px;
  border: 1px solid rgba(255, 110, 180, 0.2);
}

.typing-dots {
  color: var(--primary-color);
  font-weight: bold;
  position: relative;
}

.typing-dots::after {
  content: '';
  animation: typing 1.5s infinite;
}

@keyframes typing {
  0%, 20% { content: ''; }
  40% { content: '.'; }
  60% { content: '..'; }
  80%, 100% { content: '...'; }
}

/* 状态指示器样式 */
.streaming-status,
.loading-status {
  font-size: 0.9rem;
  font-weight: normal;
  opacity: 0.8;
  margin-left: 10px;
}

.streaming-status {
  color: #4CAF50;
  animation: pulse 1.5s infinite;
}

.loading-status {
  color: #FF9800;
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

  /* 移动端AI分析按钮适配 */
  .header-content {
    flex-direction: column;
    gap: 15px;
  }

  .ai-analysis-btn {
    margin-left: 0;
    width: 100%;
    max-width: 200px;
  }

  /* 移动端AI弹窗适配 */
  .ai-modal {
    width: 95vw;
    max-height: 95vh;
    border-radius: 16px;
  }

  .ai-modal-header {
    padding: 15px 20px;
    flex-direction: column;
    gap: 10px;
    align-items: stretch;
  }

  .ai-modal-header h3 {
    font-size: 1.1rem;
    text-align: center;
    line-height: 1.3;
  }

  .ai-modal-actions {
    justify-content: center;
    flex-wrap: wrap;
    gap: 8px;
  }

  .download-btn,
  .copy-btn {
    padding: 6px 12px;
    font-size: 0.8rem;
    min-width: 60px;
  }

  .close-btn {
    width: 32px;
    height: 32px;
    font-size: 0.9rem;
  }

  .ai-modal-content {
    padding: 15px;
    max-height: calc(95vh - 120px);
  }

  .ai-result {
    padding: 15px;
    border-radius: 12px;
    font-size: 14px;
  }

  /* 移动端触摸优化 */
  .ai-modal-overlay {
    padding: 10px;
    align-items: flex-start;
    padding-top: 5vh;
  }

  .ai-modal {
    animation: slideInUp 0.3s ease-out;
  }

  @keyframes slideInUp {
    from {
      opacity: 0;
      transform: translateY(50px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  /* 移动端按钮触摸优化 */
  .download-btn,
  .copy-btn,
  .close-btn {
    -webkit-tap-highlight-color: transparent;
    touch-action: manipulation;
  }

  .download-btn:active,
  .copy-btn:active,
  .close-btn:active {
    transform: scale(0.95);
  }

  .ai-result h1 {
    font-size: 1.6em;
    padding-bottom: 0.3em;
  }

  .ai-result h2 {
    font-size: 1.4em;
    padding-bottom: 0.2em;
  }

  .ai-result h3 {
    font-size: 1.2em;
  }

  .ai-result h4 {
    font-size: 1.1em;
  }

  .ai-result p {
    font-size: 0.9em;
    line-height: 1.6;
    margin: 0.5em 0;
  }

  .ai-result ul, .ai-result ol {
    padding: 10px 10px 10px 1.5em;
    margin: 0.6em 0;
  }

  .ai-result li {
    padding: 4px 8px;
    margin: 0.3em 0;
    font-size: 0.85em;
    line-height: 1.5;
  }

  .ai-result code {
    font-size: 0.8em;
    padding: 2px 4px;
    word-break: break-all;
  }

  .ai-result pre {
    padding: 12px;
    margin: 0.6em 0;
    font-size: 0.8em;
    overflow-x: auto;
  }

  .ai-result blockquote {
    padding: 12px 15px;
    margin: 0.6em 0;
    font-size: 0.9em;
  }

  .ai-result table {
    font-size: 0.75em;
    margin: 0.6em 0;
    overflow-x: auto;
    display: block;
  }

  .ai-result table th,
  .ai-result table td {
    padding: 6px 8px;
    min-width: 80px;
  }

  .ai-result hr {
    margin: 15px 0;
  }
}
</style>