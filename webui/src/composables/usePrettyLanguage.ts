import MIMEType from 'whatwg-mimetype'
import XmlFormatter from 'xml-formatter'

export function usePrettyLanguage() {
    function format(s: string, contentType: string): string {
        if (!s){
            return s
        }
        const mimeType = new MIMEType(contentType)
        switch (mimeType.subtype){
            case 'json':
                return JSON.stringify(JSON.parse(s), null, 2)
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

    return {format, getLanguage}
}