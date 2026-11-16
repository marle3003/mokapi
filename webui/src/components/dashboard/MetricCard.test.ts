import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import MetricCard from './MetricCard.vue'

const RouterLinkStub = {
    props: ['to'],
    template: '<a :href="to.path"><slot /></a>',
}

type Props = InstanceType<typeof MetricCard>['$props']

function mountCard(props: Partial<Props>, slots: Record<string, string> = {}) {
    return mount(MetricCard, { props: { title: 'Kafka Messages', ...props }, slots, global: { stubs: { RouterLink: RouterLinkStub } } })
}

describe('MetricCard', () => {
    it('is a plain card without a link when no route is given', () => {
        const wrapper = mountCard({ value: 12 })

        expect(wrapper.element.tagName).toBe('DIV')
        expect(wrapper.find('a').exists()).toBe(false)
        expect(wrapper.classes()).not.toContain('metric-tile')
        expect(wrapper.find('.card-title').text()).toBe('Kafka Messages')
        expect(wrapper.find('.card-text').text()).toBe('12')
    })

    it('is a link with title and value when a route is given', () => {
        const wrapper = mountCard({ value: 12, to: { path: '/dashboard/kafka' } })

        expect(wrapper.element.tagName).toBe('A')
        expect(wrapper.attributes('href')).toBe('/dashboard/kafka')
        expect(wrapper.classes()).toContain('metric-tile')
        expect(wrapper.find('.card-title').text()).toBe('Kafka Messages')
        expect(wrapper.find('.card-text').text()).toBe('12')
    })

    it('renders the value slot instead of the value', () => {
        const wrapper = mountCard({ value: 12 }, { value: '<span class="custom">4 / 1</span>' })

        expect(wrapper.find('.card-text .custom').text()).toBe('4 / 1')
        expect(wrapper.find('.card-text').text()).toBe('4 / 1')
    })

    it('uses the given aria-live on a plain card', () => {
        expect(mountCard({ value: 1, live: 'off' }).find('.card-text').attributes('aria-live')).toBe('off')
        expect(mountCard({ value: 1 }).find('.card-text').attributes('aria-live')).toBeUndefined()
    })

    it('does not announce updates of a linked card', () => {
        const wrapper = mountCard({ value: 1, live: 'polite', to: { path: '/dashboard/kafka' } })

        expect(wrapper.find('.card-text').attributes('aria-live')).toBe('off')
    })

    it('shows the additional text with its label', () => {
        const wrapper = mountCard({ value: 1, additional: '2022-05-08', additionalLabel: 'Started at' })

        expect(wrapper.find('.card-additional').attributes('aria-label')).toBe('Started at')
        expect(wrapper.find('.card-additional').text()).toBe('2022-05-08')
    })
})
