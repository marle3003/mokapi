<script setup lang="ts">
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import { useGuid } from '@/composables/guid';
import type { PropType } from 'vue';
import { useFetch } from '@/composables/fetch';

const props = defineProps({
    schema: { type: Object as PropType<Schema>, required: true }
})

const {createGuid} = useGuid();
const {formatLanguage} = usePrettyLanguage()
const example = useFetch('/api/schema/example', {method: 'POST', body: JSON.stringify(props.schema), headers: {'Content-Type': 'application/json', 'Accept': 'application/json'}}, false)

const id = createGuid()
</script>

<template>
    <button type="button" class="btn btn-primary btn-sm" data-bs-toggle="modal" :data-bs-target="'#'+id">Example</button>
    <div class="modal fade" :id="id" tabindex="-1"  aria-hidden="true">
        <div class="modal-dialog modal-dialog-centered modal-dialog-scrollable">
            <div class="modal-content">
                <div class="modal-body">
                    <div class="codeBlock">
                        <pre v-highlightjs="formatLanguage(JSON.stringify(example.data), 'application/json')"><code class="json"></code></pre>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>