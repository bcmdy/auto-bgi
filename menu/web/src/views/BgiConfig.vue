<template>
  <div class="bgi-config-page">
    <!-- 装饰性的二次元卡通元素 -->
    <div class="floating-elements">
      <div class="float-item flower">🌸</div>
      <div class="float-item sparkle">✨</div>
      <div class="float-item cat">😺</div>
      <div class="float-item ribbon">🎀</div>
    </div>

    <!-- 可爱吉祥物角标 -->
    <div class="mascot" role="img" aria-label="mascot" @click="goHome">👧</div>
    
    <div class="page-header">
      <div class="header-left">
        <a-typography-title :level="4">一条龙配置管理</a-typography-title>
      </div>

    </div>

    <div class="config-select-row">
      <div class="config-select-box">
        <a-select
          v-model:value="currentName"
          :options="configList.map(n => ({ value: n, label: n }))"
          placeholder="请选择配置目录"
          style="width: 220px;"
          @change="selectConfig"
          allow-clear
        />
      </div>
    </div>
    <div class="config-detail-row">
      <a-card :title="currentName ? `配置：${currentName}` : '请选择一个配置'" class="card-detail">
        <div v-if="!currentName" class="placeholder">请选择上方配置名以查看详情</div>
        <div v-else>
                <div class="task-scroll">
                  <a-list :dataSource="visibleTasks" bordered>
                    <template #renderItem="{ item, index }">
                      <a-list-item style="border: 1px solid burlywood; display: flex; align-items: center; justify-content: space-between;">
                        <div style="display:flex; align-items:center; gap:12px;">
                          <div class="task-name">{{ `${index + 1}. ${item.Name}` }}</div>
                          <div class="task-switch">
                            <!-- 控制显示副本的开关，不直接修改原始 taskList 引用 -->
                            <a-switch v-model:checked="visibleEnabled[index]" size="small" />
                          </div>
                        </div>
                        <div style="display:flex; gap:8px;">
                          <a-button size="small" @click="moveUp(index)" :disabled="index===0">上移</a-button>
                          <a-button size="small" @click="moveDown(index)" :disabled="index===visibleTasks.length-1">下移</a-button>
                        </div>
                      </a-list-item>
                    </template>
                  </a-list>
                </div>
          <div class="detail-actions">
            <a-button type="primary" @click="saveConfig" :loading="saving" block>保存配置</a-button>
          </div>
        </div>
      </a-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { apiMethods } from '@/utils/api'
import { useRouter } from 'vue-router'

const configList = ref([])
const currentName = ref('')
const taskList = ref([])
// visibleTasks 用于展示和调序（为不修改原始 taskList，使用副本）
const visibleTasks = ref([])
// visibleEnabled 保存每一项的启用状态的副本
const visibleEnabled = ref([])
const saving = ref(false)
const router = useRouter()

const loadConfigList = async () => {
  try {
    const res = await apiMethods.getBgiConfigAll()
    configList.value = Array.isArray(res.msg) ? res.msg : []
  } catch (err) {
    console.error(err)
    message.error('获取配置目录失败')
  }
}

const selectConfig = async (name) => {
  currentName.value = name
  try {
    const res = await apiMethods.findBgiConfig(name)
    const data = res.msg || {}
    // 原始数据保留到 taskList
    taskList.value = Array.isArray(data.TaskEnabledList) ? data.TaskEnabledList.map(t => ({ ...t })) : []
    // visibleTasks 作为展示副本，初始顺序与 taskList 相同
    visibleTasks.value = taskList.value.map(t => ({ ...t }))
    // visibleEnabled 作为启用状态副本，避免直接修改 taskList
    visibleEnabled.value = visibleTasks.value.map(t => !!t.Enabled)
  } catch (err) {
    console.error(err)
    message.error('读取配置失败')
  }
}

