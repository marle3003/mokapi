import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { reactive } from 'vue'
import { removeToolbarItem, useToolbarItem } from '@/composables/toolbar-items'

const route = reactive<{ query: Record<string, string> }>({ query: {} })

vi.mock('@/router', () => ({
    useRoute: () => route,
}))

import ServiceInfoCard from './ServiceInfoCard.vue'

const service = {
    name: 'Swagger Petstore',
    description: '',
    version: '1.0.7',
    contact: null,
    type: 'HTTP',
    metrics: {},
    configs: [],
} as Service

let itemIds: number[] = []

function addItems(onClick = vi.fn()) {
    itemIds = useToolbarItem(
        { label: 'OpenAPI specification', icon: 'bi-filetype-json', onClick },
        { label: 'Bruno collection', icon: 'bi-box-arrow-down', onClick },
    )
    return onClick
}

function mountCard() {
    return mount(ServiceInfoCard, { props: { service }, attachTo: document.body })
}

afterEach(() => {
    itemIds.forEach(removeToolbarItem)
    itemIds = []
    route.query = {}
    document.body.innerHTML = ''
})

describe('ServiceInfoCard export button', () => {
    it('shows only the type badge without items', () => {
        const wrapper = mountCard()

        expect(wrapper.find('button').exists()).toBe(false)
        expect(wrapper.find('.dropdown').exists()).toBe(false)
        expect(wrapper.find('[data-testid="service-type"]').text()).toBe('HTTP')
    })

    it('shows a labeled Export button with an icon', () => {
        addItems()
        const button = mountCard().find('button.dropdown-toggle')

        expect(button.text()).toBe('Export')
        expect(button.find('i.bi-box-arrow-up').attributes('aria-hidden')).toBe('true')
    })

    it('lists the item labels as menu entries, right-aligned', () => {
        addItems()
        const wrapper = mountCard()

        const entries = wrapper.findAll('.dropdown-menu .dropdown-item').map(e => e.text())
        expect(entries).toEqual(['OpenAPI specification', 'Bruno collection'])
        expect(wrapper.find('.dropdown-menu').classes()).toContain('dropdown-menu-end')
    })

    it('replaces the kebab button', () => {
        addItems()
        const wrapper = mountCard()

        expect(wrapper.find('[aria-label^="Actions for"]').exists()).toBe(false)
        expect(wrapper.find('.card-toolbar').exists()).toBe(false)
    })

    it('shows the type badge in front of the service name, not next to the button', () => {
        addItems()
        const wrapper = mountCard()

        const badge = wrapper.find('[data-testid="service-type"]').element
        const name = wrapper.find('[data-testid="service-name"]').element
        expect(badge.nextElementSibling).toBe(name)
        expect(wrapper.find('.dropdown').element.parentElement?.contains(badge)).toBe(false)
        expect(wrapper.find('[data-testid="service-name"]').text()).toBe('Swagger Petstore')
    })

    it('runs onClick once per click on an entry', async () => {
        const onClick = addItems()
        const wrapper = mountCard()

        await wrapper.findAll('.dropdown-item')[1].trigger('click')

        expect(onClick).toHaveBeenCalledTimes(1)
    })
})

describe('ServiceInfoCard demo=bruno', () => {
    it('opens the dropdown and highlights Bruno collection', async () => {
        addItems()
        route.query = { demo: 'bruno' }
        const wrapper = mountCard()
        await flushPromises()

        expect(wrapper.find('.dropdown-menu').classes()).toContain('show')
        const highlighted = wrapper.findAll('.menu-highlight').map(e => e.text())
        expect(highlighted).toEqual(['Bruno collection'])
    })

    it('highlights nothing for another demo value', async () => {
        addItems()
        route.query = { demo: 'other' }
        const wrapper = mountCard()
        await flushPromises()

        expect(wrapper.find('.dropdown-menu').classes()).not.toContain('show')
        expect(wrapper.find('.menu-highlight').exists()).toBe(false)
    })
})
