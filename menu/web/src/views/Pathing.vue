<template>
  <div class="pathing-page">
    <h2>地图追踪更新配置</h2>
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else>
      <!-- PC端表格 -->
      <table class="pathing-table desktop-table">
        <thead>
          <tr>
            <th>配置组名</th>
            <th>地图追踪文件夹路径</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(item, idx) in pathingList" :key="idx">
            <td>
              <!-- 下拉框替换输入框 -->
              <select v-model="item.name" required>
                <option value="" disabled>请选择配置组名</option>
                <option v-for="opt in configOptions" :key="opt.value" :value="opt.label">
                  {{ opt.label }}
                </option>
              </select>
            </td>
            <td class="path-cell">
              <div class="path-scroll">
                <!-- 级联选择替换输入框 -->
                <a-cascader
                  v-model:value="item.folderName"
                  :options="cascaderOptions"
                  placeholder="请选择地图追踪文件夹路径"
                  change-on-select
                  style="width: 95%;"
                />
              </div>
            </td>
            <td>
              <button
                :disabled="savingIdx === idx"
                @click="updateSingle(idx)"
              >
                {{ savingIdx === idx ? '新增或者更新中...' : '新增或者更新' }}
              </button>
              <!-- 新增清理按钮（每行） -->
              <button
                :disabled="cleaningIdx === idx"
                @click="cleanSingle(idx)"
                style="margin-left:8px;"
              >
                {{ cleaningIdx === idx ? '清理中...' : '清理' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <!-- 移动端卡片 -->
      <div class="mobile-list">
        <div v-for="(item, idx) in pathingList" :key="idx" class="mobile-card">
          <div class="mobile-row">
            <span class="mobile-label">配置组：</span>
            <!-- 下拉框替换输入框 -->
            <select v-model="item.name" required>
              <option value="" disabled>请选择配置组名</option>
              <option v-for="opt in configOptions" :key="opt.value" :value="opt.label">
                {{ opt.label }}
              </option>
            </select>
          </div>
          <div class="mobile-row">
            <span class="mobile-label">地图追踪文件夹：</span>
            <!-- 级联选择替换输入框 -->
            <a-cascader
              v-model:value="item.folderName"
              :options="cascaderOptions"
              placeholder="请选择地图追踪文件夹路径"
              change-on-select
              style="width: 100%;"
            />
          </div>
          <div class="mobile-row">
            <button
              :disabled="savingIdx === idx"
              @click="updateSingle(idx)"
              class="mobile-btn"
            >
              {{ savingIdx === idx ? '新增或者更新中...' : '新增或者更新' }}
            </button>
            <!-- 新增清理按钮（移动端每卡片） -->
            <button
              :disabled="cleaningIdx === idx"
              @click="cleanSingle(idx)"
              class="mobile-btn"
              style="margin-left:8px;"
            >
              {{ cleaningIdx === idx ? '清理中...' : '清理' }}
            </button>
          </div>
        </div>
      </div>
      <button type="button" @click="addPathing" style="margin-top:16px;">添加地图追踪</button>
      <!-- 新增保存·按钮 -->
      <!-- <button
        type="button"
        @click="saveAll"
        :disabled="savingAll"
        style="margin-left:16px;margin-top:16px;"
      >
        {{ savingAll ? '保存中...' : '保存·' }}
      </button> -->

      <!-- 刷新按钮 -->
        <!-- <button
        type="button"
        @click="ListPathingUpdatePaths()"
        :disabled="refresh"
        style="margin-left:16px;margin-top:16px;"
      >
        {{ refresh ? '刷新中...' : '刷新' }}
      </button> -->

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiMethods } from '@/utils/api'
import { Cascader as ACascader,message, Modal } from "ant-design-vue"


const pathingList = ref([])
const loading = ref(true)
const savingIdx = ref(-1)
const savingAll = ref(false)
const refresh = ref(false)
const cleaningIdx = ref(-1)
const configOptions = ref([])
const pathOptions = ref([]) // 原始数据
const cascaderOptions = ref([]) // 级联选择用数据

const fetchPathing = async () => {
  loading.value = true
  try {
    const res = await fetch('/api/scriptGroup/ConfigPathing')
    const json = await res.json()
    // folderName字符串转数组
    pathingList.value = (json.data || []).map(item => ({
      ...item,
      folderName: item.folderName ? item.folderName.split('\\') : []
    }))
  } catch (e) {
    alert('加载路径配置失败')
    pathingList.value = []
  } finally {
    loading.value = false
  }
}

// 获取配置选项
const fetchConfigOptions = async () => {
  try {
    const response = await apiMethods.getListGroups()
    // 直接用字符串数组
    configOptions.value = (response || []).map(item => ({
      label: item,
      value: item
    }))
  } catch (error) {
    console.error('获取配置选项失败:', error)
  }
}

// 递归转换为Cascader格式
function convertToCascaderOptions(data) {
  if (!Array.isArray(data)) return []
  return data.map(item => ({
    label: item.fileName,
    value: item.fileName,
    children: item.fileNameChild ? convertToCascaderOptions(item.fileNameChild) : undefined
  }))
}

// 获取地图追踪文件夹选项（递归结构）
const fetchPathOptions = async () => {
  try {
    const res = await fetch('/api/scriptGroup/listAllGroups')
    const json = await res.json()
    pathOptions.value = json.data || []
    cascaderOptions.value = convertToCascaderOptions(pathOptions.value)
  } catch (error) {
    console.error('获取地图追踪路径失败:', error)
    pathOptions.value = []
    cascaderOptions.value = []
  }
}

// 判断是否选到最后一级
function isLastLevel(arr) {
  if (!Array.isArray(arr) || arr.length === 0) return false
  let opts = cascaderOptions.value
  for (let i = 0; i < arr.length; i++) {
    const found = opts.find(o => o.value === arr[i])
    if (!found) return false
    if (i === arr.length - 1) {
      // 最后一级不能有children
      return !found.children || found.children.length === 0
    }
    opts = found.children || []
  }
  return false
}

const updateSingle = async (idx) => {
  const item = pathingList.value[idx]
  if (!isLastLevel(item.folderName)) {
    alert('请将地图追踪文件夹路径选择到最后一级！')
    return // 直接返回，不做任何更新操作
  }
  savingIdx.value = idx
  try {
    // folderName数组转字符串
    const payload = {
      ...pathingList.value[idx],
      folderName: Array.isArray(pathingList.value[idx].folderName)
        ? pathingList.value[idx].folderName.join('\\')
        : pathingList.value[idx].folderName
    }
    const res = await fetch('/api/scriptGroup/UpdatePathing', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
    const json = await res.json()
    if (json.status === 'success') {
      alert(json.data || '保存成功')
      await fetchPathing()
    } else {
      alert(json.data || '保存失败')
    }
  } catch (e) {
    alert('保存失败')
  } finally {
    savingIdx.value = -1
  }
}

const ListPathingUpdatePaths = async () => {
  Modal.confirm({
    title: '确认刷新？',
    content: '刷新将会重新读取配置？',
    okText: '确定',
    cancelText: '取消',
    onOk: async () => {
      try {
        await apiMethods.listPathingUpdatePaths()
        message.success('刷新成功！')
        // 刷新当前列表
        await fetchPathing()
      } catch (error) {
        message.error('刷新失败！')
      }
    }
  })
}

const saveAll = async () => {
  // 校验所有项
  for (const item of pathingList.value) {
    if (!isLastLevel(item.folderName)) {
      alert('请将所有地图追踪文件夹路径选择到最后一级！')
      savingAll.value = false
      return
    }
  }
  savingAll.value = true
  try {
    // 批量转换folderName为字符串
    const payload = pathingList.value.map(item => ({
      ...item,
      folderName: Array.isArray(item.folderName)
        ? item.folderName.join('\\')
        : item.folderName
    }))
    const res = await fetch('/api/scriptGroup/savePathing', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
    const json = await res.json()
    if (json.status === 'success') {
      alert(json.message || '批量保存成功')
      await fetchPathing()
    } else {
      alert(json.msg || '批量保存失败')
    }
  } catch (e) {
    alert('批量保存失败')
  } finally {
    savingAll.value = false
  }
}

const cleanSingle = async (idx) => {
  const item = pathingList.value[idx]
  cleaningIdx.value = idx
  try {
    // folderName数组转字符串
    const payload = {
      ...item,
      folderName: Array.isArray(item.folderName)
        ? item.folderName.join('\\')
        : item.folderName
    }
    const res = await fetch('/api/scriptGroup/cleanAllPathing', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
    let json
    try {
      json = await res.json()
    } catch (err) {
      // 如果不是标准json，尝试文本解析
      const text = await res.text()
      try {
        json = JSON.parse(text)
      } catch {
        json = { status: 'error', msg: text }
      }
    }
    console.log('清理返回:', json)
    if (json.status === 'success') {
      alert(json.message || '清理成功')
      await fetchPathing()
    } else {
      alert(json.message || '清理失败')
    }
  } catch (e) {
    alert('清理失败',e)
  } finally {
    cleaningIdx.value = -1
  }
}

const addPathing = () => {
  pathingList.value.push({ name: '', folderName: [] })
}

onMounted(() => {
  fetchPathing()
  fetchConfigOptions()
  fetchPathOptions()
})
</script>

<style scoped>
.pathing-page {
  max-width: 1200px;
  margin: 40px auto;
  background: #fff6fb;
  border-radius: 20px;
  box-shadow: 0 8px 32px #ffb6c1;
  padding: 40px;
}
h2 {
  color: #ff6eb4;
  margin-bottom: 32px;
  font-size: 2.2rem;
  text-align: center;
}
.loading {
  color: #e91e63;
  font-size: 1.2em;
  text-align: center;
}
.pathing-table {
  width: 100%;
  border-collapse: collapse;
  background: #fff;
  margin-bottom: 18px;
  font-size: 1.15rem;
}
.pathing-table th, .pathing-table td {
  border: 2px solid #ffc0da;
  padding: 16px;
  text-align: left;
}
.pathing-table th {
  background: #ffe4ee;
  color: #e91e63;
  font-size: 1.1rem;
}
input, select {
  width: 95%;
  padding: 10px 14px;
  border: 1px solid #ffc0da;
  border-radius: 8px;
  font-size: 1rem;
  background: #fff;
}
button {
  background: linear-gradient(90deg, #ff6eb4, #ff8cc8);
  color: #fff;
  border: none;
  border-radius: 10px;
  padding: 10px 24px;
  font-weight: bold;
  cursor: pointer;
  transition: background 0.2s;
  font-size: 1rem;
}
button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
button[type="button"] {
  background: #fff;
  color: #ff6eb4;
  border: 2px solid #ff6eb4;
  margin-top: 10px;
}
.desktop-table {
  display: table;
}
.mobile-list {
  display: none;
}
.path-cell {
  max-width: 380px;
  min-width: 180px;
  padding: 0;
}
.path-scroll {
  overflow-x: auto;
  width: 100%;
}
.path-scroll input {
  min-width: 320px;
  width: 100%;
  font-size: 1rem;
  padding: 10px 14px;
  border: 1px solid #ffc0da;
  border-radius: 8px;
  background: #fff;
  box-sizing: border-box;
}

/* 手机端适配 */
@media (max-width: 768px) {
  .pathing-page {
    max-width: 100%;
    margin: 10px auto;
    padding: 10px;
    border-radius: 10px;
  }
  h2 {
    font-size: 1.3rem;
    margin-bottom: 18px;
  }
  .desktop-table {
    display: none;
  }
  .mobile-list {
    display: block;
  }
  .mobile-card {
    background: #fff;
    border-radius: 12px;
    box-shadow: 0 2px 8px #ffe4ee;
    padding: 16px;
    margin-bottom: 16px;
    font-size: 1rem;
  }
  .mobile-row {
    display: flex;
    align-items: center;
    margin-bottom: 10px;
    gap: 8px;
  }
  .mobile-label {
    min-width: 70px;
    color: #e91e63;
    font-size: 0.98em;
  }
  .mobile-btn {
    width: 100%;
    padding: 10px 0;
    font-size: 1rem;
    border-radius: 8px;
  }
  input, select {
    width: 100%;
    padding: 8px 10px;
    font-size: 1rem;
  }
}
</style>
