import { flushPromises, type mount } from '@vue/test-utils'
import { expect } from 'vitest'

// Queries for the DOM of a mounted component, named like the Playwright locators they stand in for.
// They take `wrapper.element` (or a shadow root) and return what a user reads, trimmed.

export function text(element: Element | ShadowRoot | null | undefined): string | undefined {
  return element?.textContent?.replace(/\s+/g, ' ').trim()
}

// `getByLabel(label, { exact: true })`: an element named by an aria-label or by the element its
// aria-labelledby points to (looked up inside `parent`).
export function labelled(parent: ParentNode, label: string): Element | undefined {
  return [...parent.querySelectorAll('[aria-label], [aria-labelledby]')].find(el => {
    const labelledBy = el.getAttribute('aria-labelledby')
    return (labelledBy ? text(parent.querySelector(`#${labelledBy}`)) : el.getAttribute('aria-label')) === label
  })
}

// `getByRole('region', { name })`: a section named by an aria-label or a (hidden) heading.
export function region(parent: ParentNode, name: string): Element {
  const found = labelled(parent, name)
  if (!found) {
    throw new Error(`no region "${name}"`)
  }
  return found
}

// `getByText(label, { exact: true })`: an element whose whole text is `label`, for absence checks.
export function hasElementWithText(parent: Element, label: string): boolean {
  return [...parent.querySelectorAll('*')].some(el => text(el) === label)
}

// The value below a `p.label`, like useRequestView().field in the Playwright tests.
export function field(parent: ParentNode, label: string): Element | undefined {
  return [...parent.querySelectorAll('p.label')].find(l => text(l) === label)?.nextElementSibling ?? undefined
}

// MailBody renders an HTML mail into a shadow root, which the queries above don't reach.
export function shadowRootOf(parent: ParentNode): ShadowRoot {
  const host = [...parent.querySelectorAll('*')].find(el => el.shadowRoot)
  if (!host?.shadowRoot) {
    throw new Error('no element with a shadow root')
  }
  return host.shadowRoot
}

// The sub tabs of a detail panel (a mounted wrapper, not a DOM element: `openTab` clicks).
type Panel = ReturnType<typeof mount>

function tabs(panel: Panel) {
  return panel.findAll('[role="tab"]')
}

export function tabNames(panel: Panel) {
  return tabs(panel).map(t => t.text())
}

export function selectedTab(panel: Panel) {
  return tabs(panel).filter(t => t.attributes('aria-selected') === 'true').map(t => t.text())
}

export async function openTab(panel: Panel, name: string) {
  await tabs(panel).find(t => t.text() === name)!.trigger('click')
  await flushPromises()
}

// the content of the selected sub tab
export function content(panel: Panel) {
  return panel.get('[role="tabpanel"]').element
}

// The cell texts of each body row, instead of Playwright's expectRows.
export function tableRows(table: Element): string[][] {
  return [...table.querySelectorAll('tbody tr')].map(row => [...row.querySelectorAll('td')].map(td => text(td) ?? ''))
}

// `toHaveCount(1)` of Playwright: a missing or duplicated element fails the test. `what` names it in the failure.
export function single(elements: Element[], what: string): Element {
  expect(elements, what).toHaveLength(1)
  return elements[0]
}
