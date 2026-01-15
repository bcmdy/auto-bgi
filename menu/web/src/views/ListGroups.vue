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
        <div class="modern-loading">
          <div class="loading-orbits">
            <div class="orbit orbit-1"></div>
            <div class="orbit orbit-2"></div>
            <div class="orbit orbit-3"></div>
            <div class="loading-core">
              <div class="core-inner"></div>
            </div>
          </div>
          <div class="loading-particles">
            <div class="particle" v-for="i in 12" :key="i" :style="{
              '--i': i,
              '--delay': i * 0.1 + 's'
            }"></div>
          </div>
          <div class="loading-text-container">
            <p class="loading-text">正在加载配置组</p>
            <div class="loading-dots">
              <span class="dot"></span>
              <span class="dot"></span>
              <span class="dot"></span>
            </div>
          </div>
        </div>
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
            <button class="btn ghost" @click="$router.push('/')">
              <span class="btn-icon">🏠</span> 
              <span class="btn-text">主页</span>
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

    <!-- 全局加载遮罩 -->
    <transition name="fade-modal">
      <div class="loading-overlay" v-if="showLoadingOverlay">
        <div class="loading-spinner-container">
          <div class="spinner-ring"></div>
          <div class="spinner-ring"></div>
          <div class="spinner-ring"></div>
          <div class="loading-message">
            <div class="message-icon">🚀</div>
            <h3>正在启动服务...</h3>
            <p>已选择 {{ selectedGroups.length }} 个服务</p>
            <div class="progress-dots">
              <span></span><span></span><span></span>
            </div>
          </div>
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
const showLoadingOverlay = ref(false)
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
  showLoadingOverlay.value = true
  try {
    await apiMethods.startGroups(selectedGroups.value)
    message.success(`成功启动 ${selectedGroups.value.length} 个服务`)
    clearSelection()
  } catch (error) {
    message.error('批量启动失败')
  } finally {
    isStarting.value = false
    showLoadingOverlay.value = false
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
  padding-bottom: 100px;
  animation: pageFadeIn 0.6s ease-out;
}

@keyframes pageFadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 移动端优化 */
@media (max-width: 768px) {
  .list-groups-page {
    padding-bottom: 120px;
  }
}

/* 头部样式优化 - 现代化设计 */
.page-header {
  position: relative;
  height: 180px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
  border-radius: 0 0 30px 30px;
  overflow: hidden;
  box-shadow: 0 8px 32px rgba(255,110,180,0.2);
  background: linear-gradient(135deg, #ff6eb4 0%, #ff8cc8 50%, #ffb3d9 100%);
  margin-bottom: 30px;
  position: relative;
  isolation: isolate;
}

.page-header::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 100%;
  background: linear-gradient(45deg, transparent 30%, rgba(255,255,255,0.1) 50%, transparent 70%);
  animation: shimmer 3s infinite linear;
  z-index: 1;
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

@media (max-width: 768px) {
  .page-header {
    height: 140px;
    border-radius: 0 0 20px 20px;
    margin-bottom: 20px;
  }
}

@media (max-width: 480px) {
  .page-header {
    height: 120px;
    border-radius: 0 0 15px 15px;
    margin-bottom: 15px;
  }
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
  font-size: 2.2rem;
  color: white;
  margin: 0;
  text-shadow: 0 4px 12px rgba(0,0,0,0.2);
  font-weight: 800;
  letter-spacing: -0.5px;
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 12px;
  animation: titleGlow 3s ease-in-out infinite alternate;
}

@keyframes titleGlow {
  0% { text-shadow: 0 4px 12px rgba(0,0,0,0.2), 0 0 20px rgba(255,255,255,0.3); }
  100% { text-shadow: 0 4px 20px rgba(0,0,0,0.3), 0 0 30px rgba(255,255,255,0.5); }
}

.title-decoration {
  font-size: 1.8rem;
  animation: floatDecoration 3s ease-in-out infinite;
  display: inline-block;
}

.title-decoration:nth-child(1) { animation-delay: 0s; }
.title-decoration:nth-child(2) { animation-delay: 1.5s; }

@keyframes floatDecoration {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-8px) rotate(10deg); }
}

