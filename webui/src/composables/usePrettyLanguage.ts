import MIMEType from 'whatwg-mimetype'
import XmlFormatter from 'xml-formatter'

export function usePrettyLanguage() {
    function formatLanguage(s: string, contentType: string): string {
        if (!s){
            return s
        }
        const mimeType = new MIMEType(contentType)
        switch (mimeType.subtype){
            case 'json':
                try{ 
                    return JSON.stringify(JSON.parse(s), null, 2)
                }catch {
                    console.error("unable to parse json: "+s)
                    return s
                }
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