<template>
  <div class="archive-page">
    <!-- 装饰性元素 -->
    <div class="decorative-elements">
      <div class="floating-heart heart-1">💖</div>
      <div class="floating-heart heart-2">✨</div>
      <div class="floating-heart heart-3">🌸</div>
      <div class="floating-heart heart-4">🎀</div>
      <div class="floating-heart heart-5">💕</div>
    </div>

    <header class="cute-header">
      <div class="header-content">
        <button class="btn cute-btn" @click="goHome">
          <span class="btn-icon">🏠</span>
          <span class="btn-text">返回首页</span>
        </button>
        <h1 class="cute-title">
          <span class="title-icon">📚</span>
          <span class="title-text">归档记录列表</span>
          <span class="title-sparkle">✨</span>
        </h1>
         <a-button type="primary" class="deleteBtn" @click="allDelete">全部删除</a-button>
      </div>
     
    </header>

    <div class="container">
      <section class="panel cute-panel" id="archiveListPanel">
        <div class="panel-header">
          <h2 class="cute-subtitle">
            <span class="subtitle-icon">📋</span>
            <span class="subtitle-text">归档记录</span>

          </h2>

          <div class="panel-decoration">
            <div class="corner-decoration corner-tl">🌸</div>
            <div class="corner-decoration corner-tr">🌸</div>
            <div class="corner-decoration corner-bl">🌸</div>
            <div class="corner-decoration corner-br">🌸</div>
          </div>
        </div>
        
        <div id="archiveListContainer" class="table-container">
          <!-- 加载状态 -->
          <div v-if="loading" class="loading-state">
            <div class="loading-spinner">⏳</div>
            <div class="loading-text">加载中...</div>
          </div>
          
          <!-- 错误状态 -->
          <div v-else-if="error" class="error-state">
            <div class="error-icon">⚠️</div>
            <div class="error-text">{{ error }}</div>
            <button class="btn retry-btn" @click="loadArchiveList">
              <span class="retry-icon">🔄</span>
              重试
            </button>
          </div>
          
          <!-- 桌面端表格 -->
          <table v-else id="archiveTable" class="cute-table desktop-table">
            <thead>
              <tr class="table-header-row">
                <th data-key="title" @click="sortBy('title')" class="sortable-header">
                  📝 标题
                  <span class="sort-indicator" v-if="currentSort.key === 'title'">
                    {{ currentSort.asc ? '↑' : '↓' }}
                  </span>
                </th>
                <th data-key="execute_time" @click="sortBy('execute_time')" class="sortable-header">
                  ⏱️ 执行时长
                  <span class="sort-indicator" v-if="currentSort.key === 'execute_time'">
                    {{ currentSort.asc ? '↑' : '↓' }}
                  </span>
                </th>
                <th class="action-header">
                  🎮 操作
                </th>
              </tr>
            </thead>
            <tbody id="archiveTableBody">
              <tr v-if="archiveData.length === 0" class="empty-row">
                <td colspan="3" class="empty-message">
                  <div class="empty-content">
                    <span class="empty-icon">📭</span>
                    <span class="empty-text">暂无归档记录</span>
                    <span class="empty-sparkle">✨</span>
                  </div>
                </td>
              </tr>
              <tr v-for="(item, index) in sortedData" 
                  :key="item.id" 
                  :class="{ 'fade-out': item.deleting, 'table-row': true }"
                  :style="{ animationDelay: index * 0.1 + 's' }">
                <td class="title-cell">{{ item.title }}</td>
                <td class="time-cell">{{ item.execute_time }}</td>
                <td class="action-cell">
                  <button class="btn delete-btn cute-delete-btn" 
                          @click="deleteItem(item)" 
                          :disabled="item.deleting">
                    <span class="delete-icon">{{ item.deleting ? '⏳' : '🗑️' }}</span>
                    <span class="delete-text">{{ item.deleting ? '删除中...' : '删除' }}</span>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>

          <!-- 移动端卡片列表 -->
          <div class="mobile-list">
            <div v-if="archiveData.length === 0" class="empty-mobile">
              <div class="empty-content">
                <span class="empty-icon">📭</span>
                <span class="empty-text">暂无归档记录</span>
                <span class="empty-sparkle">✨</span>
              </div>
            </div>
            <div v-for="(item, index) in sortedData" 
                 :key="item.id" 
                 :class="{ 'fade-out': item.deleting, 'mobile-card': true }"
                 :style="{ animationDelay: index * 0.1 + 's' }">
              <div class="card-header">
                <div class="card-title">
                  <span class="title-icon">📝</span>
                  <span class="title-text">{{ item.title }}</span>
                </div>
                <div class="card-time">
                  <span class="time-icon">⏱️</span>
                  <span class="time-text">{{ item.execute_time }}</span>
                </div>
              </div>
              <div class="card-actions">
                <button class="btn delete-btn mobile-delete-btn" 
                        @click="deleteItem(item)" 
                        :disabled="item.deleting">
                  <span class="delete-icon">{{ item.deleting ? '⏳' : '🗑️' }}</span>
                  <span class="delete-text">{{ item.deleting ? '删除中...' : '删除' }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- 底部装饰 -->
    <div class="bottom-decoration">
      <div class="bottom-ribbon"></div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { apiMethods } from '@/utils/api'
import { message, Modal } from 'ant-design-vue'

export default {
  name: 'Archive',
  setup() {
    const router = useRouter()
    const archiveData = ref([])
    const currentSort = ref({ key: 'created_at', asc: false })
    const loading = ref(false)
    const error = ref(null)

    // 计算排序后的数据
    const sortedData = computed(() => {
      const sorted = [...archiveData.value]
      sorted.sort((a, b) => {
        const valA = String(a[currentSort.value.key] || '')
        const valB = String(b[currentSort.value.key] || '')
        return currentSort.value.asc
          ? valA.localeCompare(valB)
          : valB.localeCompare(valA)
      })
      return sorted
    })

    // 排序方法
    const sortBy = (key) => {
      if (currentSort.value.key === key) {
        currentSort.value.asc = !currentSort.value.asc
      } else {
        currentSort.value.key = key
        currentSort.value.asc = true
      }
    }

    // 加载归档列表
    const loadArchiveList = async () => {
      try {
        loading.value = true
        error.value = null
        const data = await apiMethods.getArchiveList()
        archiveData.value = data
      } catch (err) {
        console.error('加载归档列表失败:', err)
        error.value = '加载归档列表失败，请稍后重试'
        archiveData.value = []
      } finally {
        loading.value = false
      }
    }

    // 删除归档记录
    const deleteItem = async (item) => {
      if (!confirm(`确认删除归档「${item.title}」吗？`)) return
      
      try {
        item.deleting = true
        await apiMethods.deleteArchive(item.id)
        
        // 添加淡出效果
        setTimeout(() => {
          archiveData.value = archiveData.value.filter(r => r.id !== item.id)
        }, 500)
      } catch (error) {
        console.error('删除失败:', error)
        alert('删除失败')
        item.deleting = false
      }
    }

    // 返回首页
    const goHome = () => {
      router.push('/')
    }

    // 全部删除归档记录
    const allDelete =() => {

      Modal.confirm({
        title: '确认删除?',
        content: '确认删除所有归档记录吗？',
        okText: '确定',
        cancelText: '取消',
        onOk: async () => {
          try {
            await apiMethods.deleteAllArchive()
            message.success('全部归档记录已删除！')
          } catch (error) {
            console.log('删除失败:', error)
            message.error('删除失败，请稍后重试')
          }
        }
      })

    }
  

    onMounted(() => {
      loadArchiveList()
    })

    return {
      archiveData,
      sortedData,
      currentSort,
      loading,
      error,
      sortBy,
      deleteItem,
      goHome,
      allDelete
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
  --accent-color: #ff8fab;
  --shadow-color: rgba(255, 110, 180, 0.2);
}

* { 
  box-sizing: border-box; 
  margin: 0; 
  padding: 0; 
}

.archive-page {
  font-family: "Comic Sans MS", "Segoe UI", sans-serif;
  background: linear-gradient(135deg, #ff6eb4 0%, #ffe6f2 50%, #fff0f8 100%);
  color: var(--text-color);
  min-height: 100vh;
  padding-bottom: 50px;
  position: relative;
}

/* 装饰性浮动元素 */
.decorative-elements {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 1;
}

.floating-heart {
  position: absolute;
  font-size: 1.5rem;
  animation: float 6s ease-in-out infinite;
  opacity: 0.6;
}

.heart-1 { top: 10%; left: 10%; animation-delay: 0s; }
.heart-2 { top: 20%; right: 15%; animation-delay: 1s; }
.heart-3 { top: 60%; left: 5%; animation-delay: 2s; }
.heart-4 { top: 40%; right: 10%; animation-delay: 3s; }
.heart-5 { top: 80%; right: 20%; animation-delay: 4s; }

@keyframes float {
  0%, 100% { transform: translateY(0px) rotate(0deg); }
  50% { transform: translateY(-20px) rotate(5deg); }
}

/* 头部样式 */
.cute-header {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.9) 0%, rgba(255, 240, 248, 0.9) 100%);
  padding: 30px 0 20px;
  text-align: center;
  box-shadow: 0 8px 32px var(--shadow-color);
  border-radius: 0 0 40px 40px;
  position: sticky;
  top: 0;
  z-index: 10;
}

.header-content {
  position: relative;
  z-index: 2;
}

.cute-title {
  color: var(--primary-color);
  font-size: 2.2rem;
  text-shadow: 2px 2px 4px rgba(255, 110, 180, 0.3);
  margin-top: 15px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-weight: bold;
}

.title-icon, .title-sparkle {
  animation: sparkle 2s ease-in-out infinite;
}

@keyframes sparkle {
  0%, 100% { transform: scale(1) rotate(0deg); }
  50% { transform: scale(1.2) rotate(10deg); }
}

.cute-subtitle {
  color: var(--primary-color);
  font-size: 1.8rem;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: bold;
  text-shadow: 1px 1px 2px rgba(255, 110, 180, 0.2);
}

.subtitle-icon {
  animation: bounce 2s ease-in-out infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-5px); }
}

