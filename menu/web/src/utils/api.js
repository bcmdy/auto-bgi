import axios from 'axios'
import router from '../router'

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
    
    // 从 localStorage 获取 token 并添加到 Authorization 头
    const token = localStorage.getItem('aBgiToken')
    if (token) {
      config.headers.Authorization = `${token}`
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
    // 1. 处理 401 未授权错误
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('aBgiToken')
      console.warn('认证已过期，请重新登录')
      router.push('/login')
      return Promise.reject(error)
    }

    //处理888错误，强制更新前端
    if (error.response && error.response.status === 888) {
      console.warn('等待重启中，请稍后...')
      router.push('/')
      return Promise.reject("等待重启中，请稍后...")
    }
  


    // 2. 【新增】处理网络错误、后端未启动、连接超时等情况
    // 如果 error.response 不存在，说明根本没有收到后端的响应
    if ( error.code === 'ERR_BAD_RESPONSE') {
            console.error('连接失败：后端未启动或网络异常', error.message)
            var token = localStorage.getItem('aBgiToken')
            if (token==null) {
              localStorage.removeItem('aBgiToken')
                // 只有当前不在登录页时才跳转，防止重复跳转报错
            if (router.currentRoute._rawValue.path !== '/login') {
              router.push('/login')
              return
            }
      }
      
      
    
      return Promise.reject(error)
    }
    
    // 3. 其他业务错误情况
    return error.response?.data || Promise.reject(error)
  }
)

// API方法定义
export const apiMethods = {
  // 认证相关
  login: (username, password) => api.post('/api/auth/login', { username, password }),
  getSystemConfig: () => api.get('/api/auth/getSystemConfig'),
  
  // 获取系统状态
  getStatus: () => api.get('/api/index'),
  indexSX: () => api.get('/api/indexSX'),
  updateABgi: () => api.post('/api/updateABgi'),
  
  // 获取轮播图图片列表
  getImages: () => api.get('/api/images'),

  // 发送桌面截图
  sendImage: () => api.post('/api/sendImage'),

  //查看桌面图片
  viewImage: () => api.get('/api/aBgiJt?t=' + new Date().getTime()),

  // 米游社手动签到
  mysSignIn: () => api.post('/api/mysSignIn'),
  // 设置米游社签到推送（将 NoticeType 发送到后端）
  mysPush: (NoticeType) => api.post('/api/mysPush?NoticeType=' + encodeURIComponent(NoticeType)),
  
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
  // BGI 配置相关
  getBgiConfigAll: () => api.get('/api/bgiConfig/allConfig'),
  findBgiConfig: (configName) => api.get('/api/bgiConfig/findConfig', { params: { configName } }),
  saveBgiConfig: (data) => api.post('/api/bgiConfig/saveConfig', data),
  
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
  addBagStatistics: (name) => api.post(`/api/BagStatistics/ADD?name=${encodeURIComponent(name)}`),
  getCDAwareAutoGather: (status = '3') => api.get('/api/CD-Aware-AutoGather/ReadInfo', { params: { status } }),
  // 采集管理
  getCollectionManagement: (name) => api.get('/api/CDCollectionManagement/list', { params: { name } }),
  getAllUserFiles: () => api.get('/api/CDCollectionManagement/AllUserFile'),
  getPickupHistory: (name) => api.get('/api/CDCollectionManagement/ReadPickup', { params: { name } }),
  // 更新是否加入背包统计
  CDAllMaterial : () => api.post('/api/CD-Aware-AutoGather/CDAllMaterial'),
  // 一键更新全部材料
  UpdateAllCD : () => api.post('/api/CD-Aware-AutoGather/UpdateAllCD'),
  
  // 启动配置组
  startGroups: (names) => {
    const payload = Array.isArray(names) ? names : [names]
    return api.post('/api/startGroups', payload)
  },

  //脚本组管理
  //批量更新脚本
  batchUpdate: () => api.post('/api/batchUpdate'),
  //更新指定脚本
  updateJs: (name) => api.post(`/api/updateJs/${name}`),
  //获取脚本列表
  getJsNames: () => api.get('/api/jsNames'),
  //重置仓库
  resetRepo: () => api.post('/api/resetRepo'),
  //查询所有脚本
  getAllScripts: (search) => api.get('/api/jsNamesAll?search=' + search),

  // 在你的 api.js 或 api定义文件中添加：
subscribeScript: (scriptName) => {
    return api.post('/api/repo/subscribe?ScriptName=' + scriptName, null, {
      headers: {
        'Content-Type': 'multipart/form-data'
      },
      timeout: 10 * 60 * 1000 // 10 分钟
    })
},

  // 狗粮联机
  StartOnline: (typeKey,runDebug) => api.post('/api/abgiSSE/connect/'+typeKey+"?runDebug="+runDebug),
  offline:() => api.post('/api/abgiSSE/disconnect'),
  DogFooddisconnect: () => api.post('/api/abgiSSE/disconnect'),
  reportBomb: (payload) => api.post('/api/abgiSSE/reportBomb', payload),
  // 查询上线次数
  getNumberOfLaunches: () => api.get('/api/NumberOfLaunches'),
  // 清零上线次数
  clearNumberOfLaunches: () => api.post('/api/abgiSSE/clearNumberOfLaunches'),
  
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
  // 删除所有视频
  DeleteAllVideo: () => api.post('/api/abgiObs/DeleteAllVideo'),

  // 定时任务管理
  getTaskCronList: () => api.get('/api/taskCron/list'),
  getAvailableTaskCronNames: () => api.get('/api/taskCron/getTasks'),
  addTaskCron: (payload) => api.post('/api/taskCron/add', payload),
  updateTaskCron: (payload) => api.post('/api/taskCron/update', payload),
  pauseTaskCron: (id) => api.post(`/api/taskCron/pause?id=${id}`),
  
  resumeTaskCron: (id) => api.post(`/api/taskCron/resume?id=${id}`),
  // 删除定时任务
  removeTaskCron: (id,EntryID) => api.post(`/api/taskCron/remove?id=${id}&EntryID=${EntryID}`),
  // 立即执行定时任务
  AtOnceRunTaskCron: (type,data) => api.post(`/api/taskCron/AtOnceRun?type=${type}&data=${data}`),

  GetAppInfo: () => api.get('/api/appInfo'),

  // BGI更新
  uploadBgi: (file) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/api/uploadBgi', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      },
      timeout: 10 * 60 * 1000 // 10 分钟
    })
  }
  ,
  // 版本检查相关
  aBgiGetCurrentVersion: () => api.get('/api/aBgiUpdate/version'),
  aBgiGetLastVersion: () => api.post('/api/aBgiUpdate/GetLastVersion'),
  // 新的合并接口：获取当前和最新版本并返回 canUpdate
  aBgiGetVersions: () => api.get('/api/aBgiUpdate/GetBgiVersion'),
  aBgiUpdate: () => api.post('/api/aBgiUpdate/Update', {}, {
    timeout: 10 * 60 * 1000 // 10 分钟
  })
  ,
  // 通过 URL 下载并更新 BGI
// 通过 URL 下载并更新 BGI（单独超时）
  downloadBgi: () =>
    api.post(
      '/api/UpdateBgi/Download',
      {},
      {
        timeout: 10 * 60 * 1000 // 10 分钟
      }
    ),
  getBgiDownloadStatus: () => api.get('/api/UpdateBgi/DownloadStatus')
    
  
}


export default api
