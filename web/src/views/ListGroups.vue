<template>
  <div class="list-groups-page">
    <!-- 动态背景装饰 -->
    <div class="floating-hearts">
      <div class="heart" v-for="i in 15" :key="i" :style="{ animationDelay: (i * 0.5) + 's' }">♡</div>
    </div>
    
    <!-- 页面头部 -->
    <header class="page-header">
      <!-- 轮播图 -->
      <div class="header-carousel" v-if="carouselImages.length > 0">
        <div class="carousel-container">
          <div v-for="(image, index) in carouselImages" :key="index" class="carousel-slide" :class="{ active: currentImageIndex === index }">
            <img :src="image" :alt="`carousel-${index}`" />
          </div>
        </div>
      </div>
      
      <div class="header-decoration">
        <div class="sparkle">✨</div>
        <div class="sparkle">⭐</div>
        <div class="sparkle">💫</div>
      </div>
      <div class="container">
        <div class="btn-container">
          <button class="btn home-btn" @click="$router.push('/')">
            <span class="btn-icon">🏠</span>
            返回主页
          </button>
        </div>
        <h1 class="page-title">
          <span class="title-decoration">🌸</span>
          {{ pageTitle }}
          <span class="title-decoration">🌸</span>
        </h1>
        <div class="subtitle">管理您的配置组，让一切井井有条 ✨</div>
      </div>
    </header>

    <div class="container">
      <!-- 加载状态 -->
      <div v-if="loading" class="loading-container">
        <div class="loading-animation">
          <div class="loading-heart">💖</div>
          <div class="loading-dots">
            <span></span>
            <span></span>
            <span></span>
          </div>
        </div>
        <p class="loading-text">正在加载配置组...</p>
      </div>

      <!-- 配置组列表 -->
      <div v-else-if="groups.length > 0" class="groups-container">
        <div class="groups-header">
          <h2>
            <span class="header-icon">📋</span>
            配置组列表
            <span class="groups-count">({{ groups.length }})</span>
          </h2>
          <div class="header-actions">
            <button class="btn ghost" @click="selectedGroups.length === groups.length ? clearSelection() : selectAll()">
              <span class="btn-icon">✅</span>
              {{ selectedGroups.length === groups.length && groups.length > 0 ? '取消全选' : '全选' }}
            </button>
            <button class="btn primary" @click="startSelected" :disabled="isStarting || selectedGroups.length === 0">
              <span class="btn-icon">{{ isStarting ? '⏳' : '🚀' }}</span>
              启动所选 ({{ selectedGroups.length }})
            </button>
          </div>
        </div>
        
        <div class="groups-grid">
          <div 
            v-for="(group, index) in groups" 
            :key="group" 
            class="group-card"
            :class="{ selected: isSelected(group) }"
            :style="{ animationDelay: (index * 0.1) + 's' }"
          >
            <label class="select-checkbox">
              <input type="checkbox" :checked="isSelected(group)" @change="toggleSelect(group)" />
              <span class="checkbox-ui"></span>
            </label>
            <div class="card-header">
              <div class="group-icon">⚙️</div>
              <h3 class="group-name">{{ group }}</h3>
            </div>
  
            <div class="card-actions">
              <button 
                class="btn start-btn" 
                @click="startGroup(group)"
                :disabled="isStarting"
              >
                <span class="btn-icon">{{ isStarting ? '⏳' : '🚀' }}</span>
                {{ isStarting ? '启动中...' : '启动' }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-else class="empty-state">
        <div class="empty-icon">📭</div>
        <h3>暂无配置组</h3>
        <p>还没有任何配置组，点击下方按钮重新加载试试吧！</p>
        <button class="btn reload-btn" @click="loadGroups">
          <span class="btn-icon">🔄</span>
          重新加载
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
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

// 获取轮播图图片
const getImages = async () => {
  try {
    const response = await fetch('/api/images')
    if (!response.ok) {
      throw new Error('Failed to fetch images')
    }
    const data = await response.json()
    console.log('轮播图数据:', data)
    carouselImages.value = data.images || []
    
    // 启动轮播
    if (carouselImages.value.length > 0) {
      console.log('轮播图数量:', carouselImages.value.length)
      startCarousel()
    }
  } catch (error) {
    console.error('获取轮播图失败:', error)
    // 如果API失败，使用默认图片
    carouselImages.value = ['/img/bd.jpg', '/img/ff.png', '/img/ng.jpg', '/img/sh.jpg']
    startCarousel()
  }
}

// 启动轮播
const startCarousel = () => {
  console.log('启动轮播，图片数量:', carouselImages.value.length)
  if (carouselImages.value.length > 1) {
    carouselInterval = setInterval(() => {
      currentImageIndex.value = (currentImageIndex.value + 1) % carouselImages.value.length
      console.log('切换到图片:', currentImageIndex.value)
    }, 8000) // 每8秒切换一张图片
  }
}

// 加载配置组数据
const loadGroups = async () => {
  loading.value = true
  try {
    const response = await apiMethods.getListGroups()
    // 根据后端返回的数据结构调整
    if (response && response.items) {
      groups.value = response.items
      pageTitle.value = response.title || '配置组列表'
    } else if (Array.isArray(response)) {
      groups.value = response
    } else {
      groups.value = []
    }
  } catch (error) {
    console.error('获取配置组失败:', error)
    message.error('获取配置组失败')
    groups.value = []
  } finally {
    loading.value = false
  }
}

// 启动配置组
const startGroup = async (groupName) => {
  if (isStarting.value) return
  
  isStarting.value = true
  try {
    await apiMethods.startGroups([groupName])
    message.success('启动成功！')
  } catch (error) {
    console.error('启动失败:', error)
    message.error('启动失败！')
  } finally {
    isStarting.value = false
  }
}

// 多选相关（保留原有交互，轻度美化）
const isSelected = (groupName) => selectedGroups.value.includes(groupName)
const toggleSelect = (groupName) => {
  if (isSelected(groupName)) {
    selectedGroups.value = selectedGroups.value.filter(name => name !== groupName)
  } else {
    selectedGroups.value = [...selectedGroups.value, groupName]
  }
}
const selectAll = () => {
  selectedGroups.value = Array.isArray(groups.value) ? [...new Set(groups.value)] : []
}
const clearSelection = () => {
  selectedGroups.value = []
}
const startSelected = async () => {
  if (isStarting.value) return
  if (selectedGroups.value.length === 0) {
    message.warning('请先选择配置组')
    return
  }
  isStarting.value = true
  try {
    await apiMethods.startGroups(selectedGroups.value)
    message.success('启动成功！')
  } catch (error) {
    console.error('启动失败:', error)
    message.error('启动失败！')
  } finally {
    isStarting.value = false
  }
}

// 生命周期
onMounted(() => {
  loadGroups()
  getImages()
})

// 清理定时器
onUnmounted(() => {
  if (carouselInterval) {
    clearInterval(carouselInterval)
  }
})
</script>

<style scoped>
:root {
  --primary-color: #ff6eb4;
  --secondary-color: #ff8cc8;
  --accent-color: #ffb3d9;
  --background-light: #fff6fb;
  --background-gradient: linear-gradient(135deg, #fff6fb 0%, #ffe6f2 50%, #ffd6eb 100%);
  --text-color: #ff6eb4; /* 强调色 */
  --text-dark: #e91e63; /* 深粉强调 */
  --text-base: #333333; /* 主体文字 */
  --text-muted: #666666; /* 次要文字 */
  --border-color: #ffc0da;
  --hover-color: rgba(255, 192, 218, 0.3);
  --card-shadow: 0 8px 32px rgba(255, 110, 180, 0.15);
  --glow-color: rgba(255, 110, 180, 0.4);
}

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

.list-groups-page {
  font-family: "Comic Sans MS", "Segoe UI", sans-serif;
  background: var(--background-gradient);
  color: var(--text-base);
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
}

/* 动态背景装饰 */
.floating-hearts {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 1;
}

.heart {
  position: absolute;
  font-size: 20px;
  color: var(--accent-color);
  opacity: 0.6;
  animation: float 6s infinite ease-in-out;
}

.heart:nth-child(odd) {
  left: 10%;
  animation-duration: 8s;
}

.heart:nth-child(even) {
  right: 10%;
  animation-duration: 7s;
}

.heart:nth-child(3n) {
  left: 50%;
  animation-duration: 9s;
}

@keyframes float {
  0%, 100% {
    transform: translateY(100vh) rotate(0deg);
    opacity: 0;
  }
  10% {
    opacity: 0.6;
  }
  90% {
    opacity: 0.6;
  }
  50% {
    transform: translateY(-20px) rotate(180deg);
    opacity: 1;
  }
}

/* 页面头部 */
.page-header {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95) 0%, rgba(255, 246, 251, 0.9) 100%);
  padding: 40px 0 30px;
  text-align: center;
  box-shadow: 0 8px 32px var(--glow-color);
  border-radius: 0 0 40px 40px;
  position: relative;
  z-index: 10;
  backdrop-filter: blur(10px);
  min-height: 300px; /* 设置最小高度 */
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
  object-fit: cover; /* 可选值: cover, contain, fill, scale-down */
  object-position: center; /* 图片位置控制: center, top, bottom, left, right, 或具体坐标如 50% 25% */
  border-radius: 0 0 40px 40px;
}

