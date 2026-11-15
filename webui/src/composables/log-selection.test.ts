import { describe, expect, it } from 'vitest'
import { logKeyAction, neighbourId, selectedLogId, withSelection } from './log-selection'

function keyEvent(key: string, init: KeyboardEventInit = {}, target?: Element): KeyboardEvent {
  const event = new KeyboardEvent('keydown', { key, bubbles: true, ...init })
  if (target) {
    // the target of a dispatched event is set by the browser, a plain object is enough here
    Object.defineProperty(event, 'target', { value: target })
  }
  return event
}

const open = { panelOpen: true, popoverOpen: false }

describe('log selection', () => {
  describe('selectedLogId', () => {
    it('reads the request of the query', () => {
      expect(selectedLogId({ request: 'abc' }, 'request')).toBe('abc')
    })

    it('takes the first when the key is repeated', () => {
      expect(selectedLogId({ request: ['abc', 'def'] }, 'request')).toBe('abc')
    })

    it('has no selection without a usable value', () => {
      expect(selectedLogId({}, 'request')).toBeUndefined()
      expect(selectedLogId({ request: '' }, 'request')).toBeUndefined()
      expect(selectedLogId({ request: null }, 'request')).toBeUndefined()
      expect(selectedLogId({ request: [] }, 'request')).toBeUndefined()
      expect(selectedLogId({ request: [null] }, 'request')).toBeUndefined()
    })
  })

  describe('withSelection', () => {
    it('adds the request and keeps the other parameters', () => {
      expect(withSelection({ refresh: '5' }, 'request', 'abc')).toEqual({ refresh: '5', request: 'abc' })
    })

    it('replaces the selected request', () => {
      expect(withSelection({ request: 'abc', refresh: '5' }, 'request', 'def')).toEqual({ refresh: '5', request: 'def' })
    })

    it('removes the key on close, and keeps refresh', () => {
      const result = withSelection({ request: 'abc', refresh: '5' }, 'request', undefined)

      expect(result).toEqual({ refresh: '5' })
      expect('request' in result).toBe(false)
    })

    it('does not change the given query', () => {
      const query = { request: 'abc' }

      withSelection(query, 'request', undefined)

      expect(query).toEqual({ request: 'abc' })
    })
  })

  describe('with the message key', () => {
    it('reads and sets the selected message, and ignores a selected request', () => {
      expect(selectedLogId({ message: 'abc', request: 'def' }, 'message')).toBe('abc')
      expect(selectedLogId({ request: 'def' }, 'message')).toBeUndefined()
      expect(withSelection({ refresh: '5' }, 'message', 'abc')).toEqual({ refresh: '5', message: 'abc' })
    })

    it('removes only the message key on close, and keeps refresh', () => {
      const result = withSelection({ message: 'abc', refresh: '5' }, 'message', undefined)

      expect(result).toEqual({ refresh: '5' })
      expect('message' in result).toBe(false)
    })
  })

  describe('neighbourId', () => {
    const ids = ['a', 'b', 'c']

    it('moves down and up', () => {
      expect(neighbourId(ids, 'b', 'down')).toBe('c')
      expect(neighbourId(ids, 'b', 'up')).toBe('a')
    })

    it('stops at the last and at the first row', () => {
      expect(neighbourId(ids, 'c', 'down')).toBe('c')
      expect(neighbourId(ids, 'a', 'up')).toBe('a')
    })

    it('selects the first row when the current request is not on the page', () => {
      expect(neighbourId(ids, 'other', 'down')).toBe('a')
      expect(neighbourId(ids, 'other', 'up')).toBe('a')
      expect(neighbourId(ids, undefined, 'down')).toBe('a')
    })

    it('has no neighbour on an empty page', () => {
      expect(neighbourId([], 'a', 'down')).toBeUndefined()
    })
  })

  describe('logKeyAction', () => {
    it('maps the arrow keys while the panel is open', () => {
      expect(logKeyAction(keyEvent('ArrowDown'), open)).toBe('down')
      expect(logKeyAction(keyEvent('ArrowUp'), open)).toBe('up')
    })

    it('closes the panel with Escape', () => {
      expect(logKeyAction(keyEvent('Escape'), open)).toBe('close-panel')
    })

    it('closes an open popover before the panel', () => {
      expect(logKeyAction(keyEvent('Escape'), { panelOpen: true, popoverOpen: true })).toBe('close-popover')
    })

    it('does nothing while the panel is closed', () => {
      const closed = { panelOpen: false, popoverOpen: false }

      expect(logKeyAction(keyEvent('ArrowDown'), closed)).toBeUndefined()
      expect(logKeyAction(keyEvent('Escape'), closed)).toBeUndefined()
    })

    it('closes an open popover while the panel is closed', () => {
      expect(logKeyAction(keyEvent('Escape'), { panelOpen: false, popoverOpen: true })).toBe('close-popover')
    })

    it('ignores other keys', () => {
      expect(logKeyAction(keyEvent('a'), open)).toBeUndefined()
      expect(logKeyAction(keyEvent('Enter'), open)).toBeUndefined()
    })

    it.each(['input', 'textarea', 'select'])('ignores keys typed in a %s', (tag) => {
      const target = document.createElement(tag)

      expect(logKeyAction(keyEvent('ArrowDown', {}, target), open)).toBeUndefined()
      expect(logKeyAction(keyEvent('Escape', {}, target), { panelOpen: true, popoverOpen: true })).toBeUndefined()
    })

    it('ignores keys typed in editable content', () => {
      const target = document.createElement('div')
      target.contentEditable = 'true'
      // jsdom doesn't compute isContentEditable
      Object.defineProperty(target, 'isContentEditable', { value: true })

      expect(logKeyAction(keyEvent('ArrowDown', {}, target), open)).toBeUndefined()
    })

    it('handles keys of other elements', () => {
      expect(logKeyAction(keyEvent('ArrowDown', {}, document.createElement('button')), open)).toBe('down')
    })

    it.each(['ctrlKey', 'metaKey', 'altKey', 'shiftKey'])('ignores keys with %s', (modifier) => {
      expect(logKeyAction(keyEvent('ArrowDown', { [modifier]: true }), open)).toBeUndefined()
      expect(logKeyAction(keyEvent('Escape', { [modifier]: true }), open)).toBeUndefined()
    })

    it('ignores keys while text is composed', () => {
      expect(logKeyAction(keyEvent('Escape', { isComposing: true }), open)).toBeUndefined()
    })
  })
})
