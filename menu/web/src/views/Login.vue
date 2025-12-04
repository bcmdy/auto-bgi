<template>
  <div class="login-container">
    <!-- 背景装饰 -->
    <div class="bg-decoration">
      <div class="star star-1"></div>
      <div class="star star-2"></div>
      <div class="star star-3"></div>
      <div class="star star-4"></div>
      <div class="star star-5"></div>
      <div class="circle circle-1"></div>
      <div class="circle circle-2"></div>
      <div class="circle circle-3"></div>
    </div>

    <!-- 登录卡片 -->
    <div class="login-card-wrapper">
      <div class="card-header">
        <h1 class="system-title">~~~{{ systemName }}~~~</h1>
        <p class="subtitle">欢迎回来~</p>
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
          <a-input
            v-model:value="formState.username"
            placeholder="👤 输入你的用户名"
            class="form-input"
            @keyup.enter="handleEnter"
          />
        </a-form-item>

        <a-form-item
          name="password"
          :rules="[{ required: true, message: '请输入密码哦~' }]"
        >
          <a-input-password
            v-model:value="formState.password"
            placeholder="🔐 输入你的密码"
            class="form-input"
            @keyup.enter="handleEnter"
          />
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            :loading="loading"
            block
            class="login-button"
          >
            {{ loading ? '登录中✨' : '进入系统 →' }}
          </a-button>
        </a-form-item>

        <transition name="fade">
          <div v-if="errorMessage" class="error-message">
            ⚠️ {{ errorMessage }}
          </div>
        </transition>
      </a-form>

      <div class="card-footer">
        <p class="footer-text">嘿~这是一个神秘的地方呢🎀</p>
        <p class="footer-text">qq群：215053644</p>
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

// 页面挂载时获取系统配置
onMounted(async () => {
  try {
    const response = await apiMethods.getSystemConfig()
    if (response.systemName) {
      systemName.value = response.systemName
    }
  } catch (error) {
    console.error('获取系统配置失败:', error)
    // 保持默认标题 "登录"
  }
})

const handleEnter = () => {
  if (formRef.value) {
    formRef.value.submit()
  }
}

const onFinish = async () => {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await apiMethods.login(
      formState.value.username,
      formState.value.password
    )

    // 检查响应中的 code 字段
    if (response.code === 401 || response.error) {
      // 登录失败
      errorMessage.value = response.error || '登录失败，请检查用户名和密码'
      message.error('登录失败：' + (response.error || '未知错误'))
    } else if (response.code === 200 && response.aBgiToken) {
      // 登录成功
      localStorage.setItem('aBgiToken', response.aBgiToken)
      message.success('登录成功！')
      // 重定向到首页
      router.push('/')
    } else if (response.aBgiToken) {
      // 兼容没有 code 字段的情况
      localStorage.setItem('aBgiToken', response.aBgiToken)
      message.success('登录成功！')
      router.push('/')
    } else {
      // 未知响应格式
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
/* ========== 容器样式 ========== */
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  position: relative;
  overflow: hidden;
  /* 渐变背景：柔和粉色到紫色 */
  background: linear-gradient(135deg, #ffc0e9 0%, #f5a9e6 25%, #e099d4 50%, #d598d9 75%, #b5d0e6 100%);
  font-family: 'Comic Sans MS', 'Segoe UI', sans-serif;
}

/* ========== 背景装饰 ========== */
.bg-decoration {
  position: absolute;
  width: 100%;
  height: 100%;
  overflow: hidden;
  z-index: 1;
}

/* 星星装饰 */
.star {
  position: absolute;
  width: 4px;
  height: 4px;
  background: white;
  border-radius: 50%;
  opacity: 0.8;
  animation: twinkle 3s ease-in-out infinite;
}

.star-1 { top: 10%; left: 10%; animation-delay: 0s; }
.star-2 { top: 20%; right: 15%; animation-delay: 0.5s; }
.star-3 { top: 40%; left: 5%; animation-delay: 1s; }
.star-4 { bottom: 20%; right: 10%; animation-delay: 1.5s; }
.star-5 { bottom: 30%; left: 15%; animation-delay: 2s; }

@keyframes twinkle {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 1; }
}

/* 圆形装饰 */
.circle {
  position: absolute;
  border-radius: 50%;
  opacity: 0.1;
}

.circle-1 {
  width: 400px;
  height: 400px;
  background: #ff69b4;
  top: -100px;
  left: -150px;
  animation: float 6s ease-in-out infinite;
}

.circle-2 {
  width: 300px;
  height: 300px;
  background: #ff99cc;
  bottom: -80px;
  right: -100px;
  animation: float 8s ease-in-out infinite reverse;
}

.circle-3 {
  width: 200px;
  height: 200px;
  background: #ffb3d9;
  top: 50%;
  right: 5%;
  animation: float 7s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0px); }
  50% { transform: translateY(20px); }
}

