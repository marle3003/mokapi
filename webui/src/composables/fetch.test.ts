import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

// refresh-manager reads the refresh interval of the current route and would otherwise import all views
vi.mock('@/router', async () => {
  const { ref } = await import('vue')
  return { default: { currentRoute: ref({ query: { refresh: '1' } }) } }
})

const networkError = () => Promise.reject(new TypeError('Failed to fetch'))
const ok = () => Promise.resolve(new Response('{"name":"x"}', { headers: { 'content-type': 'application/json' } }))
const serverError = () => Promise.resolve(new Response('boom', { status: 500 }))

const path = '/api/test'
const heartbeatPath = '/api/info'

let useFetch: typeof import('./fetch').useFetch
let connection: ReturnType<typeof import('./connection').useConnection>
let respondHeartbeat: () => Promise<Response>
let respondPath: () => Promise<Response>
let fetchMock: ReturnType<typeof vi.fn>

function pathCalls() {
  return fetchMock.mock.calls.filter(([url]) => url === path)
}

async function loseConnection() {
  respondHeartbeat = networkError
  await vi.advanceTimersByTimeAsync(5000)
  expect(connection.isConnected.value).toBe(false)
}

async function restoreConnection() {
  respondHeartbeat = ok
  await vi.advanceTimersByTimeAsync(5000)
  expect(connection.isConnected.value).toBe(true)
}

describe('useFetch', () => {
  beforeEach(async () => {
    vi.useFakeTimers()
    vi.resetModules()
    respondHeartbeat = ok
    respondPath = ok
    fetchMock = vi.fn((url: string) => url === heartbeatPath ? respondHeartbeat() : respondPath())
    vi.stubGlobal('fetch', fetchMock)
    vi.spyOn(console, 'error').mockImplementation(() => {})
    connection = (await import('./connection')).useConnection()
    useFetch = (await import('./fetch')).useFetch
    connection.start()
  })

  afterEach(() => {
    connection.stop()
    vi.useRealTimers()
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('stays connected on an HTTP 500 and sets the response error', async () => {
    respondPath = serverError

    const response = useFetch(path, undefined, false)
    await vi.advanceTimersByTimeAsync(0)

    expect(response.error).toContain('boom')
    expect(response.isLoading).toBe(false)
    expect(connection.isConnected.value).toBe(true)
  })

  it('reports a network error as lost connection without waiting for the heartbeat', async () => {
    respondPath = networkError

    const response = useFetch(path, undefined, false)
    await vi.advanceTimersByTimeAsync(0)

    expect(connection.isConnected.value).toBe(false)
    expect(response.isLoading).toBe(true)
    expect(response.error).toBeNull()
  })

  it('sets the response error on a network error when the connection is not monitored', async () => {
    connection.stop()
    respondPath = networkError

    const response = useFetch(path, undefined, false)
    await vi.advanceTimersByTimeAsync(0)

    expect(response.error).toContain('Failed to fetch')
    expect(response.isLoading).toBe(false)
    expect(connection.isConnected.value).toBe(true)
  })

  describe('one-off fetch', () => {
    it('is repeated once after reconnect', async () => {
      respondPath = networkError
      const response = useFetch(path, undefined, false)
      await vi.advanceTimersByTimeAsync(0)
      expect(pathCalls()).toHaveLength(1)

      respondPath = ok
      await restoreConnection()
      expect(pathCalls()).toHaveLength(2)
      expect(response.data).toEqual({ name: 'x' })

      await vi.advanceTimersByTimeAsync(10000)
      expect(pathCalls()).toHaveLength(2)
    })

    it('repeats a POST with its body once after reconnect', async () => {
      const init = { method: 'POST', body: '{"schema":1}' }
      respondPath = networkError
      useFetch(path, init, false, false)
      await vi.advanceTimersByTimeAsync(0)

      respondPath = ok
      await restoreConnection()
      await vi.advanceTimersByTimeAsync(10000)

      expect(pathCalls()).toHaveLength(2)
      expect(pathCalls()[1]).toEqual([path, init])
    })

    it('is not repeated again after a later reconnect', async () => {
      respondPath = networkError
      useFetch(path, undefined, false)
      await vi.advanceTimersByTimeAsync(0)
      respondPath = ok
      await restoreConnection()
      expect(pathCalls()).toHaveLength(2)

      await loseConnection()
      await restoreConnection()

      expect(pathCalls()).toHaveLength(2)
    })

    it('is not repeated when it succeeded', async () => {
      useFetch(path, undefined, false)
      await vi.advanceTimersByTimeAsync(0)

      await loseConnection()
      await restoreConnection()

      expect(pathCalls()).toHaveLength(1)
    })
  })

  describe('refreshable fetch', () => {
    it('is loaded once after reconnect', async () => {
      const response = useFetch(path)
      await vi.advanceTimersByTimeAsync(0)
      expect(pathCalls()).toHaveLength(1)

      await loseConnection()
      await restoreConnection()
      expect(pathCalls()).toHaveLength(2)

      await vi.advanceTimersByTimeAsync(0)
      expect(pathCalls()).toHaveLength(2)
      response.close()
    })
  })
})
