import { afterEach, describe, expect, it } from 'vitest'
import { createMessageLogFilter } from './message-log-filter'

type Field = 'cluster' | 'topic' | 'clientId'

const {
  addableFields,
  defaultFilter,
  fieldLabel,
  filterChips,
  filterStorageKey,
  fixedLabel,
  loadFilter,
  removeFilter,
  restoreFilter,
  saveFilter,
  toEventQuery,
} = createMessageLogFilter<Field>({
  storagePrefix: 'test-filter',
  addressKind: 'topic',
  fields: [
    { field: 'cluster', label: 'Cluster', trait: 'name', fixedBy: 'serviceName' },
    { field: 'topic', label: 'Topic', trait: 'topic', fixedBy: 'address' },
    { field: 'clientId', label: 'Client', trait: 'clientId', fixedBy: 'clientId' },
  ],
})

function filterWith(changes: Partial<Record<Field, string>>) {
  return { ...defaultFilter(), ...changes }
}

const all = {}
const cluster = { serviceName: 'Svc' }
const topic = { serviceName: 'Svc', address: 'orders' }
const instance = { serviceName: 'Svc', address: 'a.{id}', instance: 'x' }
const client = { serviceName: 'Svc', clientId: 'c1' }

describe('message log filter', () => {
  it('has no value by default', () => {
    expect(defaultFilter()).toEqual({ cluster: '', topic: '', clientId: '' })
  })

  describe('filterStorageKey', () => {
    it('uses the prefix of the log', () => {
      expect(filterStorageKey(all)).toBe('test-filter')
      expect(filterStorageKey(cluster)).toBe('test-filter-Svc')
      expect(filterStorageKey(topic)).toBe('test-filter-Svc-topic-orders')
      expect(filterStorageKey(instance)).toBe('test-filter-Svc-topic-a.{id}-x')
      expect(filterStorageKey(client)).toBe('test-filter-Svc-client-c1')
    })

    it('names the kind of the address of the log in the key', () => {
      const { filterStorageKey: channelKey } = createMessageLogFilter<Field>({
        storagePrefix: 'test-filter',
        addressKind: 'channel',
        fields: [{ field: 'topic', label: 'Path', trait: 'channel', fixedBy: 'address' }],
      })

      expect(channelKey(all)).toBe('test-filter')
      expect(channelKey(cluster)).toBe('test-filter-Svc')
      expect(channelKey(topic)).toBe('test-filter-Svc-channel-orders')
      expect(channelKey(instance)).toBe('test-filter-Svc-channel-a.{id}-x')
      expect(channelKey(client)).toBe('test-filter-Svc-client-c1')
    })

    it('does not mix up a topic and a client with the same name', () => {
      expect(filterStorageKey({ serviceName: 'Svc', address: 'x' }))
        .not.toBe(filterStorageKey({ serviceName: 'Svc', clientId: 'x' }))
    })
  })

  describe('restoreFilter', () => {
    it('is the default filter without a saved one', () => {
      expect(restoreFilter(null)).toEqual(defaultFilter())
      expect(restoreFilter(undefined)).toEqual(defaultFilter())
      expect(restoreFilter('')).toEqual(defaultFilter())
    })

    it('restores a saved filter and keeps defaults for missing fields', () => {
      expect(restoreFilter(JSON.stringify(filterWith({ topic: 'orders', clientId: 'c' }))))
        .toEqual(filterWith({ topic: 'orders', clientId: 'c' }))
      expect(restoreFilter('{"topic":"orders"}')).toEqual(filterWith({ topic: 'orders' }))
    })

    it('ignores a value of the wrong type and unknown fields', () => {
      expect(restoreFilter('{"topic":5,"clientId":"c","partition":"1"}')).toEqual(filterWith({ clientId: 'c' }))
    })

    it('is the default filter for corrupt or foreign content', () => {
      for (const saved of ['{oops', '42', 'null', '["orders"]']) {
        expect(restoreFilter(saved)).toEqual(defaultFilter())
      }
    })
  })

  describe('loadFilter and saveFilter', () => {
    afterEach(() => localStorage.clear())

    it('loads what was saved under the key', () => {
      saveFilter('test-filter-a', filterWith({ topic: 'orders' }))

      expect(loadFilter('test-filter-a')).toEqual(filterWith({ topic: 'orders' }))
      expect(loadFilter('test-filter-b')).toEqual(defaultFilter())
    })
  })

  describe('toEventQuery', () => {
    it('has no labels for the default filter', () => {
      expect(toEventQuery(defaultFilter(), all)).toEqual({})
    })

    it('maps every field to its trait and drops empty values', () => {
      expect(toEventQuery(filterWith({ cluster: 'Svc', clientId: 'c' }), all)).toEqual({
        labels: [{ name: 'name', value: 'Svc' }, { name: 'clientId', value: 'c' }],
      })
    })

    it('drops the fields the context fixes', () => {
      const filter = filterWith({ cluster: 'o', topic: 'o', clientId: 'o' })

      expect(toEventQuery(filter, topic)).toEqual({ labels: [{ name: 'clientId', value: 'o' }] })
      expect(toEventQuery(filter, client)).toEqual({ labels: [{ name: 'topic', value: 'o' }] })
      expect(toEventQuery(filterWith({ cluster: 'o' }), cluster)).toEqual({})
    })
  })

  describe('filterChips', () => {
    it('names the field, not the trait', () => {
      expect(filterChips(filterWith({ cluster: 'Svc', topic: 'orders' }), all).map(c => [c.field, c.label]))
        .toEqual([['cluster', 'Cluster: Svc'], ['topic', 'Topic: orders']])
    })

    it('has no chip for a field the context fixes', () => {
      const filter = filterWith({ cluster: 'o', topic: 'o', clientId: 'o' })

      expect(filterChips(filter, topic).map(c => c.field)).toEqual(['clientId'])
    })

    it('cuts a long label at 40 characters and keeps the full text as title', () => {
      const long = 'orders.'.repeat(10)

      const [chip] = filterChips(filterWith({ topic: long }), all)

      expect(chip.title).toBe(`Topic: ${long}`)
      expect(chip.label.length).toBe(40)
    })
  })

  describe('removeFilter', () => {
    it('resets only the given field and does not change the filter', () => {
      const filter = filterWith({ topic: 'orders', clientId: 'c' })

      expect(removeFilter(filter, 'topic')).toEqual(filterWith({ clientId: 'c' }))
      expect(filter.topic).toBe('orders')
    })
  })

  describe('addableFields', () => {
    it('offers the fields without chip in the order of the definition', () => {
      expect(addableFields(filterWith({ topic: 'orders' }), all).map(f => f.field)).toEqual(['cluster', 'clientId'])
    })

    it('does not offer the fields the context fixes', () => {
      expect(addableFields(defaultFilter(), topic).map(f => f.field)).toEqual(['clientId'])
      expect(addableFields(defaultFilter(), client).map(f => f.field)).toEqual(['topic'])
    })

    it('labels the fields', () => {
      expect(addableFields(defaultFilter(), all).map(f => f.label)).toEqual(['Cluster', 'Topic', 'Client'])
      expect(fieldLabel('clientId')).toBe('Client')
    })
  })

  describe('fixedLabel', () => {
    it('is the label of the trait that fixes a part of the context', () => {
      expect(fixedLabel('serviceName', cluster)).toEqual([{ name: 'name', value: 'Svc' }])
      expect(fixedLabel('address', topic)).toEqual([{ name: 'topic', value: 'orders' }])
      expect(fixedLabel('clientId', client)).toEqual([{ name: 'clientId', value: 'c1' }])
    })

    it('is empty for a part the context does not fix', () => {
      expect(fixedLabel('address', all)).toEqual([])
      expect(fixedLabel('clientId', topic)).toEqual([])
      expect(fixedLabel('address', { address: '' })).toEqual([])
    })
  })
})
