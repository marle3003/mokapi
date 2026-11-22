import { afterEach, describe, expect, it } from 'vitest'
import {
  addableFields,
  defaultFilter,
  fieldLabel,
  filterChips,
  filterStorageKey,
  loadFilter,
  removeFilter,
  restoreFilter,
  saveFilter,
  toEventQuery,
  type KafkaMessageFilter,
} from './kafka-message-filter'

function filterWith(changes: Partial<KafkaMessageFilter>): KafkaMessageFilter {
  return { ...defaultFilter(), ...changes }
}

const all = {}
const cluster = { serviceName: 'Kafka World' }
const topic = { serviceName: 'Kafka World', address: 'orders' }
const instance = { serviceName: 'Kafka World', address: 'shipments.{region}.updates', instance: 'eu' }
const client = { serviceName: 'Kafka World', clientId: 'consumer-1' }

describe('kafka message filter', () => {
  describe('storageKey', () => {
    it('is one key for all clusters', () => {
      expect(filterStorageKey(all)).toBe('kafka-messages-filter')
    })

    it('is one key per cluster', () => {
      expect(filterStorageKey(cluster)).toBe('kafka-messages-filter-Kafka World')
    })

    it('is one key per topic of a cluster', () => {
      expect(filterStorageKey(topic)).toBe('kafka-messages-filter-Kafka World-topic-orders')
    })

    it('is one key per instance of a template topic', () => {
      expect(filterStorageKey(instance)).toBe('kafka-messages-filter-Kafka World-topic-shipments.{region}.updates-eu')
    })

    it('is one key per client of a cluster', () => {
      expect(filterStorageKey(client)).toBe('kafka-messages-filter-Kafka World-client-consumer-1')
    })

    it('does not mix up a topic and a client with the same name', () => {
      expect(filterStorageKey({ serviceName: 'Kafka World', address: 'x' }))
        .not.toBe(filterStorageKey({ serviceName: 'Kafka World', clientId: 'x' }))
    })
  })

  describe('restoreFilter', () => {
    it('is the default filter without a saved one', () => {
      expect(restoreFilter(null)).toEqual(defaultFilter())
      expect(restoreFilter(undefined)).toEqual(defaultFilter())
      expect(restoreFilter('')).toEqual(defaultFilter())
    })

    it('has no value by default', () => {
      expect(defaultFilter()).toEqual({ cluster: '', topic: '', partition: '', clientId: '' })
    })

    it('restores a saved filter', () => {
      const saved = filterWith({ cluster: 'Kafka World', topic: 'orders', partition: '1', clientId: 'consumer-1' })

      expect(restoreFilter(JSON.stringify(saved))).toEqual(saved)
    })

    it('keeps the default of a field that is missing', () => {
      expect(restoreFilter('{"topic":"orders"}')).toEqual(filterWith({ topic: 'orders' }))
    })

    it('ignores a value of the wrong type and unknown fields', () => {
      expect(restoreFilter('{"topic":5,"partition":null,"clientId":"c","other":"x"}')).toEqual(filterWith({ clientId: 'c' }))
    })

    it('is the default filter for corrupt or foreign content', () => {
      expect(restoreFilter('{oops')).toEqual(defaultFilter())
      expect(restoreFilter('42')).toEqual(defaultFilter())
      expect(restoreFilter('null')).toEqual(defaultFilter())
      expect(restoreFilter('["orders"]')).toEqual(defaultFilter())
    })
  })

  describe('loadFilter and saveFilter', () => {
    afterEach(() => localStorage.clear())

    it('loads the default filter when nothing is saved', () => {
      expect(loadFilter('kafka-messages-filter')).toEqual(defaultFilter())
    })

    it('loads what was saved under the key', () => {
      const filter = filterWith({ topic: 'orders' })

      saveFilter('kafka-messages-filter-a', filter)

      expect(loadFilter('kafka-messages-filter-a')).toEqual(filter)
      expect(loadFilter('kafka-messages-filter-b')).toEqual(defaultFilter())
    })

    it('loads the default filter for corrupt content', () => {
      localStorage.setItem('kafka-messages-filter', '{oops')

      expect(loadFilter('kafka-messages-filter')).toEqual(defaultFilter())
    })
  })

  describe('toEventQuery', () => {
    it('has no labels for the default filter', () => {
      expect(toEventQuery(defaultFilter(), all)).toEqual({})
    })

    it('maps every field to its trait', () => {
      const filter = filterWith({ cluster: 'Kafka World', topic: 'orders', partition: '2', clientId: 'consumer-1' })

      expect(toEventQuery(filter, all)).toEqual({
        labels: [
          { name: 'name', value: 'Kafka World' },
          { name: 'topic', value: 'orders' },
          { name: 'partition', value: '2' },
          { name: 'clientId', value: 'consumer-1' },
        ],
      })
    })

    it('keeps partition 0', () => {
      expect(toEventQuery(filterWith({ partition: '0' }), all)).toEqual({ labels: [{ name: 'partition', value: '0' }] })
    })

    it('drops empty values', () => {
      expect(toEventQuery(filterWith({ topic: '', clientId: 'c' }), all)).toEqual({ labels: [{ name: 'clientId', value: 'c' }] })
    })

    it('drops the cluster when the context fixes it', () => {
      expect(toEventQuery(filterWith({ cluster: 'Other', partition: '1' }), cluster)).toEqual({ labels: [{ name: 'partition', value: '1' }] })
    })

    it('drops the topic when the context fixes it', () => {
      expect(toEventQuery(filterWith({ topic: 'other', clientId: 'c' }), topic)).toEqual({ labels: [{ name: 'clientId', value: 'c' }] })
    })

    it('drops the client when the context fixes it', () => {
      expect(toEventQuery(filterWith({ clientId: 'other', topic: 'orders' }), client)).toEqual({ labels: [{ name: 'topic', value: 'orders' }] })
    })

    it('has no labels when only fixed fields are set', () => {
      expect(toEventQuery(filterWith({ cluster: 'Other' }), cluster)).toEqual({})
    })
  })

  describe('filterChips', () => {
    it('has no chip for the default filter', () => {
      expect(filterChips(defaultFilter(), all)).toEqual([])
    })

    it('names the field, not the trait', () => {
      const filter = filterWith({ cluster: 'Kafka World', topic: 'orders', partition: '0', clientId: 'consumer-1' })

      expect(filterChips(filter, all).map(c => [c.field, c.label])).toEqual([
        ['cluster', 'Cluster: Kafka World'],
        ['topic', 'Topic: orders'],
        ['partition', 'Partition: 0'],
        ['clientId', 'Client: consumer-1'],
      ])
    })

    it('has no chip for a field the context fixes', () => {
      const filter = filterWith({ cluster: 'Other', topic: 'other', clientId: 'other', partition: '1' })

      expect(filterChips(filter, topic).map(c => c.field)).toEqual(['partition', 'clientId'])
      expect(filterChips(filter, client).map(c => c.field)).toEqual(['topic', 'partition'])
    })

    it('cuts a long label and keeps the full text as title', () => {
      const long = 'orders.'.repeat(10)

      const [chip] = filterChips(filterWith({ topic: long }), all)

      expect(chip.title).toBe(`Topic: ${long}`)
      expect(chip.label.length).toBe(40)
      expect(chip.label.endsWith('…')).toBe(true)
    })

    it('does not cut a label of 40 characters', () => {
      const value = 'x'.repeat(40 - 'Topic: '.length)

      expect(filterChips(filterWith({ topic: value }), all)[0].label).toBe(`Topic: ${value}`)
    })
  })

  describe('removeFilter', () => {
    it('resets only the given field', () => {
      const filter = filterWith({ topic: 'orders', partition: '1' })

      expect(removeFilter(filter, 'topic')).toEqual(filterWith({ partition: '1' }))
    })

    it('does not change the given filter', () => {
      const filter = filterWith({ topic: 'orders' })

      removeFilter(filter, 'topic')

      expect(filter.topic).toBe('orders')
    })
  })

  describe('addableFields', () => {
    it('offers all fields for all clusters, in the order of the chips', () => {
      expect(addableFields(defaultFilter(), all).map(f => f.field)).toEqual(['cluster', 'topic', 'partition', 'clientId'])
    })

    it('does not offer a field that has a chip', () => {
      expect(addableFields(filterWith({ topic: 'orders' }), all).map(f => f.field)).toEqual(['cluster', 'partition', 'clientId'])
    })

    it('does not offer the fields the context fixes', () => {
      expect(addableFields(defaultFilter(), cluster).map(f => f.field)).toEqual(['topic', 'partition', 'clientId'])
      expect(addableFields(defaultFilter(), topic).map(f => f.field)).toEqual(['partition', 'clientId'])
      expect(addableFields(defaultFilter(), client).map(f => f.field)).toEqual(['topic', 'partition'])
    })

    it('labels the fields', () => {
      expect(addableFields(defaultFilter(), all).map(f => f.label)).toEqual(['Cluster', 'Topic', 'Partition', 'Client'])
      expect(fieldLabel('clientId')).toBe('Client')
    })
  })
})
