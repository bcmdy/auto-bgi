<template>
  <div class="config-container">
    <div class="floating-elements">
      <div class="floating-heart" v-for="i in 6" :key="`heart-${i}`" :style="{ animationDelay: `${i * 0.5}s` }">💖</div>
      <div class="floating-star" v-for="i in 4" :key="`star-${i}`" :style="{ animationDelay: `${i * 0.7}s` }">✨</div>
    </div>

    <div class="content-wrapper">
      <h1 class="page-title">
        <span class="title-icon">🌸</span>
        配置设置
        <span class="title-icon">🌸</span>
      </h1>

      <a-form :model="formData" layout="vertical" @finish="handleSubmit" ref="formRef" class="config-form">
        
        <a-card title="基础设置" class="config-card">
          <template #extra><span class="card-icon">⚙️</span></template>
          <a-row :gutter="24">
            <a-col :xs="24" :md="12">
               <a-form-item label="BetterGI 本地路径" name="BetterGIAddress" class="form-item-enhanced">
                  <a-input v-model:value="formData.BetterGIAddress" placeholder="例如 D:\subject\lua\BetterGI" class="enhanced-input">
                    <template #prefix><span class="input-icon">📂</span></template>
                  </a-input>
                  <div class="help-text">💡 填写路径，不能包含逗号，请使用单斜杠</div>
               </a-form-item>
            </a-col>
            <a-col :xs="24" :md="12">
              <a-form-item label="端口号" name="post" class="form-item-enhanced">
                <a-input v-model:value="formData.post" placeholder="默认 8082" class="enhanced-input">
                  <template #prefix><span class="input-icon">🔌</span></template>
                </a-input>
              </a-form-item>
            </a-col>
            <a-col :span="24">
              <a-form-item label="BGI 关闭提示内容" name="content" class="form-item-enhanced">
                <a-textarea v-model:value="formData.content" :rows="3" placeholder="BGI关闭后发送到企业微信的内容" class="enhanced-textarea" />
                <div class="help-text">💡 可填写网页链接，例如背包地址</div>
              </a-form-item>
            </a-col>
          </a-row>
        </a-card>

        <a-row :gutter="24">
          <!-- <a-col :xs="24" :lg="12">
            <a-card title="背包统计" class="config-card">
              <template #extra>
                <div class="card-extra">
                  <span class="card-icon">🎒</span>
                  <a-tooltip title="配合仓库背包统计、摩拉OCR使用">
                    <QuestionCircleOutlined class="help-icon-btn" />
                  </a-tooltip>
                </div>
              </template>
              <div class="dynamic-list">
                <div v-for="(item, index) in formData.bagKeywords" :key="`bag-${index}`" class="list-item-wrapper">
                  <a-input v-model:value="formData.bagKeywords[index]" placeholder="输入材料名称" class="enhanced-input">
                    <template #prefix>💎</template>
                  </a-input>
                  <a-button type="primary" danger shape="circle" @click="removeBagKeyword(index)" v-if="formData.bagKeywords.length > 1" class="icon-btn">
                    <DeleteOutlined />
                  </a-button>
                </div>
                <a-button type="dashed" block @click="addBagKeyword" class="add-btn">
                  <PlusOutlined /> 添加材料
                </a-button>
              </div>
            </a-card>
          </a-col> -->

          <a-col :xs="24" :lg="24">
            <a-card title="关注关键字" class="config-card">
              <template #extra>
                <div class="card-extra">
                  <span class="card-icon">📜</span>
                  <a-tooltip title="触发关键词将通知企业微信">
                    <QuestionCircleOutlined class="help-icon-btn" />
                  </a-tooltip>
                </div>
              </template>
              <div class="dynamic-list">
                <div v-for="(item, index) in formData.LogKeywords" :key="`log-${index}`" class="list-item-wrapper">
                  <a-input v-model:value="formData.LogKeywords[index]" placeholder="输入日志关键字" class="enhanced-input">
                    <template #prefix>🔑</template>
                  </a-input>
                  <a-button type="primary" danger shape="circle" @click="removeLogKeyword(index)" v-if="formData.LogKeywords.length > 1" class="icon-btn">
                    <DeleteOutlined />
                  </a-button>
                </div>
                <a-button type="dashed" block @click="addLogKeyword" class="add-btn">
                  <PlusOutlined /> 添加关键字
                </a-button>
              </div>
            </a-card>
          </a-col>
        </a-row>

        <a-card title="功能开关" class="config-card">
          <template #extra><span class="card-icon">🎮</span></template>
          <div class="switch-grid">
            <div class="switch-item">
               <a-checkbox v-model:checked="formData.OneLong.AutoUpdateJs" class="enhanced-checkbox">🔄 自动更新 JS</a-checkbox>
            </div>
            <div class="switch-item">
               <a-checkbox v-model:checked="formData.Control.OBSReplayBuffer" class="enhanced-checkbox">📼 OBS 重放缓冲</a-checkbox>
            </div>
            <div class="switch-item">
               <a-checkbox v-model:checked="formData.Control.SendWeChatImage" class="enhanced-checkbox">📸 每小时发截图</a-checkbox>
            </div>
            <div class="switch-item">
               <a-checkbox v-model:checked="formData.Control.AbgiScreen" class="enhanced-checkbox">🖥️ 实时屏幕 (高功耗)</a-checkbox>
            </div>
            <div class="switch-item">
               <a-checkbox v-model:checked="formData.Control.IsCloseYuanShen" class="enhanced-checkbox">❌ BGI关闭时关闭原神</a-checkbox>
            </div>
          </div>
        </a-card>

        <a-card title="1Remote 远程监控" class="config-card">
          <template #extra><span class="card-icon">💻</span></template>
          <a-form-item class="checkbox-item">
            <a-checkbox v-model:checked="formData.OneRemote.IsMonitor" class="enhanced-checkbox">启用远程监控</a-checkbox>
          </a-form-item>
          <div v-if="formData.OneRemote.IsMonitor" class="fade-in-section">
            <a-form-item label="日志文件夹路径" class="form-item-enhanced">
              <a-input v-model:value="formData.OneRemote.LogFilePath" placeholder="例如 ...\.logs" class="enhanced-input">
                <template #prefix>📂</template>
              </a-input>
            </a-form-item>
            </div>
        </a-card>

        <a-card title="录屏设置 (OBS)" class="config-card">
          <template #extra>
            <div class="card-extra">
              <span class="card-icon">🔍</span>
              <a-tooltip title="需开启OBS Websocket">
                <QuestionCircleOutlined class="help-icon-btn" />
              </a-tooltip>
            </div>
          </template>
          <a-form-item class="checkbox-item">
            <a-checkbox v-model:checked="formData.ScreenRecord.IsRecord" class="enhanced-checkbox">启用录屏功能</a-checkbox>
          </a-form-item>
          <a-row :gutter="24" v-if="formData.ScreenRecord.IsRecord" class="fade-in-section">
            <a-col :xs="24" :md="12">
              <a-form-item label="OBS 地址" class="form-item-enhanced">
                <a-input-password v-model:value="formData.ScreenRecord.StartScreen" placeholder="ws://..." class="enhanced-input" />
              </a-form-item>
            </a-col>
             <a-col :xs="24" :md="12">
              <a-form-item label="密码" class="form-item-enhanced">
                <a-input-password v-model:value="formData.ScreenRecord.EndScreen" placeholder="根据后端需求填写" class="enhanced-input" />
              </a-form-item>
            </a-col>
          </a-row>
        </a-card>

        <a-card title="消息通知" class="config-card">
          <template #extra><span class="card-icon">📢</span></template>
          
          <a-form-item label="通知类型" class="form-item-enhanced">
            <a-select v-model:value="formData.Notice.Type" class="enhanced-select">
              <a-select-option value="Wechat">企业微信 (Wechat)</a-select-option>
              <a-select-option value="TG">Telegram</a-select-option>
              <a-select-option value="FeiShu">飞书 (FeiShu)</a-select-option>
              <a-select-option value="OneBot">OneBot</a-select-option>
            </a-select>
          </a-form-item>

          <div v-if="formData.Notice.Type === 'Wechat'" class="fade-in-section">
            <a-form-item label="Webhook 地址" class="form-item-enhanced">
              <a-input-password v-model:value="formData.Notice.Wechat" class="enhanced-input"><template #prefix>💬</template></a-input-password>
            </a-form-item>
          </div>

          <div v-if="formData.Notice.Type === 'TG'" class="fade-in-section">
            <a-form-item label="Bot Token" class="form-item-enhanced">
              <a-input-password v-model:value="formData.Notice.TGNotice.TGToken" class="enhanced-input"><template #prefix>🤖</template></a-input-password>
            </a-form-item>
            <a-form-item label="Chat ID" class="form-item-enhanced">
              <a-input-password v-model:value="formData.Notice.TGNotice.ChatID" class="enhanced-input"><template #prefix>🆔</template></a-input-password>
            </a-form-item>
             <a-form-item label="代理地址" class="form-item-enhanced">
              <a-input-password v-model:value="formData.Notice.TGNotice.Proxy" class="enhanced-input"><template #prefix>🌐</template></a-input-password>
            </a-form-item>
          </div>

          <div v-if="formData.Notice.Type === 'FeiShu'" class="fade-in-section">
             <a-form-item label="Webhook URL" class="form-item-enhanced">
              <a-input-password v-model:value="formData.Notice.FeiShu.FeiShuWebhookURL" class="enhanced-input"><template #prefix>🔗</template></a-input-password>
            </a-form-item>
            <a-form-item label="App ID" class="form-item-enhanced">
              <a-input-password v-model:value="formData.Notice.FeiShu.AppID" class="enhanced-input"><template #prefix>🆔</template></a-input-password>
            </a-form-item>
            <a-form-item label="App Secret" class="form-item-enhanced">
              <a-input-password v-model:value="formData.Notice.FeiShu.AppSecret" class="enhanced-input"><template #prefix>🔑</template></a-input-password>
            </a-form-item>
          </div>

          <div v-if="formData.Notice.Type === 'OneBot'" class="fade-in-section">
             <a-form-item label="API Base" class="form-item-enhanced">
              <a-input-password v-model:value="formData.Notice.OneBot.APIBase" class="enhanced-input"><template #prefix>🌐</template></a-input-password>
            </a-form-item>
            <a-form-item label="Token" class="form-item-enhanced">
              <a-input-password v-model:value="formData.Notice.OneBot.Token" class="enhanced-input"><template #prefix>🎟️</template></a-input-password>
            </a-form-item>
             <a-form-item label="QQ号" class="form-item-enhanced">
              <a-input-password v-model:value="formData.Notice.OneBot.QQNum" class="enhanced-input"><template #prefix>🐧</template></a-input-password>
            </a-form-item>
          </div>
        </a-card>

        <a-card title="命令机器人" class="config-card">
           <template #extra><span class="card-icon">🤖</span></template>
           <div class="switch-grid">
             <div class="switch-item">
               <a-checkbox v-model:checked="formData.CommandBot.TgBOT" class="enhanced-checkbox">启用 TG 机器人</a-checkbox>
             </div>
             <div class="switch-item">
               <a-checkbox v-model:checked="formData.CommandBot.FeiShuBot" class="enhanced-checkbox">启用飞书机器人</a-checkbox>
             </div>
           </div>
        </a-card>

        <a-card title="联机设置" class="config-card account-card">
          <template #extra><span class="card-icon">🔗</span></template>
          <a-row :gutter="24">
            <a-col :xs="24" :md="12">
              <a-form-item label="UID" class="form-item-enhanced">
                <a-input-password v-model:value="formData.Account.Uid" class="enhanced-input"><template #prefix>🆔</template></a-input-password>
              </a-form-item>
            </a-col>
            <a-col :xs="24" :md="12">
              <a-form-item label="旅行者名字" class="form-item-enhanced">
                <a-input-password v-model:value="formData.Account.Name" class="enhanced-input"><template #prefix>🥵</template></a-input-password>
              </a-form-item>
            </a-col>
          </a-row>
          
          <div class="highlight-box">
             <a-checkbox v-model:checked="formData.Account.IsMultiUser" class="enhanced-checkbox">批发是否是多账号</a-checkbox>
          </div>

          <a-row :gutter="24">
             <a-col :xs="24" :md="12">
              <a-form-item label="狗粮联机配置组" class="form-item-enhanced">
                <a-input v-model:value="formData.Account.GouLangGroupName" class="enhanced-input"><template #prefix>🐶</template></a-input>
              </a-form-item>
            </a-col>
            <a-col :xs="24" :md="12">
              <a-form-item label="联机 SecretKey" class="form-item-enhanced">
                <a-input-password v-model:value="formData.Account.SecretKey" class="enhanced-input"><template #prefix>🔑</template></a-input-password>
              </a-form-item>
            </a-col>
          </a-row>
        </a-card>

        <div class="submit-section">
          <a-button type="primary" size="large" @click="handleSubmit" :loading="loading" class="submit-btn">
            <span class="submit-icon">💾</span> 保存所有配置 <span class="submit-icon">💾</span>
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
// 假设这是你的API路径，保持不变
import { apiMethods } from '@/utils/api'

