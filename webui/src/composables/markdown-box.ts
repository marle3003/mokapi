import type { Options } from "markdown-it"
import type MarkdownIt from "markdown-it"
import type Token from "markdown-it/lib/token"
import type Renderer from "markdown-it/lib/renderer"

export function MarkdownItBox(md: MarkdownIt, opts: Options) {
    var defaultRender = md.renderer.rules.fence!,
        unescapeAll = md.utils.unescapeAll,
        re = /box=(\w*)/

    function getInfo(token: Token) {
        return token.info ? unescapeAll(token.info).trim() : ''
    }

    function getAlertName(token: Token) {
        var info = getInfo(token) 
        return re.exec(info)?.slice(1)
    }

    function fenceGroup(tokens: Token[], idx: number, options: Options, env: any, slf: Renderer): string {
        if (tokens[idx].hidden) { return ''; }

        const name = getAlertName(tokens[idx])
        if (name == null) {
            return defaultRender(tokens, idx, options, env, slf)
        }
        
        var alert = ''
        for (let i = idx; i < tokens.length; i++) {
            const token = tokens[i]
            const name = getAlertName(token)
            if (name == null) { 
                break;
            }

            token.info = token.info.replace(re, '')
            token.hidden = true

            alert += `<div class="box ${name}" role="alert">
                     <p class="box-heading">${name}</p>
                     <p class="box-body">${token.content}</p>
                    </div>`
        }

        return alert
    
    }

    md.renderer.rules.fence = fenceGroup
};