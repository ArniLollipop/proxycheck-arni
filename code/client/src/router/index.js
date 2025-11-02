import Vue from 'vue'
import VueRouter from 'vue-router'
import ProxyView from '../views/ProxyView.vue'
import LogsView from '../views/LogsView.vue'

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
    
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

export default router
