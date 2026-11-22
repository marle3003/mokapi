import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick, ref } from 'vue'
import { mount } from '@vue/test-utils'

const services = ref<unknown[]>([])
const close = vi.fn()
const getServices = vi.fn()

vi.mock('@/composables/dashboard', () => ({
    useDashboard: () => ({ dashboard: ref({ getServices }) }),
}))

import { useServiceList } from './use-service-list'

function mountUseServiceList(needed: boolean) {
    let list!: ReturnType<typeof useServiceList>
    const wrapper = mount(defineComponent({
        setup() {
            list = useServiceList('kafka', needed)
            return () => null
        },
    }))
    return { list, wrapper }
}

describe('useServiceList', () => {
    beforeEach(() => {
        services.value = []
        vi.clearAllMocks()
        getServices.mockImplementation(() => ({ services, close }))
    })

    it('lists the services of the protocol', async () => {
        const { list } = mountUseServiceList(true)

        services.value = [{ name: 'Kafka World' }]
        await nextTick()

        expect(getServices).toHaveBeenCalledWith('kafka', false)
        expect(list.value).toEqual([{ name: 'Kafka World' }])
    })

    it('loads nothing when the list is not needed', () => {
        const { list } = mountUseServiceList(false)

        expect(getServices).not.toHaveBeenCalled()
        expect(list.value).toEqual([])
    })

    it('stops loading when the view is removed', () => {
        const { wrapper } = mountUseServiceList(true)

        wrapper.unmount()

        expect(close).toHaveBeenCalledOnce()
    })
})
