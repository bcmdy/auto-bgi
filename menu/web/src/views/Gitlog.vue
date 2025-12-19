<template>
  <div class="js-names-page">
    <header class="page-header">
      <div class="header-carousel" v-if="headerCarouselImages.length > 0">
        <div class="carousel-container">
          <div 
            v-for="(image, index) in headerCarouselImages" 
            :key="index" 
            class="carousel-slide" 
            :class="{ active: headerCurrentImageIndex === index }"
          >
            <img :src="image" :alt="`header-bg-${index}`" />
          </div>
        </div>
        <div class="carousel-overlay"></div>
      </div>

      <div class="header-content">
        <button class="btn home-btn" @click="goHome">
          <span class="icon">🏠</span> 返回首页
        </button>
        <div class="title-section">
          <h1 class="main-title">✨ 仓库提交记录 ✨</h1>
          <p class="sub-title">追踪最新的脚本与策略更新</p>
        </div>
      </div>
    </header>

    <div class="container">
      <section class="filter-panel">
        <div class="filter-item">
          <span class="filter-label">📂 分组筛选:</span>
          <select v-model="selectedGroup" class="custom-select">
            <option value="">全部显示</option>
            <option v-for="g in groupOptions" :key="g.value" :value="g.value">
              {{ g.label }}
            </option>
          </select>
        </div>

        <div class="filter-item">
          <span class="filter-label">👤 作者筛选:</span>
          <select v-model="selectedAuthor" class="custom-select">
            <option value="">全部作者</option>
            <option v-for="author in authorOptions" :key="author" :value="author">
              {{ author }}
            </option>
          </select>
        </div>
      </section>

      <div v-if="gitLogLoading" class="loading-state">
        <div class="spinner"></div>
        <p>正在从异世界获取数据...</p>
      </div>

      <div v-else>
        <div v-if="sortedGitLogs.length === 0" class="empty-state">
          <span class="empty-icon">🍃</span>
          <p>暂无相关提交记录</p>
        </div>

        <div v-else class="data-wrapper">
          
          <div class="table-view hidden-mobile">
            <div class="table-wrapper">
              <table>
                <thead>
                  <tr>
                    <th>📦 类型</th>
                    <th>📄 文件路径</th>
                    <th>👤 作者</th>
                    <th @click="sortTable('LastUpdated')" class="sortable">
                      🕒 更新时间
                      <span :class="['sort-icon', getSortIcon('LastUpdated')]"></span>
                    </th>
                    <th>🏷️ 标签</th>
                    <th>🔢 版本</th>
                    <th>📝 描述</th>
                    <th>🔍 操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(item, index) in sortedGitLogs" :key="index">
                    <td><span class="tag-badge">{{ item.TypeName }}</span></td>
                    <td class="col-path" :title="item.FilePath">{{ item.FilePath }}</td>
                    <td>{{ item.Authors }}</td>
                    <td class="col-date">{{ item.LastUpdated }}</td>
                    <td>{{ item.Tags }}</td>
                    <td><span class="version-badge">{{ item.Version }}</span></td>
                    <td class="col-desc" :title="item.Description">
                      {{ item.Description ? item.Description.slice(0, 30) + (item.Description.length > 30 ? '...' : '') : '-' }}
                    </td>
                    <td>
                      <button 
                        class="btn-action"
                        @click="openDetailFromFile(item.FilePath)"
                        :disabled="isLoadingDetail[getRepoKey(item.FilePath)]"
                      >
                        {{ isLoadingDetail[getRepoKey(item.FilePath)] ? '⏳' : '📖' }} 详情
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="card-view hidden-desktop">
            <div class="mobile-card" v-for="(item, index) in sortedGitLogs" :key="index">
              <div class="card-header">
                <span class="tag-badge">{{ item.TypeName }}</span>
                <span class="version-badge">{{ item.Version }}</span>
              </div>
              <div class="card-body">
                <div class="card-row">
                  <span class="row-icon">📄</span>
                  <span class="row-text path">{{ item.FilePath }}</span>
                </div>
                <div class="card-row">
                  <span class="row-icon">👤</span>
                  <span class="row-text">{{ item.Authors }}</span>
                </div>
                <div class="card-row">
                  <span class="row-icon">🕒</span>
                  <span class="row-text">{{ item.LastUpdated }}</span>
                </div>
                <div class="card-desc" v-if="item.Description">
                  {{ item.Description }}
                </div>
              </div>
              <div class="card-footer">
                <button 
                  class="btn-card-action"
                  @click="openDetailFromFile(item.FilePath)"
                  :disabled="isLoadingDetail[getRepoKey(item.FilePath)]"
                >
                  {{ isLoadingDetail[getRepoKey(item.FilePath)] ? '加载中...' : '查看详情 README' }}
                </button>
              </div>
            </div>
          </div>

        </div>
      </div>
    </div>

    <div class="modal-overlay" v-if="showDetailModal" @click.self="closeDetailModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>📖 {{ currentJsName }} - 详情</h3>
          <button class="close-btn" @click="closeDetailModal">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="!jsDetailContent" class="loading-state small">
            <div class="spinner"></div>
            <p>正在读取卷轴...</p>
          </div>
          <div v-else class="markdown-body" v-html="jsDetailHtml"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, reactive } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useRouter } from 'vue-router'
