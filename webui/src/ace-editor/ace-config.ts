import ace from 'ace-builds'
import modeJsonUrl from 'ace-builds/src-noconflict/mode-json?url'
import xmlUrl from 'ace-builds/src-noconflict/mode-xml?url'
import yamlUrl from 'ace-builds/src-noconflict/mode-yaml?url'
import jsUrl from 'ace-builds/src-noconflict/mode-javascript?url'
import mokapi_dark from '@/ace-editor/ace-theme-mokapi-dark.js?url'
import mokapi_light from '@/ace-editor/ace-theme-mokapi-light.js?url'


ace.config.setModuleUrl('ace/mode/json', modeJsonUrl)
ace.config.setModuleUrl('ace/mode/xml', xmlUrl)
ace.config.setModuleUrl('ace/mode/yaml', yamlUrl)
ace.config.setModuleUrl('ace/mode/javascript', jsUrl)

ace.config.setModuleUrl('ace/theme/mokapi-dark', mokapi_dark)
ace.config.setModuleUrl('ace/theme/mokapi-light', mokapi_light)
