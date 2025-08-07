<template>
  <div class="bg-container">
    <a-card title="背景管理" class="bg-card">
      <a-tabs v-model:activeKey="activeTab" type="card">
        <a-tab-pane key="wallpaper" tab="壁纸设置">
          <a-row :gutter="[16, 16]">
            <a-col :span="6" v-for="wallpaper in wallpapers" :key="wallpaper.id">
              <a-card 
                class="wallpaper-card"
                :class="{ active: currentWallpaper === wallpaper.id }"
                @click="setWallpaper(wallpaper)"
              >
                <img :src="wallpaper.thumbnail" :alt="wallpaper.name" class="wallpaper-preview" />
                <div class="wallpaper-info">
                  <div class="wallpaper-name">{{ wallpaper.name }}</div>
                  <div class="wallpaper-size">{{ wallpaper.size }}</div>
                </div>
              </a-card>
            </a-col>
          </a-row>
        </a-tab-pane>

        <a-tab-pane key="theme" tab="主题设置">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-card title="颜色主题" size="small">
                <a-space direction="vertical" style="width: 100%">
                  <div v-for="theme in themes" :key="theme.name" class="theme-option">
                    <a-radio 
                      :value="theme.name" 
                      v-model:checked="currentTheme"
                      @change="setTheme(theme)"
                    >
                      <span class="theme-name">{{ theme.displayName }}</span>
                      <div class="theme-preview">
                        <span 
                          v-for="color in theme.colors" 
                          :key="color"
                          class="color-dot"
                          :style="{ backgroundColor: color }"
                        ></span>
                      </div>
                    </a-radio>
                  </div>
                </a-space>
              </a-card>
            </a-col>

            <a-col :span="12">
              <a-card title="动画设置" size="small">
                <a-form layout="vertical">
                  <a-form-item label="樱花动画">
                    <a-switch v-model:checked="animationSettings.sakura" @change="updateAnimations" />
                  </a-form-item>
                  <a-form-item label="背景粒子">
                    <a-switch v-model:checked="animationSettings.particles" @change="updateAnimations" />
                  </a-form-item>
                  <a-form-item label="动画速度">
                    <a-slider 
                      v-model:value="animationSettings.speed" 
                      :min="0.5" 
                      :max="2" 
                      :step="0.1"
                      @change="updateAnimations"
                    />
                  </a-form-item>
                </a-form>
              </a-card>
            </a-col>
          </a-row>
        </a-tab-pane>

        <a-tab-pane key="upload" tab="上传背景">
          <a-upload-dragger
            v-model:fileList="fileList"
            :before-upload="beforeUpload"
            @change="handleUpload"
            accept="image/*"
            :multiple="false"
          >
            <p class="ant-upload-drag-icon">
              <InboxOutlined />
            </p>
            <p class="ant-upload-text">点击或拖拽文件到此区域上传</p>
            <p class="ant-upload-hint">
              支持 JPG、PNG、GIF 格式，建议尺寸 1920x1080
            </p>
          </a-upload-dragger>
        </a-tab-pane>
      </a-tabs>
    </a-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { InboxOutlined } from '@ant-design/icons-vue'
import { apiMethods } from '@/utils/api'

const activeTab = ref('wallpaper')
const currentWallpaper = ref('')
const currentTheme = ref('default')
const fileList = ref([])

const wallpapers = ref([
  {
    id: 'bg1',
    name: '二次元风格1',
    thumbnail: '/static/image/bd.jpg',
    size: '1920x1080'
  },
  {
    id: 'bg2',
    name: '二次元风格2',
    thumbnail: '/static/image/ff.png',
    size: '1920x1080'
  },
  {
    id: 'bg3',
    name: '二次元风格3',
    thumbnail: '/static/image/ng.jpg',
    size: '1920x1080'
  },
  {
    id: 'bg4',
    name: '二次元风格4',
    thumbnail: '/static/image/sh.jpg',
    size: '1920x1080'
  }
])

