import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick, ref, type Ref } from 'vue'
import { mount } from '@vue/test-utils'

const event = ref<unknown>(null)
const isLoading = ref(true)
const channel = ref<unknown>(null)
const closeEvent = vi.fn()
const closeChannel = vi.fn()
const getEvent = vi.fn()
const getKafkaChannel = vi.fn()

vi.mock('@/composables/dashboard', () => ({
    useDashboard: () => ({ dashboard: ref({ getEvent, getKafkaChannel }) }),
}))

import { useKafkaMessageEvent } from './kafka-message-event'

const json = { name: 'OrderEvent', contentType: 'application/json', payload: { schema: { type: 'object' } }, key: { schema: { type: 'string' } } }
const avro = { name: 'avro', contentType: 'avro/binary', payload: { format: 'application/vnd.apache.avro;version=1.9.0' } }

function recorded(data: object = {}) {
    return {
        id: 'e1',
        traits: { name: 'Kafka World', topic: 'orders' },
        data: { messageId: 'order', message: { value: '{"id":1}' }, ...data },
    }
}

function mountUseKafkaMessageEvent(id: Ref<string>) {
    let result!: ReturnType<typeof useKafkaMessageEvent>
    const wrapper = mount(defineComponent({
        setup() {
            result = useKafkaMessageEvent(() => id.value)
            return () => null
        },
    }))
    return { result, wrapper }
}

describe('useKafkaMessageEvent', () => {
    beforeEach(() => {
        event.value = null
        isLoading.value = true
        channel.value = null
        vi.clearAllMocks()
        getEvent.mockImplementation(() => ({ event, isLoading, close: closeEvent }))
        getKafkaChannel.mockImplementation(() => ({ channel, close: closeChannel }))
    })

    it('loads the message by id and is loading until it arrives', () => {
        const { result } = mountUseKafkaMessageEvent(ref('e1'))

        expect(getEvent).toHaveBeenCalledWith('e1')
        expect(result.isLoading.value).toBe(true)
        expect(result.event.value).toBeNull()
        expect(result.data.value).toBeUndefined()
        expect(result.message.value).toBeUndefined()
        expect(getKafkaChannel).not.toHaveBeenCalled()
    })

    it('looks up the topic of the message in its cluster', async () => {
        const { result } = mountUseKafkaMessageEvent(ref('e1'))

        event.value = recorded()
        isLoading.value = false
        await nextTick()

        expect(getKafkaChannel).toHaveBeenCalledWith('Kafka World', 'orders')
        expect(result.data.value).toEqual(recorded().data)
        expect(result.isLoading.value).toBe(false)
    })

    it('shows the value formatted for the content type, and the key type', async () => {
        const { result } = mountUseKafkaMessageEvent(ref('e1'))
        event.value = recorded()
        channel.value = { messages: { order: json } }
        await nextTick()

        expect(result.message.value).toEqual({
            source: { preview: { content: '{\n  "id": 1\n}', contentType: 'application/json', contentTypeTitle: 'application/json', description: undefined } },
            contentType: 'application/json',
            contentTypeTitle: 'application/json',
            keyType: 'string',
        })
    })

    it('shows an Avro message as JSON and keeps the binary form', async () => {
        const { result } = mountUseKafkaMessageEvent(ref('e1'))
        event.value = recorded({ messageId: 'avro', message: { value: '{"Name":"Carol"}', binary: btoa('abc') } })
        channel.value = { messages: { avro } }
        await nextTick()

        expect(result.message.value?.contentType).toBe('application/json')
        expect(result.message.value?.contentTypeTitle).toBe('avro/binary')
        expect(result.message.value?.source.preview?.description).toBe('Avro content in JSON format')
        expect(result.message.value?.source.binary).toEqual({ content: 'abc', contentType: 'avro/binary' })
    })

    it('has no message when its configuration is not known', async () => {
        const { result } = mountUseKafkaMessageEvent(ref('e1'))
        vi.spyOn(console, 'error').mockImplementation(() => {})
        event.value = recorded({ messageId: 'gone' })
        channel.value = { messages: { order: json } }
        await nextTick()

        expect(result.message.value).toBeUndefined()
    })

    it('tells whether event handlers ran', async () => {
        const { result } = mountUseKafkaMessageEvent(ref('e1'))
        expect(result.hasActions.value).toBe(false)

        event.value = recorded({ actions: [{ duration: 20 }] })
        await nextTick()
        expect(result.hasActions.value).toBe(true)

        event.value = recorded({ actions: [] })
        await nextTick()
        expect(result.hasActions.value).toBe(false)
    })

    it('loads again when the id changes and closes the previous message', async () => {
        const id = ref('e1')
        mountUseKafkaMessageEvent(id)

        id.value = 'e2'
        await nextTick()

        expect(getEvent).toHaveBeenLastCalledWith('e2')
        expect(closeEvent).toHaveBeenCalledOnce()
    })

    it('closes the message and its topic when the view is removed', async () => {
        const { wrapper } = mountUseKafkaMessageEvent(ref('e1'))
        event.value = recorded()
        await nextTick()

        wrapper.unmount()

        expect(closeEvent).toHaveBeenCalledOnce()
        expect(closeChannel).toHaveBeenCalledOnce()
    })
})
