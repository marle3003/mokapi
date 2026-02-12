<script setup lang="ts">
import { computed, getCurrentInstance, ref, watch } from 'vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import { usePrettyBytes } from '@/composables/usePrettyBytes'
import { VAceEditor } from 'vue3-ace-editor'
import '@/ace-editor/ace-config'
import HexEditor from '../HexEditor.vue'
import { Range } from 'ace-builds'
import { useRoute, useRouter } from '@/router'

interface AceMouseEvent {
  domEvent: MouseEvent
  getDocumentPosition(): { row: number }
}

const props = withDefaults(defineProps<{
  source: Source
  filename?: string
  url?: string
  deprecated?: boolean
  height?: number
  maxHeight?: number
  hideContentType?: boolean
  hideHeader?: boolean
  readonly?: boolean
}>(), { readonly: true })

const preview = ref<HTMLElement>()
const binary = ref<HTMLElement>()

const emit = defineEmits<{
  (e: 'update', value: { content: string, type: string}): void,
  (e: 'switch', value: 'preview' | 'binary'): void
}>()

const route = useRoute();
const router = useRouter();
const { getLanguage } = usePrettyLanguage()
const { format } = usePrettyBytes()

let current = ref<{data: Data, type: string}>()
if (props.source.preview) {
  current.value = { data: props.source.preview, type: 'preview' }
} else if (props.source.binary) {
  current.value = { data: props.source.binary, type: 'binary' }
} else {
  throw new Error('preview and binary not defined')
}
watch(() => props.source, (source) => {
    if (!current.value) {
      return
    }
    
    if (current.value?.type === 'preview') {
      current.value.data = source.preview!
    } else {
      current.value.data = source.binary!
    }
}, { deep: true })

const lines = computed(() => {
  let content = props.source.preview?.content
  if (props.source.binary) {
    content = props.source.binary.content
  }
  return content?.split('\n').length ?? 0
})

const size = computed(() => {
  let content = props.source.preview?.content
  if (props.source.binary) {
    content = props.source.binary.content
  }
  return format(new Blob([content!]).size)
})

const viewHeight = computed(() => {
  if (props.height) {
    return props.height
  }
  if (props.readonly) {
    let height = lines.value * 23 + 10
    if (props.maxHeight && height > props.maxHeight) {
      return props.maxHeight
    }
    return Math.max(height, 250)
  }
  return 500
})

const theme = computed(() => {
  const theme = document.documentElement.getAttribute('data-theme')
  return `mokapi-${theme}`
})

function download(event: MouseEvent) {
  var element = document.createElement('a')
  element.setAttribute('href', `data:${current.value?.data.contentType};charset=utf-8,${encodeURIComponent(current.value?.data.content!)}`)
  let filename = props.filename
  if (!filename) {
    filename = 'file.dat'
  }
  element.setAttribute('download', filename)

  element.style.display = 'none'
  document.body.appendChild(element)
  element.click()
  document.body.removeChild(element)
  event.preventDefault()
}

function copyToClipboard(event: MouseEvent) {
  navigator.clipboard.writeText(current.value?.data.content!)
  event.preventDefault()
}

function switchPreview() {
  current.value = { data: props.source.preview!, type: 'preview' }
  preview.value?.classList.add('active')
  binary.value?.classList.remove('active')
  emit('switch', 'preview')
}
function switchCode() {
  current.value = { data: props.source.binary!, type: 'binary' }
  binary.value?.classList.add('active')
  preview.value?.classList.remove('active')
  emit('switch', 'binary')
}
const editor = ref<typeof VAceEditor | null>(null)
let lineSelectionMarker: number | null = null // persistent marker id

function updateSelectionDirect(startRow: number, endRow: number) {
  if (!editor.value) return
  const session = editor.value.session

  if (lineSelectionMarker !== null) {
    session.removeMarker(lineSelectionMarker)
  }

  lineSelectionMarker = session.addMarker(
    new Range(startRow, 0, endRow, Infinity),
    'marker',
    'fullLine',
    false // back layer
  )

  updateGutterMarker(startRow, endRow);
}

function updateGutterMarker(startRow: number, endRow: number) {
  if (!editor.value) return

  const session = editor.value.session

  // remove old annotation for simplicity
  session.clearAnnotations()

  // prepare annotations array
  const annotations = []
  for (let row = startRow; row <= startRow; row++) {
    annotations.push({
      row,
      column: 0,
      text: `Line ${row+1}`,
      type: 'info'        // "error" | "warning" | "info"
    })
  }

  session.setAnnotations(annotations)
}

const startRow = ref<number | undefined>()

const onEditorInit = (e: typeof VAceEditor) => {
  editor.value = e
  e.setOptions({
    readOnly: props.readonly
  });

  applyHashSelection();

  // Gutter drag logic
  editor.value.on('guttermousedown', (e: AceMouseEvent) => {
    let row = e.getDocumentPosition().row
    let from: number
    let to: number
    if (e.domEvent.shiftKey && startRow.value) {
      from = Math.min(startRow.value, row)
      to = Math.max(startRow.value, row)
    } else {
      from = row
      to = row
    }

    // Update URL hash without triggering remount issues
    let hash = from === to ? `#L${from + 1}` : `#L${from + 1}:L${to + 1}`
    if (route.hash === hash) {
      hash = ''
    }
    router.replace({ 
      name: route.name,
      params: route.params,
      hash
    })
    updateSelectionDirect(from, to)
  })
};

