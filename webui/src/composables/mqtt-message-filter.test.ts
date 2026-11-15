import { describe, expect, it } from 'vitest'
import {
  addableFields,
  defaultFilter,
  filterChips,
  filterStorageKey,
  removeFilter,
  restoreFilter,
  toEventQuery,
  type MqttMessageFilter,
} from './mqtt-message-filter'

function filterWith(changes: Partial<MqttMessageFilter>): MqttMessageFilter {
  return { ...defaultFilter(), ...changes }
}

const all = {}
const cluster = { serviceName: 'MQTT World' }
const topic = { serviceName: 'MQTT World', topic: 'sensors/{sensorId}/data' }
const instance = { serviceName: 'MQTT World', topic: 'sensors/{sensorId}/data', instance: 'sensors/12345/data' }
const client = { serviceName: 'MQTT World', clientId: 'mqtt-client-1' }

describe('mqtt message filter', () => {
  it('has the fields cluster, topic and client, no partition', () => {
    expect(defaultFilter()).toEqual({ cluster: '', topic: '', clientId: '' })
  })

  it('has a storage key per context with the prefix mqtt-messages-filter', () => {
    expect(filterStorageKey(all)).toBe('mqtt-messages-filter')
    expect(filterStorageKey(cluster)).toBe('mqtt-messages-filter-MQTT World')
    expect(filterStorageKey(topic)).toBe('mqtt-messages-filter-MQTT World-topic-sensors/{sensorId}/data')
    expect(filterStorageKey(instance)).toBe('mqtt-messages-filter-MQTT World-topic-sensors/{sensorId}/data-sensors/12345/data')
    expect(filterStorageKey(client)).toBe('mqtt-messages-filter-MQTT World-client-mqtt-client-1')
  })

  it('restores a saved filter and ignores the partition of a Kafka filter', () => {
    expect(restoreFilter('{"topic":"a/b","partition":"1"}')).toEqual(filterWith({ topic: 'a/b' }))
    expect(restoreFilter('{oops')).toEqual(defaultFilter())
  })

  it('maps cluster, topic and client to name, topic and clientId', () => {
    const filter = filterWith({ cluster: 'MQTT World', topic: 'sensors/{sensorId}/data', clientId: 'mokapi-script' })

    expect(toEventQuery(filter, all)).toEqual({
      labels: [
        { name: 'name', value: 'MQTT World' },
        { name: 'topic', value: 'sensors/{sensorId}/data' },
        { name: 'clientId', value: 'mokapi-script' },
      ],
    })
  })

  it('drops the fields a context fixes', () => {
    const filter = filterWith({ cluster: 'o', topic: 'o', clientId: 'o' })

    expect(toEventQuery(filter, cluster)).toEqual({ labels: [{ name: 'topic', value: 'o' }, { name: 'clientId', value: 'o' }] })
    expect(toEventQuery(filter, instance)).toEqual({ labels: [{ name: 'clientId', value: 'o' }] })
    expect(toEventQuery(filter, client)).toEqual({ labels: [{ name: 'topic', value: 'o' }] })
  })

  it('labels the chips Cluster, Topic and Client and cuts at 40 characters', () => {
    const long = 'sensors/'.repeat(10)

    const chips = filterChips(filterWith({ cluster: 'MQTT World', topic: long, clientId: 'c' }), all)

    expect(chips.map(c => c.field)).toEqual(['cluster', 'topic', 'clientId'])
    expect(chips[0].label).toBe('Cluster: MQTT World')
    expect(chips[1].label.length).toBe(40)
    expect(chips[1].title).toBe(`Topic: ${long}`)
    expect(chips[2].label).toBe('Client: c')
  })

  it('removes a field', () => {
    expect(removeFilter(filterWith({ topic: 't', clientId: 'c' }), 'topic')).toEqual(filterWith({ clientId: 'c' }))
  })

  it('offers per context the fields that are not fixed and have no chip', () => {
    expect(addableFields(defaultFilter(), all).map(f => f.field)).toEqual(['cluster', 'topic', 'clientId'])
    expect(addableFields(defaultFilter(), cluster).map(f => f.field)).toEqual(['topic', 'clientId'])
    expect(addableFields(defaultFilter(), topic).map(f => f.field)).toEqual(['clientId'])
    expect(addableFields(defaultFilter(), client).map(f => f.field)).toEqual(['topic'])
    expect(addableFields(filterWith({ topic: 't' }), all).map(f => f.field)).toEqual(['cluster', 'clientId'])
  })
})
