<script setup lang="ts">
import { computed } from 'vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import { usePrettyBytes } from '@/composables/usePrettyBytes'

const props = defineProps<{
  source: string
  contentType: string
  filename?: string
  url?: string
  deprecated?: boolean
  height?: string
  hideContentType?: boolean
}>()

const { formatLanguage } = usePrettyLanguage()
const { format } = usePrettyBytes()

const code = computed(() => {
  return formatLanguage(props.source, props.contentType)
})

const lines = computed(() => {
  return code.value.split('\n').length
})

const size = computed(() => {
  return format(new Blob([props.source]).size)
})

const highlightClass = computed(() => {
  if (!props.contentType) {
    return ''
  }
  switch (props.contentType) {
    case 'application/json': return 'json'
    case 'application/yaml': return 'yaml'
    case 'application/xml': return 'xml'
    default: return ''
  }
})

function download(event: MouseEvent) {
  var element = document.createElement('a')
  element.setAttribute('href', `data:${props.contentType};charset=utf-8,${encodeURIComponent(props.source)}`)
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
  navigator.clipboard.writeText(code.value)
  event.preventDefault()
}

</script>

<template>
  <div>
    <div class="header">
      <div>
        <span v-if="!hideContentType">{{ contentType }}</span>
        <span>{{ lines }} lines</span>
        <span>{{ size }}</span>
        <span v-if="deprecated"><i class="bi bi-exclamation-triangle-fill yellow"></i> deprecated</span>
      </div>
      <div class="controls">
        <a  v-if="url" :href="url">Raw</a>
        <a href="" @click="copyToClipboard" title="Copy source"><i class="bi bi-copy"></i></a>
        <a href="" @click="download" title="Download source"><i class="bi bi-download"></i></a>
      </div>
    </div>
    <div class="source">
      <pre v-highlightjs="code" class="overflow-auto" :style="(height) ? 'max-height: '+height : ''"><code :class="highlightClass"></code></pre>
    </div>
  </div>
</template>

<style scoped>
.header {
  padding: 8px;
  border: 1px solid var(--color-tabs-border);
  border-radius: 6px 6px 0 0;
  color: var(--color-text-light);
  display: flex;
  justify-content: space-between;
}
.header .controls a {
  border: 1px solid var(--color-tabs-border);
  padding: 5px 8px;
}
.header .controls a:hover {
  border-color: var(--color-text-light);
  color: var(--color-link);
}
.header .controls a:first-child {
  border-top-left-radius: 6px;
  border-bottom-left-radius: 6px;
}
.header .controls a:last-child {
  border-top-right-radius: 6px;
  border-bottom-right-radius: 6px;
}
.header span + span::before {
  content: ' · ';
}
.source {
  border: 1px solid var(--color-tabs-border);
  border-top: 0;
  border-radius: 0 0 6px 6px;
  overflow: hidden;
}
.source pre {
  margin-bottom: 0;
  background-color: transparent;
}
.source pre .hljs {
  border-radius: 0 0 6px 6px;
}
</style>