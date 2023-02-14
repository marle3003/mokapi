<script setup lang="ts">
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';

const {formatLanguage, getLanguage} = usePrettyLanguage()
const props = defineProps({
    contentType: { type: String, required: true },
    body: {type: String, required: true}
})
</script>

<template>
    <div class="codeBlock">
        <span class="label">{{ contentType }}</span>
        <pre v-highlightjs="formatLanguage(body, contentType)"><code :class="getLanguage(contentType)"></code></pre>
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