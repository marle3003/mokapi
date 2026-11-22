import { describe, expect, it } from 'vitest'
import { formatKey, messageKey } from './kafka'

describe('formatKey', () => {
  it.each([
    [null, 'null'],
    [undefined, 'null'],
    ['"abc"', 'abc'],
    ['null', 'null'],
    ['{"a":1}', '{"a":1}'],
    ['42', '42'],
    ['true', 'true'],
    ['plain text', 'plain text'],
  ])('%s is shown as %s', (json, expected) => {
    expect(formatKey(json)).toBe(expected)
  })
})

describe('messageKey', () => {
  it('shows a text key without JSON quotes', () => {
    expect(messageKey({ value: '"GGOEWXXX0827"' })).toBe('GGOEWXXX0827')
    expect(messageKey({ value: 'GGOEWXXX0827' })).toBe('GGOEWXXX0827')
  })

  it('decodes a binary key as text', () => {
    expect(messageKey({ binary: btoa('order-1') })).toBe('order-1')
  })

  it('prefers the value over the binary form', () => {
    expect(messageKey({ value: 'text', binary: btoa('binary') })).toBe('text')
  })

  it('is empty without a key', () => {
    expect(messageKey(null)).toBe('')
    expect(messageKey(undefined)).toBe('')
    expect(messageKey({})).toBe('')
  })
})
