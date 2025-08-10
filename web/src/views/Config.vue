<template>
  <div class="config-container">
    <!-- 装饰性背景元素 -->
    <div class="floating-elements">
      <div class="floating-heart" v-for="i in 6" :key="i" :style="{ animationDelay: `${i * 0.5}s` }">💖</div>
      <div class="floating-star" v-for="i in 4" :key="`star-${i}`" :style="{ animationDelay: `${i * 0.7}s` }">✨</div>
    </div>

    <div class="content-wrapper">
      <a-typography-title :level="2" class="page-title">
        <span class="title-icon">🌸</span>
        配置设置
        <span class="title-icon">🌸</span>
      </a-typography-title>

      <a-form :model="formData" layout="vertical" @finish="handleSubmit" ref="formRef" class="config-form">
        <!-- 基础设置 -->
        <a-card title="基础设置" class="config-card basic-settings">
          <template #extra>
            <span class="card-icon">⚙️</span>
          </template>
          
          <div class="form-grid">
            <a-form-item 
              label="BetterGI 路径" 
              name="BetterGIAddress"
              class="form-item-enhanced"
            >
              <div class="input-wrapper">
                <span class="input-icon">📁</span>
                <a-input 
                  v-model:value="formData.BetterGIAddress" 
                  placeholder="D:\subject\lua\BetterGI" 
                  class="enhanced-input"
                />
              </div>
              <div class="help-text">
                <span class="help-icon">💡</span>
                填写 BetterGI 的本地路径，例如 D:\subject\lua\BetterGI。不能有逗号，单斜杠
              </div>
            </a-form-item>

            <a-form-item 
              label="企业机器人地址" 
              name="webhookURL"
              class="form-item-enhanced"
            >
              <div class="input-wrapper">
                <span class="input-icon">🤖</span>
                <a-input 
                  v-model:value="formData.webhookURL" 
                  placeholder="企业微信机器人地址"
                  class="enhanced-input"
                />
              </div>
              <div class="help-text">
                <span class="help-icon">💡</span>
                企业微信机器人地址，具体怎么申请，自行解决
              </div>
            </a-form-item>

            <a-form-item 
              label="bgi关闭提示内容" 
              name="content"
              class="form-item-enhanced"
            >
              <div class="input-wrapper">
                <span class="input-icon">📝</span>
                <a-input 
                  v-model:value="formData.content" 
                  placeholder="可以填 autobgi 的网页链接"
                  class="enhanced-input"
                />
              </div>
              <div class="help-text">
                <span class="help-icon">💡</span>
                bgi关闭之后，将会把这个内容发送到企业微信，可以填 autobgi 的网页链接，比如背包的网页地址
              </div>
            </a-form-item>

            <a-form-item 
              label="端口号" 
              name="post"
              class="form-item-enhanced"
            >
              <div class="input-wrapper">
                <span class="input-icon">🔌</span>
                <a-input-group compact class="port-input-group">
                  <a-input style="width: 30px" value=":" disabled class="port-separator" />
                  <a-input 
                    v-model:value="formData.post" 
                    placeholder="8082" 
                    style="width: calc(100% - 30px)" 
                    class="enhanced-input"
                  />
                </a-input-group>
              </div>
              <div class="help-text">
                <span class="help-icon">💡</span>
                端口号，默认是 8082，如果冲突了可以改成你喜欢的
              </div>
            </a-form-item>
          </div>
        </a-card>

        <!-- 一条龙配置 -->
        <a-card title="一条龙配置" class="config-card onelong-config">
          <template #extra>
            <div class="card-extra">
              <span class="card-icon">🐉</span>
              <a-tooltip title="这个是你的一条龙配置的名称，根据今天周几，来启动哪一个一条龙，必须要有7个，从星期天开始。">
                <QuestionCircleOutlined class="help-icon-btn" />
              </a-tooltip>
            </div>
          </template>
          
          <div class="weekday-grid">
            <div 
              v-for="(day, index) in weekdays" 
              :key="index"
              class="weekday-item"
              :class="{ 'active': formData.ConfigNames[index] }"
            >
              <div class="weekday-label">
                <span class="day-icon">{{ getDayIcon(index) }}</span>
                {{ day }}
              </div>
              <a-select 
                v-model:value="formData.ConfigNames[index]" 
                placeholder="选择配置"
                :options="configOptions"
                size="large"
                class="weekday-select"
                :class="{ 'has-value': formData.ConfigNames[index] }"
              />
            </div>
          </div>
        </a-card>

        <!-- 背包统计 -->
        <a-card title="背包统计" class="config-card bag-statistics">
          <template #extra>
            <div class="card-extra">
              <span class="card-icon">🎒</span>
              <a-tooltip title="背包统计功能，需要配合仓库的背包统计，摩拉ocr和原石统计js一起使用，将你需要关注的材料填上，摩拉和原石可以不用填，只要执行了脚本，自动有。">
                <QuestionCircleOutlined class="help-icon-btn" />
              </a-tooltip>
            </div>
          </template>

          <div class="dynamic-list">
            <div 
              v-for="(keyword, index) in formData.bagKeywords" 
              :key="`bag-${index}`"
              class="list-item"
            >
              <div class="item-content">
                <span class="item-icon">💎</span>
                <a-input 
                  v-model:value="formData.bagKeywords[index]" 
                  placeholder="材料关键字"
                  class="enhanced-input"
                />
              </div>
              <a-button 
                type="primary" 
                danger 
                @click="removeBagKeyword(index)"
                class="remove-btn"
                :disabled="formData.bagKeywords.length <= 1"
              >
                <DeleteOutlined />
              </a-button>
            </div>

            <a-button type="dashed" @click="addBagKeyword" class="add-btn">
              <PlusOutlined /> 添加材料
            </a-button>
          </div>
        </a-card>

        <!-- 需要通知的关键字 -->
        <a-card title="需要关注关键字" class="config-card script-names">
          <template #extra>
            <div class="card-extra">
              <span class="card-icon">📜</span>
              <a-tooltip title="需要关注的关键字，这里填写后，只要bgi触发了关键词就会通知企业微信">
                <QuestionCircleOutlined class="help-icon-btn" />
              </a-tooltip>
            </div>
          </template>

          <div class="dynamic-list">
            <div
                v-for="(LogKeyword, index) in formData.LogKeywords"
                :key="`LogKeyword-${index}`"
                class="list-item"
            >
              <div class="item-content">
                <span class="item-icon">🔧</span>
                <a-input
                    v-model:value="formData.LogKeywords[index]"
                    placeholder="关键字"
                    class="enhanced-input"
                />
              </div>
              <a-button
                  type="primary"
                  danger
                  @click="removeLogKeyword(index)"
                  class="remove-btn"
                  :disabled="formData.LogKeywords.length <= 1"
              >
                <DeleteOutlined />
              </a-button>
            </div>

            <a-button type="dashed" @click="addLogKeyword" class="add-btn">
              <PlusOutlined /> 添加关键字
            </a-button>
          </div>
        </a-card>


        <!-- 需要备份的文件或文件夹路径 -->
        <a-card title="需要备份的文件或者文件夹路径" class="config-card backup-paths">
          <template #extra>
            <div class="card-extra">
              <span class="card-icon">💾</span>
              <a-tooltip title="需要备份的文件或者文件夹路径，bgi目录下的文件夹路径">
                <QuestionCircleOutlined class="help-icon-btn" />
              </a-tooltip>
            </div>
          </template>

          <div class="dynamic-list">
            <div 
              v-for="(backup, index) in formData.backups" 
              :key="`backup-${index}`"
              class="list-item"
            >
              <div class="item-content">
                <span class="item-icon">📂</span>
                <a-input 
                  v-model:value="formData.backups[index]" 
                  placeholder="文件夹名称"
                  class="enhanced-input"
                />
              </div>
              <a-button 
                type="primary" 
                danger 
                @click="removeBackup(index)"
                class="remove-btn"
                :disabled="formData.backups.length <= 1"
              >
                <DeleteOutlined />
              </a-button>
            </div>

            <a-button type="dashed" @click="addBackup" class="add-btn">
              <PlusOutlined /> 添加文件或者文件夹
            </a-button>
          </div>
        </a-card>

        <!-- 一条龙时间设置 -->
        <a-card title="一条龙时间设置" class="config-card time-settings">
          <template #extra>
            <div class="card-extra">
              <span class="card-icon">⏰</span>
              <a-tooltip title="每天定时启动一条龙，记住，不管你当前bgi在干什么，都会结束，开始一条龙">
                <QuestionCircleOutlined class="help-icon-btn" />
              </a-tooltip>
            </div>
          </template>

          <div class="time-settings-content">
            <a-form-item class="checkbox-item">
              <a-checkbox v-model:checked="formData.OneLong.isStartTimeLong" class="enhanced-checkbox">
                <span class="checkbox-label">
                  <span class="checkbox-icon">🎯</span>
                  启用一条龙定时启动
                </span>
              </a-checkbox>
            </a-form-item>

            <div class="time-inputs" v-show="formData.OneLong.isStartTimeLong">
              <div class="time-input-group">
                <a-form-item label="启动小时（0~23）" name="OneLongHour" class="time-form-item">
                  <a-input-number 
                    v-model:value="formData.OneLong.OneLongHour" 
                    :min="0" 
                    :max="23" 
                    placeholder="例如：7"
                    class="enhanced-input-number"
                  />
                </a-form-item>
              </div>
              <div class="time-input-group">
                <a-form-item label="启动分钟（0~59）" name="OneLongMinute" class="time-form-item">
                  <a-input-number 
                    v-model:value="formData.OneLong.OneLongMinute" 
                    :min="0" 
                    :max="59" 
                    placeholder="例如：30"
                    class="enhanced-input-number"
                  />
                </a-form-item>
              </div>
            </div>
          </div>
        </a-card>

        <!-- 控制配置 -->
        <a-card title="控制配置" class="config-card control-settings">
          <template #extra>
            <span class="card-icon">🎮</span>
          </template>

          <div class="control-options">
            <a-form-item class="checkbox-item">
              <a-checkbox v-model:checked="formData.Control.IsCloseYuanShen" class="enhanced-checkbox">
                <span class="checkbox-label">
                  <span class="checkbox-icon">🎮</span>
                  bgi关闭是否需要关闭原神
                </span>
              </a-checkbox>
            </a-form-item>

            <a-form-item class="checkbox-item">
              <a-checkbox v-model:checked="formData.Control.SendWeChatImage" class="enhanced-checkbox">
                <span class="checkbox-label">
                  <span class="checkbox-icon">📸</span>
                  是否开启每隔一小时发送截图
                </span>
              </a-checkbox>
            </a-form-item>
          </div>
        </a-card>

        <!-- 米游社签到设置 -->
        <a-card title="米游社签到设置" class="config-card mys-sign">
          <template #extra>
            <div class="card-extra">
              <span class="card-icon">🎁</span>
              <a-tooltip title="米游社签到设置，需要配合米游社签到脚本使用，没有就关闭">
                <QuestionCircleOutlined class="help-icon-btn" />
              </a-tooltip>
            </div>
          </template>

          <div class="mys-sign-content">
            <a-form-item class="checkbox-item">
              <a-checkbox v-model:checked="formData.MySign.isMysSignIn" class="enhanced-checkbox">
                <span class="checkbox-label">
                  <span class="checkbox-icon">🎁</span>
                  启用米游社签到
                </span>
              </a-checkbox>
            </a-form-item>

            <a-form-item label="签到 URL" name="mysUrl" class="form-item-enhanced" v-show="formData.MySign.isMysSignIn">
              <div class="input-wrapper">
                <span class="input-icon">🔗</span>
                <a-input 
                  v-model:value="formData.MySign.url" 
                  placeholder="http://localhost:8888"
                  class="enhanced-input"
                />
              </div>
            </a-form-item>
          </div>
        </a-card>

        <!-- 1Remote 配置 -->
        <a-card title="1Remote远程监控" class="config-card one-remote-config">
          <template #extra>
            <div class="card-extra">
              <span class="card-icon">🖥️</span>
              <a-tooltip title="远程监控配置，填写日志路径和关键字，支持自动监控远程连接断开等事件">
                <QuestionCircleOutlined class="help-icon-btn" />
              </a-tooltip>
            </div>
          </template>
          <div class="one-remote-content">
            <a-form-item class="checkbox-item">
              <a-checkbox v-model:checked="formData.OneRemote.IsMonitor" class="enhanced-checkbox">
                <span class="checkbox-label">
                  <span class="checkbox-icon">🖥️</span>
                  启用远程监控
                </span>
              </a-checkbox>
            </a-form-item>
            <!-- 只有勾选时才显示下面内容 -->
            <div v-show="formData.OneRemote.IsMonitor">
              <a-form-item label="日志文件路径" name="OneRemoteLogFilePath" class="form-item-enhanced">
                <div class="input-wrapper">
                  <span class="input-icon">📄</span>
                  <a-input
                    v-model:value="formData.OneRemote.LogFilePath"
                    placeholder="C:\Users\Administrator\Desktop\1Remote-1.2.0-net9-x64\.logs"
                    class="enhanced-input"
                  />
                </div>
                <div class="help-text">
                  <span class="help-icon">💡</span>
                  填写 OneRemote 的日志文件夹路径，到 .logs 目录下，例如 C:\Users\Administrator\Desktop\1Remote-1.2.0-net9-x64\.logs
                </div>
              </a-form-item>
              <div class="dynamic-list">
                <div
                  v-for="(keyword, index) in formData.OneRemote.LogKeywords"
                  :key="`oneremote-keyword-${index}`"
                  class="list-item"
                >
                  <div class="item-content">
                    <span class="item-icon">🔑</span>
                    <a-input
                      v-model:value="formData.OneRemote.LogKeywords[index]"
                      placeholder="远程监控关键字"
                      class="enhanced-input"
                    />
                  </div>
                  <a-button
                    type="primary"
                    danger
                    @click="removeOneRemoteKeyword(index)"
                    class="remove-btn"
                    :disabled="formData.OneRemote.LogKeywords.length <= 1"
                  >
                    <DeleteOutlined />
                  </a-button>
                </div>
                <a-button type="dashed" @click="addOneRemoteKeyword" class="add-btn">
                  <PlusOutlined /> 添加关键字
                </a-button>
              </div>
            </div>
          </div>
        </a-card>

        <!-- 提交按钮 -->
        <div class="submit-section">
          <a-button 
            type="primary" 
            html-type="submit" 
            size="large" 
            :loading="loading" 
            class="submit-btn"
          >
            <span class="submit-icon">💾</span>
            保存配置
            <span class="submit-icon">💾</span>
          </a-button>
        </div>
      </a-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { QuestionCircleOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons-vue'
import { useRouter } from 'vue-router'
import { apiMethods } from '@/utils/api'

const router = useRouter()
const formRef = ref()
const loading = ref(false)

const weekdays = ['星期天', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六']
const configOptions = ref([])

// 获取星期图标
const getDayIcon = (index) => {
  const icons = ['🌞', '🌙', '⭐', '🌸', '🌺', '🌻', '🌹']
  return icons[index] || '🌸'
}

// 表单数据
const formData = reactive({
  BetterGIAddress: '',
  webhookURL: '',
  content: '',
  post: '8082',
  ConfigNames: new Array(7).fill(''),
  bagKeywords: [''],
  LogKeywords: [''],
  backups: [''],
  OneLong: {
    isStartTimeLong: false,
    OneLongHour: 7,
    OneLongMinute: 30
  },
  Control: {
    IsCloseYuanShen: false,
    SendWeChatImage: false
  },
  MySign: {
    isMysSignIn: false,
    url: ''
  },
  OneRemote: {
    IsMonitor: true,
    LogFilePath: '',
    LogKeywords: ['']
  }
})

// 获取配置选项
const fetchConfigOptions = async () => {
  try {
    const response = await apiMethods.getOneLongAllName()
    configOptions.value = response.data?.map(item => ({
      label: item.replace('.json', ''),
      value: item
    })) || []
  } catch (error) {
    console.error('获取配置选项失败:', error)
  }
}

// 添加/删除背包关键字
const addBagKeyword = () => {
  formData.bagKeywords.push('')
}

const removeBagKeyword = (index) => {
  if (formData.bagKeywords.length > 1) {
    formData.bagKeywords.splice(index, 1)
  }
}


const addLogKeyword = () => {
  formData.LogKeywords.push('')
}

const removeLogKeyword = (index) => {
  if (formData.LogKeywords.length > 1) {
    formData.LogKeywords.splice(index, 1)
  }
}


const addBackup = () => {
  formData.backups.push('')
}

const removeBackup = (index) => {
  if (formData.backups.length > 1) {
    formData.backups.splice(index, 1)
  }
}

const addOneRemoteKeyword = () => {
  formData.OneRemote.LogKeywords.push('')
}
const removeOneRemoteKeyword = (index) => {
  if (formData.OneRemote.LogKeywords.length > 1) {
    formData.OneRemote.LogKeywords.splice(index, 1)
  }
}

// 加载配置
const loadConfig = async () => {
  try {
    const response = await apiMethods.getConfig()
    const data = response.data

    if (data) {
      formData.BetterGIAddress = data.BetterGIAddress || ''
      formData.webhookURL = data.webhookURL || ''
      formData.content = data.content || ''
      formData.post = (data.post || '').replace(':', '')
      
      formData.ConfigNames = data.ConfigNames || new Array(7).fill('')
      
      if (data.BagStatistics) {
        formData.bagKeywords = data.BagStatistics.split(',').map(k => k.trim()).filter(k => k)
      }
      if (formData.bagKeywords.length === 0) {
        formData.bagKeywords = ['']
      }


      //日志通知关键字
      if (data.LogKeywords && Array.isArray(data.LogKeywords)) {
        formData.LogKeywords = data.LogKeywords.filter(LogKeywords => LogKeywords)
      }
      if (formData.LogKeywords.length === 0) {
        formData.LogKeywords = ['']
      }

      if (data.backups && Array.isArray(data.backups)) {
        formData.backups = data.backups.filter(backup => backup)
      }
      if (formData.backups.length === 0) {
        formData.backups = ['']
      }

      if (data.OneLong) {
        Object.assign(formData.OneLong, data.OneLong)
      }

      if (data.Control) {
        Object.assign(formData.Control, data.Control)
      }

      if (data.MySign) {
        Object.assign(formData.MySign, data.MySign)
      }

      if (data.OneRemote) {
        Object.assign(formData.OneRemote, data.OneRemote)
      }
    }
  } catch (error) {
    message.error('加载配置失败: ' + error.message)
  }
}

// 提交表单
const handleSubmit = async () => {
  loading.value = true
  try {
    const payload = {
      BetterGIAddress: formData.BetterGIAddress,
      webhookURL: formData.webhookURL,
      content: formData.content,
      ConfigNames: formData.ConfigNames,
      BagStatistics: formData.bagKeywords.filter(k => k.trim()).join(','),
      post: ':' + formData.post,
      backups: formData.backups.filter(backup => backup.trim()),
      LogKeywords: formData.LogKeywords.filter(k => k.trim()).length > 0
        ? formData.LogKeywords.filter(k => k.trim())
        : [''],
      OneLong: formData.OneLong,
      Control: formData.Control,
      MySign: formData.MySign,
      OneRemote: formData.OneRemote
    }

    await apiMethods.updateConfig(payload)
    message.success('保存成功！')
    
    setTimeout(() => {
      router.push('/')
    }, 2000)
  } catch (error) {
    message.error('保存失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchConfigOptions()
  loadConfig()
})
</script>

<style scoped>
.config-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #ffeaf4 0%, #fff0f7 50%, #ffe6f2 100%);
  position: relative;
  overflow-x: hidden;
  padding: 20px;
}

/* 装饰性背景元素 */
.floating-elements {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 1;
}

.floating-heart {
  position: absolute;
  font-size: 20px;
  animation: floatHeart 6s ease-in-out infinite;
  opacity: 0.6;
}

.floating-heart:nth-child(1) { top: 10%; left: 10%; }
.floating-heart:nth-child(2) { top: 20%; right: 15%; }
.floating-heart:nth-child(3) { top: 60%; left: 5%; }
.floating-heart:nth-child(4) { top: 80%; right: 10%; }
.floating-heart:nth-child(5) { top: 40%; left: 80%; }
.floating-heart:nth-child(6) { top: 70%; right: 5%; }

.floating-star {
  position: absolute;
  font-size: 16px;
  animation: floatStar 8s ease-in-out infinite;
  opacity: 0.4;
}

.floating-star:nth-child(7) { top: 15%; left: 20%; }
.floating-star:nth-child(8) { top: 45%; right: 20%; }
.floating-star:nth-child(9) { top: 75%; left: 15%; }
.floating-star:nth-child(10) { top: 25%; right: 5%; }

@keyframes floatHeart {
  0%, 100% { transform: translateY(0px) rotate(0deg); }
  50% { transform: translateY(-20px) rotate(180deg); }
}

@keyframes floatStar {
  0%, 100% { transform: translateX(0px) rotate(0deg); opacity: 0.4; }
  50% { transform: translateX(15px) rotate(90deg); opacity: 0.8; }
}

.content-wrapper {
  position: relative;
  z-index: 2;
  max-width: 1200px;
  margin: 0 auto;
}

.page-title {
  color: #ff5599 !important;
  text-align: center;
  margin-bottom: 40px;
  position: relative;
  text-shadow: 0 2px 4px rgba(255, 85, 153, 0.2);
  animation: titleGlow 3s ease-in-out infinite alternate;
  font-size: 32px !important;
  font-weight: 700;
}

.title-icon {
  display: inline-block;
  margin: 0 15px;
  animation: bounce 2s ease-in-out infinite;
}

@keyframes titleGlow {
  0% { text-shadow: 0 2px 4px rgba(255, 85, 153, 0.2); }
  100% { text-shadow: 0 2px 8px rgba(255, 85, 153, 0.4), 0 0 20px rgba(255, 85, 153, 0.1); }
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-5px); }
}