/* 图片位置预设类 */
.carousel-slide img.position-top {
  object-position: top;
}

.carousel-slide img.position-bottom {
  object-position: bottom;
}

.carousel-slide img.position-left {
  object-position: left;
}

.carousel-slide img.position-right {
  object-position: right;
}

.carousel-slide img.position-center-top {
  object-position: center top;
}

.carousel-slide img.position-center-bottom {
  object-position: center bottom;
}

.carousel-slide img.position-left-top {
  object-position: left top;
}

.carousel-slide img.position-right-top {
  object-position: right top;
}

.carousel-slide img.position-left-bottom {
  object-position: left bottom;
}

.carousel-slide img.position-right-bottom {
  object-position: right bottom;
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

.page-header .container {
  position: relative;
  z-index: 1;
}

.header-decoration {
  position: absolute;
  top: 10px;
  left: 0;
  right: 0;
  display: flex;
  justify-content: space-around;
  align-items: center;
  padding: 0 20px;
}

.sparkle {
  font-size: 24px;
  animation: sparkle 2s infinite ease-in-out;
}

.sparkle:nth-child(2) {
  animation-delay: 0.5s;
}

.sparkle:nth-child(3) {
  animation-delay: 1s;
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

.page-title {
  color: var(--primary-color);
  font-size: 2.5rem;
  text-shadow: 0 0 20px var(--glow-color);
  margin: 20px 0 10px;
  animation: titleGlow 3s infinite ease-in-out;
}

@keyframes titleGlow {
  0%, 100% {
    text-shadow: 0 0 20px var(--glow-color);
  }
  50% {
    text-shadow: 0 0 30px var(--glow-color), 0 0 40px var(--primary-color);
  }
}

.title-decoration {
  display: inline-block;
  animation: bounce 2s infinite;
  margin: 0 15px;
}

.title-decoration:nth-child(3) {
  animation-delay: 0.5s;
}

@keyframes bounce {
  0%, 20%, 50%, 80%, 100% {
    transform: translateY(0);
  }
  40% {
    transform: translateY(-10px);
  }
  60% {
    transform: translateY(-5px);
  }
}

.subtitle {
  font-size: 1.1rem;
  color: var(--text-muted);
  margin-top: 10px;
  opacity: 0.9;
}

.container {
  max-width: 1200px;
  margin: 40px auto;
  padding: 20px;
  position: relative;
  z-index: 5;
}

.btn-container {
  margin: 20px auto;
  text-align: center;
}

.btn {
  background: linear-gradient(135deg, #fff 0%, #fff6fb 100%);
  color: var(--primary-color);
  border: 2px solid var(--primary-color);
  border-radius: 50px;
  padding: 12px 24px;
  font-size: 1rem;
  cursor: pointer;
  box-shadow: var(--card-shadow);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  margin: 8px;
  font-weight: bold;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  backdrop-filter: blur(5px);
  outline: none;
  -webkit-tap-highlight-color: transparent;
  -webkit-appearance: none;
  appearance: none;
}

.btn.primary {
  background: linear-gradient(135deg, #ff2f9d 0%, #ff7cc8 100%);
  color: #ffffff;
  border: 1px solid #c81b74;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.18);
  border-radius: 14px;
  padding: 10px 18px;
  box-shadow: 0 10px 24px rgba(255, 110, 180, 0.32);
}
.btn.primary:hover {
  background: linear-gradient(135deg, #ff1890 0%, #ff6eb4 100%);
  color: #ffffff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.25);
  box-shadow: 0 14px 28px rgba(255, 110, 180, 0.38);
}
.btn.primary:focus { outline: none; box-shadow: 0 0 0 3px rgba(255,110,180,0.28), 0 12px 26px rgba(255,110,180,0.36); }
.btn:focus { outline: none; }
.btn:focus-visible { outline: none; }
.btn::-moz-focus-inner { border: 0; }
.btn.primary:disabled {
  opacity: 1;
  background: linear-gradient(135deg, #e97bad 0%, #f8c2da 100%);
  color: rgba(255, 255, 255, 0.92);
  cursor: not-allowed;
  border: 1px solid rgba(0,0,0,0.04);
  filter: none;
}
.btn.ghost {
  background: rgba(255, 255, 255, 0.85);
  color: var(--text-base);
  border-color: rgba(0,0,0,0.12);
  box-shadow: 0 6px 18px rgba(255, 110, 180, 0.18);
}
.btn.ghost:hover {
  background: rgba(255, 255, 255, 0.95);
  color: var(--primary-color);
  box-shadow: 0 12px 40px var(--glow-color);
}
.btn.ghost:disabled {
  opacity: 1;
  color: var(--text-muted);
  border-color: #e9b8cf;
}

.btn-icon {
  font-size: 1.2em;
  display: inline-block;
  transition: transform 0.3s ease;
}

.btn:hover {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--secondary-color) 100%);
  color: #ec38b9;
  box-shadow: 0 12px 40px var(--glow-color);
  transform: translateY(-3px) scale(1.05);
}

.btn:hover .btn-icon {
  transform: scale(1.2) rotate(10deg);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

/* 加载状态 */
.loading-container {
  text-align: center;
  padding: 80px 20px;
  background: rgba(255, 255, 255, 0.8);
  border-radius: 30px;
  box-shadow: var(--card-shadow);
  backdrop-filter: blur(10px);
}

.loading-animation {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.loading-heart {
  font-size: 60px;
  animation: heartbeat 1.5s infinite ease-in-out;
}

@keyframes heartbeat {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.3);
  }
}

.loading-dots {
  display: flex;
  gap: 8px;
}

.loading-dots span {
  width: 12px;
  height: 12px;
  background: var(--primary-color);
  border-radius: 50%;
  animation: loadingDots 1.4s infinite ease-in-out both;
}

.loading-dots span:nth-child(1) { animation-delay: -0.32s; }
.loading-dots span:nth-child(2) { animation-delay: -0.16s; }

@keyframes loadingDots {
  0%, 80%, 100% {
    transform: scale(0);
  }
  40% {
    transform: scale(1);
  }
}

.loading-text {
  margin-top: 20px;
  font-size: 1.3rem;
  color: var(--primary-color);
  font-weight: bold;
}

/* 配置组容器 */
.groups-container {
  animation: fadeInUp 0.6s ease-out;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.groups-header {
  text-align: center;
  margin-bottom: 30px;
}

.groups-header h2 {
  font-size: 1.8rem;
  color: var(--text-base);
  display: inline-flex;
  align-items: center;
  gap: 10px;
  background: rgba(255, 255, 255, 0.35);
  padding: 8px 16px;
  border-radius: 16px;
  box-shadow: 0 8px 22px rgba(255, 110, 180, 0.22);
  border: 1px solid rgba(0,0,0,0.06);
}

.header-actions {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
  margin-top: 12px;
}

.search-input {
  padding: 10px 14px;
  border: 2px solid var(--border-color);
  border-radius: 30px;
  outline: none;
  min-width: 220px;
  background: #fff;
  color: var(--text-base);
  box-shadow: var(--card-shadow);
}
.search-input:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 4px var(--hover-color);
}

.header-icon {
  font-size: 1.5em;
  animation: wiggle 3s infinite ease-in-out;
}

@keyframes wiggle {
  0%, 100% { transform: rotate(0deg); }
  25% { transform: rotate(-5deg); }
  75% { transform: rotate(5deg); }
}

.groups-count {
  background: linear-gradient(135deg, #ff379f 0%, #ff8bcf 100%);
  color: #ffffff;
  padding: 3px 12px;
  border-radius: 999px;
  font-size: 0.9em;
  margin-left: 10px;
  border: 1px solid rgba(0,0,0,0.12);
  box-shadow: 0 10px 24px rgba(255, 110, 180, 0.35);
}

/* 配置组网格 */
.groups-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 25px;
  padding: 20px 0;
}

.group-card {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95) 0%, rgba(255, 246, 251, 0.9) 100%);
  border-radius: 25px;
  padding: 25px;
  box-shadow: var(--card-shadow);
  border: 1px solid var(--border-color);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  backdrop-filter: blur(10px);
  animation: cardSlideIn 0.6s ease-out both;
  position: relative;
  overflow: hidden;
}

.select-checkbox {
  position: absolute;
  top: 14px;
  right: 14px;
  width: 26px;
  height: 26px;
  z-index: 2;
}
.select-checkbox input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}
.select-checkbox .checkbox-ui {
  display: inline-block;
  width: 26px;
  height: 26px;
  border-radius: 50%;
  border: 3px solid #d14e8f;
  background: #ffffff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  position: relative;
  transition: all 0.18s ease-in-out;
}
.select-checkbox .checkbox-ui:hover {
  transform: scale(1.04);
}
.select-checkbox input:checked + .checkbox-ui {
  background: linear-gradient(135deg, #ff2f9d 0%, #ff7cc8 100%);
  border-color: #b5166b;
  box-shadow: 0 0 0 2px rgba(255,110,180,0.28), 0 8px 20px rgba(255, 110, 180, 0.35);
  transform: scale(1.02);
}
.select-checkbox .checkbox-ui::after {
  content: '✓';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -55%) scale(0.6);
  color: #ffffff;
  font-weight: 900;
  font-size: 20px;
  line-height: 1;
  opacity: 0;
  transition: opacity 0.15s ease, transform 0.15s ease;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.28), 0 0 1px rgba(0, 0, 0, 0.35);
}
.select-checkbox input:checked + .checkbox-ui::after {
  opacity: 1;
  transform: translate(-50%, -55%) scale(1);
}
/* 移除内圈白色高光环 */

