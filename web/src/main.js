import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'
import App from './App.vue'

// 导入所有页面组件
import Home from './views/Home.vue'
import Config from './views/Config.vue'
import Log from './views/Log.vue'
import LogAnalysis from './views/LogAnalysis.vue'
import Archive from './views/Archive.vue'
import Other from './views/Other.vue'
import JsNames from './views/JsNames.vue'
import ListGroups from './views/ListGroups.vue'
import AutoArtifactsPro from './views/AutoArtifactsPro.vue'
import AutoArtifactsPro2 from './views/AutoArtifactsPro2.vue'
import Harvest from './views/Harvest.vue'
import Bg from './views/Bg.vue'
import OneLong from './views/OneLong.vue'
import Error from './views/Error.vue'
import CalculateTaskEnabledList from './views/CalculateTaskEnabledList.vue'
import BagStatistics from './views/BagStatistics.vue'

// 路由配置
const routes = [
  { path: '/', component: Home },
  { path: '/Config', component: Config },
  { path: '/log', component: Log },
  { path: '/logAnalysis', component: LogAnalysis },
  { path: '/archive', component: Archive },
  { path: '/other', component: Other },
  { path: '/jsNames', component: JsNames },
  { path: '/listGroups', component: ListGroups },
  { path: '/getAutoArtifactsPro', component: AutoArtifactsPro },
  { path: '/getAutoArtifactsPro2', component: AutoArtifactsPro2 },
  { path: '/harvest', component: Harvest },
  { path: '/bg', component: Bg },
  { path: '/onelong', component: OneLong },
  { path: '/error', component: Error },
  { path: '/CalculateTaskEnabledList', component: CalculateTaskEnabledList },
  { path: '/BagStatistics', component: BagStatistics }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

const app = createApp(App)
app.use(Antd)
app.use(router)
app.mount('#app') 