const themes = ref([
  {
    name: 'default',
    displayName: '默认粉色',
    colors: ['#ffecf5', '#ff6699', '#ffd1e0', '#ff99cc']
  },
  {
    name: 'blue',
    displayName: '蓝色主题',
    colors: ['#e6f7ff', '#1890ff', '#91d5ff', '#40a9ff']
  },
  {
    name: 'purple',
    displayName: '紫色主题',
    colors: ['#f9f0ff', '#722ed1', '#d3adf7', '#9254de']
  },
  {
    name: 'green',
    displayName: '绿色主题',
    colors: ['#f6ffed', '#52c41a', '#b7eb8f', '#73d13d']
  }
])

const animationSettings = reactive({
  sakura: true,
  particles: false,
  speed: 1.0
})

const setWallpaper = (wallpaper) => {
  currentWallpaper.value = wallpaper.id
  message.success(`已设置壁纸: ${wallpaper.name}`)
  
  // 这里可以调用API保存设置
  // apiMethods.setWallpaper({ wallpaperId: wallpaper.id })
}

const setTheme = (theme) => {
  currentTheme.value = theme.name
  message.success(`已切换到: ${theme.displayName}`)
  
  // 这里可以动态修改CSS变量来改变主题
  document.documentElement.style.setProperty('--primary-color', theme.colors[1])
}

const updateAnimations = () => {
  message.success('动画设置已更新')
  // 这里可以调用API保存动画设置
}

const beforeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  if (!isImage) {
    message.error('只能上传图片文件!')
    return false
  }
  
  const isLt10M = file.size / 1024 / 1024 < 10
  if (!isLt10M) {
    message.error('图片大小不能超过10MB!')
    return false
  }
  
  return true
}

const handleUpload = (info) => {
  if (info.file.status === 'done') {
    message.success('上传成功!')
    // 刷新壁纸列表
  } else if (info.file.status === 'error') {
    message.error('上传失败!')
  }
}

onMounted(() => {
  // 加载当前设置
  currentWallpaper.value = 'bg1'
  currentTheme.value = 'default'
})
</script>

<style scoped>
.bg-container {
  padding: 20px;
  background: #ffeaf4;
  min-height: 100vh;
}

.bg-card {
  border-radius: 16px !important;
  box-shadow: 0 4px 12px rgba(255, 105, 180, 0.1) !important;
}

.wallpaper-card {
  cursor: pointer;
  border-radius: 12px !important;
  transition: all 0.3s ease;
  border: 2px solid transparent !important;
}

.wallpaper-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(255, 102, 170, 0.2) !important;
}

.wallpaper-card.active {
  border-color: #ff99cc !important;
  box-shadow: 0 0 15px rgba(255, 102, 170, 0.3) !important;
}

.wallpaper-preview {
  width: 100%;
  height: 120px;
  object-fit: cover;
  border-radius: 8px;
}

.wallpaper-info {
  margin-top: 8px;
  text-align: center;
}

.wallpaper-name {
  font-weight: bold;
  color: #ff5599;
}

.wallpaper-size {
  font-size: 12px;
  color: #999;
}

.theme-option {
  padding: 12px;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  margin-bottom: 8px;
  transition: all 0.3s ease;
}

.theme-option:hover {
  border-color: #ff99cc;
  background: #fff0f7;
}

.theme-name {
  margin-left: 8px;
  font-weight: bold;
  color: #ff5599;
}

.theme-preview {
  display: flex;
  gap: 4px;
  margin-top: 4px;
  margin-left: 24px;
}

.color-dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 1px solid #ddd;
}

:deep(.ant-card-head-title) {
  color: #ff5599;
  font-size: 18px;
  font-weight: bold;
}

:deep(.ant-upload-drag) {
  border: 2px dashed #ff99cc;
  border-radius: 12px;
}

:deep(.ant-upload-drag:hover) {
  border-color: #ff6699;
}

:deep(.ant-upload-drag-icon) {
  color: #ff99cc;
}
</style>