<template>
  <div class="error-container">
    <a-card class="error-card">
      <a-result
        status="error"
        title="出现了一些问题"
        sub-title="抱歉，页面遇到了错误。请稍后再试或联系管理员。"
      >
        <template #extra>
          <a-space>
            <a-button type="primary" @click="goHome">
              返回首页
            </a-button>
            <a-button @click="refreshPage">
              刷新页面
            </a-button>
          </a-space>
        </template>
      </a-result>

      <a-divider />

      <a-collapse v-if="errorDetails">
        <a-collapse-panel key="1" header="错误详情">
          <pre class="error-details">{{ errorDetails }}</pre>
        </a-collapse-panel>
      </a-collapse>
    </a-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()
const errorDetails = ref('')

const goHome = () => {
  router.push('/')
}

const refreshPage = () => {
  window.location.reload()
}

onMounted(() => {
  // 从路由参数获取错误详情
  if (route.query.error) {
    errorDetails.value = decodeURIComponent(route.query.error)
  }
})
</script>

<style scoped>
.error-container {
  padding: 20px;
  background: #ffeaf4;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
}

.error-card {
  max-width: 600px;
  width: 100%;
  border-radius: 16px !important;
  box-shadow: 0 4px 12px rgba(255, 105, 180, 0.1) !important;
}

.error-details {
  background: #f8f9fa;
  padding: 12px;
  border-radius: 8px;
  font-size: 12px;
  white-space: pre-wrap;
  word-wrap: break-word;
}

:deep(.ant-result-title) {
  color: #ff5599;
}

:deep(.ant-btn-primary) {
  background: #ff99cc;
  border-color: #ff99cc;
}

:deep(.ant-btn-primary:hover) {
  background: #ff6699 !important;
  border-color: #ff6699 !important;
}
</style>