import { apiMethods } from '../utils/api'
import '../assets/markdown.css' 
import api from '@/utils/api'

export default {
  name: 'JsNames',
  setup() {
    const router = useRouter()
    const pluginData = ref([])
    const gitLogs = ref([])
    const gitLogLoading = ref(true)
    const currentSort = ref({ key: 'ChineseName', asc: true })
    const isUpdating = reactive({})

    // 详情模态框相关
    const showDetailModal = ref(false)
    const currentJsName = ref('')
    const jsDetailContent = ref('')
    const jsDetailHtml = ref('')
    const isLoadingDetail = reactive({})

    const renderMarkdownToHtml = (markdownText) => {
      try {
        const rawHtml = marked.parse(markdownText || '')
        return DOMPurify.sanitize(rawHtml)
      } catch (e) {
        return ''
      }
    }

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
        headerCarouselImages.value = data.images || []
        
        // 启动header轮播
        if (headerCarouselImages.value.length > 0) {
          startHeaderCarousel()
        }
      } catch (error) {
        console.error('获取Header轮播图失败:', error)
        headerCarouselImages.value = ['/img/bd.jpg', '/img/ff.png', '/img/ng.jpg', '/img/sh.jpg']
        startHeaderCarousel()
      }
    }

    // 启动header轮播
    const startHeaderCarousel = () => {
      if (headerCarouselImages.value.length > 1) {
        if(headerCarouselInterval) clearInterval(headerCarouselInterval)
        headerCarouselInterval = setInterval(() => {
          headerCurrentImageIndex.value = (headerCurrentImageIndex.value + 1) % headerCarouselImages.value.length
        }, 7000)
      }
    }

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

    // 扁平化 gitLogs
    const flatGitLogs = computed(() => {
      if (!gitLogs.value || !Array.isArray(gitLogs.value)) return []
      const arr = []
      for (const group of gitLogs.value) {
        if (group && group.Repo && Array.isArray(group.Repo)) {
          for (const file of group.Repo) {
            arr.push({
              TypeName: group.TypeName,
              ...file
            })
          }
        }
      }
      return arr
    })

    // 分组筛选相关
    const groupOptions = [
      { value: 'pathing', label: '地图追踪' },
      { value: 'js', label: '脚本' },
      { value: 'combat', label: '战斗策略' }
    ]
    const selectedGroup = ref('')

    // 作者筛选相关
    const selectedAuthor = ref('')

    // 计算所有可用的作者列表
    const authorOptions = computed(() => {
      if (!gitLogs.value || !Array.isArray(gitLogs.value)) return []
      const authorsSet = new Set()
      for (const group of gitLogs.value) {
        if (group && group.Repo && Array.isArray(group.Repo)) {
          for (const file of group.Repo) {
            if (file.Authors) {
              const authors = file.Authors.split(',').map(author => author.trim())
              authors.forEach(author => {
                if (author) authorsSet.add(author)
              })
            }
          }
        }
      }
      return Array.from(authorsSet).sort()
    })

    // 排序和筛选逻辑
    const sortedGitLogs = computed(() => {
      let logs = flatGitLogs.value
      
      // 分组筛选
      if (selectedGroup.value) {
        logs = logs.filter(item => {
          if (selectedGroup.value == 'pathing') return item.TypeName?.toLowerCase().includes('pathing')
          if (selectedGroup.value == 'js') return item.TypeName?.toLowerCase().includes('js')
          if (selectedGroup.value === 'combat') return item.TypeName?.toLowerCase().includes('combat')
          return false
        })
      }
      
      // 作者筛选
      if (selectedAuthor.value) {
        logs = logs.filter(item => {
          if (!item.Authors) return false
          const authors = item.Authors.split(',').map(author => author.trim())
          return authors.includes(selectedAuthor.value)
        })
      }
      
      // 按最后更新时间排序（降序）
      return [...logs].sort((a, b) => {
        const timeA = a.LastUpdated ? new Date(a.LastUpdated).getTime() : 0
        const timeB = b.LastUpdated ? new Date(b.LastUpdated).getTime() : 0
        return timeB - timeA
      })
    });

    const loadGitLog = async () => {
      try {
        gitLogLoading.value = true
        const response = await apiMethods.getLog()
        console.log('gitLog接口返回:', response)
        gitLogs.value = response.gitLog || []
      } catch (error) {
        console.error('加载提交记录失败：', error)
        gitLogs.value = []
      } finally {
        gitLogLoading.value = false
      }
    }

    const isRepoTriplePath = (filePath) => {
      return true
    }

    const getRepoSegments = (filePath) => {
      const match = filePath.match(/^repo\/(\/+)\/(\/+)\//)
      if (!match) return { group: '', name: '' }
      return { group: match[1], name: match[2] }
    }

    const getRepoKey = (filePath) => {
      const { group, name } = getRepoSegments(filePath)
      return group && name ? `${group}/${name}` : filePath
    }

    const openDetailFromFile = async (filePath) => {
      if (!isRepoTriplePath(filePath)) return
      const { group, name } = getRepoSegments(filePath)
      const key = `${group}/${name}`

      currentJsName.value = name
      showDetailModal.value = true
      isLoadingDetail[key] = true
      jsDetailContent.value = ''
      jsDetailHtml.value = ''

      try {
        const result = await api.get(`/api/md?filePath=${filePath}`)
        if (result.status === 'success') {
          jsDetailContent.value = result.data || ''
          jsDetailHtml.value = renderMarkdownToHtml(jsDetailContent.value)
        } else {
          jsDetailContent.value = '获取README内容失败'
        }
      } catch (error) {
        console.error('获取README失败：', error)
        jsDetailContent.value = '获取README内容失败：' + error.message
        jsDetailHtml.value = ''
      } finally {
        isLoadingDetail[key] = false
      }
    }

    const closeDetailModal = () => {
      showDetailModal.value = false
      currentJsName.value = ''
      jsDetailContent.value = ''
      jsDetailHtml.value = ''
    }

    onMounted(() => {
      loadGitLog()
      getHeaderImages()
    })

    return {
      pluginData,
      gitLogs,
      gitLogLoading,
      goHome,
      sortTable,
      getSortIcon,
      headerCarouselImages,
      headerCurrentImageIndex,
      showDetailModal,
      currentJsName,
      jsDetailContent,
      jsDetailHtml,
      isLoadingDetail,
      isRepoTriplePath,
      getRepoSegments,
      getRepoKey,
      openDetailFromFile,
      closeDetailModal,
      flatGitLogs,
      sortedGitLogs,
      groupOptions,
      selectedGroup,
      selectedAuthor,
      authorOptions
    }
  }
}
</script>

