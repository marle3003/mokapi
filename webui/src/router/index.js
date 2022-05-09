import Vue from 'vue'
import Router from 'vue-router'
import ServiceList from '@/views/ServiceList'
import Service from '@/views/Service'
import Endpoint from '@/components/Endpoint'
import Endpoints from '@/components/Endpoints'
import Models from '@/components/Models'
import ServiceOverview from '@/components/ServiceOverview'
import Dashboard from '@/views/Dashboard'
import Docs from '@/views/Docs'
import HttpRequest from '@/views/dashboard/HttpRequest'
import SmtpMail from '@/views/dashboard/SmtpMail'
import SmtpService from '@/views/smtp/Service'
import KafkaCluster from '@/views/kafka/Cluster'
import KafkaTopic from '@/views/kafka/Topic'

import AsyncApiService from '@/views/asyncapi/Service'

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
      component: Dashboard
    },
    {
      path: '/dashboard/http/request/:id',
      name: 'httpRequest',
      component: HttpRequest
    },
    {
      path: '/dashboard/kafka',
      name: 'kafka',
      component: Dashboard
    },
    {
      path: '/dashboard/kafka/:cluster',
      name: 'kafkaCluster',
      component: KafkaCluster
    },
    {
      path: '/dashboard/kafka/:cluster/topics/:topic',
      name: 'kafkaTopic',
      component: KafkaTopic
    },
    {
      path: '/dashboard/smtp',
      name: 'smtp',
      component: Dashboard
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
      component: AsyncApiService
    },
    {
      path: '/services/smtp/:name',
      name: 'smtpService',
      component: SmtpService
    },
    {
      path: '/services/http/:name',
      name: 'httpService',
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
    {
      path: '/services/ldap/:name',
      name: 'ldap',
      component: LdapService
    }
  ]
})
