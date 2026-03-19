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
      <div class="action-item jelly-hover" @click="batchUpdate" :class="{ 'disabled-action': isGlobalLoading }">
        <div class="action-icon bg-pink">
          <span :class="{ 'loading-spin': isBatchUpdating }">{{ isBatchUpdating ? '🌀' : '🚀' }}</span>
        </div>
        <span class="action-text">{{ isBatchUpdating ? '升级中(2-5m)...' : '一键升级' }}</span>
      </div>
      <div class="action-item jelly-hover" @click="!isGlobalLoading && openSubscribeModal()" :class="{ 'disabled-action': isGlobalLoading }">
        <div class="action-icon bg-purple">🎀</div>
        <span class="action-text">签订契约</span>
      </div>
      <div class="action-item jelly-hover" @click="!isGlobalLoading && resetRepo()" :class="{ 'disabled-action': isGlobalLoading }">
        <div class="action-icon bg-yellow">💫</div>
        <span class="action-text">重置仓库</span>
      </div>
      <div class="action-item jelly-hover" @click="!isGlobalLoading && loadPluginList()" :class="{ 'disabled-action': isGlobalLoading }">
        <div class="action-icon bg-blue">
           <span :class="{ 'loading-spin': isRefreshing }">🌀</span>
        </div>
        <span class="action-text">{{ isRefreshing ? '刷新中(2-5m)...' : '刷新列表' }}</span>
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
          
          <div class="card-action">
            <button
              class="btn-history"
              :disabled="isGlobalLoading || isHistoryButtonLoading[item.Name]"
              @click="openHistoryModal(item)"
            >
              <span v-if="isHistoryButtonLoading[item.Name]" class="loading-spin">🌀</span>
              {{ isHistoryButtonLoading[item.Name] ? '加载中...' : '🕰️ 历史版本' }}
            </button>

            <button 
              v-if="item.Mark === '有更新' || isUpdating[item.Name]"
              class="btn-update" 
              :disabled="isGlobalLoading || isUpdating[item.Name]"
              @click="updatePlugin(item.Name)"
            >
              <span v-if="isUpdating[item.Name]" class="loading-spin">🍬</span>
              {{ isUpdating[item.Name] ? '升级中(2-5m)...' : '✨ 立即升级' }}
            </button>
          </div>
        </div>
      </div>
    </main>

    <transition name="pop">
      <div v-if="showModal" class="modal-mask" @click.self="closeSubscribeModal">
        <div class="modal-panel">
          <div class="modal-decorative-bg"></div>
          <div class="modal-header">
            <h3>📝 签订契约</h3>
            <button class="close-btn" @click="closeSubscribeModal">✕</button>
          </div>
          <div class="modal-body">
            <p class="input-label">神秘的脚本</p>
            
            <div class="input-wrapper">
              <span class="input-icon">🔮</span>
              <input 
                v-model="subscribeInput" 
                type="text" 
                class="modal-input" 
                placeholder="输入关键词搜索或填写脚本名"
                ref="subInputRef"
                @input="handleSearchInput"
                @focus="showSearchResult = true"
              />
              
              <transition name="fade">
                <div class="search-dropdown" v-if="showSearchResult && searchList.length > 0">
                  <ul>
                    <li 
                      v-for="(script, idx) in searchList" 
                      :key="idx" 
                      @click="selectScript(script)"
                      class="search-item"
                    >
                      <div class="item-name">{{ script.name }}</div>
                      <div class="item-desc">
                        {{ formatDesc(script.description) }}
                      </div>
                    </li>
                  </ul>
                </div>
              </transition>
              
              <div class="loading-indicator" v-if="isSearchingScript">
                <span class="loading-spin">🌸</span>
              </div>
            </div>

            <div class="modal-tips">
              提示：点击列表选中，或直接输入脚本文件夹名字
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

    <transition name="pop">
      <div v-if="showHistoryModal" class="modal-mask" @click.self="closeHistoryModal">
        <div class="modal-panel history-panel">
          <div class="modal-decorative-bg"></div>
          <div class="modal-header">
            <h3>🕰️ 历史版本</h3>
            <button class="close-btn" @click="closeHistoryModal">✕</button>
          </div>
          <div class="modal-body">
            <div class="history-title">
              <span class="history-title-label">脚本：</span>
              <span class="history-title-name">{{ historyChineseName || historyJsName }}</span>
            </div>

            <div v-if="isHistoryLoading" class="history-loading">
              <span class="loading-spin">🌀</span>
              加载中...
            </div>

            <div v-else-if="historyVersions.length === 0" class="history-empty">
              暂无历史版本
            </div>

            <div v-else class="history-list">
              <label v-for="ver in historyVersions" :key="ver" class="history-item">
                <input
                  type="radio"
                  name="historyVer"
                  :value="ver"
                  v-model="selectedHistoryVersion"
                />
                <span class="history-ver">v{{ ver }}</span>
                <span v-if="historyNowVersion && ver === historyNowVersion" class="history-current">当前</span>
              </label>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn-cancel" @click="closeHistoryModal">关闭</button>
            <button
              class="btn-confirm"
              :disabled="!selectedHistoryVersion || isHistoryLoading || isHistoryRollingBack"
              @click="rollbackSelectedHistoryVersion"
            >
              <span v-if="isHistoryRollingBack" class="loading-spin">🍬</span>
              回滚
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
    let carouselTimer = null

    // 交互状态
    const searchText = ref('')
    const filterTab = ref('all') // 'all' or 'update'
    const showModal = ref(false)
    const subscribeInput = ref('')
    const subInputRef = ref(null)
    const isRefreshing = ref(false)
    const isBatchUpdating = ref(false)

    // 搜索脚本相关状态
    const searchList = ref([])
    const showSearchResult = ref(false)
    const isSearchingScript = ref(false)
    let searchTimer = null

    const showHistoryModal = ref(false)
    const isHistoryLoading = ref(false)
    const isHistoryRollingBack = ref(false)
    const historyJsName = ref('')
    const historyChineseName = ref('')
    const historyNowVersion = ref('')
    const historyVersions = ref([])
    const selectedHistoryVersion = ref('')
    const isHistoryButtonLoading = reactive({})

    // --- 计算属性 ---
    const isGlobalLoading = computed(() => isRefreshing.value || isBatchUpdating.value)

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

    // --- 方法 ---

    const loadPluginList = async () => {
      if (isRefreshing.value) return
      isRefreshing.value = true
      try {
        const data = await apiMethods.getJsNames()
        pluginData.value = data.data || []
      } catch (e) {
        console.error(e)
        pluginData.value = []
      } finally {
        isRefreshing.value = false
      }
    }

    const updatePlugin = async (name) => {
      if (isGlobalLoading.value) return
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
      if (isGlobalLoading.value) return
      if (updateCount.value === 0) return alert('当前没有需要更新的脚本哦~')
      if (!confirm(`准备好批量更新 ${updateCount.value} 个脚本了吗？`)) return
      
      isBatchUpdating.value = true
      try {
        await apiMethods.batchUpdate()
        alert('请求已发送，正在努力更新中... (可能需要2-5分钟)')
        loadPluginList()
      } catch (e) {
        alert('操作失败')
      } finally {
        isBatchUpdating.value = false
      }
    }

    const resetRepo = async () => {
      if (isGlobalLoading.value) return
      if (!confirm('⚠️ 警告：重置仓库会覆盖本地修改，真的要重置吗？')) return
      try {
        await apiMethods.resetRepo()
        alert('仓库已重置完毕')
        loadPluginList()
      } catch (e) {
        alert('重置失败')
      }
    }

    const closeHistoryModal = () => {
      showHistoryModal.value = false
      setTimeout(() => {
        historyVersions.value = []
        selectedHistoryVersion.value = ''
        historyJsName.value = ''
        historyChineseName.value = ''
        historyNowVersion.value = ''
        isHistoryLoading.value = false
        isHistoryRollingBack.value = false
      }, 300)
    }

    const openHistoryModal = async (item) => {
      if (isGlobalLoading.value) return
      if (!item || !item.Name) return

      historyJsName.value = item.Name
      historyChineseName.value = item.ChineseName || ''
      historyNowVersion.value = item.NowVersion || ''
      historyVersions.value = []
      selectedHistoryVersion.value = ''
      showHistoryModal.value = true

      isHistoryLoading.value = true
      isHistoryButtonLoading[item.Name] = true
      try {
        const res = await apiMethods.queryHistoryVersion(item.Name)
        if (!res || res.success !== true) {
          throw new Error(res?.message || '获取历史版本失败')
        }
        historyVersions.value = Array.isArray(res.versions) ? res.versions : []
        if (historyVersions.value.length > 0) {
          selectedHistoryVersion.value = historyVersions.value[0]
        }
      } catch (e) {
        alert('获取历史版本失败: ' + (e?.message || e))
      } finally {
        isHistoryLoading.value = false
        isHistoryButtonLoading[item.Name] = false
      }
    }

    const rollbackSelectedHistoryVersion = async () => {
      if (isHistoryLoading.value || isHistoryRollingBack.value) return
      if (!historyJsName.value || !selectedHistoryVersion.value) return
      const displayName = historyChineseName.value || historyJsName.value
      if (!confirm(`确认将「${displayName}」回滚到 v${selectedHistoryVersion.value} 吗？`)) return

      isHistoryRollingBack.value = true
      try {
        const res = await apiMethods.rollbackHistoryVersion({
          version: selectedHistoryVersion.value,
          jsName: historyJsName.value
        })
        if (!res || res.success !== true) {
          throw new Error(res?.message || '回滚失败')
        }
        alert(res.message || '更新并还原成功')
        closeHistoryModal()
        await loadPluginList()
      } catch (e) {
        alert('回滚失败: ' + (e?.message || e))
      } finally {
        isHistoryRollingBack.value = false
      }
    }

    // --- 订阅/搜索模态框逻辑 ---
    const openSubscribeModal = () => {
      if (isGlobalLoading.value) return
      subscribeInput.value = ''
      searchList.value = []
      showSearchResult.value = false
      showModal.value = true
      nextTick(() => {
        if(subInputRef.value) subInputRef.value.focus()
      })
    }

    const closeSubscribeModal = () => {
      showModal.value = false
      // 延迟清空，防止闪烁
      setTimeout(() => {
        searchList.value = []
      }, 300)
    }

    // 处理输入，带防抖的搜索
    const handleSearchInput = () => {
      if (searchTimer) clearTimeout(searchTimer)
      
      const query = subscribeInput.value.trim()
      if (!query) {
        searchList.value = []
        showSearchResult.value = false
        return
      }

      searchTimer = setTimeout(async () => {
        isSearchingScript.value = true
        try {
          // 调用新增的接口
          const res = await apiMethods.getAllScripts(query)
          if (res && res.code === 200) {
            searchList.value = res.data || []
            showSearchResult.value = true
          }
        } catch (error) {
          console.error("搜索失败", error)
        } finally {
          isSearchingScript.value = false
        }
      }, 500) // 500ms 防抖
    }

    // 选中联想词
    const selectScript = (item) => {
      subscribeInput.value = item.name
      showSearchResult.value = false
    }

    // 格式化描述信息 (AAA狗粮批发~|~直接利用...)
    const formatDesc = (desc) => {
      if (!desc) return ''
      const parts = desc.split('~|~')
      return parts.length > 1 ? parts[1] : parts[0]
    }

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
      if (searchTimer) clearTimeout(searchTimer)
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
      // 新增状态
      searchList,
      showSearchResult,
      isSearchingScript,
      isRefreshing,
      isGlobalLoading,
      showHistoryModal,
      isHistoryLoading,
      isHistoryRollingBack,
      historyJsName,
      historyChineseName,
      historyNowVersion,
      historyVersions,
      selectedHistoryVersion,
      isHistoryButtonLoading,
      // 方法
      loadPluginList,
      updatePlugin,
      batchUpdate,
      resetRepo,
      openHistoryModal,
      closeHistoryModal,
      rollbackSelectedHistoryVersion,
      openSubscribeModal,
      closeSubscribeModal,
      confirmSubscribe,
      getTagClass,
      handleSearchInput,
      selectScript,
      formatDesc,
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
  font-family: "Varela Round", "PingFang SC", "Microsoft YaHei", sans-serif;
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

