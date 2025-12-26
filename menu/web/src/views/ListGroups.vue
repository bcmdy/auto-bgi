<template>
  <div class="list-groups-page">
    <div class="floating-hearts">
      <div class="heart" v-for="i in 15" :key="i" :style="{ animationDelay: (i * 0.5) + 's' }">♡</div>
    </div>
    
    <header class="page-header">
      <div class="header-carousel" v-if="carouselImages.length > 0">
        <div class="carousel-container">
          <div v-for="(image, index) in carouselImages" :key="index" class="carousel-slide" :class="{ active: currentImageIndex === index }">
            <img :src="image" :alt="`carousel-${index}`" />
          </div>
        </div>
      </div>
      
      <div class="header-overlay"></div> <div class="header-decoration">
        <div class="sparkle">✨</div>
        <div class="sparkle">⭐</div>
        <div class="sparkle">💫</div>
      </div>
      
      <div class="container header-content">
        <h1 class="page-title">
           <span class="title-decoration">🎀</span>
           {{ pageTitle }}
           <span class="title-decoration">🎀</span>
        </h1>
        <p class="subtitle">管理并启动您的配置实例</p>
      </div>
    </header>

    <div class="container main-content">
      <div v-if="loading" class="loading-container">
        <div class="loading-animation">
          <div class="loading-heart">💖</div>
          <div class="loading-dots">
            <span></span><span></span><span></span>
          </div>
        </div>
        <p class="loading-text">正在加载配置组...</p>
      </div>

      <div v-else-if="groups.length > 0" class="groups-container">
        <div class="groups-header">
          <div class="header-left">
            <h2>
              <span class="header-icon">📋</span>
              配置列表
              <span class="groups-count">{{ groups.length }}</span>
            </h2>
          </div>
          
          <div class="header-right">
            <button class="btn home-btn" @click="$router.push('/')">
              <span class="btn-icon">🏠</span> <span class="btn-text">主页</span>
            </button>
            <button class="btn ghost" @click="toggleSelectAll">
              <span class="btn-icon">✅</span>
              <span class="btn-text">{{ isAllSelected ? '取消全选' : '全选' }}</span>
            </button>
          </div>
        </div>

        <div class="groups-grid">
          <div 
            v-for="(group, index) in groups" 
            :key="group" 
            class="group-card"
            :class="{ selected: isSelected(group) }"
            :style="{ animationDelay: (index * 0.05) + 's' }"
            @click="toggleSelect(group)"
          >
            <div class="select-checkbox-wrapper">
              <div class="checkbox-ui" :class="{ checked: isSelected(group) }"></div>
            </div>

            <div class="card-header">
              <div class="group-icon">⚙️</div>
              <h3 class="group-name">{{ group }}</h3>
            </div>
            
            <div class="card-status" v-if="isSelected(group)">
               <span class="selected-tag">已选择</span>
            </div>
  
       
          </div>
        </div>
      </div>

      <div v-else class="empty-state">
        <div class="empty-icon">📭</div>
        <h3>暂无配置组</h3>
        <button class="btn reload-btn" @click="loadGroups">
          <span class="btn-icon">🔄</span> 重新加载
        </button>
      </div>
    </div>

<transition name="slide-up">
      <div class="selection-footer" v-if="selectedGroups.length > 0">
        <div class="selection-info">
          
          <div class="count-badge clickable" @click="showDetailModal = true">
            {{ selectedGroups.length }}
            <span class="badge-hint">🔍</span>
          </div>
          
          <div class="selection-text-col">
            <span class="label">已选择:</span>
            <span class="preview-text">{{ selectionPreview }}</span>
          </div>
        </div>

        <div class="selection-actions">
          <button class="btn ghost small" @click="clearSelection">取消</button>
          <button class="btn primary glow" @click="startSelected" :disabled="isStarting">
            <span class="btn-icon">🚀</span> 启动
          </button>
        </div>
      </div>
    </transition>

    <transition name="fade-modal">
      <div class="modal-overlay" v-if="showDetailModal" @click="showDetailModal = false">
        <div class="modal-content" @click.stop>
          <div class="modal-header">
            <h3>已选列表 ({{ selectedGroups.length }})</h3>
            <button class="close-btn" @click="showDetailModal = false">✕</button>
          </div>
          <div class="modal-body">
            <div v-for="name in selectedGroups" :key="name" class="modal-item">
              <span class="item-icon">⚙️</span>
              <span class="item-text">{{ name }}</span>
              <button class="item-remove" @click.stop="toggleSelect(name)">✕</button>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn primary full-width" @click="showDetailModal = false">确定</button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { message } from 'ant-design-vue'
