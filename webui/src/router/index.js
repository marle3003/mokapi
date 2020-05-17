import Vue from 'vue'
import Router from 'vue-router'
import Home from '@/views/Home'
import ServiceList from '@/views/ServiceList'
import Service from '@/views/Service'
import Endpoints from '@/components/Endpoints'
import Endpoint from '@/components/Endpoint'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
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
          component: Endpoints
        },
        {
          path: ':path',
          name: 'endpoint',
          component: Endpoint
        }
      ]
    },
  ]
})
