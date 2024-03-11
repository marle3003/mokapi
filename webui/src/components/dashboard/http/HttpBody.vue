<script setup lang="ts">
import { computed } from 'vue'
import SourceView from '../SourceView.vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'

const props = defineProps<{
    contentType: string,
    body: string
}>()

const { formatLanguage } = usePrettyLanguage()

const source = computed(() => {
  return formatLanguage(props.body, props.contentType)
})
</script>

<template>
  <source-view v-if="body" :source="source" :content-type="contentType" :max-height="500"></source-view>
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