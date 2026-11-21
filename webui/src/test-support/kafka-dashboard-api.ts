import type { Component } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import { clusters, events as kafkaEvents } from '../../e2e/mocks/kafka'
import { json, mountWithRoutes, notFound, stubFetch, viewlessRoutes, type MockEvent } from './dashboard-api'

// The Kafka data of the mock dashboard API (e2e/mocks/kafka.ts) for component tests.
export { clusters, kafkaEvents as events }

// Same answer as the operation `kafkaChannel` of e2e/mocks/kafka.ts, but without `metrics`; keep it
// in line with it. The message views need the channel for Key Type, Content Type and the value
// (`getKafkaChannel` in composables/dashboard.backend.ts). The mock handler isn't exported.
function answerChannel(url: URL) {
  const topicName = url.pathname.match(/^\/api\/services\/kafka\/[^/]+\/channels\/([^/]+)$/)?.[1]
  if (!topicName) {
    return undefined
  }
  const topic = clusters[0].topics.find(t => t.name === decodeURIComponent(topicName))
  if (!topic) {
    return notFound()
  }
  return json(Object.fromEntries(Object.entries(topic).filter(([key]) => key !== 'partitions')))
}

// `events` can be overridden to inject data the mock doesn't have.
export function stubDashboardFetch(events: MockEvent[] = kafkaEvents) {
  stubFetch(events, answerChannel)
}

// The routes the Kafka views link to, as in src/router/index.ts.
const routes = viewlessRoutes([
  { path: '/kafka/service/:service', name: 'kafkaService' },
  { path: '/kafka/log/:service', name: 'kafkaServiceLog' },
  { path: '/kafka/service/:service/topics/:topic', name: 'kafkaTopic' },
  { path: '/kafka/service/:service/groups/:group', name: 'kafkaGroup' },
  { path: '/kafka/service/:service/groups/:group/:member', name: 'kafkaGroupMember' },
  { path: '/kafka/messages/:id', name: 'kafkaMessage' },
  { path: '/kafka/service/:service/requests/:id', name: 'kafkaRequest' },
  { path: '/kafka/service/:service/clients/:clientId', name: 'kafkaClient' },
])

export function mountAt(component: Component, props: Record<string, unknown>, location: RouteLocationRaw) {
  return mountWithRoutes(routes, component, props, location)
}
