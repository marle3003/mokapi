import Search from '@/components/dashboard/Search.vue'
import { getRouteName } from '@/composables/dashboard'
import { siteBasePath } from '@/composables/site-base'
import { removeHeadTag, setHeadTag } from '@/composables/meta'
import { createRouter, createWebHistory, createMemoryHistory, useRoute as baseRoute, type RouteLocationRaw, type RouteRecordRaw, type MatcherLocation } from 'vue-router'

const base = siteBasePath()

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

  function httpOperation(service: Service, path: HttpPath, operation: HttpOperationInfo){
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
            meta: { 
              service: 'http',
              title: ({ params }: MatcherLocation) => `Dashboard - HTTP Request – ${params.id}`,
              noindex: true
            }
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
            meta: { 
              service: 'kafka',
              title: ({ params }: MatcherLocation) => `Dashboard - Kafka Message – ${params.id}`,
              noindex: true
            }
          },
          {
            path: 'service/:service/requests/:id',
            name: getRouteName('kafkaRequest'),
            component: dashboardView,
            meta: { 
              service: 'kafka',
              title: ({ params }: MatcherLocation) => `Dashboard - Kafka Request – ${params.id}`,
              noindex: true
            }
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
            meta: { 
              service: 'ldap',
              title: ({ params }: MatcherLocation) => `Dashboard - LDAP Request – ${params.id}`,
              noindex: true
            }
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
            meta: { 
              service: 'mail',
              title: ({ params }: MatcherLocation) => `Dashboard - Mail – ${params.id}`,
              noindex: true
            }
          },
        ]
      },
      {
        path: 'mqtt',
        name: getRouteName('mqtt'),
        component: dashboardView,
        meta: { title: 'Dashboard - MQTT' },
        children: [
          {
            path: 'service/:service',
            name: getRouteName('mqttService'),
            component: dashboardView,
            meta: { service: 'mqtt', title: ({ params }: MatcherLocation) => `Dashboard - MQTT – ${params.service}` }
          },
          {
            path: 'service/:service/servers/:server',
            name: getRouteName('mqttServer'),
            component: dashboardView,
            meta: { service: 'mqtt', title: ({ params }: MatcherLocation) => `Dashboard - MQTT Server – ${params.server}` },
          },
          {
            path: 'service/:service/topics/:topic',
            name: getRouteName('mqttTopic'),
            component: dashboardView,
            meta: {
              service: 'mqtt',
              title: ({ params }: MatcherLocation) => `Dashboard - MQTT Topic – ${params.topic}`
            },
          },
          {
            path: 'messages/:id',
            name: getRouteName('mqttMessage'),
            component: dashboardView,
            meta: { 
              service: 'mqtt',
              title: ({ params }: MatcherLocation) => `Dashboard - MQTT Message – ${params.id}`,
              noindex: true
            }
          },
          {
            path: 'service/:service/requests/:id',
            name: getRouteName('mqttRequest'),
            component: dashboardView,
            meta: { 
              service: 'mqtt',
              title: ({ params }: MatcherLocation) => `Dashboard - MQTT Request – ${params.id}`,
              noindex: true
            }
          },
          {
            path: 'service/:service/clients/:clientId',
            name: getRouteName('mqttClient'),
            component: dashboardView,
            meta: { service: 'mqtt', title: ({ params }: MatcherLocation) => `Dashboard - MQTT Client – ${params.clientId}` }
          }
        ]
      },
      {
        path: 'websocket',
        name: getRouteName('websocket'),
        component: dashboardView,
        meta: { title: 'Dashboard - Websocket' },
        children: [
          {
            path: 'service/:service',
            name: getRouteName('websocketService'),
            component: dashboardView,
            meta: { service: 'websocket', title: ({ params }: MatcherLocation) => `Dashboard - Websocket – ${params.service}` }
          },
          {
            path: 'service/:service/servers/:server',
            name: getRouteName('websocketServer'),
            component: dashboardView,
            meta: { service: 'websocket', title: ({ params }: MatcherLocation) => `Dashboard - Websocket Server – ${params.server}` },
          },
          {
            path: 'service/:service/channels/:channel',
            name: getRouteName('websocketChannel'),
            component: dashboardView,
            meta: {
              service: 'websocket',
              title: ({ params }: MatcherLocation) => `Dashboard - Websocket Channel – ${params.channel}`
            },
          },
          {
            path: 'messages/:id',
            name: getRouteName('websocketMessage'),
            component: dashboardView,
            meta: { 
              service: 'websocket',
              title: ({ params }: MatcherLocation) => `Dashboard - Websocket Message – ${params.id}`,
              noindex: true
            }
          },
          {
            path: 'service/:service/clients/:id',
            name: getRouteName('websocketClient'),
            component: dashboardView,
            meta: { service: 'websocket', title: ({ params }: MatcherLocation) => `Dashboard - Websocket Client – ${params.id}` }
          },
          {
            path: 'service/:service/events/:id',
            name: getRouteName('websocketEvent'),
            component: dashboardView,
            meta: { 
              service: 'websocket',
              title: ({ params }: MatcherLocation) => `Dashboard - Websocket Event – ${params.id}`,
              noindex: true
            }
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
        meta: { 
          title: ({ params }: MatcherLocation) => `Dashboard - Config – ${params.id}`,
          noindex: true
        }
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

// Landing pages and resources (blogs, tutorials) are only part of the website build, the
// dashboard embedded in the Mokapi binary links to mokapi.io instead.
const isWebsite = import.meta.env.VITE_WEBSITE === 'true'
const websitePages: RouteRecordRaw[] = isWebsite ? [
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
    path: '/mqtt',
    component: () => import('@/views/Mqtt.vue')
  },
  {
    path: '/websocket',
    component: () => import('@/views/Websocket.vue')
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
    path: '/compare',
    component: () => import('@/views/Compare.vue')
  },
] : []
const websiteResources: RouteRecordRaw[] = isWebsite ? [
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
] : []

const router = createRouter({
  history: import.meta.env.SSR ? createMemoryHistory(base) : createWebHistory(base),
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
    ...websitePages,
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
    ...websiteResources,
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/PageNotFound.vue'),
      meta: { noindex: true }
    }
  ]
})

