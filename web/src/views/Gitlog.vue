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

        <h1 class="header-title">📜 仓库提交记录 📜</h1>
        <p class="header-subtitle">近三天的提交 ✨</p>
                   <button class="btn home-btn" @click="goHome">返回首页</button>
      </div>
    </header>

    <div class="container">
           
      <section class="panel">
        <h2>仓库提交记录</h2>
        <!-- 分组筛选下拉框 -->
        <div style="margin-bottom:16px;">
          <label style="font-weight:bold;margin-right:8px;">分组筛选：</label>
          <select v-model="selectedGroup" class="group-select">
            <option value="">全部</option>
            <option v-for="g in groupOptions" :key="g.value" :value="g.value">{{ g.label }}</option>
          </select>
        </div>
        <div id="gitLogContainer" class="table-container git-log-container">
          <table id="gitLogTable" class="desktop-table">
            <thead>
              <tr>
                <th>分组</th>
                <th>文件路径</th>
                <th>作者</th>
                <th>最后更新时间</th>
                <th>标签</th>
                <th>版本</th>
                <th>描述</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody id="gitLogBody">
              <tr v-if="sortedGitLogs.length === 0">
                <td colspan="8" style="text-align:center;">
                  {{ gitLogLoading ? '加载中...' : '暂无提交记录。' }}
                </td>
              </tr>
              <tr v-else v-for="(item, idx) in sortedGitLogs" :key="item.TypeName + '-' + idx"
                  :class="{ highlight: item.Tags && item.Tags.includes('更新') }">
                <td>{{ item.TypeName }}</td>
                <td>{{ item.FilePath }}</td>
                <td>{{ item.Authors }}</td>
                <td>{{ item.LastUpdated }}</td>
                <td>{{ item.Tags }}</td>
                <td>{{ item.Version }}</td>
                <td>
                  <span v-if="item.Description && item.Description.length > 20">
                    <span class="desc-multiline">{{ item.Description.slice(0, 20) }}</span>
                    <br>
                    <span class="desc-multiline">{{ item.Description.slice(20) }}</span>
                  </span>
                  <span v-else class="desc-multiline">{{ item.Description || '' }}</span>
                </td>
                <td>
                  <button
                    v-if="isRepoTriplePath(item.FilePath)"
                    class="btn update-btn"
                    style="min-width:60px"
                    :disabled="isLoadingDetail[getRepoKey(item.FilePath)]"
                    @click="openDetailFromFile(item.FilePath)"
                  >
                    {{ isLoadingDetail[getRepoKey(item.FilePath)] ? '加载中...' : '查看详情' }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
          <!-- 移动端卡片列表，样式与脚本信息一致 -->
          <div class="mobile-list">
            <div v-if="flatGitLogs.length === 0" class="empty-mobile">
              <div class="empty-content">
                <span class="empty-icon">📭</span>
                <span class="empty-text">{{ gitLogLoading ? '加载中...' : '暂无提交记录' }}</span>
                <span class="empty-sparkle">✨</span>
              </div>
            </div>
            <div v-else v-for="(item, idx) in flatGitLogs" :key="item.TypeName + '-' + idx"
                 :class="{ highlight: item.Tags && item.Tags.includes('更新'), 'mobile-card': true }">
              <div class="card-header">
                <div class="card-title">
                  <span class="title-icon">📦</span>
                  <span class="title-text">{{ item.TypeName }}</span>
                </div>
                <div class="card-versions">
                  <div class="version-item">
                    <span class="version-label">作者:</span>
                    <span class="version-value">{{ item.Authors }}</span>
                  </div>
                  <div class="version-item">
                    <span class="version-label">更新时间:</span>
                    <span class="version-value">{{ item.LastUpdated }}</span>
                  </div>
                </div>
                <div class="card-status">
                  <span class="status-icon">🏷️</span>
                  <span class="status-text">{{ item.Tags }}</span>
                </div>
              </div>
              <div class="card-message">
                <span class="message-icon">📄</span>
                <span class="message-text">{{ item.FilePath }}</span>
              </div>
              <div class="card-files">
                <span class="files-icon">🔢</span>
                <div class="files-content">
                  <div>版本: {{ item.Version }}</div>
                  <div>描述: {{ item.Description || '' }}</div>
                  <button
                    v-if="isRepoTriplePath(item.FilePath)"
                    class="btn update-btn mobile-update-btn"
                    style="margin-top:8px"
                    :disabled="isLoadingDetail[getRepoKey(item.FilePath)]"
                    @click="openDetailFromFile(item.FilePath)"
                  >
                    <span class="update-icon">{{ isLoadingDetail[getRepoKey(item.FilePath)] ? '⏳' : '🔍' }}</span>
                    <span class="update-text">{{ isLoadingDetail[getRepoKey(item.FilePath)] ? '加载中...' : '查看详情' }}</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- 详情模态框 -->
    <div v-if="showDetailModal" class="modal-overlay" @click="closeDetailModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>📖 {{ currentJsName }} - README详情</h3>
          <button class="modal-close-btn" @click="closeDetailModal">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="isLoadingDetail[currentJsName]" class="loading-content">
            <div class="loading-spinner"></div>
            <p>正在加载README内容...</p>
          </div>
          <div v-else-if="jsDetailHtml" class="detail-content markdown-body" v-html="jsDetailHtml"></div>
          <div v-else class="no-content">
            <p>暂无README内容</p>
          </div>
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

    // 配置 GitHub Flavored Markdown 渲染
    marked.setOptions({
      gfm: true,
      breaks: true
    })

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



    // 新增：将 gitLogs 扁平化为每个文件一行，带 TypeName，并支持分组筛选
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
      // 分组筛选
      if (selectedGroup.value) {
        return arr.filter(item => {
          if (selectedGroup.value === 'pathing') return item.TypeName?.toLowerCase().includes('pathing')
          if (selectedGroup.value === 'js') return item.TypeName?.toLowerCase().includes('js')
          if (selectedGroup.value === 'combat') return item.TypeName?.toLowerCase().includes('combat')
          return false
        })
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

    // 新增：按分组筛选并按最后更新时间排序（降序）
    const sortedGitLogs = computed(() => {
      let logs = flatGitLogs.value
      if (selectedGroup.value) {
        logs = logs.filter(item => {
          // 支持TypeName为中英文或英文
          if (selectedGroup.value === 'pathing') return item.TypeName?.toLowerCase().includes('pathing')
          if (selectedGroup.value === 'js') return item.TypeName?.toLowerCase().includes('js')
          if (selectedGroup.value === 'combat') return item.TypeName?.toLowerCase().includes('combat')
          return false
        })
      }
      return [...logs].sort((a, b) => {
        const timeA = a.LastUpdated ? new Date(a.LastUpdated).getTime() : 0
        const timeB = b.LastUpdated ? new Date(b.LastUpdated).getTime() : 0
        return timeB - timeA
      })
    });

    const loadGitLog = async () => {
      try {
        gitLogLoading.value = true
        const response = await fetch('/api/gitLog')
        const json = await response.json()
        console.log('gitLog接口返回:', json) // 调试输出
        gitLogs.value = json.gitLog || []
      } catch (error) {
        console.error('加载提交记录失败：', error)
        gitLogs.value = []
      } finally {
        gitLogLoading.value = false
      }
    }


    // 判断是否为 repo/**/**/** 结构，至少包含 repo/<group>/<name>/...
    const isRepoTriplePath = (filePath) => {
      // return /^repo\/[^^\/]+\/[^^\/]+\//.test(filePath)
      return true
    }

    // 提取 repo/<group>/<name> 的两个段
    const getRepoSegments = (filePath) => {
      const match = filePath.match(/^repo\/([^\/]+)\/([^\/]+)\//)
      if (!match) return { group: '', name: '' }
      return { group: match[1], name: match[2] }
    }

    // 作为加载状态键值
    const getRepoKey = (filePath) => {
      const { group, name } = getRepoSegments(filePath)
      return group && name ? `${group}/${name}` : filePath
    }



    // 从文件路径打开详情：提取 repo/<group>/<name>，并调用 /api/md?group=&name=
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
        const response = await fetch(`/api/md?filePath=${filePath}`)
        const result = await response.json()

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

    // 关闭详情模态框
    const closeDetailModal = () => {
      showDetailModal.value = false
      currentJsName.value = ''
      jsDetailContent.value = ''
      jsDetailHtml.value = ''
    }

    onMounted(() => {
      loadGitLog()
      getHeaderImages() // 在组件挂载时获取header轮播图

      
    })

    return {
      pluginData,
      gitLogs,
      gitLogLoading,
      goHome,
      sortTable,
      getSortIcon,
      headerCarouselImages, // 暴露header轮播图数据
      headerCurrentImageIndex, // 暴露header轮播图当前图片索引
      // 详情模态框相关
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
      selectedGroup
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

html {
  color-scheme: light;
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
  
  /* 禁用触摸设备上的hover效果，防止闪烁 */
  @media (hover: none) {
    .mobile-card:hover,
    .mobile-card:hover::before {
      /* 取消hover效果 */
      background-color: inherit !important;
      box-shadow: none !important;
      transform: none !important;
    }
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

/* 详情按钮样式 */
.detail-btn {
  background: linear-gradient(135deg, #fff, #fff6fb);
  color: var(--primary-color);
  border: 2px solid var(--primary-color);
  border-radius: 15px;
  padding: 4px 8px;
  font-size: 0.7rem;
  cursor: pointer;
  transition: all 0.3s ease;
  font-weight: bold;
  margin-left: 8px;
  min-width: 50px;
}

.detail-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, var(--primary-color), #ff8e8e);
  color: #fff;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(255, 110, 180, 0.3);
}

.detail-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  background: #f5f5f5;
  color: #999;
  border-color: #ddd;
}

.desktop-detail-btn {
  font-size: 0.6rem;
  padding: 2px 6px;
  margin-left: 5px;
}

/* 模态框样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.28); /* 降低遮罩层暗度 */
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(5px);
}

.modal-content {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(255, 246, 251, 0.95));
  border-radius: 20px;
  box-shadow: 0 20px 60px rgba(255, 110, 180, 0.3);
  max-width: 90%;
  max-height: 90%;
  width: 800px;
  overflow: hidden;
  border: 2px solid rgba(255, 110, 180, 0.2);
  backdrop-filter: blur(10px);
}

.modal-header {
  background: linear-gradient(135deg, #ff69b4, #ff1493);
  color: white;
  padding: 20px 25px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
}

.modal-header h3 {
  margin: 0;
  font-size: 1.3rem;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.modal-close-btn {
  background: rgba(255, 105, 180, 0.3);
  border: none;
  color: white;
  font-size: 1.2rem;
  cursor: pointer;
  border-radius: 50%;
  width: 35px;
  height: 35px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
}

.modal-close-btn:hover {
  background: rgba(255, 20, 147, 0.5);
  transform: scale(1.1);
}

.modal-body {
  padding: 25px;
  max-height: 60vh;
  overflow-y: auto;
}

.loading-content {
  text-align: center;
  padding: 40px 20px;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid rgba(255, 110, 180, 0.2);
  border-top: 4px solid var(--primary-color);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 20px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.detail-content {
  background: rgba(255, 255, 255, 0.8);
  border-radius: 10px;
  padding: 20px;
  border: 1px solid rgba(255, 110, 180, 0.25);
}

.detail-content pre {
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
  line-height: 1.5;
  color: #333;
  margin: 0;
  overflow-x: auto;
}

/* GitHub风格 Markdown 简要样式 */
.markdown-body {
  color: #1f2328;
  line-height: 1.6;
}
.markdown-body h1, .markdown-body h2, .markdown-body h3,
.markdown-body h4, .markdown-body h5, .markdown-body h6 {
  margin: 1em 0 0.6em;
  font-weight: 600;
}
.markdown-body h1 { font-size: 1.6em; }
.markdown-body h2 { font-size: 1.4em; }
.markdown-body h3 { font-size: 1.2em; }
.markdown-body p { margin: 0.6em 0; }
.markdown-body ul, .markdown-body ol { padding-left: 1.5em; }
.markdown-body code {
  background: rgba(27,31,35,0.05);
  padding: 0.2em 0.4em;
  border-radius: 4px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
}
.markdown-body pre code {
  display: block;
  padding: 1em;
  overflow-x: auto;
}
.markdown-body blockquote {
  margin: 0.8em 0;
  padding: 0.5em 1em;
  color: #6a737d;
  border-left: 0.25em solid #dfe2e5;
  background: rgba(0,0,0,0.02);
}
.markdown-body table {
  border-collapse: collapse;
  width: 100%;
}
.markdown-body table th,
.markdown-body table td {
  border: 1px solid #dfe2e5;
  padding: 6px 13px;
}
.markdown-body a { color: #0969da; }

/* 强化滚动条外观（更清晰） */
.modal-body::-webkit-scrollbar { width: 10px; }
.modal-body::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg, #ff69b4, #ff1493);
  border-radius: 8px;
}
.modal-body::-webkit-scrollbar-track {
  background: rgba(255, 105, 180, 0.2);
  border-radius: 8px;
}

.no-content {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-color);
  font-style: italic;
}

/* 仓库提交记录表格与脚本信息表格尺寸一致，宽度自适应容器 */
#gitLogTable,
#pluginTable {
  width: 100%;
  min-width: 0;
  table-layout: auto;
  font-size: 1rem;
}

/* 保证表格容器可横向滚动但宽度跟随容器 */
.table-container {
  overflow-x: auto;
}

/* 响应式缩放：表格字体和padding随屏幕宽度变化 */
@media (max-width: 1200px) {
  #gitLogTable,
  #pluginTable {
    font-size: 0.95rem;
  }
  #gitLogTable th, #gitLogTable td,
  #pluginTable th, #pluginTable td {
    padding: 10px 8px;
  }
}
@media (max-width: 900px) {
  #gitLogTable,
  #pluginTable {
    font-size: 0.9rem;
  }
  #gitLogTable th, #gitLogTable td,
  #pluginTable th, #pluginTable td {
    padding: 8px 6px;
  }
}
@media (max-width: 700px) {
  #gitLogTable,
  #pluginTable {
    font-size: 0.85rem;
  }
  #gitLogTable th, #gitLogTable td,
  #pluginTable th, #pluginTable td {
    padding: 6px 4px;
  }
}
@media (max-width: 600px) {
  #gitLogTable,
  #pluginTable {
    font-size: 0.8rem;
  }
  #gitLogTable th, #gitLogTable td,
  #pluginTable th, #pluginTable td {
    padding: 4px 2px;
  }
}

