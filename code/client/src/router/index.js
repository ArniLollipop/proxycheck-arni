import Vue from 'vue'
import VueRouter from 'vue-router'
import ProxyView from '../views/ProxyView.vue'
import LogsView from '../views/LogsView.vue'
import SpeedView from '../views/SpeedLogView.vue'
import IpView from '../views/IpLogView.vue'
Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    name: 'proxy',
    component: ProxyView
  },
  {
    path: '/visit_logs',
    name: 'visit_logs',
    component: LogsView
    
  },
  {
    path: '/speed_logs',
    name: 'speed_logs',
    component: SpeedView
    
  },
  {
    path: '/ip_logs',
    name: 'ip_logs',
    component: IpView
    
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

export default router