// add refresh query parameter if clicked link does not have it but current route has it
router.beforeEach((to, from) => {
  if (!to.path.startsWith('/dashboard') || to.path.startsWith('/dashboard-demo')) {
    return 
  }
        
  const hadRefresh = !!from.query.refresh;
  const hasRefresh = !!to.query.refresh;

  if (hadRefresh && !hasRefresh) {
    return{
      path: to.path,
      query: {
        ...to.query,
        refresh: from.query.refresh
      },
      hash: to.hash,
    }
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
      } catch {
        resolve(false)
      }
    }

    check()
  })
}

// to.path is encoded and the router encodes the params again
function decodeSegment(segment: string): string {
  try {
    return decodeURIComponent(segment)
  } catch {
    return segment // malformed escape, e.g. /100%
  }
}

// 404.html is what the web server answers for every unknown URL, also for /docs/foo, which the router
// would match to DocsView. Its first navigation goes to the generic not found page instead, so the
// client keeps the pre-rendered page. A guard runs before the lazy view of the matched route loads.
if (!import.meta.env.SSR && document.getElementById('app')?.hasAttribute('data-not-found')) {
  const removeGuard = router.beforeEach((to) => {
    removeGuard()
    if (to.name !== 'not-found') {
      const pathMatch = to.path.slice(1).split('/').map(decodeSegment)
      return { name: 'not-found', params: { pathMatch }, query: to.query, hash: to.hash, replace: true }
    }
  })
}

router.afterEach((to) => {
  if (to.meta.noindex) {
    setHeadTag({ selector: 'meta[name="robots"]', tagName: 'meta', attributes: { name: 'robots', content: 'noindex' } })
  } else {
    removeHeadTag('meta[name="robots"]')
  }
})
