import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import {
  buildSearchPath,
  facetTitle,
  fullSearchLocation,
  isOpenResult,
  isSearchable,
  searchStatusClass,
  searchTotalLabel,
  useSearch,
} from './search'

describe('isSearchable', () => {
  it('needs at least 3 characters', () => {
    expect(isSearchable('')).toBe(false)
    expect(isSearchable('ab')).toBe(false)
    expect(isSearchable('abc')).toBe(true)
  })

  it('is false while a field name is typed without a value', () => {
    expect(isSearchable('statusCode:')).toBe(false)
    expect(isSearchable('statusCode:>=400')).toBe(true)
  })
})

describe('buildSearchPath', () => {
  it('encodes the text and sets the limit', () => {
    expect(buildSearchPath('petstore statusCode:>=400', {}, { index: 0, limit: 20 }))
      .toBe('/api/search/query?q=petstore%20statusCode%3A%3E%3D400&limit=20')
  })

  it('adds the page index only for a page after the first', () => {
    expect(buildSearchPath('abc', {}, { index: 2, limit: 20 })).toBe('/api/search/query?q=abc&limit=20&index=2')
  })

  it('adds facets with a value and skips empty ones', () => {
    expect(buildSearchPath('abc', { type: 'HTTP & more', other: '', unset: undefined }, { index: 0, limit: 10 }))
      .toBe('/api/search/query?q=abc&limit=10&type=HTTP%20%26%20more')
  })
})

describe('facetTitle', () => {
  it('names the type facet', () => {
    expect(facetTitle('type')).toBe('Type')
  })

  it('says so for a facet without title', () => {
    expect(facetTitle('color')).toBe("title for 'color' not defined")
  })
})

describe('searchTotalLabel', () => {
  it('names all results when everything is shown', () => {
    expect(searchTotalLabel(1, 1)).toBe('Showing 1 result')
    expect(searchTotalLabel(7, 7)).toBe('Showing 7 results')
    expect(searchTotalLabel(0, 0)).toBe('Showing 0 results')
  })

  it('shows how many of all results are loaded', () => {
    expect(searchTotalLabel(20, 57)).toBe('Showing 20 of 57 results')
  })
})

describe('fullSearchLocation', () => {
  it('opens the search view with the text and the facets with a value', () => {
    expect(fullSearchLocation('abc', { type: 'Kafka', other: '' }))
      .toEqual({ name: 'search', query: { q: 'abc', type: 'Kafka' } })
  })

  it('has no location without text', () => {
    expect(fullSearchLocation('', { type: 'Kafka' })).toBeUndefined()
  })
})

describe('searchStatusClass', () => {
  it('has no class below 400', () => {
    expect(searchStatusClass('200')).toBe('')
    expect(searchStatusClass('304')).toBe('')
    expect(searchStatusClass('399')).toBe('')
  })

  it('is client-error from 400 to 499', () => {
    expect(searchStatusClass('400')).toBe('client-error')
    expect(searchStatusClass('499')).toBe('client-error')
  })

  it('is error from 500', () => {
    expect(searchStatusClass('500')).toBe('error')
    expect(searchStatusClass('503')).toBe('error')
  })

  it('has no class without a number', () => {
    expect(searchStatusClass('')).toBe('')
    expect(searchStatusClass('n/a')).toBe('')
  })
})

describe('isOpenResult', () => {
  it('is true for the same path', () => {
    expect(isOpenResult('/dashboard/http/service/Petstore', '/dashboard/http/service/Petstore')).toBe(true)
  })

  it('ignores the refresh setting and the hash', () => {
    expect(isOpenResult('/dashboard/kafka/service/K/topic/t', '/dashboard/kafka/service/K/topic/t?refresh=5#messages')).toBe(true)
  })

  it('keeps other query parameters', () => {
    expect(isOpenResult('/dashboard/x?a=1', '/dashboard/x?a=2')).toBe(false)
  })

  it('is false for another path', () => {
    expect(isOpenResult('/dashboard/http/service/A', '/dashboard/http/service/B')).toBe(false)
  })
})

