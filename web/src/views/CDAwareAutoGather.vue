<template>
  <div class="cd-aware-container">
    <!-- 樱花动画背景 -->
    <canvas ref="animeCanvas" class="anime-canvas"></canvas>
    
    <!-- 页面标题 -->
    <div class="page-header">
      <h1>🔄 CD管理自动采集</h1>
      <div class="header-actions">
        <a-select
          v-model:value="filterStatus"
          @change="handleFilterChange"
          class="filter-select"
          :disabled="loading"
          placeholder="选择筛选条件"
        >
          <a-select-option value="3">显示全部</a-select-option>
          <a-select-option value="1">仅显示可采集</a-select-option>
          <a-select-option value="2">仅显示未到时间</a-select-option>
        </a-select>
        <button @click="refreshData" class="refresh-btn" :disabled="loading">
          {{ loading ? '加载中...' : '刷新数据' }}
        </button>
      </div>
      <button @click="CDAllMaterial" class="refresh-btn">是否加入背包统计</button>
      <button @click="UpdateAllCD" class="refresh-btn">一键更新全部材料</button>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <a-spin size="large" />
      <p>正在加载数据...</p>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="error-container">
      <a-alert
        :message="error"
        type="error"
        show-icon
        closable
        @close="error = ''"
      />
    </div>

    <!-- 数据展示 -->
    <div v-else-if="data.length > 0" class="data-container">
      <div v-for="(account, accountIndex) in data" :key="accountIndex" class="account-section">
        <div class="account-header">
          <h2>👤 {{ account.UID }}---({{ getFilteredGathers(account.CDAwareAutoGather).length }})</h2>
        </div>
        
        <div class="gather-list">
          <div 
            v-for="(gather, gatherIndex) in getFilteredGathers(account.CDAwareAutoGather)" 
            :key="gatherIndex" 
            class="gather-item"
          >
            <div 
              class="gather-header" 
              @click="toggleGather(accountIndex, getOriginalGatherIndex(account.CDAwareAutoGather, gather))"
            >
              <div class="gather-title">
                <h3>📄 {{ gather.TextName }}</h3>
                <span class="file-count">{{ gather.Detail.length }} 个文件</span>
              </div>
              <div class="expand-icon" :class="{ 'expanded': gather.expanded }">
                ▼
              </div>
            </div>
            
            <a-collapse 
              :activeKey="gather.expanded ? ['details'] : []"
              :bordered="false"
              class="detail-collapse"
            >
              <a-collapse-panel key="details" :showArrow="false">
                <div class="detail-list">
                  <div 
                    v-for="(detail, detailIndex) in gather.Detail" 
                    :key="detailIndex"
                    class="detail-item"
                    :class="{ 'expired': detail.CDExpired }"
                  >
                    <div class="file-info">
                      <div class="file-name">{{ detail.FileName }}</div>
                      <div class="cd-time">
                        <span class="time-label">CD时间:</span>
                        <span class="time-value">{{ detail.CDTime }}</span>
                      </div>
                    </div>
                    <div class="status-badge" :class="{ 'expired': detail.CDExpired }">
                      {{ detail.CDExpired ? '可采集' : '未到时间' }}
                    </div>
                  </div>
                </div>
              </a-collapse-panel>
            </a-collapse>
          </div>
        </div>
      </div>
    </div>

    <!-- 空数据状态 -->
    <div v-else class="empty-container">
      <a-empty description="暂无数据" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { apiMethods } from '../utils/api.js'

// 响应式数据
const loading = ref(false)
const error = ref('')
const data = ref([])
const animeCanvas = ref(null)
const filterStatus = ref('1') // 默认显示可采集

// 动画相关
let petals = []
let animationId = null

// 樱花花瓣类
class Petal {
  constructor(canvas) {
    this.canvas = canvas
    this.x = Math.random() * canvas.width
    this.y = -10
    this.size = Math.random() * 3 + 1
    this.speedY = Math.random() * 2 + 1
    this.speedX = (Math.random() - 0.5) * 0.5
    this.rotation = Math.random() * 360
    this.rotationSpeed = (Math.random() - 0.5) * 4
    this.opacity = Math.random() * 0.7 + 0.3
    this.color = `hsl(${Math.random() * 30 + 320}, 70%, 80%)`
  }

  update() {
    this.y += this.speedY
    this.x += this.speedX
    this.rotation += this.rotationSpeed

    // 添加轻微的摆动
    this.x += Math.sin(this.y * 0.01) * 0.2

    // 如果花瓣飘出屏幕底部，重新从顶部开始
    if (this.y > this.canvas.height) {
      this.y = -10
      this.x = Math.random() * this.canvas.width
    }
  }