import api, { apiMethods } from '@/utils/api'

// 响应式数据
const pageTitle = ref('配置组列表')
const groups = ref([])
const loading = ref(true)
const isStarting = ref(false)
const selectedGroups = ref([])
const carouselImages = ref([])
const currentImageIndex = ref(0)
let carouselInterval = null
const showDetailModal = ref(false)

// 计算属性
const isAllSelected = computed(() => {
  return groups.value.length > 0 && selectedGroups.value.length === groups.value.length
})

// 显示选中的文本预览（如：GroupA, GroupB...）
const selectedGroupsText = computed(() => {
  const text = selectedGroups.value.join(', ')
  return text.length > 20 ? text.substring(0, 20) + '...' : text
})

//智能生成预览文字
const selectionPreview = computed(() => {
  const count = selectedGroups.value.length
  if (count === 0) return ''
  
  // 策略：手机屏幕小，只显示前 2 个名字，剩下的显示数量
  const maxNames = 2 
  
  if (count <= maxNames) {
    return selectedGroups.value.join(', ')
  } else {
    const firstFew = selectedGroups.value.slice(0, maxNames).join(', ')
    const remaining = count // 这里显示总数，或者 count - maxNames 显示剩余数
    return `${firstFew} 等 ${remaining} 个`
  }
})

// 获取轮播图图片
const getImages = async () => {
  try {
    const response = await fetch('/api/images')
    if (!response.ok) throw new Error('Failed')
    const data = await response.json()
    carouselImages.value = data.images || []
    if (carouselImages.value.length > 0) startCarousel()
  } catch (error) {
    // 默认图片
    carouselImages.value = ['/img/bd.jpg', '/img/ff.png']
    startCarousel()
  }
}

const startCarousel = () => {
  if (carouselImages.value.length > 1) {
    carouselInterval = setInterval(() => {
      currentImageIndex.value = (currentImageIndex.value + 1) % carouselImages.value.length
    }, 5000)
  }
}

// 加载配置组数据
const loadGroups = async () => {
  loading.value = true
  try {
    const response = await apiMethods.getListGroups()
    if (response && response.items) {
      groups.value = response.items
      pageTitle.value = response.title || '配置组列表'
    } else if (Array.isArray(response)) {
      groups.value = response
    } else {
      groups.value = []
    }
  } catch (error) {
    console.error('API Error:', error)
    // 模拟数据用于展示效果 (实际使用请删除)
    groups.value = ['LoginServer', 'GameServer', 'ChatService', 'Database', 'Gateway']
  } finally {
    loading.value = false
  }
}

// 启动单个
const startGroup = async (groupName) => {
  if (isStarting.value) return
  isStarting.value = true
  try {
    await apiMethods.startGroups([groupName])
    message.success(`已启动: ${groupName}`)
  } catch (error) {
    message.error('启动失败')
  } finally {
    isStarting.value = false
  }
}

// 多选逻辑优化
const isSelected = (groupName) => selectedGroups.value.includes(groupName)

const toggleSelect = (groupName) => {
  if (isSelected(groupName)) {
    selectedGroups.value = selectedGroups.value.filter(name => name !== groupName)
  } else {
    selectedGroups.value = [...selectedGroups.value, groupName]
  }
}

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedGroups.value = []
  } else {
    selectedGroups.value = [...groups.value]
  }
}

const clearSelection = () => {
  selectedGroups.value = []
}

