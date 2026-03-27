<template>
  <div class="anime-container">
    <div class="layout-wrapper">
      
      <aside class="control-panel">
        <div class="panel-header">
          <h1>🌸 联机管理</h1>
          <p class="subtitle">Online Center</p>
        </div>

        <div class="control-group switch-group">
          <div class="switch-label">
            <span>🔧 调机模式</span>
            <span class="status-text" :class="{ active: isDebugMode }">
              {{ isDebugMode ? 'ON' : 'OFF' }}
            </span>
          </div>
          <div class="sakura-switch" :class="{ active: isDebugMode }" @click="isDebugMode = !isDebugMode">
            <div class="switch-handle"></div>
          </div>
        </div>

        <div class="control-group status-group">
          <div class="switch-label">
            <span>📡 在线状态</span>
          </div>
          <div class="status-badge" @click="fetchOnlineStatus" :title="statusLoading ? '正在刷新' : '点击刷新'">
            <span v-if="statusLoading">刷新中...</span>
            <span class="badge-online">{{ onlineStatus }}</span>  
            <!-- <span v-else-if="onlineStatus === true" class="badge-online">在线</span>
            <span v-else-if="onlineStatus === false" class="badge-offline">离线</span>
            <span v-else class="badge-unknown">未知</span> -->
          </div>
        </div>

        <div class="launch-count-panel">
          <div class="count-header">
            <span class="icon">🚀</span>
            <span class="label">上线次数</span>
          </div>
          <div class="count-display">{{ launchCount }}</div>
          <button class="clear-btn" @click="clearLaunchCount">
            <span>🧽</span> 清零
          </button>
        </div>

        <div class="action-buttons">
          <button class="anime-btn btn-online" @click="StartOnline(null)">
            <span class="icon">🐶</span> 
            <span>一键上线</span>
          </button>
          
          <button class="anime-btn btn-refresh" @click="refreshAll">
            <span class="icon">🔄</span> 
            <span>刷新详情</span>
          </button>
          
          <button class="anime-btn btn-report" @click="openReportBomb">
            <span class="icon">🧨</span> 
            <span>举报炸弹</span>
          </button>
          
          <button class="anime-btn btn-offline" @click="offline(null)">
            <span class="icon">💤</span> 
            <span>一键下线</span>
          </button>
          
          <button class="anime-btn btn-create" @click="openCreateRoom">
            <span class="icon">🧩</span> 
            <span>创建房间</span>
          </button>
          
          <button class="anime-btn btn-home" @click="goHome">
            <span class="icon">🏠</span> 
            <span>返回主页</span>
          </button>
        </div>
      </aside>

      <main class="content-area">
        <div v-if="detailList.length === 0" class="empty-state">
          <div class="empty-icon">🍃</div>
          <p>暂无房间数据，请点击刷新...</p>
        </div>

        <div class="room-grid" v-else>
          <div 
            v-for="(item, index) in detailList" 
            :key="item.key || index" 
            class="room-card"
          >
            <div v-if="item.is_homeowner" class="room-ribbon">我的房间</div>
            <div class="card-header">
              <h3 class="room-title">{{ item.title }}</h3>
              <div class="card-actions">
                <button
                  v-if="item.is_homeowner"
                  type="button"
                  class="room-edit-btn"
                  @click="openRoomEditor(item)"
                >
                  编辑
                </button>
                <span class="room-count" :class="{ 'has-people': item.count > 0 }">
                  {{ item.count }} 人在线
                </span>
              </div>
            </div>
            
            <p class="room-desc">{{ item.description || '暂无描述' }}</p>

            <div class="divider"></div>

            <div class="member-area">
              <div v-if="item.members && item.members.length > 0" class="member-list">
                <div 
                  v-for="(member, mIndex) in item.members" 
                  :key="mIndex" 
                  class="member-pill"
                >
                  <span class="avatar">👤</span>
                  <span class="name">{{ member.name }}</span>
                  <span class="status-tag" :class="member.abgi_type === 'debug' ? 'tag-debug' : 'tag-run'">
                    {{ member.abgi_type === 'noDebug' ? '正常跑' : (member.abgi_type === 'debug' ? '调试中' : member.abgi_type) }}
                  </span>
                </div>
              </div>
              <div v-else class="no-member">
                (｡•́︿•̀｡) 暂无人员
              </div>
            </div>
          </div>
        </div>
      </main>
      
      <a-modal
        v-model:open="reportModal.open"
        title="举报炸弹"
        :confirm-loading="reportModal.loading"
        :width="isMobile ? '95vw' : 520"
        centered
        @ok="handleReportOk"
        @cancel="handleReportCancel"
        ok-text="提交"
        cancel-text="取消"
        class="anime-modal"
      >
        <div style="display:flex; flex-direction: column; gap:12px; padding-top: 8px;">
          <div>
            <div style="font-weight:700; margin-bottom:6px;">炸弹人</div>
            <a-input v-model:value="reportModal.BombName" placeholder="炸弹人" />
          </div>
          <div>
            <div style="font-weight:700; margin-bottom:6px;">行为</div>
            <a-input v-model:value="reportModal.BombAction" placeholder="行为" />
          </div>
        </div>
      </a-modal>

      <a-modal
        v-model:open="roomEditModal.open"
        title="编辑房间"
        :confirm-loading="roomEditModal.loading"
        :width="isMobile ? '95vw' : 520"
        centered
        :ok-button-props="{ disabled: roomEditModal.queryLoading }"
        @ok="handleRoomEditOk"
        @cancel="handleRoomEditCancel"
        ok-text="确认"
        cancel-text="取消"
        class="anime-modal"
      >
        <a-spin :spinning="roomEditModal.queryLoading">
          <div style="display:flex; flex-direction: column; gap:12px; padding-top: 8px;">
            <div>
              <div style="font-weight:700; margin-bottom:6px;">房间名</div>
              <a-input v-model:value="roomEditModal.group_name" disabled />
            </div>
            <div>
              <div style="font-weight:700; margin-bottom:6px;">描述</div>
              <a-textarea
                v-model:value="roomEditModal.description"
                :auto-size="{ minRows: 2, maxRows: 6 }"
                placeholder="房间描述"
              />
            </div>
            <div style="display:flex; gap: 16px; flex-wrap: wrap;">
              <div>
                <div style="font-weight:700; margin-bottom:6px;">是否启用</div>
                <a-switch v-model:checked="roomEditModal.is_active" />
              </div>
              <div>
                <div style="font-weight:700; margin-bottom:6px;">是否开放</div>
                <a-switch v-model:checked="roomEditModal.is_open" />
              </div>
            </div>
          </div>
        </a-spin>
      </a-modal>

      <a-modal
        v-model:open="createRoomModal.open"
        title="创建房间"
        :confirm-loading="createRoomModal.loading"
        :width="isMobile ? '95vw' : 520"
        centered
        @ok="handleCreateRoomOk"
        @cancel="handleCreateRoomCancel"
        ok-text="确认"
        cancel-text="取消"
        class="anime-modal"
      >
        <div style="display:flex; flex-direction: column; gap:12px; padding-top: 8px;">
          <div>
            <div style="font-weight:700; margin-bottom:6px;">房间名</div>
            <a-input v-model:value="createRoomModal.name" placeholder="房间名" allow-clear />
          </div>
          <div>
            <div style="font-weight:700; margin-bottom:6px;">描述</div>
            <a-textarea
              v-model:value="createRoomModal.description"
              :auto-size="{ minRows: 2, maxRows: 6 }"
              placeholder="房间描述"
              allow-clear
            />
          </div>
        </div>
      </a-modal>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import api, { apiMethods } from '@/utils/api'
