<script setup lang="ts">
import { useGuid } from '@/composables/guid'
import { useExample } from '@/composables/example'
import { watch } from 'vue';

const props = defineProps<{
  schema: Schema
  contentType: string
}>()

const { createGuid } = useGuid()

var example = useExample().fetchExample(props.schema, props.contentType)
watch(() => props.schema, (schema) => {
    const r = useExample().fetchExample(schema, props.contentType)
    example = r
})

const id = createGuid()
</script>

<template>
    <div data-testid="example">
        <button type="button" class="btn btn-primary btn-sm" data-bs-toggle="modal" :data-bs-target="'#'+id">Example</button>
        <div class="modal fade" :id="id" tabindex="-1"  aria-hidden="true">
            <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
                <div class="modal-content">
                    <div class="modal-body" v-if="example">
                        <pre v-highlightjs="example"><code class="json"></code></pre>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>