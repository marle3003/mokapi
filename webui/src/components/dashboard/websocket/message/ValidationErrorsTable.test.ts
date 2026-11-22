import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ValidationErrorsTable from './ValidationErrorsTable.vue'

const validationErrors = {
    ChatMessage: 'Validation error count 1:\n\t- #/type: invalid type, expected object but got string',
    Foo: 'Validation error count 2:\n\t- #/bar/type: invalid type\n\t- #/required: required properties are missing: foo',
}

const mounted: ReturnType<typeof mount>[] = []

function mountTable() {
    const wrapper = mount(ValidationErrorsTable, {
        props: { validationErrors },
        attachTo: document.body,
        global: { stubs: { SourceView: true, SchemaExpand: true, SchemaValidate: true } },
    })
    mounted.push(wrapper)
    return wrapper
}

// A click on the header of an error with selected text must not collapse or expand it.
function clickOnErrorWithSelectedText(wrapper: ReturnType<typeof mountTable>): MouseEvent {
    vi.spyOn(window, 'getSelection').mockReturnValue({ toString: () => 'selected' } as Selection)
    const event = new MouseEvent('click', { bubbles: true, cancelable: true })
    wrapper.get('[data-bs-toggle="collapse"]').element.dispatchEvent(event)
    return event
}

describe('ValidationErrorsTable', () => {
    afterEach(() => {
        // a table that stays mounted keeps its click listener on the window
        mounted.splice(0).forEach(w => w.unmount())
        vi.restoreAllMocks()
        document.body.innerHTML = ''
    })

    it('lists the errors with their summary', () => {
        const wrapper = mountTable()

        const rows = wrapper.findAll('[role="rowgroup"][class~="table-row"]')
        expect(rows.map(r => r.findAll('[role="cell"]').slice(0, 2).map(c => c.text()))).toEqual([
            ['ChatMessage', 'Validation error count 1'],
            ['Foo', 'Validation error count 2'],
        ])
        expect(wrapper.get('[role="table"]').attributes('aria-label')).toBe('Validation Errors')
    })

    it('lists the single errors of a message', () => {
        const wrapper = mountTable()

        expect(wrapper.get('ul[aria-label="Validation errors for Foo"]').findAll('li').map(li => li.text())).toEqual([
            ' #/bar/type: invalid type',
            ' #/required: required properties are missing: foo',
        ].map(t => t.trim()))
    })

    it('keeps an error open or closed while text is selected', () => {
        const wrapper = mountTable()

        expect(clickOnErrorWithSelectedText(wrapper).defaultPrevented).toBe(true)
    })

    it('toggles an error when no text is selected', () => {
        const wrapper = mountTable()
        vi.spyOn(window, 'getSelection').mockReturnValue({ toString: () => '' } as Selection)
        const event = new MouseEvent('click', { bubbles: true, cancelable: true })

        wrapper.get('[data-bs-toggle="collapse"]').element.dispatchEvent(event)

        expect(event.defaultPrevented).toBe(false)
    })

    it('stops watching clicks when it is removed, as the log panel mounts it for every message', () => {
        const wrapper = mountTable()
        const trigger = wrapper.get('[data-bs-toggle="collapse"]').element
        wrapper.unmount()
        document.body.appendChild(trigger)
        vi.spyOn(window, 'getSelection').mockReturnValue({ toString: () => 'selected' } as Selection)
        const event = new MouseEvent('click', { bubbles: true, cancelable: true })

        trigger.dispatchEvent(event)

        expect(event.defaultPrevented).toBe(false)
    })
})
