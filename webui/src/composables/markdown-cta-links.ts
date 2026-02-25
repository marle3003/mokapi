import type { Options } from "markdown-it"
import type MarkdownIt from "markdown-it"
import type Token from "markdown-it/lib/token"
import type Renderer from "markdown-it/lib/renderer"

export function MarkdownItCtaLinks(metadata: any): (md: MarkdownIt, opts: Options) => void {
    return function(md: MarkdownIt, opts: Options): void {
        var defaultRender = md.renderer.rules.text!,
            cta = /{{\s*cta-grid\s.*}}/,
            kv = /([^\s=]+="[^"]*")/g

        function getCta(token: Token) {
            if (cta.exec(token.content) === null) {
                return null
            }

            const matches = [...token.content.matchAll(kv)]
            if (matches === null || matches.length === 0) {
                return null
            }
            const data: {[key: string]: string} = {}
            for (const match of matches) {
                const kv = match[0].split('=')
                data[kv[0]!] = kv[1]!.substring(1, kv[1]!.length - 1)
            }
            return data
        }

        function text(tokens: Token[], idx: number, options: Options, env: any, slf: Renderer): string {
            if (!tokens[idx] || tokens[idx].hidden) { return ''; }

            const cta = getCta(tokens[idx])
            if (cta === null) {
                return defaultRender(tokens, idx, options, env, slf)
            }

            if (!cta.key) {
                return defaultRender(tokens, idx, options, env, slf)
            }

            const links = metadata[cta.key]
            if (!links) {
                console.error('missing metadata for cta-links: '+cta.key)
            }

            let items = ''
            for (const item of links.items) {
                items += `<div class="col">
                            <div class="card cta h-100">
                                <div class="card-inner">
                                    <a href="${item.href}" class="d-flex flex-column h-100">
                                        <div class="card-body">
                                            <div class="card-title">${item.title}</div>
                                        </div>
                                    </a>
                                </div>
                            </div>
                        </div>`
            }

            return `<div class="row row-cols-1 row-cols-md-2 g-4 card-grid mt-1">${items}</div>`;
        }

        md.renderer.rules.text = text
    }
};