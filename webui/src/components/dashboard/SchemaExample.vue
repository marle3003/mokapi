<script setup lang="ts">
import { useGuid } from '@/composables/guid';
import type { PropType } from 'vue';
import { useExample } from '@/composables/example';

const props = defineProps({
    schema: { type: Object as PropType<Schema>, required: true }
})

const {createGuid} = useGuid();
const example = useExample().fetchExample(props.schema)

const id = createGuid()
</script>

<template>
    <button type="button" class="btn btn-primary btn-sm" data-bs-toggle="modal" :data-bs-target="'#'+id">Example</button>
    <div class="modal fade" :id="id" tabindex="-1"  aria-hidden="true">
        <div class="modal-dialog modal-dialog-centered modal-dialog-scrollable">
            <div class="modal-content">
                <div class="modal-body">
                    <pre v-highlightjs="example"><code class="json"></code></pre>
                </div>
            </div>
        </div>
    </div>
</template>