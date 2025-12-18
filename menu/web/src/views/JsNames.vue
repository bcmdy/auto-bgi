<template>
  <div class="app-container anime-theme">
    <div class="bg-pattern"></div>
    
    <header class="navbar">
      <div class="nav-content">
        <div class="nav-left">
          <span class="app-title">✨ 脚本屋 🌸</span>
          <span class="badge-total" v-if="pluginData.length">{{ pluginData.length }}</span>
        </div>
        <div class="nav-right">
          <button class="icon-btn" @click="goHome" title="首页">🏰</button>
        </div>
      </div>
      
      <div class="filter-bar">
        <div class="search-box">
          <span class="search-icon">🔎</span>
          <input 
            v-model="searchText" 
            type="text" 
            placeholder="寻找神奇脚本..." 
            class="search-input"
          />
          <span v-if="searchText" class="clear-icon" @click="searchText=''">✖</span>
        </div>
        <div class="tabs">
          <button 
            class="tab-item" 
            :class="{ active: filterTab === 'all' }" 
            @click="filterTab = 'all'"
          >全部</button>
          <button 
            class="tab-item" 
            :class="{ active: filterTab === 'update' }" 
            @click="filterTab = 'update'"
          >
            待升级
            <span class="dot" v-if="updateCount > 0"></span>
          </button>
        </div>
      </div>
    </header>

    <section class="quick-actions">
      <div class="action-item jelly-hover" @click="batchUpdate">
        <div class="action-icon bg-pink">🚀</div>
        <span class="action-text">一键升级</span>
      </div>
      <div class="action-item jelly-hover" @click="openSubscribeModal">
        <div class="action-icon bg-purple">🎀</div>
        <span class="action-text">订阅契约</span>
      </div>
      <div class="action-item jelly-hover" @click="resetRepo">
        <div class="action-icon bg-yellow">💫</div>
        <span class="action-text">重置仓库</span>
      </div>
      <div class="action-item jelly-hover" @click="loadPluginList">
        <div class="action-icon bg-blue">🌀</div>
        <span class="action-text">刷新列表</span>
      </div>
    </section>

  

    <main class="main-list">
      
      <div v-if="filteredList.length === 0" class="empty-state">
        <div class="empty-img">(｡•́︿•̀｡)</div>
        <p>这里空空如也...</p>
      </div>

      <div class="script-list">
        <div 
          v-for="item in filteredList" 
          :key="item.Name" 
          class="script-card"
          :class="{ 'needs-update': item.Mark === '有更新' }"
        >
          <div class="card-main">
            <div class="card-icon">
              {{ item.Mark === '有更新' ? '⚡' : '📜' }}
            </div>
            <div class="card-info">
              <div class="card-header">
                <h3 class="script-name">{{ item.ChineseName }}</h3>
                <span class="tag" :class="getTagClass(item.Mark)">{{ item.Mark }}</span>
              </div>
              <div class="card-meta">
                <div class="version-row">
                  <span class="ver-badge cur">v{{ item.NowVersion }}</span>
                  <span class="arrow" v-if="item.Mark === '有更新'">➜</span>
                  <span class="ver-badge new" v-if="item.Mark === '有更新'">v{{ item.NewVersion }}</span>
                </div>
                <div class="time-text">{{ item.LastUpdated }}</div>
              </div>
            </div>
          </div>
          
          <div class="card-action" v-if="item.Mark === '有更新' || isUpdating[item.Name]">
              <button 
                class="btn-update" 
                :disabled="isUpdating[item.Name]"
                @click="updatePlugin(item.Name)"
              >
                <span v-if="isUpdating[item.Name]" class="loading-spin">🍬</span>
                {{ isUpdating[item.Name] ? '升级中...' : '✨ 立即升级' }}
              </button>
          </div>
        </div>
      </div>
    </main>

    <transition name="pop">
      <div v-if="showModal" class="modal-mask" @click.self="closeSubscribeModal">
        <div class="modal-panel">
          <div class="modal-header">
            <h3>📝 签订契约</h3>
            <button class="close-btn" @click="closeSubscribeModal">✕</button>
          </div>
          <div class="modal-body">
            <p class="input-label">神秘的脚本</p>
            <input 
              v-model="subscribeInput" 
              type="text" 
              class="modal-input" 
              placeholder="在此输入脚本"
              ref="subInputRef"
            />
            <div class="modal-tips">
              提示：请输入脚本文件夹名字
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn-cancel" @click="closeSubscribeModal">再想想</button>
            <button class="btn-confirm" @click="confirmSubscribe" :disabled="!subscribeInput">
              确认添加 ❤️
            </button>
          </div>
        </div>
      </div>
    </transition>

  </div>
