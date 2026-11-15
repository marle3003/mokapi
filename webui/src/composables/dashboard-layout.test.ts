import { beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

// useLocalStorage and the dashboard mode keep state at module level, so every test gets fresh modules.
async function load() {
  vi.resetModules()
  const { useDashboard } = await import('./dashboard')
  const { useDashboardLayout } = await import('./dashboard-layout')
  return { dashboard: useDashboard(), layout: useDashboardLayout() }
}

describe('dashboard layout', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  describe('layout', () => {
    it('is the new layout in the live dashboard by default', async () => {
      const { layout } = await load()

      expect(layout.layout.value).toBe('new')
    })

    it('is read from localStorage in the live dashboard', async () => {
      localStorage.setItem('dashboard-layout', JSON.stringify('classic'))

      const { layout } = await load()

      expect(layout.layout.value).toBe('classic')
    })

    it('treats an unknown stored value as the new layout', async () => {
      localStorage.setItem('dashboard-layout', JSON.stringify('compact'))

      const { layout } = await load()

      expect(layout.layout.value).toBe('new')
    })

    it('treats a stored value that is not valid JSON as the new layout', async () => {
      localStorage.setItem('dashboard-layout', 'classic')

      const { layout } = await load()

      expect(layout.layout.value).toBe('new')
    })

    it('can be switched to the classic layout and back, and is saved', async () => {
      const { layout } = await load()

      layout.setLayout('classic')
      await nextTick()

      expect(layout.layout.value).toBe('classic')
      expect(localStorage.getItem('dashboard-layout')).toBe(JSON.stringify('classic'))

      layout.setLayout('new')
      await nextTick()

      expect(layout.layout.value).toBe('new')
      expect(localStorage.getItem('dashboard-layout')).toBe(JSON.stringify('new'))
    })

    it('is always the new layout in the demo dashboard', async () => {
      localStorage.setItem('dashboard-layout', JSON.stringify('classic'))
      const { dashboard, layout } = await load()

      dashboard.setMode('demo')

      expect(layout.layout.value).toBe('new')
    })
  })

  describe('sidebar', () => {
    it('is open by default in a wide window', async () => {
      const { layout } = await load()

      expect(layout.isSidebarOpen(false)).toBe(true)
    })

    it('toggle collapses and expands the sidebar and persists the state', async () => {
      const { layout } = await load()

      layout.toggleSidebar(false)
      await nextTick()

      expect(layout.isSidebarOpen(false)).toBe(false)
      expect(localStorage.getItem('dashboard-sidebar-collapsed')).toBe('true')

      layout.toggleSidebar(false)

      expect(layout.isSidebarOpen(false)).toBe(true)
    })

    it('restores the collapsed state from localStorage', async () => {
      localStorage.setItem('dashboard-sidebar-collapsed', 'true')

      const { layout } = await load()

      expect(layout.isSidebarOpen(false)).toBe(false)
    })

    it('starts collapsed in a narrow window, whatever was saved', async () => {
      localStorage.setItem('dashboard-sidebar-collapsed', 'false')

      const { layout } = await load()

      expect(layout.isSidebarOpen(true)).toBe(false)
    })

    it('opens as overlay in a narrow window without changing the saved state', async () => {
      const { layout } = await load()

      layout.toggleSidebar(true)
      await nextTick()

      expect(layout.isSidebarOpen(true)).toBe(true)
      expect(localStorage.getItem('dashboard-sidebar-collapsed')).toBeNull()
    })

    it('closeOverlay closes the overlay but keeps the wide state', async () => {
      const { layout } = await load()
      layout.toggleSidebar(true)

      layout.closeOverlay()

      expect(layout.isSidebarOpen(true)).toBe(false)
      expect(layout.isSidebarOpen(false)).toBe(true)
    })
  })

  describe('protocol groups', () => {
    it('are expanded by default', async () => {
      const { layout } = await load()

      expect(layout.isGroupCollapsed('http')).toBe(false)
    })

    it('toggle collapses one group and persists it', async () => {
      const { layout } = await load()

      layout.toggleGroup('http')
      await nextTick()

      expect(layout.isGroupCollapsed('http')).toBe(true)
      expect(layout.isGroupCollapsed('kafka')).toBe(false)
      expect(JSON.parse(localStorage.getItem('dashboard-sidebar-collapsed-groups')!)).toEqual(['http'])

      layout.toggleGroup('http')

      expect(layout.isGroupCollapsed('http')).toBe(false)
    })

    it('restores collapsed groups from localStorage', async () => {
      localStorage.setItem('dashboard-sidebar-collapsed-groups', JSON.stringify(['mail']))

      const { layout } = await load()

      expect(layout.isGroupCollapsed('mail')).toBe(true)
    })
  })
})