  draw(ctx) {
    ctx.save()
    ctx.translate(this.x, this.y)
    ctx.rotate((this.rotation * Math.PI) / 180)
    ctx.globalAlpha = this.opacity
    ctx.fillStyle = this.color
    ctx.beginPath()
    ctx.ellipse(0, 0, this.size, this.size * 0.6, 0, 0, 2 * Math.PI)
    ctx.fill()
    ctx.restore()
  }
}

// 初始化樱花动画
const initAnime = () => {
  const canvas = animeCanvas.value
  if (!canvas) return

  const ctx = canvas.getContext('2d')
  canvas.width = window.innerWidth
  canvas.height = window.innerHeight

  // 创建花瓣
  petals = []
  for (let i = 0; i < 50; i++) {
    petals.push(new Petal(canvas))
  }

  // 动画循环
  const animate = () => {
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    
    petals.forEach(petal => {
      petal.update()
      petal.draw(ctx)
    })

    animationId = requestAnimationFrame(animate)
  }

  animate()

  // 窗口尺寸变化处理
  const handleResize = () => {
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight
  }
  window.addEventListener('resize', handleResize)

  return () => {
    window.removeEventListener('resize', handleResize)
    if (animationId) {
      cancelAnimationFrame(animationId)
    }
  }
}

// 获取数据
const fetchData = async () => {
  loading.value = true
  error.value = ''
  
  try {
    const response = await apiMethods.getCDAwareAutoGather(filterStatus.value)
    // 为每个gather项添加expanded属性用于控制折叠状态
    const processedData = (response || []).map(account => ({
      ...account,
      CDAwareAutoGather: account.CDAwareAutoGather.map(gather => ({
        ...gather,
        expanded: false
      }))
    }))
    data.value = processedData
  } catch (err) {
    console.error('获取CD管理自动采集数据失败:', err)
    error.value = '获取数据失败: ' + (err.message || '未知错误')
  } finally {
    loading.value = false
  }
}

// 切换折叠状态
const toggleGather = (accountIndex, gatherIndex) => {
  data.value[accountIndex].CDAwareAutoGather[gatherIndex].expanded = 
    !data.value[accountIndex].CDAwareAutoGather[gatherIndex].expanded
}

// 处理筛选条件变化
const handleFilterChange = () => {
  fetchData()
}

// 过滤采集任务（过滤掉0个文件的）
const getFilteredGathers = (gathers) => {
  return gathers.filter(gather => gather.Detail && gather.Detail.length > 0)
}

// 获取原始采集任务索引
const getOriginalGatherIndex = (originalGathers, filteredGather) => {
  return originalGathers.findIndex(gather => gather === filteredGather)
}

// 刷新数据
const refreshData = () => {
  fetchData()
}

// 是否加入背包统计
const CDAllMaterial = async () => {
  try {
    await apiMethods.CDAllMaterial()
    fetchData()
  } catch (err) {
    console.error('CDAllMaterial失败:', err)
  }
}

const UpdateAllCD = async () => {
  try {
    await apiMethods.UpdateAllCD()
    alert('更新成功,具体请看日志:logs/。没有更新的材料，请联系abgi')
  } catch (err) {
    console.error('CDAllMaterial失败:', err)
  }
}

// 组件挂载
onMounted(() => {
  fetchData()
  // 延迟初始化动画，确保DOM已渲染
  setTimeout(() => {
    initAnime()
  }, 100)
})

// 组件卸载
onUnmounted(() => {
  if (animationId) {
    cancelAnimationFrame(animationId)
  }
})
</script>

