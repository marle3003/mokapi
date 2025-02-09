import type { Options } from "markdown-it"
import type MarkdownIt from "markdown-it"
import type Token from "markdown-it/lib/token"
import type Renderer from "markdown-it/lib/renderer"

export function MarkdownItBox(md: MarkdownIt, opts: Options) {
    var defaultRender = md.renderer.rules.fence!,
        unescapeAll = md.utils.unescapeAll,
        boxExpr = /box=(\w*)/,
        noTitleExpre = /noTitle/,
        title = /title=([^\s]*)/,
        titleQuote = /title=\"([^"]*)\"/,
        url = /url=\[([^\]]*)\]\(([^\)]*)\)/

    function getInfo(token: Token) {
        return token.info ? unescapeAll(token.info).trim() : ''
    }

    function getAlertName(token: Token) {
        var info = getInfo(token) 
        return boxExpr.exec(info)?.slice(1)[0]
    }

    function getUrl(token: Token) {
        var info = getInfo(token) 
        
        const v = url.exec(info)
        if (!v) {
            return ''
        }

        var u = v[2]
        if (u.endsWith('.md')) {
            u = new URL(u.replace('.md', ''), document.location.href).toString()
        }

        return `<p class="mt-1 pb-2 ps-2"><a href="${u}">${v[1]}</a></p>`
    }

    function showTitle(token: Token): boolean {
        var info = getInfo(token)
        return noTitleExpre.exec(info) == null
    }

    function getTitle(token: Token) {
        var info = getInfo(token) 

        const r = titleQuote.exec(info)
        if (r && r.length > 1) {
            return r.slice(1)[0]
        }

        return title.exec(info)?.slice(1)[0]
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

            token.info = token.info.replace(boxExpr, '')
            token.hidden = true

            const url = getUrl(token)

            if (showTitle(tokens[i])) {
                let title = getTitle(token)
                let heading = ''
                if (!title) {
                    title = name
                } else {
                    heading = 'box-custom-heading'
                }

                let icon = ''
                switch (name) {
                    case 'tip':
                        icon = '<i class="bi bi-lightbulb me-1"></i>'
                        break
                    case 'warning':
                        icon = '<i class="bi bi-exclamation-triangle me-2"></i>'
                        break
                    case 'info':
                        icon = '<i class="bi bi-info-circle me-1"></i>'
                        break
                }

                alert += `<div class="box ${name}" role="alert">
                        <p class="box-heading ${heading}">${icon}${title}</p>
                        <p class="box-body">${token.content}</p>
                        ${url}
                        </div>`
            }
            else {
                alert += `<div class="box ${name} no-title" role="alert">
                        <p class="box-body">${token.content}</p>
                        ${url}
                        </div>`
            }
        }

        return alert
    
    }

    md.renderer.rules.fence = fenceGroup
};