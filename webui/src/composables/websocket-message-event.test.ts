import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick, ref, type Ref } from 'vue'
import { mount } from '@vue/test-utils'

const event = ref<unknown>(null)
const isLoading = ref(true)
const service = ref<unknown>(null)
const closeEvent = vi.fn()
const closeService = vi.fn()
const getEvent = vi.fn()
const getService = vi.fn()

vi.mock('@/composables/dashboard', () => ({
    useDashboard: () => ({ dashboard: ref({ getEvent, getService }) }),
}))

import { useWebsocketMessageEvent } from './websocket-message-event'

const json = { name: 'ChatMessage', contentType: 'application/json', payload: { schema: { type: 'object' } } }
const avro = { name: 'avro', contentType: 'avro/binary', payload: { format: 'application/vnd.apache.avro;version=1.9.0' } }
const chats = {
    name: '/chats/{chatId}',
    messages: { ChatMessage: json, avro },
    instances: [{ name: '/chats/1234' }],
}

function recorded(data: object = {}, traits: object = {}) {
    return {
        id: 'e1',
        traits: { name: 'Chat', channel: '/chats/{chatId}', instance: '/chats/1234', ...traits },
        data: { channel: '/chats/1234', messageId: 'ChatMessage', message: { value: '{"text":"hi"}' }, ...data },
    }
}

function mountUseWebsocketMessageEvent(id: Ref<string>) {
    let result!: ReturnType<typeof useWebsocketMessageEvent>
    const wrapper = mount(defineComponent({
        setup() {
            result = useWebsocketMessageEvent(() => id.value)
            return () => null
        },
    }))
    return { result, wrapper }
}