const router = useRouter()
const formRef = ref()
const loading = ref(false)
const configOptions = ref([])

// 表单数据 - 保持你的原始结构不变
const formData = reactive({
  BetterGIAddress: '',
  content: '',
  post: '10086',
  bagKeywords: [''],
  LogKeywords: [''],
  OneLong: {
    AutoUpdateJs: true
  },
  Control: {
    IsCloseYuanShen: false,
    SendWeChatImage: false,
    AbgiScreen: false,
    OBSReplayBuffer: false
  },
  OneRemote: {
    IsMonitor: false,
    LogFilePath: '',
    LogKeywords: ['']
  },
  ScreenRecord: {
    IsRecord: false,
    StartScreen: '',
    EndScreen: '',
  },
  BgiLog: '',
  basePath: '',
  Notice: {
    Type: 'Wechat',
    Wechat: '',
    TGNotice: {
      TGToken: '',
      ChatID: 0,
      Proxy: ''
    },
    OneBot: {
      APIBase: "",
      Token: "",
      QQNum: 0,
      groupNum: 0
    },
    FeiShu: {
      FeiShuWebhookURL: '',
      AppID: '',
      AppSecret: ''
    },
  },
  RepoUrl: "",
  Account: {
    Uid: "",
    Name: "",
    IsMultiUser: false,
    GouLangGroupName: "",
    SecretKey: "",
  },
  CommandBot: {
    TgBOT: false,
    FeiShuBot: false,
  },
  AbgiAiConfig: {
    IsAbgiAi: false,
    ApiKey: "",
    ApiUrl: "",
    Model: "",
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

// 动态数组操作
const addBagKeyword = () => { formData.bagKeywords.push('') }
const removeBagKeyword = (index) => {
  if (formData.bagKeywords.length > 1) formData.bagKeywords.splice(index, 1)
}

const addLogKeyword = () => { formData.LogKeywords.push('') }
const removeLogKeyword = (index) => {
  if (formData.LogKeywords.length > 1) formData.LogKeywords.splice(index, 1)
}

// 加载配置
const loadConfig = async () => {
  try {
    const response = await apiMethods.getConfig()
    const data = response.data
    console.log(data)

    if (data) {
      formData.BetterGIAddress = data.BetterGIAddress || ''
      formData.content = data.content || ''
      formData.post = (data.post || '').replace(':', '')
      formData.basePath = data.basePath || ''
      formData.BgiLog = data.BgiLog || ''
      formData.RepoUrl = data.RepoUrl || ''
      
      if (data.BagStatistics) {
        formData.bagKeywords = data.BagStatistics.split(',').map(k => k.trim()).filter(k => k)
      }
      if (formData.bagKeywords.length === 0) formData.bagKeywords = ['']

      if (data.LogKeywords && Array.isArray(data.LogKeywords)) {
        formData.LogKeywords = data.LogKeywords.filter(LogKeywords => LogKeywords)
      }
      if (formData.LogKeywords.length === 0) formData.LogKeywords = ['']

      if (data.OneLong) Object.assign(formData.OneLong, data.OneLong)
      if (data.Control) Object.assign(formData.Control, data.Control)
      if (data.OneRemote) Object.assign(formData.OneRemote, data.OneRemote)
      if (data.ScreenRecord) Object.assign(formData.ScreenRecord, data.ScreenRecord)
      if (data.Notice) Object.assign(formData.Notice, data.Notice)
      if (data.Account) Object.assign(formData.Account, data.Account)
      if (data.CommandBot) Object.assign(formData.CommandBot, data.CommandBot)
      if (data.AbgiAiConfig) Object.assign(formData.AbgiAiConfig, data.AbgiAiConfig)
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
      content: formData.content,
      BagStatistics: formData.bagKeywords.filter(k => k.trim()).join(','),
      post: ':' + formData.post,
      LogKeywords: formData.LogKeywords.filter(k => k.trim()).length > 0
        ? formData.LogKeywords.filter(k => k.trim())
        : [''],
      OneLong: formData.OneLong,
      Control: formData.Control,
      OneRemote: formData.OneRemote,
      ScreenRecord: formData.ScreenRecord,
      BgiLog: formData.BgiLog,
      basePath: formData.basePath,
      Notice: formData.Notice,
      Account: formData.Account,
      RepoUrl: formData.RepoUrl,
      CommandBot: formData.CommandBot,
      AbgiAiConfig: formData.AbgiAiConfig
    }

    console.log('提交的配置:', payload)
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
/* 全局容器 - 增加渐变背景 */
.config-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #ffeaf4 0%, #fff0f7 50%, #ffe6f2 100%);
  position: relative;
  overflow-x: hidden;
  padding: 40px 20px;
}

/* 装饰性背景元素 */
.floating-elements {
  position: fixed;
  top: 0; left: 0; width: 100%; height: 100%;
  pointer-events: none; z-index: 1;
}

.floating-heart {
  position: absolute;
  font-size: 24px;
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
  font-size: 18px;
  animation: floatStar 8s ease-in-out infinite;
  opacity: 0.5;
}
.floating-star:nth-child(1) { top: 15%; left: 20%; }
.floating-star:nth-child(2) { top: 45%; right: 20%; }
.floating-star:nth-child(3) { top: 75%; left: 15%; }
.floating-star:nth-child(4) { top: 25%; right: 5%; }

@keyframes floatHeart {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-25px) rotate(15deg); }
}

@keyframes floatStar {
  0%, 100% { transform: scale(1) rotate(0deg); opacity: 0.5; }
  50% { transform: scale(1.3) rotate(180deg); opacity: 0.9; }
}

/* 内容区域 */
.content-wrapper {
  position: relative;
  z-index: 2;
  max-width: 1000px;
  margin: 0 auto;
}

.page-title {
  color: #ff5599;
  text-align: center;
  margin-bottom: 40px;
  font-size: 32px;
  font-weight: 800;
  text-shadow: 2px 2px 0px rgba(255, 255, 255, 0.8);
  animation: titlePulse 2s infinite ease-in-out;
}
@keyframes titlePulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.05); }
}

