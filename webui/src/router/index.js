import Vue from 'vue'
import Router from 'vue-router'
import ServiceList from '@/views/ServiceList'

import Dashboard from '@/views/Dashboard'
import Docs from '@/views/Docs'
import KafkaCluster from '@/components/kafka/Cluster'
import KafkaTopic from '@/components/kafka/Topic'
import SmtpMail from '@/views/dashboard/SmtpMail'

import HttpService from '@/views/http/Service'
import ServiceOverview from '@/components/http/ServiceOverview'
import Endpoint from '@/components/http/Endpoint'
import Endpoints from '@/components/http/Endpoints'
import Models from '@/components/http/Models'

import SmtpService from '@/views/smtp/Service'

import KafkaService from '@/views/kafka/Service'

import LdapService from '@/views/ldap/Service'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'home',
      redirect: to => ({
        name: 'dashboard',
        query: {refresh: '20'}
      })
    },
    {
      path: '/docs',
      redirect: '/docs/welcome',
      name: 'docsStart'
    },
    {
      path: '/docs/:topic/:subject?',
      name: 'docs',
      component: Docs
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: Dashboard
    },
    {
      path: '/dashboard/http',
      name: 'http',
      component: Dashboard,
      meta: {showMetrics: true}
    },
    {
      path: '/dashboard/http/api/:service',
      name: 'httpService2',
      component: Dashboard
    },
    {
      path: '/dashboard/http/api/:service/:path',
      name: 'httpPath',
      component: Dashboard
    },
    {
      path: '/dashboard/http/requests/:id',
      name: 'httpRequest',
      component: Dashboard
    },
    {
      path: '/dashboard/kafka',
      name: 'kafka',
      component: Dashboard,
      meta: {showMetrics: true},
      children: [
        {
          path: '/dashboard/kafka/api/:cluster',
          name: 'kafkaCluster',
          component: KafkaCluster
        },
        {
          path: '/dashboard/kafka/api/:cluster/topics/:topic',
          name: 'kafkaTopic',
          component: KafkaTopic
        }
      ]
    },
    {
      path: '/dashboard/smtp',
      name: 'smtp',
      component: Dashboard,
      meta: {showMetrics: true}
    },
    {
      path: '/dashboard/smtp/mails/:id',
      name: 'smtpMail',
      component: SmtpMail
    },
    {
      path: '/services',
      name: 'serviceList',
      component: ServiceList
    },
    {
      path: '/services/http',
      name: 'httpServices',
      component: ServiceList
    },
    {
      path: '/services/kafka',
      name: 'kafkaServices',
      component: ServiceList
    },
    {
      path: '/services/ldap',
      name: 'ldapServices',
      component: ServiceList
    },
    {
      path: '/services/smtp',
      name: 'smtpServices',
      component: ServiceList
    },
    {
      path: '/services/kafka/:name',
      name: 'kafkaService',
      component: KafkaService
    },
    {
      path: '/services/smtp/:name',
      name: 'smtpService',
      component: SmtpService
    },
    {
      path: '/services/http/:name',
      name: 'httpService',
      component: HttpService,
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
          name: 'http-path',
          component: Endpoint
        }
      ]
    },
    {
      path: '/services/ldap/:name',
      name: 'ldap',
      component: LdapService
    }
  ]
})
