export function selectCode(event: MouseEvent) {
  const target = event.target as HTMLElement
  if (!target.closest('pre')) {
    return
  }
  const range = document.createRange()
  range.selectNodeContents(target)
  const selection = getSelection()
  selection?.removeAllRanges()
  selection?.addRange(range)
}

export async function copyCode(event: MouseEvent) {
  const target = event.target as HTMLElement
  const button = target.closest('[data-copy]')
  if (!button) {
    return
  }

  const codeBlock = button.closest('.code')
  if (!codeBlock) {
    return
  }
  const activePane = codeBlock.querySelector('.tab-pane.active')
  if (!activePane) {
    return
  }
  const codeElement = activePane.querySelector('pre code')
  if (!codeElement) {
    return
  }

  const text = codeElement.textContent ?? ''

  try {
    await navigator.clipboard.writeText(text)
    button.classList.add('copied')
    setTimeout(() => button.classList.remove('copied'), 1500)
  } catch (err) {
    console.error('Failed to copy code:', err)
  }
}