<style scoped>
.cd-aware-container {
  min-height: 100vh;
  position: relative;
  background: linear-gradient(135deg, #ffecf5 0%, #f0f8ff 100%);
  padding: 20px;
}

.anime-canvas {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 1;
}

.page-header {
  position: relative;
  z-index: 10;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  background: rgba(255, 255, 255, 0.9);
  padding: 20px;
  border-radius: 15px;
  box-shadow: 0 8px 32px rgba(255, 182, 193, 0.3);
  backdrop-filter: blur(10px);
}

.page-header h1 {
  color: #ff6699;
  margin: 0;
  font-size: 2.5rem;
  text-shadow: 2px 2px 4px rgba(255, 182, 193, 0.5);
}

.header-actions {
  display: flex;
  gap: 15px;
  align-items: center;
}

.filter-select {
  min-width: 180px;
}

.filter-select .ant-select-selector {
  background: linear-gradient(45deg, #4ecdc4, #44a08d) !important;
  border: none !important;
  border-radius: 25px !important;
  box-shadow: 0 4px 15px rgba(78, 205, 196, 0.4) !important;
  transition: all 0.3s ease !important;
}

.filter-select:hover .ant-select-selector {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(78, 205, 196, 0.6) !important;
}

.filter-select .ant-select-selection-item {
  color: white !important;
  font-weight: bold !important;
}

.filter-select .ant-select-arrow {
  color: white !important;
}

.filter-select.ant-select-disabled .ant-select-selector {
  opacity: 0.6;
  cursor: not-allowed;
}

.refresh-btn {
  background: linear-gradient(45deg, #ff6699, #ff8fab);
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 25px;
  font-size: 1rem;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 4px 15px rgba(255, 105, 153, 0.4);
}

.refresh-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(255, 105, 153, 0.6);
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.loading-container, .error-container, .empty-container {
  position: relative;
  z-index: 10;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 15px;
  box-shadow: 0 8px 32px rgba(255, 182, 193, 0.3);
  backdrop-filter: blur(10px);
}

.data-container {
  position: relative;
  z-index: 10;
}

.account-section {
  background: rgba(255, 255, 255, 0.9);
  border-radius: 15px;
  margin-bottom: 20px;
  box-shadow: 0 8px 32px rgba(255, 182, 193, 0.3);
  backdrop-filter: blur(10px);
  overflow: hidden;
}

.account-header {
  background: linear-gradient(45deg, #ff6699, #ff8fab);
  color: white;
  padding: 20px;
}

.account-header h2 {
  margin: 0;
  font-size: 1.8rem;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
}

.gather-list {
  padding: 20px;
}

.gather-item {
  background: rgba(255, 255, 255, 0.7);
  border-radius: 12px;
  margin-bottom: 15px;
  padding: 15px;
  border: 2px solid rgba(255, 182, 193, 0.3);
  transition: all 0.3s ease;
}

.gather-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(255, 182, 193, 0.4);
}

.gather-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  cursor: pointer;
  transition: all 0.3s ease;
  border-radius: 8px;
  background: rgba(255, 182, 193, 0.1);
}

.gather-header:hover {
  background: rgba(255, 182, 193, 0.2);
  transform: translateY(-1px);
}

.gather-title {
  display: flex;
  align-items: center;
  gap: 15px;
}

.gather-title h3 {
  margin: 0;
  color: #ff6699;
  font-size: 1.3rem;
}

.file-count {
  background: linear-gradient(45deg, #ff6699, #ff8fab);
  color: white;
  padding: 5px 12px;
  border-radius: 15px;
  font-size: 0.9rem;
  font-weight: bold;
}

.expand-icon {
  font-size: 1.2rem;
  color: #ff6699;
  transition: transform 0.3s ease;
  user-select: none;
}

.expand-icon.expanded {
  transform: rotate(180deg);
}

.detail-collapse {
  background: transparent;
  border: none;
}

.detail-collapse .ant-collapse-item {
  border: none;
}

.detail-collapse .ant-collapse-content {
  background: transparent;
  border: none;
}

.detail-collapse .ant-collapse-content-box {
  padding: 0 15px 15px 15px;
}

.detail-list {
  display: grid;
  gap: 10px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: rgba(255, 255, 255, 0.8);
  border-radius: 8px;
  border: 1px solid rgba(255, 182, 193, 0.2);
  transition: all 0.3s ease;
}

.detail-item:hover {
  background: rgba(255, 182, 193, 0.1);
  transform: translateX(5px);
}

.detail-item.expired {
  background: rgba(255, 182, 193, 0.2);
  border-color: #ff6b6b;
}

.file-info {
  flex: 1;
}

.file-name {
  font-weight: bold;
  color: #333;
  margin-bottom: 5px;
  font-size: 1rem;
}

.cd-time {
  display: flex;
  align-items: center;
  gap: 8px;
}

.time-label {
  color: #666;
  font-size: 0.9rem;
}

.time-value {
  color: #ff6699;
  font-weight: bold;
  font-size: 0.9rem;
}

.status-badge {
  background: linear-gradient(45deg, #4ecdc4, #44a08d);
  color: white;
  padding: 6px 12px;
  border-radius: 15px;
  font-size: 0.8rem;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.status-badge.expired {
  background: linear-gradient(45deg, #ff6b6b, #ee5a52);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: 15px;
    text-align: center;
  }
  
  .page-header h1 {
    font-size: 2rem;
  }
  
  .header-actions {
    flex-direction: column;
    gap: 10px;
    width: 100%;
  }
  
  .filter-select, .refresh-btn {
    width: 100%;
    max-width: 200px;
  }
  
  .gather-header {
    flex-direction: column;
    gap: 10px;
    text-align: center;
  }
  
  .gather-title {
    flex-direction: column;
    gap: 10px;
  }
  
  .detail-item {
    flex-direction: column;
    gap: 10px;
    text-align: center;
  }
}
</style>