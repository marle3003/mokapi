<script setup lang="ts">
import { useGuid } from '@/composables/guid'
import { useExample } from '@/composables/example'
import { watch } from 'vue'
import SourceView from './SourceView.vue'

const props = defineProps<{
  schema: Schema
  contentType: string
  title?: string
  source?: {
    filename: string
  }
}>()

const { createGuid } = useGuid()
const { fetchExample } = useExample()

var example = fetchExample(props.schema, props.contentType)
watch(() => props.schema, (schema) => {
    example = fetchExample(schema, props.contentType)
})

const id = createGuid()
</script>

<template>
    <div data-testid="example">
        <button type="button" class="btn btn-primary btn-sm" data-bs-toggle="modal" :data-bs-target="'#'+id">Example</button>
        <div class="modal fade" :id="id" tabindex="-1"  aria-hidden="true" :aria-labelledby="id+'title'">
            <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
                <div class="modal-content">
                    <div class="modal-header" v-if="title">
                        <h6 :id="id+'title'" class="modal-title">{{ title }}</h6>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                    </div>
                    <div class="modal-body">
                        <source-view v-if="example" :source="example.data" :content-type="contentType" :filename="source?.filename" />
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>