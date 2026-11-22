import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick, ref, type Ref } from 'vue'
import { mount } from '@vue/test-utils'

const event = ref<unknown>(null)
const isLoading = ref(true)
const closeEvent = vi.fn()
const getEvent = vi.fn()

vi.mock('@/composables/dashboard', () => ({
    useDashboard: () => ({ dashboard: ref({ getEvent }) }),
}))

import { useRequestEvent } from './request-event'

type JoinGroupData = { header: { requestName: string } }

const recorded = { id: 'e1', time: '2025-02-13T09:50:00+01:00', traits: { name: 'Kafka World' }, data: { header: { requestName: 'JoinGroup' } } }

function mountUseRequestEvent(id: Ref<string>) {
    let result!: ReturnType<typeof useRequestEvent<JoinGroupData>>
    const wrapper = mount(defineComponent({
        setup() {
            result = useRequestEvent<JoinGroupData>(() => id.value)
            return () => null
        },
    }))
    return { result, wrapper }
}

describe('useRequestEvent', () => {
    beforeEach(() => {
        event.value = null
        isLoading.value = true
        vi.clearAllMocks()
        getEvent.mockImplementation(() => ({ event, isLoading, close: closeEvent }))
    })

    it('loads the request by id and is loading until it arrives', () => {
        const { result } = mountUseRequestEvent(ref('e1'))

        expect(getEvent).toHaveBeenCalledWith('e1')
        expect(result.isLoading.value).toBe(true)
        expect(result.event.value).toBeNull()
        expect(result.data.value).toBeUndefined()
    })

    it('exposes the event and its data', async () => {
        const { result } = mountUseRequestEvent(ref('e1'))

        event.value = recorded
        isLoading.value = false
        await nextTick()

        expect(result.event.value).toEqual(recorded)
        expect(result.data.value).toEqual(recorded.data)
        expect(result.isLoading.value).toBe(false)
    })

    it('ends loading without an event for an unknown id', async () => {
        const { result } = mountUseRequestEvent(ref('gone'))

        isLoading.value = false
        await nextTick()

        expect(result.isLoading.value).toBe(false)
        expect(result.event.value).toBeNull()
        expect(result.data.value).toBeUndefined()
    })

    it('loads again when the id changes and closes the previous request', async () => {
        const id = ref('e1')
        mountUseRequestEvent(id)

        id.value = 'e2'
        await nextTick()

        expect(getEvent).toHaveBeenLastCalledWith('e2')
        expect(closeEvent).toHaveBeenCalledOnce()
    })

    it('closes the request when the view is removed', () => {
        const { wrapper } = mountUseRequestEvent(ref('e1'))

        wrapper.unmount()

        expect(closeEvent).toHaveBeenCalledOnce()
    })
})