// Parse route.hash and update marker
function applyHashSelection() {
  if (!editor.value) return
  const m = route.hash.match(/L([0-9]+)(:L([0-9]+))?/)
  if (!m) {
    return
  }

  startRow.value = Number(m[1]) - 1
  const endRow = Number(m[3] || m[1]) - 1
  updateSelectionDirect(startRow.value, endRow)
}

watch(
  () => route.hash,
  () => {
    applyHashSelection()
  }
)
</script>

<template>
  <section aria-label="Source" class="source-view">
    <div class="header" v-if="!hideHeader">
      <div class="view controls" v-if="source.preview && source.binary">
        <button ref="preview" type="button" class="btn btn-link active" @click="switchPreview()">JSON</button>
        <button ref="binary" type="button" class="btn btn-link" @click="switchCode()">Binary</button>
      </div>
      <div class="info">
        <span style="" v-if="!hideContentType" aria-label="Content Type">{{ current?.data.contentTypeTitle ?? current?.data.contentType }}</span>
        <span aria-label="Lines of Code" v-if="!source.binary">{{ lines }} lines</span>
        <span aria-label="Size of Code">{{ size }}</span>
        <span v-if="deprecated"><span class="bi bi-exclamation-triangle-fill yellow"></span> deprecated</span>
        <span v-if="current?.data.description">{{ current?.data.description }}</span>
      </div>
      <div class="controls">
        <a  v-if="url" :href="url">Raw</a>
        <button type="button" class="btn btn-link" @click="copyToClipboard" title="Copy raw content" aria-label="Copy raw content"><span class="bi bi-copy"></span></button>
        <button type="button" class="btn btn-link" @click="download" title="Download raw content" aria-label="Download raw content"><span class="bi bi-download"></span></button>
      </div>
    </div>
    <section class="source" aria-label="Content" :class="getLanguage(current?.data.contentType!)" v-if="current?.type == 'preview' && current && current.data.content">
      <v-ace-editor
        :id="getCurrentInstance()?.uid"
        :value="current?.data.content!"
        @update:value="emit('update', { content: $event, type: current?.type! })"
        :lang="getLanguage(current?.data.contentType!)"
        :theme="theme"
        style="font-size: 16px;"
        :style="`height: ${viewHeight}px`"
        @init="onEditorInit"
      />
    </section>
    <section class="source" v-else-if="current?.type == 'binary' && source.binary" :class="'ace-'+theme">
      <HexEditor :data="source.binary.content" :style="`height: ${viewHeight}px`" @update="emit('update', { content: $event, type: 'binary'})" :readonly="readonly" />
    </section>
  </section>
</template>

<style scoped>
.source-view .header {
  padding: 8px;
  border: 1px solid var(--source-border);
  border-radius: 6px 6px 0 0;
  color: var(--color-text-light);
  display: flex;

}
.source-view .header .info span {
  vertical-align: middle;
  line-height: 30px;
}
.source-view .header .controls {
  border: 1px solid var(--source-border);
  border-radius: 6px;
}
.source-view .header .view {
  margin-right: 7px;
}
.source-view .header .controls:not(.view) {
  margin-left: auto;
}
.source-view .header button {
  background: none !important;
  font-size: 0.9rem;
  vertical-align: middle;
  color: var(--source-header-color);
  display: inline-grid;
  place-content: center;
  border-right: 1px solid var(--source-border);
}
.source-view .header button.active {
  background-color: black !important;
  outline: 1px solid var(--source-border);
  border-radius: 6px;
}
.source-view .header .controls > button, .header .controls > a {
  padding: 5px 8px;
  height: 28px;
  line-height: 18px;
  position: relative;
  text-decoration: none;
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
}
.source-view .header .controls > button + button, .header .controls > a + a {
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
}
.source-view .header .controls > button:last-child, .header .controls > a:last-child {
  border-top-right-radius: 6px;
  border-bottom-right-radius: 6px;
}
.source-view .header .controls a {
  display: inline-block;
  box-sizing: border-box;
  vertical-align: middle;
}
.source-view .header .controls > *:last-child {
  border-right: 0;
}
.source-view .header span + span::before {
  content: ' · ';
}
.source-view .header .controls > *:hover:not(.active) {
  background-color: var(--source-header-background-active) !important;
}
.source-view .source {
  border: 1px solid var(--source-border);
  border-top: 0;
  border-radius: 0 0 6px 6px;
  overflow: hidden;
  color: var(--color-text);
}
.source-view .source pre {
  margin-bottom: 0;
  background-color: transparent;
}
.source-view .source pre .hljs {
  border-radius: 0 0 6px 6px;
  padding-left: 0.5rem;
}
.source-view .source code {
  display: inline-block;
  min-width: 100%;
}
</style>

<style>
.source .hljs-deletion {
  background-color: transparent;
  color: #C9D1DE;
}
.marker, .source-view .ace_marker-layer .ace_selection {
  position:absolute;
  background-color: rgba(56, 139, 253, 0.15);
  z-index:20
}
[data-theme="light"] .marker, .source-view .ace_marker-layer .ace_selection {
  background-color: rgba(9, 105, 218, 0.12);
}
.ace_gutter-cell {
  cursor: pointer;
}
.ace_gutter-cell.ace_info,
.ace_gutter-cell.ace_warning,
.ace_gutter-cell.ace_error {
  background-image: none !important;
}
.ace_gutter-cell.ace_info .ace_icon::before {
  font-family: bootstrap-icons;
  content: "\F5D4";
  display: inline-block;
  width: 20px;
  text-align: center;
  margin-right: 1px;
  border: 1px solid black;
  border-radius: 3px;
  margin-left: 3px;
  padding-inline: 1px;
}
</style>