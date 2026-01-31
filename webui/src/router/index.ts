import Search from '@/components/dashboard/Search.vue'
import { getRouteName } from '@/composables/dashboard'
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
        name: getRouteName(`${type}Service`).value,
        params: { service: name },
    }
  }

  function httpPath(service: Service, path: HttpPath): RouteLocationRaw {
    return {
      name: getRouteName('httpEndpoint').value,
      params: { service: service.name, endpoint: path.path.substring(1).split('/') },
    }
  }

  function httpOperation(service: Service, path: HttpPath, operation: HttpOperation){
    const endpoint = path.path.substring(1).split('/')
    endpoint.push(operation.method)

    return {
        name: getRouteName('httpEndpoint').value,
        params: { service: service.name, endpoint: endpoint },
    }
  }

  return {service, httpPath, httpOperation, context, router, ...route}
}

const dashboardView = () => import('@/views/DashboardView.vue')

function createDashboardRoute(mode: 'live' | 'demo'): RouteRecordRaw {
  const getRouteName = (name: string) => mode === 'live' ? name : name + '-demo';

  return {
    path: mode === 'live' ? '/dashboard' : '/dashboard-demo',
    name: getRouteName('dashboard'),
    meta: {
        mode: mode
    },
    component: dashboardView,
    children: [
      {
        path: 'services',
        name: getRouteName('serviceList'),
        component: dashboardView
      },
      {
        path: 'search',
        name: getRouteName('search'),
        component: Search
      },
      {
        path: 'http',
        name: getRouteName('http'),
        component: dashboardView,
        children: [
          {
            path: 'services/:service',
            name: getRouteName('httpService'),
            component: dashboardView,
            meta: {service: 'http'}
          },
          {
            path: 'requests/:id',
            name: getRouteName('httpRequest'),
            component: dashboardView,
            meta: {service: 'http'}
          },
          {
            path: 'services/:service/:endpoint(.*)*',
            name: getRouteName('httpEndpoint'),
            component: dashboardView,
            meta: {service: 'http'}
          }
        ]
      },
      {
        path: 'kafka',
        name: getRouteName('kafka'),
        component: dashboardView,
        children: [
          {
            path: 'service/:service',
            name: getRouteName('kafkaService'),
            component: dashboardView,
            meta: {service: 'kafka'}
          },
          {
            path: 'service/:service/servers/:server',
            name: getRouteName('kafkaServer'),
            component: dashboardView,
            meta: {service: 'kafka'},
          },
          {
            path: 'service/:service/topics/:topic',
            name: getRouteName('kafkaTopic'),
            component: dashboardView,
            meta: {service: 'kafka'},
          },
          {
            path: 'service/:service/groups/:group',
            name: getRouteName('kafkaGroup'),
            component: dashboardView,
            meta: {service: 'kafka'}
          },
          {
            path: 'service/:service/groups/:group/:member',
            name: getRouteName('kafkaGroupMember'),
            component: dashboardView,
            meta: {service: 'kafka'}
          },
          {
            path: 'messages/:id',
            name: getRouteName('kafkaMessage'),
            component: dashboardView,
            meta: {service: 'kafka'}
          },
          {
            path: 'service/:service/requests/:id',
            name: getRouteName('kafkaRequest'),
            component: dashboardView,
            meta: {service: 'kafka'}
          },
          {
            path: 'service/:service/clients/:clientId',
            name: getRouteName('kafkaClient'),
            component: dashboardView,
            meta: {service: 'kafka'}
          }
        ]
      },
      {
        path: 'ldap',
        name: getRouteName('ldap'),
        component: dashboardView,
        children: [
          {
            path: 'service/:service',
            name: getRouteName('ldapService'),
            component: dashboardView,
            meta: { service: 'ldap' },
          },
          {
            path: 'ldap/requests/:id',
            name: getRouteName('ldapRequest'),
            component: dashboardView,
            meta: { service: 'ldap' }
          },
        ]
      },
      {
        path: 'mail',
        name: getRouteName('mail'),
        component: dashboardView,
        children: [
          {
            path: 'service/:service',
            name: getRouteName('mailService'),
            component: dashboardView,
            meta: { service: 'mail' },
          },
          {
            path: 'service/:service/maiboxes/:name',
            name: getRouteName('smtpMailbox'),
            component: dashboardView,
            meta: { service: 'mail' }
          },
          {
            path: 'mail/mails/:id',
            name: getRouteName('smtpMail'),
            component: dashboardView,
            meta: { service: 'mail' }
          },
        ]
      },
      {
        path: 'jobs',
        name: getRouteName('jobs'),
        component: dashboardView,
      },
      {
        path: 'configs',
        name: getRouteName('configs'),
        component: dashboardView,
      },
      {
        path: 'config/:id',
        name: getRouteName('config'),
        component: dashboardView,
      },
      {
        path: 'tree',
        name: getRouteName('tree'),
        component: dashboardView,
      },
    ]
  }
}

let startPageRoute: RouteRecordRaw
if (import.meta.env.VITE_WEBSITE == 'false') {
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
  scrollBehavior: async (to, from, savedPosition) => {
    if (savedPosition) {
      return savedPosition
    }
    if (to.hash) {
      await waitForElement(to.hash)
      // if anchor is set go to element
      return {
        el: to.hash,
        behavior: 'smooth' as const,
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

router.afterEach((to, from) => {
  if (!to.path.startsWith('/dashboard') || to.path.startsWith('/dashboard-demo')) {
    return
  }
        
  const hadRefresh = !!from.query.refresh;
  const hasRefresh = !!to.query.refresh;

  if (hadRefresh && !hasRefresh) {
    router.replace({
      ...to,
      query: {
        ...to.query,
        refresh: from.query.refresh
      }
    });
  }
});

if (import.meta.env.VITE_DASHBOARD === 'true') {
  router.addRoute(createDashboardRoute('live'));
}
if (import.meta.env.VITE_USE_DEMO === 'true') {
  router.addRoute(createDashboardRoute('demo'));
}


export default router

export function useRouter() {
  return router
}

function waitForElement(selector: string, timeout = 2000) {
  return new Promise((resolve) => {
    const start = Date.now()

    const check = () => {
      if (document.querySelector(selector)) {
        resolve(true)
      } else if (Date.now() - start > timeout) {
        resolve(false)
      } else {
        requestAnimationFrame(check)
      }
    }

    check()
  })
}