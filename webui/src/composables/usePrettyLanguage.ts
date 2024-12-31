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

        switch (contentType) {
            case 'avro/binary':
                try{ 
                    return JSON.stringify(JSON.parse(s), null, 2)
                }catch {
                    console.error("unable to parse json: "+s)
                    return s
                }
        }

        return s
    }

    function getLanguage(contentType: string) {
        const mimeType = new MIMEType(contentType)
        switch (mimeType.subtype){
            case 'rss+xml':
                return 'xml'
            case 'plain':
                return 'text'
            case 'javascript':
                return 'javascript'
            case 'typescript':
                return 'javascript'
        }

        switch (contentType) {
            case 'avro/binary':
                // display avro content as JSON
                return "javascript"
            default:
                return 'text'
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

    function formatSchema(s: Schema | undefined): string {
        if (!s) {
            return ''
        }
        return formatLanguage(JSON.stringify(s), 'application/json')
    }

    return { formatLanguage, getLanguage, getContentType, formatSchema }
}