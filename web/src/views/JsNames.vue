<template>
  <div class="js-names-page">
    <!-- Header轮播图 -->
    <header class="page-header" v-if="headerCarouselImages.length > 0">
      <div class="header-carousel">
        <div class="carousel-container">
          <div v-for="(image, index) in headerCarouselImages" :key="index" class="carousel-slide" :class="{ active: headerCurrentImageIndex === index }">
            <img :src="image" :alt="`header-carousel-${index}`" />
          </div>
        </div>
      </div>
      <div class="header-content">

        <h1 class="header-title">📜 脚本更新列表 📜</h1>
        <p class="header-subtitle">管理您的脚本，保持最新状态 ✨</p>
      </div>
    </header>

    <div class="container">
      <section class="panel">
        <h2>脚本信息   <button class="btn home-btn" @click="goHome">返回首页</button></h2>
             
        <div id="pluginListContainer" class="table-container">
          <!-- 桌面端表格 -->
          <table id="pluginTable" class="desktop-table">
            <thead>
              <tr>
                <th data-key="ChineseName" @click="sortTable('ChineseName')" class="sortable">
                  <span>脚本中文名</span>
                  <i class="sort-icon" :class="getSortIcon('ChineseName')"></i>
                </th>
                <th data-key="NowVersion" @click="sortTable('NowVersion')" class="sortable">
                  <span>当前版本</span>
                  <i class="sort-icon" :class="getSortIcon('NowVersion')"></i>
                </th>
                <th data-key="NewVersion" @click="sortTable('NewVersion')" class="sortable">
                  <span>最新版本</span>
                  <i class="sort-icon" :class="getSortIcon('NewVersion')"></i>
                </th>
                <th data-key="Mark" @click="sortTable('Mark')" class="sortable">
                  <span>状态</span>
                  <i class="sort-icon" :class="getSortIcon('Mark')"></i>
                </th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody id="pluginTableBody">
              <tr v-if="pluginData.length === 0">
                <td colspan="5" style="text-align:center;">暂无插件数据。</td>
              </tr>
              <tr v-else v-for="item in sortedPluginData" :key="item.Name" 
                  :class="{ highlight: item.Mark === '有更新' }">
                <td>{{ item.ChineseName }}</td>
                <td>{{ item.NowVersion }}</td>
                <td>{{ item.NewVersion }}</td>
                <td>{{ item.Mark }}</td>
                <td>
                  <button 
                    class="btn update-btn" 
                    :disabled="item.Mark !== '有更新' || isUpdating[item.Name]"
                    @click="updatePlugin(item.Name)"
                  >
                    {{ isUpdating[item.Name] ? '更新中...' : (item.Mark === '有更新' ? '更新' : '已更新') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>

          <!-- 移动端卡片列表 -->
          <div class="mobile-list">
            <div v-if="pluginData.length === 0" class="empty-mobile">
              <div class="empty-content">
                <span class="empty-icon">📭</span>
                <span class="empty-text">暂无插件数据</span>
                <span class="empty-sparkle">✨</span>
              </div>
            </div>
            <div v-else v-for="item in sortedPluginData" :key="item.Name" 
                 :class="{ highlight: item.Mark === '有更新', 'mobile-card': true }">
              <div class="card-header">
                <div class="card-title">
                  <span class="title-icon">📝</span>
                  <span class="title-text">{{ item.ChineseName }}</span>
                </div>
                <div class="card-versions">
                  <div class="version-item">
                    <span class="version-label">当前版本:</span>
                    <span class="version-value">{{ item.NowVersion }}</span>
                  </div>
                  <div class="version-item">
                    <span class="version-label">最新版本:</span>
                    <span class="version-value">{{ item.NewVersion }}</span>
                  </div>
                </div>
                <div class="card-status">
                  <span class="status-icon">📊</span>
                  <span class="status-text" :class="{ 'status-update': item.Mark === '有更新' }">{{ item.Mark }}</span>
                </div>
              </div>
              <div class="card-actions">
                <button 
                  class="btn update-btn mobile-update-btn" 
                  :disabled="item.Mark !== '有更新' || isUpdating[item.Name]"
                  @click="updatePlugin(item.Name)"
                >
                  <span class="update-icon">{{ isUpdating[item.Name] ? '⏳' : '🔄' }}</span>
                  <span class="update-text">{{ isUpdating[item.Name] ? '更新中...' : (item.Mark === '有更新' ? '更新' : '已更新') }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="panel">
        <h2>仓库提交记录</h2>
        <div id="gitLogContainer" class="table-container git-log-container">
          <!-- 桌面端表格 -->
          <table id="gitLogTable" class="desktop-table">
            <thead>
              <tr>
                <th>提交时间</th>
                <th>作者</th>
                <th>提交信息</th>
                <th>相关文件</th>
              </tr>
            </thead>
            <tbody id="gitLogBody">
              <tr v-if="gitLogs.length === 0">
                <td colspan="4" style="text-align:center;">
                  {{ gitLogLoading ? '加载中...' : '暂无提交记录。' }}
                </td>
              </tr>
              <tr v-else v-for="log in gitLogs" :key="log.CommitTime + log.Author">
                <td>{{ log.CommitTime }}</td>
                <td>{{ log.Author }}</td>
                <td v-html="log.Message.replace(/\n/g, '<br/>')"></td>
                <td>
                  <ul v-if="log.Files && log.Files.length > 0" style="padding-left:10px;margin:0;font-size: 32px;">
                    <li v-for="file in log.Files" :key="file" style="margin-left:15px;font-size: 16px;">🎉  {{ file }}</li>
                  </ul>
                  <ul v-else style="padding-left:10px;margin:0;">
                    <li>无文件</li>
                  </ul>
                </td>
              </tr>
            </tbody>
          </table>

          <!-- 移动端卡片列表 -->
          <div class="mobile-list">
            <div v-if="gitLogs.length === 0" class="empty-mobile">
              <div class="empty-content">
                <span class="empty-icon">📭</span>
                <span class="empty-text">{{ gitLogLoading ? '加载中...' : '暂无提交记录' }}</span>
                <span class="empty-sparkle">✨</span>
              </div>
            </div>
            <div v-else v-for="(log,index) in gitLogs" :key="log.CommitTime + log.Author" class="mobile-card">
              <div class="card-header">
                <h3 style="margin-bottom: 4px;">#{{index+1 }}</h3>
                <div class="card-time">
              
                  <span class="time-icon">⏰</span>
                  <span class="time-text">{{ log.CommitTime }}</span>
                </div>
                <div class="card-author">
                  <span class="author-icon">👤</span>
                  <span class="author-text">{{ log.Author }}</span>
                </div>
              </div>
              <div class="card-message">
                <span class="message-icon">💬</span>
                <div class="message-text" v-html="log.Message.replace(/\n/g, '<br/>')"></div>
              </div>
              <div class="card-files">
                <span class="files-icon">📁</span>
                <div class="files-content">
                  <ul v-if="log.Files && log.Files.length > 0">
                    <li v-for="file in log.Files" :key="file">{{ file }}</li>
                  </ul>
                  <span v-else class="no-files">无文件</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'

export default {
  name: 'JsNames',
  setup() {
    const router = useRouter()
    const pluginData = ref([])
    const gitLogs = ref([])
    const gitLogLoading = ref(true)
    const currentSort = ref({ key: 'ChineseName', asc: true })
    const isUpdating = reactive({})
    
    // Header轮播图相关
    const headerCarouselImages = ref([])
    const headerCurrentImageIndex = ref(0)
    let headerCarouselInterval = null

    // 获取header轮播图图片
    const getHeaderImages = async () => {
      try {
        const response = await fetch('/api/images')
        if (!response.ok) {
          throw new Error('Failed to fetch header images')
        }
        const data = await response.json()
        console.log('Header轮播图数据:', data)
        headerCarouselImages.value = data.images || []
        
        // 启动header轮播
        if (headerCarouselImages.value.length > 0) {
          console.log('Header轮播图数量:', headerCarouselImages.value.length)
          startHeaderCarousel()
        }
      } catch (error) {
        console.error('获取Header轮播图失败:', error)
        // 如果API失败，使用默认图片
        headerCarouselImages.value = ['/img/bd.jpg', '/img/ff.png', '/img/ng.jpg', '/img/sh.jpg']
        startHeaderCarousel()
      }
    }

    // 启动header轮播
    const startHeaderCarousel = () => {
      console.log('启动Header轮播，图片数量:', headerCarouselImages.value.length)
      if (headerCarouselImages.value.length > 1) {
        headerCarouselInterval = setInterval(() => {
          headerCurrentImageIndex.value = (headerCurrentImageIndex.value + 1) % headerCarouselImages.value.length
          console.log('切换到Header图片:', headerCurrentImageIndex.value)
        }, 7000) // 每7秒切换一张图片
      }
    }

    const sortedPluginData = computed(() => {
      const sorted = [...pluginData.value]
      sorted.sort((a, b) => {
        const valA = String(a[currentSort.value.key] || '')
        const valB = String(b[currentSort.value.key] || '')
        return currentSort.value.asc ? valA.localeCompare(valB) : valB.localeCompare(valA)
      })
      return sorted
    })

    const goHome = () => {
      router.push('/')
    }

    const sortTable = (key) => {
      if (currentSort.value.key === key) {
        currentSort.value.asc = !currentSort.value.asc
      } else {
        currentSort.value.key = key
        currentSort.value.asc = true
      }
    }

    const getSortIcon = (key) => {
      if (currentSort.value.key !== key) {
        return 'sort-default'
      }
      return currentSort.value.asc ? 'sort-asc' : 'sort-desc'
    }

    const loadPluginList = async () => {
      try {
        const response = await fetch('/api/jsNames')
        const json = await response.json()
        pluginData.value = json.data || []
      } catch (error) {
        console.error('加载插件列表失败：', error)
        pluginData.value = []
      }
    }

    const loadGitLog = async () => {
      try {
        gitLogLoading.value = true
        const response = await fetch('/api/gitLog')
        const json = await response.json()
        gitLogs.value = json.gitLog || []
      } catch (error) {
        console.error('加载提交记录失败：', error)
        gitLogs.value = []
      } finally {
        gitLogLoading.value = false
      }
    }

    const updatePlugin = async (pluginName) => {
  isUpdating[pluginName] = true

  try {
    const response = await fetch('/api/updateJs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: pluginName })
    })
    const result = await response.json()

    if (result.success) {
      await loadPluginList() // ✅ 重新加载插件列表，插件数据会更新
    } else {
      throw new Error(result.message || '更新失败')
    }
  } catch (error) {
    alert('更新失败：' + error.message)
  } finally {
    isUpdating[pluginName] = false
  }
}

    onMounted(() => {
      loadPluginList()
      loadGitLog()
      getHeaderImages() // 在组件挂载时获取header轮播图
    })

    return {
      pluginData,
      gitLogs,
      gitLogLoading,
      sortedPluginData,
      isUpdating,
      goHome,
      sortTable,
      updatePlugin,
      getSortIcon,
      headerCarouselImages, // 暴露header轮播图数据
      headerCurrentImageIndex // 暴露header轮播图当前图片索引
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
  --grid-color: rgba(255, 182, 193, 0.1);
}

* { 
  box-sizing: border-box; 
  margin: 0; 
  padding: 0; 
}

/* ============ Header轮播图样式 ============ */
.page-header {
  position: relative;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95) 0%, rgba(255, 246, 251, 0.9) 100%);
  padding: 40px 0 30px;
  text-align: center;
  box-shadow: 0 8px 32px rgba(255, 110, 180, 0.15);
  border-radius: 0 0 40px 40px;
  margin-bottom: 20px;
  overflow: hidden;
}

.header-carousel {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  z-index: -1;
  border-radius: 0 0 40px 40px;
}

.carousel-container {
  position: relative;
  width: 100%;
  height: 100%;
}

.carousel-slide {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  transition: opacity 1.5s ease-in-out;
}

.carousel-slide.active {
  opacity: 1;
}

.carousel-slide img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 0 0 40px 40px;
}

/* 添加渐变遮罩，确保文字可读性 */
.page-header::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.8) 0%,
    rgba(255, 246, 251, 0.7) 50%,
    rgba(255, 214, 235, 0.6) 100%
  );
  z-index: 0;
  border-radius: 0 0 40px 40px;
}

