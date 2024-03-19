declare module 'vue3-markdown-it'
declare module 'vue3-highlightjs'
declare module 'markdown-it-highlightjs'
declare module 'markdown-it-codetabs'
declare module '@ssthouse/vue3-tree-chart'
declare module 'highlight.js' {
    export interface HLJS{
        highlightAuto: function()
    }
    const hljs: HLJS
    export default hljs
}