<script setup lang="ts">
import { useGuid } from '@/composables/guid'
import { useExample, type Example } from '@/composables/example'
import { watch, reactive } from 'vue'
import '../../ace-editor/ace-config';
import SourceView from './SourceView.vue'
import { transformPath } from '@/composables/fetch'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';

const props = withDefaults(defineProps<{
  schema: SchemaFormat
  name?: string
  title?: string
  source: {
        filename?: string
        preview?: Data,
        binary?: Data
  }
}>(), {
    title: 'Data Validator'
})
const { createGuid } = useGuid()
const { fetchExample } = useExample()
const { formatLanguage } = usePrettyLanguage()

const id = createGuid()

const state = reactive({
    source: { preview: props.source.preview, binary: props.source.binary },
    errors: '',
    validated: false,
    view: 'preview'
})

const exampleRequest = {
    name: props.name,
    schema: props.schema,
    contentTypes: Array()
}
if (props.source.preview) {
    exampleRequest.contentTypes.push(props.source.preview.contentType)
}
if (props.source.binary) {
    exampleRequest.contentTypes.push(props.source.binary.contentType)
}

var example = fetchExample(exampleRequest)
watch(() => props.schema, (schema) => {
    exampleRequest.schema = schema
    example = fetchExample(exampleRequest)
})

function setExample() {
    if (example.error) {
        state.errors = example.error
        return
    }else{
        state.errors = ''
    }
    
    if (state.source.preview) {
        state.source.preview.content = example.data[0].value
    }
    if (state.source.binary) {
        let index = 0
        if (example.data.length > 1) {
            index = 1
        }
        state.source.binary.content = example.data[index].value
    }
    example.next()
}

function validate() {
    const body = {
        format: props.schema.format,
        schema: props.schema.schema,
        data: state.view === 'preview' ? state.source.preview?.content : btoa(state.source.binary!.content),
        contentType: state.view === 'preview' ? props.source.preview!.contentType : props.source.binary!.contentType
    }
    let query = ''
    if (state.source.binary) {
        query = '?outputFormat=' + (state.view === 'preview' ? state.source.binary.contentType : state.source.preview!.contentType)
    }
    fetch(transformPath('/api/schema/validate'+query), { 
        method: 'POST', 
        body: JSON.stringify(body), 
        headers: { 'Content-Type': 'application/json' } 
    })
    .then(res => {
        const contentType = res.headers.get("content-type");
        if (contentType && contentType.indexOf("application/json") !== -1) {
            return res.json()
        } else{
            return res.text()
        }
    }).then(body => {
        state.validated = true

        if (typeof body === 'string' || body instanceof String) {
            const errors = <string>body
            state.errors = errors.replaceAll('\n', '<br />')
        } else {
            let examples = body as Example[]
            state.errors = ''
            
            if (examples.length === 0) {
                return
            }
            if (state.view === 'preview' && state.source.binary) {
                const v = atob(examples[0].value)
                state.source.binary.content = v
            } else if (state.source.preview) {
                const v = atob(examples[0].value)
                state.source.preview.content = formatLanguage(v, examples[0].contentType)
            }
        }    
    })
}

function update(event: {content: string, type: string}) {
    if (event.type === 'preview') {
        state.source.preview!.content = event.content
    }
    if (event.type === 'binary') {
        state.source.binary!.content = event.content
    }
}
function switchSource(e: string) {
    state.view = e
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
                            <span v-html="state.errors"></span>
                        </div>
                        <source-view ref="source" :filename="source?.filename" v-model:source="state.source" :readonly="false" @update="update" @switch="switchSource" />
                    </div>
                    <div class="modal-footer justify-content-between">
                        <span class="float-start">
                            <span v-if="state.errors.length == 0 && state.validated" class="status" role="status">
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
.modal-footer button {
    margin-left: 15px;
}
.modal-footer .status {
    color: var(--color-green);
    font-size: 16px;
}
</style>