.subtitle {
  color: rgba(255,255,255,0.9);
  margin-top: 8px;
  font-size: 1rem;
  font-weight: 500;
  letter-spacing: 0.5px;
  text-shadow: 0 2px 4px rgba(0,0,0,0.2);
}

@media (max-width: 768px) {
  .page-title {
    font-size: 1.6rem;
    gap: 8px;
  }

  .title-decoration {
    font-size: 1.3rem;
  }

  .subtitle {
    font-size: 0.85rem;
    margin-top: 6px;
  }
}

@media (max-width: 480px) {
  .page-title {
    font-size: 1.3rem;
    gap: 6px;
  }

  .title-decoration {
    font-size: 1.1rem;
  }

  .subtitle {
    font-size: 0.75rem;
    margin-top: 4px;
  }
}

/* 布局容器 */
.container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 20px;
}

@media (max-width: 768px) {
  .container {
    padding: 0 12px;
  }
}

.main-content {
  margin-top: 20px;
  position: relative;
  z-index: 3;
}

@media (max-width: 768px) {
  .main-content {
    margin-top: 15px;
  }
}

/* 操作栏 */
.groups-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}

.groups-header h2 {
  margin: 0;
  font-size: 1.1rem;
  display: flex;
  align-items: center;
  color: #333;
  font-weight: 600;
}

@media (max-width: 768px) {
  .groups-header {
    margin-bottom: 12px;
    gap: 8px;
  }
  .groups-header h2 {
    font-size: 0.9rem;
  }
}

.groups-count {
  background: var(--primary-color);
  color: white;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 0.8rem;
  margin-left: 8px;
}

@media (max-width: 768px) {
  .groups-count {
    padding: 1px 6px;
    font-size: 0.7rem;
    margin-left: 6px;
  }
}

.header-right {
  display: flex;
  gap: 8px;
}

/* 网格系统 - 现代化响应式设计 */
.groups-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 24px;
  animation: gridAppear 0.6s ease-out;
}

@keyframes gridAppear {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 超小屏幕 - 单列布局 */
@media (max-width: 480px) {
  .groups-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }
}

/* 小屏幕 - 双列 */
@media (min-width: 481px) and (max-width: 768px) {
  .groups-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
  }
}

/* 平板 - 三列 */
@media (min-width: 769px) and (max-width: 1024px) {
  .groups-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
  }
}

/* 中等屏幕 - 四列 */
@media (min-width: 1025px) and (max-width: 1440px) {
  .groups-grid {
    grid-template-columns: repeat(4, 1fr);
    gap: 24px;
  }
}

/* 大屏幕 - 五列 */
@media (min-width: 1441px) {
  .groups-grid {
    grid-template-columns: repeat(5, 1fr);
    gap: 28px;
  }
}

/* 卡片样式优化 - 现代化设计 */
.group-card {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95) 0%, rgba(255, 255, 255, 0.85) 100%);
  border-radius: 20px;
  padding: 24px 20px;
  position: relative;
  border: 2px solid rgba(255, 255, 255, 0.8);
  box-shadow:
    0 8px 32px rgba(255, 110, 180, 0.12),
    0 2px 8px rgba(0, 0, 0, 0.05),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  backdrop-filter: blur(12px);
  cursor: pointer;
  user-select: none;
  min-height: 160px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  overflow: hidden;
  animation: cardAppear 0.5s ease-out forwards;
  opacity: 0;
  transform: translateY(20px);
}

.group-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, #ff6eb4, #ff8cc8, #ffb3d9);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.group-card:hover::before {
  opacity: 1;
}

@keyframes cardAppear {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 768px) {
  .group-card {
    border-radius: 16px;
    padding: 20px 16px;
    min-height: 140px;
  }
}

