import type { Options } from "markdown-it"
import type MarkdownIt from "markdown-it"
import type Token from "markdown-it/lib/token"

export function MarkdownItLinks(md: MarkdownIt, opts: Options) {

    function replace(s: string): string{
        s = s.replace('.md', '')
        return s
    }

    md.core.ruler.after('inline', 'link', function(state){
        state.tokens.forEach(function (blockToken: Token) {
            if (blockToken.type === 'inline' && blockToken.children) {
              blockToken.children.forEach(function (token) {
                if (token.type == 'link_open' && token.attrs) {
                    for (let attr of token.attrs){
                        if (attr[0] == 'href' && attr[1].endsWith('.md')){
                            attr[1] = replace(attr[1])
                        }
                    }
                }
              })
            }
          })
    })
};