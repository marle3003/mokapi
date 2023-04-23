import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '@/views/DashboardView.vue'

let base = document.querySelector("base")?.href ?? '/'
base = base.replace(document.location.origin, '')

const router = createRouter({
  history: createWebHistory(base),
  routes: [
    {
      path: '/',
      name: 'home',
      redirect: to => {
        if (import.meta.env.VITE_DASHBOARD == 'true') {
          return {name: 'dashboard', query: {refresh: 20}}
        }
        return {path: '/home'}
      }
    },
    {
      path: '/home',
      component: () => import('@/views/Home.vue')
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: DashboardView,
      children: [
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
              path: '/dashboard/http/services/:service',
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
            {
              path: '/dashboard/http/services/:service/:path',
              name: 'httpPath',
              component: DashboardView,
              meta: {service: 'http'}
            },
            {
              path: '/dashboard/http/services/:service/:path/:operation',
              name: 'httpOperation',
              component: DashboardView,
              meta: {service: 'http'}
            },
            {
              path: '/dashboard/http/services/:service/:path/:operation/parameters/:parameter',
              name: 'httpParameter',
              component: DashboardView,
              meta: {service: 'http'}
            }
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
      ]
    },
    {
      path: '/docs',
      redirect: ({
        name: 'docs',
        params: {level1: 'Welcome'}
      }),
      name: 'docsStart',
      children: [
        {
          path: '/docs/:level1/:level2?/:level3?',
          name: 'docs',
          component: () => import('@/views/DocsView.vue')
        },
      ]
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/PageNotFound.vue')
    }
  ]
})

export default router
