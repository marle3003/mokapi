import Search from '@/components/dashboard/Search.vue'
import { createRouter, createWebHistory, useRoute as baseRoute, type RouteLocationRaw, type RouteRecordRaw } from 'vue-router'

let base = document.querySelector("base")?.href ?? '/'
base = base.replace(document.location.origin, '')

export function useRoute() {
  const route = baseRoute()
  const context = {
    service: route.params.service?.toString(),
    path: route.params.path?.toString(),
    operation: route.params.operation?.toString()
  }

  function service(service: Service | string, type: string): RouteLocationRaw{
    let name;
    if (typeof service === 'string') {
      name = service
    } else {
      name = service.name
    }
    return {
        name: `${type}Service`,
        params: { service: name },
        query: { refresh: route.query.refresh }
    }
  }

  function httpPath(service: Service, path: HttpPath): RouteLocationRaw {
    return {
      name: 'httpEndpoint',
      params: { service: service.name, endpoint: path.path.substring(1).split('/') },
      query: { refresh: route.query.refresh }
    }
  }

  function httpOperation(service: Service, path: HttpPath, operation: HttpOperation){
    const endpoint = path.path.substring(1).split('/')
    endpoint.push(operation.method)

    return {
        name: 'httpEndpoint',
        params: { service: service.name, endpoint: endpoint },
        query: { refresh: route.query.refresh }
    }
  }

  return {service, httpPath, httpOperation, context, router, ...route}
}

const dashboardView = () => import('@/views/DashboardView.vue')

let startPageRoute: RouteRecordRaw
if (import.meta.env.VITE_DASHBOARD == 'true') {
  startPageRoute = {
    path: '/',
    name: 'home',
    redirect: () => {
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
    if (savedPosition) {
      return savedPosition
    }
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
      path: '/mail',
      component: () => import('@/views/Mail.vue')
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: dashboardView,
      children: [
        {
          path: '/services',
          name: 'serviceList',
          component: dashboardView
        },
        {
          path: '/dashboard/search',
          name: 'search',
          component: Search
        },
        {
          path: '/dashboard/http',
          name: 'http',
          component: dashboardView,
          children: [
            {
              path: '/dashboard/http/services/:service',
              name: 'httpService',
              component: dashboardView,
              meta: {service: 'http'}
            },
            {
              path: '/dashboard/http/requests/:id',
              name: 'httpRequest',
              component: dashboardView,
              meta: {service: 'http'}
            },
            {
              path: '/dashboard/http/services/:service/:endpoint(.*)*',
              name: 'httpEndpoint',
              component: dashboardView,
              meta: {service: 'http'}
            }
          ]
        },
        {
          path: '/dashboard/kafka',
          name: 'kafka',
          component: dashboardView,
          children: [
            {
              path: '/dashboard/kafka/service/:service',
              name: 'kafkaService',
              component: dashboardView,
              meta: {service: 'kafka'}
            },
            {
              path: '/dashboard/kafka/service/:service/topic/:topic',
              name: 'kafkaTopic',
              component: dashboardView,
              meta: {service: 'kafka'}
            },
            {
              path: '/dashboard/kafka/messages/:id',
              name: 'kafkaMessage',
              component: dashboardView,
              meta: {service: 'kafka'}
            }
          ]
        },
        {
          path: '/dashboard/ldap',
          name: 'ldap',
          component: dashboardView,
          children: [
            {
              path: '/dashboard/ldap/service/:service',
              name: 'ldapService',
              component: dashboardView,
              meta: { service: 'ldap' },
            },
            {
              path: '/dashboard/ldap/requests/:id',
              name: 'ldapRequest',
              component: dashboardView,
              meta: { service: 'ldap' }
            },
          ]
        },
        {
          path: '/dashboard/mail',
          name: 'mail',
          component: dashboardView,
          children: [
            {
              path: '/dashboard/mail/service/:service',
              name: 'mailService',
              component: dashboardView,
              meta: { service: 'mail' },
            },
            {
              path: '/dashboard/mail/service/:service/maiboxes/:name',
              name: 'smtpMailbox',
              component: dashboardView,
              meta: { service: 'mail' }
            },
            {
              path: '/dashboard/mail/mails/:id',
              name: 'smtpMail',
              component: dashboardView,
              meta: { service: 'mail' }
            },
          ]
        },
        {
          path: '/dashboard/jobs',
          name: 'jobs',
          component: dashboardView,
        },
        {
          path: '/dashboard/configs',
          name: 'configs',
          component: dashboardView,
        },
        {
          path: '/dashboard/config/:id',
          name: 'config',
          component: dashboardView,
        },
        {
          path: '/dashboard/tree',
          name: 'tree',
          component: dashboardView,
        },
      ]
    },
    {
      path: '/docs/examples/:pathMatch(.*)*',
      redirect: to => {
        if (typeof to.params.pathMatch === 'string') {
          return `/docs/resources/${to.params.pathMatch}`
        }
        return `/docs/resources/${to.params.pathMatch!.join('/')}`
      }
    },
    {
      path: '/docs',
      redirect: ({
        name: 'docs',
        params: {level1: 'guides', level2: 'welcome'}
      }),
      name: 'docsStart',
      children: [
        {
          path: '/docs/:level1/:level2?/:level3?/:level4?',
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

export function useRouter() {
  return router
}