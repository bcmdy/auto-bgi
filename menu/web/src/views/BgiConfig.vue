<template>
  <div class="bgi-config-page">
    <div class="bg-pattern"></div>

    <div class="floating-elements">
      <div class="float-item flower">🌸</div>
      <div class="float-item sparkle">✨</div>
      <div class="float-item cat">😺</div>
      <div class="float-item ribbon">🎀</div>
      <div class="float-item star">⭐</div>
    </div>

    <div class="mascot" role="img" aria-label="mascot" @click="goHome">
      <span class="mascot-emoji">👧</span>
      <span class="mascot-tip">Back</span>
    </div>

    <div class="main-container">
      <div class="page-header">
        <div class="header-left">
          <h1 class="cute-title">✨ 一条龙配置管理 ✨</h1>
        </div>
      </div>

      <div class="config-select-row">
        <div class="config-select-box">
          <span class="label-prefix">当前配置：</span>
          <a-select
              v-model:value="currentName"
              :options="configList.map(n => ({ value: n, label: n }))"
              placeholder="请选择配置目录..."
              class="cute-select"
              @change="selectConfig"
              allow-clear
          />
        </div>
      </div>

      <div class="config-detail-row">
        <a-card :bordered="false" class="card-detail glass-card">
          <template #title>
            <div class="card-header">
              <span class="header-icon">📂</span>
              <span>{{ currentName ? `配置：${currentName}` : '等待选择...' }}</span>
            </div>
          </template>

          <div v-if="!currentName" class="placeholder">
            <div class="empty-state">
              <div class="empty-icon">🎐</div>
              <p>请在上方选择配置名以查看详情哦~</p>
            </div>
          </div>

          <div v-else>
            <div class="task-scroll custom-scrollbar">
              <a-list :dataSource="visibleTasks" :split="false">
                <template #renderItem="{ item, index }">
                  <a-list-item class="cute-list-item">
                    <div class="item-content-wrapper">
                      <div class="item-left">
                        <div class="item-index-badge">{{ index + 1 }}</div>
                        <div class="task-name" :title="item.Name">{{ item.Name }}</div>
                        <div class="task-switch">
                          <a-switch
                              v-model:checked="visibleEnabled[index]"
                              checked-children="开"
                              un-checked-children="关"
                              class="cute-switch"
                          />
                        </div>
                      </div>

                      <div class="item-right">
                        <a-button
                            type="text"
                            class="action-btn up-btn"
                            @click="moveUp(index)"
                            :disabled="index===0"
                            title="上移"
                        >
                          ⬆
                        </a-button>
                        <a-button
                            type="text"
                            class="action-btn down-btn"
                            @click="moveDown(index)"
                            :disabled="index===visibleTasks.length-1"
                            title="下移"
                        >
                          ⬇
                        </a-button>
                      </div>
                    </div>
                  </a-list-item>
                </template>
              </a-list>
            </div>

            <div class="detail-actions">
              <div class="batch-btns">
                <a-button class="cute-btn" @click="enableAll">✅ 全部开启</a-button>
                <a-button class="cute-btn" @click="disableAll">⛔ 全部关闭</a-button>
              </div>
              <a-button
                  type="primary"
                  @click="saveConfig"
                  :loading="saving"
                  class="save-btn"
                  size="large"
                  block
              >
                {{ saving ? '保存中...' : '💾 保存当前配置' }}
              </a-button>
            </div>
          </div>
        </a-card>
      </div>
    </div>
  </div>
</template>

