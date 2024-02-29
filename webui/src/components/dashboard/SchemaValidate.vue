<script setup lang="ts">
import { useGuid } from '@/composables/guid'
import { useExample } from '@/composables/example'
import { watch, reactive, computed } from 'vue'
import { VAceEditor } from 'vue3-ace-editor'
import ace from 'ace-builds'
import themeGithubUrl from 'ace-builds/src-noconflict/theme-github_dark?url'

import modeJsonUrl from 'ace-builds/src-noconflict/mode-json?url';
import { transformPath } from '@/composables/fetch'
ace.config.setModuleUrl('ace/mode/json', modeJsonUrl);

ace.config.setModuleUrl('ace/theme/github_dark', themeGithubUrl)

const props = withDefaults(defineProps<{
  schema: Schema
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
var example = fetchExample(props.schema, props.contentType)
watch(() => props.schema, (schema) => {
    example = fetchExample(schema, props.contentType)
})

function setExample() {
    updateContent(example.value!)
}

function validate() {
    const body = {
        schema: props.schema,
        data: state.content
    }
    console.log(state.content)
    fetch(transformPath('/api/schema/validate'), { method: 'POST', body: JSON.stringify(body), headers: { 'Data-Content-Type': props.contentType, 'Content-Type': 'application/json' } })
    .then(res => {
        if (res.status === 200){
            console.log('Valid')
            return ''
        }else {
            console.log('not valid')
            return res.text()
        }
    }).then(errors => {
        if (errors === '') {
            state.errors = []
        } else{
            state.errors = errors.split('\n').filter(i => i).slice(1)
        }
        state.validated = true
    })
}
const editorHeight = computed(() => Math.max(500, state.content.split('\n').length * 19) + 'px')

function updateContent(event: string) {
    state.content = event
    state.validated = false
}
</script>

<template>
    <div>
        <button type="button" class="btn btn-primary btn-sm" data-bs-toggle="modal" :data-bs-target="'#'+id">Validate</button>
        <div class="modal fade" :id="id" tabindex="-1"  aria-hidden="true" :aria-labelledby="id+'title'">
            <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
                <div class="modal-content">
                    <div class="modal-header">
                        <h6 :id="id+'title'" class="modal-title">{{ title }}</h6>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                    </div>
                    <div class="modal-body">
                        <div class="alert alert-danger" role="alert" v-if="state.errors.length > 0 && state.validated">
                            <ul class="error-list">
                                <li v-for="err in state.errors">{{ err }}</li>
                            </ul>
                        </div>
                        <v-ace-editor
                            :value="state.content"
                            @update:value="updateContent($event)"
                            lang="json"
                            theme="github_dark"
                            style="font-size: 16px;" :style="'height: '+editorHeight"/>
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