import { useRouter } from 'vue-router'

const isDebugMode = ref(false)
const detailList = ref([])
const router = useRouter()
const launchCount = ref(0)
const onlineStatus = ref(null) // null = unknown, true = online, false = offline
const statusLoading = ref(false)
const isMobile = ref(window.innerWidth <= 480)

const reportModal = reactive({
  open: false,
  loading: false,
  BombName: '',
  BombAction: ''
})

const roomEditModal = reactive({
  open: false,
  loading: false,
  queryLoading: false,
  group_name: '',
  description: '',
  is_active: false,
  is_open: false
})

const createRoomModal = reactive({
  open: false,
  loading: false,
  name: '',
  description: ''
})

const openReportBomb = () => {
  reportModal.BombName = ''
  reportModal.BombAction = ''
  reportModal.open = true
}

const handleReportCancel = () => {
  reportModal.open = false
}

const handleReportOk = async () => {
  if (!reportModal.BombName || !reportModal.BombAction) {
    message.warning('请填写完整信息')
    return
  }
  reportModal.loading = true
  try {
    const res = await apiMethods.reportBomb({
      BombName: reportModal.BombName,
      BombAction: reportModal.BombAction
    })
    
    // 优先显示后端返回的 message (兼容 200 和被拦截器处理的 500)
    if (res && res.message) {
      Modal.info({
        title: '举报结果',
        content: res.message,
        okText: '确定',
        width: isMobile.value ? 360 : 520,
        centered: true,
        class: 'anime-modal'
      })
    } else {
      message.success('举报成功')
    }

    reportModal.open = false
  } catch (e) {
    // 处理未被拦截器转换为数据的错误
    const errorMsg = e.response?.data?.message || e.message || '举报失败'
    message.error(errorMsg)
  } finally {
    reportModal.loading = false
  }
}

