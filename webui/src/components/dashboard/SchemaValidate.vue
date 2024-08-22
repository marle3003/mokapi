<script setup lang="ts">
import { useGuid } from '@/composables/guid'
import { useExample } from '@/composables/example'
import { watch, reactive } from 'vue'
import '../../ace-editor/ace-config';
import SourceView from './SourceView.vue'

import { transformPath } from '@/composables/fetch'

const props = withDefaults(defineProps<{
  schema: Schema
  name?: string
  contentType: string
  title?: string
}>(), {
    title: 'Data Validator'
})
const { createGuid } = useGuid()
const { fetchExample } = useExample()

const id = createGuid()

const state = reactive({
    content: '',
    errors: <string[]>[],
    validated: false
})
var example = fetchExample({name: props.name, schema: props.schema, contentType: props.contentType})
watch(() => props.schema, (schema) => {
    example = fetchExample({name: props.name, schema: schema, contentType: props.contentType})
})

function setExample() {
    if (example.error) {
        state.errors = [example.error]
        return
    }else{
        state.errors = []
    }
    state.content = example.data
    example.next()
}

function validate() {
    const body = {
        schema: props.schema,
        data: state.content
    }
    fetch(transformPath('/api/schema/validate'), { method: 'POST', body: JSON.stringify(body), headers: { 'Data-Content-Type': props.contentType, 'Content-Type': 'application/json' } })
    .then(res => {
        if (res.status === 200){
            return ''
        }else {
            return res.text()
        }
    }).then(errors => {
        if (errors === '') {
            state.errors = []
        } else{
            state.errors = errors.split('\n').filter(i => i)
        }
        state.validated = true
    })
}

</script>

<template>
    <div>
        <button type="button" class="btn btn-primary btn-sm" data-bs-toggle="modal" :data-bs-target="'#'+id">Example & Validate</button>
        <div class="modal fade" :id="id" tabindex="-1"  aria-hidden="true" :aria-labelledby="id+'title'">
            <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
                <div class="modal-content">
                    <div class="modal-header">
                        <h6 :id="id+'title'" class="modal-title">{{ title }}</h6>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                    </div>
                    <div class="modal-body">
                        <div class="alert alert-danger" role="alert" v-if="state.errors.length > 0">
                            <ul class="error-list">
                                <li v-for="err in state.errors">{{ err }}</li>
                            </ul>
                        </div>
                        <source-view ref="source" v-model:source="state.content" :content-type="contentType" :readonly="false" @update="(source) => state.content = source" />
                    </div>
                    <div class="modal-footer justify-content-between">
                        <span class="float-start">
                            <span v-if="state.errors.length == 0 && state.validated" class="status">
                                <i class="bi bi-check-circle-fill"> Valid</i>
                            </span>
                        </span>
                        <div class="float-end">
                            <button type="button" class="btn btn-primary btn-sm" @click="setExample()">Example</button>
                            <button type="button" class="btn btn-primary btn-sm" @click="validate()">Validate</button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style>
.error-list {
    margin-bottom: 0;
}
.modal-footer button {
    margin-left: 15px;
}
.modal-footer .status {
    color: var(--color-green);
    font-size: 16px;
}
</style>