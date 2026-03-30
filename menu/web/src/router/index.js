import { createRouter, createWebHistory } from 'vue-router'

// 路由配置
const routes = [
  { path: '/login', component: () => import('../views/Login.vue'), meta: { requiresAuth: false } },
  { path: '/', component: () => import('../views/Home.vue'), meta: { requiresAuth: true } },
  { path: '/Config', component: () => import('../views/Config.vue'), meta: { requiresAuth: true } },
  { path: '/log', component: () => import('../views/Log.vue'), meta: { requiresAuth: true } },
  { path: '/logAnalysis', component: () => import('../views/LogAnalysis.vue'), meta: { requiresAuth: true } },
  { path: '/autoLog', component: () => import('../views/AutoLog.vue'), meta: { requiresAuth: true } },
  { path: '/archive', component: () => import('../views/Archive.vue'), meta: { requiresAuth: true } },
  { path: '/other', component: () => import('../views/Other.vue'), meta: { requiresAuth: true } },
  { path: '/jsNames', component: () => import('../views/JsNames.vue'), meta: { requiresAuth: true } },
  { path: '/Pathing', component: () => import('../views/Pathing.vue'), meta: { requiresAuth: true } },
  { path: '/listGroups', component: () => import('../views/ListGroups.vue'), meta: { requiresAuth: true } },
  { path: '/getAutoArtifactsPro', component: () => import('../views/AutoArtifactsPro.vue'), meta: { requiresAuth: true } },
  { path: '/getAutoArtifactsPro2', component: () => import('../views/AutoArtifactsPro2.vue'), meta: { requiresAuth: true } },
  { path: '/harvest', component: () => import('../views/Harvest.vue'), meta: { requiresAuth: true } },
  { path: '/bg', component: () => import('../views/Bg.vue'), meta: { requiresAuth: true } },
  { path: '/error', component: () => import('../views/Error.vue'), meta: { requiresAuth: true } },
  { path: '/CalculateTaskEnabledList', component: () => import('../views/CalculateTaskEnabledList.vue'), meta: { requiresAuth: true } },
  { path: '/BagStatistics', component: () => import('../views/BagStatistics.vue'), meta: { requiresAuth: true } },
  { path: '/MyBag', component: () => import('../views/MyBag.vue'), meta: { requiresAuth: true } },
  { path: '/MaterialTrend', component: () => import('../views/MaterialTrend.vue'), meta: { requiresAuth: true } },
  { path: '/Gitlog', component: () => import('../views/Gitlog.vue'), meta: { requiresAuth: true } },
  { path: '/Online', component: () => import('../views/online.vue'), meta: { requiresAuth: true } },
  { path: '/CDAwareAutoGather', component: () => import('../views/CDAwareAutoGather.vue'), meta: { requiresAuth: true } },
  { path: '/screen', component: () => import('../views/screen.vue'), meta: { requiresAuth: true } },
  { path: '/obsVideo', component: () => import('../views/obsVideo.vue'), meta: { requiresAuth: true } },
  { path: '/TaskCron', component: () => import('../views/TaskCron.vue'), meta: { requiresAuth: true } },
  { path: '/OneLongPlan', component: () => import('../views/OneLongPlan.vue'), meta: { requiresAuth: true } },
  { path: '/BgiConfig', component: () => import('../views/BgiConfig.vue'), meta: { requiresAuth: true } },
  { path: '/Update', component: () => import('../views/Update.vue'), meta: { requiresAuth: true } },
  { path: '/CollectionManagement', component: () => import('../views/CollectionManagement.vue'), meta: { requiresAuth: true } },
  { path: '/logDetail', component: () => import('../views/LogDetail.vue'), meta: { requiresAuth: true } },
  { path: '/Morale', component: () => import('../views/Morale.vue'), meta: { requiresAuth: true } },
  { path: '/HotKey', component: () => import('../views/HotKey.vue'), meta: { requiresAuth: true } },
  { path: '/OneRemote', component: () => import('../views/OneRemote.vue'), meta: { requiresAuth: true } }
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
