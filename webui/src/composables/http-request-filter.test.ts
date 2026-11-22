import { afterEach, describe, expect, it } from 'vitest'
import {
  activateFilter,
  addableFields,
  defaultFilter,
  fieldLabel,
  filterChips,
  filterStorageKey,
  loadFilter,
  removeFilter,
  restoreFilter,
  saveFilter,
  toEventQuery,
  type HttpRequestFilter,
} from './http-request-filter'

function filterWith(changes: (filter: HttpRequestFilter) => void): HttpRequestFilter {
  const filter = defaultFilter()
  changes(filter)
  return filter
}

const all = {}
const petstore = { serviceName: 'Petstore' }

describe('http request filter', () => {
  describe('storageKey', () => {
    it('is one key for all APIs', () => {
      expect(filterStorageKey()).toBe('http-requests-filter')
    })

    it('has a key per API', () => {
      expect(filterStorageKey('Petstore')).toBe('http-requests-filter-Petstore')
    })

    it('has a key per API, path and method', () => {
      expect(filterStorageKey('Petstore', '/pet', 'GET')).toBe('http-requests-filter-Petstore-/pet-GET')
    })

    it('keeps the key of earlier versions for a path without method', () => {
      expect(filterStorageKey('Petstore', '/pet')).toBe('http-requests-filter-Petstore-/pet-undefined')
    })
  })

  describe('defaultFilter', () => {
    it('has no active filter', () => {
      const filter = defaultFilter()

      expect(filterChips(filter, all)).toEqual([])
      expect(toEventQuery(filter, all)).toEqual({})
    })

    it('returns a new object each time', () => {
      expect(defaultFilter()).not.toBe(defaultFilter())
      expect(defaultFilter().method.value).not.toBe(defaultFilter().method.value)
    })
  })

  describe('restoreFilter', () => {
    it('uses the defaults without a saved filter', () => {
      expect(restoreFilter(null)).toEqual(defaultFilter())
      expect(restoreFilter('')).toEqual(defaultFilter())
    })

    it('uses the defaults for a saved text that is not JSON', () => {
      expect(restoreFilter('{not json')).toEqual(defaultFilter())
    })

    it('uses the defaults for saved JSON that is not an object', () => {
      expect(restoreFilter('42')).toEqual(defaultFilter())
      expect(restoreFilter('null')).toEqual(defaultFilter())
      expect(restoreFilter('[]')).toEqual(defaultFilter())
    })

    it('restores a saved filter', () => {
      const saved = filterWith(f => {
        f.method = { checkbox: true, state: 'Multi', value: ['GET', 'POST'] }
        f.url = { checkbox: true, value: '/pet' }
        f.request.headers = { checkbox: true, value: [{ name: 'Accept', value: 'application/json' }, { name: '', value: '' }] }
        f.response.statusCode = { checkbox: true, value: '4xx' }
        f.clientIP = { checkbox: true, value: '10.0.0.1' }
      })

      expect(restoreFilter(JSON.stringify(saved))).toEqual(saved)
    })

    it('keeps the defaults of what a partial filter leaves out', () => {
      const restored = restoreFilter(JSON.stringify({ url: { checkbox: true, value: '/pet' } }))

      expect(restored).toEqual(filterWith(f => { f.url = { checkbox: true, value: '/pet' } }))
    })

    it('restores a custom method', () => {
      const restored = restoreFilter(JSON.stringify({ method: { state: 'Single', checkbox: true, value: ['CUSTOM'], custom: 'link' } }))

      expect(restored.method.custom).toBe('link')
    })

    it('ignores values of the wrong type', () => {
      const restored = restoreFilter(JSON.stringify({
        service: { state: 'Sometimes', value: 'Petstore' },
        method: { checkbox: 'yes', value: [1, 2] },
        url: { checkbox: 1, value: 5 },
        request: { headers: { value: [{ name: 1 }] } },
        response: { statusCode: { value: ['4xx'] } },
        clientIP: 'localhost',
      }))

      expect(restored).toEqual(defaultFilter())
    })

    it('ignores unknown properties, also the tooltip texts saved by earlier versions', () => {
      const restored = restoreFilter(JSON.stringify({
        unknown: 1,
        clientIP: { checkbox: true, value: '10.0.0.1', title: 'old tooltip' },
      }))

      expect(restored).toEqual(filterWith(f => { f.clientIP = { checkbox: true, value: '10.0.0.1' } }))
    })
  })

  describe('loadFilter and saveFilter', () => {
    afterEach(() => localStorage.clear())

    it('loads the defaults when nothing is saved', () => {
      expect(loadFilter('http-requests-filter')).toEqual(defaultFilter())
    })

    it('saves and loads a filter under its key', () => {
      const filter = filterWith(f => { f.url = { checkbox: true, value: '/pet' } })

      saveFilter('http-requests-filter-Petstore', filter)

      expect(loadFilter('http-requests-filter-Petstore')).toEqual(filter)
      expect(loadFilter('http-requests-filter')).toEqual(defaultFilter())
    })

    it('loads the defaults for corrupt storage', () => {
      localStorage.setItem('http-requests-filter', '{oops')

      expect(loadFilter('http-requests-filter')).toEqual(defaultFilter())
    })
  })

  describe('toEventQuery', () => {
    it('sends the service as a name label only when the filter is on and no service is given', () => {
      const filter = filterWith(f => { f.service = { checkbox: true, state: 'Multi', value: ['A', 'B'] } })

      expect(toEventQuery(filter, all)).toEqual({ labels: [{ name: 'name', value: 'A' }, { name: 'name', value: 'B' }] })
      expect(toEventQuery(filter, petstore)).toEqual({})
      expect(toEventQuery(defaultFilter(), all)).toEqual({})
    })

    it('sends a single method', () => {
      const filter = filterWith(f => { f.method = { checkbox: true, state: 'Single', value: ['POST'] } })

      expect(toEventQuery(filter, all)).toEqual({ methods: ['POST'] })
    })

    it('sends several methods', () => {
      const filter = filterWith(f => { f.method = { checkbox: true, state: 'Multi', value: ['GET', 'POST'] } })

      expect(toEventQuery(filter, all)).toEqual({ methods: ['GET', 'POST'] })
    })

    it('sends the custom methods in upper case, split at spaces and commas', () => {
      const single = filterWith(f => { f.method = { checkbox: true, state: 'Single', value: ['CUSTOM'], custom: 'link, connect' } })
      const multi = filterWith(f => { f.method = { checkbox: true, state: 'Multi', value: ['GET', 'CUSTOM'], custom: 'link connect' } })

      expect(toEventQuery(single, all)).toEqual({ methods: ['LINK', 'CONNECT'] })
      expect(toEventQuery(multi, all)).toEqual({ methods: ['GET', 'LINK', 'CONNECT'] })
    })

    it('sends CUSTOM itself when a multi selection has no custom method', () => {
      const filter = filterWith(f => { f.method = { checkbox: true, state: 'Multi', value: ['CUSTOM'] } })

      expect(toEventQuery(filter, all)).toEqual({ methods: ['CUSTOM'] })
    })

    it('sends the URL only when the filter is on and not empty', () => {
      expect(toEventQuery(filterWith(f => { f.url = { checkbox: true, value: '/pet' } }), all)).toEqual({ url: '/pet' })
      expect(toEventQuery(filterWith(f => { f.url = { checkbox: false, value: '/pet' } }), all)).toEqual({})
      expect(toEventQuery(filterWith(f => { f.url = { checkbox: true, value: '' } }), all)).toEqual({})
    })

    it('sends header filters without the empty rows', () => {
      const headers = [{ name: 'Accept', value: 'json' }, { name: 'X-Only-Name', value: '' }, { name: '', value: '' }]
      const filter = filterWith(f => {
        f.request.headers = { checkbox: true, value: headers }
        f.response.headers = { checkbox: true, value: headers }
      })

      expect(toEventQuery(filter, all)).toEqual({
        requestHeaders: [{ name: 'Accept', value: 'json' }],
        responseHeaders: [{ name: 'Accept', value: 'json' }],
      })
    })

    it('leaves header filters out when they are off', () => {
      const filter = filterWith(f => { f.request.headers = { checkbox: false, value: [{ name: 'Accept', value: 'json' }] } })

      expect(toEventQuery(filter, all)).toEqual({})
    })

    it('sends the status code and the client IP when the filters are on', () => {
      const filter = filterWith(f => {
        f.response.statusCode = { checkbox: true, value: '4xx' }
        f.clientIP = { checkbox: true, value: '10.0.0.1' }
      })

      expect(toEventQuery(filter, all)).toEqual({ statusCode: '4xx', clientIP: '10.0.0.1' })
    })

    it('sends an empty status code of an active filter, as before', () => {
      const filter = filterWith(f => { f.response.statusCode = { checkbox: true, value: '' } })

      expect(toEventQuery(filter, all)).toEqual({ statusCode: '' })
    })

    it('does not send a client IP that is empty', () => {
      const filter = filterWith(f => { f.clientIP = { checkbox: true, value: '' } })

      expect(toEventQuery(filter, all)).toEqual({})
    })
  })

  describe('filterChips', () => {
    it('shows the service with its name', () => {
      const filter = filterWith(f => { f.service = { checkbox: true, state: 'Single', value: ['Petstore'] } })

      expect(filterChips(filter, all)).toEqual([{ field: 'service', label: 'Service: Petstore', title: 'Service: Petstore' }])
    })

    it('shows several services', () => {
      const filter = filterWith(f => { f.service = { checkbox: true, state: 'Multi', value: ['A', 'B'] } })

      expect(filterChips(filter, all)[0].label).toBe('Service: A, B')
    })

    it('has no service chip when the log belongs to a service', () => {
      const filter = filterWith(f => { f.service = { checkbox: true, state: 'Single', value: ['Petstore'] } })

      expect(filterChips(filter, petstore)).toEqual([])
    })

    it('shows the methods', () => {
      const filter = filterWith(f => { f.method = { checkbox: true, state: 'Multi', value: ['GET', 'POST'] } })

      expect(filterChips(filter, all)[0]).toMatchObject({ field: 'method', label: 'Method: GET, POST' })
    })

    it('shows the custom methods instead of CUSTOM', () => {
      const filter = filterWith(f => { f.method = { checkbox: true, state: 'Multi', value: ['GET', 'CUSTOM'], custom: 'link' } })

      expect(filterChips(filter, all)[0].label).toBe('Method: GET, LINK')
    })

    it('shows a method filter that has no method yet', () => {
      const filter = filterWith(f => { f.method = { checkbox: true, state: 'Single', value: ['CUSTOM'] } })

      expect(filterChips(filter, all)).toHaveLength(1)
    })

    it('shows URL, status code and client IP', () => {
      const filter = filterWith(f => {
        f.url = { checkbox: true, value: '/pet' }
        f.response.statusCode = { checkbox: true, value: '4xx' }
        f.clientIP = { checkbox: true, value: '10.0.0.1' }
      })

      expect(filterChips(filter, all).map(c => c.label)).toEqual(['URL: /pet', 'Status: 4xx', 'Client IP: 10.0.0.1'])
    })

    it('shows a header filter as name:value', () => {
      const filter = filterWith(f => {
        f.request.headers = { checkbox: true, value: [{ name: 'Accept', value: 'application/json' }, { name: '', value: '' }] }
      })

      expect(filterChips(filter, all)[0].label).toBe('Request header: Accept:application/json')
    })

    it('shows header filters with all their rows', () => {
      const filter = filterWith(f => {
        f.request.headers = { checkbox: true, value: [{ name: 'Accept', value: 'application/json' }, { name: 'X-A', value: '1' }, { name: '', value: '' }] }
        f.response.headers = { checkbox: true, value: [{ name: 'Content-Type', value: 'text/plain' }, { name: '', value: '' }] }
      })

      expect(filterChips(filter, all).map(c => [c.field, c.title])).toEqual([
        ['requestHeaders', 'Request header: Accept:application/json, X-A:1'],
        ['responseHeaders', 'Response header: Content-Type:text/plain'],
      ])
    })

    it('has no chip for a header row with only a name', () => {
      const filter = filterWith(f => {
        f.request.headers = { checkbox: true, value: [{ name: 'Accept', value: '' }, { name: '', value: '' }] }
      })

      expect(filterChips(filter, all)).toEqual([])
    })

    it('lists only the header rows with name and value', () => {
      const filter = filterWith(f => {
        f.request.headers = { checkbox: true, value: [{ name: 'Accept', value: 'application/json' }, { name: 'X-A', value: '' }, { name: '', value: '' }] }
      })

      expect(filterChips(filter, all).map(c => c.title)).toEqual(['Request header: Accept:application/json'])
    })

    it('has no chip for a filter that matches nothing yet', () => {
      const filter = filterWith(f => {
        f.url = { checkbox: true, value: '' }
        f.request.headers = { checkbox: true, value: [{ name: '', value: '' }] }
        f.response.headers = { checkbox: false, value: [{ name: 'Accept', value: 'json' }, { name: '', value: '' }] }
        f.response.statusCode = { checkbox: true, value: '' }
        f.clientIP = { checkbox: true, value: '' }
      })

      expect(filterChips(filter, all)).toEqual([])
    })

    it('cuts a long label and keeps the full text as title', () => {
      const url = '/pet/findByStatus/with/a/very/long/path/that/goes/on'
      const filter = filterWith(f => { f.url = { checkbox: true, value: url } })

      const [chip] = filterChips(filter, all)

      expect(chip.title).toBe(`URL: ${url}`)
      expect(chip.label.length).toBeLessThanOrEqual(40)
      expect(chip.label.endsWith('…')).toBe(true)
      expect(chip.title.startsWith(chip.label.slice(0, -1))).toBe(true)
    })

    it('does not cut a label of 40 characters', () => {
      const filter = filterWith(f => { f.url = { checkbox: true, value: '/'.padEnd(35, 'x') } })

      expect(filterChips(filter, all)[0].label).toBe(`URL: ${'/'.padEnd(35, 'x')}`)
    })

    it('has as many chips as the filter count of the former Filter button', () => {
      const filter = filterWith(f => {
        f.service = { checkbox: true, state: 'Single', value: ['Petstore'] }
        f.method = { checkbox: true, state: 'Single', value: ['GET'] }
        f.url = { checkbox: true, value: '/pet' }
        f.request.headers = { checkbox: true, value: [{ name: 'A', value: 'b' }, { name: '', value: '' }] }
        f.response.headers = { checkbox: true, value: [{ name: 'A', value: 'b' }, { name: '', value: '' }] }
        f.response.statusCode = { checkbox: true, value: '200' }
        f.clientIP = { checkbox: true, value: '10.0.0.1' }
      })

      expect(filterChips(filter, all)).toHaveLength(7)
    })
  })

  describe('removeFilter', () => {
    const everything = filterWith(f => {
      f.service = { checkbox: true, state: 'Single', value: ['Petstore'] }
      f.method = { checkbox: true, state: 'Multi', value: ['GET', 'POST'], custom: 'link' }
      f.url = { checkbox: true, value: '/pet' }
      f.request.headers = { checkbox: true, value: [{ name: 'A', value: 'b' }, { name: '', value: '' }] }
      f.response.headers = { checkbox: true, value: [{ name: 'A', value: 'b' }, { name: '', value: '' }] }
      f.response.statusCode = { checkbox: true, value: '200' }
      f.clientIP = { checkbox: true, value: '10.0.0.1' }
    })

    it.each([
      ['service', 'Service'],
      ['method', 'Method'],
      ['url', 'URL'],
      ['requestHeaders', 'Request header'],
      ['responseHeaders', 'Response header'],
      ['statusCode', 'Status'],
      ['clientIP', 'Client IP'],
    ] as const)('removes only the %s filter', (field, labelPrefix) => {
      const result = removeFilter(everything, field)

      const labels = filterChips(result, all).map(c => c.label)
      expect(labels).toHaveLength(6)
      expect(labels.some(l => l.startsWith(labelPrefix + ':'))).toBe(false)
    })

    it('resets the removed filter to its default', () => {
      const result = removeFilter(everything, 'method')

      expect(result.method).toEqual(defaultFilter().method)
    })

    it('does not change the given filter', () => {
      const before = JSON.stringify(everything)

      removeFilter(everything, 'url')

      expect(JSON.stringify(everything)).toBe(before)
    })
  })

  describe('addableFields', () => {
    it('offers every field without chip, in the order of the former dialog', () => {
      expect(addableFields(defaultFilter(), all).map(f => f.field)).toEqual(
        ['service', 'method', 'url', 'requestHeaders', 'statusCode', 'responseHeaders', 'clientIP'])
    })

    it('names the fields', () => {
      expect(addableFields(defaultFilter(), all).map(f => f.label)).toEqual(
        ['API Name', 'Method', 'URL', 'Request Header', 'Status Code', 'Response Header', 'Client IP'])
    })

    it('offers the API name only for the log of all APIs', () => {
      expect(addableFields(defaultFilter(), petstore).map(f => f.field)).not.toContain('service')
    })

    it('does not offer a field that has a chip', () => {
      const filter = filterWith(f => { f.url = { checkbox: true, value: '/pet' } })

      expect(addableFields(filter, all).map(f => f.field)).not.toContain('url')
    })

    it('offers a field again after its chip is removed', () => {
      const filter = filterWith(f => { f.url = { checkbox: true, value: '/pet' } })

      expect(addableFields(removeFilter(filter, 'url'), all).map(f => f.field)).toContain('url')
    })
  })

  describe('fieldLabel', () => {
    it('names a field like the entry of the Add filter menu', () => {
      expect(fieldLabel('requestHeaders')).toBe('Request Header')
      expect(fieldLabel('clientIP')).toBe('Client IP')
    })
  })

  describe('activateFilter', () => {
    it('switches a single-value filter on', () => {
      const filter = defaultFilter()

      activateFilter(filter, 'statusCode')
      activateFilter(filter, 'clientIP')
      activateFilter(filter, 'url')

      expect(filter.response.statusCode.checkbox).toBe(true)
      expect(filter.clientIP.checkbox).toBe(true)
      expect(filter.url.checkbox).toBe(true)
    })

    it('switches the header filters on', () => {
      const filter = defaultFilter()

      activateFilter(filter, 'requestHeaders')
      activateFilter(filter, 'responseHeaders')

      expect(filter.request.headers.checkbox).toBe(true)
      expect(filter.response.headers.checkbox).toBe(true)
    })

    it('lets the method and the API select one value first', () => {
      const filter = defaultFilter()

      activateFilter(filter, 'method')
      activateFilter(filter, 'service')

      expect(filter.method).toMatchObject({ state: 'Single', checkbox: true, value: ['GET'] })
      expect(filter.service).toMatchObject({ state: 'Single', checkbox: true })
    })

    it('shows a chip only once the filter has a value', () => {
      const filter = defaultFilter()

      activateFilter(filter, 'statusCode')
      expect(filterChips(filter, all)).toEqual([])

      filter.response.statusCode.value = '404'
      expect(filterChips(filter, all).map(c => c.label)).toEqual(['Status: 404'])
    })
  })
})
