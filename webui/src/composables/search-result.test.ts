import { describe, expect, it } from 'vitest'
import { resultRoute } from './search-result'

function item(type: string, params: { [name: string]: string }): SearchItem {
  return { type, title: 'title', fragments: [], params: { type: type.toLowerCase(), ...params } }
}

describe('resultRoute: http', () => {
  it('opens the service', () => {
    const route = resultRoute(item('HTTP', { service: 'Petstore' }))

    expect(route).toEqual({ name: 'httpService', params: { type: 'http', service: 'Petstore' } })
  })

  it('opens the path as endpoint', () => {
    const route = resultRoute(item('HTTP', { service: 'Petstore', path: '/pet/store' }))

    expect(route).toMatchObject({ name: 'httpEndpoint', params: { endpoint: ['pet', 'store'], service: 'Petstore' } })
  })

  it('opens the operation of the path with the lower case method', () => {
    const route = resultRoute(item('HTTP', { service: 'Petstore', path: '/pet', method: 'POST' }))

    expect(route).toMatchObject({ name: 'httpEndpoint', params: { endpoint: ['pet', 'post'] } })
  })
})

describe('resultRoute: kafka', () => {
  it('opens the service', () => {
    expect(resultRoute(item('Kafka', { service: 'K' }))).toMatchObject({ name: 'kafkaService', params: { service: 'K' } })
  })

  it('opens the topic', () => {
    expect(resultRoute(item('Kafka', { service: 'K', topic: 't' }))).toMatchObject({ name: 'kafkaTopic', params: { topic: 't' } })
  })
})

describe('resultRoute: mail', () => {
  it('opens the service', () => {
    expect(resultRoute(item('Mail', { service: 'M' }))).toMatchObject({ name: 'mailService', params: { service: 'M' } })
  })

  it('opens the mailbox', () => {
    expect(resultRoute(item('Mail', { service: 'M', mailbox: 'alice' })))
      .toMatchObject({ name: 'smtpMailbox', params: { name: 'alice', service: 'M', mailbox: 'alice' } })
  })
})

describe('resultRoute: ldap and config', () => {
  it('opens the LDAP service', () => {
    expect(resultRoute(item('LDAP', { service: 'L' }))).toMatchObject({ name: 'ldapService', params: { service: 'L' } })
  })

  it('opens the config', () => {
    expect(resultRoute(item('Config', { id: 'abc' }))).toMatchObject({ name: 'config', params: { id: 'abc' } })
  })
})

describe('resultRoute: event', () => {
  const event = (namespace: string, params: { [name: string]: string } = {}) =>
    item('Event', { 'traits.namespace': namespace, ...params })

  it('opens an HTTP request', () => {
    expect(resultRoute(event('http', { id: '1' }))).toMatchObject({ name: 'httpRequest', params: { id: '1' } })
  })

  it('has no route for an HTTP log entry', () => {
    expect(resultRoute(event('http', { 'traits.type': 'log' }))).toBeUndefined()
  })

  it('opens a Kafka message', () => {
    expect(resultRoute(event('kafka', { id: '1' }))).toMatchObject({ name: 'kafkaMessage', params: { id: '1' } })
  })

  it('opens a mail by its message id', () => {
    expect(resultRoute(event('mail', { messageId: 'm@x' }))).toMatchObject({ name: 'smtpMail', params: { id: 'm@x' } })
  })

  it('opens an LDAP request', () => {
    expect(resultRoute(event('ldap', { id: '1' }))).toMatchObject({ name: 'ldapRequest', params: { id: '1' } })
  })

  it('has no route for an unknown namespace', () => {
    expect(resultRoute(event('mqtt'))).toBeUndefined()
  })
})

describe('resultRoute: unknown type', () => {
  it('has no route', () => {
    expect(resultRoute(item('Other', {}))).toBeUndefined()
  })
})
