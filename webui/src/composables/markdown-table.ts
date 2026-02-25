import type MarkdownIt from "markdown-it"

export function MarkdownItTable(md: MarkdownIt) {

    // Override table_open
    md.renderer.rules.table_open = function (tokens, idx, options, env, self) {
        return (
        '<div class="table-responsive-sm">\n' +
        '<table class="table">\n'
        )
    }

    // Override table_close
    md.renderer.rules.table_close = function (tokens, idx, options, env, self) {
        return '</table>\n</div>\n'
    }
};