import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick, reactive, ref } from 'vue'
import { mount } from '@vue/test-utils'

const data = ref<unknown>(null)
const isLoading = ref(true)
const close = vi.fn()
const getEvents = vi.fn()
let mode = 'live'
const route = reactive<{ query: Record<string, unknown> }>({ query: {} })

vi.mock('vue-router', () => ({ useRoute: () => route }))

vi.mock('@/composables/dashboard', () => ({
    useDashboard: () => ({ dashboard: ref({ getEvents }), getMode: () => mode }),
}))

import { useEventLog } from './use-event-log'

function mountUseEventLog(labels = ref<Label[]>([{ name: 'namespace', value: 'kafka' }])) {
    let result!: ReturnType<typeof useEventLog>
    const wrapper = mount(defineComponent({
        setup() {
            result = useEventLog(() => ({ labels: labels.value }))
            return () => null
        },
    }))
    return { result, labels, wrapper }
}

function mountUseEventLogWith(query: () => { labels: Label[], http?: object }) {
    mount(defineComponent({
        setup() {
            useEventLog(query)
            return () => null
        },
    }))
}

describe('useEventLog', () => {
    beforeEach(() => {
        localStorage.clear()
        data.value = null
        isLoading.value = true
        mode = 'live'
        route.query = { refresh: '20' }
        vi.clearAllMocks()
        getEvents.mockImplementation(() => ({ data, isLoading, close }))
    })

    it('asks for the first page of the query and exposes the items', async () => {
        const { result } = mountUseEventLog()

        expect(getEvents).toHaveBeenCalledWith({ page: 1, size: 25, labels: [{ name: 'namespace', value: 'kafka' }] })
        expect(result.isLoading.value).toBe(true)

        data.value = { items: [{ id: 'a' }], total: 1 }
        isLoading.value = false
        await nextTick()

        expect(result.events.value).toEqual([{ id: 'a' }])
        expect(result.isLoading.value).toBe(false)
    })

    it('asks again and closes the previous request when the page changes', async () => {
        const { result } = mountUseEventLog()

        result.paging.current = 2
        await nextTick()

        expect(getEvents).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2 }))
        expect(close).toHaveBeenCalledOnce()
    })

    it('stops asking while paused, keeps the shown events and asks again on resume', async () => {
        const { result } = mountUseEventLog()
        data.value = { items: [{ id: 'a' }], total: 1 }
        await nextTick()

        result.togglePause()
        await nextTick()

        expect(close).toHaveBeenCalledOnce()
        expect(result.events.value).toEqual([{ id: 'a' }])
        const calls = getEvents.mock.calls.length

        result.togglePause()
        await nextTick()

        expect(getEvents.mock.calls.length).toBe(calls + 1)
    })

    it('ends the pause when the query changes', async () => {
        const { result, labels } = mountUseEventLog()
        result.togglePause()
        await nextTick()

        labels.value = [{ name: 'namespace', value: 'http' }]
        await nextTick()

        expect(result.paused.value).toBe(false)
        expect(getEvents).toHaveBeenLastCalledWith(expect.objectContaining({ labels: [{ name: 'namespace', value: 'http' }] }))
    })

    it('stays paused when the query is rebuilt with the same content', async () => {
        // e.g. the service of the log is a new object after every poll of the parent
        const rebuilt = ref(0)
        let result!: ReturnType<typeof useEventLog>
        mount(defineComponent({
            setup() {
                result = useEventLog(() => {
                    void rebuilt.value
                    return { labels: [{ name: 'namespace', value: 'kafka' }] }
                })
                return () => null
            },
        }))
        result.togglePause()
        await nextTick()
        const calls = getEvents.mock.calls.length

        rebuilt.value++
        await nextTick()

        expect(result.paused.value).toBe(true)
        expect(getEvents.mock.calls.length).toBe(calls)
    })

    it('asks again when a value inside the query was edited in place', async () => {
        // the query shares the (reactive) rows of the filter editor, which are edited while typing
        const row = reactive({ name: 'Set-Cookie', value: 'a' })
        mountUseEventLogWith(() => ({ labels: [], http: { responseHeaders: [row] } }))

        row.value = 'b'
        await nextTick()

        expect(getEvents).toHaveBeenLastCalledWith(expect.objectContaining({ http: { responseHeaders: [{ name: 'Set-Cookie', value: 'b' }] } }))
        expect(getEvents).toHaveBeenCalledTimes(2)
    })

    describe('canPause', () => {
        it.each([
            ['20', true],
            ['0.5', true],
            [undefined, false],
            ['', false],
            ['0', false],
            ['abc', false],
            ['-5', false],
            ['Infinity', false],
            [['5', '10'], false],
        ])('in the live dashboard with refresh=%j is %s', (refresh, expected) => {
            route.query = refresh === undefined ? {} : { refresh }

            expect(mountUseEventLog().result.canPause.value).toBe(expected)
        })

        it('is false in the demo dashboard, even with refresh', () => {
            mode = 'demo'

            expect(mountUseEventLog().result.canPause.value).toBe(false)
        })

        it('follows the refresh of the current route', async () => {
            route.query = {}
            const { result } = mountUseEventLog()
            expect(result.canPause.value).toBe(false)

            route.query = { refresh: '20' }
            await nextTick()
            expect(result.canPause.value).toBe(true)

            route.query = {}
            await nextTick()
            expect(result.canPause.value).toBe(false)
        })
    })

    it('closes the request when the log is removed', () => {
        const { wrapper } = mountUseEventLog()

        wrapper.unmount()

        expect(close).toHaveBeenCalledOnce()
    })
})
