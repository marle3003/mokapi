import { describe, expect, it } from 'vitest'
import {
  fakeRequest,
  fakeRequestInit,
  generatorError,
  isTreeNode,
  parseSchema,
  sourceNote,
  splitPropertyPath,
  type FakeMatch,
} from './faker-playground'

function match(node: string, source = 'node'): FakeMatch {
  return { propertyPath: 'a', node, tokens: [], source, value: 1 }
}

describe('faker playground', () => {
  describe('splitPropertyPath', () => {
    it('splits the dot notation into path segments', () => {
      expect(splitPropertyPath('user.email')).toEqual(['user', 'email'])
    })

    it('keeps a name without dots as one segment', () => {
      expect(splitPropertyPath('customerEmail')).toEqual(['customerEmail'])
    })

    it('trims segments and drops empty ones', () => {
      expect(splitPropertyPath(' user . .email. ')).toEqual(['user', 'email'])
    })

    it('has no segments for an empty field', () => {
      expect(splitPropertyPath('  ')).toEqual([])
    })
  })

  describe('parseSchema', () => {
    it('has no schema for an empty field', () => {
      expect(parseSchema('  \n')).toEqual({})
    })

    it('parses the schema', () => {
      expect(parseSchema('{"type": "string", "format": "email"}')).toEqual({ schema: { type: 'string', format: 'email' } })
    })

    it('reports invalid JSON', () => {
      const result = parseSchema('{"type": ')

      expect(result).toHaveProperty('error')
      expect((result as { error: string }).error).toMatch(/^Invalid JSON: /)
    })
  })

  describe('fakeRequest', () => {
    it('has neither path nor schema when both fields are empty', () => {
      expect(fakeRequest('', '')).toEqual({ request: { path: undefined, schema: undefined } })
    })

    it('takes path and schema', () => {
      expect(fakeRequest('user.email', '{"type":"string"}')).toEqual({
        request: { path: ['user', 'email'], schema: { type: 'string' } }
      })
    })

    it('reports an invalid schema instead of a request', () => {
      expect(fakeRequest('user', '{')).toHaveProperty('error')
    })
  })

  it('sends the request as JSON by POST', () => {
    const init = fakeRequestInit({ path: ['a'] })

    expect(init.method).toBe('POST')
    expect(init.body).toBe('{"path":["a"]}')
    expect(init.headers).toMatchObject({ 'Content-Type': 'application/json' })
  })

  it('selects only nodes of the tree', () => {
    expect(isTreeNode(match('/person/email'))).toBe(true)
    expect(isTreeNode(match('fakeBySchema'))).toBe(false)
    expect(isTreeNode(match('-', 'enum'))).toBe(false)
  })

  it('tells why a value is not from a node', () => {
    expect(sourceNote(match('/email'))).toBe('')
    expect(sourceNote(match('-', 'enum'))).toBe('from enum')
  })

  describe('generatorError', () => {
    it('shows the message of the generator', () => {
      expect(generatorError('Error: Bad Request: invalid schema: unknown type')).toBe('invalid schema: unknown type')
    })

    it('keeps an error that is not an HTTP error', () => {
      expect(generatorError('TypeError: Failed to fetch')).toBe('TypeError: Failed to fetch')
    })
  })
})