.config-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.config-card {
  border-radius: 24px !important;
  border: 2px solid rgba(255, 153, 204, 0.3) !important;
  box-shadow: 
    0 12px 40px rgba(255, 105, 180, 0.15),
    0 4px 12px rgba(255, 105, 180, 0.1) !important;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  backdrop-filter: blur(15px);
  background: rgba(255, 255, 255, 0.95) !important;
  position: relative;
  overflow: hidden;
  animation: slideInUp 0.6s ease-out;
}

.config-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 6px;
  background: linear-gradient(90deg, #ff66aa, #ff99cc, #ffb3d9, #ff66aa);
  background-size: 200% 100%;
  animation: gradientMove 3s ease-in-out infinite;
}

@keyframes gradientMove {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

.config-card:hover {
  transform: translateY(-6px);
  box-shadow: 
    0 20px 60px rgba(255, 105, 180, 0.25),
    0 8px 20px rgba(255, 105, 180, 0.2);
  border-color: rgba(255, 153, 204, 0.6) !important;
}

:deep(.ant-card-head) {
  background: linear-gradient(135deg, rgba(255, 240, 247, 0.9), rgba(255, 250, 253, 0.95)) !important;
  border-bottom: 2px solid rgba(255, 193, 221, 0.4) !important;
  border-radius: 24px 24px 0 0 !important;
  padding: 20px 24px 16px;
}

:deep(.ant-card-head-title) {
  color: #ff5599;
  font-size: 20px;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 12px;
}

.card-icon {
  font-size: 24px;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}

.card-extra {
  display: flex;
  align-items: center;
  gap: 8px;
}

.help-icon-btn {
  color: #ff99cc;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.help-icon-btn:hover {
  color: #ff66aa;
  transform: scale(1.1);
}

:deep(.ant-card-body) {
  background: rgba(255, 255, 255, 0.98);
  padding: 28px;
}

:deep(.ant-form-item-label > label) {
  color: #2c3e50;
  font-weight: 600;
  font-size: 15px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  margin-bottom: 8px;
  display: block;
}

:deep(.ant-tooltip-inner) {
  background: linear-gradient(135deg, #ff66aa, #ff99cc);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(255, 102, 170, 0.3);
}

:deep(.ant-tooltip-arrow-content) {
  background: #ff99cc;
}

:deep(.ant-input-group-addon) {
  background: linear-gradient(135deg, rgba(255, 240, 247, 0.8), rgba(255, 250, 253, 0.9));
  border: 1px solid rgba(255, 153, 204, 0.3);
  border-radius: 12px 0 0 12px;
  color: #ff5599;
  font-weight: 500;
}

:deep(.ant-form-item-explain) {
  color: #666;
  font-size: 12px;
  margin-top: 4px;
  padding-left: 4px;
}

:deep(.ant-btn) {
  border-radius: 12px;
  font-weight: 500;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

:deep(.ant-btn-primary) {
  background: linear-gradient(135deg, #ff66aa, #ff99cc);
  border: none;
  box-shadow: 0 4px 15px rgba(255, 102, 170, 0.3);
  color: white;
}

:deep(.ant-btn-primary):hover {
  background: linear-gradient(135deg, #ff3388, #ff66aa) !important;
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(255, 102, 170, 0.4) !important;
}

:deep(.ant-btn-primary):active {
  transform: translateY(-1px);
}

:deep(.ant-btn-dashed) {
  border: 2px dashed rgba(255, 153, 204, 0.5);
  background: linear-gradient(135deg, rgba(255, 240, 247, 0.3), rgba(255, 250, 253, 0.5));
  color: #ff5599;
  font-weight: 500;
}

:deep(.ant-btn-dashed):hover {
  border-color: #ff99cc !important;
  background: linear-gradient(135deg, rgba(255, 240, 247, 0.6), rgba(255, 250, 253, 0.8)) !important;
  color: #ff3388 !important;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(255, 153, 204, 0.2);
}

:deep(.ant-btn-primary.ant-btn-dangerous) {
  background: linear-gradient(135deg, #ff6b6b, #ff8e8e);
  border: none;
}

:deep(.ant-btn-primary.ant-btn-dangerous):hover {
  background: linear-gradient(135deg, #ff5252, #ff6b6b) !important;
  transform: translateY(-2px);
}

:deep(.ant-checkbox-wrapper) {
  color: #333;
  font-weight: 500;
  transition: all 0.3s ease;
}

:deep(.ant-checkbox-wrapper):hover {
  color: #ff5599;
}

:deep(.ant-checkbox-checked::after) {
  border-color: #ff99cc;
}

/* 表单网格布局 */
.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;
}

.form-item-enhanced {
  margin-bottom: 0 !important;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 12px;
  z-index: 2;
  font-size: 16px;
  color: #ff99cc;
}

.enhanced-input {
  padding-left: 40px !important;
  height: 48px !important;
  border-radius: 16px !important;
  border: 2px solid rgba(255, 153, 204, 0.3) !important;
  background: linear-gradient(135deg, rgba(255, 240, 247, 0.8), rgba(255, 250, 253, 0.9)) !important;
  transition: all 0.3s ease;
  font-size: 14px;
}

.enhanced-input:focus {
  border-color: #ff99cc !important;
  box-shadow: 0 0 0 3px rgba(255, 153, 204, 0.2) !important;
  background: rgba(255, 255, 255, 0.95) !important;
  transform: translateY(-2px);
}

.enhanced-input:hover {
  border-color: #ff99cc !important;
  background: rgba(255, 255, 255, 0.9) !important;
  transform: translateY(-1px);
}

.help-text {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  color: #666;
  font-size: 12px;
  padding: 8px 12px;
  background: rgba(255, 240, 247, 0.5);
  border-radius: 8px;
  border-left: 3px solid #ff99cc;
}

.help-icon {
  font-size: 14px;
}

/* 端口输入组 */
.port-input-group {
  display: flex;
  align-items: center;
}

.port-separator {
  background: linear-gradient(135deg, rgba(255, 240, 247, 0.8), rgba(255, 250, 253, 0.9)) !important;
  border: 2px solid rgba(255, 153, 204, 0.3) !important;
  border-radius: 16px 0 0 16px !important;
  color: #ff5599;
  font-weight: 600;
  text-align: center;
}

/* 星期网格 */
.weekday-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-top: 16px;
}

.weekday-item {
  background: linear-gradient(135deg, rgba(255, 240, 247, 0.6), rgba(255, 250, 253, 0.8));
  border-radius: 16px;
  padding: 16px;
  border: 2px solid rgba(255, 153, 204, 0.2);
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.weekday-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, #ff66aa, #ff99cc);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.weekday-item.active::before {
  opacity: 1;
}

.weekday-item:hover {
  transform: translateY(-2px);
  border-color: rgba(255, 153, 204, 0.4);
  box-shadow: 0 8px 20px rgba(255, 153, 204, 0.2);
}

.weekday-label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  font-weight: 600;
  color: #ff5599;
  font-size: 14px;
}

.day-icon {
  font-size: 18px;
  animation: rotate 3s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.weekday-select {
  width: 100%;
}

.weekday-select.has-value :deep(.ant-select-selector) {
  border-color: #ff99cc !important;
  background: rgba(255, 255, 255, 0.95) !important;
}

:deep(.ant-select-selector) {
  height: 44px !important;
  border-radius: 12px !important;
  border: 2px solid rgba(255, 153, 204, 0.3) !important;
  background: linear-gradient(135deg, rgba(255, 240, 247, 0.8), rgba(255, 250, 253, 0.9)) !important;
  transition: all 0.3s ease;
}

:deep(.ant-select:hover .ant-select-selector) {
  border-color: #ff99cc !important;
  background: rgba(255, 255, 255, 0.9) !important;
}

:deep(.ant-select-focused .ant-select-selector) {
  border-color: #ff99cc !important;
  box-shadow: 0 0 0 3px rgba(255, 153, 204, 0.2) !important;
  background: rgba(255, 255, 255, 0.95) !important;
}

/* 动态列表 */
.dynamic-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.list-item {
  display: flex;
  align-items: center;
  gap: 12px;
  background: linear-gradient(135deg, rgba(255, 240, 247, 0.6), rgba(255, 250, 253, 0.8));
  border-radius: 16px;
  padding: 12px;
  border: 2px solid rgba(255, 153, 204, 0.2);
  transition: all 0.3s ease;
}

.list-item:hover {
  transform: translateX(4px);
  border-color: rgba(255, 153, 204, 0.4);
  box-shadow: 0 4px 12px rgba(255, 153, 204, 0.2);
}

.item-content {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
}

.item-icon {
  font-size: 16px;
  color: #ff99cc;
}

.remove-btn {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #ff6b6b, #ff8e8e) !important;
  border: none;
  transition: all 0.3s ease;
}

.remove-btn:hover {
  background: linear-gradient(135deg, #ff5252, #ff6b6b) !important;
  transform: scale(1.1);
}

.add-btn {
  border: 2px dashed rgba(255, 153, 204, 0.5);
  background: linear-gradient(135deg, rgba(255, 240, 247, 0.3), rgba(255, 250, 253, 0.5));
  color: #ff5599;
  font-weight: 600;
  border-radius: 16px;
  height: 48px;
  transition: all 0.3s ease;
}

.add-btn:hover {
  border-color: #ff99cc !important;
  background: linear-gradient(135deg, rgba(255, 240, 247, 0.6), rgba(255, 250, 253, 0.8)) !important;
  color: #ff3388 !important;
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(255, 153, 204, 0.3);
}

/* 时间设置 */
.time-settings-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.checkbox-item {
  margin-bottom: 0 !important;
}

.enhanced-checkbox {
  display: flex;
  align-items: center;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #333;
}

.checkbox-icon {
  font-size: 16px;
  color: #ff99cc;
}

:deep(.ant-checkbox) {
  transform: scale(1.2);
}

:deep(.ant-checkbox-checked .ant-checkbox-inner) {
  background: linear-gradient(135deg, #ff66aa, #ff99cc);
  border-color: #ff99cc;
}

:deep(.ant-checkbox:hover .ant-checkbox-inner) {
  border-color: #ff99cc;
  background: rgba(255, 153, 204, 0.1);
}

.time-inputs {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-top: 16px;
  padding: 20px;
  background: linear-gradient(135deg, rgba(255, 240, 247, 0.4), rgba(255, 250, 253, 0.6));
  border-radius: 16px;
  border: 2px solid rgba(255, 153, 204, 0.2);
}

.time-form-item {
  margin-bottom: 0 !important;
}

.enhanced-input-number {
  width: 100% !important;
  height: 48px !important;
  border-radius: 16px !important;
  border: 2px solid rgba(255, 153, 204, 0.3) !important;
  background: linear-gradient(135deg, rgba(255, 240, 247, 0.8), rgba(255, 250, 253, 0.9)) !important;
  transition: all 0.3s ease;
}

.enhanced-input-number:hover {
  border-color: #ff99cc !important;
  background: rgba(255, 255, 255, 0.9) !important;
}

.enhanced-input-number:focus {
  border-color: #ff99cc !important;
  box-shadow: 0 0 0 3px rgba(255, 153, 204, 0.2) !important;
  background: rgba(255, 255, 255, 0.95) !important;
}

/* 控制选项 */
.control-options {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 米游社签到 */
.mys-sign-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 提交按钮 */
.submit-section {
  text-align: center;
  margin-top: 40px;
  padding: 20px;
}

.submit-btn {
  background: linear-gradient(135deg, #ff66aa, #ff99cc, #ffb3d9) !important;
  border: none !important;
  border-radius: 24px !important;
  font-size: 18px !important;
  font-weight: 700 !important;
  padding: 20px 60px !important;
  height: auto !important;
  box-shadow: 
    0 12px 35px rgba(255, 102, 170, 0.4),
    0 6px 15px rgba(255, 102, 170, 0.2) !important;
  position: relative;
  overflow: hidden;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.submit-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.4), transparent);
  transition: left 0.6s ease;
}

.submit-btn:hover::before {
  left: 100%;
}

.submit-btn:hover {
  background: linear-gradient(135deg, #ff3388, #ff66aa, #ff99cc) !important;
  transform: translateY(-4px) scale(1.02);
  box-shadow: 
    0 16px 45px rgba(255, 102, 170, 0.5),
    0 8px 20px rgba(255, 102, 170, 0.3) !important;
}

.submit-btn:active {
  transform: translateY(-2px) scale(1.01);
}

.submit-icon {
  display: inline-block;
  margin: 0 8px;
  animation: bounce 2s ease-in-out infinite;
}

/* 动画效果 */
.config-card {
  animation: slideInUp 0.6s ease-out;
}

.config-card:nth-child(1) { animation-delay: 0.1s; }
.config-card:nth-child(2) { animation-delay: 0.2s; }
.config-card:nth-child(3) { animation-delay: 0.3s; }
.config-card:nth-child(4) { animation-delay: 0.4s; }
.config-card:nth-child(5) { animation-delay: 0.5s; }
.config-card:nth-child(6) { animation-delay: 0.6s; }
.config-card:nth-child(7) { animation-delay: 0.7s; }

@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(40px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .config-container {
    padding: 16px;
  }
  
  .page-title {
    font-size: 24px !important;
    margin-bottom: 24px;
  }
  
  .config-card {
    margin-bottom: 20px;
    border-radius: 20px !important;
  }
  
  :deep(.ant-card-body) {
    padding: 20px;
  }
  
  .form-grid {
    grid-template-columns: 1fr;
    gap: 20px;
  }
  
  .weekday-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  
  .time-inputs {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  
  .submit-btn {
    padding: 16px 40px !important;
    font-size: 16px !important;
  }
  
  .floating-elements {
    display: none;
  }
}

/* 深色模式支持 */
@media (prefers-color-scheme: dark) {
  .config-container {
    background: linear-gradient(135deg, #2a1a2e 0%, #16213e 50%, #1a1a2e 100%);
  }
  
  .config-card {
    background: rgba(40, 44, 52, 0.95) !important;
    border-color: rgba(255, 153, 204, 0.3) !important;
  }
  
  :deep(.ant-card-body) {
    background: rgba(40, 44, 52, 0.98);
    color: #e6e6e6;
  }
  
  .help-text {
    background: rgba(255, 240, 247, 0.1);
    color: #ccc;
  }
}
</style>