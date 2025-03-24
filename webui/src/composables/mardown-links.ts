import type { Options } from "markdown-it"
import type MarkdownIt from "markdown-it"
import Token from "markdown-it/lib/token"
import { useRoute } from "vue-router"

export function MarkdownItLinks(md: MarkdownIt, opts: Options) {

    function replace(s: string): string{
        s = s.replace('.md', '')
        const u = new URL(s, document.location.href)
        return u.toString()
    }

    const route = useRoute()

    md.core.ruler.after('inline', 'link', function(state){
        state.tokens.forEach(function (blockToken: Token) {
            if (blockToken.type === 'inline' && blockToken.children) {
              blockToken.children.forEach(function (token) {
                if (token.type == 'link_open' && token.attrs) {
                    for (let attr of token.attrs){
                        if (attr[0] == 'href' && attr[1].endsWith('.md')){
                            attr[1] = replace(attr[1])
                        }
                        else if (attr[0] == 'href' && attr[1].includes('.md#')){
                            attr[1] = replace(attr[1])
                        }
                        else if (attr[0] == 'href' && attr[1].startsWith('#')){
                            attr[1] = route.path + attr[1]
                        }
                    }
                }
              })
            }
          })
    })
};