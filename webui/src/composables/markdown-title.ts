import type { Options } from "markdown-it";
import type MarkdownIt from "markdown-it"
import type Token from "markdown-it/lib/token.mjs";

export function MarkdownItTitle(meta: DocMeta): (md: MarkdownIt, opts: Options) => void {

    return function(md: MarkdownIt): void {

        md.renderer.rules.heading_open = (tokens, idx, options, env, self) => {
            const token = tokens[idx] as Token;
            
            let heading = self.renderToken(tokens, idx, options);

            if (token.tag !== 'h1') {
                return heading;
            }

            let html = '';
            if (meta) {
                if (meta.tags) {
                    html += `<div class="mb-2">`
                    for (const tag of meta.tags) {
                        html += `<span class="badge text-bg-primary">${md.utils.escapeHtml(tag)}</span>`
                    }
                    html += `</div>`
                }
            }

            return html + heading;
        }

        md.renderer.rules.heading_close = (tokens, idx, options, env, self) => {
            const token = tokens[idx] as Token;
            
            let html = self.renderToken(tokens, idx, options);

            if (token.tag !== 'h1') {
                return html;
            }

            if (meta) {
                if (meta.subtitle) {
                    html += `<p class="subtitle">${meta.subtitle}</p>`;
                }
            }

            return html;
        };

    }
}