import { createRouter, createWebHashHistory } from 'vue-router'
import Dashboard from '../pages/Dashboard.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: Dashboard
    },
    {
      path: '/stream',
      name: 'stream',
      component: () => import('../pages/LiveStream.vue')
    },
    {
      path: '/alerts',
      name: 'alerts',
      component: () => import('../pages/Alerts.vue')
    },
    {
      path: '/rules',
      name: 'rules',
      component: () => import('../pages/Rules.vue')
    },
    {
      path: '/workloads',
      name: 'workloads',
      component: () => import('../pages/Workloads.vue')
    },
    {
      path: '/profiler',
      name: 'profiler',
      component: () => import('../pages/Profiler.vue')
    },
    {
      path: '/kernel',
      name: 'kernel',
      component: () => import('../pages/KernelXRay.vue')
    }
  ]
})

export default router

