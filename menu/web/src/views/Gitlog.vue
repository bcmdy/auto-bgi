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

      <button class="btn home-btn" @click="goHome">
        <span class="icon">🏠</span> 
        <span class="text">返回首页</span>
      </button>

      <div class="header-content">
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
                    <td><span class="author-tag">{{ item.Authors }}</span></td>
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
            <div class="mobile-card" v-for="(item, index) in sortedGitLogs" :key="index" :style="{ '--delay': index * 0.1 + 's' }">
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
        <div class="modal-body custom-scrollbar">
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
    const currentSort = ref({ key: 'LastUpdated', asc: false }) // 默认按时间倒序
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

    const getHeaderImages = async () => {
      try {
        const response = await fetch('/api/images')
        if (!response.ok) throw new Error('Failed')
        const data = await response.json()
        headerCarouselImages.value = data.images || []
        if (headerCarouselImages.value.length > 0) startHeaderCarousel()
      } catch (error) {
        // 默认图片回退
        headerCarouselImages.value = ['/img/bd.jpg', '/img/ff.png']
        startHeaderCarousel()
      }
    }

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
      if (currentSort.value.key !== key) return 'sort-default'
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

    const groupOptions = [
      { value: 'pathing', label: '地图追踪' },
      { value: 'js', label: '脚本' },
      { value: 'combat', label: '战斗策略' }
    ]
    const selectedGroup = ref('')
    const selectedAuthor = ref('')

    const authorOptions = computed(() => {
      if (!gitLogs.value || !Array.isArray(gitLogs.value)) return []
      const authorsSet = new Set()
      for (const group of gitLogs.value) {
        if (group && group.Repo && Array.isArray(group.Repo)) {
          for (const file of group.Repo) {
            if (file.Authors) {
              const authors = file.Authors.split(',').map(author => author.trim())
              authors.forEach(author => { if (author) authorsSet.add(author) })
            }
          }
        }
      }
      return Array.from(authorsSet).sort()
    })

    const sortedGitLogs = computed(() => {
      let logs = flatGitLogs.value
      
      if (selectedGroup.value) {
        logs = logs.filter(item => {
          if (selectedGroup.value == 'pathing') return item.TypeName?.toLowerCase().includes('pathing')
          if (selectedGroup.value == 'js') return item.TypeName?.toLowerCase().includes('js')
          if (selectedGroup.value === 'combat') return item.TypeName?.toLowerCase().includes('combat')
          return false
        })
      }
      
      if (selectedAuthor.value) {
        logs = logs.filter(item => {
          if (!item.Authors) return false
          const authors = item.Authors.split(',').map(author => author.trim())
          return authors.includes(selectedAuthor.value)
        })
      }
      
      return [...logs].sort((a, b) => {
        let valA = a[currentSort.value.key]
        let valB = b[currentSort.value.key]

        // 特殊处理日期
        if(currentSort.value.key === 'LastUpdated') {
           valA = new Date(valA).getTime()
           valB = new Date(valB).getTime()
        }

        if (valA < valB) return currentSort.value.asc ? -1 : 1
        if (valA > valB) return currentSort.value.asc ? 1 : -1
        return 0
      })
    });

    const loadGitLog = async () => {
      try {
        gitLogLoading.value = true
        const response = await apiMethods.getLog()
        gitLogs.value = response.gitLog || []
      } catch (error) {
        console.error('加载提交记录失败：', error)
        gitLogs.value = []
      } finally {
        gitLogLoading.value = false
      }
    }

    const isRepoTriplePath = (filePath) => true // 简化判断

    const getRepoSegments = (filePath) => {
      const match = filePath.match(/^repo\/(\/+)\/(\/+)\//)
      if (!match) return { group: '', name: '' }
      return { group: match[1], name: match[2] }
    }

    const getRepoKey = (filePath) => filePath

    const openDetailFromFile = async (filePath) => {
      // 提取文件名作为标题
      const parts = filePath.split('/')
      const name = parts[parts.length - 1] || filePath
      const key = filePath

      currentJsName.value = name
      showDetailModal.value = true
      isLoadingDetail[key] = true
      jsDetailContent.value = ''
      
      try {
        // 注意：这里需要根据实际后端API路径调整
        const result = await api.get(`/api/md?filePath=${encodeURIComponent(filePath)}`)
        if (result && (result.data || result.content)) {
          jsDetailContent.value = result.data || result.content
          jsDetailHtml.value = renderMarkdownToHtml(jsDetailContent.value)
        } else {
          jsDetailContent.value = '# 暂无详情\n无法读取该文件的说明文档。'
          jsDetailHtml.value = renderMarkdownToHtml(jsDetailContent.value)
        }
      } catch (error) {
        jsDetailContent.value = '获取README失败：' + error.message
        jsDetailHtml.value = renderMarkdownToHtml(jsDetailContent.value)
      } finally {
        isLoadingDetail[key] = false
      }
    }

    const closeDetailModal = () => {
      showDetailModal.value = false
    }

    onMounted(() => {
      loadGitLog()
      getHeaderImages()
    })

    return {
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
      getRepoKey,
      openDetailFromFile,
      closeDetailModal,
      sortedGitLogs,
      groupOptions,
      selectedGroup,
      selectedAuthor,
      authorOptions
    }
  }
}
</script>

