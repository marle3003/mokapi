import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick, ref, type Ref } from 'vue'
import { mount } from '@vue/test-utils'

const event = ref<unknown>(null)
const isLoading = ref(true)
const operations = ref<unknown[] | null>(null)
const closeEvent = vi.fn()
const closeOperations = vi.fn()
const getEvent = vi.fn()
const getHttpOperations = vi.fn()

vi.mock('@/composables/dashboard', () => ({
    useDashboard: () => ({ dashboard: ref({ getEvent, getHttpOperations }) }),
}))

import { useHttpEvent } from './http-event'

const recorded = {
    id: 'e1',
    traits: { name: 'Books API', path: '/books', method: 'GET' },
    data: { response: { headers: { 'Content-Type': 'application/json' } } },
}

function mountUseHttpEvent(id: Ref<string>) {
    let result!: ReturnType<typeof useHttpEvent>
    const wrapper = mount(defineComponent({
        setup() {
            result = useHttpEvent(() => id.value)
            return () => null
        },
    }))
    return { result, wrapper }
}

describe('useHttpEvent', () => {
    beforeEach(() => {
        event.value = null
        isLoading.value = true
        operations.value = null
        vi.clearAllMocks()
        getEvent.mockImplementation(() => ({ event, isLoading, close: closeEvent }))
        getHttpOperations.mockImplementation(() => ({ operations, close: closeOperations }))
    })

    it('loads the request by id and is loading until it arrives', () => {
        const { result } = mountUseHttpEvent(ref('e1'))

        expect(getEvent).toHaveBeenCalledWith('e1')
        expect(result.isLoading.value).toBe(true)
        expect(result.event.value).toBeNull()
        expect(result.eventData.value).toBeUndefined()
    })

    it('exposes the data and the content type of the response', async () => {
        const { result } = mountUseHttpEvent(ref('e1'))

        event.value = recorded
        isLoading.value = false
        await nextTick()

        expect(result.eventData.value).toEqual(recorded.data)
        expect(result.responseContentType.value).toBe('application/json')
        expect(result.isLoading.value).toBe(false)
    })

    it('has an empty content type without a request or without the header', async () => {
        const { result } = mountUseHttpEvent(ref('e1'))
        expect(result.responseContentType.value).toBe('')

        event.value = { ...recorded, data: { response: {} } }
        await nextTick()

        expect(result.responseContentType.value).toBe('')
    })

    it('looks up the operation that handled the request, with a lowercase method', async () => {
        const { result } = mountUseHttpEvent(ref('e1'))
        expect(getHttpOperations).not.toHaveBeenCalled()

        event.value = recorded
        await nextTick()
        operations.value = [{ operationId: 'listBooks' }]

        expect(getHttpOperations).toHaveBeenCalledWith('Books API', '/books', 'get')
        expect(result.operation.value).toEqual({ operationId: 'listBooks' })
    })

    it('has no operation for a request that is not known', () => {
        const { result } = mountUseHttpEvent(ref('e1'))

        expect(result.operation.value).toBeUndefined()
    })

    it('loads again when the id changes and closes the previous request', async () => {
        const id = ref('e1')
        mountUseHttpEvent(id)

        id.value = 'e2'
        await nextTick()

        expect(getEvent).toHaveBeenLastCalledWith('e2')
        expect(closeEvent).toHaveBeenCalledOnce()
    })

    it('closes the request and its operation when the view is removed', async () => {
        const { wrapper } = mountUseHttpEvent(ref('e1'))
        event.value = recorded
        await nextTick()

        wrapper.unmount()

        expect(closeEvent).toHaveBeenCalledOnce()
        expect(closeOperations).toHaveBeenCalledOnce()
    })
})
