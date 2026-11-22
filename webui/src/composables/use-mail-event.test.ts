import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick, ref, type Ref } from 'vue'
import { mount } from '@vue/test-utils'

const event = ref<unknown>(null)
const eventLoading = ref(true)
const mail = ref<unknown>(null)
const mailLoading = ref(true)
const getEvent = vi.fn()
const getMail = vi.fn()

vi.mock('@/composables/dashboard', () => ({
    useDashboard: () => ({ dashboard: ref({ getEvent, getMail }) }),
}))

import { useMailEvent } from './use-mail-event'

function recorded(data: object = {}) {
    return {
        id: 'e1',
        traits: { namespace: 'mail', name: 'Mail Testserver' },
        data: { from: 'alice@mokapi.io', to: ['bob@mokapi.io'], messageId: 'm1', subject: 'Hi', duration: 3, error: '', actions: null, ...data },
    }
}

function mountUseMailEvent(id: Ref<string>) {
    let result!: ReturnType<typeof useMailEvent>
    mount(defineComponent({
        setup() {
            result = useMailEvent(() => id.value)
            return () => null
        },
    }))
    return result
}

async function eventArrives(e: object) {
    event.value = e
    eventLoading.value = false
    await nextTick()
}

describe('useMailEvent', () => {
    beforeEach(() => {
        event.value = null
        eventLoading.value = true
        mail.value = null
        mailLoading.value = true
        vi.clearAllMocks()
        getEvent.mockImplementation(() => ({ event, isLoading: eventLoading, close: vi.fn() }))
        getMail.mockImplementation(() => ({ mail, isLoading: mailLoading }))
    })

    it('is loading and fetches no mail until the event arrives', () => {
        const result = mountUseMailEvent(ref('e1'))

        expect(getEvent).toHaveBeenCalledWith('e1')
        expect(getMail).not.toHaveBeenCalled()
        expect(result.isLoading.value).toBe(true)
        expect(result.mailGone.value).toBe(false)
    })

    it('loads the mail an accepted event names', async () => {
        const result = mountUseMailEvent(ref('e1'))

        await eventArrives(recorded())
        mail.value = { service: 'Mail Testserver', data: { messageId: 'm1' } }
        mailLoading.value = false
        await nextTick()

        expect(getMail).toHaveBeenCalledWith('m1')
        expect(result.mail.value).toEqual(mail.value)
        expect(result.rejected.value).toBe(false)
        expect(result.mailGone.value).toBe(false)
        expect(result.isLoading.value).toBe(false)
    })

    it('is loading while the mail loads', async () => {
        const result = mountUseMailEvent(ref('e1'))

        await eventArrives(recorded())

        expect(result.isLoading.value).toBe(true)
        expect(result.mailGone.value).toBe(false)
    })

    it('knows that the mail is gone when it is not found', async () => {
        const result = mountUseMailEvent(ref('e1'))

        await eventArrives(recorded())
        mailLoading.value = false
        await nextTick()

        expect(result.mail.value).toBeNull()
        expect(result.mailGone.value).toBe(true)
        expect(result.data.value?.subject).toBe('Hi')
    })

    it('never fetches the mail of a rejected event', async () => {
        const result = mountUseMailEvent(ref('e1'))

        await eventArrives(recorded({ messageId: 'm1', error: 'Mail rejected by rule spam' }))

        expect(getMail).not.toHaveBeenCalled()
        expect(result.rejected.value).toBe(true)
        expect(result.mailGone.value).toBe(false)
        expect(result.isLoading.value).toBe(false)
    })

    it('has no mail and ends loading for an unknown event id', async () => {
        const result = mountUseMailEvent(ref('unknown'))

        eventLoading.value = false
        await nextTick()

        expect(result.event.value).toBeNull()
        expect(result.isLoading.value).toBe(false)
        expect(result.mailGone.value).toBe(false)
    })

    it('loads again when the id changes', async () => {
        const id = ref('e1')
        mountUseMailEvent(id)
        await eventArrives(recorded())

        id.value = 'e2'
        await nextTick()

        expect(getEvent).toHaveBeenLastCalledWith('e2')
    })

    it('fetches the mail of the next event', async () => {
        const result = mountUseMailEvent(ref('e1'))
        await eventArrives(recorded({ messageId: 'm1' }))
        expect(getMail).toHaveBeenLastCalledWith('m1')

        await eventArrives(recorded({ messageId: 'm2' }))

        expect(getMail).toHaveBeenLastCalledWith('m2')
        expect(result.data.value?.messageId).toBe('m2')
    })
})
