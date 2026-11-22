import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

// The panel state and useLocalStorage keep state at module level, so every test gets fresh modules.
async function load() {
  vi.resetModules()
  const { useSearchPanel } = await import('./search-panel')
  return useSearchPanel()
}

function respond(body: unknown, ok = true) {
  const fetchMock = vi.fn().mockResolvedValue({ ok, json: () => Promise.resolve(body) })
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

const facets = (...values: string[]) => ({ type: values.map(value => ({ value, count: 1 })) })

describe('search panel state', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('starts empty', async () => {
    const panel = await load()

    expect(panel.queryText.value).toBe('')
    expect(panel.facets.value).toEqual({})
  })

  it('restores text and facets from localStorage', async () => {
    localStorage.setItem('dashboard-search-query', JSON.stringify('petstore'))
    localStorage.setItem('dashboard-search-facets', JSON.stringify({ type: 'HTTP' }))

    const panel = await load()

    expect(panel.queryText.value).toBe('petstore')
    expect(panel.facets.value).toEqual({ type: 'HTTP' })
  })

  it('starts empty when the saved values are corrupt', async () => {
    localStorage.setItem('dashboard-search-query', 'petstore')
    localStorage.setItem('dashboard-search-facets', '{type:')

    const panel = await load()

    expect(panel.queryText.value).toBe('')
    expect(panel.facets.value).toEqual({})
  })

  it('saves text and facets when they change', async () => {
    const panel = await load()

    panel.queryText.value = 'petstore'
    panel.facets.value = { type: 'Kafka' }
    await nextTick()

    expect(JSON.parse(localStorage.getItem('dashboard-search-query')!)).toBe('petstore')
    expect(JSON.parse(localStorage.getItem('dashboard-search-facets')!)).toEqual({ type: 'Kafka' })
  })

  it('searches with the restored facet and 20 results per page', async () => {
    localStorage.setItem('dashboard-search-query', JSON.stringify('petstore'))
    localStorage.setItem('dashboard-search-facets', JSON.stringify({ type: 'HTTP' }))
    const fetchMock = respond({ results: [], total: 0, facets: facets('HTTP') })
    const panel = await load()

    await panel.searchNow()

    expect(fetchMock.mock.calls[0][0]).toContain('q=petstore&limit=20&type=HTTP')
  })

  it('keeps a saved facet value that is among the options', async () => {
    localStorage.setItem('dashboard-search-query', JSON.stringify('petstore'))
    localStorage.setItem('dashboard-search-facets', JSON.stringify({ type: 'HTTP' }))
    const fetchMock = respond({ results: [], total: 0, facets: facets('HTTP', 'Kafka') })
    const panel = await load()

    await panel.searchNow()
    await nextTick()

    expect(panel.facets.value).toEqual({ type: 'HTTP' })
    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('drops a saved facet value that is not among the options and searches again without it', async () => {
    localStorage.setItem('dashboard-search-query', JSON.stringify('petstore'))
    localStorage.setItem('dashboard-search-facets', JSON.stringify({ type: 'Gone' }))
    const fetchMock = respond({ results: [], total: 0, facets: facets('HTTP', 'Kafka') })
    const panel = await load()

    await panel.searchNow()
    await vi.waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2))

    expect(panel.facets.value).toEqual({ type: '' })
    expect(fetchMock.mock.calls[1][0]).not.toContain('type=')
  })

  it('leaves a facet alone when the response has no options for it', async () => {
    localStorage.setItem('dashboard-search-query', JSON.stringify('petstore'))
    localStorage.setItem('dashboard-search-facets', JSON.stringify({ type: 'HTTP' }))
    respond({ results: [], total: 0 })
    const panel = await load()

    await panel.searchNow()
    await nextTick()

    expect(panel.facets.value).toEqual({ type: 'HTTP' })
  })
})
