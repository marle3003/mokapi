import type MarkdownIt from "markdown-it"

export function MarkdownItBlockquote(md: MarkdownIt) {

    md.core.ruler.after('inline', 'blockquote', function(state){
        state.tokens.forEach(token => {
            if (token.type === 'blockquote_open') {
            token.attrPush(['class', 'blockquote lead']);
            }
        });
    });
}