const getErrorMessage = (e, fallback) => {
  if (!e) return fallback
  if (typeof e === 'string') return e
  return e.response?.data?.message || e.data?.message || e.message || e.msg || fallback
}

const openRoomEditor = async (room) => {
  roomEditModal.group_name = room?.title || room?.key || ''
  roomEditModal.description = ''
  roomEditModal.is_active = false
  roomEditModal.is_open = false
  roomEditModal.open = true
  roomEditModal.queryLoading = true

  try {
    const res = await api.post('/api/abgiSSE/HomeownerQueryGroup', {
      group_name: roomEditModal.group_name
    })
    const data = res?.message
    if (!data) {
      message.error('查询房间信息失败')
      return
    }
    roomEditModal.group_name = data.group_name || roomEditModal.group_name
    roomEditModal.description = data.description || ''
    roomEditModal.is_active = data.is_active === true || data.is_active === 1 || data.is_active === '1' || data.is_active === 'true'
    roomEditModal.is_open = data.is_open === true || data.is_open === 1 || data.is_open === '1' || data.is_open === 'true'
  } catch (e) {
    message.error(getErrorMessage(e, '查询房间信息失败'))
  } finally {
    roomEditModal.queryLoading = false
  }
}

const handleRoomEditCancel = () => {
  roomEditModal.open = false
}

const handleRoomEditOk = async () => {
  if (!roomEditModal.group_name) {
    message.warning('房间名不能为空')
    return
  }

  roomEditModal.loading = true
  try {
    const res = await api.post('/api/abgiSSE/HomeownerUpdateGroup', {
      group_name: roomEditModal.group_name,
      description: roomEditModal.description,
      is_active: roomEditModal.is_active,
      is_open: roomEditModal.is_open
    })
    message.success(res?.message || '房间信息更新成功')
    roomEditModal.open = false
    await fetchOnlineDetail()
  } catch (e) {
    message.error(getErrorMessage(e, '更新房间信息失败'))
  } finally {
    roomEditModal.loading = false
  }
}

const openCreateRoom = () => {
  createRoomModal.name = ''
  createRoomModal.description = ''
  createRoomModal.open = true
}

const handleCreateRoomCancel = () => {
  createRoomModal.open = false
}

const handleCreateRoomOk = async () => {
  const name = (createRoomModal.name || '').trim()
  if (!name) {
    message.warning('请填写房间名')
    return
  }

  createRoomModal.loading = true
  try {
    const res = await api.post('/api/abgiSSE/HomeownerCreateGroup', {
      name,
      description: createRoomModal.description || ''
    })
    message.success(res?.message || '创建房间成功')
    createRoomModal.open = false
    await fetchOnlineDetail()
  } catch (e) {
    message.error(getErrorMessage(e, '创建房间失败'))
  } finally {
    createRoomModal.loading = false
  }
}


// /**
//  * 检测是否为 WebView 环境
//  * @returns {boolean} true 表示是 WebView
//  */
// const isWebView = () => {
//   const ua = navigator.userAgent.toLowerCase()
  
