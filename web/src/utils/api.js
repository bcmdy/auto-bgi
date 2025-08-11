import axios from 'axios'

// 创建axios实例
const api = axios.create({
  baseURL: '', // 使用相对路径，让Vite代理处理
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  config => {
    if (process.env.NODE_ENV !== 'production') {
      console.log('API请求:', config.method?.toUpperCase(), config.url)
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    if (process.env.NODE_ENV !== 'production') {
      console.log('API响应:', response.status, response.config.url)
    }
    return response.data
  },
  error => {
    if (process.env.NODE_ENV !== 'production') {
      console.error('API请求错误:', error)
    }
    return Promise.reject(error)
  }
)

// API方法定义
export const apiMethods = {
  // 获取系统状态
  getStatus: () => api.get('/api/index'),
  
  // 获取轮播图图片列表
  getImages: () => api.get('/api/images'),
  
  // 系统操作
  startOneLong: () => api.post('/oneLong'),
  closeBgi: () => api.post('/closeBgi'),
  closeYuanShen: () => api.post('/closeYuanShen'),
  backup: () => api.post('/backup'),
  autoUpdateJsAndPathing: () => api.post('/autoUpdateJsAndPathing'),
  
  // 配置相关
  getConfig: () => api.get('/api/config'),
  updateConfig: (data) => api.post('/api/saveConfig', data),
  getOneLongAllName: () => api.get('/api/oneLongAllName'),
  
  // 日志相关
  getLog: () => api.get('/log'),
  getLogFiles: () => api.get('/api/logFiles'),
  getLogAnalysis: (file) => api.get('/api/logAnalysis', { params: { file } }),
  
  // 归档查询
  getArchive: (params) => api.get('/archive', { params }),
  
  // 归档列表相关
  getArchiveList: () => api.get('/api/archiveList'),
  deleteArchive: (id) => api.delete(`/api/archive?id=${id}`),
  
  // 其他功能
  getOther: () => api.get('/other'),
  getJsNames: () => api.get('/api/jsNames'),
  getListGroups: () => api.get('/api/listGroups'),
  getAutoArtifactsPro: () => api.get('/api/getAutoArtifactsPro'),
  getAutoArtifactsPro2: () => api.get('/api/getAutoArtifactsPro2'),
  getHarvest: () => api.get('/harvest'),
  getBg: () => api.get('/bg'),
  getOneLong: () => api.get('/onelong'),
  getError: () => api.get('/error'),
  getCalculateTaskEnabledList: () => api.get('/CalculateTaskEnabledList'),
  getBagStatistics: () => api.get('/api/BagStatistics'),
  
  // 启动配置组
  startGroups: (names) => {
    const payload = Array.isArray(names) ? names : [names]
    return api.post('/api/startGroups', payload)
  }
}

export default api