.header-content {
  position: relative;
  z-index: 1;
}

.header-title {
  color: #ff6eb4;
  font-size: 2.5rem;
  text-shadow: 0 0 20px rgba(255, 110, 180, 0.4);
  margin: 20px 0 10px;
  animation: titleGlow 3s infinite ease-in-out;
}

@keyframes titleGlow {
  0%, 100% {
    text-shadow: 0 0 20px rgba(255, 110, 180, 0.4);
  }
  50% {
    text-shadow: 0 0 30px rgba(255, 110, 180, 0.4), 0 0 40px #ff6eb4;
  }
}

.header-subtitle {
  font-size: 1.1rem;
  color: #e91e63;
  margin-top: 10px;
  opacity: 0.8;
}

.home-btn {
  position: absolute;
  top: 20px;
  left: 20px;
  z-index: 2;
  background: linear-gradient(135deg, #fff 0%, #fff6fb 100%);
  color: #ff6eb4;
  border: 2px solid #ff6eb4;
  border-radius: 50px;
  padding: 12px 24px;
  font-size: 1rem;
  cursor: pointer;
  box-shadow: 0 8px 32px rgba(255, 110, 180, 0.15);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  font-weight: bold;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  backdrop-filter: blur(5px);
}

.home-btn:hover {
  background: linear-gradient(135deg, #ff6eb4 0%, #ff8cc8 100%);
  color: rgb(255, 255, 255);
  box-shadow: 0 12px 40px rgba(255, 110, 180, 0.4);
  transform: translateY(-3px) scale(1.05);
}

/* ============ 基础样式 ============ */
.js-names-page {
  min-height: 100vh;
  background: 
    linear-gradient(90deg, var(--grid-color) 1px, transparent 1px),
    linear-gradient(0deg, var(--grid-color) 1px, transparent 1px);
  background-size: 20px 20px;
  background-color: var(--background-light);
  background-image: 
    radial-gradient(circle at 20px 20px, rgba(255, 214, 235, 0.3) 2px, transparent 2px),
    radial-gradient(circle at 70px 70px, rgba(255, 192, 218, 0.4) 3px, transparent 3px),
    linear-gradient(90deg, var(--grid-color) 1px, transparent 1px),
    linear-gradient(0deg, var(--grid-color) 1px, transparent 1px);
  background-size: 100px 100px, 100px 100px, 20px 20px, 20px 20px;
  background-position: 0 0, 0 0, 0 0, 0 0;
}

body {
  font-family: "Comic Sans MS", "Segoe UI", sans-serif;
  color: var(--text-color);
  padding-bottom: 50px;
}

header {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.9), rgba(255, 240, 250, 0.9));
  padding: 30px 0 10px;
  text-align: center;
  box-shadow: 0 4px 20px rgba(255, 110, 180, 0.3);
  border-radius: 0 0 30px 30px;
  position: sticky;
  top: 0;
  z-index: 10;
  backdrop-filter: blur(10px);
  border-bottom: 2px solid rgba(255, 110, 180, 0.2);
}

