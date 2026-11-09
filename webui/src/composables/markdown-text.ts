import { load as loadYaml } from 'js-yaml'

const metadataRegex = /^---([\s\S]*?)---/

export function parseMetadata(data: string): DocMeta {
    if (!data) {
        return {} as DocMeta
    }

    const metadataMatch = data.match(metadataRegex)
    if (!metadataMatch) {
        return {} as DocMeta
    }

    return loadYaml(metadataMatch[1]!) as DocMeta
}

export function removeMetadata(markdown: string): string {
    return markdown.replace(metadataRegex, '')
}

export function replaceAppVersion(markdown: string, version: string): string {
    return markdown.replaceAll(/__APP_VERSION__/g, version)
}

// The text the search looks in: the Markdown without front matter, as authored.
export function toSearchText(markdown: string, version: string): string {
    return removeMetadata(replaceAppVersion(markdown, version))
}
