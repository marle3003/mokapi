declare module 'vue3-markdown-it';
declare module 'vue3-highlightjs'
declare module 'highlight.js' {
    export interface HLJS{
        highlightAuto: function()
    }
    const hljs: HLJS
    export default hljs
}