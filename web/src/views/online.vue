<template>
  <div class="online-container">
    <div class="header">
      <h1>在线联机管理</h1>
    </div>

    <div class="main-content">
      <!-- 左侧操作卡片 -->
      <div class="card">
        <!-- 是否调机模式选择 -->
        <div class="refresh-wrap switch-wrap">
          <span class="switch-label">调机模式</span>
          <div class="switch" :class="{ 'switch-on': isDebugMode }" @click="isDebugMode = !isDebugMode">
            <div class="switch-handle"></div>
            <span class="switch-text">{{ isDebugMode ? '开' : '关' }}</span>
          </div>
        </div>

        <!-- 上线（通用按钮） -->
        <div class="refresh-wrap">
          <button @click="StartOnline()">🐶 上线</button>
        </div>

        <!-- 刷新按钮 -->
        <div class="refresh-wrap">
          <button @click="fetchOnlineDetail">🔄 刷新详情</button>
        </div>

        <!-- 下线（通用按钮） -->
        <div class="refresh-wrap">
          <button @click="offline()">下线</button>
        </div>

        <!-- 返回主页 -->
        <div class="refresh-wrap">
          <button @click="goHome">返回主页</button>
        </div>
      </div>

      <!-- 右侧详情面板 -->
      <div class="details-wrap">
        <div class="detail-panel" v-for="item in detailList" :key="item.key">
          <h2>{{ item.title }}</h2>
          <div class="detail-content">
            <div class="status-row" v-if="item.members && item.members.length > 0">
              <span class="label">在线人员：</span>
              <span>
                <span v-for="member in item.members" :key="member.name" class="member-name">
                  {{ member.name }}
                 
                    <span >
                        ({{ member.abgi_type === 'noDebug' ? '正常跑' : member.abgi_type === 'debug' ? '调试':member.abgi_type}})
              
                    </span>
                 
                </span>
              
              </span>
            </div>
            <div v-else class="status-row">
              <span class="label">在线人员：</span>
              <span class="status offline">暂无</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue' // 保留消息弹窗
import api, { apiMethods } from '@/utils/api'
import { useRouter } from 'vue-router'

const isDebugMode = ref(false)
const detailList = ref([])
const router = useRouter()