function page(titles: string[], total: number, facets: SearchResult['facets'] = { type: [{ value: 'HTTP', count: total }] }): SearchResult {
  return { results: titles.map(title => ({ type: 'HTTP', title, fragments: [], params: { type: 'http' } })), total, facets }
}

function json(body: unknown, status = 200) {
  return Promise.resolve(new Response(JSON.stringify(body), { status }))
}

describe('useSearch', () => {
  let fetchMock: ReturnType<typeof vi.fn>
  let search: ReturnType<typeof useSearch>

  const requested = () => fetchMock.mock.calls.map(c => String(c[0]))
  const lastRequest = () => requested()[requested().length - 1]

  beforeEach(async () => {
    vi.useFakeTimers()
    fetchMock = vi.fn(() => json(page(['a'], 1)))
    vi.stubGlobal('fetch', fetchMock)
    search = useSearch({ limit: 20 })
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  async function type(text: string) {
    search.queryText.value = text
    await vi.advanceTimersByTimeAsync(0)
  }

  describe('trigger', () => {
    it('searches 1 s after the last key press', async () => {
      await type('abc')
      await vi.advanceTimersByTimeAsync(999)
      expect(fetchMock).not.toHaveBeenCalled()

      await vi.advanceTimersByTimeAsync(1)

      expect(requested()).toEqual(['/api/search/query?q=abc&limit=20'])
    })

    it('restarts the timer with every key press', async () => {
      await type('abc')
      await vi.advanceTimersByTimeAsync(600)
      await type('abcd')
      await vi.advanceTimersByTimeAsync(600)
      expect(fetchMock).not.toHaveBeenCalled()

      await vi.advanceTimersByTimeAsync(400)

      expect(requested()).toEqual(['/api/search/query?q=abcd&limit=20'])
    })

    it('searches immediately on searchNow, e.g. Enter', async () => {
      search.queryText.value = 'abc'

      await search.searchNow()

      expect(requested()).toEqual(['/api/search/query?q=abc&limit=20'])
    })

    it('does not search for less than 3 characters', async () => {
      await type('ab')
      await vi.advanceTimersByTimeAsync(2000)

      expect(fetchMock).not.toHaveBeenCalled()
    })

    it('does not search while the text ends with a colon', async () => {
      await type('statusCode:')
      await vi.advanceTimersByTimeAsync(2000)

      expect(fetchMock).not.toHaveBeenCalled()
    })

    it('drops a pending search when the text is no longer searchable', async () => {
      await type('abc')
      await vi.advanceTimersByTimeAsync(500)
      await type('ab')
      await vi.advanceTimersByTimeAsync(2000)

      expect(fetchMock).not.toHaveBeenCalled()
    })

    it('does not search for the initial text', async () => {
      const initial = useSearch({ limit: 10, queryText: ref('petstore') })

      await vi.advanceTimersByTimeAsync(2000)

      expect(initial.result.value).toBeUndefined()
      expect(fetchMock).not.toHaveBeenCalled()
    })
  })

  describe('result', () => {
    it('keeps the results, total and facet options', async () => {
      fetchMock.mockReturnValue(json(page(['a', 'b'], 5)))
      search.queryText.value = 'abc'

      await search.searchNow()

      expect(search.result.value?.results.map(r => r.title)).toEqual(['a', 'b'])
      expect(search.result.value?.total).toBe(5)
      expect(Object.keys(search.result.value!.facets)).toEqual(['type'])
    })

    it('has an empty result for no match', async () => {
      fetchMock.mockReturnValue(json({ results: [], total: 0 }))
      search.queryText.value = 'abc'

      await search.searchNow()

      expect(search.result.value?.total).toBe(0)
      expect(search.errorMessage.value).toBeUndefined()
    })

    it('shows the message of the API on an error', async () => {
      fetchMock.mockReturnValue(json({ message: 'invalid query' }, 400))
      search.queryText.value = 'abc'

      await search.searchNow()

      expect(search.errorMessage.value).toBe('invalid query')
      expect(search.result.value).toBeUndefined()
      expect(search.loading.isLoading.value).toBe(false)
    })

    it('shows the message of a network error', async () => {
      fetchMock.mockReturnValue(Promise.reject(new TypeError('Failed to fetch')))
      search.queryText.value = 'abc'

      await search.searchNow()

      expect(search.errorMessage.value).toBe('Failed to fetch')
    })

    it('clears the error with the next search', async () => {
      fetchMock.mockReturnValueOnce(json({ message: 'invalid query' }, 400))
      search.queryText.value = 'abc'
      await search.searchNow()

      await search.searchNow()

      expect(search.errorMessage.value).toBeUndefined()
      expect(search.result.value?.total).toBe(1)
    })

    it('is loading until the response arrived', async () => {
      let respond: (r: Response) => void = () => {}
      fetchMock.mockReturnValue(new Promise<Response>(resolve => { respond = resolve }))
      search.queryText.value = 'abc'

      const pending = search.searchNow()
      expect(search.loading.isLoading.value).toBe(true)

      respond(new Response(JSON.stringify(page(['a'], 1))))
      await pending

      expect(search.loading.isLoading.value).toBe(false)
    })

    it('ignores a response that arrives after a newer search', async () => {
      let respondFirst: (r: Response) => void = () => {}
      fetchMock.mockReturnValueOnce(new Promise<Response>(resolve => { respondFirst = resolve }))
      fetchMock.mockReturnValueOnce(json(page(['new'], 1)))
      search.queryText.value = 'abc'

      const first = search.searchNow()
      await search.searchNow()
      respondFirst(new Response(JSON.stringify(page(['old'], 1))))
      await first

      expect(search.result.value?.results.map(r => r.title)).toEqual(['new'])
    })
  })

  describe('facets', () => {
    it('adds a selectable entry per returned facet', async () => {
      search.queryText.value = 'abc'

      await search.searchNow()

      expect(search.facets.value).toEqual({ type: '' })
    })

    it('sends the selected facet and starts again at the first page', async () => {
      fetchMock.mockReturnValue(json(page(['a'], 40)))
      search.queryText.value = 'abc'
      await search.searchNow()
      await search.loadMore()

      search.facets.value.type = 'HTTP'
      await search.searchNow()

      expect(lastRequest()).toBe('/api/search/query?q=abc&limit=20&type=HTTP')
    })
  })

  describe('load more', () => {
    beforeEach(async () => {
      fetchMock.mockImplementation((url: string) =>
        json(url.includes('index=1') ? page(['c', 'd'], 3) : page(['a', 'b'], 3)))
      search.queryText.value = 'abc'
      await search.searchNow()
    })

    it('has more while not all results are loaded', () => {
      expect(search.hasMore.value).toBe(true)
    })

    it('requests the next page and appends it', async () => {
      await search.loadMore()

      expect(lastRequest()).toBe('/api/search/query?q=abc&limit=20&index=1')
      expect(search.result.value?.results.map(r => r.title)).toEqual(['a', 'b', 'c', 'd'])
      expect(search.result.value?.total).toBe(3)
    })

    it('has no more once all results are loaded', async () => {
      await search.loadMore()

      expect(search.hasMore.value).toBe(false)
    })

    it('replaces the loaded results with the first page of a new text', async () => {
      await search.loadMore()

      search.queryText.value = 'abcd'
      await search.searchNow()

      expect(lastRequest()).toBe('/api/search/query?q=abcd&limit=20')
      expect(search.result.value?.results.map(r => r.title)).toEqual(['a', 'b'])
    })

    it('keeps the loaded results when the next page fails', async () => {
      fetchMock.mockReturnValue(json({ message: 'boom' }, 500))

      await search.loadMore()

      expect(search.result.value?.results.map(r => r.title)).toEqual(['a', 'b'])
      expect(search.errorMessage.value).toBe('boom')
    })
  })

  describe('showPage', () => {
    it('replaces the results with the requested page', async () => {
      fetchMock.mockReturnValue(json(page(['x'], 30)))
      search.queryText.value = 'abc'

      await search.showPage(2)

      expect(requested()).toEqual(['/api/search/query?q=abc&limit=20&index=2'])
      expect(search.pageIndex.value).toBe(2)
    })
  })
})
