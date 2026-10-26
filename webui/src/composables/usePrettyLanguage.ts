import { MIMEType } from 'whatwg-mimetype'
import XmlFormatter from 'xml-formatter'

export function usePrettyLanguage() {
    function formatLanguage(s: string, contentType: string, truncated?: boolean): string {
        if (!s){
            return s
        }
        const mimeType = new MIMEType(contentType)
        switch (mimeType.subtype){
            case 'json':
                if (!truncated) {
                    try {
                    return JSON.stringify(JSON.parse(s), null, 2)
                    } catch {
                        console.error("unable to parse json: "+s)
                        return s
                    }
                }
                return prettyJson(s)
            case 'xml':
            case 'rss+xml':
                return XmlFormatter(s, {collapseContent: true})
        }

        return s
    }

    function getLanguage(contentType: string) {
        const mimeType = new MIMEType(contentType)
        switch (mimeType.subtype){
            case 'xml':
            case 'problem+xml':
            case 'rss+xml':
                return 'xml'
            case 'plain':
                return 'text'
            case 'javascript':
                return 'javascript'
            case 'typescript':
                return 'javascript'
            case 'yaml':
                return 'yaml'
            case 'json':
            case 'problem+json':
                return 'json'
            case 'ldif':
                return 'ldif'
        }

        switch (contentType) {
            case 'avro/binary':
            case 'application/avro':
            case 'application/octet-stream':
                // display avro content as JSON
                return 'hex'
            default:
                return mimeType.subtype
        }
    }

    function getContentType(url: string): string | null {
        const index = url.lastIndexOf('.')
        if (index < 0) {
            return null
        }
        const ext = url.substring(index)
        switch (ext) {
            case '.json':
                return 'application/json'
            case '.yaml':
            case '.yml':
                return 'application/yaml'
            case '.js':
                return 'text/javascript'
            case '.ts':
                return 'text/typescript'
            case '.ldif':
                return 'application/ldif'
            default:
                return null
        }
    }

    function formatSchema(s: Schema | SchemaFormat | undefined): string {
        if (!s) {
            return ''
        }
        const sf = s as SchemaFormat
        if (sf && sf.schema) {
            return formatLanguage(JSON.stringify(sf.schema), 'application/json')
        }
        return formatLanguage(JSON.stringify(s), 'application/json')
    }

    return { formatLanguage, getLanguage, getContentType, formatSchema }
}

function toBinary(s: string) {
    const codeUnits = new Uint16Array(s.length);
    for (let i = 0; i < codeUnits.length; i++) {
      codeUnits[i] = s.charCodeAt(i);
    }
    return btoa(String.fromCharCode(...new Uint8Array(codeUnits.buffer)));
  }

  /**
 * Pretty-prints JSON without parsing it.
 * Works on truncated input (e.g. a body preview): the text is only
 * re-indented, never validated. Whitespace outside of strings is dropped.
 */
export function prettyJson(s: string, indent = 2): string {
  const pad = ' '.repeat(indent)
  const out: string[] = []
  let level = 0
  let inString = false
  let escaped = false

  const newline = () => out.push('\n', pad.repeat(level))

  for (let i = 0; i < s.length; i++) {
    const c = s[i]

    if (inString) {
      out.push(c)
      if (escaped) escaped = false
      else if (c === '\\') escaped = true
      else if (c === '"') inString = false
      continue
    }

    switch (c) {
      case '"':
        inString = true
        out.push(c)
        break
      case '{':
      case '[': {
        // keep empty {} and [] on one line
        let j = i + 1
        while (j < s.length && /\s/.test(s[j])) j++
        if (s[j] === (c === '{' ? '}' : ']')) {
          out.push(c, s[j])
          i = j
          break
        }
        out.push(c)
        level++
        newline()
        break
      }
      case '}':
      case ']':
        level = Math.max(0, level - 1)
        newline()
        out.push(c)
        break
      case ',':
        out.push(c)
        newline()
        break
      case ':':
        out.push(': ')
        break
      case ' ':
      case '\t':
      case '\n':
      case '\r':
        break // drop original formatting
      default:
        out.push(c)
    }
  }
  return out.join('')
}