describe('useWebsocketMessageEvent', () => {
    beforeEach(() => {
        event.value = null
        isLoading.value = true
        service.value = null
        vi.clearAllMocks()
        getEvent.mockImplementation(() => ({ event, isLoading, close: closeEvent }))
        getService.mockImplementation(() => ({ service, close: closeService }))
    })

    it('loads the message by id and is loading until it arrives', () => {
        const { result } = mountUseWebsocketMessageEvent(ref('e1'))

        expect(getEvent).toHaveBeenCalledWith('e1')
        expect(result.isLoading.value).toBe(true)
        expect(result.event.value).toBeNull()
        expect(result.data.value).toBeUndefined()
        expect(result.channel.value).toBeUndefined()
        expect(result.value.value).toBeUndefined()
        expect(getService).not.toHaveBeenCalled()
    })

    it('looks up the service and the channel of the message', async () => {
        const { result } = mountUseWebsocketMessageEvent(ref('e1'))

        event.value = recorded()
        isLoading.value = false
        service.value = { channels: [{ name: '/chat', messages: {} }, chats] }
        await nextTick()

        expect(getService).toHaveBeenCalledWith('Chat', 'websocket')
        expect(result.data.value).toEqual(recorded().data)
        expect(result.channel.value).toEqual(chats)
        expect(result.isLoading.value).toBe(false)
    })

    it('has no channel when the service has no channel of that name', async () => {
        const { result } = mountUseWebsocketMessageEvent(ref('e1'))

        event.value = recorded()
        service.value = { channels: [{ name: '/chat', messages: {} }] }
        await nextTick()

        expect(result.channel.value).toBeUndefined()
        expect(result.message.value).toBeUndefined()
    })

    it('follows a message of another channel', async () => {
        const { result } = mountUseWebsocketMessageEvent(ref('e1'))
        event.value = recorded()
        service.value = { channels: [chats, { name: '/chat', messages: { ChatMessage: json } }] }
        await nextTick()
        expect(result.channel.value).toEqual(chats)

        event.value = recorded({ channel: '/chat' }, { channel: '/chat', instance: undefined })
        await nextTick()
        expect(result.channel.value?.name).toBe('/chat')

        event.value = recorded({}, { channel: '/gone' })
        await nextTick()
        expect(result.channel.value).toBeUndefined()
        expect(result.message.value).toBeUndefined()
    })

    it('shows the value formatted for the content type', async () => {
        const { result } = mountUseWebsocketMessageEvent(ref('e1'))
        event.value = recorded()
        service.value = { channels: [chats] }
        await nextTick()

        expect(result.message.value?.source.preview?.content).toBe('{\n  "text": "hi"\n}')
        expect(result.message.value?.contentType).toBe('application/json')
        expect(result.message.value?.contentTypeTitle).toBe('application/json')
        expect(result.value.value?.source.preview?.content).toBe('{\n  "text": "hi"\n}')
        expect(result.value.value?.hideContentType).toBe(false)
    })

    it('shows an Avro message as JSON and keeps the binary form', async () => {
        const { result } = mountUseWebsocketMessageEvent(ref('e1'))
        event.value = recorded({ messageId: 'avro', message: { value: '{"Name":"Carol"}', binary: btoa('abc') } })
        service.value = { channels: [chats] }
        await nextTick()

        expect(result.message.value?.contentType).toBe('application/json')
        expect(result.message.value?.source.binary).toEqual({ content: 'abc', contentType: 'avro/binary' })
    })

    it('shows binary content that has no text value', async () => {
        const { result } = mountUseWebsocketMessageEvent(ref('e1'))
        event.value = recorded({ messageId: 'avro', message: { binary: btoa('abc') } })
        service.value = { channels: [chats] }
        await nextTick()

        expect(result.message.value?.source.preview).toBeUndefined()
        expect(result.message.value?.source.binary).toEqual({ content: 'abc', contentType: 'avro/binary' })
    })

    it('shows the raw value of an invalid message that has no message id', async () => {
        const { result } = mountUseWebsocketMessageEvent(ref('e1'))
        event.value = recorded({ messageId: undefined, message: { value: '{"bar": "not a number"}' } })
        service.value = { channels: [chats] }
        await nextTick()

        expect(result.message.value).toBeUndefined()
        expect(result.value.value?.source.preview).toEqual({ content: '{"bar": "not a number"}', contentType: 'application/json' })
        expect(result.value.value?.hideContentType).toBe(true)
    })

    it('shows the raw value when the message configuration is not known', async () => {
        const { result } = mountUseWebsocketMessageEvent(ref('e1'))
        event.value = recorded({ messageId: 'gone' })
        service.value = { channels: [chats] }
        await nextTick()

        expect(result.message.value).toBeUndefined()
        expect(result.value.value?.source.preview?.content).toBe('{"text":"hi"}')
    })

    it('has no value to show for a message without any content and configuration', async () => {
        const { result } = mountUseWebsocketMessageEvent(ref('e1'))
        event.value = recorded({ messageId: undefined, message: {} })
        service.value = { channels: [chats] }
        await nextTick()

        expect(result.value.value).toBeUndefined()
    })

    it('tells whether event handlers ran', async () => {
        const { result } = mountUseWebsocketMessageEvent(ref('e1'))
        expect(result.hasActions.value).toBe(false)

        event.value = recorded({ actions: [{ duration: 20 }] })
        await nextTick()
        expect(result.hasActions.value).toBe(true)

        event.value = recorded({ actions: null })
        await nextTick()
        expect(result.hasActions.value).toBe(false)
    })

    it('tells whether a client is still known to the service', async () => {
        const { result } = mountUseWebsocketMessageEvent(ref('e1'))
        event.value = recorded()
        service.value = { channels: [chats], clients: [{ id: 'c1' }] }
        await nextTick()

        expect(result.isClientAvailable('c1')).toBe(true)
        expect(result.isClientAvailable('c2')).toBe(false)
    })

    it('knows no client when the service lists none', async () => {
        const { result } = mountUseWebsocketMessageEvent(ref('e1'))
        event.value = recorded()
        service.value = { channels: [chats] }
        await nextTick()

        expect(result.isClientAvailable('c1')).toBe(false)
    })

    it('loads again when the id changes and closes the previous message', async () => {
        const id = ref('e1')
        mountUseWebsocketMessageEvent(id)

        id.value = 'e2'
        await nextTick()

        expect(getEvent).toHaveBeenLastCalledWith('e2')
        expect(closeEvent).toHaveBeenCalledOnce()
    })

    it('closes the message and its service when the view is removed', async () => {
        const { wrapper } = mountUseWebsocketMessageEvent(ref('e1'))
        event.value = recorded()
        await nextTick()

        wrapper.unmount()

        expect(closeEvent).toHaveBeenCalledOnce()
        expect(closeService).toHaveBeenCalledOnce()
    })
})
