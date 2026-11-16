import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'

const services = ref<unknown[]>([])
const close = vi.fn()

vi.mock('@/composables/dashboard', () => ({
    useDashboard: () => ({
        dashboard: ref({ getServices: () => ({ services, close }) }),
    }),
}))

import HttpRequestMetricCard from './HttpRequestMetricCard.vue'

const RouterLinkStub = {
    props: ['to'],
    template: '<a :href="to.path"><slot /></a>',
}

function mountCard(props: Record<string, unknown> = {}) {
    services.value = [
        { metrics: { requests_total: 3, requests_errors_total: 1 } },
        { metrics: { requests_total: 1 } },
    ]
    return mount(HttpRequestMetricCard, { props, global: { stubs: { RouterLink: RouterLinkStub } } })
}

describe('HttpRequestMetricCard', () => {
    it('shows the requests of all services', () => {
        const wrapper = mountCard()

        expect(wrapper.find('.card-title').text()).toBe('HTTP Requests')
        expect(wrapper.find('.card-text').text()).toBe('4')
    })

    it('shows requests and errors', () => {
        const wrapper = mountCard({ includeError: true })

        expect(wrapper.find('.card-title').text()).toBe('HTTP Requests / Errors')
        expect(wrapper.find('.card-text').text()).toBe('4 / 1')
        expect(wrapper.find('.text-danger').text()).toBe('1')
    })

    it('shows only the errors', () => {
        const wrapper = mountCard({ onlyError: true })

        expect(wrapper.find('.card-title').text()).toBe('HTTP Errors')
        expect(wrapper.find('.card-text').text()).toBe('1')
    })

    it('does not mark zero errors as danger', () => {
        const wrapper = mountCard({ includeError: true })
        services.value = [{ metrics: { requests_total: 2 } }]

        return wrapper.vm.$nextTick().then(() => {
            expect(wrapper.find('.card-text').text()).toBe('2 / 0')
            expect(wrapper.find('.text-danger').exists()).toBe(false)
        })
    })

    it('shows 0 as long as the services are not loaded', () => {
        const wrapper = mountCard()
        services.value = null as unknown as unknown[]

        return wrapper.vm.$nextTick().then(() => {
            expect(wrapper.find('.card-text').text()).toBe('0')
        })
    })

    it('keeps the test id and is a labelled status', () => {
        const wrapper = mountCard()

        expect(wrapper.attributes('data-testid')).toBe('metric-http-requests')
        const value = wrapper.find('.card-text')
        expect(value.attributes('role')).toBe('status')
        expect(value.attributes('aria-labelledby')).toBe(wrapper.find('.card-title').attributes('id'))
    })

    it('is a plain card without a route', () => {
        expect(mountCard().element.tagName).toBe('DIV')
    })

    it('is a link when a route is given', () => {
        const wrapper = mountCard({ to: { path: '/dashboard/http' } })

        expect(wrapper.element.tagName).toBe('A')
        expect(wrapper.attributes('href')).toBe('/dashboard/http')
    })

    it('closes the services subscription when unmounted', () => {
        close.mockClear()
        mountCard().unmount()

        expect(close).toHaveBeenCalled()
    })
})
