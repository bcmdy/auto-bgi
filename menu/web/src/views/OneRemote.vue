<template>
  <div class="one-remote-page">
    <div class="page-bg" aria-hidden="true"></div>

    <div class="page-shell">
      <header class="topbar">
        <div class="topbar-left">
          <a-button class="topbar-back" @click="router.push('/')">返回首页</a-button>
          <div class="topbar-titles">
            <div class="title"><span class="title-icon">🌸</span>1Remote 管理<span class="title-icon">✨</span></div>
            <div class="sub">会话查看 / 启动 / 注销 · 二次元小助手版</div>
          </div>
        </div>

        <div class="topbar-right">
          <div class="topbar-refresh">
            <span class="refresh-label">自动刷新啾</span>
            <a-switch v-model:checked="autoRefresh" />
          </div>
          <a-button type="default" @click="refreshSessions" :loading="loading">刷新</a-button>
        </div>
      </header>

      <div class="stats">
        <div class="stat">
          <div class="stat-label">会话总数</div>
          <div class="stat-value">{{ totalCount }}</div>
        </div>
        <div class="stat">
          <div class="stat-label">活跃</div>
          <div class="stat-value stat-green">{{ activeCount }}</div>
        </div>
        <div class="stat">
          <div class="stat-label">断开</div>
          <div class="stat-value stat-red">{{ disconnectedCount }}</div>
        </div>
        <div class="stat">
          <div class="stat-label">监听</div>
          <div class="stat-value stat-blue">{{ listeningCount }}</div>
        </div>
      </div>

      <div class="content-grid">
        <div class="panel">
        <div class="panel-header">
          <div class="panel-title">快速操作</div>
          <div class="panel-tools">
            <a-tag v-if="selectedID" color="processing">已选择 ID: {{ selectedID }}</a-tag>
          </div>
        </div>
        <div class="panel-body">
          <div class="form-row">
            <a-input v-model:value="launcher" placeholder="输入 launcher / 用户名" allow-clear />
            <a-button type="primary" :loading="startLoading" @click="handleStart">启动 1Remote</a-button>
            <a-button danger :loading="logoffLoading" @click="handleLogoff">注销</a-button>
          </div>
          <div class="form-hint">
            启动：优先使用输入框；未填写时会使用选中会话的用户名/会话名。注销：优先按选中 ID 注销，否则按输入的用户名查找会话再注销。
          </div>
        </div>
      </div>

        <div class="panel">
        <div class="panel-header">
          <div class="panel-title">当前会话</div>
          <div class="panel-tools">
            <a-input v-model:value="search" placeholder="搜索 用户名 / 会话名" allow-clear style="width: 260px" />
          </div>
        </div>
        <div class="panel-body">
          <a-table
            :data-source="filteredSessions"
            :columns="columns"
            size="middle"
            :row-key="(r) => r.ID"
            :loading="loading"
            :pagination="{ pageSize: 10 }"
            :row-selection="{
              selectedRowKeys,
              onChange: onSelectChange,
              type: 'radio'
            }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'State'">
                <a-tag :color="stateColor(record.State)">{{ record.State }}</a-tag>
              </template>
            </template>
          </a-table>
          <div class="selected-tip">
            选择的会话 ID：{{ selectedID || '-' }}
          </div>
        </div>
      </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch, onUnmounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'
import { useRouter } from 'vue-router'

const router = useRouter()

const loading = ref(false)
const startLoading = ref(false)
const logoffLoading = ref(false)

const sessions = ref([])
const selectedRowKeys = ref([])
const search = ref('')
const launcher = ref('')
const autoRefresh = ref(true)
let refreshTimer = null

const columns = [
  { title: '会话名', dataIndex: 'SessionName', key: 'SessionName', width: 180 },
  { title: '用户名', dataIndex: 'Username', key: 'Username', width: 180 },
  { title: 'ID', dataIndex: 'ID', key: 'ID', width: 80 },
  { title: '状态', dataIndex: 'State', key: 'State', width: 140 }
]

const selectedID = computed(() => {
  const k = selectedRowKeys.value[0]
  return k ? String(k) : ''
})

const filteredSessions = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return sessions.value
  return sessions.value.filter((s) => {
    return (
      String(s.Username || '').toLowerCase().includes(q) ||
      String(s.SessionName || '').toLowerCase().includes(q)
    )
  })
})

const totalCount = computed(() => sessions.value.length)
const activeCount = computed(() => sessions.value.filter((s) => stateColor(s.State) === 'green').length)
const disconnectedCount = computed(() => sessions.value.filter((s) => stateColor(s.State) === 'red').length)
const listeningCount = computed(() => sessions.value.filter((s) => stateColor(s.State) === 'blue').length)

const stateColor = (state) => {
  const s = String(state || '').toLowerCase()
  if (s.includes('active') || s.includes('已活动')) return 'green'
  if (s.includes('disc') || s.includes('断开')) return 'red'
  if (s.includes('listen') || s.includes('监听')) return 'blue'
  return 'default'
}