</template>

<script>
import { ref, computed, onMounted, onUnmounted, reactive, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { apiMethods } from '@/utils/api'

export default {
  name: 'JsNamesAnimeTheme',
  setup() {
    const router = useRouter()
    
    // 数据状态
    const pluginData = ref([])
    const isUpdating = reactive({})
    const headerCurrentImageIndex = ref(0)
    let carouselTimer = null

    // 交互状态
    const searchText = ref('')
    const filterTab = ref('all') // 'all' or 'update'
    const showModal = ref(false)
    const subscribeInput = ref('')
    const subInputRef = ref(null)

    // --- 计算属性 ---
    const updateCount = computed(() => {
      return pluginData.value.filter(i => i.Mark === '有更新').length
    })

    const filteredList = computed(() => {
      let list = pluginData.value

      if (filterTab.value === 'update') {
        list = list.filter(item => item.Mark === '有更新')
      }

      if (searchText.value) {
        const key = searchText.value.toLowerCase()
        list = list.filter(item => 
          (item.ChineseName && item.ChineseName.toLowerCase().includes(key)) ||
          (item.Name && item.Name.toLowerCase().includes(key))
        )
      }

      return list.sort((a, b) => {
        if (a.Mark === '有更新' && b.Mark !== '有更新') return -1
        if (a.Mark !== '有更新' && b.Mark === '有更新') return 1
        return 0
      })
    })

    // --- 方法 (保持原有逻辑不变) ---

    const loadPluginList = async () => {
      try {
        const data = await apiMethods.getJsNames()
        pluginData.value = data.data || []
      } catch (e) {
        console.error(e)
        // 模拟数据用于预览效果
        pluginData.value = [
  
        ]
      }
    }

    const updatePlugin = async (name) => {
      if (!name) return
      isUpdating[name] = true
      try {
        await apiMethods.updateJs(name)
        await loadPluginList()
      } catch (e) {
        alert('更新失败: ' + e.message)
      } finally {
        isUpdating[name] = false
      }
    }

    const batchUpdate = async () => {
      if (updateCount.value === 0) return alert('当前没有需要更新的脚本哦~')
      if (!confirm(`准备好批量更新 ${updateCount.value} 个脚本了吗？`)) return
      
      try {
        await apiMethods.batchUpdate()
        alert('请求已发送，正在努力更新中...')
        loadPluginList()
      } catch (e) {
        alert('操作失败')
      }
    }

    const resetRepo = async () => {
      if (!confirm('⚠️ 警告：重置仓库会覆盖本地修改，真的要重置吗？')) return
      try {
        await apiMethods.resetRepo()
        alert('仓库已重置完毕')
        loadPluginList()
      } catch (e) {
        alert('重置失败')
      }
    }

    const openSubscribeModal = () => {
      subscribeInput.value = ''
      showModal.value = true
      nextTick(() => {
        if(subInputRef.value) subInputRef.value.focus()
      })
    }
    const closeSubscribeModal = () => showModal.value = false
    const confirmSubscribe = async () => {
      if (!subscribeInput.value) return
      try {
        const res = await apiMethods.subscribeScript(subscribeInput.value.trim())
        if (res.message && res.message.includes('成功')) {
            alert('🎉 契约签订成功！')
            closeSubscribeModal()
            loadPluginList()
        } else {
            throw new Error(res.error || '未知错误')
        }
      } catch (e) {
        alert('订阅失败: ' + e.message)
      }
    }



    const getTagClass = (mark) => {
      if (mark === '有更新') return 'tag-update'
      if (mark === '未知') return 'tag-unknown'
      return 'tag-normal'
    }
    


    onMounted(() => {
      loadPluginList()
    })

    onUnmounted(() => {
      if (carouselTimer) clearInterval(carouselTimer)
    })

    return {
      pluginData,
      filteredList,
      updateCount,
      isUpdating,
      searchText,
      filterTab,
      showModal,
      subscribeInput,
      subInputRef,
      loadPluginList,
      updatePlugin,
      batchUpdate,
      resetRepo,
      openSubscribeModal,
      closeSubscribeModal,
      confirmSubscribe,
      getTagClass,
      goHome: () => router.push('/')
    }
  }
}
</script>

<style scoped>
/* --- 🌸 二次元主题变量 --- */
:root {
  --primary: #FB7299;      /* B站粉 */
  --primary-hover: #FF85AD;
  --secondary: #23ADE5;    /* 配合粉色的天蓝 */
  --bg-color: #FFF1F5;     /* 极浅粉背景 */
  --card-bg: #FFFFFF;
  --text-main: #505050;    /* 柔和的深灰 */
  --text-sub: #999999;
  --border-radius: 20px;   /* 大圆角 */
  --shadow-soft: 0 8px 24px rgba(251, 114, 153, 0.15); /* 粉色光晕 */
  --shadow-hover: 0 12px 32px rgba(251, 114, 153, 0.25);
  --gap-safe: env(safe-area-inset-bottom);
}

.app-container {
  min-height: 100vh;
  background-color: var(--bg-color);
  font-family: "Varela Round", "PingFang SC", "Microsoft YaHei", sans-serif; /* 圆润字体优先 */
  color: var(--text-main);
  padding-bottom: calc(20px + var(--gap-safe));
  position: relative;
  overflow-x: hidden;
}

/* 背景波点纹理 */
.bg-pattern {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background-image: radial-gradient(#FFD1DC 10%, transparent 10%);
  background-size: 24px 24px;
  opacity: 0.4;
  z-index: 0;
  pointer-events: none;
}

/* --- 1. 顶部导航栏 (玻璃拟态) --- */
.navbar {
  position: sticky;
  top: 10px;
  z-index: 100;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  margin: 10px 16px;
  border-radius: 24px;
  padding: 12px 16px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.05);
  border: 2px solid #FFF;
}

.nav-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.app-title {
  font-size: 22px;
  font-weight: 900;
  color: var(--primary);
  letter-spacing: 1px;
  text-shadow: 1px 1px 0 #FFF;
}

.badge-total {
  font-size: 12px;
  background: var(--primary);
  color: #e60d79;
  padding: 2px 10px;
  border-radius: 12px;
  margin-left: 8px;
  font-weight: bold;
  box-shadow: 0 2px 6px rgba(251, 114, 153, 0.4);
}

.icon-btn {
  background: #FFF;
  border: 2px solid #FFE6EE;
  border-radius: 50%;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  transition: transform 0.2s;
}
.icon-btn:active { transform: scale(0.9); }

/* 搜索与 Tab */
.filter-bar {
  display: flex;
  gap: 12px;
}

.search-box {
  flex: 1;
  background: #FFF;
  border: 2px solid #FFE6EE;
  border-radius: 20px;
  display: flex;
  align-items: center;
  padding: 0 12px;
  height: 40px;
  transition: border-color 0.3s;
}
.search-box:focus-within {
  border-color: var(--primary);
}

.search-input {
  border: none;
  background: transparent;
  width: 100%;
  font-size: 14px;
  color: var(--text-main);
  outline: none;
  margin-left: 8px;
}
.search-input::placeholder { color: #FFC0CB; }

.tabs {
  display: flex;
  background: #FFF;
  border: 2px solid #FFE6EE;
  border-radius: 20px;
  padding: 4px;
}

.tab-item {
  border: none;
  background: transparent;
  padding: 0 16px;
  font-size: 14px;
  color: #e009a0;
  border-radius: 16px;
  height: 32px;
  font-weight: 600;
  transition: all 0.3s;
  position: relative;
}

.tab-item.active {
  background: var(--primary);
  color: #470ce9;
  box-shadow: 0 2px 8px rgba(251, 114, 153, 0.4);
}

.dot {
  position: absolute;
  top: 4px;
  right: 6px;
  width: 8px;
  height: 8px;
  background: #FFD700;
  border: 1px solid #FFF;
  border-radius: 50%;
}

/* --- 2. 快捷功能区 (图标拟物化) --- */
.quick-actions {
  display: flex;
  justify-content: space-between;
  padding: 8px 20px;
  margin-bottom: 8px;
  position: relative;
  z-index: 1;
}

.action-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  width: 25%;
  cursor: pointer;
}

.action-icon {
  width: 48px;
  height: 48px;
  border-radius: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  border: 3px solid #FFF;
  box-shadow: 0 6px 12px rgba(251, 114, 153, 0.15);
  transition: all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.jelly-hover:active .action-icon {
  transform: scale(0.9);
}

.bg-pink { background: linear-gradient(135deg, #FF9A9E 0%, #FECFEF 100%); }
.bg-purple { background: linear-gradient(135deg, #A18CD1 0%, #FBC2EB 100%); }
.bg-yellow { background: linear-gradient(135deg, #F6D365 0%, #FDA085 100%); }
.bg-blue { background: linear-gradient(135deg, #89f7fe 0%, #66a6ff 100%); }

.action-text {
  font-size: 12px;
  font-weight: bold;
  color: #777;
}

/* --- 3. 轮播图 --- */
.banner-area {
  padding: 0 16px 16px;
  position: relative;
  z-index: 1;
}
.banner-wrapper {
  position: relative;
  height: 130px;
  border-radius: 20px;
  overflow: hidden;
  box-shadow: var(--shadow-soft);
  border: 4px solid #FFF;
}
.banner-img {
  width: 100%;
  height: 400%;
  object-fit: cover;
}
.banner-dots {
  position: absolute;
  bottom: 8px;
  right: 12px; /* 移到右侧 */
  display: flex;
  gap: 6px;
}
.banner-dots .dot {
  width: 8px;
  height: 8px;
  background: rgba(255,255,255,0.6);
  border-radius: 50%;
  position: static;
  transition: all 0.3s;
}
.banner-dots .dot.active {
  background: #FFF;
  transform: scale(1.2);
  width: 20px;
  border-radius: 10px;
}

/* --- 4. 列表卡片 --- */
.main-list {
  padding: 0 16px;
  position: relative;
  z-index: 1;
}

.empty-state {
  text-align: center;
  padding: 60px 0;
  color: #FFB7C5;
}
.empty-img { font-size: 60px; margin-bottom: 10px; text-shadow: 2px 2px 0 #FFF; }

.script-card {
  background: #FFF;
  border-radius: 24px;
  margin-bottom: 16px;
  padding: 18px;
  box-shadow: var(--shadow-soft);
  border: 2px solid transparent;
  transition: all 0.3s;
  position: relative;
  overflow: hidden;
}

.script-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-hover);
}

/* 待更新状态：粉色边框和浅色背景 */
.script-card.needs-update {
  border-color: #FFB6C1;
  background: #FFF8FA;
}
/* 装饰性丝带效果 */
.script-card.needs-update::before {
  content: '';
  position: absolute;
  top: 0; left: 0; width: 6px; height: 100%;
  background: var(--primary);
  border-radius: 4px 0 0 4px;
}

.card-main {
  display: flex;
  gap: 14px;
}

.card-icon {
  width: 48px;
  height: 48px;
  background: #FFF0F5;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  border: 1px solid #FFE4EA;
}

.card-info {
  flex: 1;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.script-name {
  margin: 0;
  font-size: 17px;
  font-weight: 700;
  color: #444;
}

.card-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.version-row {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-family: monospace; /* 保持版本号清晰 */
  background: #F7F8FA;
  padding: 4px 8px;
  border-radius: 10px;
}

.ver-badge {
  color: #666;
}

.ver-badge.new {
  color: var(--primary);
  font-weight: 800;
}

.arrow { color: #FFC0CB; font-weight: bold; }

.time-text {
  font-size: 12px;
  color: #BBB;
  font-weight: 500;
}

.tag {
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 10px;
  font-weight: bold;
}
.tag-update { 
    color: #f30606; 
    background: var(--primary); 
    box-shadow: 0 2px 6px rgba(251, 114, 153, 0.3);
}
.tag-normal { color: #52C41A; background: #F6FFED; border: 1px solid #B7EB8F; }
.tag-unknown { color: #B0B0B0; background: #F5F5F5; }

/* 按钮区域 */
.card-action {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 2px dashed #FFE4EA; /* 虚线分隔更可爱 */
}

.btn-update {
  width: 100%;
  border: none;
  background: linear-gradient(90deg, #FB7299, #FF5C8A);
  color: white;
  padding: 12px;
  border-radius: 16px;
  font-weight: 700;
  font-size: 15px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  box-shadow: 0 4px 12px rgba(251, 114, 153, 0.4);
  transition: all 0.2s;
}
.btn-update:active { transform: scale(0.97); }
.btn-update:disabled { 
    background: #E0E0E0; 
    box-shadow: none; 
    color: #999;
}

.loading-spin {
  animation: spin 1s infinite linear;
  display: inline-block;
}
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

/* --- 5. 弹窗 (Pop-up Style) --- */
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(255, 192, 203, 0.3); /* 粉色半透明遮罩 */
  backdrop-filter: blur(5px);
  z-index: 999;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-panel {
  background: #FFF;
  width: 80%;
  max-width: 320px;
  border-radius: 30px;
  padding: 24px;
  box-shadow: 0 20px 60px rgba(251, 114, 153, 0.3);
  border: 4px solid #FFF;
  text-align: center;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.modal-header h3 { margin: 0; font-size: 20px; color: var(--primary); }
.close-btn { 
    background: #F0F0F0; 
    border: none; 
    width: 32px; height: 32px; 
    border-radius: 50%; 
    color: #999; 
    font-weight: bold;
}

.input-label { 
    font-size: 15px; 
    color: #666; 
    margin-bottom: 12px; 
    font-weight: bold; 
    text-align: left;
}

.modal-input {
  width: 100%;
  box-sizing: border-box;
  padding: 14px;
  border: 2px solid #FFE6EE;
  border-radius: 16px;
  font-size: 16px;
  background: #FFFBFD;
  outline: none;
  color: var(--primary);
  text-align: center;
  transition: border 0.3s;
}
.modal-input:focus { 
    border-color: var(--primary); 
    background: #FFF; 
    box-shadow: 0 0 0 4px rgba(251, 114, 153, 0.1);
}

.modal-tips { font-size: 12px; color: #FFB7C5; margin-top: 10px; }

.modal-footer {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}
.btn-cancel, .btn-confirm {
  flex: 1;
  padding: 12px;
  border-radius: 18px;
  border: none;
  font-size: 15px;
  font-weight: 700;
}
.btn-cancel { background: #F5F5F5; color: #999; }
.btn-confirm { 
    background: var(--primary); 
    color: white; 
    box-shadow: 0 4px 10px rgba(251, 114, 153, 0.4);
}
.btn-confirm:disabled { background: #FFD1DC; box-shadow: none; }

/* 弹窗动画 */
.pop-enter-active, .pop-leave-active { transition: all 0.4s cubic-bezier(0.68, -0.55, 0.265, 1.55); }
.pop-enter-from, .pop-leave-to { transform: scale(0.5); opacity: 0; }

/* 桌面端适配 */
@media (min-width: 768px) {
  .app-container {
    background-color: #FFE6EE; /* 桌面端更深的粉色背景 */
  }
  .navbar, .quick-actions, .banner-area, .main-list {
    max-width: 800px;
    margin: 0 auto;
  }
  .navbar { margin-top: 20px; margin-bottom: 20px; }
  .script-list {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 20px;
  }
  .script-card { margin-bottom: 0; }
}
</style>