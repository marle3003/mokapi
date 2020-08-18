import Vue from 'vue'
import Router from 'vue-router'
import ServiceList from '@/views/ServiceList'
import Service from '@/views/Service'
import Endpoint from '@/components/Endpoint'
import Endpoints from '@/components/Endpoints'
import Models from '@/components/Models'
import ServiceOverview from '@/components/ServiceOverview'
import Dashboard from '@/views/Dashboard'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'home',
      redirect: '/dashboard'
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: Dashboard
    },
    {
      path: '/services',
      name: 'serviceList',
      component: ServiceList
    },
    {
      path: '/services/:name',
      name: 'service',
      component: Service,
      redirect: {
        name: 'default'
      },
      children: [
        {
          path: '',
          name: 'default',
          component: ServiceOverview
        },
        {
          path: 'endpoints',
          name: 'endpoints',
          component: Endpoints
        },
        {
          path: 'models',
          name: 'models',
          component: Models
        },
        {
          path: 'endpoint/:path',
          name: 'endpoint',
          component: Endpoint
        }
      ]
    },
  ]
})
