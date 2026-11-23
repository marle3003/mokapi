import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import NavDocItem from './NavDocItem.vue'
import { region } from '@/test-support/dom-queries'

const RouterLinkStub = {
  props: ['to'],
  template: '<a :href="to.path"><slot /></a>'
}

const quickStart: DocEntry = { label: 'Quick Start', path: '/docs/http/quick-start' }
const http: DocEntry = { label: 'HTTP', path: '/docs/http/overview', items: [quickStart] }
const kafka: DocEntry = { label: 'Kafka', items: [{ label: 'Config', path: '/docs/kafka/config' }] }
const protocol: DocEntry = { label: 'Protocol', type: 'headline', items: [http, kafka] }
const getStarted: DocEntry = {
  label: 'Get Started', type: 'headline', path: '/docs/welcome',
  items: [{ label: 'Installation', path: '/docs/get-started/installation' }]
}

function render(current: DocEntry, actives?: DocEntry[]) {
  return mount(NavDocItem, {
    props: { current, actives },
    global: { stubs: { RouterLink: RouterLinkStub } }
  })
}

describe('NavDocItem', () => {
  describe('headline', () => {
    it('is a link to its page when it has a path', () => {
      const wrapper = render(getStarted)

      const link = wrapper.get('a[href="/docs/welcome"]')
      expect(link.text()).toBe('Get Started')
      expect(wrapper.find('button').exists()).toBe(false)
    })

    it('is a plain label without a path', () => {
      const wrapper = render(protocol)

      expect(wrapper.find('.headline > a, .headline > button').exists()).toBe(false)
      expect(wrapper.get('.headline > div').text()).toBe('Protocol')
    })

    it('labels its always expanded section', () => {
      const wrapper = render(protocol)

      const section = region(wrapper.element, 'Protocol')
      expect(section.classList.contains('show')).toBe(true)
    })

    it('is active only on its own page', () => {
      expect(render(getStarted, [getStarted]).get('a[href="/docs/welcome"]').classes()).toContain('active')
      expect(render(getStarted, [getStarted, getStarted.items![0]!]).get('a[href="/docs/welcome"]').classes())
        .not.toContain('active')
    })
  })

  describe('group with a page', () => {
    it('links its name to the page and toggles with the chevron', () => {
      const wrapper = render(http)

      const link = wrapper.get('a[href="/docs/http/overview"]')
      expect(link.text()).toBe('HTTP')
      expect(link.attributes('data-bs-toggle')).toBeUndefined()
      const chevron = wrapper.get('button')
      expect(chevron.attributes('aria-label')).toBe('Toggle HTTP')
      expect(chevron.attributes('data-bs-toggle')).toBe('collapse')
      expect(chevron.attributes('aria-controls')).toBe(wrapper.get('section').attributes('id'))
    })

    it('labels its region by the link', () => {
      const wrapper = render(http)

      const link = wrapper.get('a[href="/docs/http/overview"]')
      expect(wrapper.get('section').attributes('aria-labelledby')).toBe(link.attributes('id'))
      expect(region(wrapper.element, 'HTTP')).toBe(wrapper.get('section').element)
    })

    it('is collapsed by default', () => {
      const wrapper = render(http)

      expect(wrapper.get('section').classes()).not.toContain('show')
      expect(wrapper.get('button').attributes('aria-expanded')).toBe('false')
    })

    it('is active and expanded on its own page', () => {
      const wrapper = render(http, [protocol, http])

      expect(wrapper.get('a[href="/docs/http/overview"]').classes()).toContain('active')
      expect(wrapper.get('section').classes()).toContain('show')
      expect(wrapper.get('button').attributes('aria-expanded')).toBe('true')
    })

    it('is expanded but not active on a child page', () => {
      const wrapper = render(http, [protocol, http, quickStart])

      expect(wrapper.get('a[href="/docs/http/overview"]').classes()).not.toContain('active')
      expect(wrapper.get('section').classes()).toContain('show')
      expect(wrapper.get('a[href="/docs/http/quick-start"]').classes()).toContain('active')
    })
  })

  describe('group without a page', () => {
    it('toggles through its name and the chevron', () => {
      const wrapper = render(kafka)

      const buttons = wrapper.findAll('button')
      expect(buttons.map(b => b.text() || b.attributes('aria-label'))).toEqual(['Kafka', 'Toggle Kafka'])
      expect(buttons.every(b => b.attributes('data-bs-toggle') === 'collapse')).toBe(true)
    })

    it('is expanded when marked as expanded', () => {
      const wrapper = render({ ...kafka, expanded: true })

      expect(wrapper.get('section').classes()).toContain('show')
    })
  })

  describe('entry without items', () => {
    it('is a link', () => {
      const wrapper = render(quickStart, [quickStart])

      const link = wrapper.get('a')
      expect(link.attributes('href')).toBe('/docs/http/quick-start')
      expect(link.classes()).toContain('active')
    })

    it('is not shown when hidden in the navigation', () => {
      const wrapper = render({ ...quickStart, hideInNavigation: true })

      expect(wrapper.find('a').exists()).toBe(false)
    })
  })
})
