import { describe, expect, it } from 'vitest'
import { useDocNav } from './doc-nav'

const overview: DocEntry = { label: 'Overview', path: '/docs/http/overview', source: 'http/overview.md' }
const quickStart: DocEntry = { label: 'Quick Start', path: '/docs/http/quick-start', source: 'http/quick-start.md' }
const http: DocEntry = { label: 'HTTP', path: '/docs/http/overview', source: 'http/overview.md', items: [quickStart] }
const kafka: DocEntry = { label: 'Kafka', items: [{ label: 'Config', path: '/docs/kafka/config' }] }
const protocol: DocEntry = { label: 'Protocol', type: 'headline', items: [http, kafka] }
const getStarted: DocEntry = {
  label: 'Get Started', type: 'headline', path: '/docs/welcome', source: 'get-started/welcome.md',
  items: [{ label: 'Installation', path: '/docs/get-started/installation' }]
}

function navFor(actives: DocEntry[] | undefined) {
  return useDocNav(() => actives)
}

describe('doc-nav', () => {
  describe('kind', () => {
    it.each([
      ['a headline with a page', getStarted, 'headline-link'],
      ['a headline without a page', protocol, 'headline-label'],
      ['a group with a page', http, 'group-link'],
      ['a group without a page', kafka, 'group-button'],
      ['an entry without items', overview, 'leaf'],
      ['a hidden entry without items', { ...overview, hideInNavigation: true }, 'hidden'],
    ] as [string, DocEntry, string][])('is %s', (_, entry, kind) => {
      expect(navFor(undefined).kind(entry)).toBe(kind)
    })

    it('treats a group with an empty item list as a leaf', () => {
      expect(navFor(undefined).kind({ label: 'Empty', path: '/docs/empty', items: [] })).toBe('leaf')
    })
  })

  describe('isActive', () => {
    it('is false without a breadcrumb', () => {
      expect(navFor(undefined).isActive(http)).toBe(false)
    })

    it('is true for every entry of the breadcrumb', () => {
      const nav = navFor([protocol, http, quickStart])

      expect(nav.isActive(protocol)).toBe(true)
      expect(nav.isActive(http)).toBe(true)
      expect(nav.isActive(quickStart)).toBe(true)
      expect(nav.isActive(kafka)).toBe(false)
    })
  })

  describe('isCurrentPage', () => {
    it('is true only for the last entry of the breadcrumb', () => {
      const nav = navFor([protocol, http, quickStart])

      expect(nav.isCurrentPage(quickStart)).toBe(true)
      expect(nav.isCurrentPage(http)).toBe(false)
    })

    it('is true for a group opened on its own page', () => {
      expect(navFor([protocol, http]).isCurrentPage(http)).toBe(true)
    })

    it('is false without a breadcrumb', () => {
      expect(navFor(undefined).isCurrentPage(http)).toBe(false)
      expect(navFor([]).isCurrentPage(http)).toBe(false)
    })
  })

  describe('isExpanded', () => {
    it('is true for a headline', () => {
      expect(navFor(undefined).isExpanded(protocol)).toBe(true)
    })

    it('is false for a collapsed group', () => {
      expect(navFor([getStarted]).isExpanded(http)).toBe(false)
    })

    it('is true for a group that contains the current page', () => {
      expect(navFor([protocol, http, quickStart]).isExpanded(http)).toBe(true)
    })

    it('is true for a group opened on its own page', () => {
      expect(navFor([protocol, http]).isExpanded(http)).toBe(true)
    })

    it('is true for a group marked as expanded', () => {
      expect(navFor(undefined).isExpanded({ ...kafka, expanded: true })).toBe(true)
    })

    it('keeps several groups open at the same time', () => {
      const nav = navFor([protocol, http, quickStart])

      expect(nav.isExpanded(http)).toBe(true)
      expect(nav.isExpanded({ ...kafka, expanded: true })).toBe(true)
    })
  })

  describe('getId', () => {
    it('is built from the path', () => {
      expect(navFor(undefined).getId(http)).toBe('nav-docs-http-overview')
    })

    it('is built from the label without a path', () => {
      expect(navFor(undefined).getId({ label: 'Use Cases', items: [] })).toBe('nav-Use-Cases')
    })
  })
})
