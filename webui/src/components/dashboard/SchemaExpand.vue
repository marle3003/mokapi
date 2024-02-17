<script setup lang="ts">
import { useGuid } from '@/composables/guid'
import SourceView from './SourceView.vue'

defineProps<{
    schema: Schema
    title?: string
    source?: {
        filename: string
    }
}>()

const { createGuid } = useGuid();
const id = createGuid()
</script>

<template>
    <div data-testid="expand">
        <button type="button" class="btn btn-primary btn-sm" data-bs-toggle="modal" :data-bs-target="'#'+id">Expand</button>
        <div class="modal fade" :id="id" tabindex="-1" aria-hidden="true" :aria-labelledby="id+'title'">
            <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
                <div class="modal-content">
                    <div class="modal-header" v-if="title">
                        <h6 :id="id+'title'" class="modal-title">{{ title }}</h6>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                    </div>
                    <div class="modal-body">
                        <div class="codeBlock">
                            <source-view :source="JSON.stringify(schema)" content-type="application/json" :hide-content-type="true" :filename="source?.filename" />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.modal-header {
    border-color: var(--color-tabs-border);
    padding-bottom: 15px;
    padding-top: 15px;
}
</style>