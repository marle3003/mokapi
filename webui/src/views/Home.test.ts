import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import Home from './Home.vue'

const jsonLdSelector = 'script[type="application/ld+json"]'

function mountHome() {
  return mount(Home, { global: { stubs: { RouterLink: true, Footer: true, ImageDialog: true } } })
}

function jsonLdVersion() {
  const tag = document.head.querySelector(jsonLdSelector)
  return tag && JSON.parse(tag.textContent!).softwareVersion
}

function githubResponse(body: object) {
  return { json: () => Promise.resolve(body) }
}

describe('Home', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    document.head.innerHTML = ''
  })

  it('removes the JSON-LD when the page is left', async () => {
    vi.stubGlobal('fetch', vi.fn((url: string) =>
      Promise.resolve(githubResponse(url.endsWith('/releases/latest') ? { tag_name: 'v1.2.3' } : {}))))

    const home = mountHome()
    await flushPromises()
    expect(jsonLdVersion()).toBe('1.2.3')

    home.unmount()

    expect(document.head.querySelector(jsonLdSelector)).toBeNull()
  })

  it('does not insert the JSON-LD again when the GitHub response arrives after leaving', async () => {
    let resolveRelease!: (response: ReturnType<typeof githubResponse>) => void
    const release = new Promise<ReturnType<typeof githubResponse>>((resolve) => { resolveRelease = resolve })
    vi.stubGlobal('fetch', vi.fn((url: string) =>
      url.endsWith('/releases/latest') ? release : Promise.resolve(githubResponse({}))))

    const home = mountHome()
    home.unmount()
    resolveRelease(githubResponse({ tag_name: 'v9.9.9' }))
    await flushPromises()

    expect(document.head.querySelector(jsonLdSelector)).toBeNull()
  })
})
