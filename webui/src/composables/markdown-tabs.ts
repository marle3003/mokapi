import type { Options } from "markdown-it"
import type MarkdownIt from "markdown-it"
import type Token from "markdown-it/lib/token"
import type Renderer from "markdown-it/lib/renderer"


export function MarkdownItTabs(md: MarkdownIt, opts: Options) {
    var defaultRender = md.renderer.rules.fence!,
        unescapeAll = md.utils.unescapeAll,
        simple = /tab=([^\s]*)/,
        quote = /tab="([^"]*)/,
        counter = 0

    function getInfo(token: Token) {
        return token.info ? unescapeAll(token.info).trim() : ''
    }

    function getTabName(token: Token) {
        var info = getInfo(token)
        const r = quote.exec(info)
        if (r && r.length > 1) {
            return r.slice(1).toString()
        }

        return simple.exec(info)?.slice(1).toString() ?? ''
    }

    function getLanguage(token: Token): string {
        const info = getInfo(token)
        const lang = info.split(/\s+/)[0] || 'text'
        switch (lang) {
            case 'javascript': return 'JavaScript';
            case 'typescript': return 'TypeScript';
            case 'bash': return 'Bash';
            case 'json': return 'JSON';
            case 'yaml': return 'YAML'
            default: return capitalizeFirstLetter(lang);
        }
    }

    function capitalizeFirstLetter(val: string) {
        return val.charAt(0).toUpperCase() + val.slice(1);
    }

    function hasTab(token: Token): boolean {
        return !!getTabName(token)
    }

    function fenceGroup(tokens: Token[], idx: number, options: Options, env: any, slf: Renderer): string {
        const token = tokens[idx]
        if (!token || token.hidden) return ''

        const explicitTab = hasTab(token)

        if (!explicitTab) {
            counter++

            const lang = getLanguage(token)
            const tabName = lang || 'code'
            const tabId = `tab-${counter}-${tabName}`
            const tabPanelId = `tabPanel-${counter}-${tabName}`

            return `
            <div class="code">
                <div class="nav code-tabs" role="tablist">
                    <button class="active"
                            id="${tabId}"
                            data-bs-toggle="tab"
                            data-bs-target="#${tabPanelId}"
                            type="button"
                            role="tab"
                            aria-selected="true">
                        ${tabName}
                    </button>
                    <button type="button" class="btn btn-link control" title="Copy content" aria-label="Copy content" data-copy>
                        <span class="bi bi-copy"></span>
                    </button>
                    <div class="tabs-border"></div>
                </div>
                <div class="tab-content code">
                    <div class="tab-pane fade show active"
                        id="${tabPanelId}"
                        role="tabpanel"
                        aria-labelledby="${tabId}">
                        ${defaultRender(tokens, idx, options, env, slf)}
                    </div>
                </div>
            </div>`
        }


        counter++

        let tabs = ''
        let contents = ''

        for (let i = idx; i < tokens.length; i++) {
            const t = tokens[i]
            if (!t) {
                continue
            }
            const tabName = getTabName(t)

            if (!tabName) break

            t.info = t.info.replace(quote, '').replace(simple, '')
            t.hidden = true

            const safeName = tabName.replace('.', '-')
            const tabId = `tab-${counter}-${safeName}`
            const tabPanelId = `tabPanel-${counter}-${safeName}`
            const active = i === idx

            tabs += `
                <button class="${active ? 'active' : ''}"
                        id="${tabId}"
                        data-bs-toggle="tab"
                        data-bs-target="#${tabPanelId}"
                        type="button"
                        role="tab"
                        aria-selected="${active}">
                    ${tabName}
                </button>`

            contents += `
                <div class="tab-pane fade ${active ? 'show active' : ''}"
                    id="${tabPanelId}"
                    role="tabpanel"
                    aria-labelledby="${tabId}">
                    ${defaultRender(tokens, i, options, env, slf)}
                </div>`
        }

        return `
            <div class="code">
                <div class="nav code-tabs" role="tablist">
                    ${tabs}
                    <button type="button" class="btn btn-link control" title="Copy content" aria-label="Copy content" data-copy>
                        <span class="bi bi-copy"></span>
                    </button>
                    <div class="tabs-border"></div>
                </div>
                <div class="tab-content code">
                    ${contents}
                </div>
            </div>`
    }

    md.renderer.rules.fence = fenceGroup
}