<script setup lang="ts">
import { computed } from 'vue'
import hljs from '@/plugins/highlight'

// Highlighted while rendering, so the code is highlighted in the pre-rendered HTML too (the
// `highlightjs` directive only runs in the browser).
const props = defineProps<{ code: string, language: string }>()

const highlighted = computed(() => hljs.highlight(props.code, { language: props.language }).value)
</script>

<template>
  <pre><code :class="`hljs language-${language}`" v-html="highlighted"></code></pre>
</template>
