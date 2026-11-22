import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import MarkdownIt from 'markdown-it'
import { MarkdownItTabs } from './markdown-tabs'
import { copyCode, selectCode } from './code-block'

const markdown = [
  '```javascript tab=First',
  'const first = 1',
  '```',
  '',
  '```javascript tab=Second',
  'const second = 2',
  '```'
].join('\n')

let content: HTMLElement
let writeText: ReturnType<typeof vi.fn>

function renderCodeBlock() {
  const md = new MarkdownIt()
  md.use(MarkdownItTabs)
  content = document.createElement('div')
  content.innerHTML = md.render(markdown)
  content.addEventListener('click', copyCode)
  content.addEventListener('dblclick', selectCode)
  document.body.appendChild(content)
}

function copyButton() {
  return content.querySelector('[data-copy]')!
}

function activateSecondTab() {
  const panes = content.querySelectorAll('.tab-pane')
  panes[0]!.classList.remove('show', 'active')
  panes[1]!.classList.add('show', 'active')
}

function click(target: Element) {
  target.dispatchEvent(new MouseEvent('click', { bubbles: true }))
}

function doubleClick(target: Element) {
  target.dispatchEvent(new MouseEvent('dblclick', { bubbles: true }))
}

describe('code block', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true })
    renderCodeBlock()
  })

  afterEach(() => {
    vi.useRealTimers()
    delete (navigator as { clipboard?: unknown }).clipboard
    content.remove()
    getSelection()?.removeAllRanges()
  })

  describe('copy button', () => {
    it('copies the code of the active tab and marks the button as copied', async () => {
      activateSecondTab()

      click(copyButton())
      await vi.advanceTimersByTimeAsync(0)

      expect(writeText).toHaveBeenCalledTimes(1)
      expect(writeText).toHaveBeenCalledWith(expect.stringContaining('const second = 2'))
      expect(writeText).not.toHaveBeenCalledWith(expect.stringContaining('const first = 1'))
      expect(copyButton().classList.contains('copied')).toBe(true)
    })

    it('removes the copied mark after 1.5 seconds', async () => {
      click(copyButton())
      await vi.advanceTimersByTimeAsync(1499)
      expect(copyButton().classList.contains('copied')).toBe(true)

      await vi.advanceTimersByTimeAsync(1)

      expect(copyButton().classList.contains('copied')).toBe(false)
    })

    it('ignores clicks outside the copy button', async () => {
      click(content.querySelector('pre code')!)
      await vi.advanceTimersByTimeAsync(0)

      expect(writeText).not.toHaveBeenCalled()
    })

    it('does not mark the button as copied when writing to the clipboard fails', async () => {
      const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
      writeText.mockRejectedValue(new Error('denied'))

      click(copyButton())
      await vi.advanceTimersByTimeAsync(0)

      expect(copyButton().classList.contains('copied')).toBe(false)
      expect(consoleError).toHaveBeenCalled()
      consoleError.mockRestore()
    })
  })

  describe('double click', () => {
    it('selects the code', () => {
      const code = content.querySelector('pre code')!

      doubleClick(code)

      expect(getSelection()!.toString()).toContain('const first = 1')
    })

    it('ignores a double click outside of code', () => {
      const tab = content.querySelector('[role="tab"]')!

      doubleClick(tab)

      expect(getSelection()!.toString()).toBe('')
    })
  })
})