/* 按钮样式 */
.cute-btn {
  background: linear-gradient(135deg, #fde0e0 0%, #dda6c2 100%);
  color: var(--primary-color);
  border: 3px solid var(--primary-color);
  border-radius: 25px;
  padding: 10px 20px;
  font-size: 2rem;
  cursor: pointer;
  box-shadow: 0 4px 15px var(--shadow-color);
  transition: all 0.3s ease;
  font-weight: bold;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.cute-btn:hover {
  background: linear-gradient(135deg, #bde1e2 0%, #dda6c2 100%);
  color: #31bfe2;
  box-shadow: 0 6px 20px var(--shadow-color);
  transform: translateY(-2px) scale(1.05);
}

.btn-icon {
  font-size: 1.1rem;
}

/* 面板样式 */
.cute-panel {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95) 0%, rgba(255, 248, 252, 0.95) 100%);
  box-shadow: 0 10px 30px var(--shadow-color);
  border-radius: 25px;
  padding: 30px;
  margin-bottom: 30px;
  position: relative;
  border: 2px solid rgba(255, 192, 218, 0.3);
}

.panel-header {
  position: relative;
  margin-bottom: 25px;
}

.panel-decoration {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.corner-decoration {
  position: absolute;
  font-size: 1.2rem;
  opacity: 0.6;
  animation: rotate 4s linear infinite;
}

.corner-tl { top: 10px; left: 10px; }
.corner-tr { top: 10px; right: 10px; }
.corner-bl { bottom: 10px; left: 10px; }
.corner-br { bottom: 10px; right: 10px; }

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 表格样式 */
.table-container {
  position: relative;
  overflow: hidden;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.5);
  padding: 5px;
}

.cute-table {
  width: 100%;
  border-collapse: collapse;
  border-spacing: 0;
  border-radius: 15px;
  overflow: hidden;
  box-shadow: 0 5px 15px var(--shadow-color);
  table-layout: fixed;
}

.desktop-table {
  display: table; /* Ensure it's a block-level element for desktop */
}

.mobile-list {
  display: none; /* Hide by default on desktop */
}

.table-header-row {
  background: linear-gradient(135deg, rgba(255, 182, 193, 0.3) 0%, rgba(255, 192, 218, 0.3) 100%);
}

.sortable-header, .action-header {
  padding: 15px 12px;
  text-align: left;
  font-weight: bold;
  cursor: pointer;
  border: none;
  position: relative;
  transition: all 0.3s ease;
  vertical-align: middle;
}

.sortable-header {
  text-align: left;
}

.action-header {
  text-align: center;
}

.sortable-header:hover {
  background: rgba(255, 192, 218, 0.4);
  transform: translateY(-1px);
}

.header-text {
  display: flex;
  align-items: center;
  gap: 5px;
  flex: 1;
}

.sort-indicator {
  font-size: 0.9rem;
  color: var(--accent-color);
  animation: pulse 1s ease-in-out infinite;
  margin-left: 5px;
}

@keyframes pulse {
  0%, 100% { opacity: 0.7; }
  50% { opacity: 1; }
}

.table-row {
  background: rgba(255, 255, 255, 0.8);
  transition: all 0.3s ease;
  animation: slideIn 0.5s ease-out forwards;
  opacity: 0;
  transform: translateY(20px);
}

@keyframes slideIn {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.table-row:hover {
  background: rgba(255, 240, 248, 0.9);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px var(--shadow-color);
}

.title-cell, .time-cell, .action-cell {
  padding: 12px 15px;
  border: none;
  border-bottom: 5px solid rgba(245, 118, 171, 0.3);
  transition: all 0.3s ease;
  display: table-cell;
  vertical-align: middle;
}

.title-cell {
  font-weight: 500;
  color: var(--text-color);
  width: 40%;
}

.time-cell {
  color: var(--accent-color);
  font-weight: 500;
  width: 35%;
}

.action-cell {
  text-align: center;
  width: 25%;
}

/* 删除按钮样式 */
.cute-delete-btn {
  background: linear-gradient(135deg, #fff 0%, #fff6fb 100%);
  color: #ff6b6b;
  border: 2px solid #ff6b6b;
  border-radius: 20px;
  padding: 8px 16px;
  font-size: 0.9rem;
  cursor: pointer;
  box-shadow: 0 3px 10px rgba(255, 107, 107, 0.2);
  transition: all 0.3s ease;
  font-weight: bold;
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.cute-delete-btn:hover {
  background: linear-gradient(135deg, #ff6b6b 0%, #ff8a8a 100%);
  color: #fff;
  box-shadow: 0 5px 15px rgba(255, 107, 107, 0.3);
  transform: translateY(-1px) scale(1.05);
}

.delete-icon {
  font-size: 1rem;
}

/* 空状态样式 */
.empty-row {
  background: transparent;
}

.empty-message {
  text-align: center;
  padding: 40px 20px;
  border: none;
}

.empty-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: var(--accent-color);
}

.empty-icon {
  font-size: 3rem;
  opacity: 0.6;
  animation: float 3s ease-in-out infinite;
}

.empty-text {
  font-size: 1.2rem;
  font-weight: 500;
}

.empty-sparkle {
  font-size: 1.5rem;
  animation: sparkle 2s ease-in-out infinite;
}

/* 淡出动画 */
tr.fade-out {
  opacity: 0;
  transform: translateX(-20px);
  transition: all 0.5s ease;
}

/* 加载和错误状态样式 */
.loading-state, .error-state {
  text-align: center;
  padding: 40px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
}

.loading-spinner {
  font-size: 2rem;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.loading-text {
  font-size: 1.1rem;
  color: var(--accent-color);
  font-weight: 500;
}

.error-icon {
  font-size: 2.5rem;
  opacity: 0.7;
}

.error-text {
  font-size: 1rem;
  color: #ff6b6b;
  font-weight: 500;
  max-width: 300px;
  line-height: 1.4;
}

.retry-btn {
  background: linear-gradient(135deg, #fff 0%, #fff6fb 100%);
  color: var(--primary-color);
  border: 2px solid var(--primary-color);
  border-radius: 20px;
  padding: 10px 20px;
  font-size: 0.9rem;
  cursor: pointer;
  box-shadow: 0 3px 10px var(--shadow-color);
  transition: all 0.3s ease;
  font-weight: bold;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.retry-btn:hover {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: #fff;
  box-shadow: 0 5px 15px var(--shadow-color);
  transform: translateY(-1px) scale(1.05);
}

.retry-icon {
  font-size: 1rem;
}

/* 移动端卡片样式 */
.mobile-card {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.9) 0%, rgba(255, 248, 252, 0.9) 100%);
  border: 1px solid rgba(198, 13, 90, 0.799);
  border-radius: 15px;
  padding: 15px;
  margin-bottom: 15px;
  box-shadow: 0 3px 10px var(--shadow-color);
  transition: all 0.3s ease;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 15px;
}

.mobile-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 6px 15px var(--shadow-color);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}

.card-title {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--text-color);
  display: flex;
  align-items: center;
  gap: 5px;
  flex: 1;
  min-width: 0;
}

.card-title .title-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-time {
  font-size: 0.9rem;
  color: var(--accent-color);
  display: flex;
  align-items: center;
  gap: 5px;
  flex-shrink: 0;
}

.card-time .time-text {
  white-space: nowrap;
}

.card-actions {
  display: flex;
  justify-content: flex-end;
}

.mobile-delete-btn {
  background: linear-gradient(135deg, #fff 0%, #fff6fb 100%);
  color: #ff6b6b;
  border: 2px solid #ff6b6b;
  border-radius: 15px;
  padding: 8px 12px;
  font-size: 0.8rem;
  cursor: pointer;
  box-shadow: 0 3px 10px rgba(255, 107, 107, 0.2);
  transition: all 0.3s ease;
  font-weight: bold;
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.mobile-delete-btn:hover {
  background: linear-gradient(135deg, #ff6b6b 0%, #ff8a8a 100%);
  color: #fff;
  box-shadow: 0 5px 15px rgba(255, 107, 107, 0.3);
  transform: translateY(-1px) scale(1.05);
}

.delete-text {
  font-size: 0.8rem;
}

/* 容器样式 */
.container {
  max-width: 1000px;
  margin: 30px auto;
  padding: 0 20px;
  position: relative;
  z-index: 2;
}

/* 底部装饰 */
.bottom-decoration {
  position: fixed;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 20px;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  opacity: 0.3;
  z-index: 1;
}

.bottom-ribbon {
  height: 100%;
  background: repeating-linear-gradient(
    45deg,
    transparent,
    transparent 10px,
    rgba(255, 255, 255, 0.1) 10px,
    rgba(255, 255, 255, 0.1) 20px
  );
}

.deleteBtn{
  text-align: center;
  margin-top: 20px;
  border: 1px solid #ff6b6b;
  background-color: #ffe6f2;
  color: #ff6b6b;
  height: 40px;
  width: 100px;
  border-radius: 10px;
  font-size: 1.2rem;
  font-family: Arial, Helvetica, sans-serif;
  font-weight: bold;
}

.deleteBtn:hover{
  background-color: #ff6b6b;
  color: white;
  border: 1px solid #ff6b6b;
  cursor: pointer;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .cute-title {
    font-size: 1.6rem;
    flex-direction: column;
    gap: 5px;
  }
  
  .title-text {
    font-size: 1.4rem;
  }
  
  .cute-subtitle {
    font-size: 1.3rem;
  }
  
  .subtitle-text {
    font-size: 1.1rem;
  }
  
  .cute-panel {
    padding: 15px;
    margin: 0 10px 20px;
    border-radius: 20px;
  }
  
  .table-container {
    overflow-x: auto;
  }
  
  .cute-table {
    min-width: 650px;
  }

  .desktop-table {
    display: none; /* Hide desktop table on mobile */
  }

  .mobile-list {
    display: block; /* Show mobile list on mobile */
  }
  
  .mobile-card {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
    padding: 12px;
  }
  
  .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .card-title {
    font-size: 1.1rem;
    width: 100%;
  }
  
  .card-time {
    font-size: 0.85rem;
    width: 100%;
  }
  
  .card-actions {
    justify-content: center;
    width: 100%;
  }
  
  .mobile-delete-btn {
    width: 100%;
    justify-content: center;
    padding: 10px;
    font-size: 0.9rem;
  }
  
  .cute-btn {
    padding: 8px 15px;
    font-size: 0.9rem;
  }
  
  .btn-text {
    font-size: 0.9rem;
  }
  
  .floating-heart {
    display: none;
  }
  
  .container {
    padding: 0 15px;
    margin: 20px auto;
  }
  
  .cute-header {
    padding: 20px 0 15px;
    border-radius: 0 0 30px 30px;
  }
  
  .header-content {
    padding: 0 15px;
  }
  
  .empty-mobile {
    text-align: center;
    padding: 40px 20px;
  }
  
  .empty-mobile .empty-content {
    flex-direction: column;
    gap: 15px;
  }
  
  .empty-mobile .empty-icon {
    font-size: 2.5rem;
  }
  
  .empty-mobile .empty-text {
    font-size: 1rem;
  }
}

/* 超小屏幕适配 */
@media (max-width: 480px) {
  .cute-title {
    font-size: 1.4rem;
  }
  
  .title-text {
    font-size: 1.2rem;
  }
  
  .cute-subtitle {
    font-size: 1.1rem;
  }
  
  .subtitle-text {
    font-size: 2rem;
    margin-left: 20px;
  }
  
  .cute-panel {
    padding: 12px;
    margin: 0 8px 15px;
  }
  
  .mobile-card {
    padding: 10px;
    margin-bottom: 12px;
  }
  
  .card-title {
    font-size: 1rem;
  }
  
  .card-time {
    font-size: 0.8rem;
  }
  
  .mobile-delete-btn {
    padding: 8px;
    font-size: 0.85rem;
  }
  
  .container {
    padding: 0 10px;
    margin: 15px auto;
  }
  
  .cute-header {
    padding: 15px 0 10px;
  }
  
  .header-content {
    padding: 0 10px;
  }
  
  .loading-state, .error-state {
    padding: 30px 15px;
  }
  
  .loading-spinner {
    font-size: 1.8rem;
  }
  
  .loading-text {
    font-size: 1rem;
  }
  
  .error-icon {
    font-size: 2rem;
  }
  
  .error-text {
    font-size: 0.9rem;
    max-width: 250px;
  }
  
  .retry-btn {
    padding: 8px 16px;
    font-size: 0.85rem;
  }
}

/* 禁用状态 */
.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.btn:disabled:hover {
  transform: none;
  box-shadow: none;
}

/* 移动端触摸优化 */
@media (max-width: 768px) {
  .btn {
    min-height: 44px; /* 确保触摸目标足够大 */
    touch-action: manipulation; /* 优化触摸响应 */
  }
  
  .mobile-delete-btn {
    min-height: 40px;
  }
  
  .sortable-header {
    min-height: 50px;
    padding: 12px 8px;
  }
  
  /* 移动端滚动优化 */
  .table-container {
    -webkit-overflow-scrolling: touch;
  }
  
  /* 移动端卡片触摸反馈 */
  .mobile-card {
    -webkit-tap-highlight-color: transparent;
    user-select: none;
  }
  
  .mobile-card:active {
    transform: scale(0.98);
  }
  
  /* 移动端按钮触摸反馈 */
  .btn:active {
    transform: scale(0.95);
  }
  
  .mobile-delete-btn:active {
    transform: scale(0.95);
  }
}

/* 横屏模式优化 */
@media (max-width: 768px) and (orientation: landscape) {
  .cute-header {
    padding: 15px 0 10px;
  }
  
  .cute-title {
    font-size: 1.4rem;
  }
  
  .title-text {
    font-size: 1.2rem;
  }
  
  .container {
    margin: 15px auto;
  }
  
  .mobile-card {
    flex-direction: row;
    align-items: center;
    padding: 10px 12px;
  }
  
  .card-header {
    flex-direction: row;
    align-items: center;
    flex: 1;
  }
  
  .card-actions {
    flex-shrink: 0;
  }
  
  .mobile-delete-btn {
    width: auto;
    padding: 8px 12px;
  }
}
</style>