//   // 1. 微信、QQ 等常见 APP 内核
//   if (ua.match(/micromessenger|qq\/|weibo/i)) {
//     return true
//   }

//   // 2. Android WebView 特征 (通常包含 'wv' 或 'version/x.x')
//   // 很多安卓内置浏览器 UserAgent 会包含 Version/4.0 这种标识，而 Chrome 浏览器通常不会
//   const isAndroid = ua.indexOf('android') > -1
//   if (isAndroid && (ua.indexOf('wv') > -1 || ua.indexOf('version/') > -1)) {
//     return true
//   }
//   return false
// }


const fetchOnlineDetail = async () => {
  try {
    const res = await api.get('/api/abgiSSE/getOnlineUser')
    detailList.value = res.map(item => ({
      key: item.group_name,
      title: item.group_name,
      description: item.description,
      count: item.count,
      members: Array.isArray(item.members) ? item.members : [],
      is_homeowner: item.is_homeowner === true || item.is_homeowner === 1 || item.is_homeowner === '1' || item.is_homeowner === 'true',
      status: item.count > 0,
      time: ''
    }))
  } catch (e) {
    message.error('获取联机详情失败')
  }
}

const fetchOnlineStatus = async () => {
  statusLoading.value = "未知"
  try {
    const res = await api.get('/api/abgiSSE/getOnlineStatus')
    // 接口返回 true/false；确保布尔值
    onlineStatus.value = res
  } catch (e) {
    console.error('获取在线状态失败', e)
    message.error('获取在线状态失败')
    onlineStatus.value = null
  } finally {
    statusLoading.value = false
  }
}

const fetchLaunchCount = async () => {
  try {
    const res = await apiMethods.getNumberOfLaunches()
    launchCount.value = res.number || 0
  } catch (e) {
    console.error('获取上线次数失败', e)
  }
}

// 刷新所有：详情 + 在线状态
const refreshAll = async () => {
  await Promise.all([fetchOnlineDetail(), fetchOnlineStatus()])
}

const clearLaunchCount = async () => {
  Modal.confirm({
    title: '确认清零？',
    content: '确定要清空上线次数吗？',
    okText: '确定',
    cancelText: '取消',
    centered: true,
    class: 'anime-modal',
    async onOk() {
      try {
        await apiMethods.clearNumberOfLaunches()
        Modal.destroyAll()
        message.success('清零成功')
        fetchLaunchCount()
      } catch (e) {
        message.error(e.message || '清零失败')
      }
    }
  })
}

const offline = (typeKey) => {
  Modal.confirm({
    title: '确认下线吗？',
    content: typeKey ? `下线【${typeKey}】？` : '确认全部下线？',
    okText: '确定',
    cancelText: '取消',
    centered: true,
    class: 'anime-modal',
    async onOk() {
      try {
        await apiMethods.offline(typeKey || 'all')
        Modal.destroyAll()
        Modal.info({ title: '下线结果', content: '下线成功', okText: '关闭', centered: true })
        await fetchOnlineDetail()
        // 下线后同时刷新在线状态
        await fetchOnlineStatus()
      } catch (e) {
        message.error(e.message || '操作失败')
      }
    }
  })
}

const StartOnline = (typeKey) => {
  Modal.confirm({
    title: '确认上线吗？',
    content: typeKey ? `上线【${typeKey}】？` : '确认一键上线？',
    okText: '确定',
    cancelText: '取消',
    centered: true,
    class: 'anime-modal',
    async onOk() {
      try {
        const response = await apiMethods.StartOnline(typeKey || 'noDebug', isDebugMode.value)
        console.log(response)
        Modal.destroyAll()
        Modal.info({ title: '上线结果', content: response, okText: '关闭', centered: true })
        await fetchOnlineDetail()
        await fetchLaunchCount() // 上线成功后刷新上线次数
        // 上线后同时刷新在线状态
        await fetchOnlineStatus()
      } catch (e) {
        console.log("=====", e)
        const errorMsg = e.response && e.response.data ? e.response.data : '上线失败';
        message.error(errorMsg)
      }
    }
  })
}

const goHome = () => {
  router.push('/')
}

