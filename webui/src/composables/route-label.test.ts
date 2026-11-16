import { describe, expect, it } from 'vitest'
import { routeLabel } from './route-label'

describe('routeLabel', () => {
  it('shows the path of an HTTP endpoint', () => {
    expect(routeLabel('httpEndpoint', { service: 'Swagger Petstore', endpoint: ['pet', '{petId}'] })).toBe('/pet/{petId}')
  })

  it('shows the path of an HTTP operation without its method', () => {
    expect(routeLabel('httpEndpoint', { service: 'Swagger Petstore', endpoint: ['pet', '{petId}', 'get'] })).toBe('/pet/{petId}')
  })

  it('keeps a path segment that only looks like a method when it is the only one', () => {
    expect(routeLabel('httpEndpoint', { service: 'Api', endpoint: ['get'] })).toBe('/get')
  })

  it('names the request log of one service after the service', () => {
    expect(routeLabel('httpServiceLog', { service: 'Books API' })).toBe('Books API requests')
  })

  it('names the request log of all services "Requests"', () => {
    expect(routeLabel('httpLog', {})).toBe('Requests')
  })

  it('treats a demo request log like the live one', () => {
    expect(routeLabel('httpServiceLog-demo', { service: 'Books API' })).toBe('Books API requests')
    expect(routeLabel('httpLog-demo', {})).toBe('Requests')
  })

  it('names the message log of one cluster after the cluster', () => {
    expect(routeLabel('kafkaServiceLog', { service: 'Kafka World' })).toBe('Kafka World messages')
  })

  it('names the message log of all clusters "Messages"', () => {
    expect(routeLabel('kafkaLog', {})).toBe('Messages')
  })

  it('treats a demo message log like the live one', () => {
    expect(routeLabel('kafkaServiceLog-demo', { service: 'Kafka World' })).toBe('Kafka World messages')
    expect(routeLabel('kafkaLog-demo', {})).toBe('Messages')
  })

  it('names the message log of one MQTT cluster after the cluster', () => {
    expect(routeLabel('mqttServiceLog', { service: 'MQTT World' })).toBe('MQTT World messages')
  })

  it('names the message log of all MQTT clusters "Messages"', () => {
    expect(routeLabel('mqttLog', {})).toBe('Messages')
  })

  it('treats a demo MQTT message log like the live one', () => {
    expect(routeLabel('mqttServiceLog-demo', { service: 'MQTT World' })).toBe('MQTT World messages')
    expect(routeLabel('mqttLog-demo', {})).toBe('Messages')
  })

  it('names the message log of one WebSocket service after the service', () => {
    expect(routeLabel('websocketServiceLog', { service: 'Chat' })).toBe('Chat messages')
  })

  it('names the message log of all WebSocket services "Messages"', () => {
    expect(routeLabel('websocketLog', {})).toBe('Messages')
  })

  it('treats a demo WebSocket message log like the live one', () => {
    expect(routeLabel('websocketServiceLog-demo', { service: 'Chat' })).toBe('Chat messages')
    expect(routeLabel('websocketLog-demo', {})).toBe('Messages')
  })

  it.each([
    ['httpService', { service: 'Books API' }, 'Books API'],
    ['kafkaService', { service: 'Orders' }, 'Orders'],
    ['kafkaTopic', { service: 'Orders', topic: 'mokapi.shop.products' }, 'mokapi.shop.products'],
    ['kafkaServer', { service: 'Orders', server: 'broker' }, 'broker'],
    ['kafkaGroup', { service: 'Orders', group: 'shop' }, 'shop'],
    ['kafkaGroupMember', { service: 'Orders', group: 'shop', member: 'm1' }, 'm1'],
    ['kafkaClient', { service: 'Orders', clientId: 'client-1' }, 'client-1'],
    ['mqttTopic', { service: 'Home', topic: 'home/livingroom/temperature' }, 'home/livingroom/temperature'],
    ['mqttClient', { service: 'Home', clientId: 'c1' }, 'c1'],
    ['websocketChannel', { service: 'Chat', channel: '/chat' }, '/chat'],
    ['websocketClient', { service: 'Chat', id: 'w1' }, 'w1'],
    ['smtpMailbox', { service: 'Mail', name: 'alice@foo.bar' }, 'alice@foo.bar'],
    ['ldapService', { service: 'Directory' }, 'Directory'],
  ])('shows the name in %s', (name, params, expected) => {
    expect(routeLabel(name, params)).toBe(expected)
  })

  it.each([
    ['dashboard', 'Overview'],
    ['serviceList', 'Services'],
    ['search', 'Search'],
    ['http', 'HTTP'],
    ['kafka', 'Kafka'],
    ['ldap', 'LDAP'],
    ['mail', 'Mail'],
    ['mqtt', 'MQTT'],
    ['websocket', 'Websocket'],
    ['jobs', 'Jobs'],
    ['configs', 'Configs'],
    ['config', 'Config'],
    ['tree', 'Faker'],
    ['httpRequest', 'HTTP Request'],
    ['kafkaMessage', 'Kafka Message'],
    ['smtpMail', 'Mail'],
    ['websocketEvent', 'Websocket Event'],
  ])('names the page %s "%s"', (name, expected) => {
    expect(routeLabel(name, { id: 'x' })).toBe(expected)
  })

  it('treats a demo route like the live route', () => {
    expect(routeLabel('kafkaTopic-demo', { topic: 'orders' })).toBe('orders')
    expect(routeLabel('tree-demo', {})).toBe('Faker')
  })

  it('returns undefined for a route it does not know', () => {
    expect(routeLabel('docs', {})).toBeUndefined()
    expect(routeLabel(undefined, {})).toBeUndefined()
  })
})
