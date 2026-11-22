import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'

const metrics = ref<{ name: string; value: number }[]>([])
const close = vi.fn()

vi.mock('@/composables/dashboard', () => ({
    useDashboard: () => ({
        dashboard: ref({ getMetrics: () => ({ data: metrics.value, close }) }),
    }),
}))

import JobCountCard from './JobCountCard.vue'

const RouterLinkStub = {
    props: ['to'],
    template: '<a :href="to.path"><slot /></a>',
}

function mountCard(props: Record<string, unknown> = {}) {
    return mount(JobCountCard, { props, global: { stubs: { RouterLink: RouterLinkStub } } })
}

describe('JobCountCard', () => {
    it('is hidden when no job ran', () => {
        metrics.value = [{ name: 'app_job_run_total', value: 0 }]

        expect(mountCard().find('.card').exists()).toBe(false)
    })

    it('shows the number of job runs', () => {
        metrics.value = [
            { name: 'app_job_run_total{name="a"}', value: 2 },
            { name: 'app_job_run_total{name="b"}', value: 3 },
            { name: 'app_memory_usage_bytes', value: 100 },
        ]
        const wrapper = mountCard()

        expect(wrapper.find('.card-title').text()).toBe('Jobs')
        expect(wrapper.find('.card-text').text()).toBe('5')
        expect(wrapper.attributes('data-testid')).toBe('metric-job-count')
        expect(wrapper.element.tagName).toBe('DIV')
    })

    it('is a link when a route is given', () => {
        metrics.value = [{ name: 'app_job_run_total', value: 1 }]
        const wrapper = mountCard({ to: { path: '/dashboard/jobs' } })

        expect(wrapper.element.tagName).toBe('A')
        expect(wrapper.attributes('href')).toBe('/dashboard/jobs')
    })

    it('closes the metrics subscription when unmounted', () => {
        metrics.value = []
        close.mockClear()
        mountCard().unmount()

        expect(close).toHaveBeenCalled()
    })
})