h1 {
  color: var(--primary-color);
  font-size: 2rem;
  text-shadow: 0 2px 10px rgba(255, 110, 180, 0.3);
  margin-top: 15px;
  background: linear-gradient(45deg, #ff6eb4, #ff8e8e);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.btn {
  color: var(--primary-color);
  border: 2px solid var(--primary-color);
  border-radius: 50px;
  padding: 8px 16px;
  font-size: 0.9rem;
  cursor: pointer;
  box-shadow: 0 4px 15px rgba(255, 110, 180, 0.2);
  transition: all 0.3s ease;
  font-weight: bold;
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(243, 5, 104, 0.3);
}

.btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.4), transparent);
  transition: left 0.5s;
}

.btn:hover::before {
  left: 100%;
}

.btn:hover {
  box-shadow: 0 6px 20px rgba(255, 110, 180, 0.4);
  transform: translateY(-2px);
}

.container {
  max-width: 1600px;
  margin: 30px auto;
  padding: 0 20px;
}

section.panel {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.9), rgba(255, 250, 255, 0.9));
  box-shadow: 0 8px 25px rgba(255, 204, 230, 0.3);
  border-radius: 20px;
  padding: 25px 30px;
  margin-bottom: 30px;
  border: 1px solid rgba(255, 192, 218, 0.3);
  backdrop-filter: blur(10px);
  position: relative;
  overflow: hidden;
}

