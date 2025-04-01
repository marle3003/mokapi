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
    result: null,
    validated: false,
    validating: false,
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
    state.validating = true
    state.validated = false

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
            return res.json().then((data) => ({
                status: res.status,
                body: data,
            }))
        } else{
            return res.text().then((txt) => ({
                status: res.status,
                body: txt,
            }))
        }
    }).then(({ status, body }) => {
        state.validating = false
        state.validated = true

        if (status === 500) {
            const errors = <string>body
            state.errors = errors.replaceAll('\n', '<br />')
        } else if (status === 400) { 
            state.errors = ''
            state.result = body
        } else {
            let examples = body as Example[]
            state.errors = ''
            state.result = null
            
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
function formatResult(result: any): string {
    const [s, count] = formatResultItem(result)
    return `<div>error count ${count}: ${s}</div>`
}
function formatResultItem(result: any): [string, number] {
    if (Array.isArray(result)) {
        let s = '<ul>'
        let sum = 0
        for (const item of result) {
            const [msg, count] = formatResultItem(item)
            s += `<li>${msg}</li>`
            sum += count
        }
        return [s + '</ul>', result.length]
    }else if (typeof result === "object" && result !== null) {
        if (result.errors) {
            let s = `${result.schema}: ${result.message}<ul>`
            let sum = 0
            for (const item of result.errors) {
                const [msg, count] = formatResultItem(item)
                s += `<li>${msg}</li>`
                sum += count
            }
            return [s + '</ul>', result.length]
        }
        return [`${result.schema}: ${result.message}`, 1]
    }
    return [result, 1]
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
                        <div class="alert alert-danger" role="alert" v-if="state.result">
                            <span v-html="formatResult(state.result)"></span>
                        </div>
                        <source-view ref="source" :filename="source?.filename" v-model:source="state.source" :readonly="false" @update="update" @switch="switchSource" />
                    </div>
                    <div class="modal-footer justify-content-between">
                        <span class="float-start">
                            <span  class="status" role="status">
                                <span v-if="state.validating" class="loading"><i class="bi bi-arrow-repeat spin"></i> Validating...</span>
                                <i v-else-if="state.errors.length == 0 && state.result == null && state.validated" class="valid bi bi-check-circle-fill fade-in"> Valid</i>
                                <i v-else-if="state.validated" class="failed bi bi-x-circle-fill fade-in"> Failed</i>
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
    font-size: 16px;
}
.modal-footer .status .valid {
    color: var(--color-green);

}
.modal-footer .status .failed {
    color: var(--color-red);
}
.modal-footer .status .loading {
    color: var(--color-blue);
}
.spin {
    display: inline-block;
    animation: spin 1s linear infinite;
}
@keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
}
.fade-in {
    animation: fadeIn 0.5s ease-in-out;
}
@keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
}
</style>