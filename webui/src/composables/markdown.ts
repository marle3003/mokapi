import MarkdownIt from 'markdown-it';
import { MarkdownItTabs } from '@/composables/markdown-tabs';
import { MarkdownItBox } from '@/composables/markdown-box';
import { MarkdownItLinks } from '@/composables/markdown-links';
import { MarkdownItCard } from '@/composables/markdown-card';
import { MarkdownItCarousel } from './markdown-carousel';
import { MarkdownItBlockquote } from './markdown-blockquote';
import { MarkdownItTabContent } from './markdown-tab-content';
import { MarkdownItTitle } from './markdown-title';
import { MarkdownItTable } from './markdown-table';
import { imageCaption } from './markdown-image';
import { MarkdownItCtaLinks } from './markdown-cta-links';
import hljs from '@/plugins/highlight'
import { images } from '#docs-images'
import { sitePath } from './site-base'
import { parseMetadata, removeMetadata, replaceAppVersion } from './markdown-text'

export function useMarkdown(content: string | undefined, linkify = false): {content: string | undefined, metadata?: DocMeta} {
    if (!content) {
        return { content }
    }
    try {
        const metadata = parseMetadata(content)

        content = removeMetadata(replaceAppVersion(content, APP_VERSION))

        if (content) {
            const md = new MarkdownIt({
                    linkify: linkify,
                    highlight: (str, lang): string => {
                        if (lang === 'typescript') {
                            lang = 'javascript'
                        }
                        if (lang && hljs.getLanguage(lang)) {
                            return `<pre class="hljs"><code>${hljs.highlight(str, { language: lang }).value}</code></pre>`
                        }
                        return `<pre class="hljs"><code>${md.utils.escapeHtml(str)}</code></pre>`
                    }
                })
                .use(MarkdownItTitle(metadata))
                .use(MarkdownItBlockquote)
                .use(MarkdownItTabs)
                .use(MarkdownItTabContent)
                .use(MarkdownItBox)
                .use(MarkdownItLinks)
                .use(MarkdownItTable)
                .use(MarkdownItCarousel(metadata))
                .use(MarkdownItCard(metadata))
                .use(MarkdownItCtaLinks(metadata))
                .use(imageCaption)
                .set({html: true})
            content = md.render(content)
        }

        content = replaceImageUrls(content) 

        return {content, metadata}
    } catch (e) {
        console.error('invalid markdown: '+e)
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
                imageUrl = sitePath(m[2]!)
                data = data.replace(m[0], `<img${m[1]} src="${imageUrl}"`)
            }
        }
    } while(m)
    return data
}
