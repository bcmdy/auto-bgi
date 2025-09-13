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
                 
                    <span v-if="member.is_simulated">
                        (小号)
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
        const response = await apiMethods.StartOnline(typeKey || 'all', isDebugMode.value)
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
/* 左侧开关样式 */
.switch-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 16px;
}

.switch-label {
  font-size: 18px;
  font-weight: 500;
}

.switch {
  position: relative;
  width: 100px;
  height: 36px;
  background-color: #ccc;
  border-radius: 20px;
  cursor: pointer;
  transition: background-color 0.25s;
  display: flex;
  align-items: center;
  padding: 0 4px;
}

.switch-on {
  background-color: #409eff;
}

.switch-handle {
  width: 32px;
  height: 32px;
  background: #fff;
  border-radius: 50%;
  transition: transform 0.25s;
}

.switch-on .switch-handle {
  transform: translateX(65px);
}

.switch-text {
  position: absolute;
  width: 100%;
  text-align: center;
  font-weight: 600;
  color: #fff;
  pointer-events: none;
  font-size: 14px;
}

/* 其余样式保持原来的按钮和详情面板 */
.member-name {
  display: inline-block;
  background: #f0f5ff;
  color: #409eff;
  padding: 2px 8px;
  margin: 0 4px 4px 0;
  border-radius: 12px;
  font-size: 14px;
}

.online-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #f0f4ff 0%, #fdfbff 100%);
  padding: 40px 0;
}
.header { text-align: center; margin-bottom: 32px; font-family: 'Segoe UI', sans-serif; }
.main-content { display: flex; justify-content: center; gap: 32px; flex-wrap: wrap; align-items: flex-start; }
.card { width: 360px; background: #fff; border-radius: 18px; box-shadow: 0 6px 20px rgba(64,158,255,0.12); padding: 32px 24px; display: flex; flex-direction: column; align-items: center; }
button { padding: 14px 0; background: linear-gradient(90deg, #409eff 0%, #66b1ff 100%); color: #fff; border: none; border-radius: 8px; cursor: pointer; font-size: 16px; font-weight: 500; box-shadow: 0 3px 12px rgba(64,158,255,0.1); transition: all 0.25s ease; width: 100%; }
button:hover { transform: translateY(-2px); background: linear-gradient(90deg, #66b1ff 0%, #409eff 100%); }
.refresh-wrap { margin-top: 16px; width: 100%; }
.refresh-wrap button { background: linear-gradient(90deg, #67c23a 0%, #85d13a 100%); }
.refresh-wrap button:hover { background: linear-gradient(90deg, #85d13a 0%, #67c23a 100%); }

.details-wrap { display: grid; grid-template-columns: repeat(2, 1fr); gap: 24px; align-items: stretch; min-width: 360px; }
.detail-panel { background: #fff; border-radius: 18px; box-shadow: 0 6px 20px rgba(64,158,255,0.12); padding: 28px 22px; display: flex; flex-direction: column; min-width: 260px; }
.detail-panel h2 { font-size: 18px; font-weight: 600; color: #409eff; margin-bottom: 16px; text-align: center; }
.detail-content { display: flex; flex-direction: column; gap: 12px; }
.status-row { display: flex; justify-content: space-between; align-items: center; font-size: 15px; }
.label { color: #888; font-weight: 500; }
.status { font-weight: bold; padding: 4px 12px; border-radius: 16px; font-size: 15px; text-align: center; }
.status.online { background: #e6f7ff; color: #409eff; border: 1px solid #b3e0ff; }
.status.offline { background: #f5f5f5; color: #bbb; border: 1px solid #eee; }

@media (max-width: 1200px) { .main-content { flex-direction: column; align-items: center; gap: 20px; } .card, .detail-panel { width: 95%; } .details-wrap { grid-template-columns: 1fr; } }
@media (max-width: 600px) { .card { padding: 20px 10px; } .detail-panel { padding: 16px 10px; } }
</style>
