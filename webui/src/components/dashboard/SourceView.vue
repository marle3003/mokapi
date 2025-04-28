<script setup lang="ts">
import { computed, getCurrentInstance, ref, watch } from 'vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import { usePrettyBytes } from '@/composables/usePrettyBytes'
import { VAceEditor } from 'vue3-ace-editor'
import '@/ace-editor/ace-config'
import HexEditor from '../HexEditor.vue'

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

const editor = ref()
const preview = ref<HTMLElement>()
const binary = ref<HTMLElement>()

const emit = defineEmits<{
  (e: 'update', value: { content: string, type: string}): void,
  (e: 'switch', value: 'preview' | 'binary'): void
}>()

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
  return content!.split('\n').length
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
</script>

<template>
  <section aria-label="Source">
    <div class="header" v-if="!hideHeader">
      <div class="view controls" v-if="source.preview && source.binary">
        <button ref="preview" type="button" class="btn btn-link active" @click="switchPreview()">JSON</button>
        <button ref="binary" type="button" class="btn btn-link" @click="switchCode()">Binary</button>
      </div>
      <div class="info">
        <span style="" v-if="!hideContentType">{{ current?.data.contentTypeTitle ?? current?.data.contentType }}</span>
        <span aria-label="Lines of Code" v-if="!source.binary">{{ lines }} lines</span>
        <span aria-label="Size of Code">{{ size }}</span>
        <span v-if="deprecated"><i class="bi bi-exclamation-triangle-fill yellow"></i> deprecated</span>
        <span v-if="current?.data.description">{{ current?.data.description }}</span>
      </div>
      <div class="controls">
        <a  v-if="url" :href="url">Raw</a>
        <button type="button" class="btn btn-link" @click="copyToClipboard" title="Copy raw content" aria-label="Copy raw content"><i class="bi bi-copy"></i></button>
        <button type="button" class="btn btn-link" @click="download" title="Download raw content" aria-label="Download raw content"><i class="bi bi-download"></i></button>
      </div>
    </div>
    <section class="source" aria-label="Content" :class="getLanguage(current?.data.contentType!)" v-if="current?.type == 'preview'">
      <v-ace-editor
        ref="editor"
        :id="getCurrentInstance()?.uid"
        :value="current?.data.content!"
        @update:value="emit('update', { content: $event, type: current?.type! })"
        :lang="getLanguage(current?.data.contentType!)"
        :theme="theme"
        style="font-size: 16px;"
        :style="`height: ${viewHeight}px`"
        :options="{
          readOnly: props.readonly
        }"
      />
    </section>
    <section class="source" v-else-if="current?.type == 'binary' && source.binary" :class="'ace-'+theme">
      <HexEditor :data="source.binary.content" :style="`height: ${viewHeight}px`" @update="emit('update', { content: $event, type: 'binary'})" :readonly="readonly" />
    </section>
  </section>
</template>

<style scoped>
.header {
  padding: 8px;
  border: 1px solid var(--source-border);
  border-radius: 6px 6px 0 0;
  color: var(--color-text-light);
  display: flex;

}
.header .info span {
  vertical-align: middle;
  line-height: 30px;
}
.header .controls {
  border: 1px solid var(--source-border);
  border-radius: 6px;
}
.header .view {
  margin-right: 7px;
}
.header .controls:not(.view) {
  margin-left: auto;
}
.header button {
  background: none !important;
  font-size: 0.9rem;
  vertical-align: middle;
  color: var(--source-header-color);
  display: inline-grid;
  place-content: center;
  border-right: 1px solid var(--source-border);
}
.header button.active {
  background-color: black !important;
  outline: 1px solid var(--source-border);
  border-radius: 6px;
}
.header .controls > button, .header .controls > a {
  padding: 5px 8px;
  height: 28px;
  line-height: 18px;
  position: relative;
  text-decoration: none;
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
}
.header .controls > button + button, .header .controls > a + a {
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
}
.header .controls > button:last-child, .header .controls > a:last-child {
  border-top-right-radius: 6px;
  border-bottom-right-radius: 6px;
}
.header .controls a {
  display: inline-block;
  box-sizing: border-box;
  vertical-align: middle;
}
.header .controls > *:last-child {
  border-right: 0;
}
.header span + span::before {
  content: ' · ';
}
.header .controls > *:hover:not(.active) {
  background-color: var(--source-header-background-active) !important;
}
.source {
  border: 1px solid var(--source-border);
  border-top: 0;
  border-radius: 0 0 6px 6px;
  overflow: hidden;
  color: var(--color-text);
}
.source pre {
  margin-bottom: 0;
  background-color: transparent;
}
.source pre .hljs {
  border-radius: 0 0 6px 6px;
  padding-left: 0.5rem;
}
.source code {
  display: inline-block;
  min-width: 100%;
}
</style>

<style>
.source .hljs-deletion {
  background-color: transparent;
  color: #C9D1DE;
}
</style>