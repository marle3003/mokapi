import type { MarkdownIt } from "markdown-it"

export function MarkdownItTable(md: MarkdownIt) {

    // Override table_open
    md.renderer.rules.table_open = function () {
        return (
        '<div class="table-responsive-sm">\n' +
        '<table class="table">\n'
        )
    }

    // Override table_close
    md.renderer.rules.table_close = function () {
        return '</table>\n</div>\n'
    }
};