import Search from '@/components/dashboard/Search.vue'
import { getRouteName } from '@/composables/dashboard'
import { createRouter, createWebHistory, useRoute as baseRoute, type RouteLocationRaw, type RouteRecordRaw, type MatcherLocation } from 'vue-router'

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
        mode: mode,
        title: 'Dashboard'
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
        component: Search,
        meta: { title: 'Dashboard - Search'}
      },
      {
        path: 'http',
        name: getRouteName('http'),
        component: dashboardView,
        meta: { title: 'Dashboard - HTTP' },
        children: [
          {
            path: 'services/:service',
            name: getRouteName('httpService'),
            component: dashboardView,
            meta: { service: 'http', title: ({ params }: MatcherLocation) => `Dashboard - HTTP – ${params.service}` }
          },
          {
            path: 'requests/:id',
            name: getRouteName('httpRequest'),
            component: dashboardView,
            meta: { service: 'http', title: ({ params }: MatcherLocation) => `Dashboard - HTTP Request – ${params.id}` }
          },
          {
            path: 'services/:service/:endpoint(.*)*',
            name: getRouteName('httpEndpoint'),
            component: dashboardView,
            meta: { service: 'http', title: ({ params }: MatcherLocation) => `Dashboard - HTTP Endpoint – ${params.service}/${params.endpoint}` }
          }
        ]
      },
      {
        path: 'kafka',
        name: getRouteName('kafka'),
        component: dashboardView,
        meta: { title: 'Dashboard - Kafka' },
        children: [
          {
            path: 'service/:service',
            name: getRouteName('kafkaService'),
            component: dashboardView,
            meta: { service: 'kafka', title: ({ params }: MatcherLocation) => `Dashboard - Kafka – ${params.service}` }
          },
          {
            path: 'service/:service/servers/:server',
            name: getRouteName('kafkaServer'),
            component: dashboardView,
            meta: { service: 'kafka', title: ({ params }: MatcherLocation) => `Dashboard - Kafka Server – ${params.server}` },
          },
          {
            path: 'service/:service/topics/:topic',
            name: getRouteName('kafkaTopic'),
            component: dashboardView,
            meta: {
              service: 'kafka',
              title: ({ params }: MatcherLocation) => `Dashboard - Kafka Topic – ${params.topic}`
            },
          },
          {
            path: 'service/:service/groups/:group',
            name: getRouteName('kafkaGroup'),
            component: dashboardView,
            meta: { service: 'kafka', title: ({ params }: MatcherLocation) => `Dashboard - Kafka Group – ${params.group}` }
          },
          {
            path: 'service/:service/groups/:group/:member',
            name: getRouteName('kafkaGroupMember'),
            component: dashboardView,
            meta: { service: 'kafka', title: ({ params }: MatcherLocation) => `Dashboard - Kafka Group Member – ${params.member}` }
          },
          {
            path: 'messages/:id',
            name: getRouteName('kafkaMessage'),
            component: dashboardView,
            meta: { service: 'kafka', title: ({ params }: MatcherLocation) => `Dashboard - Kafka Message – ${params.id}` }
          },
          {
            path: 'service/:service/requests/:id',
            name: getRouteName('kafkaRequest'),
            component: dashboardView,
            meta: { service: 'kafka', title: ({ params }: MatcherLocation) => `Dashboard - Kafka Request – ${params.id}` }
          },
          {
            path: 'service/:service/clients/:clientId',
            name: getRouteName('kafkaClient'),
            component: dashboardView,
            meta: { service: 'kafka', title: ({ params }: MatcherLocation) => `Dashboard - Kafka Client – ${params.clientId}` }
          }
        ]
      },
      {
        path: 'ldap',
        name: getRouteName('ldap'),
        component: dashboardView,
        meta: { title: 'Dashboard - LDAP' },
        children: [
          {
            path: 'service/:service',
            name: getRouteName('ldapService'),
            component: dashboardView,
            meta: { service: 'ldap', title: ({ params }: MatcherLocation) => `Dashboard - LDAP Service – ${params.service}` },
          },
          {
            path: 'ldap/requests/:id',
            name: getRouteName('ldapRequest'),
            component: dashboardView,
            meta: { service: 'ldap', title: ({ params }: MatcherLocation) => `Dashboard - LDAP Request – ${params.id}` }
          },
        ]
      },
      {
        path: 'mail',
        name: getRouteName('mail'),
        component: dashboardView,
        meta: { title: 'Dashboard - Mail' },
        children: [
          {
            path: 'service/:service',
            name: getRouteName('mailService'),
            component: dashboardView,
            meta: { service: 'mail', title: ({ params }: MatcherLocation) => `Dashboard - Mail Service – ${params.service}` },
          },
          {
            path: 'service/:service/maiboxes/:name',
            name: getRouteName('smtpMailbox'),
            component: dashboardView,
            meta: { service: 'mail', title: ({ params }: MatcherLocation) => `Dashboard - Mailbox – ${params.name}` }
          },
          {
            path: 'mail/mails/:id',
            name: getRouteName('smtpMail'),
            component: dashboardView,
            meta: { service: 'mail', title: ({ params }: MatcherLocation) => `Dashboard - Mail – ${params.id}` }
          },
        ]
      },
      {
        path: 'jobs',
        name: getRouteName('jobs'),
        component: dashboardView,
        meta: { title: 'Dashboard - Jobs' }
      },
      {
        path: 'configs',
        name: getRouteName('configs'),
        component: dashboardView,
        meta: { title: 'Dashboard - Configs' }
      },
      {
        path: 'config/:id',
        name: getRouteName('config'),
        component: dashboardView,
        meta: { title: ({ params }: MatcherLocation) => `Dashboard - Config – ${params.id}` }
      },
      {
        path: 'tree',
        name: getRouteName('tree'),
        component: dashboardView,
        meta: { title: 'Dashboard - Tree' }
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
      if (await waitForElement(to.hash) ) {
        // if anchor is set go to element
        return {
          el: to.hash,
          behavior: 'smooth' as const,
        }
      }
    }

    const sameRoute =
      to.path === from.path &&
      JSON.stringify(to.query) === JSON.stringify(from.query) &&
      JSON.stringify(to.params) === JSON.stringify(from.params)

    if (sameRoute) {
      // only hash changed → keep scroll position
      return false
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
          return `/resources/${to.params.pathMatch}`
        }
        return `/resources/${to.params.pathMatch!.join('/')}`
      }
    },
    {
      path: '/docs/resources/:pathMatch(.*)*',
      redirect: to => {
        if (typeof to.params.pathMatch === 'string') {
          return `/resources/${to.params.pathMatch}`
        }
        return `/resources/${to.params.pathMatch!.join('/')}`
      }
    },
    {
      path: '/docs/guides/:pathMatch(.*)*',
      redirect: to => {
        if (typeof to.params.pathMatch === 'string') {
          return `/docs/${to.params.pathMatch}`
        }
        return `/docs/${to.params.pathMatch!.join('/')}`
      }
    },
    {
      path: '/docs',
      redirect: ({
        name: 'docs',
        params: { level1: 'welcome' }
      }),
       children: [
        {
          path: ':level1/:level2?/:level3?/:level4?',
          name: 'docs',
          component: () => import('@/views/DocsView.vue')
        },
      ]
    },
    {
      path: '/resources',
      component: () => import('@/views/DocsView.vue'),
      children: [
        {
          path: ':level1/:level2?/:level3?/:level4?',
          name: 'resources',
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

function waitForElement(selector: string, timeout = 2000): Promise<boolean> {
  return new Promise<boolean>((resolve) => {
    const start = Date.now()

    const check = () => {
      try {
        if (document.querySelector(selector)) {
          resolve(true)
        } else if (Date.now() - start > timeout) {
          resolve(false)
        } else {
          requestAnimationFrame(check)
        }
      } catch(e) {
        resolve(false)
      }
    }

    check()
  })
}