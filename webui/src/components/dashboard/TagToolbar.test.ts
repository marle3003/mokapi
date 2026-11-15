import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import TagToolbar from './TagToolbar.vue'

const pet = { name: 'pet', description: 'Everything about your pets' }
const store = { name: 'store' }

function mountToolbar(tags: { name: string, description?: string }[], props: Record<string, unknown> = {}) {
    return mount(TagToolbar, {
        props: { tags, noun: 'topics', selected: ['__all'], grouping: 'az', ...props },
    })
}

describe('TagToolbar visibility', () => {
    it('renders nothing without tags', () => {
        const toolbar = mountToolbar([])

        expect(toolbar.find('fieldset').exists()).toBe(false)
        expect(toolbar.find('[role="group"]').exists()).toBe(false)
    })

    it('shows only the switch with one tag', () => {
        const toolbar = mountToolbar([pet])

        expect(toolbar.find('fieldset').exists()).toBe(false)
        expect(toolbar.find('[aria-label="Group rows"]').exists()).toBe(true)
    })

    it('shows filter and switch with two tags', () => {
        const toolbar = mountToolbar([pet, store])

        expect(toolbar.find('fieldset').exists()).toBe(true)
        expect(toolbar.find('[aria-label="Group rows"]').exists()).toBe(true)
    })
})

describe('TagToolbar filter', () => {
    it('names the list in legend and help text and uses straight quotes', () => {
        const toolbar = mountToolbar([pet, store], { noun: 'paths' })

        expect(toolbar.get('legend').text()).toBe('Filter paths by tags')
        expect(toolbar.get('#tag-help').text()).toBe('Select one or more tags to filter the paths. Selecting "All" enables all tags.')
    })

    it('points aria-describedby of the fieldset to the help text', () => {
        const toolbar = mountToolbar([pet, store])

        expect(toolbar.get('fieldset').attributes('aria-describedby')).toBe('tag-help')
        expect(toolbar.find('#tag-help').exists()).toBe(true)
    })

    it('shows the description of a tag as title and hidden text', () => {
        const toolbar = mountToolbar([pet, store])

        const withDescription = toolbar.get('#tag-list .form-check:nth-child(1)')
        expect(withDescription.attributes('title')).toBe('Everything about your pets')
        expect(withDescription.get('#desc-pet').text()).toBe('Everything about your pets')
        expect(toolbar.get('input[id="pet"]').attributes('aria-describedby')).toBe('desc-pet')

        expect(toolbar.find('#desc-store').exists()).toBe(false)
        expect(toolbar.get('input[id="store"]').attributes('aria-describedby')).toBeUndefined()
    })

    it('checks every box while all tags are selected', () => {
        const toolbar = mountToolbar([pet, store])

        for (const id of ['all', 'pet', 'store']) {
            expect((toolbar.get(`input[id="${id}"]`).element as HTMLInputElement).checked).toBe(true)
        }
    })

    it('checks only the selected tags', () => {
        const toolbar = mountToolbar([pet, store], { selected: ['pet'] })

        expect((toolbar.get('input[id="all"]').element as HTMLInputElement).checked).toBe(false)
        expect((toolbar.get('input[id="pet"]').element as HTMLInputElement).checked).toBe(true)
        expect((toolbar.get('input[id="store"]').element as HTMLInputElement).checked).toBe(false)
    })

    it('emits the new selection when a tag is unchecked', async () => {
        const toolbar = mountToolbar([pet, store])

        await toolbar.get('input[id="store"]').trigger('change')

        expect(toolbar.emitted('update:selected')).toEqual([[['pet']]])
    })

    it('emits "All" when the All box is checked', async () => {
        const toolbar = mountToolbar([pet, store], { selected: ['pet'] })

        await toolbar.get('input[id="all"]').trigger('change')

        expect(toolbar.emitted('update:selected')).toEqual([[['__all']]])
    })
})

describe('TagToolbar switch', () => {
    it('follows the grouping with aria-pressed', () => {
        const toolbar = mountToolbar([pet], { grouping: 'tag' })

        const [az, byTag] = toolbar.findAll('button')
        expect(az.attributes('aria-pressed')).toBe('false')
        expect(byTag.attributes('aria-pressed')).toBe('true')
    })

    it('emits the chosen grouping', async () => {
        const toolbar = mountToolbar([pet])

        await toolbar.findAll('button')[1].trigger('click')

        expect(toolbar.emitted('update:grouping')).toEqual([['tag']])
    })
})
