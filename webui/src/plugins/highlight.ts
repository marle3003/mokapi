import hljs from 'highlight.js/lib/core'
import json from 'highlight.js/lib/languages/json'
import yaml from 'highlight.js/lib/languages/yaml'
import javascript from 'highlight.js/lib/languages/javascript'
import bash from 'highlight.js/lib/languages/bash'
import ldif from 'highlight.js/lib/languages/ldif'

hljs.registerLanguage('json', json)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('ldif', ldif)

export default hljs