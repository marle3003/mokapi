<script setup lang="ts">
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';


const {format, getLanguage} = usePrettyLanguage()
const props = defineProps({
    contentType: { type: String, required: true },
    body: {type: String, required: true}
})
</script>

<template>
    <p class="label">Content Type</p>
    <p>{{ contentType }}</p>
    <div v-if="body">
        <pre v-highlightjs="format(body, contentType)"><code :class="getLanguage(contentType)"></code></pre>
    </div>
</template>

<style scoped>
.w-10{
    width: 10%;
}
</style>