<template>
  <div class="one-remote-page">
    <div class="page-bg" aria-hidden="true"></div>

    <div class="bg-ornaments" aria-hidden="true">
      <span class="geom geom-1 geom-circle"></span>
      <span class="geom geom-2 geom-triangle"></span>
      <span class="geom geom-3 geom-star"></span>
    </div>

    <div class="page-shell">
      <header class="topbar">
        <div class="topbar-left">
          <a-button class="topbar-back" @click="router.push('/')">返回首页</a-button>
          <div class="topbar-titles">
            <div class="title"><span class="title-icon">🌸</span>1Remote 管理<span class="title-icon">✨</span></div>
          </div>
        </div>

        <div class="topbar-right">
          <a-button class="topbar-refresh-btn" type="default" @click="refreshSessions" :loading="refreshLoading">刷新</a-button>
        </div>
      </header>

      <div class="stats">
        <div class="stat">
          <span class="sticker sticker-1"></span>
          <span class="sticker sticker-2"></span>
          <div class="stat-label">会话总数</div>
          <div class="stat-value">{{ totalCount }}</div>
        </div>
        <div class="stat">
          <span class="sticker sticker-1"></span>
          <span class="sticker sticker-2"></span>
          <div class="stat-label">活跃</div>
          <div class="stat-value stat-green">{{ activeCount }}</div>
        </div>
        <div class="stat">
          <span class="sticker sticker-1"></span>
          <span class="sticker sticker-2"></span>
          <div class="stat-label">断开</div>
          <div class="stat-value stat-red">{{ disconnectedCount }}</div>
        </div>
        <div class="stat">
          <span class="sticker sticker-1"></span>
          <span class="sticker sticker-2"></span>
          <div class="stat-label">监听</div>
          <div class="stat-value stat-blue">{{ listeningCount }}</div>
        </div>
      </div>

      <div class="content-grid">
        <div class="left-stack">
          <div class="panel">
            <div class="panel-header">
              <div class="panel-title">快速操作</div>
              <div class="panel-tools">
                <a-tag v-if="selectedID" color="processing">已选择 ID: {{ selectedID }}</a-tag>
              </div>
            </div>
            <div class="panel-body">
              <div class="action-form">
                <a-input v-model:value="launcher" placeholder="输入  用户名（可选）" allow-clear />
                <div class="action-row">
                  <a-button type="primary" size="large" :loading="startLoading" @click="handleStart">✨ 启动 1Remote</a-button>
                  <a-button danger size="large" :loading="logoffLoading" @click="handleLogoff">💥 注销选中会话</a-button>
                </div>
                <div class="form-hint">
                  使用方式：可直接输入用户名进行启动；也可以在右侧列表中选择会话后再操作。注销仅对选中会话生效。
                </div>
              </div>
            </div>
          </div>

          <div class="panel performance-panel">
            <div class="panel-header">
              <div class="panel-title">电脑性能监控</div>
              <div class="panel-tools">
                <a-tag color="processing" v-if="performanceLoading">刷新中</a-tag>
              </div>
            </div>
            <div class="panel-body performance-body">
              <div class="performance-list">
                <div class="performance-item">
                  <div class="performance-label">CPU</div>
                  <div class="performance-value">{{ performanceData.CPUUsage || '-' }}</div>
                </div>
                <div class="performance-item">
                  <div class="performance-label">内存</div>
                  <div class="performance-value">{{ performanceData.MemoryUsage || '-' }}</div>
                </div>
                <div class="performance-item">
                  <div class="performance-label">GPU</div>
                  <div class="performance-value">{{ performanceData.GPUUsage || '-' }}</div>
                </div>
              </div>
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
          <div class="panel-body table-body">
            <a-table
              :data-source="filteredSessions"
              :columns="columns"
              size="middle"
              :row-key="(r) => r.ID"
              :loading="refreshLoading"
              :pagination="{ pageSize: 10 }"
              :row-selection="{
                selectedRowKeys,
                onChange: onSelectChange,
                type: 'radio'
              }"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'State'">
                  <a-tag :class="['state-tag', `state-${stateColor(record.State)}`]">{{ record.State }}</a-tag>
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
const performanceLoading = ref(false)

