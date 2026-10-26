<script setup lang="ts">
import { watch, reactive, useId, ref, onMounted, onBeforeUnmount } from 'vue'
import '../../plugins/ace-editor/ace-config';
import SourceView from './SourceView.vue'
import { transformPath } from '@/composables/fetch'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import { useDashboard } from '@/composables/dashboard';
import type { Example } from '@/types/dashboard';

declare interface State {
    source: { preview: Data | undefined, binary: Data | undefined }
    errors: string
    result: any
    validated: boolean
    validating: boolean
    view: 'preview' | 'binary'
}

declare interface ExampleRequest {
    name?: string
    path?: string[]
}

const props = withDefaults(defineProps<{
    schema: SchemaFormat
    example?: ExampleRequest
    title?: string
    source: {
        filename?: string
        preview?: Data,
        binary?: Data
    }
    exampleEnabled?: boolean
}>(), {
    title: 'Data Validator',
    exampleEnabled: true
})

const { formatLanguage } = usePrettyLanguage()
const { dashboard } = useDashboard()
const id = useId()
const modal = ref<HTMLElement>()

const state = reactive<State>({
    source: { preview: undefined, binary: undefined },
    errors: '',
    result: null,
    validated: false,
    validating: false,
    view: 'preview'
})

type ExampleHandle = ReturnType<typeof dashboard.value.getExample>

let example: ExampleHandle | null = null
let isOpen = false

function buildExampleRequest() {
    const contentTypes: string[] = []
    if (state.source.preview) contentTypes.push(state.source.preview.contentType)
    if (state.source.binary) contentTypes.push(state.source.binary.contentType)
    return {
        name: props.example?.name,
        path: props.example?.path,
        schema: props.schema,
        contentTypes
    }
}

function loadExample() {
    if (!props.exampleEnabled) return
    example = dashboard.value.getExample(buildExampleRequest())
}

// Schema/Source haben sich geändert: altes Example verwerfen.
// Neu geladen wird nur, wenn der Dialog gerade offen ist.
function invalidateExample() {
    example = null
    if (isOpen) loadExample()
}

function onShow() {
    isOpen = true
    if (!example) loadExample()
}
function onHidden() {
    isOpen = false
}

onMounted(() => {
    modal.value?.addEventListener('show.bs.modal', onShow)
    modal.value?.addEventListener('hidden.bs.modal', onHidden)
})
onBeforeUnmount(() => {
    modal.value?.removeEventListener('show.bs.modal', onShow)
    modal.value?.removeEventListener('hidden.bs.modal', onHidden)
})

watch(() => props.source, (source) => {
    const content = state.source.preview?.content
    state.source.preview = source.preview
        ? { ...source.preview, content: content || source.preview.content }
        : undefined
    state.source.binary = source.binary ? { ...source.binary } : undefined
    invalidateExample()
}, { immediate: true })

watch([() => props.schema, () => props.example], invalidateExample)

function setExample() {
    if (!example) loadExample()
    const ex = example
    if (!ex) return

    if (ex.error) {
        state.errors = ex.error
        return
    }
    state.errors = ''

    const [first, second] = ex.data
    if (state.source.preview) {
        if (first?.error) state.errors = first.error
        else if (first) state.source.preview.content = first.value
    }
    if (state.source.binary) {
        const item = ex.data.length > 1 ? second : first
        if (item?.error) state.errors = item.error
        else if (item) state.source.binary.content = item.value
    }
    ex.next()
}