@media (max-width: 480px) {
  .group-card {
    border-radius: 14px;
    padding: 18px 14px;
    min-height: 130px;
  }
}

/* Hover效果 - 仅PC端 */
@media (min-width: 769px) {
  .group-card:hover {
    transform: translateY(-8px) scale(1.02);
    box-shadow:
      0 20px 60px rgba(255, 110, 180, 0.25),
      0 8px 24px rgba(0, 0, 0, 0.1),
      inset 0 1px 0 rgba(255, 255, 255, 0.9);
    border-color: rgba(255, 110, 180, 0.4);
  }
}

.group-card:active {
  transform: scale(0.98) !important;
  transition: transform 0.1s ease;
}

/* 选中状态 */
.group-card.selected {
  border-color: var(--primary-color);
  background: linear-gradient(135deg, rgba(255, 110, 180, 0.1) 0%, rgba(255, 140, 200, 0.15) 100%);
  box-shadow:
    0 16px 48px rgba(255, 110, 180, 0.25),
    0 4px 16px rgba(0, 0, 0, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.9);
  animation: selectedPulse 2s ease-in-out infinite;
  position: relative;
  overflow: hidden;
}

.group-card.selected::after {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: linear-gradient(
    45deg,
    transparent 30%,
    rgba(255, 255, 255, 0.1) 50%,
    transparent 70%
  );
  animation: selectedShimmer 3s linear infinite;
  z-index: 1;
  pointer-events: none;
}

@keyframes selectedShimmer {
  0% { transform: translateX(-100%) translateY(-100%) rotate(45deg); }
  100% { transform: translateX(100%) translateY(100%) rotate(45deg); }
}

