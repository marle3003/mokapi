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
