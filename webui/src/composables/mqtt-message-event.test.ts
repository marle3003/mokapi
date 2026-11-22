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

import { useMqttMessageEvent } from './mqtt-message-event'

const json = { name: 'Temperature', contentType: 'application/json', payload: { schema: { type: 'object' } } }
const avro = { name: 'avro', contentType: 'avro/binary', payload: { format: 'application/vnd.apache.avro;version=1.9.0' } }
const sensors = { name: 'sensors/{sensorId}/data', messages: { temperature: json, avro } }

function recorded(data: object = {}, traits: object = {}) {
    return {
        id: 'e1',
        traits: { name: 'MQTT World', topic: 'sensors/{sensorId}/data', ...traits },
        data: { topic: 'sensors/12345/data', messageId: 'temperature', message: { value: '{"temp":1}' }, ...data },
    }
}

function mountUseMqttMessageEvent(id: Ref<string>) {
    let result!: ReturnType<typeof useMqttMessageEvent>
    const wrapper = mount(defineComponent({
        setup() {
            result = useMqttMessageEvent(() => id.value)
            return () => null
        },
    }))
    return { result, wrapper }
}

describe('useMqttMessageEvent', () => {
    beforeEach(() => {
        event.value = null
        isLoading.value = true
        service.value = null
        vi.clearAllMocks()
        getEvent.mockImplementation(() => ({ event, isLoading, close: closeEvent }))
        getService.mockImplementation(() => ({ service, close: closeService }))
    })

    it('loads the message by id and is loading until it arrives', () => {
        const { result } = mountUseMqttMessageEvent(ref('e1'))

        expect(getEvent).toHaveBeenCalledWith('e1')
        expect(result.isLoading.value).toBe(true)
        expect(result.event.value).toBeNull()
        expect(result.data.value).toBeUndefined()
        expect(result.topic.value).toBeUndefined()
        expect(result.message.value).toBeUndefined()
        expect(getService).not.toHaveBeenCalled()
    })

    it('looks up the cluster and the topic of the message', async () => {
        const { result } = mountUseMqttMessageEvent(ref('e1'))

        event.value = recorded()
        isLoading.value = false
        service.value = { topics: [{ name: 'other', messages: {} }, sensors] }
        await nextTick()

        expect(getService).toHaveBeenCalledWith('MQTT World', 'mqtt')
        expect(result.data.value).toEqual(recorded().data)
        expect(result.topic.value).toEqual(sensors)
        expect(result.isLoading.value).toBe(false)
    })

    it('has no topic when the cluster has no topic of that name', async () => {
        const { result } = mountUseMqttMessageEvent(ref('e1'))

        event.value = recorded()
        service.value = { topics: [{ name: 'other', messages: {} }] }
        await nextTick()

        expect(result.topic.value).toBeUndefined()
    })

    it('forgets the topic when another message of another topic is shown', async () => {
        const id = ref('e1')
        const { result } = mountUseMqttMessageEvent(id)
        event.value = recorded()
        service.value = { topics: [sensors, { name: 'home', messages: { temperature: json } }] }
        await nextTick()
        expect(result.topic.value).toEqual(sensors)

        event.value = recorded({}, { topic: 'gone' })
        await nextTick()

        expect(result.topic.value).toBeUndefined()
        expect(result.message.value).toBeUndefined()
    })

    it('shows the value formatted for the content type', async () => {
        const { result } = mountUseMqttMessageEvent(ref('e1'))
        event.value = recorded()
        service.value = { topics: [sensors] }
        await nextTick()

        expect(result.message.value?.source.preview?.content).toBe('{\n  "temp": 1\n}')
        expect(result.message.value?.contentType).toBe('application/json')
        expect(result.message.value?.contentTypeTitle).toBe('application/json')
    })

    it('shows an Avro message as JSON and keeps the binary form', async () => {
        const { result } = mountUseMqttMessageEvent(ref('e1'))
        event.value = recorded({ messageId: 'avro', message: { value: '{"Name":"Carol"}', binary: btoa('abc') } })
        service.value = { topics: [sensors] }
        await nextTick()

        expect(result.message.value?.contentType).toBe('application/json')
        expect(result.message.value?.source.binary).toEqual({ content: 'abc', contentType: 'avro/binary' })
    })

    it('shows binary content that has no text value', async () => {
        const { result } = mountUseMqttMessageEvent(ref('e1'))
        event.value = recorded({ messageId: 'avro', message: { binary: btoa('abc') } })
        service.value = { topics: [sensors] }
        await nextTick()

        expect(result.message.value?.source.preview).toBeUndefined()
        expect(result.message.value?.source.binary).toEqual({ content: 'abc', contentType: 'avro/binary' })
    })

    it('has no message when its configuration is not known', async () => {
        const { result } = mountUseMqttMessageEvent(ref('e1'))
        vi.spyOn(console, 'error').mockImplementation(() => {})
        event.value = recorded({ messageId: 'gone' })
        service.value = { topics: [sensors] }
        await nextTick()

        expect(result.message.value).toBeUndefined()
    })

    it('tells whether event handlers ran', async () => {
        const { result } = mountUseMqttMessageEvent(ref('e1'))
        expect(result.hasActions.value).toBe(false)

        event.value = recorded({ actions: [{ duration: 20 }] })
        await nextTick()
        expect(result.hasActions.value).toBe(true)

        event.value = recorded({ actions: null })
        await nextTick()
        expect(result.hasActions.value).toBe(false)
    })

    it('loads again when the id changes and closes the previous message', async () => {
        const id = ref('e1')
        mountUseMqttMessageEvent(id)

        id.value = 'e2'
        await nextTick()

        expect(getEvent).toHaveBeenLastCalledWith('e2')
        expect(closeEvent).toHaveBeenCalledOnce()
    })

    it('closes the message and its cluster when the view is removed', async () => {
        const { wrapper } = mountUseMqttMessageEvent(ref('e1'))
        event.value = recorded()
        await nextTick()

        wrapper.unmount()

        expect(closeEvent).toHaveBeenCalledOnce()
        expect(closeService).toHaveBeenCalledOnce()
    })
})
