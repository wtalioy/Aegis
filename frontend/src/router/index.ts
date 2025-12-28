import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/observatory',
      name: 'observatory',
      component: () => import('../pages/Observatory.vue')
    },
    {
      path: '/',
      redirect: '/observatory'
    },
    {
      path: '/sentinel',
      name: 'sentinel',
      component: () => import('../pages/Sentinel.vue')
    },
    {
      path: '/policy-studio',
      name: 'policy-studio',
      component: () => import('../pages/PolicyStudio.vue')
    },
    {
      path: '/investigation',
      name: 'investigation',
      component: () => import('../pages/Investigation.vue')
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('../pages/Settings.vue')
    },
    {
      path: '/rule-validation',
      name: 'rule-validation',
      component: () => import('../pages/RuleValidation.vue')
    }
  ]
})

export default router

