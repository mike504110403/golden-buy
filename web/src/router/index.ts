import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('../views/dashboard/Index.vue'),
    meta: { title: '儀表板' }
  },
  {
    path: '/orders',
    name: 'Orders',
    component: () => import('../views/orders/Index.vue'),
    meta: { title: '訂單管理' }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// 路由守衛
router.beforeEach((to, _from, next) => {
  // 設置頁面標題
  document.title = `${to.meta.title || 'Golden Buy'} - 貴金屬交易平台`
  next()
})

export default router

