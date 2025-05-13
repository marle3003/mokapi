import type { Options } from "markdown-it"
import type MarkdownIt from "markdown-it"
import type Token from "markdown-it/lib/token"
import type Renderer from "markdown-it/lib/renderer"

export function MarkdownItCarousel(md: MarkdownIt, opts: Options) {
    var defaultRender = md.renderer.rules.fence!,
        unescapeAll = md.utils.unescapeAll,
        carousel = /carousel=([^\s]*)/;

    function getInfo(token: Token) {
        return token.info ? unescapeAll(token.info).trim() : ''
    }

    function getCarousel(token: Token) {
        var info = getInfo(token) 
        return carousel.exec(info)?.slice(1).toString() ?? ''
    }

    function fenceGroup(tokens: Token[], idx: number, options: Options, env: any, slf: Renderer): string {
        if (tokens[idx].hidden) { return ''; }

        const found = getCarousel(tokens[idx])
        if (found === null || found === '') {
            return defaultRender(tokens, idx, options, env, slf)
        }

        const imgs = tokens[idx].content.split('\n').filter(x => x !== '')
        let items = ''
        for (const [index, img] of imgs.entries()) {
            items += `<div class="carousel-item ${ index === 0?'active': ''}">
                      <img src="${img}" class="d-block w-100" alt="...">
                      </div>`
        }

        return `
<div id="carousel-${idx}" class="carousel slide">
  <div class="carousel-inner">
    ${items}
  </div>
  <button class="carousel-control-prev" type="button" data-bs-target="#carousel-${idx}" data-bs-slide="prev">
    <span class="carousel-control-prev-icon" aria-hidden="true"></span>
    <span class="visually-hidden">Previous</span>
  </button>
  <button class="carousel-control-next" type="button" data-bs-target="#carousel-${idx}" data-bs-slide="next">
    <span class="carousel-control-next-icon" aria-hidden="true"></span>
    <span class="visually-hidden">Next</span>
  </button>
</div>`
    }

    md.renderer.rules.fence = fenceGroup
};