<style scoped>
/* ================= 全局变量 ================= */
:root {
  --primary-pink: #ff9ecd;
  --dark-pink: #ff6eb4;
  --light-pink: #fff0f6;
  --accent-blue: #87ceeb;
  --text-main: #5a5a5a;
  --glass-bg: rgba(255, 255, 255, 0.85);
  --glass-border: 1px solid rgba(255, 255, 255, 0.6);
  --shadow-soft: 0 8px 32px rgba(255, 158, 205, 0.2);
  --radius-lg: 24px;
  --radius-md: 16px;
  --radius-sm: 8px;
}

.js-names-page {
  min-height: 100vh;
  /* 可爱的背景纹理 */
  background-color: #fffafc;
  background-image: 
    radial-gradient(#ffdef0 2px, transparent 2px), 
    radial-gradient(#ffdef0 2px, transparent 2px);
  background-size: 40px 40px;
  background-position: 0 0, 20px 20px;
  color: var(--text-main);
  font-family: 'Varela Round', 'Nunito', 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  padding-bottom: 60px;
}

/* ================= Header 样式 ================= */
.page-header {
  position: relative;
  height: 280px;
  border-bottom-left-radius: var(--radius-lg);
  border-bottom-right-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-soft);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 30px;
  background: linear-gradient(135deg, #ffcce6, #d4f0ff);
}

.header-carousel {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
}

.carousel-slide {
  position: absolute;
  width: 100%;
  height: 100%;
  opacity: 0;
  transition: opacity 1s ease-in-out;
}

.carousel-slide.active {
  opacity: 1;
}

.carousel-slide img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.carousel-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(to bottom, rgba(255,255,255,0.4), rgba(255,240,246,0.9));
  backdrop-filter: blur(2px);
}

.header-content {
  position: relative;
  z-index: 2;
  text-align: center;
  width: 100%;
  max-width: 1200px;
}

.title-section {
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0% { transform: translateY(0px); }
  50% { transform: translateY(-10px); }
  100% { transform: translateY(0px); }
}

.main-title {
  font-size: 2.5rem;
  color: var(--dark-pink);
  margin: 0;
  text-shadow: 2px 2px 0px #fff, 4px 4px 0px rgba(255, 158, 205, 0.3);
  font-weight: 800;
  letter-spacing: 1px;
}

.sub-title {
  font-size: 1.1rem;
  color: #888;
  margin-top: 10px;
  background: rgba(255,255,255,0.6);
  display: inline-block;
  padding: 4px 16px;
  border-radius: 20px;
}

.home-btn {
  position: absolute;
  top: 20px;
  left: 20px;
  background: var(--glass-bg);
  border: var(--glass-border);
  color: var(--dark-pink);
  padding: 10px 20px;
  border-radius: 30px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
  display: flex;
  align-items: center;
  gap: 8px;
}

.home-btn:hover {
  transform: translateY(-2px);
  background: #fff;
  box-shadow: 0 6px 16px rgba(255, 110, 180, 0.2);
}

/* ================= 容器与筛选 ================= */
.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.filter-panel {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  background: var(--glass-bg);
  padding: 20px;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-soft);
  margin-bottom: 24px;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255,255,255,0.8);
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.filter-label {
  font-weight: bold;
  color: var(--dark-pink);
}

