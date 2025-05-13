import type { Options } from "markdown-it"
import type MarkdownIt from "markdown-it"
import type Token from "markdown-it/lib/token"
import type Renderer from "markdown-it/lib/renderer"

export function MarkdownItCarousel(metadata: any): (md: MarkdownIt, opts: Options) => void {
    return function(md: MarkdownIt, opts: Options): void {
        var defaultRender = md.renderer.rules.text!,
            carousel = /{{\s*carousel\s.*}}/,
            kv = /([^\s=]+="[^"]*")/g

        function getCarousel(token: Token) {
            if (carousel.exec(token.content) === null) {
                return null
            }

            const matches = [...token.content.matchAll(kv)]
            if (matches === null || matches.length === 0) {
                return null
            }
            const data: { [key: string]: string } = {}
            for (const match of matches) {
                const kv = match[0].split('=')
                data[kv[0]] = kv[1].substring(1, kv[1].length - 1)
            }

            return data
        }

        function text(tokens: Token[], idx: number, options: Options, env: any, slf: Renderer): string {
            if (tokens[idx].hidden) { return ''; }

            const carousel = getCarousel(tokens[idx])
            if (carousel === null) {
                return defaultRender(tokens, idx, options, env, slf)
            }

            if (!carousel.key) {
                return defaultRender(tokens, idx, options, env, slf) 
            }

            const items = metadata[carousel.key]
            if (!items) {
                console.error('missing metadata for carousel: ' + carousel.key)
            }

            let indicators = ''
            let images = ''
            for (const [index, item] of items.entries()) {
                indicators += `<button type="button" data-bs-target="#carousel-${carousel.key}" data-bs-slide-to="${index}" ${index === 0 ? 'class="active" aria-current="true"' : ''} aria-label="Slide ${index+1}"></button>`
                images += `<div class="carousel-item ${index === 0 ? 'active' : ''}">
                               <img src="${item.img}" class="d-block w-100" alt="${item.alt}">
                               <div class="carousel-caption d-none d-md-block">
                                   <h6>${item.title}</h6>
                                   <p>${item.description}</p>
                               </div>
                           </div>`
            }

            const content = `
<div id="carousel-${carousel.key}" class="carousel slide">
  <div class="carousel-indicators">
    ${indicators}
  </div>
  <div class="carousel-inner">
    ${images}
  </div>
  <button class="carousel-control-prev" type="button" data-bs-target="#carousel-${carousel.key}" data-bs-slide="prev">
    <span class="carousel-control-prev-icon" aria-hidden="true"></span>
    <span class="visually-hidden">Previous</span>
  </button>
  <button class="carousel-control-next" type="button" data-bs-target="#carousel-${carousel.key}" data-bs-slide="next">
    <span class="carousel-control-next-icon" aria-hidden="true"></span>
    <span class="visually-hidden">Next</span>
  </button>
</div>`;

            return content
        }

        md.renderer.rules.text = text
    };
};