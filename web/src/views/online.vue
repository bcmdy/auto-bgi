<template>
  <div class="online-container">
    <div class="header">
      <h1>在线联机管理</h1>
    </div>

    <div class="main-content">
      <!-- 操作卡片 -->
      <div class="card">
        <div class="actions-grid">
          <button @click="StartOnline('dogFour')">🐶 狗粮四人联机</button>
          <button @click="StartOnline('dogTwo')">🐾 狗粮二人联机</button>
        </div>
        <!-- 刷新按钮 -->
        <div class="refresh-wrap">
          <button @click="offline">🔄 刷新详情</button>
        </div>

        <!-- 下线 -->
         <div class="refresh-wrap">
          <button @click="offline ">下线</button>
        </div>

        <!--  -->
        <div class="refresh-wrap">
          <button @click="goHome">返回主页</button>
        </div>
      </div>

      <!-- 详情面板 -->
      <div class="details-wrap">
        <div class="detail-panel" v-for="item in detailList" :key="item.key">
          <h2>{{ item.title }}池</h2>
          <div class="detail-content">
            <div class="status-row" v-if="item.members && item.members.length > 0">
              <span class="label">在线人员：</span>
              <span>
                <span v-for="member in item.members" :key="member.name" class="member-name">
                  {{ member.name }}
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
import { message, Modal } from 'ant-design-vue'
import api, { apiMethods } from '@/utils/api'
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const types = [
  { key: 'dog4', title: '狗粮四人联机' },
  { key: 'dog2', title: '狗粮二人联机' },
]

const detailList = ref([])
    const router = useRouter()

async function fetchOnlineDetail() {
  try {
    const res = await api.get('/api/abgiSSE/getOnlineUser') // 返回数组
    detailList.value = types.map(t => {
      const found = res.find(i => i.group_name === t.title)
      return {
        ...t,
        count: found ? found.count : 0,
        members: (found && Array.isArray(found.members)) ? found.members : [],
        status: found ? found.count > 0 : false,
        time: ''
      }
    })
  } catch (e) {
    message.error('获取联机详情失败')
  }
}

const offline = (typeKey) => {
  Modal.confirm({
    title: '确认下线吗？',
    content: '联机下线？',
    okText: '确定',
    cancelText: '取消',
    async onOk() {
      try {
        const response = await apiMethods.offline(typeKey)
        Modal.destroyAll()
        Modal.info({
          title: '上线结果',
          content: "下线成功",
          okText: '关闭'
        })
        fetchOnlineDetail()
      } catch (error) {
        message.error(error)
      }
    }
  })
}


    const goHome = () => {
      router.push('/')
    }

onMounted(() => fetchOnlineDetail())

const StartOnline = (typeKey) => {
  Modal.confirm({
    title: '确认上线吗？',
    content: '联机上线？',
    okText: '确定',
    cancelText: '取消',
    async onOk() {
      try {
        const response = await apiMethods.StartOnline(typeKey)
        Modal.destroyAll()
        Modal.info({
          title: '上线结果',
          content: response,
          okText: '关闭'
        })
        fetchOnlineDetail()
      } catch (error) {
        message.error(error)
      }
    }
  })
}
</script>

<style scoped>
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
.header {
  text-align: center;
  margin-bottom: 32px;
  font-family: 'Segoe UI', sans-serif;
}
.main-content {
  display: flex;
  justify-content: center;
  gap: 32px;
  flex-wrap: wrap;
  align-items: flex-start;
}
.card {
  width: 360px;
  background: #fff;
  border-radius: 18px;
  box-shadow: 0 6px 20px rgba(64,158,255,0.12);
  padding: 32px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.actions-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-top: 16px;
  width: 100%;
}
button {
  padding: 14px 0;
  background: linear-gradient(90deg, #409eff 0%, #66b1ff 100%);
  color: #fff;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 500;
  box-shadow: 0 3px 12px rgba(64,158,255,0.1);
  transition: all 0.25s ease;
  width: 100%;
}
button:hover {
  transform: translateY(-2px);
  background: linear-gradient(90deg, #66b1ff 0%, #409eff 100%);
}

/* 刷新按钮特殊样式 */
.refresh-wrap {
  margin-top: 16px;
  width: 100%;
}
.refresh-wrap button {
  background: linear-gradient(90deg, #67c23a 0%, #85d13a 100%);
}
.refresh-wrap button:hover {
  background: linear-gradient(90deg, #85d13a 0%, #67c23a 100%);
}

.details-wrap {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 24px;
  align-items: stretch;
  min-width: 360px;
}
.detail-panel {
  background: #fff;
  border-radius: 18px;
  box-shadow: 0 6px 20px rgba(64,158,255,0.12);
  padding: 28px 22px;
  display: flex;
  flex-direction: column;
  min-width: 260px;
}
.detail-panel h2 {
  font-size: 18px;
  font-weight: 600;
  color: #409eff;
  margin-bottom: 16px;
  text-align: center;
}
.detail-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.status-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 15px;
}
.label {
  color: #888;
  font-weight: 500;
}
.status {
  font-weight: bold;
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 15px;
  text-align: center;
}
.status.online {
  background: #e6f7ff;
  color: #409eff;
  border: 1px solid #b3e0ff;
}
.status.offline {
  background: #f5f5f5;
  color: #bbb;
  border: 1px solid #eee;
}

/* 响应式调整 */
@media (max-width: 1200px) {
  .main-content {
    flex-direction: column;
    align-items: center;
    gap: 20px;
  }
  .card, .detail-panel {
    width: 95%;
  }
  .details-wrap {
    grid-template-columns: 1fr;
  }
}
@media (max-width: 600px) {
  .card {
    padding: 20px 10px;
  }
  .detail-panel {
    padding: 16px 10px;
  }
}
</style>