const fetchOnlineDetail = async () => {
  try {
    const res = await api.get('/api/abgiSSE/getOnlineUser')
    detailList.value = res.map(item => ({
      key: item.group_name,
      title: item.group_name,
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
    content: typeKey ? `下线【${typeKey}】？` : '确认下线？',
    okText: '确定',
    cancelText: '取消',
    async onOk() {
      try {
        await apiMethods.offline(typeKey || 'all')
        Modal.destroyAll()
        Modal.info({ title: '下线结果', content: '下线成功', okText: '关闭' })
        fetchOnlineDetail()
      } catch (e) {
        message.error(e)
      }
    }
  })
}

const StartOnline = (typeKey) => {
  Modal.confirm({
    title: '确认上线吗？',
    content: typeKey ? `上线【${typeKey}】？` : '确认上线？',
    okText: '确定',
    cancelText: '取消',
    async onOk() {
      try {
        const response = await apiMethods.StartOnline(typeKey || 'noDebug', isDebugMode.value)
        console.log(response)
        Modal.destroyAll()
        Modal.info({ title: '上线结果', content: response, okText: '关闭' })
        fetchOnlineDetail()
      } catch (e) {
        console.log("=========",e)
        message.error(e.response.data)
      }
    }
  })
}

const goHome = () => {
  router.push('/')
}

onMounted(() => fetchOnlineDetail())
</script>


<style scoped>
.online-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #ffe6f3 0%, #e8f0ff 100%);
  font-family: "Poppins", "Segoe UI", "Microsoft YaHei", sans-serif;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 50px 0;
}

/* ===== 标题区 ===== */
.header {
  text-align: center;
  margin-bottom: 40px;
}
.header h1 {
  font-size: 36px;
  font-weight: 800;
  color: #ff66a3;
  text-shadow: 0 3px 10px rgba(255, 102, 163, 0.3);
  letter-spacing: 1px;
}

/* ===== 主体布局 ===== */
.main-content {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  align-items: flex-start;
  gap: 40px;
  width: 90%;
  max-width: 1200px;
}

/* ===== 左侧控制卡片 ===== */
.card {
  width: 320px;
  background: #ffffff;
  border-radius: 24px;
  box-shadow: inset 0 2px 8px rgba(255,255,255,0.8),
              0 8px 25px rgba(255, 182, 193, 0.3);
  padding: 30px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  transition: all .3s;
}
.card:hover {
  transform: translateY(-4px);
  box-shadow: 0 10px 28px rgba(255, 105, 180, 0.25);
}

/* ===== 调机模式开关 ===== */
.switch-wrap {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 18px;
}
.switch-label {
  font-size: 18px;
  color: #444;
  font-weight: 600;
}
.switch {
  width: 80px;
  height: 36px;
  background: #eee;
  border-radius: 20px;
  position: relative;
  cursor: pointer;
  transition: 0.3s;
}
.switch-on {
  background: linear-gradient(90deg, #89f7fe, #66a6ff);
  box-shadow: 0 0 10px rgba(102,166,255,0.4);
}
.switch-handle {
  position: absolute;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: #fff;
  top: 3px;
  left: 4px;
  transition: all 0.3s;
}
.switch-on .switch-handle {
  transform: translateX(42px);
}
.switch-text {
  position: absolute;
  width: 100%;
  text-align: center;
  color: #555;
  font-weight: 600;
  top: 6px;
  font-size: 14px;
}
.switch-on .switch-text {
  color: #fff;
}

/* ===== 操作按钮 ===== */
.refresh-wrap {
  width: 100%;
  margin-top: 14px;
}
button {
  width: 100%;
  border: none;
  border-radius: 14px;
  font-size: 16px;
  font-weight: 700;
  color: white;
  padding: 12px 0;
  cursor: pointer;
  box-shadow: 0 4px 10px rgba(0,0,0,0.1);
  transition: 0.25s;
}
button:hover {
  transform: translateY(-2px) scale(1.03);
  box-shadow: 0 6px 14px rgba(0,0,0,0.15);
}
.refresh-wrap:nth-child(2) button {
  background: linear-gradient(135deg, #6ea8ff 0%, #409eff 100%);
}
.refresh-wrap:nth-child(3) button {
  background: linear-gradient(135deg, #5ee65e 0%, #36b64f 100%);
}
.refresh-wrap:nth-child(4) button {
  background: linear-gradient(135deg, #ff6b81 0%, #ff416c 100%);
}
.refresh-wrap:nth-child(5) button {
  background: linear-gradient(135deg, #ffb347 0%, #ffcc33 100%);
}

/* ===== 右侧详情卡片 ===== */
.details-wrap {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(270px, 1fr));
  gap: 30px;
  flex: 1;
}

.detail-panel {
  background: #ffffff;
  border-radius: 22px;
  box-shadow: 0 6px 20px rgba(102,166,255,0.12);
  padding: 26px 22px;
  transition: all .3s;
  border: 1px solid rgba(255,255,255,0.6);
}
.detail-panel:hover {
  transform: translateY(-4px);
  box-shadow: 0 10px 28px rgba(64,158,255,0.18);
}
.detail-panel h2 {
  font-size: 18px;
  font-weight: 700;
  color: #409eff;
  text-align: center;
  margin-bottom: 12px;
  border-bottom: 1px dashed #dce8ff;
  padding-bottom: 6px;
}
.detail-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.status-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 15px;
}
.label {
  color: #777;
  font-weight: 600;
}

/* ===== 在线人员标签 ===== */
.member-name {
  display: inline-block;
  background: #e8f0ff;
  color: #409eff;
  padding: 4px 10px;
  margin: 4px 5px 0 0;
  border-radius: 12px;
  font-size: 14px;
  transition: 0.2s;
}
.member-name:hover {
  background: #d2e3ff;
  transform: scale(1.05);
}

/* 暂无状态 */
.status.offline {
  background: #f5f5f5;
  color: #bbb;
  padding: 4px 10px;
  border-radius: 12px;
}

/* ===== 响应式优化 ===== */
@media (max-width: 1200px) {
  .main-content { flex-direction: column; align-items: center; }
  .card { width: 90%; }
  .details-wrap { width: 100%; grid-template-columns: 1fr 1fr; }
}
@media (max-width: 768px) {
  .details-wrap { grid-template-columns: 1fr; }
}
</style>

