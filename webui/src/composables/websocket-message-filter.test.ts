import { describe, expect, it } from 'vitest'
import {
  addableFields,
  defaultFilter,
  filterChips,
  filterStorageKey,
  removeFilter,
  restoreFilter,
  toEventQuery,
  type WebsocketMessageFilter,
} from './websocket-message-filter'

function filterWith(changes: Partial<WebsocketMessageFilter>): WebsocketMessageFilter {
  return { ...defaultFilter(), ...changes }
}

const all = {}
const service = { serviceName: 'WebSocket Chat API' }
const channel = { serviceName: 'WebSocket Chat API', address: '/chat' }
const instance = { serviceName: 'WebSocket Chat API', address: '/chats/{chatId}', instance: '/chats/1234' }
const client = { serviceName: 'WebSocket Chat API', clientId: 'c-1' }

describe('websocket message filter', () => {
  it('has the fields service, channel (Path) and client', () => {
    expect(defaultFilter()).toEqual({ service: '', channel: '', clientId: '' })
  })

  it('has a storage key per context with the prefix websocket-messages-filter', () => {
    expect(filterStorageKey(all)).toBe('websocket-messages-filter')
    expect(filterStorageKey(service)).toBe('websocket-messages-filter-WebSocket Chat API')
    expect(filterStorageKey(channel)).toBe('websocket-messages-filter-WebSocket Chat API-channel-/chat')
    expect(filterStorageKey(instance)).toBe('websocket-messages-filter-WebSocket Chat API-channel-/chats/{chatId}-/chats/1234')
    expect(filterStorageKey(client)).toBe('websocket-messages-filter-WebSocket Chat API-client-c-1')
  })

  it('restores a saved filter with defaults for missing fields and ignores corrupt content', () => {
    expect(restoreFilter(null)).toEqual(defaultFilter())
    expect(restoreFilter('{"channel":"/chat"}')).toEqual(filterWith({ channel: '/chat' }))
    expect(restoreFilter('{"channel":5,"clientId":"c","partition":"1"}')).toEqual(filterWith({ clientId: 'c' }))
    expect(restoreFilter('{oops')).toEqual(defaultFilter())
  })

  it('maps service, path and client to name, channel and clientId', () => {
    const filter = filterWith({ service: 'WebSocket Chat API', channel: '/chats/{chatId}', clientId: 'c-1' })

    expect(toEventQuery(filter, all)).toEqual({
      labels: [
        { name: 'name', value: 'WebSocket Chat API' },
        { name: 'channel', value: '/chats/{chatId}' },
        { name: 'clientId', value: 'c-1' },
      ],
    })
  })

  it('drops empty values and the fields a context fixes', () => {
    const filter = filterWith({ service: 'o', channel: 'o', clientId: 'o' })

    expect(toEventQuery(defaultFilter(), all)).toEqual({})
    expect(toEventQuery(filter, service)).toEqual({ labels: [{ name: 'channel', value: 'o' }, { name: 'clientId', value: 'o' }] })
    expect(toEventQuery(filter, instance)).toEqual({ labels: [{ name: 'clientId', value: 'o' }] })
    expect(toEventQuery(filter, client)).toEqual({ labels: [{ name: 'channel', value: 'o' }] })
  })

  it('labels the chips Service, Path and Client and cuts at 40 characters', () => {
    const long = '/chats/'.repeat(10)

    const chips = filterChips(filterWith({ service: 'WebSocket Chat API', channel: long, clientId: 'c' }), all)

    expect(chips.map(c => c.field)).toEqual(['service', 'channel', 'clientId'])
    expect(chips[0].label).toBe('Service: WebSocket Chat API')
    expect(chips[1].label.length).toBe(40)
    expect(chips[1].title).toBe(`Path: ${long}`)
    expect(chips[2].label).toBe('Client: c')
  })

  it('removes a field', () => {
    expect(removeFilter(filterWith({ channel: '/chat', clientId: 'c' }), 'channel')).toEqual(filterWith({ clientId: 'c' }))
  })

  it('offers per context the fields that are not fixed and have no chip', () => {
    expect(addableFields(defaultFilter(), all).map(f => f.field)).toEqual(['service', 'channel', 'clientId'])
    expect(addableFields(defaultFilter(), service).map(f => f.field)).toEqual(['channel', 'clientId'])
    expect(addableFields(defaultFilter(), channel).map(f => f.field)).toEqual(['clientId'])
    expect(addableFields(defaultFilter(), client).map(f => f.field)).toEqual(['channel'])
    expect(addableFields(filterWith({ channel: '/chat' }), all).map(f => f.field)).toEqual(['service', 'clientId'])
  })
})
