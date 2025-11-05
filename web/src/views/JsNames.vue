<template>
  <div class="js-names-page">
    <!-- Header轮播图 -->
    <header class="page-header" v-if="headerCarouselImages.length > 0">
      <div class="header-carousel">
        <div class="carousel-container">
          <div
            v-for="(image, index) in headerCarouselImages"
            :key="index"
            class="carousel-slide"
            :class="{ active: headerCurrentImageIndex === index }"
          >
            <img :src="image" :alt="`header-carousel-${index}`" loading="lazy" />
          </div>
        </div>
      </div>
      <div class="header-content">
        <h1 class="header-title">📜 脚本更新列表 📜</h1>
        <p class="header-subtitle">管理您的脚本，保持最新状态 ✨</p>
        <button class="btn home-btn" @click="goHome">返回首页</button>
        <button class="btn home-btn" style="margin-left:150px;" @click="batchUpdate">批量更新</button>
      </div>
    </header>

    <div class="container">
      <section class="panel">
        <h2>脚本信息</h2>
        <div id="pluginListContainer" class="table-container">
          <!-- 桌面端表格 -->
          <table class="desktop-table">
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
            <tbody>
              <tr v-if="pluginData.length === 0">
                <td colspan="5" style="text-align:center;">暂无插件数据。</td>
              </tr>
              <tr
                v-else
                v-for="item in sortedPluginData"
                :key="item.Name"
                :class="{ highlight: item.Mark === '有更新', highlighta: item.Mark === '未知' }"
      
              >
                <td>{{ item.ChineseName }}</td>
                <td>{{ item.NowVersion }}</td>
                <td>{{ item.NewVersion }}</td>
                <td>{{ item.Mark }}</td>
                <td>
                  <button
                    class="btn update-btn"
                    :class="{ 'haveUpdate-btn': item.Mark === '有更新' }"
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
            <div
              v-for="item in sortedPluginData"
              :key="item.Name"
              :class="{ highlight: item.Mark === '有更新', 'mobile-card': true, highlightm: item.Mark === '未知' }"
            >
              <div class="card-header">
                <div class="card-title">
                  <span class="title-icon">📝</span>
                  <span class="title-text" style="font-size: 18px;">{{ item.ChineseName }}</span>
                </div>
                <div class="card-versions">
                  <div class="version-item">
                    <span class="version-label">当前版本:</span>
                    <span class="version-value">{{ item.NowVersion }}</span>
                  </div>
                  <h1></h1>
                  <div class="version-item">
                    <span class="version-label">最新版本:</span>
                    <span class="version-value">{{ item.NewVersion }}</span>
                  </div>
                </div>
                <h1></h1>
                <div class="card-status">
                  <span class="status-icon">📊</span>
                  <span class="status-text" :class="{ 'status-update': item.Mark === '有更新' }">{{ item.Mark }}</span>
                </div>
              </div>
              <h1></h1>
              <div class="card-actions">
                <button
                  class="btn update-btn mobile-update-btn"
                  :class="{ 'haveUpdate-btn': item.Mark === '有更新' }"
                  :disabled="item.Mark !== '有更新' || isUpdating[item.Name]"
                  @click="updatePlugin(item.Name)"
                >
                  <span class="update-icon">{{ isUpdating[item.Name] ? '⏳' : '🔄' }}</span>
                  <span class="update-text">{{ isUpdating[item.Name] ? '更新中...' : (item.Mark === '有更新' ? '更新' : '已更新') }}</span>
                </button>
              </div>
            </div>
            <div v-if="pluginData.length === 0" class="empty-mobile">
              <div class="empty-content">
                <span class="empty-icon">📭</span>
                <span class="empty-text">暂无插件数据</span>
                <span class="empty-sparkle">✨</span>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, reactive, h } from 'vue'
import { useRouter } from 'vue-router'

