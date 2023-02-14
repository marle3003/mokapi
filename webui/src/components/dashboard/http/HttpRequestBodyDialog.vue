<script setup lang="ts">
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import type { PropType } from 'vue';

const props = defineProps({
    content: { type: Object as PropType<HttpMediaType>, required: true }
})

const {formatLanguage} = usePrettyLanguage()
</script>

<template>
    <div class="modal fade" id="exampleModal" tabindex="-1" aria-labelledby="exampleModalLabel" aria-hidden="true">
        <div class="modal-dialog modal-dialog-centered modal-dialog-scrollable">
            <div class="modal-content">
                <div class="modal-body">
                    <div class="codeBlock">
                        <span class="label">{{ content.type }}</span>
                        <pre v-highlightjs="formatLanguage(JSON.stringify(content.schema), 'application/json')"><code class="json"></code></pre>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.codeBlock {
  position: relative;
}
.codeBlock code {
    max-height: 300px;
}
.card-body .codeBlock .label {
    position: absolute;
    right: 4px;
    top: 4px;
    font-size: 0.8rem;
  }
</style>