const sessions = ref([])
const selectedRowKeys = ref([])
const search = ref('')
const launcher = ref('')
const performanceData = ref({ CPUUsage: '-', MemoryUsage: '-', GPUUsage: '-' })
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
const refreshLoading = computed(() => loading.value || performanceLoading.value)

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

const loadPerformance = async () => {
  try {
    performanceLoading.value = true
    const res = await apiMethods.oneRemotePerformance()
    const data = res?.data || {}
    performanceData.value = {
      CPUUsage: data.CPUUsage || '-',
      MemoryUsage: data.MemoryUsage || '-',
      GPUUsage: data.GPUUsage || '-'
    }
  } catch (e) {
    message.error('获取性能监控失败')
  } finally {
    performanceLoading.value = false
  }
}

const refreshSessions = async () => {
  await Promise.all([loadSessions(), loadPerformance()])
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
    message.warning('请先在右侧选择一条会话')
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
  if (!id) {
    message.warning('请先在右侧选择一条会话')
    return
  }
  const content = `确认注销会话 ID=${id}？`
  Modal.confirm({
    title: '确认注销',
    content,
    okText: '确定',
    cancelText: '取消',
    onOk: async () => {
      try {
        logoffLoading.value = true
        await apiMethods.oneRemoteLogoff({ id })
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
  refreshSessions()
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
  background: rgba(255, 255, 255, 0.76);
  border: 2px solid rgba(255, 154, 198, 0.32);
  box-shadow: 0 14px 38px rgba(255, 154, 198, 0.18);
  border-radius: 16px;
  backdrop-filter: blur(8px);
  padding: 12px 14px;
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
.topbar-back {
  border-radius: 14px;
  border: 2px solid rgba(255, 154, 198, 0.3);
  background: rgba(255, 255, 255, 0.9);
  font-weight: 700;
  box-shadow: 0 10px 22px rgba(255, 154, 198, 0.18);
  transition: all 0.22s ease;
}
.topbar-back:hover {
  transform: translateY(-1px);
  box-shadow: 0 14px 26px rgba(255, 154, 198, 0.24);
}
.topbar-refresh-btn {
  border: none;
  border-radius: 14px;
  background: linear-gradient(120deg, rgba(255, 182, 213, 0.95) 0%, rgba(233, 213, 255, 0.95) 60%, rgba(199, 210, 254, 0.9) 100%);
  color: #1f2937;
  font-weight: 800;
  box-shadow: 0 12px 26px rgba(233, 213, 255, 0.32);
  transition: all 0.22s ease;
}
.topbar-refresh-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 16px 32px rgba(233, 213, 255, 0.38);
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
  transition: transform 0.22s ease, box-shadow 0.22s ease;
}
.stat::before {
  content: '';
  position: absolute;
  inset: -2px;
  background: radial-gradient(220px 120px at 20% 20%, rgba(255, 154, 198, 0.22) 0%, rgba(255, 154, 198, 0) 65%),
    radial-gradient(220px 120px at 85% 0%, rgba(147, 197, 253, 0.2) 0%, rgba(147, 197, 253, 0) 65%);
  opacity: 1;
}
.stat::after {
  content: '';
  position: absolute;
  width: 120px;
  height: 120px;
  background: radial-gradient(circle at 40% 40%, rgba(199, 210, 254, 0.28), rgba(255, 255, 255, 0));
  filter: blur(18px);
  opacity: 0.7;
  bottom: -40px;
  right: -32px;
}
.stat:hover {
  transform: translateY(-2px);
  box-shadow: 0 22px 40px rgba(59, 130, 246, 0.16);
}
.stat .sticker {
  position: absolute;
  width: 68px;
  height: 68px;
  filter: blur(10px);
  opacity: 0.55;
}
.stat .sticker-1 {
  top: -18px;
  left: -12px;
  background: radial-gradient(circle at 40% 40%, rgba(255, 182, 213, 0.45), rgba(255, 255, 255, 0));
}
.stat .sticker-2 {
  bottom: -22px;
  right: -6px;
  background: radial-gradient(circle at 50% 50%, rgba(187, 247, 208, 0.32), rgba(255, 255, 255, 0));
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
.left-stack {
  display: grid;
  gap: 16px;
}
.content-grid .panel:first-child {
  background: rgba(255, 255, 255, 0.78);
  border-color: rgba(255, 182, 213, 0.32);
  box-shadow: 0 18px 45px rgba(255, 182, 213, 0.18);
}
.panel {
  background: rgba(255, 255, 255, 0.82);
  backdrop-filter: blur(8px);
  border-radius: 18px;
  border: 2px solid rgba(255, 154, 198, 0.26);
  box-shadow: 0 18px 45px rgba(255, 154, 198, 0.16);
  overflow: hidden;
  position: relative;
  transition: transform 0.22s ease, box-shadow 0.22s ease;
}
.panel:hover {
  transform: translateY(-2px);
  box-shadow: 0 22px 52px rgba(255, 154, 198, 0.2);
}
.bg-ornaments {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}
.geom {
  position: absolute;
  display: block;
  filter: blur(32px);
  opacity: 0.65;
  mix-blend-mode: screen;
}
.geom-circle {
  width: 260px;
  height: 260px;
  border-radius: 50%;
  background: radial-gradient(circle at 30% 30%, rgba(249, 168, 212, 0.45), rgba(233, 213, 255, 0.18));
}
.geom-triangle {
  width: 0;
  height: 0;
  border-left: 140px solid transparent;
  border-right: 140px solid transparent;
  border-bottom: 240px solid rgba(199, 210, 254, 0.32);
  transform: rotate(12deg);
}
.geom-star {
  width: 220px;
  height: 220px;
  background: radial-gradient(circle at 40% 40%, rgba(187, 247, 208, 0.42), rgba(233, 213, 255, 0.16));
  clip-path: polygon(50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%, 50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%);
}
.geom-1 {
  top: -60px;
  left: -40px;
}
.geom-2 {
  bottom: -90px;
  right: 80px;
}
.geom-3 {
  top: 120px;
  right: -80px;
}
@keyframes floaty {
  0% {
    transform: translateY(0) rotate(0deg);
  }
  50% {
    transform: translateY(-6px) rotate(4deg);
  }
  100% {
    transform: translateY(0) rotate(0deg);
  }
}
.geom {
  animation: floaty 8s ease-in-out infinite;
}
.geom-2 {
  animation-duration: 10s;
}
.geom-3 {
  animation-duration: 9s;
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
  background: rgba(255, 255, 255, 0.6);
  border: 1px dashed rgba(255, 182, 213, 0.22);
  border-radius: 14px;
}
.table-body {
  padding: 0;
  background: transparent;
  border: none;
}
.action-form {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}
.action-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  align-items: center;
  gap: 10px;
}
.panel-body + .panel-body {
  margin-top: 12px;
}
.action-form :deep(.ant-input) {
  border-radius: 14px;
  border: 2px solid rgba(255, 154, 198, 0.28);
  background: rgba(255, 255, 255, 0.92);
  transition: all 0.2s ease;
}
.action-form :deep(.ant-input:focus),
.action-form :deep(.ant-input-focused) {
  box-shadow: 0 0 0 3px rgba(255, 182, 213, 0.35);
  border-color: rgba(255, 154, 198, 0.48);
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
  background-image:
    repeating-linear-gradient(45deg, rgba(255, 182, 213, 0.22) 0, rgba(255, 182, 213, 0.22) 6px, transparent 6px, transparent 12px),
    radial-gradient(circle at 14% 20%, rgba(199, 210, 254, 0.2), transparent 45%),
    radial-gradient(circle at 84% 68%, rgba(187, 247, 208, 0.18), transparent 52%);
}
.action-row :deep(.ant-btn) {
  transition: all 0.2s ease;
}
.action-row :deep(.ant-btn:hover) {
  transform: translateY(-2px);
  box-shadow: 0 16px 30px rgba(255, 154, 198, 0.22);
}
.selected-tip {
  margin-top: 8px;
  font-size: 12px;
  color: #6b7280;
}
.performance-body {
  padding-top: 14px;
}
.performance-list {
  display: grid;
  gap: 10px;
}
.performance-item {
  border-radius: 12px;
  border: 2px solid rgba(147, 197, 253, 0.24);
  background: rgba(255, 255, 255, 0.85);
  padding: 10px 12px;
}
.performance-label {
  font-size: 12px;
  font-weight: 700;
  color: #6b7280;
}
.performance-value {
  margin-top: 4px;
  font-size: 13px;
  color: #1f2937;
  line-height: 1.4;
  word-break: break-word;
}
@media (max-width: 640px) {
  .page-shell {
    padding: 14px;
  }
  .stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }
  .content-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }
  .topbar {
    flex-direction: column;
    align-items: stretch;
  }
  .topbar-right {
    justify-content: space-between;
    gap: 8px;
  }
  .topbar-right .ant-btn,
  .topbar-left .ant-btn {
    width: 100%;
  }
  .title {
    font-size: 22px;
  }
  .sub {
    font-size: 13px;
  }
  .action-row {
    grid-template-columns: 1fr;
  }
  .panel-body {
    padding: 12px;
  }
  .table-body {
    padding: 0;
  }
  .table-body :deep(.ant-table-container) {
    width: 100vw;
    margin-left: -14px;
    border-radius: 0;
    border-left: none;
    border-right: none;
  }
  .table-body :deep(.ant-table-content) {
    overflow-x: auto;
  }
  .bg-ornaments {
    opacity: 0.4;
  }
}

.one-remote-page :deep(.ant-table) {
  background: transparent;
}
.one-remote-page :deep(.ant-table-container) {
  border-radius: 12px;
  overflow: hidden;
  border: 2px solid rgba(147, 197, 253, 0.25);
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(6px);
  box-shadow: 0 14px 32px rgba(147, 197, 253, 0.18);
}
.one-remote-page :deep(.ant-table-thead > tr > th) {
  background: rgba(255, 154, 198, 0.12);
  color: #6d28d9;
  font-weight: 800;
  backdrop-filter: blur(4px);
}
.one-remote-page :deep(.ant-table-tbody > tr:hover > td) {
  background: rgba(147, 197, 253, 0.16);
}
.one-remote-page :deep(.ant-input) {
  border-radius: 14px;
  border: 2px solid rgba(255, 154, 198, 0.22);
  background: rgba(255, 255, 255, 0.9);
  transition: all 0.2s ease;
}
.one-remote-page :deep(.ant-input:focus),
.one-remote-page :deep(.ant-input-focused) {
  box-shadow: 0 0 0 3px rgba(199, 210, 254, 0.4);
  border-color: rgba(199, 210, 254, 0.8);
}
.one-remote-page :deep(.ant-btn) {
  border-radius: 14px;
  font-weight: 800;
  transition: all 0.2s ease;
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
.one-remote-page :deep(.ant-btn:hover) {
  transform: translateY(-2px);
  box-shadow: 0 16px 32px rgba(251, 113, 133, 0.28);
}
.state-tag {
  border-radius: 12px !important;
  padding: 2px 10px;
  border: none !important;
  font-weight: 700;
  background: rgba(255, 255, 255, 0.75);
}
.state-green {
  color: #0ea36a !important;
  background: rgba(16, 185, 129, 0.16) !important;
}
.state-red {
  color: #f43f5e !important;
  background: rgba(244, 63, 94, 0.16) !important;
}
.state-blue {
  color: #3b82f6 !important;
  background: rgba(59, 130, 246, 0.16) !important;
}
.state-default {
  color: #6b7280 !important;
  background: rgba(107, 114, 128, 0.14) !important;
}
.one-remote-page :deep(.ant-switch) {
  background: rgba(255, 154, 198, 0.35);
}
.one-remote-page :deep(.ant-switch-checked) {
  background: #a78bfa;
}
</style>