const startSelected = async () => {
  if (isStarting.value || selectedGroups.value.length === 0) return
  
  isStarting.value = true
  try {
    await apiMethods.startGroups(selectedGroups.value)
    message.success(`成功启动 ${selectedGroups.value.length} 个服务`)
    clearSelection()
  } catch (error) {
    message.error('批量启动失败')
  } finally {
    isStarting.value = false
  }
}

onMounted(() => {
  loadGroups()
  getImages()
})

onUnmounted(() => {
  if (carouselInterval) clearInterval(carouselInterval)
})
</script>

<style scoped>
:root {
  --primary-color: #ff6eb4;
  --secondary-color: #ff8cc8;
  --bg-gradient: linear-gradient(135deg, #fff6fb 0%, #ffe6f2 100%);
  --card-bg: rgba(255, 255, 255, 0.85);
  --glass-border: 1px solid rgba(255, 255, 255, 0.6);
  --shadow-sm: 0 4px 12px rgba(255, 110, 180, 0.1);
  --shadow-hover: 0 10px 25px rgba(255, 110, 180, 0.25);
}

* { box-sizing: border-box; -webkit-tap-highlight-color: transparent; }

.list-groups-page {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  background: var(--bg-gradient);
  color: #444;
  min-height: 100vh;
  padding-bottom: 80px; /* 减小底部空间 */
}

/* 头部样式优化 */
.page-header {
  position: relative;
  height: 120px; /* 进一步减小头部高度 */
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
  border-radius: 0 0 15px 15px;
  overflow: hidden;
  box-shadow: 0 3px 15px rgba(255,110,180,0.12);
}

.header-carousel, .carousel-slide img, .header-overlay {
  position: absolute;
  top: 0; left: 0; width: 100%; height: 100%;
}

.carousel-slide img {
  object-fit: cover;
}

.header-overlay {
  background: linear-gradient(to bottom, rgba(255,246,251,0.3), rgba(255,246,251,0.9));
  z-index: 1;
}

.header-content {
  position: relative;
  z-index: 2;
}

.page-title {
  font-size: 1.3rem; /* 进一步减小标题字体 */
  color: #e91e63;
  margin: 0;
  text-shadow: 0 2px 4px rgba(255,255,255,0.8);
}

.subtitle {
  color: #888;
  margin-top: 2px;
  font-size: 0.75rem; /* 进一步减小副标题 */
}

/* 布局容器 */
.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 10px; /* 减小左右边距 */
}

.main-content {
  margin-top: -15px; /* 进一步减小上移距离 */
  position: relative;
  z-index: 3;
}

/* 操作栏 */
.groups-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px; /* 进一步减小底部间距 */
  flex-wrap: wrap;
  gap: 6px;
}

.groups-header h2 {
  margin: 0;
  font-size: 0.9rem; /* 进一步减小标题字体 */
  display: flex;
  align-items: center;
  color: #333;
}

.groups-count {
  background: var(--primary-color);
  color: white;
  padding: 1px 6px;
  border-radius: 8px;
  font-size: 0.7rem;
  margin-left: 6px;
}

.header-right {
  display: flex;
  gap: 8px;
}

/* 网格系统 - 响应式核心 */
.groups-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); /* 手机端默认双列，最小140px */
  gap: 8px;
}

@media (min-width: 500px) {
  .groups-grid { 
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); /* 更大屏幕，卡片稍大 */
  }
}
@media (min-width: 768px) {
  .groups-grid { 
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); /* 平板 */
  }
}
@media (min-width: 1024px) {
  .groups-grid { 
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); /* PC端 */
  }
}

/* 卡片样式 */
.group-card {
  background: var(--card-bg);
  border-radius: 12px; /* 进一步减小圆角 */
  padding: 8px; /* 进一步减小内边距 */
  position: relative;
  border: 2px solid transparent;
  box-shadow: var(--shadow-sm);
  transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
  backdrop-filter: blur(10px);
  cursor: pointer;
  user-select: none;
}

.group-card:active {
  transform: scale(0.98);
}

/* 选中状态 */
.group-card.selected {
  border-color: var(--primary-color);
  background: #fff0f6;
  box-shadow: var(--shadow-hover);
}