/* ========== 卡片样式 ========== */
.login-card-wrapper {
  position: relative;
  z-index: 10;
  width: 100%;
  max-width: 480px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 30px;
  padding: 40px 30px;
  box-shadow: 0 20px 60px rgba(255, 105, 180, 0.25), 
              0 0 30px rgba(200, 150, 200, 0.2);
  border: 2px solid rgba(255, 192, 233, 0.4);
  animation: slideUp 0.6s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* ========== 卡片头 ========== */
.card-header {
  text-align: center;
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 2px dashed rgba(255, 105, 180, 0.3);
}

.system-title {
  margin: 0;
  font-size: 28px;
  font-weight: bold;
  background: linear-gradient(135deg, #ff1493 0%, #ff69b4 50%, #da70d6 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  animation: shimmer 2s ease-in-out infinite;
}

.subtitle {
  margin: 8px 0 0 0;
  font-size: 14px;
  color: #da70d6;
  font-style: italic;
  letter-spacing: 1px;
}

@keyframes shimmer {
  0%, 100% { filter: brightness(1); }
  50% { filter: brightness(1.1); }
}

/* ========== 表单样式 ========== */
.login-form {
  margin: 0;
}

.form-input {
  height: 45px;
  font-size: 14px;
  border-radius: 15px;
  border: 2px solid #ffc0e9;
  background: rgba(255, 240, 250, 0.8);
  transition: all 0.3s ease;
}

.form-input:focus,
.form-input:hover,
:deep(.form-input:focus),
:deep(.form-input:hover) {
  border-color: #ff69b4;
  box-shadow: 0 0 15px rgba(255, 105, 180, 0.25);
  background: rgba(255, 255, 255, 0.95);
}

:deep(.form-input input),
:deep(.form-input input::placeholder) {
  color: #999;
}

:deep(.form-input input) {
  font-size: 14px;
}

/* ========== 按钮样式 ========== */
.login-button {
  height: 45px;
  font-size: 16px;
  font-weight: bold;
  border-radius: 15px;
  border: none;
  background: linear-gradient(135deg, #ff69b4 0%, #da70d6 100%);
  color: white;
  transition: all 0.3s ease;
  letter-spacing: 1px;
  text-transform: uppercase;
  box-shadow: 0 5px 20px rgba(255, 105, 180, 0.35);
  margin-top: 10px;
}

.login-button:hover,
:deep(.login-button:hover) {
  transform: translateY(-2px);
  box-shadow: 0 8px 30px rgba(255, 105, 180, 0.45);
  background: linear-gradient(135deg, #ff5fa8 0%, #d860d1 100%);
}

.login-button:active,
:deep(.login-button:active) {
  transform: translateY(0px);
}

/* ========== 错误提示 ========== */
.error-message {
  margin-top: 15px;
  padding: 12px;
  background: rgba(255, 77, 79, 0.1);
  border: 1.5px solid #ff4d4f;
  border-radius: 10px;
  color: #ff4d4f;
  font-size: 13px;
  text-align: center;
  font-weight: 500;
  animation: shake 0.5s ease-in-out;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-5px); }
  75% { transform: translateX(5px); }
}

/* ========== 卡片底部 ========== */
.card-footer {
  text-align: center;
  margin-top: 25px;
  padding-top: 20px;
  border-top: 2px dashed rgba(255, 105, 180, 0.3);
}

.footer-text {
  margin: 0;
  font-size: 12px;
  color: #da70d6;
  opacity: 0.7;
  font-style: italic;
}

/* ========== 过渡效果 ========== */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* ========== Ant Design 组件覆盖 ========== */
:deep(.ant-form-item) {
  margin-bottom: 20px;
}

:deep(.ant-form-item-has-error .ant-input),
:deep(.ant-form-item-has-error .ant-input-affix-wrapper) {
  border-color: #ff4d4f !important;
  background: rgba(255, 77, 79, 0.05) !important;
}

:deep(.ant-input-password-icon) {
  color: #da70d6;
}

:deep(.ant-input::placeholder) {
  color: #d3a5d3;
}

/* 响应式设计 */
@media (max-width: 480px) {
  .login-card-wrapper {
    max-width: calc(100% - 30px);
    padding: 30px 20px;
  }

  .system-title {
    font-size: 24px;
  }

  .login-form {
    margin-top: 20px;
  }
}
</style>
