<script setup lang="ts">
import { computed } from 'vue'
import SourceView from '../SourceView.vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'

const props = defineProps<{
    contentType: string,
    body?: string
    size?: number
}>()

const { formatLanguage } = usePrettyLanguage()

const source = computed(() => {
  if (!props.body) {
    return ''
  }
  return formatLanguage(props.body, props.contentType)
})

const size = computed(() => {
  if (props.size) {
    return props.size
  }
  if (!props.body) {
    return 0
  }
  return new TextEncoder().encode(props.body).length;
})
</script>

<template>
  <source-view v-if="body"
    :source="{ preview: { content: source, contentType: contentType }}"
    :size="size"
    :max-height="500">
  </source-view>
  <div v-else>No body returned for response</div>
</template>

<style scoped>
.codeBlock {
  position: relative;
}
.card-body .codeBlock .label {
    position: absolute;
    right: 4px;
    top: 4px;
    font-size: 0.8rem;
  }
</style>