export default {
  name: 'JsNames',
  setup() {
    const router = useRouter()
    const pluginData = ref([])
    const currentSort = ref({ key: 'ChineseName', asc: true })
    const isUpdating = reactive({})
    
    const headerCarouselImages = ref([])
    const headerCurrentImageIndex = ref(0)
    let headerCarouselInterval = null

    const getHeaderImages = async () => {
      try {
        const res = await fetch('/api/images')
        const data = await res.json()
        headerCarouselImages.value = data.images || []
      } catch {
        headerCarouselImages.value = ['/img/bd.jpg', '/img/ff.png', '/img/ng.jpg', '/img/sh.jpg']
      } finally {
        startHeaderCarousel()
      }
    }

    const startHeaderCarousel = () => {
      if (headerCarouselImages.value.length > 1) {
        headerCarouselInterval = setInterval(() => {
          headerCurrentImageIndex.value = (headerCurrentImageIndex.value + 1) % headerCarouselImages.value.length
        }, 7000)
      }
    }

    const sortedPluginData = computed(() => {
      return [...pluginData.value].sort((a, b) => {
        if (a.Mark === '有更新' && b.Mark !== '有更新') return -1
        if (a.Mark !== '有更新' && b.Mark === '有更新') return 1
        const valA = String(a[currentSort.value.key] || '')
        const valB = String(b[currentSort.value.key] || '')
        return currentSort.value.asc ? valA.localeCompare(valB) : valB.localeCompare(valA)
      })
    })

    const goHome = () => router.push('/')
    const sortTable = key => {
      if (currentSort.value.key === key) currentSort.value.asc = !currentSort.value.asc
      else { currentSort.value.key = key; currentSort.value.asc = true }
    }
    const getSortIcon = key => currentSort.value.key !== key ? 'sort-default' : (currentSort.value.asc ? 'sort-asc' : 'sort-desc')

    const loadPluginList = async () => {
      try {
        const res = await fetch('/api/jsNames')
        const json = await res.json()
        pluginData.value = json.data || []
      } catch {
        pluginData.value = []
      }
    }

    const batchUpdate = async () => {
      try {
        const res = await fetch('/api/batchUpdate', { method: 'POST' })
        alert('批量更新已执行')
        await loadPluginList()
      } catch {
        alert('批量更新失败')
      }
    }

    const updatePlugin = async name => {
      if (!name) return
      isUpdating[name] = true
      try {
        await fetch(`/api/updateJs/${name}`, { method: 'POST' })
        await loadPluginList()
      } finally {
        isUpdating[name] = false
      }
    }

    onMounted(() => {
      loadPluginList()
      getHeaderImages()
    })

    return {
      pluginData,
      sortedPluginData,
      currentSort,
      isUpdating,
      sortTable,
      getSortIcon,
      goHome,
      batchUpdate,
      updatePlugin,
      headerCarouselImages,
      headerCurrentImageIndex
    }
  }
}
</script>

<style scoped>
.js-names-page {
  background: #fff6fb;
  color: #333;
  font-family: 'Segoe UI', sans-serif;
}

.page-header {
  position: relative;
  text-align: center;
  padding: 2rem 0;
  background: #ffe4f0;
}

.header-carousel {
  position: relative;
  width: 100%;
  overflow: hidden;
  height: 200px;
}

.carousel-container {
  display: flex;
  transition: transform 0.5s ease-in-out;
}

.haveUpdate-btn {
  background-color: #2cce3a!important;
  color: #000!important;
}



.carousel-slide {
  flex: 0 0 100%;
  opacity: 0.9;
}

.carousel-slide img {
  width: 100%;
  height: 200px;
  object-fit: cover;
  display: block;
}



.header-content {
  margin-top: 1rem;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1rem;
}

.desktop-table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 1rem;
}

.desktop-table th,
.desktop-table td {
  border: 1px solid #ddd;
  padding: 0.75rem;
}

.sortable {
  cursor: pointer;
}

.sort-icon {
  margin-left: 0.25rem;
}

.mobile-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.mobile-card {
  border: 1px solid #f2cce7;
  border-radius: 12px;
  padding: 1rem;
  background: #fff0f7;
}

.card-header {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.card-actions {
  margin-top: 0.5rem;
}

.btn {
  padding: 0.5rem 1rem;
  border-radius: 8px;
  cursor: pointer;
  border: none;
  background-color: #f48fb1;
  color: white;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.highlight {
  background-color: #F48FB1 !important;
}

.highlighta {
  background-color: #928e90 !important;
  border : 2px solid #d18ba6 !important;
 
}

.highlightm{
  background-color: #928e90 !important;
  border : 2px solid #d18ba6 !important;
}

/* PC端隐藏移动端卡片列表 */
@media (min-width: 769px) {
  .mobile-list {
    display: none;
  }
}
/* 移动端隐藏桌面表格 */
@media (max-width: 768px) {
  .desktop-table {
    display: none;
  }
  .mobile-card {
    font-size: 0.9rem;
  }
  .highlight {
    background-color: #fff0f7 !important;
  }
}
</style>
