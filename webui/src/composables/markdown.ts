import MarkdownItHighlightjs from 'markdown-it-highlightjs';
import MarkdownIt from 'markdown-it';
import { MarkdownItTabs } from '@/composables/markdown-tabs';
import { MarkdownItBox } from '@/composables/markdown-box';
import { MarkdownItLinks } from '@/composables/mardown-links'

const images =  import.meta.glob('/src/assets/docs/**/*.png', {as: 'url', eager: true})
const metadataRegex = /^---([\s\S]*?)---/;

export function useMarkdown(content: string | undefined) {
    if (!content) {
        return {content, metadata: {}}
    }
    const metadata = parseMetadata(content)
    content = replaceImageUrls(content).replace(metadataRegex, '')

    if (content) {
        content = new MarkdownIt()
          .use(MarkdownItHighlightjs)
          .use(MarkdownItTabs)
          .use(MarkdownItBox)
          .use(MarkdownItLinks)
          .set({html: true})
          .render(content)
    }

    return {content, metadata}
}

function replaceImageUrls(data: string) {
    const regex = /<img([^>]*)src="(?:[^"\/]*)([^"]+)"/gi
    let m
    do {
        m = regex.exec(data)
        if (m) {
        const path = `/src/assets${m[2]}`
        const imageUrl = images[path]
        if (imageUrl) {
            data = data.replace(m[0], `<img${m[1]} src="${imageUrl}"`)
        }
        }
    } while(m)
    return data
}

export function parseMetadata(data: string) {
    const metadataMatch = data.match(metadataRegex);
  
    if (!metadataMatch) {
      return {};
    }
  
    const metadataLines = metadataMatch[1].split("\n")
  
    // Use reduce to accumulate the metadata as an object
    const metadata = metadataLines.reduce((acc: any, line) => {
        const i = line.indexOf(':');
        const splits = [line.slice(0,i), line.slice(i+1)];

        const [key, value] = splits.map(part => part.trim())
        if(key) {
            acc[key] = value
        }
        return acc;
    }, {});
  
    return metadata;
  };