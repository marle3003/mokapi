<script setup lang="ts">
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import { useGuid } from '@/composables/guid';
import type { PropType } from 'vue';
import SourceView from './SourceView.vue'

defineProps({
    schema: { type: Object as PropType<Schema>, required: true }
})

const {createGuid} = useGuid();
const id = createGuid()
</script>

<template>
    <div data-testid="expand">
        <button type="button" class="btn btn-primary btn-sm" data-bs-toggle="modal" :data-bs-target="'#'+id">Expand</button>
        <div class="modal fade" :id="id" tabindex="-1"  aria-hidden="true">
            <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
                <div class="modal-content">
                    <div class="modal-body">
                        <div class="codeBlock">
                            <source-view :source="JSON.stringify(schema)" content-type="application/json" :hide-content-type="true" />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>