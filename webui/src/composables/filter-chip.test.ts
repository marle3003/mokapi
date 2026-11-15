import { describe, expect, it } from 'vitest'
import { shorten } from './filter-chip'

describe('filter chip', () => {
  describe('shorten', () => {
    it('keeps a text of up to 40 characters', () => {
      expect(shorten('')).toBe('')
      expect(shorten('x'.repeat(40))).toBe('x'.repeat(40))
    })

    it('cuts a longer text to 40 characters with an ellipsis', () => {
      const result = shorten('x'.repeat(41))

      expect(result).toBe('x'.repeat(39) + '…')
      expect(result.length).toBe(40)
    })
  })
})