<style>
/* ================= 全局 & 滚动条 ================= */
:root {
  --primary-pink: #ff9ecd;
  --dark-pink: #ff6eb4;
  --light-pink: #fff0f6;
  --text-main: #5a5a5a;
  --glass-bg: rgba(255, 255, 255, 0.75);
  --shadow-soft: 0 8px 32px rgba(255, 158, 205, 0.15);
  --radius-lg: 24px;
}

/* 全局滚动条美化 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}
::-webkit-scrollbar-track {
  background: #fff5f9;
  border-radius: 4px;
}
::-webkit-scrollbar-thumb {
  background: var(--primary-pink);
  border-radius: 4px;
}
::-webkit-scrollbar-thumb:hover {
  background: var(--dark-pink);
}

.js-names-page {
  min-height: 100vh;
  background-color: #fffafc;
  /* 更细腻的背景网格 */
  background-image: 
    linear-gradient(rgba(255, 158, 205, 0.1) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 158, 205, 0.1) 1px, transparent 1px);
  background-size: 30px 30px;
  color: var(--text-main);
  font-family: 'Varela Round', sans-serif;
  padding-bottom: 60px;
}

/* ================= Header 区域 ================= */
.page-header {
  position: relative;
  height: 260px;
  border-bottom-left-radius: 40px;
  border-bottom-right-radius: 40px;
  overflow: hidden;
  box-shadow: 0 10px 40px rgba(255, 110, 180, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 30px;
  /* 确保按钮可以定位 */
  z-index: 1; 
}

.header-carousel {
  position: absolute;
  inset: 0;
  z-index: -1;
}

.carousel-slide {
  position: absolute;
  width: 100%;
  height: 100%;
  opacity: 0;
  transition: opacity 1s ease;
  transform: scale(1.05); /* 轻微放大效果 */
}
.carousel-slide.active { opacity: 1; transform: scale(1); transition: opacity 1s, transform 6s; }
.carousel-slide img { width: 100%; height: 100%; object-fit: cover; }

.carousel-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to bottom, rgba(255,255,255,0.2), rgba(255,240,246,0.95));
  backdrop-filter: blur(1px);
}

/* 修复后的按钮样式：绝对定位在Header左上角 */
.home-btn {
  position: absolute;
  z-index: 1000; /* 确保在最上层 */
  top: 20px;
  left: 20px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(10px);
  border: 2px solid #fff;
  color: var(--dark-pink);
  padding: 8px 18px;
  border-radius: 30px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  align-items: center;
  gap: 6px;
  box-shadow: 0 4px 15px rgba(0,0,0,0.05);
}

.home-btn:hover {
  background: #fff;
  transform: translateY(-2px) scale(1.05);
  box-shadow: 0 0 15px rgba(255, 110, 180, 0.4);
  color: #ff4081;
}

.header-content {
  position: relative;
  z-index: 2;
  text-align: center;
  animation: float 4s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8px); }
}

.main-title {
  font-size: 2.8rem;
  color: var(--dark-pink);
  margin: 0;
  text-shadow: 3px 3px 0px #fff;
  font-weight: 900;
  letter-spacing: 2px;
}

.sub-title {
  font-size: 1rem;
  color: #777;
  margin-top: 8px;
  background: rgba(255,255,255,0.7);
  padding: 4px 16px;
  border-radius: 20px;
  border: 1px solid #fff;
}

/* ================= 筛选栏 ================= */
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
  padding: 20px 24px;
  border-radius: 20px;
  box-shadow: var(--shadow-soft);
  margin-bottom: 24px;
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255,255,255,0.6);
}

.filter-label { color: var(--dark-pink); font-weight: 700; }

.custom-select {
  padding: 8px 16px;
  border-radius: 12px;
  border: 1px solid #ffcce6;
  background: #fff;
  color: #666;
  outline: none;
  cursor: pointer;
  transition: all 0.3s;
}
.custom-select:hover, .custom-select:focus {
  border-color: var(--dark-pink);
  box-shadow: 0 0 0 3px rgba(255, 110, 180, 0.1);
}

