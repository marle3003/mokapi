<script setup lang="ts">
import { type PropType, computed } from 'vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import { usePrettyBytes } from '@/composables/usePrettyBytes'

const props = defineProps<{
  source: string
  contentType?: string
  filename?: string
}>()

// const props = defineProps({
//     source: { type: Object as PropType<string>, required: true },
//     contentType: { type: Object as PropType<string>, required: false },
//     filename: { type: Object as PropType<string>, required: false }
// })

const { formatLanguage } = usePrettyLanguage()
const { format } = usePrettyBytes()

const contentType = computed(() => {
    if (props.contentType) {
      props.contentType
    }
    return 'application/json'
})

const code = computed(() => {
  return formatLanguage(props.source, contentType.value)
})

const lines = computed(() => {
  return code.value.split('\n').length
})

const size = computed(() => {
  return format(new Blob([props.source]).size)
})

function raw() {
  var wnd = window.open("about:blank", "", "_blank")
  wnd!.document.write(props.source)
  return false
}

function download() {
  var element = document.createElement('a')
  element.setAttribute('href', `data:${contentType};charset=utf-8,${encodeURIComponent(props.source)}`)
  let filename = props.filename
  if (!filename) {
    filename = 'file.dat'
  }
  element.setAttribute('download', filename)

  element.style.display = 'none'
  document.body.appendChild(element)
  element.click()
  document.body.removeChild(element)
  return false
}

</script>

<template>
  <div>
    <div class="header">
      <div>
        <span>{{ lines }} lines</span> · 
        <span>{{ size }}</span>
      </div>
      <div class="controls">
        <a href="" @click="raw()">Raw</a>
        <a href="" @click="download()"><i class="bi bi-download"></i></a>
      </div>
    </div>
    <div class="source">
      <pre v-highlightjs="code" class="overflow-auto"><code></code></pre>
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
.source {
  border: 1px solid var(--color-tabs-border);
  border-top: 0;
  border-radius: 0 0 6px 6px;
}
.source pre {
  margin-bottom: 0;
  background-color: transparent;
}
.source pre .hljs {
  border-radius: 0 0 6px 6px;
}
</style>