section.panel::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, #ff6eb4, #ff8e8e, #ff6eb4);
  background-size: 200% 100%;
  animation: shimmer 3s ease-in-out infinite;
}

@keyframes shimmer {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

section.panel h2 {
  color: var(--primary-color);
  font-size: 1.6rem;
  margin-bottom: 20px;
  border-bottom: 2px solid var(--primary-color);
  padding-bottom: 8px;
  display: inline-block;
  position: relative;
}

section.panel h2::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  width: 100%;
  height: 2px;
  background: linear-gradient(90deg, var(--primary-color), transparent);
}

.table-container {
  border-radius: 15px;
  overflow: hidden;
  box-shadow: 0 4px 15px rgba(255, 182, 226, 0.2);
  background: rgba(255, 255, 255, 0.7);
}

.git-log-container {
  overflow-y: auto;
}

.git-log-container::-webkit-scrollbar {
  width: 8px;
}

.git-log-container::-webkit-scrollbar-track {
  background: rgba(255, 182, 193, 0.1);
  border-radius: 4px;
}

.git-log-container::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, var(--primary-color), #ff8e8e);
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.git-log-container::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, #ff8e8e, var(--primary-color));
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 10px;
  background: rgba(255, 255, 255, 0.8);
}

th, td {
  border: 3px solid rgba(222, 32, 111, 0.4);
  padding: 12px 15px;
  text-align: left;
  position: relative;
}