.title-icon { margin: 0 10px; font-size: 28px; }

/* 卡片样式 */
.config-card {
  border-radius: 20px !important;
  border: 2px solid rgba(255, 255, 255, 0.8) !important;
  box-shadow: 0 10px 30px rgba(255, 182, 193, 0.3) !important;
  background: rgba(255, 255, 255, 0.75) !important;
  backdrop-filter: blur(12px);
  margin-bottom: 24px;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.config-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 15px 40px rgba(255, 105, 180, 0.25) !important;
  border-color: #ffb6c1 !important;
}

:deep(.ant-card-head) {
  border-bottom: 1px solid rgba(255, 105, 180, 0.1) !important;
  font-size: 18px;
  color: #ff66aa;
  font-weight: bold;
}

:deep(.ant-card-body) {
  padding: 24px;
}

/* 输入框增强 */
.enhanced-input, .enhanced-textarea, .enhanced-select :deep(.ant-select-selector) {
  border-radius: 12px !important;
  border: 1px solid #ffd1dc !important;
  background: rgba(255, 255, 255, 0.9) !important;
  box-shadow: inset 0 2px 4px rgba(0,0,0,0.02);
  transition: all 0.3s;
}

.enhanced-input:hover, .enhanced-input:focus,
.enhanced-textarea:hover, .enhanced-textarea:focus,
.enhanced-select:hover :deep(.ant-select-selector) {
  border-color: #ff88bb !important;
  box-shadow: 0 0 0 3px rgba(255, 136, 187, 0.2) !important;
}

