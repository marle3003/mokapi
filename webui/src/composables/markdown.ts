import MarkdownItHighlightjs from 'markdown-it-highlightjs';
import MarkdownIt from 'markdown-it';
import { MarkdownItTabs } from '@/composables/markdown-tabs';
import { MarkdownItBox } from '@/composables/markdown-box';
import { MarkdownItLinks } from '@/composables/mardown-links';
import { MarkdownItCard } from '@/composables/markdown-card';
import { MarkdownItCarousel } from './markdown-carousel';
import yaml from 'js-yaml'

const images =  import.meta.glob('/src/assets/docs/**/*.png', {as: 'url', eager: true})
const metadataRegex = /^---([\s\S]*?)---/;

export function useMarkdown(content: string | undefined): {content: string | undefined, metadata?: DocMeta} {
    if (!content) {
        return { content }
    }
    try {
        const metadata = parseMetadata(content)
        content = replaceImageUrls(content).replace(metadataRegex, '')

        content = content.replaceAll(/__APP_VERSION__/g, APP_VERSION)

        if (content) {
            content = new MarkdownIt()
            .use(MarkdownItHighlightjs)
            .use(MarkdownItTabs)
            .use(MarkdownItBox)
            .use(MarkdownItLinks)
            .use(MarkdownItCarousel(metadata))
            .use(MarkdownItCard(metadata))
            .set({html: true})
            .render(content)
        }

        return {content, metadata}
    } catch (e) {
        console.error('invalid markdown: '+content)
        return { content: '' }
    }
}

function replaceImageUrls(data: string) {
    const regex = /<img([^>]*)src="(?:[^"\/]*)([^"]+)"/gi
    let m
    do {
        m = regex.exec(data)
        if (m) {
            const path = `/src/assets${m[2]}`
            let imageUrl = images[path]
            if (imageUrl) {
                data = data.replace(m[0], `<img${m[1]} src="${imageUrl}"`)
            } else {
                imageUrl = transformPath(m[2])
                data = data.replace(m[0], `<img${m[1]} src="${imageUrl}"`)
            }
        }
    } while(m)
    return data
}

export function parseMetadata(data: string): DocMeta {
    if (!data) {
        return {} as DocMeta
    }

    const metadataMatch = data.match(metadataRegex)
    if (!metadataMatch) {
        return {} as DocMeta
    }
  
    return yaml.load(metadataMatch[1]) as DocMeta
}

export function transformPath(path: string): string {
    let base = document.querySelector('base')?.href
    if (base) {
        base = base.substring(0, base.length - 1)
        path = base + path
    }
    return path
}

export function parseMarkdown(markdown: string): { meta: DocMeta, content: string } {
    const metadata = parseMetadata(markdown)
    let content = replaceImageUrls(markdown).replace(metadataRegex, '')

    content = content.replaceAll(/__APP_VERSION__/g, APP_VERSION)
    return { meta: metadata, content: content }
}