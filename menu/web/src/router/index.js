import { createRouter, createWebHistory } from 'vue-router'

// 导入所有页面组件
import Home from '../views/Home.vue'
import Config from '../views/Config.vue'
import Log from '../views/Log.vue'
import LogAnalysis from '../views/LogAnalysis.vue'
import AutoLog from '../views/AutoLog.vue'
import Archive from '../views/Archive.vue'
import Other from '../views/Other.vue'
import JsNames from '../views/JsNames.vue'
import Pathing from '../views/Pathing.vue'
import ListGroups from '../views/ListGroups.vue'
import AutoArtifactsPro from '../views/AutoArtifactsPro.vue'
import AutoArtifactsPro2 from '../views/AutoArtifactsPro2.vue'
import Harvest from '../views/Harvest.vue'
import Bg from '../views/Bg.vue'
import Error from '../views/Error.vue'
import CalculateTaskEnabledList from '../views/CalculateTaskEnabledList.vue'
import BagStatistics from '../views/BagStatistics.vue'
import MaterialTrend from '../views/MaterialTrend.vue'
import Gitlog from '../views/Gitlog.vue'
import Online from '../views/online.vue'
import CDAwareAutoGather from '../views/CDAwareAutoGather.vue'
import screen from '../views/screen.vue'
import obsVideo from '../views/obsVideo.vue'
import TaskCron from '../views/TaskCron.vue'
import Login from '../views/Login.vue'
import BgiConfig from '../views/BgiConfig.vue'
import Update from '../views/Update.vue'
import CollectionManagement from '../views/CollectionManagement.vue'
import LogDetail from '../views/LogDetail.vue'
import Morale from '../views/Morale.vue'

// 路由配置
const routes = [
  { path: '/login', component: Login, meta: { requiresAuth: false } },
  { path: '/', component: Home, meta: { requiresAuth: true } },
  { path: '/Config', component: Config, meta: { requiresAuth: true } },
  { path: '/log', component: Log, meta: { requiresAuth: true } },
  { path: '/logAnalysis', component: LogAnalysis, meta: { requiresAuth: true } },
  { path: '/autoLog', component: AutoLog, meta: { requiresAuth: true } },
  { path: '/archive', component: Archive, meta: { requiresAuth: true } },
  { path: '/other', component: Other, meta: { requiresAuth: true } },
  { path: '/jsNames', component: JsNames, meta: { requiresAuth: true } },
  { path: '/Pathing', component: Pathing, meta: { requiresAuth: true } },
  { path: '/listGroups', component: ListGroups, meta: { requiresAuth: true } },
  { path: '/getAutoArtifactsPro', component: AutoArtifactsPro, meta: { requiresAuth: true } },
  { path: '/getAutoArtifactsPro2', component: AutoArtifactsPro2, meta: { requiresAuth: true } },
  { path: '/harvest', component: Harvest, meta: { requiresAuth: true } },
  { path: '/bg', component: Bg, meta: { requiresAuth: true } },
  { path: '/error', component: Error, meta: { requiresAuth: true } },
  { path: '/CalculateTaskEnabledList', component: CalculateTaskEnabledList, meta: { requiresAuth: true } },
  { path: '/BagStatistics', component: BagStatistics, meta: { requiresAuth: true } },
  { path: '/MaterialTrend', component: MaterialTrend, meta: { requiresAuth: true } },
  { path: '/Gitlog', component: Gitlog, meta: { requiresAuth: true } },
  { path: '/Online', component: Online, meta: { requiresAuth: true } },
  { path: '/CDAwareAutoGather', component: CDAwareAutoGather, meta: { requiresAuth: true } },
  { path: '/screen', component: screen, meta: { requiresAuth: true } },
  { path: '/obsVideo', component: obsVideo, meta: { requiresAuth: true } },
  { path: '/TaskCron', component: TaskCron, meta: { requiresAuth: true } }
  ,{ path: '/BgiConfig', component: BgiConfig, meta: { requiresAuth: true } }
  ,{ path: '/Update', component: Update, meta: { requiresAuth: true } }
  ,{ path: '/CollectionManagement', component: CollectionManagement, meta: { requiresAuth: true } }
  ,{ path: '/logDetail', component: LogDetail, meta: { requiresAuth: true } }
  ,{ path: '/Morale', component: Morale, meta: { requiresAuth: true } }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫 - 检查认证状态
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('aBgiToken')
  const requiresAuth = to.meta.requiresAuth

  // 如果路由需要认证但没有token，重定向到登录页
  if (requiresAuth && !token) {
    next('/login')
  }
  // 如果已登录且尝试访问登录页，重定向到首页
  else if (to.path === '/login' && token) {
    next('/')
  }
  // 其他情况正常导航
  else {
    next()
  }
})

export default router