.select-checkbox-wrapper {
  position: absolute;
  top: 6px; /* 调整位置 */
  right: 6px;
  z-index: 5;
}

.checkbox-ui {
  width: 18px; /* 进一步减小复选框 */
  height: 18px;
  border-radius: 50%;
  border: 2px solid #ddd;
  background: #fff;
  transition: all 0.2s;
  position: relative;
}

.checkbox-ui.checked {
  background: var(--primary-color);
  border-color: var(--primary-color);
}

.checkbox-ui.checked::after {
  content: '';
  position: absolute;
  top: 2px; left: 5px; /* 调整勾的位置 */
  width: 3px; height: 7px; /* 进一步减小勾的大小 */
  border: solid rgb(252, 5, 5);
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}

.card-header {
  display: flex;
  flex-direction: column; /* 手机端垂直排列更省空间 */
  align-items: center;
  text-align: center;
  margin-bottom: 4px; /* 进一步减小间距 */
  border: 2px solid rgba(245, 7, 122, 0.15); /* 进一步减小边框 */
  border-radius: 8px;
  padding: 6px 4px; /* 减小内边距 */
}

.group-icon {
  font-size: 1.4rem; /* 进一步减小图标 */
  margin-bottom: 3px;
}

.group-name {
  margin: 0;
  font-size: 0.85rem; /* 进一步减小字体 */
  color: #333;
  word-break: break-all;
  line-height: 1.2;
}

.selected-tag {
  display: block;
  text-align: center;
  font-size: 0.65rem; /* 进一步减小字体 */
  color: var(--primary-color);
  margin-bottom: 2px; /* 进一步减小间距 */
  font-weight: bold;
}

.card-actions .start-btn {
  width: 100%;
  padding: 10px;
  border-radius: 12px;
  border: 1px solid #ffcee5;
  background: white;
  color: var(--primary-color);
  font-weight: 600;
}

/* 底部悬浮操作栏 */
.selection-footer {
  position: fixed;
  bottom: 15px; /* 减小底部距离 */
  left: 50%;
  transform: translateX(-50%);
  width: 92%; /* 手机端稍微宽一点 */
  max-width: 450px; /* 减小最大宽度 */
  background: rgba(255, 255, 255, 0.98); /* 增加不透明度 */
  backdrop-filter: blur(15px);
  padding: 8px 12px; /* 减小内边距 */
  border-radius: 40px; /* 减小圆角 */
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 8px 30px rgba(0,0,0,0.18); /* 减小阴影 */
  border: 1px solid rgba(255,110,180, 0.3); /* 增加一点边框颜色 */
  z-index: 100;
  gap: 8px; /* 减小元素间距 */
}

.selection-info {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1; /* 占据剩余空间 */
  min-width: 0; /* 关键：允许flex子元素收缩，防止文字撑开容器 */
}

.count-badge {
  background: linear-gradient(135deg, #ff6eb4, #ff2f9d);
  color: white;
  width: 22px; /* 减小气泡 */
  height: 22px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 0.75rem; /* 减小字体 */
  flex-shrink: 0; /* 防止气泡被压扁 */
  box-shadow: 0 2px 6px rgba(255, 47, 157, 0.4);
}

.selection-text-col {
  display: flex;
  flex-direction: column;
  justify-content: center;
  line-height: 1.2;
  overflow: hidden; /* 隐藏溢出文字 */
}

.selection-text-col .label {
  font-size: 0.65rem; /* 减小字体 */
  color: #999;
  margin-bottom: 1px;
}

.selection-text {
  display: flex;
  flex-direction: column;
  font-size: 0.8rem; /* 减小字体 */
  overflow: hidden;
}

.preview-text {
  color: #333;
  font-weight: 600;
  font-size: 0.85rem; /* 减小字体 */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis; /* 文字过长显示省略号 */
}

.selection-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0; /* 防止按钮被压缩 */
}
/* 通用按钮样式 */
.btn {
  border: none;
  cursor: pointer;
  transition: transform 0.2s;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
}
.btn:active { transform: scale(0.95); }

