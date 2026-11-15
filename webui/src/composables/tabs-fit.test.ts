import { describe, expect, it } from 'vitest'
import { fitTabs, maxTabsThatFit, maxTabWidth, minTabWidth, moreButtonWidth } from './tabs-fit'

describe('tabs fit', () => {
  describe('maxTabsThatFit', () => {
    it('is the number of tabs that fit at minimum width', () => {
      expect(maxTabsThatFit(minTabWidth * 4 + 10)).toBe(4)
    })

    it('is at least one', () => {
      expect(maxTabsThatFit(10)).toBe(1)
      expect(maxTabsThatFit(0)).toBe(1)
    })
  })

  describe('fitTabs', () => {
    it('shows all tabs without a +N menu when they fit at minimum width', () => {
      const fit = fitTabs(3, 0, minTabWidth * 3)

      expect(fit.visible).toEqual([0, 1, 2])
      expect(fit.overflow).toEqual([])
    })

    it('shrinks the tabs to share the width, up to a maximum', () => {
      expect(fitTabs(3, 0, minTabWidth * 3 + 60).tabWidth).toBe(minTabWidth + 20)
      expect(fitTabs(1, 0, 2000).tabWidth).toBe(maxTabWidth)
      expect(fitTabs(3, 0, minTabWidth * 3).tabWidth).toBe(minTabWidth)
    })

    it('moves the tabs that do not fit into the +N menu and keeps room for its button', () => {
      const fit = fitTabs(6, 0, minTabWidth * 3 + moreButtonWidth)

      expect(fit.visible).toEqual([0, 1, 2])
      expect(fit.overflow).toEqual([3, 4, 5])
      expect(fit.tabWidth).toBe(minTabWidth)
    })

    it('keeps the active tab visible by taking the place of the last visible tab', () => {
      const fit = fitTabs(6, 5, minTabWidth * 3 + moreButtonWidth)

      expect(fit.visible).toEqual([0, 1, 5])
      expect(fit.overflow).toEqual([2, 3, 4])
    })

    it('keeps the order of the tabs when the active tab is visible anyway', () => {
      const fit = fitTabs(6, 1, minTabWidth * 3 + moreButtonWidth)

      expect(fit.visible).toEqual([0, 1, 2])
    })

    it('shows at least the active tab in a very narrow row', () => {
      const fit = fitTabs(4, 2, 50)

      expect(fit.visible).toEqual([2])
      expect(fit.overflow).toEqual([0, 1, 3])
    })

    it('has no tabs and no menu without tabs', () => {
      const fit = fitTabs(0, -1, 500)

      expect(fit.visible).toEqual([])
      expect(fit.overflow).toEqual([])
    })

    it('never loses a tab', () => {
      const fit = fitTabs(10, 7, 400)

      expect([...fit.visible, ...fit.overflow].sort((a, b) => a - b)).toEqual([0, 1, 2, 3, 4, 5, 6, 7, 8, 9])
    })
  })
})