onMounted(() => {
  const handleResize = () => {
    isMobile.value = window.innerWidth <= 480
  }
  window.addEventListener('resize', handleResize)
  
  // // === 在这里进行拦截 ===
  // if (isWebView()) {
  //   // 暴力替换整个页面内容
  //   document.body.innerHTML = `
  //     <div style="
  //       display: flex;
  //       flex-direction: column;
  //       justify-content: center;
  //       align-items: center;
  //       height: 100vh;
  //       background: #fff0f5;
  //       font-family: sans-serif;
  //       color: #555;
  //       text-align: center;
  //     ">
  //       <div style="font-size: 60px; margin-bottom: 20px;">🚫</div>
  //       <h2 style="color: #ff69b4;">非法访问</h2>
  //       <p>为了安全与体验，请点击右上角选择<br/><b>"在浏览器打开"</b> (Chrome / Safari)</p>
  //     </div>
  //   `
  //   // 阻止后续逻辑执行
  //   return
  // }
  
  // 正常环境则加载数据
  fetchOnlineDetail()
  fetchLaunchCount()
  // 页面进入时请求一次在线状态
  fetchOnlineStatus()
})

onUnmounted(() => {
  const handleResize = () => {}
  window.removeEventListener('resize', handleResize)
})

</script>

<style scoped>
@import '../assets/fonts3.css';