th {
  background: linear-gradient(135deg, rgba(255, 182, 193, 0.3), rgba(255, 192, 218, 0.2));
  font-weight: bold;
  cursor: pointer;
  color: var(--primary-color);
  font-size: 0.95rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

th.sortable {
  position: relative;
  user-select: none;
  transition: all 0.3s ease;
}

th.sortable:hover {
  background: linear-gradient(135deg, rgba(255, 182, 193, 0.4), rgba(255, 192, 218, 0.3));
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(255, 110, 180, 0.2);
}

.sort-icon {
  margin-left: 8px;
  font-size: 14px;
  transition: all 0.3s ease;
}

.sort-default::before { content: '↕️'; opacity: 0.5; }
.sort-asc::before { content: '⬆️'; opacity: 1; }
.sort-desc::before { content: '⬇️'; opacity: 1; }

tr {
  transition: all 0.3s ease;
}

tr:hover {
  background: rgba(107, 226, 205, 0.6);
  transform: scale(1.01);
  box-shadow: 3px 2px 10px rgba(251, 9, 46, 0.2);
}

tr.highlight {
  background: linear-gradient(135deg, rgba(255, 105, 180, 0.15), rgba(255, 182, 193, 0.1));
  animation: glow 2s infinite alternate;
  border-left: 4px solid var(--primary-color);
}

@keyframes glow {
  from { 
    box-shadow: 0 0 10px rgba(255, 105, 180, 0.3);
    background: linear-gradient(135deg, rgba(255, 105, 180, 0.15), rgba(255, 182, 193, 0.1));
  }
  to { 
    box-shadow: 0 0 20px rgba(255, 105, 180, 0.5);
    background: linear-gradient(135deg, rgba(255, 105, 180, 0.2), rgba(255, 182, 193, 0.15));
  }
}

td {
  color: #333;
  font-size: 0.9rem;
  line-height: 1.4;
}

#gitLogTable td ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

#gitLogTable td li {
  padding: 3px 0;
  color: var(--primary-color);
  font-size: 0.85rem;
  border-bottom: 1px solid rgba(255, 192, 218, 0.2);
  transition: all 0.3s ease;
}

#gitLogTable td li:hover {
  background: rgba(255, 240, 250, 0.5);
  padding-left: 5px;
  border-radius: 3px;
}

#gitLogTable td li:last-child {
  border-bottom: none;
}

.update-btn {
  background: linear-gradient(135deg, #fff, #fff6fb);
  color: var(--primary-color);
  border: 2px solid var(--primary-color);
  border-radius: 25px;
  padding: 6px 12px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.3s ease;
  font-weight: bold;
  min-width: 60px;
  position: relative;
  overflow: hidden;
}

.update-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  background: #f5f5f5;
  color: #999;
  border-color: #ddd;
}

