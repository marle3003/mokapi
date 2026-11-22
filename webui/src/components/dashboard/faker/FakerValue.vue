<script setup lang="ts">
import { computed } from 'vue'
import hljs from '@/plugins/highlight'

// A generated value as highlighted JSON.
const props = defineProps<{ value: unknown }>()

const html = computed(() => hljs.highlight(JSON.stringify(props.value, null, 2) ?? 'null', { language: 'json' }).value)
</script>

<template>
    <pre class="hljs faker-value" aria-label="Generated value"><code v-html="html"></code></pre>
</template>
