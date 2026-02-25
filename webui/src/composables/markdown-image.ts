import type MarkdownIt from "markdown-it"
import type Token from "markdown-it/lib/token"

export function imageCaption(md: MarkdownIt) {

    md.renderer.rules.image = function (tokens: Token[], idx, options, env, self) {
        const token = tokens[idx]
        const alt = token.content

        let img = '<img '
        let title = ''
        for (const attr of token.attrs || []) {
            if (attr.length === 2) {
                if (attr[0] === 'alt' && !attr[1]) {
                    attr[1] = alt
                }
                img += ` ${attr[0]}="${attr[1]}"`
                if (attr[0] === 'title') {
                    title = attr[1]
                }
            }
        }

        img += ' />'

        if (!title) {
            return img;
        }

        return `<div class="image-with-caption">
            ${img}
            <div class="image-caption">${md.utils.escapeHtml(title)}</div>
        </div>`
    }
}