/* ============ 新增样式 ============ */

/* 仓库提交记录表格描述列缩略 */
#gitLogTable td {
  position: relative;
  overflow: hidden;
}

#gitLogTable td::after {
  content: '...';
  position: absolute;
  top: 50%;
  left: 100%;
  transform: translateY(-50%);
  font-size: 0.8rem;
  color: rgba(255, 110, 180, 0.7);
  display: none;
}

#gitLogTable td:hover::after {
  display: block;
}

/* 移动端优化：仓库提交记录卡片 */
.card-message {
  display: -webkit-box;
  display: flex;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.card-files {
  display: -webkit-box;
  display: flex;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

/* 详情模态框内容区 */
.modal-body {
  padding: 20px;
  max-height: 70vh;
  overflow-y: auto;
}

/* 加载状态样式 */
.loading-content {
  text-align: center;
  padding: 40px 20px;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid rgba(255, 110, 180, 0.2);
  border-top: 4px solid var(--primary-color);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 20px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.desc-multiline {
  display: block;
  word-break: break-all;
  white-space: pre-line;
}

.group-select {
  padding: 6px 12px;
  border-radius: 6px;
  border: 1px solid #ffb6d5;
  font-size: 1rem;
  color: #ff6eb4;
  background: #fff6fb;
  outline: none;
  margin-left: 4px;
}
.group-select:focus {
  border-color: #ff6eb4;
  box-shadow: 0 0 0 2px #ffe0f0;
}
</style>