/* 全局容器：樱花背景 */
.anime-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #fff0f5 0%, #e6f7ff 100%);
  background-image: 
    radial-gradient(#ffc0cb 15%, transparent 16%),
    radial-gradient(#87ceeb 15%, transparent 16%);
  background-size: 30px 30px;
  background-position: 0 0, 15px 15px;
  font-family: 'Nunito', 'Fredoka', 'Microsoft YaHei', sans-serif;
  color: #555;
  padding: 20px;
  box-sizing: border-box;
}

/* 布局包装 */
.layout-wrapper {
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  gap: 30px;
  align-items: flex-start;
}

/* === 左侧控制面板 === */
.control-panel {
  width: 300px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  border-radius: 24px;
  padding: 30px 20px;
  box-shadow: 0 8px 32px rgba(255, 182, 193, 0.3);
  border: 2px solid #fff;
  position: sticky;
  top: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.panel-header {
  text-align: center;
  margin-bottom: 10px;
}

.panel-header h1 {
  font-size: 26px;
  color: #ff69b4;
  margin: 0;
  font-weight: 800;
  letter-spacing: 1px;
}

.subtitle {
  font-size: 14px;
  color: #aab;
  margin: 5px 0 0;
}

/* 开关组 */
.switch-group {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fdfdfd;
  padding: 15px;
  border-radius: 16px;
  box-shadow: inset 0 2px 6px rgba(0,0,0,0.04);
}

.status-group {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff8f9;
  padding: 12px;
  border-radius: 12px;
  box-shadow: inset 0 1px 4px rgba(0,0,0,0.03);
}

.status-text-small { font-size: 12px; color: #999; margin-top: 2px; }

.status-badge {
  min-width: 86px;
  text-align: center;
  padding: 6px 10px;
  border-radius: 12px;
  font-weight: 700;
  cursor: pointer;
  user-select: none;
}

.badge-online { background: #e6ffec; color: #237804; border: 1px solid #b7eb8f; }
.badge-offline { background: #fff1f0; color: #cf1322; border: 1px solid #ffa39e; }
.badge-unknown { background: #fffbe6; color: #614700; border: 1px solid #ffe58f; }

.switch-label {
  display: flex;
  flex-direction: column;
  font-weight: 700;
  color: #666;
}

.status-text {
  font-size: 12px;
  color: #ccc;
  margin-top: 2px;
}
.status-text.active { color: #ff69b4; }

/* 自定义樱花开关 */
.sakura-switch {
  width: 56px;
  height: 28px;
  background: #eee;
  border-radius: 14px;
  position: relative;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0.0, 0.2, 1);
}

.sakura-switch.active {
  background: #ff9ebb;
  box-shadow: 0 0 10px rgba(255, 158, 187, 0.5);
}

.switch-handle {
  width: 22px;
  height: 22px;
  background: #fff;
  border-radius: 50%;
  position: absolute;
  top: 3px;
  left: 3px;
  transition: transform 0.3s cubic-bezier(0.4, 0.0, 0.2, 1);
  box-shadow: 0 2px 4px rgba(0,0,0,0.2);
}

.sakura-switch.active .switch-handle {
  transform: translateX(28px);
}

/* 上线次数面板 */
.launch-count-panel {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  padding: 20px;
  color: #fff;
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.3);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.count-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  opacity: 0.95;
}

.count-header .icon {
  font-size: 18px;
}

.count-display {
  font-size: 48px;
  font-weight: 800;
  text-align: center;
  letter-spacing: 2px;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  padding: 10px 0;
}

.clear-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: #fff;
  padding: 10px;
  border-radius: 12px;
  font-weight: 600;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.clear-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
}

.clear-btn:active {
  transform: scale(0.95);
}

/* 按钮组 */
.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.anime-btn {
  width: 100%;
  border: none;
  padding: 14px;
  border-radius: 18px;
  font-weight: 700;
  font-size: 16px;
  color: #130202;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  transition: all 0.2s;
  box-shadow: 0 4px 15px rgba(0,0,0,0.1);
  position: relative;
  overflow: hidden;
}

.anime-btn:active {
  transform: scale(0.96);
}

.anime-btn:hover {
  transform: translateY(-3px);
  filter: brightness(1.05);
}

.btn-online {
  background: linear-gradient(135deg, #ff9a9e 0%, #fecfef 99%, #fecfef 100%);
  box-shadow: 0 4px 15px rgba(255, 154, 158, 0.4);
}

.btn-refresh {
  background: linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%);
  box-shadow: 0 4px 15px rgba(161, 140, 209, 0.4);
}

.btn-offline {
  background: linear-gradient(135deg, #84fab0 0%, #8fd3f4 100%);
  box-shadow: 0 4px 15px rgba(132, 250, 176, 0.4);
}

.btn-create {
  background: linear-gradient(135deg, #ffe259 0%, #ffa751 100%);
  box-shadow: 0 4px 15px rgba(255, 167, 81, 0.4);
}

.btn-report {
  background: linear-gradient(135deg, #ff7e5f 0%, #feb47b 100%);
  box-shadow: 0 4px 15px rgba(255, 126, 95, 0.4);
}

.btn-home {
  background: #fff;
  color: #888;
  border: 2px solid #eee;
  box-shadow: none;
}
.btn-home:hover {
  background: #f8f8f8;
  border-color: #ddd;
}

.anime-modal :deep(.ant-modal-content) {
  border-radius: 20px;
  border: 3px solid #ffcce6;
  background: #fff0f5;
  box-shadow: 0 12px 32px rgba(255, 182, 193, 0.35);
}
.anime-modal :deep(.ant-modal-header) {
  background: transparent;
  border-bottom: 2px dashed #ffb6c1;
}
.anime-modal :deep(.ant-modal-title) {
  color: #ff3385;
  text-align: center;
  font-weight: 800;
}
.anime-modal :deep(.ant-modal-footer) {
  display: flex;
  justify-content: center;
  gap: 12px;
}
.anime-modal :deep(.ant-btn) {
  border-radius: 12px;
  font-weight: 700;
}
.anime-modal :deep(.ant-btn-primary) {
  background: linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%);
  border-color: #ff85ad;
  color: #fff;
}
.anime-modal :deep(.ant-input) {
  border-radius: 12px;
  border: 2px solid #ffd6e7;
  background: #fff;
}

/* === 右侧详情内容 === */
.content-area {
  flex: 1;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px;
  background: rgba(255,255,255,0.6);
  border-radius: 20px;
  color: #aaa;
}
.empty-icon { font-size: 48px; margin-bottom: 10px; opacity: 0.5; }

.room-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 24px;
}

/* 房间卡片 */
.room-card {
  background: #ffffff;
  border-radius: 24px;
  padding: 24px;
  border: 1px solid #fff;
  box-shadow: 0 6px 20px rgba(176, 196, 222, 0.2);
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.room-ribbon {
  position: absolute;
  top: 14px;
  right: -46px;
  transform: rotate(45deg);
  transform-origin: center;
  background: linear-gradient(135deg, #ff6fb2 0%, #ffb6d5 100%);
  color: #fff;
  font-weight: 800;
  font-size: 12px;
  padding: 6px 56px;
  letter-spacing: 1px;
  box-shadow: 0 6px 14px rgba(255, 105, 180, 0.35);
  z-index: 2;
  user-select: none;
  pointer-events: none;
}

.room-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 12px 30px rgba(255, 182, 193, 0.35);
  border-color: #ffe6ea;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.card-actions {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.room-edit-btn {
  border: 1px solid #ffd6e7;
  background: #fff0f5;
  color: #ff3385;
  font-weight: 800;
  border-radius: 10px;
  padding: 4px 10px;
  cursor: pointer;
  line-height: 1;
  transition: all 0.2s ease;
}

.room-edit-btn:hover {
  border-color: #ff9ecf;
  background: #ffe6f1;
}

.room-title {
  margin: 0;
  font-size: 18px;
  color: #444;
  font-weight: 700;
}

.room-count {
  font-size: 12px;
  background: #eee;
  padding: 4px 10px;
  border-radius: 10px;
  color: #999;
}
.room-count.has-people {
  background: #e6f7ff;
  color: #1890ff;
  font-weight: 600;
}

.room-desc {
  font-size: 13px;
  color: #888;
  line-height: 1.5;
  margin-bottom: 15px;
}

.divider {
  height: 1px;
  background: repeating-linear-gradient(to right, #eee 0, #eee 5px, transparent 5px, transparent 10px);
  margin-bottom: 15px;
}

/* 成员列表 */
.member-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.member-pill {
  display: flex;
  align-items: center;
  background: #f9f9f9;
  border-radius: 12px;
  padding: 8px 12px;
  transition: background 0.2s;
}

.member-pill:hover {
  background: #fff0f5; /* 浅粉色 hover */
}

.avatar { margin-right: 8px; font-size: 14px; }
.name { flex: 1; font-size: 14px; color: #555; font-weight: 600; }

.status-tag {
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 8px;
}

.tag-run {
  background: #e6ffec;
  color: #52c41a;
  border: 1px solid #b7eb8f;
}

.tag-debug {
  background: #fffbe6;
  color: #faad14;
  border: 1px solid #ffe58f;
}

.no-member {
  text-align: center;
  color: #ccc;
  font-size: 13px;
  padding: 10px 0;
}

/* === 移动端适配 === */
@media (max-width: 900px) {
  .layout-wrapper {
    flex-direction: column;
    align-items: stretch;
  }
  
  .control-panel {
    width: 100%;
    position: relative;
    top: 0;
    margin-bottom: 20px;
    padding: 20px;
    box-sizing: border-box;
  }
  
  .action-buttons {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  .anime-btn {
    font-size: 14px;
    padding: 10px;
  }
}

@media (max-width: 480px) {
  .anime-container {
    padding: 10px;
  }
  
  .panel-header h1 {
    font-size: 22px;
  }
  
  /* 手机端按钮两行排列 */
  .action-buttons {
    grid-template-columns: 1fr 1fr;
  }
  
  .room-grid {
    grid-template-columns: 1fr; /* 手机端单列 */
  }
}
</style>

<style>
.anime-modal .ant-modal-content {
  border-radius: 20px;
  border: 3px solid #ffcce6;
  background: #fff0f5;
  box-shadow: 0 12px 32px rgba(255, 182, 193, 0.35);
}
.anime-modal .ant-modal-header {
  background: transparent;
  border-bottom: 2px dashed #ffb6c1;
}
.anime-modal .ant-modal-title {
  color: #ff3385;
  text-align: center;
  font-weight: 800;
}
.anime-modal .ant-modal-footer {
  display: flex;
  justify-content: center;
  gap: 12px;
}
.anime-modal .ant-btn {
  border-radius: 12px;
  font-weight: 700;
}
.anime-modal .ant-btn-primary {
  background: linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%);
  border-color: #ff85ad;
  color: #fff;
}
.anime-modal .ant-input {
  border-radius: 12px;
  border: 2px solid #ffd6e7;
  background: #fff;
}
</style>
