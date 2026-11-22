import { describe, expect, it } from 'vitest'
import { messageValue } from './message-value'

function config(contentType: string, format?: string): BaseMessage {
  return { name: 'm', title: '', summary: '', description: '', contentType, payload: { format } as SchemaFormat }
}

const avro = config('avro/binary', 'application/vnd.apache.avro;version=1.9.0')

describe('messageValue', () => {
  it('formats the value for the content type of the message', () => {
    const value = messageValue(config('application/json'), { value: '{"id":1}' })

    expect(value).toEqual({
      source: { preview: { content: '{\n  "id": 1\n}', contentType: 'application/json', contentTypeTitle: 'application/json', description: undefined } },
      contentType: 'application/json',
      contentTypeTitle: 'application/json',
    })
  })

  it('keeps the value of another content type as it is', () => {
    const value = messageValue(config('text/plain'), { value: 'hello' })

    expect(value.source.preview?.content).toBe('hello')
    expect(value.contentType).toBe('text/plain')
  })

  it('shows an Avro message as JSON and keeps the binary form', () => {
    const value = messageValue(avro, { value: '{"Name":"Carol"}', binary: btoa('abc') })

    expect(value.contentType).toBe('application/json')
    expect(value.contentTypeTitle).toBe('avro/binary')
    expect(value.source.preview?.description).toBe('Avro content in JSON format')
    expect(value.source.binary).toEqual({ content: 'abc', contentType: 'avro/binary' })
  })

  it('does not treat a binary content type without Avro format as JSON', () => {
    const value = messageValue(config('application/octet-stream'), { value: 'x', binary: btoa('abc') })

    expect(value.contentType).toBe('application/octet-stream')
    expect(value.source.binary).toEqual({ content: 'abc', contentType: 'application/octet-stream' })
  })

  it('shows binary content without a preview when the message has no text value', () => {
    const value = messageValue(avro, { binary: btoa('abc') })

    expect(value.source.preview).toBeUndefined()
    expect(value.source.binary).toEqual({ content: 'abc', contentType: 'avro/binary' })
  })

  it('ignores binary content of a text content type', () => {
    const value = messageValue(config('application/json'), { value: '{}', binary: btoa('abc') })

    expect(value.source.binary).toBeUndefined()
  })
})
