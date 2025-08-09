<template>
  <div class="container">
    <div class="title-row">
      <h1 class="main-title">{{ title }}</h1>
      <span v-if="jsVersion" class="title-badge">[{{ jsVersion }}]</span>
    </div>

    <div class="button-group">
      <button @click="goToHome" class="btn">返回主页</button>
    </div>

    <section class="panel" id="executionPanel">
<!--      <h2>{{ title }}</h2>-->

      <div v-if="loading" class="loading-container">
        <div class="loading-spinner"></div>
        <p>加载中...</p>
      </div>

      <div v-else-if="items.length === 0" class="no-data">
        <p>暂无数据</p>
      </div>

      <div v-for="(item, index) in items" :key="index" class="info-box">
        <div v-if="images.length > 0" class="swiper corner-icon-swiper" :ref="`swiper-${index}`">
          <div class="swiper-wrapper">
            <div v-for="(image, imgIndex) in images" :key="imgIndex" class="swiper-slide">
              <img :src="`/static/image/${image}`" :alt="`image-${imgIndex}`" />
            </div>
          </div>
        </div>
        <div class="file-header">
          <h3 class="file-title">
            <span class="file-icon">📄</span>
            {{ item.FileName }}
          </h3>
          <button class="btn chart-btn"  @click="getAutoArtifactsPro2Btn(item.FileName)">
            📊 转为折线图
          </button>
        </div>
        <h4 style="margin-bottom: 10px">
          {{ item.Mark }}
        </h4>

        <div class="file-details">
          <div class="detail-item" v-for="(detail, detailIndex) in item.Detail" :key="detailIndex">
            {{ detail }}
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script>
import { ref, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { Swiper } from 'swiper'
import 'swiper/css'
import 'swiper/css/effect-fade'
import { apiMethods } from '@/utils/api'

export default {
  name: 'AutoArtifactsPro',
  setup() {
    const router = useRouter()
    
    // 响应式数据
    const title = ref('AutoArtifactsPro')
    const jsVersion = ref('')
    const items = ref([])
    const images = ref([])
    const loading = ref(false)

    // 方法
    const goToHome = () => {
      router.push('/')
    }

    const getAutoArtifactsPro2Btn = (fileName) => {
      router.push({ path: '/getAutoArtifactsPro2', query: { fileName } })
    }

    const loadImages = async () => {
      try {
        const data = await apiMethods.getImages()
        images.value = data.data || []
      } catch (err) {
        console.error('Failed to load images:', err)
        // 如果API不可用，使用默认图片列表
        images.value = ['bd.jpg', 'ff.png', 'ng.jpg', 'sh.jpg']
      }
    }

    const initializeSwipers = async () => {
      await nextTick()
      
      const swiperContainers = document.querySelectorAll('.corner-icon-swiper')
      
      swiperContainers.forEach(swiperContainer => {
        if (images.value.length > 0) {
          new Swiper(swiperContainer, {
            loop: true,
            autoplay: {
              delay: 5000,
              disableOnInteraction: false
            },
            effect: 'fade',
            fadeEffect: {
              crossFade: true
            }
          })
        }
      })
    }

    const loadData = async () => {
      try {
        loading.value = true
        const data = await apiMethods.getAutoArtifactsPro()
        
        // 根据后端返回的数据结构调整
        if (data) {
          title.value = data.title || 'AutoArtifactsPro'
          jsVersion.value = data.JsVersion || ''
          items.value = data.items || []
        }
      } catch (err) {
        console.error('Failed to load AutoArtifactsPro data:', err)
        // 使用默认数据作为fallback
        items.value = [
          {
            FileName: 'example1.txt',
            Detail: ['暂无数据，请检查后端服务'],
            Mark: '暂无数据'
          }
        ]
      } finally {
        loading.value = false
      }
    }

    // 生命周期
    onMounted(async () => {
      await loadImages()
      await loadData()
      await initializeSwipers()
    })

    return {
      title,
      jsVersion,
      items,
      images,
      loading,
      goToHome,
      getAutoArtifactsPro2Btn
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

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

.button-group{
  display: flex;
  justify-content: flex-end;

}

.container {
  font-family: "Comic Sans MS", "Segoe UI", sans-serif;
  background-color: var(--background-light);
  color: var(--text-color);
  min-height: 100vh;
  background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100"><circle cx="20" cy="20" r="5" fill="%23ffd6eb" opacity="0.6"/><circle cx="70" cy="70" r="7" fill="%23ffc0da" opacity="0.5"/></svg>');
  padding-bottom: 50px;
  max-width: 900px;
  margin: 30px auto;
  padding: 0 20px;
  position: relative;
}

h1 {
  color: var(--primary-color);
  font-size: 2rem;
  text-shadow: 0 0 10px var(--primary-color);
  margin-top: 15px;
}

.btn {
  background-color: #f8eaea;
  color: var(--primary-color);
  border: 3px solid var(--primary-color);
  border-radius: 50px;
  padding: 10px 20px;
  font-size: 1rem;
  cursor: pointer;
  box-shadow: 0 0 10px var(--primary-color);
  transition: all 0.3s ease;
  margin: 0 10px 10px 0;
  font-weight: bold;
}

.btn:hover {
  background-color: var(--primary-color);
  color: #a22b2b;
  box-shadow: 0 0 20px var(--primary-color);
  transform: scale(1.05);
}

section.panel {
  border-radius: 20px;
  margin-bottom: 30px;
  position: relative;
}

section.panel h2 {
  color: var(--primary-color);
  font-size: 1.6rem;
  margin-bottom: 15px;
  border-bottom: 2px solid var(--primary-color);
  padding-bottom: 5px;
}

table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0 8px;
  border-radius: 20px;
  overflow: hidden;
}

th, td {
  padding: 15px 20px;
  text-align: left;
  color: var(--text-color);
}

th {
  background-color: #ffcce6;
  color: #ff5599;
  font-weight: bold;
  border-radius: 10px;
}

td {
  background-color: rgba(255, 255, 255, 0.5);
  border-radius: 10px;
}

tr:hover td {
  background-color: var(--hover-color);
}

/* 可爱元素装饰 */
.container::before {
  content: "♡";
  position: absolute;
  top: 10px;
  left: 10px;
  font-size: 30px;
  color: var(--primary-color);
  opacity: 0.5;
}

.container::after {
  content: "♡";
  position: absolute;
  bottom: 10px;
  right: 10px;
  font-size: 30px;
  color: var(--primary-color);
  opacity: 0.5;
}

@media (max-width: 600px) {
  h1 {
    font-size: 1.5rem;
  }
  h2 {
    font-size: 1.3rem;
  }
  .btn {
    font-size: 0.9rem;
    padding: 8px 16px;
  }
  th, td {
    font-size: 0.9rem;
    padding: 12px;
  }
  img {
    opacity: 0.9;
  }
}

.info-box {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(255, 246, 251, 0.95));
  border: 3px solid var(--primary-color);
  border-radius: 25px;
  padding: 25px;
  margin-bottom: 30px;
  box-shadow: 
    0 8px 25px rgba(255, 110, 180, 0.2),
    0 4px 10px rgba(255, 110, 180, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
  transition: all 0.3s ease;
  position: relative;
  backdrop-filter: blur(10px);
}

.info-box:hover {
  transform: translateY(-6px);
  box-shadow: 
    0 15px 35px rgba(255, 110, 180, 0.3),
    0 8px 15px rgba(255, 110, 180, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.9);
  border-color: #ff5599;
}

.info-box::before {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(45deg, var(--primary-color), #ff5599, var(--primary-color));
  border-radius: 25px;
  z-index: -1;
  opacity: 0.3;
}

/* 文件头部区域 */
.file-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 2px solid rgba(255, 110, 180, 0.2);
  flex-wrap: wrap;
  gap: 10px;
}

.file-title {
  font-size: 1.4rem;
  color: var(--primary-color);
  margin: 0;
  text-shadow: 0 0 8px var(--primary-color);
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: bold;
  flex: 1;
}

.file-icon {
  font-size: 1.2rem;
  filter: drop-shadow(0 0 3px var(--primary-color));
}

.chart-btn {
  background: linear-gradient(135deg, var(--primary-color), #ff5599);
  color: #ff6eb4;
  border: none;
  box-shadow: 0 4px 15px rgba(255, 110, 180, 0.3);
  font-weight: bold;
  white-space: nowrap;
}

.chart-btn:hover {
  background: linear-gradient(135deg, #ff5599, #ff4488);
  transform: scale(1.05) translateY(-2px);
  box-shadow: 0 6px 20px rgba(255, 110, 180, 0.4);
}

/* 文件详情区域 */
.file-details {
  background: rgb(244, 218, 228);
  border-radius: 15px;
  padding: 15px;
  border: 1px solid rgba(78, 6, 36, 0.5);
}

.detail-item {
  background: rgba(255, 255, 255, 0.8);
  box-shadow: 0 3px 10px rgba(235, 4, 115, 0.2);
  margin: 8px 0;
  padding: 12px 15px;
  border-radius: 12px;
  border-left: 4px solid var(--primary-color);
  font-size: 1rem;
  color: #ff6699;
  font-weight: bold;
  text-shadow: 0 0 3px rgba(255, 209, 224, 0.8);
  transition: all 0.2s ease;
  word-break: break-word;
}

.detail-item:hover {
  background: rgba(255, 255, 255, 0.95);
  transform: translateX(5px);
  box-shadow: 0 3px 10px rgba(255, 110, 180, 0.2);
}

.info-box .highlight {
  color: #ff5599;
  font-weight: bold;
}

/* 响应式调整 */
@media (max-width: 600px) {
  .file-header {
    flex-direction: column;
    align-items: stretch;
    text-align: center;
  }
  
  .file-title {
    justify-content: center;
    margin-bottom: 10px;
  }
}

.corner-icon-swiper {
  position: absolute;
  bottom: 2px;
  right: 5px;
  width: 220px;
  height: 400px;
  z-index: -1;
  pointer-events: none;
}

.corner-icon-swiper .swiper-slide img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 10px;
}

p {
  font-size: 1em;
  margin-bottom: 10px;
  text-shadow: 0 0 5px #ffd1e0;
  word-break: break-word;
  color: #ff6699;
  font-weight: bold;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  justify-content: flex-start;
  flex-wrap: wrap;
  margin-bottom: 20px;
}

.main-title {
  font-size: 2rem;
  font-weight: bold;
  color: var(--primary-color);
  text-shadow: 0 0 10px var(--primary-color);
  margin: 0;
}

.title-badge {
  font-size: 0.75rem;
  background-color: #ffeaf5;
  border: 1px solid #ffb6d9;
  color: #ff5599;
  border-radius: 20px;
  padding: 2px 8px;
  font-weight: bold;
  box-shadow: 0 0 5px rgba(255, 105, 180, 0.3);
  white-space: nowrap;
}

/* 加载状态样式 */
.loading-container {
  text-align: center;
  padding: 40px 0;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-color);
  border-top: 3px solid var(--primary-color);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 20px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.loading-container p {
  color: var(--primary-color);
  font-size: 1.1rem;
  font-weight: bold;
}

/* 无数据状态样式 */
.no-data {
  text-align: center;
  padding: 40px 0;
}

.no-data p {
  color: var(--primary-color);
  font-size: 1.2rem;
  font-weight: bold;
  opacity: 0.7;
}
</style>
