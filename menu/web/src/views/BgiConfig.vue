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
            <a-list :dataSource="taskList" bordered>
              <template #renderItem="{ item }">
                <a-list-item style="border: 1px solid burlywood;">
                  <div class="task-item">
                    <div class="task-name">{{ item.Name }}</div>
                    <div class="task-switch">
                      <a-switch v-model:checked="item.Enabled" size="small" />
                    </div>
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
    taskList.value = Array.isArray(data.TaskEnabledList) ? data.TaskEnabledList.map(t => ({ ...t })) : []
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
    const payload = {
      Name: currentName.value,
      TaskEnabledList: taskList.value.map(t => ({ Name: t.Name, Enabled: !!t.Enabled, Index: t.Index || '' }))
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