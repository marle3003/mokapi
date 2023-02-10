import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import DashboardView from '../views/DashboardView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      redirect: ({
        name: 'dashboard',
        query: {refresh: 20}
      })
    },
    {
      path: '/about',
      name: 'about',
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import('../views/AboutView.vue')
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: DashboardView
    },
    {
      path: '/222',
      name: 'serviceList',
      component: HomeView
    },
    {
      path: '/222',
      name: 'docsStart',
      component: HomeView
    },
    {
      path: '/dashboard/http',
      name: 'http',
      component: DashboardView,
    },
    {
      path: '/dashboard/kafka',
      name: 'kafka',
      component: DashboardView
    },
    {
      path: '/dashboard/smtp',
      name: 'smtp',
      component: DashboardView
    },
    {
      path: '/dashboard/http/service/:service',
      name: 'httpService',
      component: DashboardView,
      meta: {service: 'http'}
    },
    {
      path: '/dashboard/http/requests/:id',
      name: 'httpRequest',
      component: DashboardView,
      meta: {service: 'http'}
    }
  ]
})

export default router
