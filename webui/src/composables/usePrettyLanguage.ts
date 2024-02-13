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
                    console.log("unable to parse json: "+s)
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
            case 'rss+xml':
                return 'xml'
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
            default:
                return null
        }
    }

    return { formatLanguage, getLanguage, getContentType}
}