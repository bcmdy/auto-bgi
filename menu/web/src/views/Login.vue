<template>
  <div class="login-container">
    <div class="bg-layer"></div>
    <div class="grid-pattern"></div>
    
    <div class="bg-decoration">
      <div class="floating-shape shape-1"></div>
      <div class="floating-shape shape-2"></div>
      <div class="floating-shape shape-3"></div>
      <div class="star star-1">✨</div>
      <div class="star star-2">⭐</div>
      <div class="star star-3">✨</div>
      <div class="star star-4">🌟</div>
    </div>

    <div class="login-card-wrapper">
      <div class="card-ribbon">🎀</div>

      <div class="card-header">
        <h1 class="system-title">{{ systemName }}</h1>
        <div class="subtitle-badge">
          <span>✨ 欢迎回来 Master ✨</span>
        </div>
      </div>

      <a-form
        ref="formRef"
        :model="formState"
        @finish="onFinish"
        @finishFailed="onFinishFailed"
        :label-col="{ span: 0 }"
        :wrapper-col="{ span: 24 }"
        class="login-form"
      >
        <a-form-item
          name="username"
          :rules="[{ required: true, message: '请输入用户名哦~' }]"
        >
          <div class="input-group">
            <span class="input-icon">👤</span>
            <a-input
              v-model:value="formState.username"
              placeholder="请输入用户名..."
              class="kawaii-input"
              :bordered="false"
              @keyup.enter="handleEnter"
            />
          </div>
        </a-form-item>

        <a-form-item
          name="password"
          :rules="[{ required: true, message: '请输入密码哦~' }]"
        >
          <div class="input-group">
            <span class="input-icon">🔐</span>
            <a-input-password
              v-model:value="formState.password"
              placeholder="请输入密码..."
              class="kawaii-input"
              :bordered="false"
              @keyup.enter="handleEnter"
            />
          </div>
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            :loading="loading"
            block
            class="kawaii-button"
          >
            {{ loading ? '少女祈祷中...✨' : '进入异世界 →' }}
          </a-button>
        </a-form-item>

        <transition name="bounce">
          <div v-if="errorMessage" class="error-bubble">
            <span class="error-icon">💢</span> {{ errorMessage }}
          </div>
        </transition>
      </a-form>

      <div class="card-footer">
        <div class="footer-divider"></div>
        <p class="footer-text" @click="aaa">嘿~这是一个神秘的地方呢🎀</p>
        <div class="contact-pill">qq群：215053644</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { apiMethods } from '../utils/api'

const router = useRouter()
const formRef = ref(null)
const loading = ref(false)
const errorMessage = ref('')
const systemName = ref('登录')

const formState = ref({
  username: '',
  password: ''
})

// 在 script setup 顶部添加定义
const isUniappReady = ref(false) //

// 页面挂载时获取系统配置
onMounted(async () => {
  try {
    const response = await apiMethods.getSystemConfig()
    if (response.systemName) {
      systemName.value = response.systemName
    }
  } catch (error) {
    console.error('获取系统配置失败:', error)
  }



// const initUniBridge = () => {
//     isUniappReady.value = true;
//     console.log('✨ UniApp Bridge 已就绪');
//     // 自动握手一次
//     if (window.uni && window.uni.postMessage) {
//       // window.uni.postMessage({ data: { type: '思姐真可爱', msg: 'abgi已经连接' } });
//       console.log('✨ 已向 UniApp 发送握手消息');
//     }
//   };

//   // 2. 同时尝试两种检查方式
//   if (window.UniAppJSBridgeReady) {
//     initUniBridge();
//   } else {
//     document.addEventListener('UniAppJSBridgeReady', initUniBridge);
//   }


})

const handleEnter = () => {
  if (formRef.value) {
    formRef.value.submit()
  }
}


const aaa = () => {
    console.log("Check Uni Object:", window.uni);

      // 在 Uniapp WebView 中，官方 SDK 会挂载 window.uni
      if (window.uni && window.uni.postMessage) {
        window.uni.postMessage({
          data: { 
            action: '思姐真可爱',
            content: '来自神秘地方的数据🎀' 
          }
        });
        message.success('已向异世界发送信号✨');
      } else {
        console.error("【提示】当前不在 UniApp 环境，或 SDK 尚未加载。");
        message.warning('咒语失效了，请在 App 中尝试哦~');
      }
};

