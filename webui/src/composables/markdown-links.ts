import type { MarkdownIt } from "markdown-it"
import { useRoute } from "vue-router"

export function MarkdownItLinks(md: MarkdownIt) {

    function replace(s: string): string {
        s = s.replace('.md', '')
        return new URL(s, document.location.href).toString()
    }

    const route = useRoute()

    md.core.ruler.after('inline', 'link', (state) => {
        for (const blockToken of state.tokens) {
            if (blockToken.type !== 'inline' || !blockToken.children) continue

            for (const token of blockToken.children) {
                if (token.type !== 'link_open') continue

                const href = token.attrGet('href')
                if (typeof href !== 'string') continue

                if (href.endsWith('.md') || href.includes('.md#')) {
                    token.attrSet('href', replace(href))
                } else if (href.startsWith('#')) {
                    token.attrSet('href', route.path + href)
                }
            }
        }
    })
}