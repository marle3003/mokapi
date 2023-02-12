import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '../views/DashboardView.vue'
import DocsView from '../views/DocsView.vue'

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
      path: '/dashboard',
      name: 'dashboard',
      component: DashboardView
    },
    {
      path: '/services',
      name: 'serviceList',
      component: DashboardView
    },
    {
      path: '/dashboard/http',
      name: 'http',
      component: DashboardView,
      children: [
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
        },
      ]
    },
    {
      path: '/dashboard/kafka',
      name: 'kafka',
      component: DashboardView,
      children: [
        {
          path: '/dashboard/kafka/service/:service',
          name: 'kafkaService',
          component: DashboardView,
          meta: {service: 'kafka'}
        },
        {
          path: '/dashboard/kafka/service/:service/topic/:topic',
          name: 'kafkaTopic',
          component: DashboardView,
          meta: {service: 'kafka'}
        }
      ]
    },
    {
      path: '/dashboard/smtp',
      name: 'smtp',
      component: DashboardView,
      meta: {service: 'smtp'}
    },
    {
      path: '/docs',
      redirect: ({
        name: 'docs',
        params: {topic: 'Welcome'}
      }),
      name: 'docsStart'
    },
    {
      path: '/docs/:topic/:subject?',
      name: 'docs',
      component: DocsView
    },
  ]
})

export default router
