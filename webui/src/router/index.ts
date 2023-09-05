import { createRouter, createWebHistory, useRoute as baseRoute, type RouteLocationRaw, type RouteRecordRaw } from 'vue-router'
import DashboardView from '@/views/DashboardView.vue'

let base = document.querySelector("base")?.href ?? '/'
base = base.replace(document.location.origin, '')

export function useRoute() {
  const route = baseRoute()
  const context = {
    service: route.params.service?.toString(),
    path: route.params.path?.toString(),
    operation: route.params.operation?.toString()
  }
  
  function service(service: Service): RouteLocationRaw{
    return {
        name: 'httpService',
        params: { service: service.name },
        query: { refresh: route.query.refresh }
    }
  }

  function path(service: Service, path: HttpPath): RouteLocationRaw {
    return {
      name: 'httpPath',
      params: { service: service.name, path: path.path.substring(1) },
      query: { refresh: route.query.refresh }
    }
  }

  function operation(service: Service, path: HttpPath, operation: HttpOperation){
    return {
        name: 'httpOperation',
        params: { service: service.name, path: path.path.substring(1), operation: operation.method },
        query: { refresh: route.query.refresh }
    }
  }

  return {service, path, operation, context, router}
}

let startPageRoute: RouteRecordRaw
if (import.meta.env.VITE_DASHBOARD == 'true') {
  startPageRoute = {
    path: '/',
    name: 'home',
    redirect: to => {
      return {name: 'dashboard', query: {refresh: 20}}
    }
  }
} 
else {
  startPageRoute = {
    path: '/',
    name: 'home',
    component: () => {
      return import('@/views/Home.vue')
    },
  }
}

const router = createRouter({
  history: createWebHistory(base),
  scrollBehavior: (to, from, savedPosition) => {
    if (to.hash) {
      // if anchor is set go to element
      return {
        el: to.hash,
        behavior: 'smooth',
      }
    }
    // always scroll to top
    return { top: 0 }
  },
  routes: [
    startPageRoute,
    {
      path: '/home',
      component: () => import('@/views/Home.vue')
    },
    {
      path: '/http',
      component: () => import('@/views/Http.vue')
    },
    {
      path: '/kafka',
      component: () => import('@/views/Kafka.vue')
    },
    {
      path: '/ldap',
      component: () => import('@/views/Ldap.vue')
    },
    {
      path: '/smtp',
      component: () => import('@/views/Smtp.vue')
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
          path: '/dashboard/ldap',
          name: 'ldap',
          component: DashboardView,
          children: [
            {
              path: '/dashboard/ldap/service/:service',
              name: 'ldapService',
              component: DashboardView,
              meta: { service: 'ldap' },
            },
            {
              path: '/dashboard/ldap/searches/:id',
              name: 'ldapSearch',
              component: DashboardView,
              meta: { service: 'ldap' }
            },
          ]
        },
        {
          path: '/dashboard/smtp',
          name: 'smtp',
          component: DashboardView,
          children: [
            {
              path: '/dashboard/smtp/service/:service',
              name: 'smtpService',
              component: DashboardView,
              meta: { service: 'smtp' },
            },
            {
              path: '/dashboard/smtp/mails/:id',
              name: 'smtpMail',
              component: DashboardView,
              meta: { service: 'smtp' }
            },
          ]
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
