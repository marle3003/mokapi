import type { Options } from "markdown-it"
import type MarkdownIt from "markdown-it"
import type Token from "markdown-it/lib/token"
import type Renderer from "markdown-it/lib/renderer"

export function MarkdownItCard(metadata: any): (md: MarkdownIt, opts: Options) => void {
    return function(md: MarkdownIt, opts: Options): void {
        var defaultRender = md.renderer.rules.text!,
            unescapeAll = md.utils.unescapeAll,
            card = /{{\s*card-grid\s.*}}/,
            kv = /([^\s=]+="[^"]*")/g

        function getCard(token: Token) {
            if (card.exec(token.content)?.slice(1) === null) {
                return null
            }
            const matches = [...token.content.matchAll(kv)]
            if (matches === null || matches.length === 0) {
                return null
            }
            const data: {[key: string]: string} = {}
            for (const match of matches) {
                const kv = match[0].split('=')
                data[kv[0]] = kv[1].substring(1, kv[1].length - 1)
            }
            return data
        }

        function fenceGroup(tokens: Token[], idx: number, options: Options, env: any, slf: Renderer): string {
            if (tokens[idx].hidden) { return ''; }

            const card = getCard(tokens[idx])
            if (card === null) {
                return defaultRender(tokens, idx, options, env, slf)
            }

            if (!card.key) {
                return defaultRender(tokens, idx, options, env, slf)
            }

            const cards = metadata[card.key]
            if (!cards) {
                console.error('missing metadata for cards: '+card.key)
            }

            let items = ''
            for (const item of cards.items) {
                items += `<div class="col">
                            <a class="card h-100" href="${item.href}">
                                <div class="card-body">
                                    <div class="card-title">${item.title}</div>
                                    <div class="card-text">${item.description}</div>
                                </div>
                            </a>
                        </div>`
            }

            return `<div class="row row-cols-1 row-cols-md-2 g-4 card-grid mt-1">${items}</div>`;
        }

        md.renderer.rules.text = fenceGroup
    }
};