import { describe, expect, it } from 'vitest'
import { buildSidebarGroups } from './sidebar-groups'

function service(name: string, type: string, metrics: Record<string, number> = {}): Service {
  return { name, type, metrics } as Service
}

describe('sidebar groups', () => {
  it('has one group per active protocol in the fixed order', () => {
    const groups = buildSidebarGroups([], ['ldap', 'mail', 'websocket', 'mqtt', 'kafka', 'http'])

    expect(groups.map(g => g.title)).toEqual(['HTTP', 'Kafka', 'MQTT', 'Websocket', 'Mail', 'LDAP'])
  })

  it('leaves out protocols that are not active', () => {
    const groups = buildSidebarGroups([service('Orders', 'kafka')], ['kafka'])

    expect(groups.map(g => g.type)).toEqual(['kafka'])
  })

  it('keeps the order of the services within a group', () => {
    const groups = buildSidebarGroups(
      [service('A', 'http'), service('Orders', 'kafka'), service('B', 'http')],
      ['http', 'kafka'])

    expect(groups[0].services.map(s => s.name)).toEqual(['A', 'B'])
    expect(groups[1].services.map(s => s.name)).toEqual(['Orders'])
  })

  it('counts requests for HTTP and LDAP, messages for Kafka, MQTT and Websocket, mails for Mail', () => {
    const groups = buildSidebarGroups([
      service('h', 'http', { requests_total: 4 }),
      service('k', 'kafka', { messages_total: 5 }),
      service('m', 'mqtt', { messages_total: 6 }),
      service('w', 'websocket', { messages_total: 7 }),
      service('s', 'mail', { mails_total: 8 }),
      service('l', 'ldap', { requests_total: 9 }),
    ], ['http', 'kafka', 'mqtt', 'websocket', 'mail', 'ldap'])

    expect(groups.map(g => g.services[0].count)).toEqual([4, 5, 6, 7, 8, 9])
  })

  it('counts 0 when the metric is missing', () => {
    const groups = buildSidebarGroups([service('h', 'http')], ['http'])

    expect(groups[0].services[0].count).toBe(0)
  })

  it('marks an HTTP service with errors', () => {
    const groups = buildSidebarGroups([
      service('ok', 'http', { requests_errors_total: 0 }),
      service('bad', 'http', { requests_errors_total: 2 }),
    ], ['http'])

    expect(groups[0].services.map(s => s.hasErrors)).toEqual([false, true])
  })

  it('shows the error dot for HTTP only', () => {
    const groups = buildSidebarGroups([service('l', 'ldap', { requests_errors_total: 2 })], ['ldap'])

    expect(groups[0].services[0].hasErrors).toBe(false)
  })
})
