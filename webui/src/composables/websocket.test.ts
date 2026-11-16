import { describe, expect, it } from 'vitest'
import { findChannel, formatAddress, hasClient, formatType, messageDestination, messageSource, useWebsocket } from './websocket'

const client = { id: 'c1', address: '[::1]:53211', server: '127.0.0.1:8080', query: {}, headers: {} }

function messageData(direction: 'send' | 'receive'): WebsocketMessageData {
    return { direction, client } as WebsocketMessageData
}

describe('formatAddress', () => {
    it('shows the IPv6 loopback address as localhost', () => {
        expect(formatAddress('[::1]:53211')).toBe('localhost:53211')
    })

    it('keeps other addresses', () => {
        expect(formatAddress('127.0.0.1:53211')).toBe('127.0.0.1:53211')
    })

    it('keeps an empty address', () => {
        expect(formatAddress('')).toBe('')
    })
})

describe('formatType', () => {
    it('starts the type with a capital letter', () => {
        expect(formatType('connect')).toBe('Connect')
        expect(formatType('close')).toBe('Close')
    })
})

describe('useWebsocket', () => {
    it('offers the formatters', () => {
        const { formatAddress: address, formatType: type } = useWebsocket()

        expect(address('[::1]:1')).toBe('localhost:1')
        expect(type('close')).toBe('Close')
    })
})

describe('source and destination of a message', () => {
    it('is from the client to the server when the client sent the message', () => {
        const data = messageData('send')

        expect(messageSource(data)).toBe('localhost:53211')
        expect(messageDestination(data)).toBe('127.0.0.1:8080')
    })

    it('is from the server to the client when the server sent the message', () => {
        const data = messageData('receive')

        expect(messageSource(data)).toBe('127.0.0.1:8080')
        expect(messageDestination(data)).toBe('localhost:53211')
    })
})

describe('findChannel', () => {
    const service = {
        channels: [
            { name: '/chat' },
            { name: '/chats/{chatId}', instances: [{ name: '/chats/1234' }, { name: '/chats/5678' }] },
        ],
    } as WebsocketService

    it('finds a channel by its name', () => {
        expect(findChannel(service, '/chat')?.name).toBe('/chat')
        expect(findChannel(service, '/chats/{chatId}')?.name).toBe('/chats/{chatId}')
    })

    it('finds the template channel of an instance', () => {
        expect(findChannel(service, '/chats/5678')?.name).toBe('/chats/{chatId}')
    })

    it('finds nothing for an unknown path', () => {
        expect(findChannel(service, '/chats/general')).toBeUndefined()
    })
})


describe('hasClient', () => {
    it('finds a client the service lists', () => {
        expect(hasClient({ clients: [{ id: 'c1' }] } as WebsocketService, 'c1')).toBe(true)
        expect(hasClient({ clients: [{ id: 'c1' }] } as WebsocketService, 'c2')).toBe(false)
    })

    it('knows no client when the service lists none', () => {
        expect(hasClient({} as WebsocketService, 'c1')).toBe(false)
    })
})