const saveConfig = async () => {
  if (!currentName.value) {
    message.warning('未选择配置')
    return
  }
  saving.value = true
  try {
    // 构造按当前 visibleTasks 顺序的 TaskEnabledList，但不修改原始 taskList
    const orderedList = visibleTasks.value.map((t, idx) => {
      const item = {
        Name: t.Name,
        Enabled: !!visibleEnabled.value[idx]
      }
      // 只有当原始条目包含 Index（非空）时，才更新 Index 为新的顺序值；否则不发送 Index 字段
      if (t.Index !== undefined && t.Index !== null && t.Index !== '') {
        // 后端接收字符串类型的 Index，使用 1-based 的序号并转为字符串
        item.Index = String(idx + 1)
      }
      return item
    })
    const payload = {
      Name: currentName.value,
      TaskEnabledList: orderedList
    }
    await apiMethods.saveBgiConfig(payload)
    message.success('保存成功')
  } catch (err) {
    console.error(err)
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}

// 上移
const moveUp = (index) => {
  if (index <= 0) return
  const vt = visibleTasks.value
  const ve = visibleEnabled.value
  const tmp = vt[index - 1]
  vt[index - 1] = vt[index]
  vt[index] = tmp
  const tmpE = ve[index - 1]
  ve[index - 1] = ve[index]
  ve[index] = tmpE
}

// 下移
const moveDown = (index) => {
  if (index >= visibleTasks.value.length - 1) return
  const vt = visibleTasks.value
  const ve = visibleEnabled.value
  const tmp = vt[index + 1]
  vt[index + 1] = vt[index]
  vt[index] = tmp
  const tmpE = ve[index + 1]
  ve[index + 1] = ve[index]
  ve[index] = tmpE
}

onMounted(() => {
  loadConfigList()
})

const goHome = () => {
  router.push('/')
}
</script>

<style scoped>
/* 基本样式 */
.bgi-config-page {
  position: relative;
  background: linear-gradient(135deg, #fff7ff 0%, #fff0f7 50%, #fff7fb 100%);
  padding: 18px;
  border-radius: 12px;
  box-shadow: 0 8px 20px rgba(255, 105, 180, 0.1);
  min-height: 90vh;
}

/* 二次元卡通风格装饰 */
.floating-elements {
  position: absolute;
  left: 0;
  top: 10px;
  pointer-events: none;
  z-index: 1;
  width: 100%;
}

.float-item {
  position: absolute;
  font-size: 22px;
  opacity: 0.9;
}

.float-item.flower {
  left: 6%;
  top: 6%;
  animation: floatY 6s ease-in-out infinite;
}

.float-item.sparkle {
  right: 10%;
  top: 4%;
  font-size: 18px;
  animation: floatX 7s ease-in-out infinite;
}

.float-item.cat {
  left: 85%;
  top: 70%;
  font-size: 28px;
  animation: floatY 8s ease-in-out infinite;
}

.float-item.ribbon {
  left: 70%;
  top: 10%;
  font-size: 20px;
  animation: floatRotate 9s linear infinite;
}

/* 卡片和列表样式 */
:deep(.ant-card) {
  border-radius: 18px !important;
  box-shadow: 0 12px 30px rgba(255, 105, 180, 0.08) !important;
  border: 1px solid rgba(255, 200, 230, 0.4) !important;
}

:deep(.ant-list-item) {
  border-radius: 12px !important;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.home-btn {
  margin-left: 8px;
}

/* Header布局 */

.config-select-row {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  margin-bottom: 16px;
}
.config-select-box {
  min-width: 220px;
  
}

.config-detail-row {
  width: 100%;
    height: 100%;
}

.card-detail {
  min-height: 400px;
  
}

.task-scroll {
  max-height: 68vh;
  overflow: auto;
  padding-right: 6px;
}

.placeholder {
  color: #888;
  padding: 18px;
  text-align: center;
}

.detail-actions {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

a-list-item {
  padding: 12px 16px !important;
  cursor: pointer;
}

.task-item .task-switch .ant-switch {
  transform: scale(1.0);
}

/* 手机端适配 */
@media (max-width: 600px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }

  .detail-actions {
    justify-content: stretch;
  }

  .list-scroll {
    max-height: 30vh;
  }

  .task-scroll {
    max-height: 60vh;
  }

  .floating-elements .float-item {
    font-size: 16px;
  }
}

/* 右下角返回首页按钮 */
.mascot {
  position: fixed;
  right: 18px;
  bottom: 18px;
  font-size: 48px;
  z-index: 3;
  transform: translateZ(0);
  box-shadow: 0 8px 20px rgba(255, 105, 180, 0.12);
  border-radius: 50%;
  padding: 6px;
  background: linear-gradient(180deg, #fff, #ffdfee);
}

@keyframes floatY {
  0% { transform: translateY(0); }
 50% { transform: translateY(-18px); }
  100% { transform: translateY(0); }
}
@keyframes floatX {
  0% { transform: translateX(0); }
  50% { transform: translateX(12px); }
  100% { transform: translateX(0); }
}

@keyframes floatRotate {
  0% { transform: rotate(0deg); }
  50% { transform: rotate(25deg); }
  100% { transform: rotate(0deg); }
}

</style>