.btn.primary {
  background: linear-gradient(135deg, #ff6eb4, #ff8cc8);
  color: white;
  border-radius: 18px; /* 减小圆角 */
  padding: 6px 14px; /* 减小内边距 */
  font-weight: bold;
  font-size: 0.85rem; /* 减小字体 */
  box-shadow: 0 3px 12px rgba(255, 110, 180, 0.35); /* 减小阴影 */
}

.btn.ghost {
  background: #f0f0f0;
  color: #666;
  border-radius: 18px; /* 减小圆角 */
  padding: 6px 14px; /* 减小内边距 */
  font-size: 0.85rem; /* 减小字体 */
}
/* 按钮微调，适应小屏幕 */
.btn.small {
  padding: 5px 10px; /* 减小内边距 */
  font-size: 0.8rem; /* 减小字体 */
}

/* 动画 */
.slide-up-enter-active, .slide-up-leave-active { transition: all 0.3s ease; }
.slide-up-enter-from, .slide-up-leave-to { opacity: 0; transform: translate(-50%, 20px); }

/* 针对极小屏幕适配 */
@media (max-width: 380px) {
  .btn-text { display: none; } /* 隐藏部分按钮文字 */
  .page-title { font-size: 1.2rem; }
  .selection-footer { width: 95%; padding: 6px 10px; }
  .container { padding: 0 8px; }
}

/* 装饰性元素保持不变 (Heart, Sparkle等) - 省略以节省篇幅，保持原样即可 */
.floating-hearts, .heart { pointer-events: none; z-index: 0; }
.heart { position: fixed; color: #ffb3d9; animation: float 6s infinite ease-in-out; }
@keyframes float { 0% { transform: translateY(100vh); opacity: 0; } 50% { opacity: 1; } 100% { transform: translateY(-100px); opacity: 0; } }

/* --- 数字气泡交互优化 --- */
.count-badge.clickable {
  cursor: pointer;
  position: relative;
  transition: transform 0.2s;
  /* 增加点击区域 */
  border: 2px solid rgba(255,255,255,0.5);
}

.count-badge.clickable:active {
  transform: scale(0.9);
}

/* 小放大镜提示 */
.badge-hint {
  font-size: 8px;
  position: absolute;
  bottom: -2px;
  right: -2px;
  background: white;
  border-radius: 50%;
  width: 12px;
  height: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}

/* --- 弹窗样式 --- */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.4); /* 半透明遮罩 */
  backdrop-filter: blur(5px); /* 背景模糊 */
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.modal-content {
  background: rgba(255, 255, 255, 0.95);
  width: 100%;
  max-width: 320px;
  border-radius: 24px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
  max-height: 70vh; /* 防止弹窗过高 */
  animation: popIn 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.modal-header {
  padding: 15px 20px;
  border-bottom: 1px solid #eee;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--primary-color);
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.2rem;
  color: #999;
  cursor: pointer;
  padding: 5px;
}

.modal-body {
  padding: 10px;
  overflow-y: auto; /* 内容过多可滚动 */
  -webkit-overflow-scrolling: touch;
}

.modal-item {
  display: flex;
  align-items: center;
  padding: 10px 15px;
  background: #fff;
  border-radius: 12px;
  margin-bottom: 8px;
  border: 1px solid #f0f0f0;
  transition: all 0.2s;
}

.modal-item:last-child {
  margin-bottom: 0;
}

.item-icon {
  margin-right: 10px;
  font-size: 1.2rem;
}

.item-text {
  flex: 1;
  font-size: 0.95rem;
  font-weight: 500;
  color: #333;
}

.item-remove {
  background: #ffe0e0;
  color: #ff4d4f;
  border: none;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.modal-footer {
  padding: 15px;
  border-top: 1px solid #eee;
}

.full-width {
  width: 100%;
  justify-content: center;
}

/* 弹窗动画 */
@keyframes popIn {
  from { transform: scale(0.8); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}

.fade-modal-enter-active, .fade-modal-leave-active {
  transition: opacity 0.3s ease;
}
.fade-modal-enter-from, .fade-modal-leave-to {
  opacity: 0;
}
</style>