.custom-select {
  padding: 8px 16px;
  border-radius: 20px;
  border: 2px solid #ffdef0;
  background: #fff;
  color: var(--text-main);
  outline: none;
  cursor: pointer;
  transition: border-color 0.3s;
}

.custom-select:focus {
  border-color: var(--dark-pink);
}

/* ================= 表格样式 (PC) ================= */
.table-view {
  background: var(--glass-bg);
  border-radius: var(--radius-md);
  padding: 10px;
  box-shadow: var(--shadow-soft);
  backdrop-filter: blur(10px);
  overflow: hidden;
  width: 100%; /* 确保宽度撑开 */
}

.table-wrapper {
  overflow-x: auto;
  width: 100%;
}

table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  min-width: 800px; /* 防止表格过分压缩 */
}

th {
  background: rgba(255, 158, 205, 0.1);
  color: var(--dark-pink);
  padding: 16px;
  text-align: left;
  font-weight: bold;
  white-space: nowrap;
}

th:first-child { border-top-left-radius: var(--radius-sm); }
th:last-child { border-top-right-radius: var(--radius-sm); }

td {
  padding: 14px 16px;
  border-bottom: 1px solid #ffeef6;
  font-size: 0.95rem;
  color: #666;
  vertical-align: middle;
}

tr:last-child td { border-bottom: none; }