<script setup>
/** * 逻辑代码完全保持原样，未做任何功能性修改
 */
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
      // 保持原始 Index，不重新排序
      if (t.Index !== undefined && t.Index !== null && t.Index !== '') {
        item.Index = t.Index
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

const enableAll = () => {
  visibleEnabled.value = visibleEnabled.value.map(() => true)
}

const disableAll = () => {
  visibleEnabled.value = visibleEnabled.value.map(() => false)
}

onMounted(() => {
  loadConfigList()
})

const goHome = () => {
  router.push('/')
}
</script>

<style scoped>
/* ================= 全局变量与背景 ================= */
.bgi-config-page {
  --primary-pink: #ff9eb5;
  --soft-pink: #fff0f5;
  --deep-pink: #ff69b4;
  --glass-bg: rgba(255, 255, 255, 0.75);
  --glass-border: rgba(255, 255, 255, 0.9);

  position: relative;
  /* 柔和的渐变背景 */
  background: linear-gradient(135deg, #fdfbfb 0%, #ebedee 100%);
  background-image: linear-gradient(to top, #fad0c4 0%, #ffd1ff 100%);
  min-height: 100vh;
  padding: 20px;
  overflow-x: hidden;
  font-family: "Varela Round", "PingFang SC", "Microsoft YaHei", sans-serif;
}

/* 背景波点纹理 */
.bg-pattern {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-image: radial-gradient(rgba(255, 255, 255, 0.6) 2px, transparent 2px);
  background-size: 30px 30px;
  pointer-events: none;
  z-index: 0;
}

.main-container {
  position: relative;
  z-index: 2;
  max-width: 800px;
  margin: 0 auto;
  padding-bottom: 60px; /* 留出底部空间 */
}

/* ================= 漂浮装饰动画 ================= */
.floating-elements {
  position: fixed;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 1;
  overflow: hidden;
}

.float-item {
  position: absolute;
  font-size: 24px;
  filter: drop-shadow(0 2px 4px rgba(255, 105, 180, 0.2));
}

.float-item.flower { left: 5%; top: 10%; animation: floatY 6s ease-in-out infinite; font-size: 32px; }
.float-item.sparkle { right: 8%; top: 5%; animation: floatRotate 5s linear infinite; font-size: 24px; }
.float-item.cat { left: 88%; top: 60%; animation: floatY 7s ease-in-out infinite; font-size: 40px; }
.float-item.ribbon { left: 2%; bottom: 15%; animation: floatX 8s ease-in-out infinite; font-size: 28px; }
.float-item.star { right: 15%; bottom: 10%; animation: floatRotate 8s linear infinite; font-size: 22px; }

@keyframes floatY {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-20px); }
}
@keyframes floatX {
  0%, 100% { transform: translateX(0); }
  50% { transform: translateX(20px); }
}
@keyframes floatRotate {
  0% { transform: rotate(0deg) scale(1); }
  50% { transform: rotate(180deg) scale(1.1); }
  100% { transform: rotate(360deg) scale(1); }
}

/* ================= 头部标题 ================= */
.page-header {
  text-align: center;
  margin-bottom: 24px;
}

.cute-title {
  color: #fff;
  font-size: 26px;
  font-weight: 700;
  text-shadow: 2px 2px 0px var(--primary-pink), 0 0 10px rgba(255, 105, 180, 0.5);
  margin: 0;
  letter-spacing: 2px;
}

/* ================= 选择框区域 ================= */
.config-select-row {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.config-select-box {
  background: var(--glass-bg);
  padding: 10px 20px;
  border-radius: 50px;
  box-shadow: 0 4px 15px rgba(255, 105, 180, 0.15);
  display: flex;
  align-items: center;
  backdrop-filter: blur(5px);
  width: 100%;
  max-width: 400px;
}

.label-prefix {
  color: #888;
  font-weight: bold;
  margin-right: 10px;
  white-space: nowrap;
}

/* 覆盖 Ant Select 样式 */
:deep(.ant-select-selector) {
  border-radius: 20px !important;
  border: 1px solid #ffe4e1 !important;
  background-color: #fff !important;
  box-shadow: none !important;
}
:deep(.ant-select-focused .ant-select-selector) {
  border-color: var(--deep-pink) !important;
  box-shadow: 0 0 0 2px rgba(255, 105, 180, 0.2) !important;
}
:deep(.cute-select) {
  width: 100%;
}

/* ================= 卡片容器 ================= */
.glass-card {
  background: var(--glass-bg) !important;
  backdrop-filter: blur(10px);
  border-radius: 24px !important;
  border: 2px solid #fff !important;
  box-shadow: 0 10px 40px rgba(255, 182, 193, 0.3) !important;
}

:deep(.ant-card-head) {
  border-bottom: 1px dashed rgba(255, 105, 180, 0.3) !important;
  min-height: 56px;
}

.card-header {
  display: flex;
  align-items: center;
  font-size: 18px;
  color: #555;
  font-weight: 600;
}
.header-icon {
  margin-right: 8px;
  font-size: 20px;
}

/* ================= 列表区域 ================= */
.task-scroll {
  max-height: 60vh;
  overflow-y: auto;
  padding: 4px;
}

/* 自定义粉色滚动条 */
.custom-scrollbar::-webkit-scrollbar {
  width: 8px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: #fff0f5;
  border-radius: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #ffb6c1;
  border-radius: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #ff69b4;
}

/* 列表项样式重写 */
.cute-list-item {
  background: white;
  margin-bottom: 12px;
  border-radius: 16px !important;
  border: 1px solid transparent !important;
  padding: 0 !important;
  box-shadow: 0 2px 8px rgba(200, 200, 200, 0.1);
  transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
  overflow: hidden;
}

.cute-list-item:hover {
  transform: translateY(-3px);
  box-shadow: 0 5px 15px rgba(255, 105, 180, 0.25);
  border-color: #ffc0cb !important;
}

.item-content-wrapper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 12px 16px;
}

.item-left {
  display: flex;
  align-items: center;
  flex: 1;
  overflow: hidden; /* 防止文字过长 */
}

.item-index-badge {
  background: var(--soft-pink);
  color: var(--deep-pink);
  width: 28px;
  height: 28px;
  line-height: 28px;
  text-align: center;
  border-radius: 50%;
  font-weight: bold;
  font-size: 14px;
  margin-right: 12px;
  flex-shrink: 0;
}

.task-name {
  font-size: 15px;
  color: #444;
  margin-right: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

.task-switch {
  margin-left: 8px;
  flex-shrink: 0;
}

/* 定制 Switch 颜色 */
:deep(.ant-switch-checked) {
  background-color: var(--deep-pink) !important;
}
:deep(.ant-switch-checked:hover) {
  background-color: #ff1493 !important; /* hotpink darken */
}

/* 右侧按钮组 */
.item-right {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.action-btn {
  border-radius: 10px !important;
  background: #f8f8f8;
  color: #888;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  border: none;
}
.action-btn:not(:disabled):hover {
  background: var(--soft-pink);
  color: var(--deep-pink);
  transform: scale(1.1);
}
.action-btn:disabled {
  background: transparent;
  opacity: 0.3;
}

/* ================= 底部保存区 ================= */
.detail-actions {
  margin-top: 20px;
  padding-top: 10px;
  border-top: 1px dashed rgba(255, 105, 180, 0.2);
}

.batch-btns {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
}

.cute-btn {
  flex: 1;
  border-radius: 20px !important;
  border: 1px solid #ffe4e1 !important;
  color: #666;
  font-weight: 600;
  transition: all 0.3s;
}

.cute-btn:hover {
  color: var(--deep-pink);
  border-color: var(--deep-pink) !important;
  background-color: var(--soft-pink);
  transform: translateY(-2px);
}

.save-btn {
  background: linear-gradient(45deg, #ff9a9e 0%, #fad0c4 99%, #fad0c4 100%) !important;
  border: none !important;
  height: 48px !important;
  font-size: 16px !important;
  border-radius: 24px !important;
  font-weight: bold;
  box-shadow: 0 4px 15px rgba(255, 154, 158, 0.4);
  transition: transform 0.2s;
}
.save-btn:active {
  transform: scale(0.98);
}
.save-btn:hover {
  filter: brightness(1.05);
  box-shadow: 0 6px 20px rgba(255, 154, 158, 0.6);
}

/* ================= 空状态 ================= */
.placeholder {
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.empty-state {
  text-align: center;
  color: #999;
}
.empty-icon {
  font-size: 48px;
  margin-bottom: 10px;
  animation: floatY 3s ease-in-out infinite;
}

/* ================= 吉祥物返回按钮 ================= */
.mascot {
  position: fixed;
  right: 20px;
  bottom: 20px;
  width: 60px;
  height: 60px;
  background: white;
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  box-shadow: 0 5px 20px rgba(255, 105, 180, 0.3);
  cursor: pointer;
  z-index: 99;
  transition: all 0.3s;
  border: 2px solid #fff0f5;
  user-select: none;
}
.mascot-emoji {
  font-size: 28px;
  line-height: 1;
}
.mascot-tip {
  font-size: 10px;
  color: var(--deep-pink);
  font-weight: bold;
}
.mascot:hover {
  transform: scale(1.1) rotate(-10deg);
  box-shadow: 0 8px 25px rgba(255, 105, 180, 0.5);
}
.mascot:active {
  transform: scale(0.95);
}

/* ================= 移动端适配 ================= */
@media (max-width: 600px) {
  .bgi-config-page {
    padding: 10px;
  }

  .card-detail {
    border-radius: 16px !important;
  }

  :deep(.ant-card-body) {
    padding: 12px !important;
  }

  .item-content-wrapper {
    padding: 10px;
  }

  .item-index-badge {
    width: 24px;
    height: 24px;
    line-height: 24px;
    font-size: 12px;
    margin-right: 8px;
  }

  .task-name {
    font-size: 14px;
  }

  .action-btn {
    width: 32px;
    height: 32px;
  }

  .mascot {
    right: 15px;
    bottom: 15px;
    width: 50px;
    height: 50px;
  }
  .mascot-emoji { font-size: 24px; }
  .mascot-tip { display: none; }
}
</style>