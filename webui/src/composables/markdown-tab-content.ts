import type MarkdownIt from "markdown-it"
import type { Options } from "markdown-it"
import container from "markdown-it-container"
import type { Renderer } from "markdown-it/dist/index.cjs.js"
import Token from 'markdown-it/lib/token'

export function MarkdownItTabContent(md: MarkdownIt, _opts: Options) {

    md.core.ruler.push('test', function (state) {
        let hide = false
        let inTab = false

        for (let i = 0; i < state.tokens.length; i++) {
            const token = state.tokens[i]!

            switch (token.type) {
                case 'container_tabs_open':
                    hide = true
                    continue
                case 'container_tabs_close':
                    hide = false
                    inTab = false
                    continue
                case 'inline':
                    if (token.content.startsWith('@tab')) {
                        const match = token.content.trim().match(/^@tab "(.*)"$/);
                        if (!match || !match[1]) {
                            console.error('invalid tab format: ' + token.content)
                            return
                        }
                        const label = match[1];
                        const safe = label.replace(/[^\w-]/g, "-");
                        const tabIndex = tabs.length
                        const tabId = `tab-${tabIndex}-${safe}`
                        const panelId = `tabPanel-${tabIndex}-${safe}`

                        const button = `
    <button class="${tabIndex === 0 ? "active" : ""}"
            id="${tabId}"
            data-bs-toggle="tab"
            data-bs-target="#${panelId}"
            type="button"
            role="tab"
            aria-selected="${tabIndex === 0}">
      ${label}
    </button>`

                        tabs.push({
                            index: tabIndex,
                            panelId: panelId,
                            button: button,
                            panel: '',
                            tokens: []
                        })
                        inTab = true
                        hideToken(token);
                        i++; // skip closing paragraph
                        continue
                    }
            }

            if (hide) {
                hideToken(token)
                if (inTab) {
                    tabs[tabs.length - 1]!.tokens.push(token)
                }
            }
        }
    })

    function hideToken(token: Token) {
        token.hidden = true
        if (token.children) {
            for (const child of token.children) {
                hideToken(child)
            }
        }
    }

    interface TabsState {
        index: number
        panelId: string
        button: string
        panel: string
        tokens: Token[]
    }

    const tabs: TabsState[] = []

    md.use(container, 'tabs', {
        validate: (params: any) => params.trim() === 'tabs',
        render: (tokens: Token[], idx: number, options: Options, env: any, slf: Renderer) => {
            const token = tokens[idx];
            if (token?.nesting === 1) {
                return ''
            } else {
                return `
<div class="tabs">
    <div class="nav nav-tabs" role="tablist">
        ${tabs.map(x => x.button).join('')}
        <div class="tabs-border"></div>
    </div>
    <div class="tab-content">
        ${tabs.map(x => {
                    const tokens = x.tokens.map(t => clone(t))
                    const content = md.renderer.render(tokens, options, env)
                    return `
            <div class="tab-pane ${x.index === 0 ? "show active" : ""}"
                id="${x.panelId}">
                ${content}
            </div>
            `;
                }).join('')}
     </div>
</div>`
            }
        }
    });

    function clone(token: Token): Token {
        const t = Object.create(Object.getPrototypeOf(token))
        Object.assign(t, token)

        t.hidden = false
        if(token.attrs) {
            t.attrs = token.attrs.map((a: any) => [...a])
        }

        if (token.children) {
            t.children = token.children.map(clone)
        }
        return t
    }
}