.input-icon { margin-right: 8px; font-size: 16px; }

/* 帮助文本 */
.help-text {
  font-size: 12px;
  color: #ff88aa;
  margin-top: 6px;
  margin-left: 4px;
}

/* 动态列表 */
.dynamic-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.list-item-wrapper {
  display: flex;
  gap: 10px;
  align-items: center;
}
.icon-btn {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}
.add-btn {
  border-radius: 12px;
  border-style: dashed;
  border-color: #ff99cc;
  color: #ff66aa;
}
.add-btn:hover {
  border-color: #ff3388;
  color: #ff3388;
}

/* 增强Checkbox */
.enhanced-checkbox {
  font-size: 15px;
  color: #555;
  transition: color 0.3s;
  padding: 8px 0;
}
.enhanced-checkbox:hover {
  color: #ff5599;
}
:deep(.ant-checkbox-inner) {
  border-radius: 6px;
  border-color: #ff99cc;
}
:deep(.ant-checkbox-checked .ant-checkbox-inner) {
  background-color: #ff66aa;
  border-color: #ff66aa;
}

/* 开关网格 */
.switch-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}

/* 特殊高亮盒子 */
.highlight-box {
  margin: 16px 0;
  padding: 16px;
  border: 2px dashed #ff99cc;
  border-radius: 16px;
  background: rgba(255, 240, 245, 0.5);
  text-align: center;
}

/* 提交按钮 */
.submit-section {
  text-align: center;
  margin-top: 40px;
  margin-bottom: 40px;
}
.submit-btn {
  background: linear-gradient(90deg, #ff88bb, #ff6699);
  border: none;
  height: 56px;
  padding: 0 50px;
  font-size: 18px;
  border-radius: 28px;
  box-shadow: 0 10px 25px rgba(255, 102, 153, 0.4);
  transition: all 0.3s;
}
.submit-btn:hover {
  transform: translateY(-3px) scale(1.02);
  box-shadow: 0 15px 35px rgba(255, 102, 153, 0.5);
  background: linear-gradient(90deg, #ff77aa, #ff5588);
}

/* 淡入动画 */
.fade-in-section {
  animation: fadeIn 0.4s ease-out;
  margin-top: 10px;
}
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-5px); }
  to { opacity: 1; transform: translateY(0); }
}

/* 响应式调整 */
@media (max-width: 768px) {
  .config-container { padding: 20px 10px; }
  .config-card { padding: 0; }
  :deep(.ant-card-body) { padding: 16px; }
  .page-title { font-size: 24px; }
  .submit-btn { width: 100%; }
}
</style>