const onFinish = async () => {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await apiMethods.login(
      formState.value.username,
      formState.value.password
    )

    if (response.code === 401 || response.error) {
      errorMessage.value = response.error || '登录失败，请检查用户名和密码'
      message.error('登录失败：' + (response.error || '未知错误'))
    } else if (response.code === 200 && response.aBgiToken) {
      localStorage.setItem('aBgiToken', response.aBgiToken)
      message.success('登录成功！')
      router.push('/')
    } else if (response.aBgiToken) {
      localStorage.setItem('aBgiToken', response.aBgiToken)
      message.success('登录成功！')
      router.push('/')
    } else {
      errorMessage.value = '登录失败，请重试'
      message.error('登录失败，请重试')
    }
  } catch (error) {
    errorMessage.value = error.message || '登录失败，请重试'
    message.error('网络错误：' + error.message)
  } finally {
    loading.value = false
  }
}

const onFinishFailed = (errorInfo) => {
  console.log('Failed:', errorInfo)
}



</script>

<style scoped>
@import '../assets/css2.css';

/* ========== 全局容器与背景 ========== */
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  position: relative;
  overflow: hidden;
  font-family: 'Fredoka', 'Microsoft YaHei', sans-serif;
}

/* 渐变底层 */
.bg-layer {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: linear-gradient(120deg, #fccb90 0%, #d57eeb 100%);
  opacity: 0.6;
  z-index: -2;
}

/* 波点网格 */
.grid-pattern {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background-image: 
    radial-gradient(#ffffff 2px, transparent 2px),
    linear-gradient(rgba(255, 255, 255, 0.1) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.1) 1px, transparent 1px);
  background-size: 30px 30px, 50px 50px, 50px 50px;
  background-position: 0 0, 0 0, 0 0;
  z-index: -1;
}

/* ========== 漂浮装饰 ========== */
.bg-decoration {
  position: absolute;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 0;
}

.floating-shape {
  position: absolute;
  border-radius: 50%;
  filter: blur(40px);
  animation: float 10s infinite ease-in-out;
}

.shape-1 {
  width: 300px; height: 300px;
  background: #ff9a9e;
  top: -50px; left: -50px;
  opacity: 0.5;
}

.shape-2 {
  width: 400px; height: 400px;
  background: #a18cd1;
  bottom: -100px; right: -100px;
  opacity: 0.4;
  animation-delay: -5s;
}

.star {
  position: absolute;
  font-size: 24px;
  animation: twinkle 3s infinite alternate;
}
.star-1 { top: 15%; left: 10%; animation-delay: 0s; }
.star-2 { top: 25%; right: 20%; animation-delay: 1s; font-size: 18px; }
.star-3 { bottom: 20%; left: 15%; animation-delay: 2s; }
.star-4 { bottom: 10%; right: 10%; animation-delay: 1.5s; font-size: 30px;}

@keyframes float {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(20px, 30px); }
}
@keyframes twinkle {
  0% { transform: scale(1) rotate(0deg); opacity: 0.6; }
  100% { transform: scale(1.2) rotate(15deg); opacity: 1; }
}

