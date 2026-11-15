import { describe, expect, it } from 'vitest'
import {
  activateTab,
  closeTab,
  emptyTabs,
  endpointMethod,
  maxTabs,
  openTab,
  restoreTabs,
  showPath,
  tabIdentity,
  viewKey,
  tabTitle,
  type TabsState,
} from './dashboard-tabs'

// opens every path in its own tab; the last one is active
function withTabs(...paths: string[]): TabsState {
  return paths.reduce((state, path) => openTab(state, path), emptyTabs())
}

function tabPaths(state: TabsState): string[] {
  return state.tabs.map(t => t.path)
}

function activePath(state: TabsState): string | undefined {
  return state.tabs.find(t => t.id === state.active)?.path
}

function idOf(state: TabsState, index: number): number {
  return state.tabs[index].id
}

describe('dashboard tabs', () => {
  describe('openTab', () => {
    it('adds the path as a new active tab', () => {
      const state = openTab(withTabs('/dashboard'), '/dashboard/configs')

      expect(tabPaths(state)).toEqual(['/dashboard', '/dashboard/configs'])
      expect(activePath(state)).toBe('/dashboard/configs')
    })

    it('opens another tab for a path that is already open', () => {
      const state = openTab(withTabs('/dashboard', '/dashboard/configs'), '/dashboard')

      expect(tabPaths(state)).toEqual(['/dashboard', '/dashboard/configs', '/dashboard'])
      expect(state.active).toBe(idOf(state, 2))
    })

    it('gives every tab its own id', () => {
      const state = withTabs('/a', '/a', '/a')

      expect(new Set(state.tabs.map(t => t.id)).size).toBe(3)
    })

    it('never reuses the id of an open tab after other tabs were closed', () => {
      let state = withTabs('/a', '/b', '/c')
      state = closeTab(state, idOf(state, 1))

      state = openTab(state, '/d')

      expect(new Set(state.tabs.map(t => t.id)).size).toBe(3)
    })

    it('evicts the least recently active tab when the limit is reached', () => {
      const paths = Array.from({ length: maxTabs }, (_, i) => `/dashboard/http/services/s${i}`)
      let state = withTabs(...paths)
      state = activateTab(state, idOf(state, 0)) // s0 is now the most recent, s1 the least recent

      state = openTab(state, '/dashboard/configs')

      expect(tabPaths(state)).toHaveLength(maxTabs)
      expect(tabPaths(state)).not.toContain(paths[1])
      expect(tabPaths(state)).toContain(paths[0])
      expect(activePath(state)).toBe('/dashboard/configs')
    })

    it('evicts tabs until the new one fits into the row', () => {
      const state = openTab(withTabs('/a', '/b', '/c', '/d'), '/e', 3)

      expect(tabPaths(state)).toEqual(['/c', '/d', '/e'])
      expect(activePath(state)).toBe('/e')
    })

    it('keeps at least the new tab when nothing fits', () => {
      const state = openTab(withTabs('/a', '/b'), '/c', 0)

      expect(tabPaths(state)).toEqual(['/c'])
    })

    it('stores the path without the selected request', () => {
      const state = openTab(emptyTabs(), '/dashboard/http/log?request=abc')

      expect(tabPaths(state)).toEqual(['/dashboard/http/log'])
    })

    it('stores the path without the refresh query', () => {
      const state = openTab(emptyTabs(), '/dashboard/http?refresh=20&name=a')

      expect(tabPaths(state)).toEqual(['/dashboard/http?name=a'])
    })
  })

  describe('activateTab', () => {
    it('makes the tab active and the most recently active one', () => {
      let state = withTabs('/a', '/b', '/c')

      state = activateTab(state, idOf(state, 0))

      expect(activePath(state)).toBe('/a')
      expect(state.tabs[0].lastActive).toBeGreaterThan(state.tabs[2].lastActive)
    })

    it('ignores an id that is not open', () => {
      const state = withTabs('/a', '/b')

      expect(activateTab(state, 999)).toEqual(state)
    })
  })

  describe('showPath', () => {
    it('replaces the active tab in place', () => {
      let state = withTabs('/a', '/b', '/c')
      state = activateTab(state, idOf(state, 1))
      const id = state.active

      state = showPath(state, '/x')

      expect(tabPaths(state)).toEqual(['/a', '/x', '/c'])
      expect(state.active).toBe(id)
    })

    it('activates the open tab of the target instead', () => {
      const state = showPath(withTabs('/a', '/b', '/c'), '/a')

      expect(tabPaths(state)).toEqual(['/a', '/b', '/c'])
      expect(activePath(state)).toBe('/a')
    })

    it('does nothing to the other tabs when the target is the active tab', () => {
      const before = withTabs('/a', '/b')

      const state = showPath(before, '/b?refresh=10')

      expect(tabPaths(state)).toEqual(['/a', '/b'])
      expect(state.active).toBe(before.active)
    })

    it('keeps the active tab when it already shows the path, although another tab does too', () => {
      let state = withTabs('/a', '/b', '/a')
      state = activateTab(state, idOf(state, 2))
      state = activateTab(state, idOf(state, 1))
      state = activateTab(state, idOf(state, 2))

      state = showPath(state, '/a')

      expect(state.active).toBe(idOf(state, 2))
    })

    it('activates the most recently active of several tabs with the path', () => {
      let state = withTabs('/a', '/b', '/a', '/c')
      state = activateTab(state, idOf(state, 0))
      state = activateTab(state, idOf(state, 2)) // the second /a was active more recently
      state = activateTab(state, idOf(state, 3))

      state = showPath(state, '/a')

      expect(state.active).toBe(idOf(state, 2))
      expect(tabPaths(state)).toEqual(['/a', '/b', '/a', '/c'])
    })

    it('opens the first tab when there is none', () => {
      const state = showPath(emptyTabs(), '/a')

      expect(tabPaths(state)).toEqual(['/a'])
      expect(activePath(state)).toBe('/a')
    })

    it('replaces an active tab that is not a tab anymore by opening a new one', () => {
      const state = showPath({ tabs: [{ id: 1, path: '/a', lastActive: 1 }], active: 99 }, '/b')

      expect(tabPaths(state)).toEqual(['/a', '/b'])
    })
  })

  describe('closeTab', () => {
    it('activates the left neighbor when the active tab is closed', () => {
      let state = withTabs('/a', '/b', '/c')

      state = closeTab(state, idOf(state, 2))

      expect(tabPaths(state)).toEqual(['/a', '/b'])
      expect(activePath(state)).toBe('/b')
    })

    it('activates the right neighbor when the first tab is closed', () => {
      let state = withTabs('/a', '/b', '/c')
      state = activateTab(state, idOf(state, 0))

      state = closeTab(state, idOf(state, 0))

      expect(tabPaths(state)).toEqual(['/b', '/c'])
      expect(activePath(state)).toBe('/b')
    })

    it('keeps the active tab when another tab is closed', () => {
      let state = withTabs('/a', '/b', '/c')

      state = closeTab(state, idOf(state, 0))

      expect(tabPaths(state)).toEqual(['/b', '/c'])
      expect(activePath(state)).toBe('/c')
    })

    it('closes only the tab with the id when several tabs show the path', () => {
      let state = withTabs('/a', '/b', '/a')
      const first = idOf(state, 0)

      state = closeTab(state, idOf(state, 2))

      expect(tabPaths(state)).toEqual(['/a', '/b'])
      expect(state.tabs[0].id).toBe(first)
      expect(activePath(state)).toBe('/b')
    })

    it('leaves no active tab when the last tab is closed', () => {
      let state = withTabs('/a')

      state = closeTab(state, idOf(state, 0))

      expect(tabPaths(state)).toEqual([])
      expect(state.active).toBeNull()
    })

    it('opens the fallback path as the only tab when the last tab is closed', () => {
      let state = withTabs('/dashboard')

      state = closeTab(state, idOf(state, 0), '/dashboard')

      expect(tabPaths(state)).toEqual(['/dashboard'])
      expect(activePath(state)).toBe('/dashboard')
    })

    it('does not open the fallback path while other tabs are left', () => {
      let state = withTabs('/a', '/b')

      state = closeTab(state, idOf(state, 1), '/dashboard')

      expect(tabPaths(state)).toEqual(['/a'])
    })

    it('ignores an id that is not open', () => {
      const state = withTabs('/a', '/b')

      expect(closeTab(state, 999)).toEqual(state)
    })
  })

  describe('restoreTabs', () => {
    it('restores the saved tabs and the active tab', () => {
      let saved = withTabs('/a', '/b')
      saved = activateTab(saved, idOf(saved, 0))

      const state = restoreTabs(saved, '/a')

      expect(tabPaths(state)).toEqual(['/a', '/b'])
      expect(activePath(state)).toBe('/a')
    })

    it('adds the URL as an additional tab and activates it', () => {
      const state = restoreTabs(withTabs('/a', '/b'), '/c')

      expect(tabPaths(state)).toEqual(['/a', '/b', '/c'])
      expect(activePath(state)).toBe('/c')
    })

    it('activates the tab of the URL when it is already open', () => {
      const state = restoreTabs(withTabs('/a', '/b'), '/a?refresh=20')

      expect(tabPaths(state)).toEqual(['/a', '/b'])
      expect(activePath(state)).toBe('/a')
    })

    it('keeps duplicate tabs and activates the most recently active one for the URL', () => {
      let saved = withTabs('/a', '/b', '/a', '/c')
      saved = activateTab(saved, idOf(saved, 2))
      saved = activateTab(saved, idOf(saved, 3))

      const state = restoreTabs(saved, '/a')

      expect(tabPaths(state)).toEqual(['/a', '/b', '/a', '/c'])
      expect(state.active).toBe(idOf(saved, 2))
    })

    it('starts with the URL when nothing was saved', () => {
      const state = restoreTabs(undefined, '/a')

      expect(tabPaths(state)).toEqual(['/a'])
      expect(activePath(state)).toBe('/a')
    })

    it.each([
      ['a string', 'classic'],
      ['no tabs array', { active: 1 }],
      ['tabs that are no objects', { tabs: ['/a'], active: 1 }],
    ])('starts with the URL when the saved value is %s', (_, saved) => {
      const state = restoreTabs(saved as unknown as TabsState, '/b')

      expect(tabPaths(state)).toEqual(['/b'])
    })

    it('evicts the least recently active tab when the URL does not fit', () => {
      const state = restoreTabs(withTabs('/a', '/b', '/c'), '/d', 3)

      expect(tabPaths(state)).toEqual(['/b', '/c', '/d'])
    })

    describe('tabs saved without ids (earlier format, the active tab is a path)', () => {
      const saved = {
        tabs: [{ path: '/a', lastActive: 1 }, { path: '/b', lastActive: 3 }, { path: '/c', lastActive: 2 }],
        active: '/b',
      } as unknown as TabsState

      it('restores the tabs with new ids and keeps the active tab', () => {
        const state = restoreTabs(saved, '/b')

        expect(tabPaths(state)).toEqual(['/a', '/b', '/c'])
        expect(new Set(state.tabs.map(t => t.id)).size).toBe(3)
        expect(activePath(state)).toBe('/b')
      })

      it('keeps the order of last activity', () => {
        const state = restoreTabs(saved, '/c')
        const [a, b, c] = state.tabs

        expect(b.lastActive).toBeLessThan(c.lastActive)
        expect(a.lastActive).toBeLessThan(b.lastActive)
      })

      it('opens new tabs with an id none of the restored tabs has', () => {
        const state = restoreTabs(saved, '/d')

        expect(new Set(state.tabs.map(t => t.id)).size).toBe(4)
      })
    })

    it('gives unique ids to saved tabs that share an id', () => {
      const saved = {
        tabs: [{ id: 1, path: '/a', lastActive: 1 }, { id: 1, path: '/b', lastActive: 2 }],
        active: 1,
      }

      const state = restoreTabs(saved, '/b')

      expect(new Set(state.tabs.map(t => t.id)).size).toBe(2)
      expect(activePath(state)).toBe('/b')
    })
  })

  describe('tabIdentity', () => {
    it('removes the refresh query and keeps other parameters', () => {
      expect(tabIdentity('/dashboard/configs?refresh=20')).toBe('/dashboard/configs')
      expect(tabIdentity('/dashboard/configs?a=1&refresh=20&b=2')).toBe('/dashboard/configs?a=1&b=2')
      expect(tabIdentity('/dashboard/configs')).toBe('/dashboard/configs')
    })

    it('removes the selected request like the refresh query', () => {
      expect(tabIdentity('/dashboard/http/log?request=abc')).toBe('/dashboard/http/log')
      expect(tabIdentity('/dashboard/http/log?a=1&request=abc&refresh=5&b=2')).toBe('/dashboard/http/log?a=1&b=2')
      expect(tabIdentity('/dashboard/http/log?request=abc#tab-requests')).toBe('/dashboard/http/log')
    })

    it('keeps a parameter that only starts with a view-state name', () => {
      expect(tabIdentity('/dashboard/search?requests=1&refreshing=2')).toBe('/dashboard/search?requests=1&refreshing=2')
    })

    it('ignores the hash of a sub tab', () => {
      expect(tabIdentity('/dashboard/configs?refresh=20#tab-servers')).toBe('/dashboard/configs')
      expect(tabIdentity('/dashboard/configs?a=1#tab-servers')).toBe('/dashboard/configs?a=1')
    })
  })

  describe('viewKey', () => {
    it('stays the same when a request is selected, so the view stays mounted', () => {
      expect(viewKey('/dashboard/http/log?request=abc')).toBe('/dashboard/http/log')
      expect(viewKey('/dashboard/http/log?request=abc&refresh=5')).toBe('/dashboard/http/log?refresh=5')
      expect(viewKey('/dashboard/http/log?a=1&request=abc&b=2')).toBe('/dashboard/http/log?a=1&b=2')
    })

    it('still changes with the refresh interval and the hash', () => {
      expect(viewKey('/dashboard/configs?refresh=20')).toBe('/dashboard/configs?refresh=20')
      expect(viewKey('/dashboard/configs?request=abc#tab-servers')).toBe('/dashboard/configs#tab-servers')
    })

    it('keeps a parameter that only starts with request', () => {
      expect(viewKey('/dashboard/search?requests=1')).toBe('/dashboard/search?requests=1')
    })
  })

  describe('tabTitle', () => {
    it('removes the Dashboard prefix', () => {
      expect(tabTitle('Dashboard - HTTP Endpoint – Petstore/pet')).toBe('HTTP Endpoint – Petstore/pet')
    })

    it('shows the plain Dashboard title as Overview', () => {
      expect(tabTitle('Dashboard')).toBe('Overview')
    })

    it('keeps a title without the prefix', () => {
      expect(tabTitle('Mokapi')).toBe('Mokapi')
    })
  })

  describe('endpointMethod', () => {
    it('is the last segment of an endpoint when it is an HTTP method', () => {
      expect(endpointMethod(['pet', '{petId}', 'get'])).toBe('get')
    })

    it('is undefined for a path without operation', () => {
      expect(endpointMethod(['pet', '{petId}'])).toBeUndefined()
    })

    it('is undefined for a single segment', () => {
      expect(endpointMethod(['get'])).toBeUndefined()
      expect(endpointMethod(undefined)).toBeUndefined()
    })
  })
})