const onSelectChange = (keys) => {
  selectedRowKeys.value = Array.isArray(keys) ? keys.slice(0, 1) : []
}

const loadSessions = async () => {
  try {
    loading.value = true
    const res = await apiMethods.oneRemoteGetSessions()
    sessions.value = Array.isArray(res?.data) ? res.data : []
  } catch (e) {
    message.error('获取会话失败')
  } finally {
    loading.value = false
  }
}

const refreshSessions = () => {
  loadSessions()
}

const startAutoRefresh = () => {
  if (refreshTimer) clearInterval(refreshTimer)
  refreshTimer = setInterval(() => {
    loadSessions()
  }, 5000)
}

const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

const resolveLauncher = () => {
  const v = launcher.value.trim()
  if (v) return v
  const id = selectedID.value
  const picked = sessions.value.find((s) => String(s.ID) === id)
  if (!picked) return ''
  return (picked.Username || picked.SessionName || '').trim()
}

const handleStart = async () => {
  const value = resolveLauncher()
  if (!value) {
    message.warning('请输入 launcher / 选择一条会话')
    return
  }
  try {
    startLoading.value = true
    await apiMethods.oneRemoteStart(value)
    message.success('已触发启动')
  } catch (e) {
    message.error('启动失败')
  } finally {
    startLoading.value = false
  }
}

const handleLogoff = async () => {
  const id = selectedID.value
  const name = launcher.value.trim()
  if (!id && !name) {
    message.warning('请选择会话或输入用户名')
    return
  }
  const content = id ? `确认注销会话 ID=${id}？` : `确认注销用户 ${name} 的会话？`
  Modal.confirm({
    title: '确认注销',
    content,
    okText: '确定',
    cancelText: '取消',
    onOk: async () => {
      try {
        logoffLoading.value = true
        await apiMethods.oneRemoteLogoff(id ? { id } : { username: name })
        message.success('注销成功')
        await loadSessions()
      } catch (e) {
        message.error('注销失败')
      } finally {
        logoffLoading.value = false
      }
    }
  })
}

onMounted(() => {
  loadSessions()
  if (autoRefresh.value) startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})

watch(autoRefresh, (v) => {
  if (v) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
})
</script>