tr:hover td {
  background-color: rgba(255, 239, 247, 0.5);
}

.tag-badge {
  background: #ffe0f0;
  color: var(--dark-pink);
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: bold;
  display: inline-block;
}

.version-badge {
  background: #e0f7fa;
  color: #00bcd4;
  padding: 2px 8px;
  border-radius: 8px;
  font-family: monospace;
}

.col-path {
  font-family: 'Consolas', monospace;
  color: #888;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-desc {
  max-width: 250px;
  color: #777;
}

.sortable { cursor: pointer; user-select: none; }
.sort-icon::after { content: ' ↕'; opacity: 0.3; }
.sort-asc::after { content: ' ↑'; opacity: 1; }
.sort-desc::after { content: ' ↓'; opacity: 1; }

.btn-action {
  background: linear-gradient(135deg, #ff9ecd, #ff6eb4);
  color: white;
  border: none;
  padding: 6px 14px;
  border-radius: 20px;
  cursor: pointer;
  font-size: 0.85rem;
  transition: transform 0.2s;
  box-shadow: 0 2px 8px rgba(255, 110, 180, 0.3);
}

.btn-action:hover {
  transform: scale(1.05);
}
.btn-action:disabled {
  background: #ddd;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

/* ================= 卡片样式 (Mobile) ================= */
.mobile-card {
  background: #fff;
  border-radius: var(--radius-md);
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 4px 16px rgba(255, 158, 205, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.5);
  transition: transform 0.2s;
}

.mobile-card:active {
  transform: scale(0.98);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px dashed #ffeef6;
}

.card-body {
  margin-bottom: 12px;
}

.card-row {
  display: flex;
  align-items: center;
  margin-bottom: 6px;
  font-size: 0.9rem;
}

.row-icon {
  width: 24px;
  text-align: center;
  margin-right: 8px;
}

.path {
  font-family: monospace;
  color: #888;
  word-break: break-all;
}

.card-desc {
  background: #fafafa;
  padding: 8px;
  border-radius: 8px;
  font-size: 0.85rem;
  color: #666;
  margin-top: 8px;
  line-height: 1.4;
}

.card-footer {
  text-align: center;
}

.btn-card-action {
  width: 100%;
  padding: 10px;
  background: #fff0f6;
  color: var(--dark-pink);
  border: 1px solid #ffcce6;
  border-radius: 12px;
  font-weight: bold;
  cursor: pointer;
}

/* ================= 详情模态框 ================= */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(4px);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-content {
  background: rgba(255, 255, 255, 0.95);
  width: 90%;
  max-width: 800px;
  height: 80vh;
  border-radius: var(--radius-lg);
  box-shadow: 0 20px 60px rgba(0,0,0,0.2);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 2px solid #fff;
  animation: popIn 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

@keyframes popIn {
  from { transform: scale(0.8); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}

.modal-header {
  padding: 16px 24px;
  background: linear-gradient(90deg, #fff0f6, #fff);
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #eee;
}

.modal-header h3 {
  margin: 0;
  color: var(--dark-pink);
  font-size: 1.2rem;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #999;
  cursor: pointer;
}

.modal-body {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
  font-size: 0.95rem;
}

/* ================= 状态组件 ================= */
.loading-state, .empty-state {
  text-align: center;
  padding: 60px 0;
  color: #aaa;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #ffe0f0;
  border-top: 4px solid var(--dark-pink);
  border-radius: 50%;
  margin: 0 auto 16px;
  animation: spin 1s linear infinite;
}

@keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }

.empty-icon { font-size: 3rem; display: block; margin-bottom: 10px; }

/* ================= 响应式切换 (确保PC端优先显示) ================= */
.hidden-mobile { display: block; }
.hidden-desktop { display: none; }

@media (max-width: 768px) {
  .hidden-mobile { display: none; }
  .hidden-desktop { display: block; }
  
  .page-header { height: 200px; }
  .main-title { font-size: 1.8rem; }
  .filter-panel { flex-direction: column; gap: 10px; }
  .custom-select { width: 100%; }
}
</style>