/* ========== 卡片核心 ========== */
.login-card-wrapper {
  position: relative;
  z-index: 10;
  width: 90%;
  max-width: 420px;
  background: rgba(255, 255, 255, 0.75);
  backdrop-filter: blur(15px);
  -webkit-backdrop-filter: blur(15px);
  border-radius: 24px;
  padding: 40px 30px;
  box-shadow: 
    0 10px 40px rgba(255, 154, 158, 0.3),
    0 0 0 5px rgba(255, 255, 255, 0.4);
  border: 2px solid #fff;
  animation: cardEnter 0.6s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

@keyframes cardEnter {
  from { opacity: 0; transform: translateY(50px) scale(0.9); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

.card-ribbon {
  position: absolute;
  top: -25px;
  left: 50%;
  transform: translateX(-50%);
  font-size: 40px;
  filter: drop-shadow(0 5px 5px rgba(0,0,0,0.1));
  z-index: 20;
}

/* ========== 头部修改区域 (重点) ========== */
.card-header {
  display: flex;             /* 启用Flex布局 */
  flex-direction: column;    /* 垂直排列：上标题，下副标题 */
  align-items: center;       /* 水平居中 */
  justify-content: center;
  margin-bottom: 35px;
  width: 100%;
}

.system-title {
  display: block;            /* 块级元素 */
  width: 100%;               /* 占满整行宽度 */
  text-align: center;        /* 文字居中 */
  margin: 0 0 15px 0;        /* 底部留出间距，与副标题分开 */
  font-size: 26px;
  font-weight: 800;
  color: #5c5c8a;
  letter-spacing: 2px;
  text-shadow: 2px 2px 0px #fff;
  line-height: 1.4;          /* 优化行高 */
}

.subtitle-badge {
  display: inline-block;
  background: #ffebf7;
  padding: 5px 15px;
  border-radius: 20px;
  border: 1px dashed #ffb7d6;
}

.subtitle-badge span {
  color: #ff85b3;
  font-size: 13px;
  font-weight: bold;
}

/* ========== 表单控件 ========== */
.input-group {
  display: flex;
  align-items: center;
  background: #fff;
  border: 2px solid #f0f0f0;
  border-radius: 999px;
  padding: 5px 15px;
  transition: all 0.3s ease;
}

.input-group:focus-within {
  border-color: #ffb7d6;
  box-shadow: 0 0 15px rgba(255, 183, 214, 0.4);
  transform: translateY(-2px);
}

.input-icon {
  font-size: 18px;
  margin-right: 8px;
  filter: grayscale(0.5);
  transition: 0.3s;
}

.input-group:focus-within .input-icon {
  filter: grayscale(0);
  transform: scale(1.1);
}

.kawaii-input {
  flex: 1;
  background: transparent !important;
  height: 38px;
  font-size: 14px;
  color: #666;
}

:deep(.ant-input-password), 
:deep(.ant-input), 
:deep(.ant-input:focus), 
:deep(.ant-input-focused) {
  box-shadow: none !important;
  border: none !important;
}

:deep(.ant-input-password-icon) {
  color: #ffb7d6 !important;
}
:deep(.ant-input-password-icon:hover) {
  color: #ff85b3 !important;
}

/* ========== 按钮 ========== */
.kawaii-button {
  height: 48px;
  border-radius: 999px;
  background: linear-gradient(90deg, #ff9a9e 0%, #fad0c4 99%, #fad0c4 100%);
  border: none;
  font-size: 16px;
  font-weight: bold;
  color: #fff;
  text-shadow: 1px 1px 2px rgba(0,0,0,0.1);
  box-shadow: 0 6px 20px rgba(255, 154, 158, 0.4);
  transition: all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  margin-top: 10px;
}

.kawaii-button:hover, 
.kawaii-button:focus {
  background: linear-gradient(90deg, #ff85b3 0%, #ff9a9e 100%);
  transform: translateY(-3px) scale(1.02);
  box-shadow: 0 10px 25px rgba(255, 133, 179, 0.5);
}

.kawaii-button:active {
  transform: translateY(1px) scale(0.98);
}

/* ========== 错误提示 ========== */
.error-bubble {
  background: #fff1f0;
  border: 1px solid #ffccc7;
  color: #ff4d4f;
  padding: 10px;
  border-radius: 12px;
  font-size: 13px;
  text-align: center;
  margin-top: 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.bounce-enter-active {
  animation: bounce-in 0.5s;
}
.bounce-leave-active {
  animation: bounce-in 0.5s reverse;
}
@keyframes bounce-in {
  0% { transform: scale(0); }
  50% { transform: scale(1.1); }
  100% { transform: scale(1); }
}

/* ========== 底部 ========== */
.card-footer {
  margin-top: 25px;
  text-align: center;
}

.footer-divider {
  height: 2px;
  background: repeating-linear-gradient(
    90deg,
    #ffb7d6 0,
    #ffb7d6 6px,
    transparent 6px,
    transparent 12px
  );
  margin-bottom: 15px;
  opacity: 0.5;
}

.footer-text {
  color: #999;
  font-size: 12px;
  margin-bottom: 8px;
}

.contact-pill {
  display: inline-block;
  background: #f0f5ff;
  color: #85a5ff;
  padding: 4px 12px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: bold;
}

/* ========== 响应式 ========== */
@media (max-width: 480px) {
  .login-card-wrapper {
    width: 85%;
    padding: 30px 20px;
  }
  
  .system-title {
    font-size: 22px;
  }
  
  .kawaii-button {
    height: 44px;
    font-size: 15px;
  }
  
  .star-4 { display: none; }
}
</style>