/* ================= 表格 (PC) ================= */
.table-view {
  background: rgba(255, 255, 255, 0.85);
  border-radius: 24px;
  padding: 12px;
  box-shadow: var(--shadow-soft);
  backdrop-filter: blur(10px);
  border: 1px solid #fff;
}

.table-wrapper { overflow-x: auto; }
table { width: 100%; border-collapse: separate; border-spacing: 0; min-width: 900px; }

th {
  background: rgba(255, 225, 239, 0.5);
  color: var(--dark-pink);
  padding: 18px;
  text-align: left;
  font-weight: 800;
  border-bottom: 2px solid #fff;
}

td {
  padding: 16px;
  border-bottom: 1px solid #fff0f6;
  color: #666;
  font-size: 0.95rem;
  transition: background 0.2s;
}

tr:hover td { background: rgba(255, 240, 248, 0.6); }

/* 徽章样式 */
.tag-badge {
  background: linear-gradient(135deg, #ffcce6, #ffdef0);
  color: #d63384;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: bold;
}

.author-tag {
  color: #444;
  font-weight: 500;
}

.version-badge {
  background: #e3f2fd;
  color: #0288d1;
  padding: 3px 8px;
  border-radius: 6px;
  font-family: monospace;
  font-size: 0.85rem;
}

.btn-action {
  background: linear-gradient(135deg, #ff9ecd, #ff6eb4);
  color: white;
  border: none;
  padding: 8px 18px;
  border-radius: 20px;
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: bold;
  box-shadow: 0 4px 10px rgba(255, 110, 180, 0.3);
  transition: all 0.2s;
}
.btn-action:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 15px rgba(255, 110, 180, 0.5);
}
.btn-action:disabled { background: #e0e0e0; cursor: not-allowed; box-shadow: none; }

/* ================= 卡片 (Mobile) ================= */
.mobile-card {
  background: #fff;
  border-radius: 20px;
  padding: 20px;
  margin-bottom: 16px;
  box-shadow: 0 8px 20px rgba(0,0,0,0.03);
  border: 1px solid rgba(255, 255, 255, 0.8);
  
  /* 动画 */
  animation: slideUp 0.5s ease backwards;
  animation-delay: var(--delay);
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.card-header {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: 14px; padding-bottom: 10px; border-bottom: 1px dashed #eee;
}

.card-row { margin-bottom: 8px; font-size: 0.9rem; display: flex; align-items: flex-start;}
.row-icon { margin-right: 8px; min-width: 20px; }
.path { word-break: break-all; font-family: monospace; color: #ff6eb4; }

.btn-card-action {
  width: 100%; padding: 12px;
  background: #fff0f6; color: var(--dark-pink);
  border: 1px solid #ffcce6; border-radius: 14px;
  font-weight: bold; cursor: pointer; margin-top: 10px;
}

/* ================= 详情模态框 ================= */
.modal-overlay {
  position: fixed; inset: 0;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(5px);
  z-index: 1000;
  display: flex; align-items: center; justify-content: center;
}

.modal-content {
  background: rgba(255, 255, 255, 0.96);
  width: 90%; max-width: 800px; height: 80vh;
  border-radius: 24px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  display: flex; flex-direction: column;
  overflow: hidden;
  animation: popIn 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes popIn {
  from { transform: scale(0.9); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}

.modal-header {
  padding: 20px; background: #fff;
  border-bottom: 1px solid #f0f0f0;
  display: flex; justify-content: space-between; align-items: center;
}

.close-btn {
  background: #f5f5f5; border: none; width: 32px; height: 32px; border-radius: 50%;
  cursor: pointer; color: #666; transition: background 0.2s;
}
.close-btn:hover { background: #ffcce6; color: var(--dark-pink); }

.modal-body { padding: 30px; overflow-y: auto; flex: 1; }

/* 状态组件 */
.loading-state, .empty-state { text-align: center; padding: 60px 0; color: #bbb; }
.spinner {
  width: 40px; height: 40px; border: 4px solid #fff0f6;
  border-top-color: var(--dark-pink); border-radius: 50%;
  margin: 0 auto 16px; animation: spin 1s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* 响应式辅助 */
.hidden-mobile { display: block; }
.hidden-desktop { display: none; }

@media (max-width: 768px) {
  .hidden-mobile { display: none; }
  .hidden-desktop { display: block; }
  .page-header { height: 200px; }
  .main-title { font-size: 2rem; }
  .home-btn { top: 15px; left: 15px; padding: 6px 12px; }
  .home-btn .text { display: none; } /* 手机端只显示图标 */
  .filter-panel { flex-direction: column; }
  .custom-select { width: 100%; }
}
</style>