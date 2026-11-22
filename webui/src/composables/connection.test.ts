import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const networkError = () => Promise.reject(new TypeError('Failed to fetch'))
const serverError = () => Promise.resolve(new Response('error', { status: 500 }))
const ok = () => Promise.resolve(new Response('{}', { status: 200 }))

let connection: ReturnType<typeof import('./connection').useConnection>
let fetchMock: ReturnType<typeof vi.fn>

describe('connection', () => {
  beforeEach(async () => {
    vi.useFakeTimers()
    vi.resetModules()
    fetchMock = vi.fn(ok)
    vi.stubGlobal('fetch', fetchMock)
    const { useConnection } = await import('./connection')
    connection = useConnection()
  })

  afterEach(() => {
    connection.stop()
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  describe('heartbeat', () => {
    it('requests /api/info every 5 seconds', async () => {
      connection.start()

      await vi.advanceTimersByTimeAsync(4999)
      expect(fetchMock).not.toHaveBeenCalled()

      await vi.advanceTimersByTimeAsync(1)
      expect(fetchMock).toHaveBeenCalledTimes(1)
      expect(String(fetchMock.mock.calls[0]![0])).toMatch(/\/api\/info$/)

      await vi.advanceTimersByTimeAsync(5000)
      expect(fetchMock).toHaveBeenCalledTimes(2)
    })

    it('reports a network error as lost connection', async () => {
      connection.start()
      fetchMock.mockImplementation(networkError)

      await vi.advanceTimersByTimeAsync(5000)

      expect(connection.isConnected.value).toBe(false)
    })

    it('stays connected on an HTTP 500', async () => {
      connection.start()
      fetchMock.mockImplementation(serverError)

      await vi.advanceTimersByTimeAsync(5000)

      expect(connection.isConnected.value).toBe(true)
    })

    it('does not request /api/info before start', async () => {
      await vi.advanceTimersByTimeAsync(10000)

      expect(fetchMock).not.toHaveBeenCalled()
    })
  })

  describe('onReconnect', () => {
    it('runs the callback once when the connection is back and not while it stays connected', async () => {
      const onReconnect = vi.fn()
      connection.onReconnect(onReconnect)
      connection.start()

      await vi.advanceTimersByTimeAsync(5000)
      expect(onReconnect).not.toHaveBeenCalled()

      fetchMock.mockImplementation(networkError)
      await vi.advanceTimersByTimeAsync(5000)
      expect(onReconnect).not.toHaveBeenCalled()

      fetchMock.mockImplementation(ok)
      await vi.advanceTimersByTimeAsync(5000)
      expect(connection.isConnected.value).toBe(true)
      expect(onReconnect).toHaveBeenCalledTimes(1)

      await vi.advanceTimersByTimeAsync(5000)
      expect(onReconnect).toHaveBeenCalledTimes(1)
    })

    it('runs the callback again after the next disconnect', async () => {
      const onReconnect = vi.fn()
      connection.onReconnect(onReconnect)
      connection.start()

      for (let i = 0; i < 2; i++) {
        fetchMock.mockImplementation(networkError)
        await vi.advanceTimersByTimeAsync(5000)
        fetchMock.mockImplementation(ok)
        await vi.advanceTimersByTimeAsync(5000)
      }

      expect(onReconnect).toHaveBeenCalledTimes(2)
    })
  })

  describe('monitoredFetch', () => {
    it('reports a rejected GET as lost connection while monitoring', async () => {
      connection.start()
      fetchMock.mockImplementation(networkError)

      await expect(connection.monitoredFetch('/api/search')).rejects.toThrow('Failed to fetch')

      expect(connection.isConnected.value).toBe(false)
    })

    it('reports a rejected POST as lost connection while monitoring', async () => {
      connection.start()
      fetchMock.mockImplementation(networkError)

      await expect(connection.monitoredFetch('/api/schema/validate', { method: 'POST', body: '{}' }))
        .rejects.toThrow('Failed to fetch')

      expect(connection.isConnected.value).toBe(false)
      expect(fetchMock).toHaveBeenCalledWith('/api/schema/validate', { method: 'POST', body: '{}' })
    })

    it('ignores network errors when the monitor is not started', async () => {
      fetchMock.mockImplementation(networkError)

      await expect(connection.monitoredFetch('/api/search')).rejects.toThrow('Failed to fetch')

      expect(connection.isConnected.value).toBe(true)
    })

    it('stays connected on an HTTP 500', async () => {
      connection.start()
      fetchMock.mockImplementation(serverError)

      const res = await connection.monitoredFetch('/api/search')

      expect(res.status).toBe(500)
      expect(connection.isConnected.value).toBe(true)
    })
  })
})