.update-btn:not(:disabled):hover {
  background: linear-gradient(135deg, var(--primary-color), #ff8e8e);
  color: #fff;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(255, 110, 180, 0.3);
}

/* 默认隐藏移动端列表 */
.mobile-list {
  display: none;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .container {
    margin: 20px auto;
    padding: 0 15px;
  }
  
  .page-header {
    padding: 25px 0 20px;
    border-radius: 0 0 30px 30px;
  }
  
  .header-title {
    font-size: 2rem;
  }
  
  .header-subtitle {
    font-size: 1rem;
  }
  
  .home-btn {
    top: 15px;
    left: 15px;
    padding: 10px 20px;
    font-size: 0.9rem;
  }
  
  /* 桌面端表格隐藏 */
  .desktop-table {
    display: none;
  }
  
  /* 移动端卡片显示 */
  .mobile-list {
    display: block;
  }
  
  /* 移动端卡片样式 */
  .mobile-card {
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.95) 0%, rgba(255, 246, 251, 0.9) 100%);
    border-radius: 20px;
    padding: 20px;
    margin-bottom: 25px;
    box-shadow: 0 8px 32px rgba(255, 110, 180, 0.15);
    border: 3px solid rgba(180, 32, 248, 0.3);
    backdrop-filter: blur(10px);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    position: relative;
    overflow: hidden;
  }
  
  .mobile-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 4px;
    background: linear-gradient(90deg, var(--primary-color), var(--secondary-color), var(--accent-color));
    transform: scaleX(0);
    transition: transform 0.3s ease;
  }
  
  .mobile-card:hover::before {
    transform: scaleX(1);
  }
  
  .mobile-card:hover {
    transform: translateY(-5px) scale(1.02);
    background-color: #ff8e8e;
    box-shadow: 0 15px 40px rgba(255, 110, 180, 0.25);
  }
  
  .mobile-card.highlight {
    background: linear-gradient(135deg, rgba(255, 105, 180, 0.15), rgba(255, 182, 193, 0.1));
    animation: glow 2s infinite alternate;
    border-left: 4px solid var(--primary-color);
  }
  
  .card-header {
    margin-bottom: 15px;
  }
  
  .card-title {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 12px;
  }
  
  .title-icon {
    font-size: 1.2rem;
  }
  
  .title-text {
    font-size: 1.1rem;
    font-weight: bold;
    color: var(--text-dark);
  }
  
  .card-versions {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 12px;
  }
  
  .version-item {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  
  .version-label {
    font-size: 0.9rem;
    color: var(--text-color);
    min-width: 70px;
  }
  
  .version-value {
    font-size: 0.9rem;
    color: var(--text-dark);
    font-weight: 500;
  }
  
  .card-status {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  
  .status-icon {
    font-size: 1rem;
  }
  
  .status-text {
    font-size: 0.9rem;
    color: var(--text-color);
  }
  
  .status-update {
    color: #ff6b6b;
    font-weight: bold;
  }
  
  .card-actions {
    text-align: center;
  }
  
  .mobile-update-btn {
    width: 100%;
    justify-content: center;
    padding: 12px;
    font-size: 0.9rem;
  }
  
  /* Git日志移动端卡片 */
  .card-time {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 10px;
  }
  
  .time-icon {
    font-size: 1rem;
  }
  
  .time-text {
    font-size: 0.9rem;
    color: var(--text-dark);
    font-weight: 500;
  }
  
  .card-author {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
  }
  
  .author-icon {
    font-size: 1rem;
  }
  
  .author-text {
    font-size: 0.9rem;
    color: var(--text-color);
  }
  
  .card-message {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;
  }
  
  .message-icon {
    font-size: 1rem;
    flex-shrink: 0;
  }
  
  .message-text {
    font-size: 0.9rem;
    color: var(--text-dark);
    line-height: 1.4;
    flex: 1;
  }
  
  .card-files {
    display: flex;
    gap: 8px;
  }
  
  .files-icon {
    font-size: 1rem;
    flex-shrink: 0;
  }
  
  .files-content {
    flex: 1;
  }
  
  .files-content ul {
    margin: 0;
    padding-left: 15px;
    font-size: 0.8rem;
    color: var(--text-color);
  }
  
  .files-content li {
    margin-bottom: 2px;
  }
  
  .no-files {
    font-size: 0.8rem;
    color: var(--text-color);
    font-style: italic;
  }
  
  /* 空状态样式 */
  .empty-mobile {
    text-align: center;
    padding: 40px 20px;
  }
  
  .empty-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 15px;
  }
  
  .empty-icon {
    font-size: 3rem;
    opacity: 0.6;
  }
  
  .empty-text {
    font-size: 1.1rem;
    color: var(--text-color);
  }
  
  .empty-sparkle {
    font-size: 1.5rem;
    animation: sparkle 2s infinite ease-in-out;
  }
  
  @keyframes sparkle {
    0%, 100% {
      transform: scale(1) rotate(0deg);
      opacity: 0.7;
    }
    50% {
      transform: scale(1.3) rotate(180deg);
      opacity: 1;
    }
  }
  
  .table-container {
    overflow-x: visible;
  }
  
  .git-log-container {
    max-height: none;
  }
}