.group-card::before {
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

.group-card:hover::before {
  transform: scaleX(1);
}

.group-card:hover {
  transform: translateY(-8px) scale(1.02);
  box-shadow: 0 20px 50px var(--glow-color);
}

.group-card.selected {
  border: 2px solid var(--primary-color);
  box-shadow: 0 0 0 2px rgba(255,110,180,0.25), 0 16px 36px rgba(255,110,180,0.28);
  background: linear-gradient(135deg, #fff0f6 0%, #ffe4f0 100%);
}
.group-card.selected::after {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 6px;
  background: linear-gradient(180deg, #ff2f9d, #ff8bcf);
  border-top-left-radius: 24px;
  border-bottom-left-radius: 24px;
}

@keyframes cardSlideIn {
  from {
    opacity: 0;
    transform: translateX(-50px) scale(0.9);
  }
  to {
    opacity: 1;
    transform: translateX(0) scale(1);
  }
}

.card-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 15px;
}

.group-icon {
  font-size: 2.5rem;
  animation: rotate 4s infinite linear;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.group-name {
  color: var(--text-base);
  font-size: 1.4rem;
  font-weight: bold;
  margin: 0;
}

.card-content {
  margin-bottom: 20px;
}

.group-description {
  color: var(--text-muted);
  margin-bottom: 10px;
  font-size: 1rem;
}

.group-status {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  width: 10px;
  height: 10px;
  background: #4caf50;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(76, 175, 80, 0.7);
  }
  70% {
    box-shadow: 0 0 0 10px rgba(76, 175, 80, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(76, 175, 80, 0);
  }
}

.status-text {
  color: #4caf50;
  font-weight: bold;
  font-size: 0.9rem;
}

.card-actions {
  text-align: center;
}

.start-btn {
  width: 100%;
  margin: 0;
  padding: 12px 20px;
  font-size: 1rem;
  font-weight: bold;
  border: 1px solid #ffb3d9;
  border-radius: 12px;
  background: linear-gradient(135deg, #fff 0%, #fff6fb 100%);
  color: var(--primary-color);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition: all 0.2s ease-in-out;
  box-shadow: 0 8px 20px rgba(255, 110, 180, 0.18);
}
.start-btn:hover {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--secondary-color) 100%);
  color: #fff;
  border-color: transparent;
  box-shadow: 0 12px 28px rgba(255, 110, 180, 0.35);
}
.start-btn:disabled {
  background: linear-gradient(135deg, #ffe6f4 0%, #ffeef7 100%);
  color: #cc6a9e;
  border-color: #ffcee5;
  box-shadow: none;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 80px 20px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95) 0%, rgba(255, 246, 251, 0.9) 100%);
  border-radius: 30px;
  box-shadow: var(--card-shadow);
  backdrop-filter: blur(10px);
  animation: fadeInUp 0.6s ease-out;
}

.empty-icon {
  font-size: 80px;
  margin-bottom: 20px;
  animation: sway 3s infinite ease-in-out;
}

@keyframes sway {
  0%, 100% {
    transform: rotate(-5deg);
  }
  50% {
    transform: rotate(5deg);
  }
}

.empty-state h3 {
  font-size: 1.8rem;
  color: var(--text-dark);
  margin-bottom: 15px;
}

.empty-state p {
  font-size: 1.1rem;
  margin-bottom: 30px;
  color: var(--text-muted);
  line-height: 1.6;
}

.reload-btn {
  font-size: 1.1rem;
  padding: 15px 30px;
}

/* 浮动工具条 */
.floating-toolbar {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  padding: 10px 16px;
  display: flex;
  align-items: center;
  gap: 14px;
  box-shadow: var(--card-shadow);
  backdrop-filter: blur(8px);
  z-index: 50;
}
.floating-toolbar .toolbar-info {
  color: var(--text-dark);
  font-weight: bold;
}
.floating-toolbar .toolbar-actions {
  display: flex;
  gap: 10px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .page-title {
    font-size: 2rem;
  }
  
  .title-decoration {
    margin: 0 8px;
  }

  .groups-grid {
    grid-template-columns: 1fr;
    gap: 20px;
  }

  .container {
    margin: 20px auto;
    padding: 15px;
  }

  .btn {
    font-size: 0.9rem;
    padding: 10px 20px;
  }

  .group-card {
    padding: 20px;
  }
  
  /* 移动端轮播图适配 */
  .page-header {
    min-height: 250px; /* 移动端减小高度 */
  }
  
  .header-carousel {
    border-radius: 0 0 20px 20px;
  }
  
  .carousel-slide img {
    border-radius: 0 0 20px 20px;
    object-position: center top; /* 移动端使用顶部居中 */
  }
  
  .page-header::before {
    border-radius: 0 0 20px 20px;
  }
}

@media (max-width: 480px) {
  .page-title {
    font-size: 1.8rem;
  }

  .subtitle {
    font-size: 1rem;
  }

  .floating-hearts {
    display: none;
  }

  .header-decoration {
    display: none;
  }
  
  /* 小屏幕轮播图适配 */
  .header-carousel {
    border-radius: 0 0 15px 15px;
  }
  
  .carousel-slide img {
    border-radius: 0 0 15px 15px;
    object-position: center; /* 小屏幕使用居中 */
  }
  
  .page-header::before {
    border-radius: 0 0 15px 15px;
  }
}
</style>
