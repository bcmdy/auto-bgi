import axios from 'axios'

// 创建axios实例
const api = axios.create({
  baseURL: '', // 使用相对路径，让Vite代理处理
  timeout: 35000,
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
    // if (process.env.NODE_ENV !== 'production') {
    //   console.error('API请求错误:', error)
    //   alert(error.response.data.msg)
    // }
    return error.response.data
  }
)

// API方法定义
export const apiMethods = {
  // 获取系统状态
  getStatus: () => api.get('/api/index'),
  indexSX: () => api.get('/api/indexSX'),
  updateABgi: () => api.post('/api/updateABgi'),
  
  // 获取轮播图图片列表
  getImages: () => api.get('/api/images'),

  // 发送桌面截图
  sendImage: () => api.post('/api/sendImage'),

  // 米游社手动签到
  mysSignIn: () => api.post('/api/mysSignIn'),
  
  // 系统操作
  startOneLong: (data) => api.post('/api/oneLong/startOneLong', data),
  closeBgi: () => api.post('/api/closeBgi'),
  closeYuanShen: () => api.post('/closeYuanShen'),
  backup: () => api.post('/api/backup'),
  autoUpdateJsAndPathing: () => api.post('/autoUpdateJsAndPathing'),
  
  // 配置相关
  getConfig: () => api.get('/api/config'),
  updateConfig: (data) => api.post('/api/saveConfig', data),
  getOneLongAllName: () => api.get('/api/oneLong/oneLongAllName'),
  
  // 日志相关
  getLog: () => api.get('/api/gitLog'),
  getLogFiles: () => api.get('/api/logFiles'),
  getLogAnalysis: (file) => api.get('/api/logAnalysis', { params: { file } }),
  queryAutoLogs: (keyword = '') => {
    const params = {}
    if (keyword && keyword.trim()) {
      params.data = keyword.trim()
    }
    return api.get('/api/autoLog', { params })
  },
  
  // 归档
  getArchive: (params) => api.get('/archive', { params }),
  getArchiveList: () => api.get('/api/archiveList'),
  deleteArchive: (id) => api.delete(`/api/archive?id=${id}`),
  deleteAllArchive: () => api.delete(`/api/allArchives`),
  
  // 其他功能
  getOther: () => api.get('/other'),
  getJsNames: () => api.get('/api/jsNames'),
  getListGroups: () => api.get('/api/scriptGroup/listGroups'),
  // 读取配置组所有的地图追踪
  listPathingUpdatePaths: () => api.get('/api/scriptGroup/listPathingUpdatePaths'),
  getAutoArtifactsPro: () => api.get('/api/getAutoArtifactsPro'),
  getAutoArtifactsPro2: () => api.get('/api/getAutoArtifactsPro2'),
  getHarvest: () => api.get('/harvest'),
  getBg: () => api.get('/bg'),
  getOneLong: () => api.get('/onelong'),
  getError: () => api.get('/error'),
  getCalculateTaskEnabledList: () => api.get('/CalculateTaskEnabledList'),
  getBagStatistics: () => api.get('/api/BagStatistics'),
  getCDAwareAutoGather: (status = '3') => api.get('/api/CD-Aware-AutoGather/ReadInfo', { params: { status } }),
  // 更新是否加入背包统计
  CDAllMaterial : () => api.post('/api/CD-Aware-AutoGather/CDAllMaterial'),
  // 一键更新全部材料
  UpdateAllCD : () => api.post('/api/CD-Aware-AutoGather/UpdateAllCD'),
  
  // 启动配置组
  startGroups: (names) => {
    const payload = Array.isArray(names) ? names : [names]
    return api.post('/api/startGroups', payload)
  },

  // 狗粮联机
  StartOnline: (typeKey,runDebug) => api.post('/api/abgiSSE/connect/'+typeKey+"?runDebug="+runDebug),
  offline:() => api.post('/api/abgiSSE/disconnect'),
  DogFooddisconnect: () => api.post('/api/abgiSSE/disconnect'),
  
  // 黑名单相关
  getBlackList: () => api.get('/api/betterGi/blackList'),
  addBlackList: (blackList) => api.post('/api/betterGi/addBlackList', blackList),
  deleteBlackList: (blackName) => api.post('/api/betterGi/deleteBlackList', blackName),

 //获取录制状态
  IsRecording: () => api.get('/api/abgiObs/IsRecording'),
  // 开始录制
  StartRecording: () => api.post('/api/abgiObs/StartRecording'),
  // 停止录制
  StopRecording: () => api.post('/api/abgiObs/StopRecording'),
  // 获取回放缓冲区状态
  GetReplayBufferStatus: () => api.get('/api/abgiObs/GetReplayBufferStatus'),
  // 启动回放缓冲区
  StartReplayBuffer: () => api.post('/api/abgiObs/StartReplayBuffer'),
  // 停止回放缓冲区
  StopReplayBuffer: () => api.post('/api/abgiObs/StopReplayBuffer'),
  // 保存回放缓冲区
  SaveReplayBuffer: () => api.post('/api/abgiObs/SaveReplayBuffer'),
  // 获取视频信息
  GetVideoInfo: () => api.get('/api/abgiObs/GetVideoInfo'),
  //删除视频
  DeleteVideo: (filePath) => api.post('/api/abgiObs/DeleteVideo?videoName='+filePath),
  
}

export default api
