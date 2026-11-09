import type { MarkdownIt } from "markdown-it"
import { useRouter } from "@/router"
import { toWebsiteUrl } from "./website-url"

export function MarkdownItLinks(md: MarkdownIt) {

    function replace(s: string): string {
        s = s.replace('.md', '')
        return new URL(s, document.location.href).toString()
    }

    // not useRoute(): Vue may evaluate a computed that renders Markdown outside of a component
    const router = useRouter()

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
                    token.attrSet('href', router.currentRoute.value.path + href)
                } else if (href.startsWith('/')) {
                    token.attrSet('href', toWebsiteUrl(href))
                }
            }
        }
    })
}