@media (max-width: 480px) {
  /* 小屏幕Header轮播图适配 */
  .page-header {
    border-radius: 0 0 15px 15px;
    padding: 25px 0 15px;
  }
  
  .header-carousel {
    border-radius: 0 0 15px 15px;
  }
  
  .carousel-slide img {
    border-radius: 0 0 15px 15px;
  }
  
  .page-header::before {
    border-radius: 0 0 15px 15px;
  }
  
  .header-title {
    font-size: 1.8rem;
  }
  
  .header-subtitle {
    font-size: 0.9rem;
  }
  
  .home-btn {
    top: 10px;
    left: 10px;
    padding: 8px 16px;
    font-size: 0.8rem;
  }
  
  /* 小屏幕卡片优化 */
  .mobile-card {
    padding: 15px;
    margin-bottom: 12px;
  }
  
  .title-text {
    font-size: 1rem;
  }
  
  .version-label {
    font-size: 0.85rem;
    min-width: 60px;
  }
  
  .version-value {
    font-size: 0.85rem;
  }
  
  .status-text {
    font-size: 0.85rem;
  }
  
  .mobile-update-btn {
    padding: 10px;
    font-size: 0.85rem;
  }
  
  .time-text {
    font-size: 0.85rem;
  }
  
  .author-text {
    font-size: 0.85rem;
  }
  
  .message-text {
    font-size: 0.85rem;
  }
  
  .files-content ul {
    font-size: 0.75rem;
  }
  
  .no-files {
    font-size: 0.75rem;
  }
  
  .empty-icon {
    font-size: 2.5rem;
  }
  
  .empty-text {
    font-size: 1rem;
  }
}

/* 横屏模式优化 */
@media (max-width: 768px) and (orientation: landscape) {
  .container {
    margin: 10px auto;
  }
  
  section.panel {
    margin-bottom: 10px;
  }
  
  .git-log-container {
    max-height: 250px;
  }
}

/* 触摸优化 */
@media (pointer: coarse) {
  .btn, .update-btn {
    min-height: 44px;
    min-width: 44px;
  }
  
  th.sortable {
    min-height: 44px;
    padding: 12px 8px;
  }
  
  .update-btn {
    margin: 2px;
  }
}

/* 高分辨率屏幕优化 */
@media (min-resolution: 2dppx) {
  .js-names-page {
    background-size: 10px 10px, 10px 10px, 10px 10px, 10px 10px;
  }
}

/* 深色模式支持 */
@media (prefers-color-scheme: dark) {
  :root {
    --background-light: #1a1a1a;
    --text-color: #ff8e8e;
    --border-color: #ff8e8e;
  }
  
  .js-names-page {
    background-color: var(--background-light);
  }
  
  section.panel {
    background: linear-gradient(135deg, rgba(40, 40, 40, 0.9), rgba(50, 50, 50, 0.9));
  }
  
  table {
    background: rgba(40, 40, 40, 0.8);
  }
  
  td {
    color: #e0e0e0;
  }
}
</style>
