import { describe, expect, it } from 'vitest'
import { isPinTargetAvailable, isPinned, removePin, restorePinned, togglePin } from './pinned'

type PinService = Parameters<typeof isPinTargetAvailable>[1][number]

const services: PinService[] = [
  {
    name: 'Swagger Petstore', type: 'http', items: [
      { name: '/pet', method: 'post' },
      { name: '/pet/{petId}', method: 'get' },
      { name: '/pet/{petId}', method: 'delete' },
    ],
  },
  { name: 'Kafka World', type: 'kafka', items: [{ name: 'mokapi.shop.products' }] },
  { name: 'MQTT Sensors', type: 'mqtt', items: [{ name: 'sensors/{sensorId}/data' }] },
  { name: 'Chat', type: 'websocket', items: [{ name: '/chats/{chatId}' }] },
  { name: 'Mail Server', type: 'mail', items: [{ name: 'alice@mokapi.io' }] },
  { name: 'Directory', type: 'ldap' },
]

describe('pinned entries', () => {
  describe('toggle', () => {
    it('pins a path that is not pinned', () => {
      expect(togglePin(['/dashboard/configs'], '/dashboard/jobs')).toEqual(['/dashboard/configs', '/dashboard/jobs'])
    })

    it('unpins a path that is pinned', () => {
      expect(togglePin(['/dashboard/configs', '/dashboard/jobs'], '/dashboard/configs')).toEqual(['/dashboard/jobs'])
    })

    it('ignores the refresh interval and the hash', () => {
      const pinned = togglePin([], '/dashboard/jobs?refresh=20#top')

      expect(pinned).toEqual(['/dashboard/jobs'])
      expect(isPinned(pinned, '/dashboard/jobs?refresh=5')).toBe(true)
      expect(togglePin(pinned, '/dashboard/jobs?refresh=5')).toEqual([])
    })

    it('ignores the selected request of a log', () => {
      const pinned = togglePin([], '/dashboard/http/log?request=abc')

      expect(pinned).toEqual(['/dashboard/http/log'])
      expect(isPinned(pinned, '/dashboard/http/log?request=def')).toBe(true)
    })

    it('ignores the selected message of a Kafka log', () => {
      const pinned = togglePin([], '/dashboard/kafka/log?message=abc')

      expect(pinned).toEqual(['/dashboard/kafka/log'])
      expect(isPinned(pinned, '/dashboard/kafka/log?message=def')).toBe(true)
    })

    it('ignores the selected message of a WebSocket log', () => {
      const pinned = togglePin([], '/dashboard/websocket/log?message=abc')

      expect(pinned).toEqual(['/dashboard/websocket/log'])
      expect(isPinned(pinned, '/dashboard/websocket/log?message=def')).toBe(true)
    })

    it('ignores the selected message of an MQTT log', () => {
      const pinned = togglePin([], '/dashboard/mqtt/log?message=abc')

      expect(pinned).toEqual(['/dashboard/mqtt/log'])
      expect(isPinned(pinned, '/dashboard/mqtt/log?message=def')).toBe(true)
    })

    it('keeps other query parameters', () => {
      expect(togglePin([], '/dashboard/search?q=pet&refresh=20')).toEqual(['/dashboard/search?q=pet'])
    })

    it('does not change the given list', () => {
      const pinned = ['/dashboard/configs']
      togglePin(pinned, '/dashboard/jobs')
      expect(pinned).toEqual(['/dashboard/configs'])
    })
  })

  describe('remove', () => {
    it('removes a path', () => {
      expect(removePin(['/dashboard/configs', '/dashboard/jobs'], '/dashboard/jobs')).toEqual(['/dashboard/configs'])
    })

    it('leaves the list as it is when the path is not pinned', () => {
      expect(removePin(['/dashboard/configs'], '/dashboard/jobs')).toEqual(['/dashboard/configs'])
    })
  })

  describe('restore', () => {
    it('restores saved paths in their order', () => {
      expect(restorePinned(['/dashboard/jobs', '/dashboard/configs'])).toEqual(['/dashboard/jobs', '/dashboard/configs'])
    })

    it.each([undefined, null, 'jobs', 42, { a: 1 }])('starts empty for %j', saved => {
      expect(restorePinned(saved)).toEqual([])
    })

    it('drops entries that are not paths and duplicates', () => {
      expect(restorePinned(['/dashboard/jobs', 3, null, '/dashboard/jobs?refresh=20', '/dashboard/configs']))
        .toEqual(['/dashboard/jobs', '/dashboard/configs'])
    })
  })

  describe('target availability', () => {
    function available(path: string, list = services) {
      return isPinTargetAvailable(path, list)
    }

    it.each([
      '/dashboard',
      '/dashboard/search',
      '/dashboard/jobs',
      '/dashboard/configs',
      '/dashboard/faker',
      '/dashboard/tree',
      '/dashboard/http',
      '/dashboard/kafka',
      '/dashboard/mqtt',
      '/dashboard/websocket',
      '/dashboard/mail',
      '/dashboard/ldap',
    ])('the fixed page %s is always available', path => {
      expect(available(path, [])).toBe(true)
    })

    it('the full view of a job run is available', () => {
      expect(available('/dashboard/jobs/8134', [])).toBe(true)
      expect(available('/dashboard-demo/jobs/8134', [])).toBe(true)
    })

    it('works the same in the demo dashboard', () => {
      expect(available('/dashboard-demo/http/services/Swagger%20Petstore')).toBe(true)
      expect(available('/dashboard-demo/http/services/Gone')).toBe(false)
    })

    it('a page that does not exist is not available', () => {
      expect(available('/dashboard/unknown')).toBe(false)
      expect(available('/dashboard/http/unknown/Swagger%20Petstore')).toBe(false)
    })

    describe('service', () => {
      it.each([
        ['/dashboard/http/services/Swagger%20Petstore', true],
        ['/dashboard/http/services/Gone', false],
        ['/dashboard/kafka/service/Kafka%20World', true],
        ['/dashboard/kafka/service/Gone', false],
        ['/dashboard/mqtt/service/MQTT%20Sensors', true],
        ['/dashboard/websocket/service/Chat', true],
        ['/dashboard/mail/service/Mail%20Server', true],
        ['/dashboard/ldap/service/Directory', true],
        ['/dashboard/ldap/service/Gone', false],
      ])('%s', (path, expected) => {
        expect(available(path)).toBe(expected)
      })

      it.each([
        ['/dashboard/http/log', true],
        ['/dashboard/http/log/Swagger%20Petstore', true],
        ['/dashboard/http/log/Gone', false],
        ['/dashboard-demo/http/log/Swagger%20Petstore', true],
      ])('request log %s', (path, expected) => {
        expect(available(path)).toBe(expected)
      })

      it.each([
        ['/dashboard/mqtt/log', true],
        ['/dashboard/mqtt/log/MQTT%20Sensors', true],
        ['/dashboard/mqtt/log/Gone', false],
        ['/dashboard-demo/mqtt/log/MQTT%20Sensors', true],
      ])('MQTT message log %s', (path, expected) => {
        expect(available(path)).toBe(expected)
      })

      it.each([
        ['/dashboard/ldap/log', true],
        ['/dashboard/ldap/log/Directory', true],
        ['/dashboard/ldap/log/Gone', false],
        ['/dashboard-demo/ldap/log/Directory', true],
      ])('LDAP request log %s', (path, expected) => {
        expect(available(path)).toBe(expected)
      })

      it.each([
        ['/dashboard/mail/log', true],
        ['/dashboard/mail/log/Mail%20Server', true],
        ['/dashboard/mail/log/Gone', false],
        ['/dashboard-demo/mail/log/Mail%20Server', true],
      ])('mail log %s', (path, expected) => {
        expect(available(path)).toBe(expected)
      })

      it.each([
        ['/dashboard/websocket/log', true],
        ['/dashboard/websocket/log/Chat', true],
        ['/dashboard/websocket/log/Gone', false],
        ['/dashboard-demo/websocket/log/Chat', true],
      ])('WebSocket message log %s', (path, expected) => {
        expect(available(path)).toBe(expected)
      })

      it.each([
        ['/dashboard/kafka/log', true],
        ['/dashboard/kafka/log/Kafka%20World', true],
        ['/dashboard/kafka/log/Gone', false],
        ['/dashboard-demo/kafka/log/Kafka%20World', true],
      ])('message log %s', (path, expected) => {
        expect(available(path)).toBe(expected)
      })

      it('needs the service of that protocol', () => {
        expect(available('/dashboard/kafka/service/Swagger%20Petstore')).toBe(false)
      })

      it('keeps the query', () => {
        expect(available('/dashboard/http/services/Swagger%20Petstore?tab=servers')).toBe(true)
      })
    })

    describe('HTTP endpoint', () => {
      it.each([
        ['/dashboard/http/services/Swagger%20Petstore/pet', true],
        ['/dashboard/http/services/Swagger%20Petstore/pet/%7BpetId%7D', true],
        ['/dashboard/http/services/Swagger%20Petstore/pet/%7BpetId%7D/get', true],
        ['/dashboard/http/services/Swagger%20Petstore/pet/%7BpetId%7D/delete', true],
        ['/dashboard/http/services/Swagger%20Petstore/pet/%7BpetId%7D/put', false],
        ['/dashboard/http/services/Swagger%20Petstore/pet/get', false],
        ['/dashboard/http/services/Swagger%20Petstore/store', false],
        ['/dashboard/http/services/Gone/pet', false],
      ])('%s', (path, expected) => {
        expect(available(path)).toBe(expected)
      })

      it('is not available when the service has no items', () => {
        expect(available('/dashboard/http/services/Bare/pet', [{ name: 'Bare', type: 'http' }])).toBe(false)
      })
    })

    describe('topic, channel and mailbox', () => {
      it.each([
        ['/dashboard/kafka/service/Kafka%20World/topics/mokapi.shop.products', true],
        ['/dashboard/kafka/service/Kafka%20World/topics/gone', false],
        ['/dashboard/mqtt/service/MQTT%20Sensors/topics/sensors%2F%7BsensorId%7D%2Fdata', true],
        ['/dashboard/mqtt/service/MQTT%20Sensors/topics/gone', false],
        ['/dashboard/websocket/service/Chat/channels/%2Fchats%2F%7BchatId%7D', true],
        ['/dashboard/websocket/service/Chat/channels/%2Fgone', false],
        ['/dashboard/mail/service/Mail%20Server/maiboxes/alice%40mokapi.io', true],
        ['/dashboard/mail/service/Mail%20Server/maiboxes/bob%40mokapi.io', false],
      ])('%s', (path, expected) => {
        expect(available(path)).toBe(expected)
      })

      it('is not available when its service is gone', () => {
        expect(available('/dashboard/kafka/service/Gone/topics/mokapi.shop.products')).toBe(false)
      })
    })

    describe('other pages of a service', () => {
      it.each([
        '/dashboard/kafka/service/Kafka%20World/servers/Broker',
        '/dashboard/kafka/service/Kafka%20World/groups/g1',
        '/dashboard/mqtt/service/MQTT%20Sensors/clients/c1',
      ])('%s is available as long as the service exists', path => {
        expect(available(path)).toBe(true)
        expect(available(path, [])).toBe(false)
      })
    })

    describe('pages that depend on recorded data', () => {
      it.each([
        '/dashboard/http/requests/abc',
        '/dashboard/kafka/messages/abc',
        '/dashboard/config/abc',
        '/dashboard/mail/mail/mails/abc',
      ])('%s is not checked', path => {
        expect(available(path, [])).toBe(true)
      })
    })
  })
})