async function validate() {
    const isPreview = state.view === 'preview'
    const current = isPreview ? state.source.preview : state.source.binary
    const other = isPreview ? state.source.binary : state.source.preview
    if (!current) return

    state.validating = true
    state.validated = false

    const body = {
        format: props.schema.format,
        schema: props.schema.schema,
        data: isPreview ? current.content : btoa(current.content ?? ''),
        contentType: current.contentType
    }
    const query = state.source.binary && other
        ? `?outputFormat=${encodeURIComponent(other.contentType)}`
        : ''

    try {
        const res = await fetch(transformPath('/api/schema/validate' + query), {
            method: 'POST',
            body: JSON.stringify(body),
            headers: { 'Content-Type': 'application/json' }
        })
        const isJson = res.headers.get('content-type')?.includes('application/json')
        const data = isJson ? await res.json() : await res.text()
        state.validated = true

        if (res.status === 400) {
            state.errors = ''
            state.result = data
            return
        }
        if (!res.ok) {
            state.errors = typeof data === 'string' ? data : JSON.stringify(data)
            state.result = null
            return
        }

        state.errors = ''
        state.result = null
        const first = (data as Example[])[0]
        if (!first) return

        const decoded = atob(first.value)
        if (isPreview && state.source.binary) {
            state.source.binary.content = decoded
        } else if (state.source.preview) {
            state.source.preview.content = formatLanguage(decoded, first.contentType)
        }
    } catch (e) {
        state.validated = true
        state.result = null
        state.errors = e instanceof Error ? e.message : String(e)
    } finally {
        state.validating = false
    }
}

function update(event: { content: string, type: string }) {
    if (event.type === 'preview' && state.source.preview) {
        state.source.preview.content = event.content
    }
    if (event.type === 'binary' && state.source.binary) {
        state.source.binary.content = event.content
    }
}

function switchSource(e: 'preview' | 'binary') {
    state.view = e
}

function formatResult(result: any): string {
    const [html, count] = formatResultItem(result)
    return `<div>error count ${count}: ${html}</div>`
}

function formatResultItem(result: any): [string, number] {
    if (Array.isArray(result)) {
        let html = '<ul>'
        let sum = 0
        for (const item of result) {
            const [msg, count] = formatResultItem(item)
            html += `<li>${msg}</li>`
            sum += count
        }
        return [html + '</ul>', sum]
    }
    if (result !== null && typeof result === 'object') {
        const head = `${result.schema}: ${result.message}`
        if (Array.isArray(result.errors)) {
            const [list, count] = formatResultItem(result.errors)
            return [head + list, count]
        }
        return [head, 1]
    }
    return [result, 1]
}
</script>

<template>
    <div>
        <button type="button" class="btn btn-primary btn-sm" data-bs-toggle="modal" :data-bs-target="'#'+id">{{ exampleEnabled ? 'Example & Validate' : 'Validate' }}</button>
        <div ref="modal" class="modal fade" :id="id" tabindex="-1" aria-hidden="true" :aria-labelledby="id+'title'">
            <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
                <div class="modal-content">
                    <div class="modal-header">
                        <h6 :id="id+'title'" class="modal-title">{{ title }}</h6>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                    </div>
                    <div class="modal-body">
                        <div class="alert alert-danger errors" role="alert" v-if="state.errors.length > 0">{{ state.errors }}</div>
                        <div class="alert alert-danger" role="alert" v-if="state.result">
                            <span v-html="formatResult(state.result)"></span>
                        </div>
                        <source-view :filename="source?.filename" v-model:source="state.source" :readonly="false" @update="update" @switch="switchSource" />
                    </div>
                    <div class="modal-footer justify-content-between">
                        <span class="float-start">
                            <span class="status" role="status">
                                <span v-if="state.validating" class="loading"><span class="bi bi-arrow-repeat spin"></span> Validating...</span>
                                <span v-else-if="state.errors.length == 0 && state.result == null && state.validated" class="valid bi bi-check-circle-fill fade-in"> Valid</span>
                                <span v-else-if="state.validated" class="failed bi bi-x-circle-fill fade-in"> Failed</span>
                            </span>
                        </span>
                        <div class="float-end">
                            <button v-if="exampleEnabled" type="button" class="btn btn-primary btn-sm" @click="setExample()">Example</button>
                            <button type="button" class="btn btn-primary btn-sm" :disabled="state.validating" @click="validate()">Validate</button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style>
.modal-body .errors {
    white-space: pre-line;
}
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