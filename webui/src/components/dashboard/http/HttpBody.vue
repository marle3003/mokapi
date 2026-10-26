<script setup lang="ts">
import { computed } from 'vue'
import SourceView from '../SourceView.vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'

const props = defineProps<{
    contentType: string,
    preview?: HttpPreviewBody
}>()

const { formatLanguage } = usePrettyLanguage()

const source = computed(() => {
  if (!props.preview) {
    return ''
  }
  return formatLanguage(props.preview.body, props.contentType, props.preview.truncated)
})
</script>

<template>
  <source-view v-if="preview"
    :source="{ preview: { content: source, contentType: contentType }}"
    :size="preview.size" :truncated="preview.truncated"
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