@keyframes selectedPulse {
  0%, 100% {
    box-shadow:
      0 16px 48px rgba(255, 110, 180, 0.25),
      0 4px 16px rgba(0, 0, 0, 0.08),
      inset 0 1px 0 rgba(255, 255, 255, 0.9);
  }
  50% {
    box-shadow:
      0 20px 60px rgba(255, 110, 180, 0.35),
      0 6px 20px rgba(0, 0, 0, 0.12),
      inset 0 1px 0 rgba(255, 255, 255, 0.9);
  }
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
  flex-direction: column;
  align-items: center;
  text-align: center;
  margin-bottom: 12px;
  border: 2px solid rgba(255, 110, 180, 0.15);
  border-radius: 16px;
  padding: 16px 12px;
  background: rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(4px);
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.group-card:hover .card-header {
  border-color: rgba(255, 110, 180, 0.3);
  background: rgba(255, 255, 255, 0.8);
  transform: translateY(-2px);
}

.group-card.selected .card-header {
  border-color: rgba(255, 110, 180, 0.4);
  background: rgba(255, 255, 255, 0.9);
}

@media (max-width: 768px) {
  .card-header {
    padding: 12px 8px;
    margin-bottom: 8px;
    border-radius: 12px;
  }
}

@media (max-width: 480px) {
  .card-header {
    padding: 10px 6px;
    margin-bottom: 6px;
    border-radius: 10px;
  }
}

.group-icon {
  font-size: 2.5rem;
  margin-bottom: 12px;
  filter: drop-shadow(0 4px 8px rgba(0,0,0,0.15));
  animation: iconFloat 3s ease-in-out infinite;
  display: inline-block;
}

@keyframes iconFloat {
  0%, 100% { transform: translateY(0) scale(1); }
  50% { transform: translateY(-5px) scale(1.05); }
}

.group-card.selected .group-icon {
  animation: selectedIcon 1.5s ease-in-out infinite;
}

@keyframes selectedIcon {
  0%, 100% { transform: translateY(0) scale(1); filter: drop-shadow(0 4px 8px rgba(255, 110, 180, 0.3)); }
  50% { transform: translateY(-5px) scale(1.1); filter: drop-shadow(0 6px 12px rgba(255, 110, 180, 0.5)); }
}

@media (max-width: 768px) {
  .group-icon {
    font-size: 2rem;
    margin-bottom: 8px;
  }
}

@media (max-width: 480px) {
  .group-icon {
    font-size: 1.8rem;
    margin-bottom: 6px;
  }
}

.group-name {
  margin: 0;
  font-size: 1.1rem;
  color: #333;
  word-break: break-word;
  line-height: 1.4;
  font-weight: 700;
  letter-spacing: -0.3px;
  position: relative;
  padding: 0 4px;
}

.group-name::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 50%;
  transform: translateX(-50%);
  width: 0;
  height: 2px;
  background: linear-gradient(90deg, #ff6eb4, #ff8cc8);
  transition: width 0.3s ease;
}

.group-card:hover .group-name::after {
  width: 60%;
}

.group-card.selected .group-name::after {
  width: 80%;
  background: linear-gradient(90deg, #ff2f9d, #ff6eb4);
}

@media (max-width: 768px) {
  .group-name {
    font-size: 0.95rem;
    line-height: 1.3;
  }
}

@media (max-width: 480px) {
  .group-name {
    font-size: 0.9rem;
    line-height: 1.25;
  }
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

/* 底部悬浮操作栏优化 - 现代化设计 */
.selection-footer {
  position: fixed;
  bottom: 30px;
  left: 50%;
  transform: translateX(-50%);
  width: 90%;
  max-width: 700px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.98) 0%, rgba(255, 255, 255, 0.95) 100%);
  backdrop-filter: blur(25px) saturate(180%);
  padding: 16px 24px;
  border-radius: 60px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow:
    0 20px 60px rgba(255, 110, 180, 0.3),
    0 8px 32px rgba(0, 0, 0, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.9);
  border: 2px solid rgba(255, 255, 255, 0.8);
  z-index: 100;
  gap: 16px;
  animation: footerSlideUp 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes footerSlideUp {
  from {
    opacity: 0;
    transform: translate(-50%, 20px);
  }
  to {
    opacity: 1;
    transform: translate(-50%, 0);
  }
}

.selection-footer::before {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(135deg, #ff6eb4, #ff8cc8, #ffb3d9);
  border-radius: 62px;
  z-index: -1;
  opacity: 0.3;
  filter: blur(4px);
}

@media (max-width: 768px) {
  .selection-footer {
    width: 94%;
    max-width: 600px;
    padding: 14px 20px;
    bottom: 20px;
    gap: 12px;
    border-radius: 50px;
  }
}

@media (max-width: 480px) {
  .selection-footer {
    width: 96%;
    padding: 12px 16px;
    bottom: 16px;
    gap: 10px;
    border-radius: 40px;
  }
}

@media (max-width: 380px) {
  .selection-footer {
    width: 98%;
    padding: 10px 14px;
    bottom: 12px;
    border-radius: 35px;
  }
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
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 0.9rem;
  flex-shrink: 0;
  box-shadow: 0 3px 10px rgba(255, 47, 157, 0.4);
}

@media (max-width: 768px) {
  .count-badge {
    width: 26px;
    height: 26px;
    font-size: 0.8rem;
  }
}

@media (max-width: 480px) {
  .count-badge {
    width: 22px;
    height: 22px;
    font-size: 0.75rem;
  }
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
/* 通用按钮样式优化 */
.btn {
  border: none;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-weight: 600;
  outline: none;
  -webkit-tap-highlight-color: transparent;
}

.btn:active {
  transform: scale(0.95);
  transition: transform 0.1s ease;
}

/* 按钮涟漪效果 */
.btn {
  position: relative;
  overflow: hidden;
}

.btn::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 5px;
  height: 5px;
  background: rgba(255, 255, 255, 0.5);
  opacity: 0;
  border-radius: 100%;
  transform: scale(1, 1) translate(-50%, -50%);
  transform-origin: 50% 50%;
}

.btn:focus:not(:active)::after {
  animation: ripple 1s ease-out;
}

@keyframes ripple {
  0% {
    transform: scale(0, 0);
    opacity: 0.5;
  }
  20% {
    transform: scale(25, 25);
    opacity: 0.3;
  }
  100% {
    opacity: 0;
    transform: scale(40, 40);
  }
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn.primary {
  background: linear-gradient(135deg, #ff6eb4, #ff8cc8);
  color: white;
  border-radius: 24px;
  padding: 10px 20px;
  font-size: 0.95rem;
  box-shadow: 0 4px 15px rgba(255, 110, 180, 0.4);
}

@media (min-width: 769px) {
  .btn.primary:hover:not(:disabled) {
    box-shadow: 0 6px 20px rgba(255, 110, 180, 0.5);
    transform: translateY(-2px);
  }
}

@media (max-width: 768px) {
  .btn.primary {
    padding: 8px 16px;
    font-size: 0.9rem;
  }
}

@media (max-width: 480px) {
  .btn.primary {
    padding: 6px 14px;
    font-size: 0.85rem;
  }
}

.btn.ghost {
  background: #f5f5f5;
  color: #666;
  border-radius: 24px;
  padding: 10px 20px;
  font-size: 0.95rem;
  border: 1px solid #e9b0b0;
}

.btn.ghost:hover:not(:disabled) {
  background: #ffe6f2;
  border-color: #ff8cc8;
  color: #ff6eb4;
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(255, 110, 180, 0.2);
}


@media (max-width: 768px) {
  .btn.ghost {
    padding: 8px 16px;
    font-size: 0.9rem;
  }
}

@media (max-width: 480px) {
  .btn.ghost {
    padding: 6px 14px;
    font-size: 0.85rem;
  }
}

.btn.small {
  padding: 6px 12px;
  font-size: 0.85rem;
}

@media (max-width: 480px) {
  .btn.small {
    padding: 5px 10px;
    font-size: 0.8rem;
  }
}

/* 动画 */
.slide-up-enter-active, .slide-up-leave-active { transition: all 0.3s ease; }
.slide-up-enter-from, .slide-up-leave-to { opacity: 0; transform: translate(-50%, 20px); }

/* 针对极小屏幕适配 */
@media (max-width: 380px) {
  .btn-text { display: none; }
  .page-title { font-size: 1.2rem; }
  .selection-footer { width: 98%; padding: 8px 10px; }
  .container { padding: 0 10px; }
  .groups-header h2 { font-size: 0.85rem; }
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

/* --- 弹窗样式 - 现代化设计 --- */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(10px) saturate(180%);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  animation: overlayFade 0.3s ease;
}

@keyframes overlayFade {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-content {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.98) 0%, rgba(255, 255, 255, 0.95) 100%);
  width: 100%;
  max-width: 400px;
  border-radius: 28px;
  box-shadow:
    0 32px 96px rgba(0, 0, 0, 0.25),
    0 8px 32px rgba(255, 110, 180, 0.15),
    inset 0 1px 0 rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  max-height: 80vh;
  animation: modalPopIn 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
  border: 2px solid rgba(255, 255, 255, 0.8);
  overflow: hidden;
  position: relative;
}

.modal-content::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, #ff6eb4, #ff8cc8, #ffb3d9);
}

@keyframes modalPopIn {
  from {
    opacity: 0;
    transform: scale(0.9) translateY(20px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

@media (max-width: 768px) {
  .modal-overlay {
    padding: 16px;
  }

  .modal-content {
    max-width: 360px;
    border-radius: 24px;
  }
}

@media (max-width: 480px) {
  .modal-overlay {
    padding: 12px;
  }

  .modal-content {
    max-width: 320px;
    border-radius: 20px;
  }
}

@media (max-width: 380px) {
  .modal-content {
    max-width: 280px;
    border-radius: 18px;
  }
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

/* ========== 现代化加载动画 ========== */
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  padding: 40px;
}

.modern-loading {
  position: relative;
  width: 200px;
  height: 200px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.loading-orbits {
  position: relative;
  width: 160px;
  height: 160px;
}

.orbit {
  position: absolute;
  border-radius: 50%;
  border: 2px solid transparent;
  animation: orbitSpin linear infinite;
}

.orbit-1 {
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border-top-color: #ff6eb4;
  border-right-color: #ff6eb4;
  animation-duration: 3s;
}

.orbit-2 {
  top: 20px;
  left: 20px;
  right: 20px;
  bottom: 20px;
  border-top-color: #ff8cc8;
  border-right-color: #ff8cc8;
  animation-duration: 2.5s;
  animation-direction: reverse;
}

.orbit-3 {
  top: 40px;
  left: 40px;
  right: 40px;
  bottom: 40px;
  border-top-color: #ffb3d9;
  border-right-color: #ffb3d9;
  animation-duration: 2s;
}

@keyframes orbitSpin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.loading-core {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #ff6eb4, #ff8cc8);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 0 30px rgba(255, 110, 180, 0.5);
  animation: corePulse 2s ease-in-out infinite;
}

.core-inner {
  width: 16px;
  height: 16px;
  background: white;
  border-radius: 50%;
  animation: innerSpin 1.5s linear infinite;
}

@keyframes corePulse {
  0%, 100% { transform: translate(-50%, -50%) scale(1); }
  50% { transform: translate(-50%, -50%) scale(1.1); }
}

@keyframes innerSpin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.loading-particles {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

.particle {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 4px;
  height: 4px;
  background: #ff6eb4;
  border-radius: 50%;
  transform: translate(-50%, -50%) rotate(calc(var(--i) * 30deg)) translateX(80px);
  opacity: 0;
  animation: particleOrbit 2s linear infinite;
  animation-delay: var(--delay);
}

@keyframes particleOrbit {
  0% {
    opacity: 0;
    transform: translate(-50%, -50%) rotate(calc(var(--i) * 30deg)) translateX(80px);
  }
  10% {
    opacity: 1;
  }
  90% {
    opacity: 1;
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) rotate(calc(var(--i) * 30deg + 360deg)) translateX(80px);
  }
}

.loading-text-container {
  margin-top: 100px;
  text-align: center;
}

.loading-text {
  font-size: 1.1rem;
  color: #666;
  font-weight: 600;
  margin-bottom: 12px;
}

.loading-dots {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.dot {
  width: 8px;
  height: 8px;
  background: #ff6eb4;
  border-radius: 50%;
  animation: dotBounce 1.4s ease-in-out infinite;
}

.dot:nth-child(1) { animation-delay: 0s; }
.dot:nth-child(2) { animation-delay: 0.2s; }
.dot:nth-child(3) { animation-delay: 0.4s; }

@keyframes dotBounce {
  0%, 60%, 100% {
    transform: translateY(0);
    opacity: 0.6;
  }
  30% {
    transform: translateY(-10px);
    opacity: 1;
  }
}

@media (max-width: 768px) {
  .loading-container {
    min-height: 300px;
    padding: 30px;
  }

  .modern-loading {
    width: 160px;
    height: 160px;
  }

  .loading-orbits {
    width: 120px;
    height: 120px;
  }

  .orbit-2 {
    top: 15px;
    left: 15px;
    right: 15px;
    bottom: 15px;
  }

  .orbit-3 {
    top: 30px;
    left: 30px;
    right: 30px;
    bottom: 30px;
  }

  .loading-core {
    width: 32px;
    height: 32px;
  }

  .core-inner {
    width: 12px;
    height: 12px;
  }

  .particle {
    transform: translate(-50%, -50%) rotate(calc(var(--i) * 30deg)) translateX(60px);
  }

  .loading-text-container {
    margin-top: 80px;
  }

  .loading-text {
    font-size: 1rem;
  }
}

@media (max-width: 480px) {
  .loading-container {
    min-height: 250px;
    padding: 20px;
  }

  .modern-loading {
    width: 140px;
    height: 140px;
  }

  .loading-orbits {
    width: 100px;
    height: 100px;
  }

  .orbit-2 {
    top: 12px;
    left: 12px;
    right: 12px;
    bottom: 12px;
  }

  .orbit-3 {
    top: 24px;
    left: 24px;
    right: 24px;
    bottom: 24px;
  }

  .loading-core {
    width: 28px;
    height: 28px;
  }

  .core-inner {
    width: 10px;
    height: 10px;
  }

  .particle {
    transform: translate(-50%, -50%) rotate(calc(var(--i) * 30deg)) translateX(50px);
  }

  .loading-text-container {
    margin-top: 70px;
  }

  .loading-text {
    font-size: 0.9rem;
  }
}

/* ========== 全局加载遮罩样式 ========== */
.loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(8px);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.loading-spinner-container {
  position: relative;
  width: 280px;
  height: 280px;
  display: flex;
  align-items: center;
  justify-content: center;
}

@media (max-width: 768px) {
  .loading-spinner-container {
    width: 240px;
    height: 240px;
  }
}

/* 三层旋转环 */
.spinner-ring {
  position: absolute;
  border-radius: 50%;
  border: 4px solid transparent;
  border-top-color: #ff6eb4;
  animation: spin 1.5s cubic-bezier(0.68, -0.55, 0.265, 1.55) infinite;
}

.spinner-ring:nth-child(1) {
  width: 200px;
  height: 200px;
  border-width: 6px;
  animation-duration: 2s;
}

.spinner-ring:nth-child(2) {
  width: 150px;
  height: 150px;
  border-top-color: #ff8cc8;
  border-width: 5px;
  animation-duration: 1.5s;
  animation-direction: reverse;
}

.spinner-ring:nth-child(3) {
  width: 100px;
  height: 100px;
  border-top-color: #ffb3d9;
  border-width: 4px;
  animation-duration: 1s;
}

@media (max-width: 768px) {
  .spinner-ring:nth-child(1) {
    width: 160px;
    height: 160px;
  }
  .spinner-ring:nth-child(2) {
    width: 120px;
    height: 120px;
  }
  .spinner-ring:nth-child(3) {
    width: 80px;
    height: 80px;
  }
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

/* 加载消息 */
.loading-message {
  position: absolute;
  text-align: center;
  color: white;
  z-index: 10;
}

.message-icon {
  font-size: 3rem;
  margin-bottom: 12px;
  animation: bounce 1s ease-in-out infinite;
}

@media (max-width: 768px) {
  .message-icon {
    font-size: 2.5rem;
  }
}

.loading-message h3 {
  margin: 0 0 8px 0;
  font-size: 1.3rem;
  font-weight: 700;
  text-shadow: 0 2px 8px rgba(0,0,0,0.3);
}

.loading-message p {
  margin: 0 0 15px 0;
  font-size: 0.95rem;
  color: rgba(255,255,255,0.9);
}

@media (max-width: 768px) {
  .loading-message h3 {
    font-size: 1.1rem;
  }
  .loading-message p {
    font-size: 0.85rem;
  }
}

.progress-dots {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.progress-dots span {
  width: 8px;
  height: 8px;
  background: white;
  border-radius: 50%;
  animation: dotPulse 1.4s ease-in-out infinite;
}

.progress-dots span:nth-child(1) { animation-delay: 0s; }
.progress-dots span:nth-child(2) { animation-delay: 0.2s; }
.progress-dots span:nth-child(3) { animation-delay: 0.4s; }

@keyframes dotPulse {
  0%, 60%, 100% { 
    transform: scale(1); 
    opacity: 0.7;
  }
  30% { 
    transform: scale(1.5); 
    opacity: 1;
  }
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-15px); }
}
</style>