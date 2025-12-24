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

        <div class="action-buttons">
          <button class="anime-btn btn-online" @click="StartOnline(null)">
            <span class="icon">🐶</span> 
            <span>一键上线</span>
          </button>
          
          <button class="anime-btn btn-refresh" @click="fetchOnlineDetail">
            <span class="icon">🔄</span> 
            <span>刷新详情</span>
          </button>
          
          <button class="anime-btn btn-offline" @click="offline(null)">
            <span class="icon">💤</span> 
            <span>一键下线</span>
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
            <div class="card-header">
              <h3 class="room-title">{{ item.title }}</h3>
              <span class="room-count" :class="{ 'has-people': item.count > 0 }">
                {{ item.count }} 人在线
              </span>
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
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import api, { apiMethods } from '@/utils/api'
import { useRouter } from 'vue-router'

const isDebugMode = ref(false)
const detailList = ref([])
const router = useRouter()

setInterval(() => {
  debugger
}, 100)


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
      status: item.count > 0,
      time: ''
    }))
  } catch (e) {
    message.error('获取联机详情失败')
  }
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
        fetchOnlineDetail()
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
        fetchOnlineDetail()
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
  //       <p>为了安全与体验，请点击右上角选择<br/><b>“在浏览器打开”</b> (Chrome / Safari)</p>
  //     </div>
  //   `
  //   // 阻止后续逻辑执行
  //   return
  // }
  
  // 正常环境则加载数据
  fetchOnlineDetail()
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