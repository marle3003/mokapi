declare module 'markdown-it-codetabs'
// todo: remove it after whatwg-mimetype updates its @types/whatwg-mimetype
declare module 'whatwg-mimetype';

declare module 'markdown-it/lib/renderer' {
  import type { Renderer } from 'markdown-it'
  export default Renderer
}
declare module 'markdown-it/lib/token' {
  import type { Token } from 'markdown-it'
  export default Token
}
