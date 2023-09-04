<script setup lang="ts">
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';

const {formatLanguage, getLanguage} = usePrettyLanguage()
defineProps({
    contentType: { type: String, required: true },
    body: {type: String, required: true}
})
</script>

<template>
    <div class="codeBlock" v-if="body">
        <span class="label">{{ contentType }}</span>
        <pre v-highlightjs="formatLanguage(body, contentType)" class="overflow-auto" style="max-height: 500px;">
          <code :class="getLanguage(contentType)"></code>
        </pre>
    </div>
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