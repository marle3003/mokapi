import { describe, expect, it } from 'vitest'
import { filterSidebar, isFilterActive, maxItemsPerService } from './sidebar-filter'

function service(name: string, type: string, items?: ServiceItem[]): Service {
  return { name, type, items } as Service
}

const petstore = service('Petstore', 'http', [
  { name: '/pet/{petId}', method: 'get', tags: ['pet'] },
  { name: '/pet', method: 'post', tags: ['pet', 'write'] },
  { name: '/store/order', method: 'get' },
])
const orders = service('Order Platform', 'kafka', [
  { name: 'order-created', tags: ['Orders'] },
  { name: 'order-shipped' },
])

describe('sidebar filter: active', () => {
  it('is inactive for an empty or whitespace-only query', () => {
    expect(isFilterActive('')).toBe(false)
    expect(isFilterActive('   ')).toBe(false)
    expect(isFilterActive(' a ')).toBe(true)
  })

  it('finds nothing for an inactive query', () => {
    const result = filterSidebar([petstore], '  ')

    expect(result.groups).toEqual([])
    expect(result.matches).toBe(0)
  })
})

describe('sidebar filter: matching', () => {
  it('matches the service name, ignoring case', () => {
    const result = filterSidebar([petstore, orders], 'PETSTORE')

    expect(result.groups).toHaveLength(1)
    expect(result.groups[0].service).toMatchObject({ name: 'Petstore', matched: true })
    expect(result.groups[0].items).toEqual([])
  })

  it('matches an item name', () => {
    const result = filterSidebar([petstore], 'store/ORDER')

    expect(result.groups[0].service.matched).toBe(false)
    expect(result.groups[0].items.map(i => i.name)).toEqual(['/store/order'])
  })

  it('matches the HTTP method', () => {
    const result = filterSidebar([petstore], 'post')

    expect(result.groups[0].items.map(i => `${i.method} ${i.name}`)).toEqual(['post /pet'])
  })

  it('matches a tag', () => {
    const result = filterSidebar([petstore, orders], 'orders')

    expect(result.groups.map(g => g.service.name)).toEqual(['Order Platform'])
    expect(result.groups[0].items.map(i => i.name)).toEqual(['order-created'])
  })

  it('ignores whitespace around the query', () => {
    expect(filterSidebar([orders], '  shipped ').groups[0].items.map(i => i.name)).toEqual(['order-shipped'])
  })

  it('leaves out services without a match', () => {
    expect(filterSidebar([petstore, orders], 'nothing').groups).toEqual([])
  })

  it('finds a service without items by its name only', () => {
    const ldap = service('Directory', 'ldap')

    expect(filterSidebar([ldap], 'direct').groups[0].service).toMatchObject({ name: 'Directory', matched: true })
    expect(filterSidebar([ldap], 'get').groups).toEqual([])
  })

  it('shows the service and its matching items when both match', () => {
    const result = filterSidebar([orders], 'order')

    expect(result.groups[0].service.matched).toBe(true)
    expect(result.groups[0].items.map(i => i.name)).toEqual(['order-created', 'order-shipped'])
  })
})

describe('sidebar filter: groups', () => {
  it('keeps the order of the services', () => {
    const result = filterSidebar([orders, petstore], 'o')

    expect(result.groups.map(g => g.service.name)).toEqual(['Order Platform', 'Petstore'])
  })

  it('sorts items by name, then method', () => {
    const api = service('Api', 'http', [
      { name: '/b', method: 'get' },
      { name: '/a', method: 'post' },
      { name: '/a', method: 'get' },
    ])

    const items = filterSidebar([api], '/').groups[0].items

    expect(items.map(i => `${i.method} ${i.name}`)).toEqual(['get /a', 'post /a', 'get /b'])
  })

  it('shows at most 6 items and counts the rest', () => {
    const items = Array.from({ length: 9 }, (_, i) => ({ name: `topic-${i}` }))
    const result = filterSidebar([service('Big', 'kafka', items)], 'topic')

    expect(maxItemsPerService).toBe(6)
    expect(result.groups[0].items.map(i => i.name)).toEqual(items.slice(0, 6).map(i => i.name))
    expect(result.groups[0].more).toBe(3)
  })

  it('has no rest for exactly 6 items', () => {
    const items = Array.from({ length: 6 }, (_, i) => ({ name: `topic-${i}` }))

    expect(filterSidebar([service('Big', 'kafka', items)], 'topic').groups[0].more).toBe(0)
  })

  it('shows the first 6 items in sorted order, not in source order', () => {
    const items = ['g', 'f', 'e', 'd', 'c', 'b', 'a'].map(name => ({ name: `t-${name}` }))

    const [group] = filterSidebar([service('Letters', 'kafka', items)], 't-').groups

    expect(group.items.map(i => i.name)).toEqual(['t-a', 't-b', 't-c', 't-d', 't-e', 't-f'])
    expect(group.more).toBe(1)
  })
})

describe('sidebar filter: count', () => {
  it('counts every matching item and every service matched by name', () => {
    // Order Platform: name + 2 items, Petstore: 2 items (tag pet)
    const result = filterSidebar([petstore, orders], 'o')

    expect(result.matches).toBe(1 + 2 + 3)
    expect(result.summary).toBe('6 matches in 2 services')
  })

  it('counts items that are not shown because of the limit', () => {
    const items = Array.from({ length: 9 }, (_, i) => ({ name: `topic-${i}` }))
    const result = filterSidebar([service('Big', 'kafka', items)], 'topic')

    expect(result.matches).toBe(9)
    expect(result.summary).toBe('9 matches in 1 service')
  })

  it('uses the singular for one match', () => {
    expect(filterSidebar([petstore], 'findme').summary).toBe('0 matches in 0 services')
    expect(filterSidebar([petstore], '/store').summary).toBe('1 match in 1 service')
  })
})

describe('sidebar filter: link targets', () => {
  const targets = (s: Service) => filterSidebar([s], 'x').groups[0]

  it('opens a service of its protocol', () => {
    const group = filterSidebar([service('x', 'kafka')], 'x').groups[0]

    expect(group.service.to).toEqual({ name: 'kafkaService', params: { service: 'x' } })
    expect(group.moreTo).toEqual(group.service.to)
  })

  it('opens an HTTP operation with the path segments and the method', () => {
    const group = filterSidebar([service('x', 'http', [{ name: '/pet/{petId}', method: 'get' }])], 'pet').groups[0]

    expect(group.items[0].to).toEqual({ name: 'httpEndpoint', params: { service: 'x', endpoint: ['pet', '{petId}', 'get'] } })
  })

  it('opens a Kafka and an MQTT topic', () => {
    expect(targets(service('x', 'kafka', [{ name: 'x.topic' }])).items[0].to)
      .toEqual({ name: 'kafkaTopic', params: { service: 'x', topic: 'x.topic' } })
    expect(targets(service('x', 'mqtt', [{ name: 'x/{id}' }])).items[0].to)
      .toEqual({ name: 'mqttTopic', params: { service: 'x', topic: 'x/{id}' } })
  })

  it('opens a Websocket channel', () => {
    expect(targets(service('x', 'websocket', [{ name: '/x' }])).items[0].to)
      .toEqual({ name: 'websocketChannel', params: { service: 'x', channel: '/x' } })
  })

  it('opens a mailbox', () => {
    expect(targets(service('x', 'mail', [{ name: 'x@mokapi.io' }])).items[0].to)
      .toEqual({ name: 'smtpMailbox', params: { service: 'x', name: 'x@mokapi.io' } })
  })
})