<style scoped>
.one-remote-page {
  position: relative;
  isolation: isolate;
  min-height: 100vh;
  box-sizing: border-box;
  font-family: 'Comic Sans MS', 'Microsoft YaHei UI', 'Microsoft YaHei', 'Segoe UI', system-ui, -apple-system, Arial;
  color: #1f2937;
}
.page-bg {
  position: absolute;
  inset: 0;
  z-index: 0;
  background:
    radial-gradient(850px 520px at 18% 8%, rgba(255, 154, 198, 0.38) 0%, rgba(255, 154, 198, 0) 60%),
    radial-gradient(900px 520px at 90% 0%, rgba(147, 197, 253, 0.34) 0%, rgba(147, 197, 253, 0) 55%),
    radial-gradient(900px 520px at 65% 92%, rgba(167, 243, 208, 0.26) 0%, rgba(167, 243, 208, 0) 55%),
    linear-gradient(180deg, #fff7fb 0%, #f3f8ff 60%, #f7fff9 100%);
}
.page-bg::after {
  content: '';
  position: absolute;
  inset: 0;
  background-image:
    radial-gradient(rgba(255, 255, 255, 0.95) 2px, transparent 2px),
    radial-gradient(rgba(255, 255, 255, 0.75) 1px, transparent 1px);
  background-size: 48px 48px, 22px 22px;
  background-position: 0 0, 9px 13px;
  opacity: 0.6;
}
.page-bg::before {
  content: '';
  position: absolute;
  inset: -80px;
  background:
    radial-gradient(circle at 12% 18%, rgba(255, 182, 213, 0.22) 0%, rgba(255, 182, 213, 0) 45%),
    radial-gradient(circle at 82% 22%, rgba(191, 219, 254, 0.22) 0%, rgba(191, 219, 254, 0) 45%),
    radial-gradient(circle at 58% 88%, rgba(187, 247, 208, 0.18) 0%, rgba(187, 247, 208, 0) 48%);
  filter: blur(18px);
  opacity: 1;
}
.page-shell {
  position: relative;
  z-index: 1;
  padding: 24px;
  box-sizing: border-box;
  max-width: 1240px;
  margin: 0 auto;
}
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  margin-bottom: 16px;
}
.topbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}
.topbar-titles {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.topbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}
.topbar-refresh {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.78);
  border: 2px solid rgba(255, 154, 198, 0.35);
  box-shadow: 0 14px 30px rgba(255, 154, 198, 0.16);
}
.refresh-label {
  font-size: 12px;
  color: #7c3aed;
  font-weight: 700;
}
.title {
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  letter-spacing: 0.2px;
  line-height: 1.15;
  text-shadow: 0 8px 18px rgba(255, 154, 198, 0.25);
}
.title-icon {
  display: inline-block;
  margin: 0 6px;
  transform: translateY(-1px);
}
.sub {
  font-size: 14px;
  color: #6b7280;
  line-height: 1.25;
}
.stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}
.stat {
  border-radius: 14px;
  padding: 12px 14px;
  background: rgba(255, 255, 255, 0.8);
  border: 2px solid rgba(255, 154, 198, 0.25);
  box-shadow: 0 16px 34px rgba(59, 130, 246, 0.12);
  position: relative;
  overflow: hidden;
}
.stat::before {
  content: '';
  position: absolute;
  inset: -2px;
  background: radial-gradient(220px 120px at 20% 20%, rgba(255, 154, 198, 0.22) 0%, rgba(255, 154, 198, 0) 65%),
    radial-gradient(220px 120px at 85% 0%, rgba(147, 197, 253, 0.2) 0%, rgba(147, 197, 253, 0) 65%);
  opacity: 1;
}
.stat-label {
  font-size: 12px;
  color: #6b7280;
  font-weight: 700;
  position: relative;
}
.stat-value {
  margin-top: 6px;
  font-size: 22px;
  font-weight: 800;
  color: #111827;
  position: relative;
}
.stat-green {
  color: #10b981;
}
.stat-red {
  color: #fb7185;
}
.stat-blue {
  color: #60a5fa;
}
.content-grid {
  display: grid;
  grid-template-columns: 440px 1fr;
  gap: 16px;
}
.panel {
  background: rgba(255, 255, 255, 0.82);
  backdrop-filter: blur(8px);
  border-radius: 18px;
  border: 2px solid rgba(255, 154, 198, 0.26);
  box-shadow: 0 18px 45px rgba(255, 154, 198, 0.16);
  overflow: hidden;
}
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  border-bottom: 1px dashed rgba(255, 154, 198, 0.35);
  background: linear-gradient(90deg, rgba(255, 154, 198, 0.12) 0%, rgba(147, 197, 253, 0.1) 55%, rgba(167, 243, 208, 0.08) 100%);
}
.panel-title {
  font-size: 16px;
  font-weight: 600;
  color: #7c3aed;
  letter-spacing: 0.2px;
}
.panel-tools {
  display: flex;
  gap: 10px;
}
.panel-body {
  padding: 16px;
}
.form-row {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 10px;
}
.form-hint {
  margin-top: 10px;
  padding: 10px 12px;
  border-radius: 12px;
  background: rgba(255, 154, 198, 0.1);
  border: 2px dashed rgba(255, 154, 198, 0.35);
  color: #6d28d9;
  font-size: 12px;
  line-height: 1.4;
}
.selected-tip {
  margin-top: 8px;
  font-size: 12px;
  color: #6b7280;
}
@media (max-width: 640px) {
  .page-shell {
    padding: 14px;
  }
  .stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .content-grid {
    grid-template-columns: 1fr;
  }
  .topbar {
    flex-direction: column;
    align-items: stretch;
  }
  .topbar-right {
    justify-content: space-between;
  }
  .form-row {
    grid-template-columns: 1fr;
  }
}

.one-remote-page :deep(.ant-table) {
  background: transparent;
}
.one-remote-page :deep(.ant-table-container) {
  border-radius: 12px;
  overflow: hidden;
  border: 2px solid rgba(147, 197, 253, 0.25);
}
.one-remote-page :deep(.ant-table-thead > tr > th) {
  background: rgba(255, 154, 198, 0.12);
  color: #6d28d9;
  font-weight: 800;
}
.one-remote-page :deep(.ant-table-tbody > tr:hover > td) {
  background: rgba(147, 197, 253, 0.16);
}
.one-remote-page :deep(.ant-input) {
  border-radius: 14px;
  border: 2px solid rgba(255, 154, 198, 0.22);
  background: rgba(255, 255, 255, 0.9);
}
.one-remote-page :deep(.ant-btn) {
  border-radius: 14px;
  font-weight: 800;
}
.one-remote-page :deep(.ant-btn-primary) {
  border: none;
  background: linear-gradient(90deg, #fb7185 0%, #a78bfa 60%, #60a5fa 100%);
  box-shadow: 0 14px 28px rgba(251, 113, 133, 0.22);
}
.one-remote-page :deep(.ant-btn-default) {
  border: 2px solid rgba(147, 197, 253, 0.35);
  background: rgba(255, 255, 255, 0.75);
}
.one-remote-page :deep(.ant-btn-dangerous) {
  border: none;
  background: linear-gradient(90deg, #fb7185 0%, #fda4af 100%);
  box-shadow: 0 14px 28px rgba(251, 113, 133, 0.18);
}
.one-remote-page :deep(.ant-switch) {
  background: rgba(255, 154, 198, 0.35);
}
.one-remote-page :deep(.ant-switch-checked) {
  background: #a78bfa;
}
</style>