/* --- 1. 顶部导航栏 --- */
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

/* --- 2. 快捷功能区 --- */
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

.disabled-action {
  opacity: 0.6;
  cursor: not-allowed;
}
.disabled-action:active .action-icon {
  transform: none;
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

/* --- 3. 列表卡片 --- */
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

.script-card.needs-update {
  border-color: #FFB6C1;
  background: #FFF8FA;
}
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
  font-family: monospace;
  background: #F7F8FA;
  padding: 4px 8px;
  border-radius: 10px;
}

.ver-badge { color: #666; }
.ver-badge.new { color: var(--primary); font-weight: 800; }
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
.tag-update { color: #f30606; background: var(--primary); box-shadow: 0 2px 6px rgba(251, 114, 153, 0.3); }
.tag-normal { color: #52C41A; background: #F6FFED; border: 1px solid #B7EB8F; }
.tag-unknown { color: #B0B0B0; background: #F5F5F5; }

.card-action {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 2px dashed #FFE4EA;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.btn-history {
  width: 100%;
  border: none;
  background: linear-gradient(90deg, #23ADE5, #66a6ff);
  color: white;
  padding: 12px;
  border-radius: 16px;
  font-weight: 700;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  box-shadow: 0 4px 12px rgba(35, 173, 229, 0.35);
  transition: all 0.2s;
}
.btn-history:active { transform: scale(0.97); }
.btn-history:disabled { background: #E0E0E0; box-shadow: none; color: #999; }

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
.btn-update:disabled { background: #E0E0E0; box-shadow: none; color: #999; }

.loading-spin {
  animation: spin 1s infinite linear;
  display: inline-block;
}
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

/* --- 5. 弹窗 (Pop-up Style - 美化重构版) --- */
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(100, 100, 110, 0.4);
  backdrop-filter: blur(8px);
  z-index: 999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.modal-panel {
  background: linear-gradient(180deg, #FFF0F5 0%, #FFFFFF 100%);
  width: 90%;
  max-width: 720px;
  height: 700px;
  border-radius: 28px;
  padding: 24px;
  box-shadow: 0 20px 60px rgba(251, 114, 153, 0.35), 0 0 0 6px rgba(255, 255, 255, 0.6);
  border: 1px solid #FFE4EA;
  text-align: center;
  position: relative;
  overflow: visible; /* 允许下拉框超出 */
  display: flex;
  flex-direction: column;
}

.history-panel {
  max-width: 560px;
  height: auto;
  max-height: 80vh;
  overflow: hidden;
}

.history-title {
  display: flex;
  gap: 8px;
  align-items: baseline;
  margin-bottom: 12px;
  padding-left: 8px;
}
.history-title-label { font-size: 13px; color: #999; font-weight: bold; }
.history-title-name { font-size: 14px; color: var(--text-main); font-weight: 800; }

.history-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #999;
  padding: 32px 0;
}

.history-empty {
  text-align: center;
  color: #BBB;
  padding: 32px 0;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 6px 8px;
  overflow: auto;
  max-height: 46vh;
}

.history-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 16px;
  border: 2px solid #FFE6EE;
  background: #FFF;
  cursor: pointer;
}

.history-item input { accent-color: var(--primary); }
.history-ver { font-weight: 800; color: var(--primary); }
.history-current {
  margin-left: auto;
  font-size: 12px;
  color: #52C41A;
  background: #F6FFED;
  border: 1px solid #B7EB8F;
  padding: 2px 8px;
  border-radius: 999px;
  font-weight: 700;
}

/* 装饰性背景 */
.modal-decorative-bg {
    position: absolute;
    top: -20px; left: -20px; right: -20px; bottom: -20px;
    z-index: -1;
    /* background: radial-gradient(circle at 50% 0, #ffdce5 0%, transparent 60%); */
    border-radius: 40px;
    opacity: 0.5;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  position: relative;
}
.modal-header h3 { 
    margin: 0; 
    font-size: 20px; 
    color: var(--primary); 
    font-weight: 900;
    text-shadow: 1px 1px 0 #fff;
    flex: 1;
    text-align: center;
    margin-left: 32px; /* 平衡关闭按钮的位置 */
}
.close-btn { 
  background: #FFF; 
  border: 2px solid #FFE6EE; 
  width: 32px; height: 32px; 
  border-radius: 50%; 
  color: #FB7299; 
  font-weight: bold;
  cursor: pointer;
  transition: all 0.2s;
  display: flex; align-items: center; justify-content: center;
}
.close-btn:hover { background: #FB7299; color: #FFF; border-color: #FB7299; }

.modal-body {
    text-align: left;
    margin-bottom: 24px;
}

.input-label { 
  font-size: 14px; 
  color: #888; 
  margin-bottom: 8px; 
  font-weight: bold; 
  padding-left: 8px;
}

/* 输入框容器 */
.input-wrapper {
  position: relative;
  z-index: 10;
}

.input-icon {
    position: absolute;
    left: 16px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 18px;
    z-index: 2;
    pointer-events: none;
}

.modal-input {
  width: 100%;
  box-sizing: border-box;
  padding: 14px 14px 14px 46px; /* 留出图标位置 */
  border: 2px solid #FFE6EE;
  border-radius: 50px; /* 全圆角 */
  font-size: 16px;
  background: #FFF;
  outline: none;
  color: #555;
  transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
  box-shadow: inset 0 2px 4px rgba(0,0,0,0.03);
}
.modal-input:focus { 
  border-color: var(--primary); 
  background: #FFF; 
  box-shadow: 0 0 0 6px rgba(251, 114, 153, 0.15);
}
.modal-input::placeholder { 
  color: #ccc; font-size: 14px; 
}

/* 搜索下拉 */
.search-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: 12px;
  background: #FFF;
  border-radius: 20px;
  box-shadow: 0 10px 40px rgba(100, 100, 110, 0.15);
  border: 1px solid #FFE6EE;
  max-height: 240px;
  overflow-y: auto;
  z-index: 100;
  padding: 8px;
}

/* 滚动条美化 */
.search-dropdown::-webkit-scrollbar { width: 4px; }
.search-dropdown::-webkit-scrollbar-thumb { background: #FFD1DC; border-radius: 4px; }
.search-dropdown::-webkit-scrollbar-track { background: transparent; }

.search-dropdown ul {
  list-style: none;
  padding: 0;
  margin: 0;
}
.search-item {
  padding: 12px 16px;
  border-radius: 12px;
  cursor: pointer;
  transition: background 0.2s;
  margin-bottom: 4px;
}
.search-item:last-child { margin-bottom: 0; }
.search-item:hover {
  background: #FFF0F5;
}
.item-name {
  font-size: 15px;
  font-weight: bold;
  color: var(--primary);
  margin-bottom: 4px;
}
.item-desc {
  font-size: 12px;
  color: #999;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.loading-indicator {
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 16px;
  pointer-events: none;
}

.modal-tips { 
    font-size: 12px; 
    color: #BBB; 
    margin-top: 12px; 
    padding-left: 8px;
}

.modal-footer {
  display: flex;
  gap: 16px;
  margin-top: auto;
}
.btn-cancel, .btn-confirm {
  flex: 1;
  padding: 12px;
  border-radius: 50px;
  border: none;
  font-size: 15px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-cancel { 
    background: #F2F3F5; 
    color: #999; 
}
.btn-cancel:hover { background: #E5E6EB; }

.btn-confirm { 
   background: linear-gradient(90deg, #FB7299, #FF5C8A);
   color: #FFF; 
   box-shadow: 0 4px 12px rgba(251, 114, 153, 0.4);
}
.btn-confirm:hover { 
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(251, 114, 153, 0.5);
}
.btn-confirm:disabled { 
    background: #E0E0E0; 
    box-shadow: none; 
    color: #AAA; 
    transform: none;
    cursor: not-allowed;
}

/* 弹窗动画 */
.pop-enter-active, .pop-leave-active { transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1); }
.pop-enter-from, .pop-leave-to { transform: scale(0.9) translateY(20px); opacity: 0; }

.fade-enter-active, .fade-leave-active { transition: opacity 0.2s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

/* 桌面端适配 */
@media (min-width: 768px) {
  .app-container {
    background-color: #FFE6EE;
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
