<template>
  <div class="update-container">
    <div class="sponsor-note">感谢 <span class="sponsor-name">思姐</span> 赞助 ❤</div>
    <h2>ABGI版本检测与更新</h2>

    <div class="version-row">
      <div class="label">当前版本：</div>
      <div class="value">{{ currentVersion }}</div>
    </div>

    <div class="version-row">
      <div class="label">最新版本：</div>
      <div class="value">{{ latestVersion }}</div>
    </div>

    <div class="actions">
      <a-button type="default" @click="refresh" :loading="checking">刷新</a-button>
      <a-button type="primary" @click="doUpdate" :disabled="!isDifferent || loading" :loading="loading">
        {{ isDifferent ? '更新到最新版' : '已是最新版' }}
      </a-button>
    </div>

   <h2 style="margin-top: 100px;">BGI远程更新</h2>
    <div class="download-by-url">
        <div class="version-row">
          <div class="label">当前版本：</div>
          <div class="value">{{ bgiCurrentVersion }}</div>
        </div>

        <div class="version-row">
          <div class="label">最新版本：</div>
          <div class="value">{{ bgiLatestVersion }}</div>
        </div>

      <div style="margin-top:18px; display:flex; gap:8px; align-items:center;">
        <!-- <a-input v-model:value="downloadUrl" placeholder="粘贴 BGI zip 下载地址（支持 http(s)）" /> -->
        <a-button type="primary" @click="downloadByUrl" :loading="downloading" :disabled="!bgiIsDifferent || !bgiCanUpdate">在线更新</a-button>
      </div>
      <!-- <div style="margin-top:8px;color:#666;font-size:12px">注意：请输入可直接下载的压缩包链接；解压直接是文件才行，二次压缩不行</div> -->
    </div>

    <div class="note" v-if="note">{{ note }}</div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'

const currentVersion = ref('加载中...')
const latestVersion = ref('加载中...')
const loading = ref(false)
const checking = ref(false)
const note = ref('')
const downloading = ref(false)
// BGI 远程更新的版本信息
const bgiCurrentVersion = ref('加载中...')
const bgiLatestVersion = ref('加载中...')
const bgiCanUpdate = ref(false)

const normalize = (v) => (v == null ? '' : String(v).trim())

const isDifferent = computed(() => {
  return normalize(currentVersion.value) !== normalize(latestVersion.value) && latestVersion.value !== ''
})

// BGI 远程更新是否有版本差异
const bgiIsDifferent = computed(() => {
  return normalize(bgiCurrentVersion.value) !== normalize(bgiLatestVersion.value) && bgiLatestVersion.value !== ''
})

const refresh = async () => {
  checking.value = true
  note.value = ''
  try {
    const cur = await apiMethods.aBgiGetCurrentVersion()
    // 兼容不同返回形态
    currentVersion.value = cur?.version ?? cur?.data?.version ?? (typeof cur === 'string' ? cur : JSON.stringify(cur))

    const last = await apiMethods.aBgiGetLastVersion()
    latestVersion.value = last?.version ?? last?.data?.version ?? (typeof last === 'string' ? last : JSON.stringify(last))
  } catch (err) {
    console.error(err)
    message.error('获取版本信息失败')
    note.value = err?.message || String(err)
  } finally {
    checking.value = false
  }
}

const doUpdate = async () => {
  if (!isDifferent.value) return
  loading.value = true
  note.value = ''
  try {
    const res = await apiMethods.aBgiUpdate()
  
    if (res && (res.code === 200 || res.msg === '更新成功' || res.success === true || res.message === '更新成功')) {
      message.success('更新已触发，可能需要重启软件')
    } else {

      message.success((res && (res.msg || res.message)) || '更新请求已发送（请检查服务端日志）')
    }

    await refresh()
  } catch (err) {
    console.error(err)
    message.error('更新失败：' + (err?.message || String(err)))
    note.value = err?.message || String(err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  refresh()
  refreshBgiVersions()
})


const refreshBgiVersions = async () => {
  bgiCurrentVersion.value = '加载中...'
  bgiLatestVersion.value = '加载中...'
  bgiCanUpdate.value = false
  try {
   
    const res = await apiMethods.aBgiGetVersions()
    if (res && typeof res === 'object') {
      console.debug('aBgiGetVersions response:', res)
      bgiCurrentVersion.value = res.currentVersion ?? res.current ?? bgiCurrentVersion.value
      bgiLatestVersion.value = res.lastVersion ?? res.latest ?? bgiLatestVersion.value

      
      if (Object.prototype.hasOwnProperty.call(res, 'canUpdate')) {
        if (typeof res.canUpdate === 'boolean') bgiCanUpdate.value = res.canUpdate
        else if (typeof res.canUpdate === 'string') bgiCanUpdate.value = res.canUpdate === 'true'
        else bgiCanUpdate.value = Boolean(res.canUpdate)
      } else {

        bgiCanUpdate.value = normalize(bgiCurrentVersion.value) !== normalize(bgiLatestVersion.value) && bgiLatestVersion.value !== ''
      }
    } else {

      bgiCurrentVersion.value = '未知'
      bgiLatestVersion.value = ''
      bgiCanUpdate.value = false
    }
  } catch (err) {
    console.warn('刷新 BGI 版本失败', err)
    bgiCurrentVersion.value = '获取失败'
    bgiLatestVersion.value = ''
    bgiCanUpdate.value = false
  }
}

const downloadByUrl = async () => {
  downloading.value = true
  try {
    const res = await apiMethods.downloadBgi()
    message.success((res && (res.message || res.msg)) || '下载更新请求已发送')
    await refreshBgiVersions()
  } catch (err) {
    console.error(err)
    message.error('通过 URL 更新失败：' + (err?.message || String(err)))
    note.value = err?.message || String(err)
  } finally {
    downloading.value = false
  }
}
</script>

<style scoped>
.update-container {
  max-width: 720px;
  margin: 120px auto 40px;
  padding: 20px;
  background: rgba(255,255,255,0.9);
  border-radius: 12px;
  box-shadow: 0 6px 20px rgba(0,0,0,0.06);
  color: #333;
}
.version-row{display:flex;align-items:center;margin:12px 0}
.label{width:110px;color:#666;font-weight:600}
.value{flex:1;color:#111}
.actions{display:flex;gap:12px;margin-top:18px}
.note{margin-top:12px;color:#d00}

.sponsor-note{
  background: linear-gradient(90deg,#fff7f9,#fff0f5);
  border: 1px solid #ffd7ea;
  color: #d6006a;
  padding: 10px 14px;
  border-radius: 10px;
  font-weight:700;
  display:inline-block;
  margin-bottom:12px;
}
.sponsor-name{color:#ff2d7a}
</style>
