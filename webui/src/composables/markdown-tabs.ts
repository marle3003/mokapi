import type { Options } from "markdown-it";
import type MarkdownIt from "markdown-it";
import type Token from "markdown-it/lib/token";
import type Renderer from "markdown-it/lib/renderer";


export function MarkdownItTabs(md: MarkdownIt, opts: Options) {
    var defaultRender = md.renderer.rules.fence!,
        unescapeAll = md.utils.unescapeAll,
        re = /tab=(\w*)/,
        counter = 0

    function getInfo(token: Token) {
        return token.info ? unescapeAll(token.info).trim() : '';
    }

    function getTabName(token: Token) {
        var info = getInfo(token)   
        return re.exec(info)?.slice(1)
    }

    function fenceGroup(tokens: Token[], idx: number, options: Options, env: any, slf: Renderer): string {
        if (tokens[idx].hidden) { return ''; }

        const tabName = getTabName(tokens[idx]);
        if (tabName == null) {
            return defaultRender(tokens, idx, options, env, slf);
        }
        counter++
        
        var tabs = '', contents = '';
        for (let i = idx; i < tokens.length; i++) {
            const token = tokens[i];
            const tabName = getTabName(token);
            if (tabName == null) { 
                break;
            }

            token.info = token.info.replace(re, '');
            token.hidden = true;

            const tabId = `tab-${counter}-${tabName}`
            const tabPanelId = `tabPanel-${counter}-${tabName}`
            const checked = i - idx > 0 ? '' : ' checked';

            tabs += `<button class="${checked?'active':''}" id="${tabId}" data-bs-toggle="tab" data-bs-target="#${tabPanelId}" type="button" role="tab" aria-controls="${tabPanelId}" aria-selected="${checked}">${tabName}</button>`
            contents += `<div class="tab-pane fade ${checked ? 'show active':''}" id="${tabPanelId}" role="tabpanel" aria-labelledby="${tabId}">
                            ${defaultRender(tokens, i, options, env, slf)}
                        </div>`
        }

        return `<div class="nav code-tabs" id="tab-${counter}" role="tablist">
                ${tabs}
                </div>
                <div class="tab-content code" id="tabContent-${counter}">
                ${contents}
                </div>`
    
    }

    md.renderer.rules.fence = fenceGroup;
};