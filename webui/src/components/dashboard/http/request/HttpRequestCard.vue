<script setup lang="ts">
import { computed, onMounted, reactive, } from 'vue'
import HttpParameters from './HttpParameters.vue'
import SchemaExpand from '../../SchemaExpand.vue'
import SchemaValidate from '../../SchemaValidate.vue'
import SourceView from '../../SourceView.vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import { Tooltip } from 'bootstrap';

declare type Tab = 'RequestBody' | 'Parameters' | 'Security'
const { formatSchema } = usePrettyLanguage()

const props = defineProps<{
    operation: HttpOperation
    path: string
}>()
const selected = reactive({
    content: {} as HttpMediaType | null,
})

if (props.operation.requestBody?.contents?.length > 0){
    selected.content = props.operation.requestBody.contents[0]!
}
function selectedContentChange(event: any){
    for (let content of props.operation.requestBody.contents){
        if (content.type == event.target.value){
            selected.content = content
        }
    }
}

onMounted(() => {
    const tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]')
    const tooltipList = [...tooltipTriggerList].map(tooltipTriggerEl => new Tooltip(tooltipTriggerEl))
})

const name = computed(() => {
    const segments = props.path.split('/').reverse()
    for (const seg of segments) {
        if (seg === '') {
            continue
        }
        if (!seg.startsWith('{')) {
            return seg
        }
    }
    return ''
})
const schemes = computed(() => {
    if (!props.operation.security) {
        return null
    }

    const schemes: {[name: string]: HttpSecurityScheme & { name: string} } = {}
    for (let i = 0; i < props.operation.security.length; i++) {
        const item = props.operation.security[i]!
        const keys = Object.keys(item)
        for (let j = 0; j < keys.length; j++) {
            const schemeName = keys[j]!
            schemes[`${i}-${j}`] = {  ...item[schemeName], ...{ name: schemeName } } as HttpSecurityScheme & { name: string} 
        }
    }
    return schemes
})
const hasOneSecurityScheme = computed(() => {
    if (!props.operation.security) {
        return false
    }
    if (props.operation.security.length === 1) {
        return Object.keys(props.operation.security[0]!).length === 1
    }
    return false
})
const activeTab = computed<Tab>(() => {
    if (props.operation.requestBody) {
        return 'RequestBody';
    }
    if (props.operation.parameters || !props.operation.security) {
        return 'Parameters';
    }
    return 'Security'
})
function getSchemeClass(scheme: HttpSecurityScheme) {
    switch (scheme.configs['type']) {
        case 'http':
            return scheme.configs.scheme
        default:
            return scheme.configs.type
    }
}
</script>

<template>
    <section class="card" aria-labelledby="request">
        <div class="card-body">
            <h2 id="request" class="card-title text-center">Request</h2>

            <div class="nav card-tabs" role="tablist" data-testid="tabs">
              <button :class="activeTab === 'RequestBody' ? 'active' : 'disabled'" id="body-tab" :aria-disabled="!operation.requestBody" data-bs-toggle="tab" data-bs-target="#body" type="button" role="tab" aria-controls="body" :aria-selected="activeTab === 'RequestBody'"><span class="bi-file-text me-2" />Body</button>
              <button :class="{ active: activeTab === 'Parameters' }" id="parameters-tab" data-bs-toggle="tab" data-bs-target="#parameters" type="button" role="tab" aria-controls="parameters" :aria-selected="activeTab === 'Parameters'"><span class="bi-sliders me-2" />Parameters</button>
              <button :class="{ active: activeTab === 'Security', disabled: !operation.security }" id="security-tab" :aria-disabled="!operation.security" data-bs-toggle="tab" data-bs-target="#security" type="button" role="tab" aria-controls="security" :aria-selected="activeTab === 'Security'"><span class="bi-shield-lock me-2" /> Security</button>
            </div>

            <div class="tab-content" id="tabRequest">
                <div class="tab-pane fade" :class="operation.requestBody ? 'show active' : ''" id="body" role="tabpanel" aria-labelledby="body-tab" v-if="operation.requestBody">
                    <div v-if="operation.requestBody">
                        <div class="row mb-2" v-if="operation.requestBody.description">
                            <div class="col">
                                <p id="request-body-description" class="label">Description</p>
                                <p aria-labelledby="request-body-description">{{  operation.requestBody.description }}</p>
                            </div>
                        </div>
                        <div class="row mb-2">
                            <div class="col-2">
                                <p id="request-body-content-type" class="label" v-if="operation.requestBody.contents.length == 1">Request content type</p>
                                <p aria-labelledby="request-body-content-type" v-if="operation.requestBody.contents.length == 1">{{ operation.requestBody.contents[0]?.type }}</p>
                            </div>
                            <div class="col">
                                <p id="request-body-required" class="label">Required</p>
                                <p aria-labelledby="request-body-required">{{ operation.requestBody.required }}</p>
                            </div>
                        </div>
                        
                        <source-view 
                            :source="{ preview: { content: formatSchema(selected.content?.schema), contentType: 'application/json' }}" 
                            :deprecated="selected.content?.schema.deprecated" 
                            :hide-content-type="true"
                            :height="250" class="mb-2">
                        </source-view>

                        <div class="row">
                            <div class="col-auto pe-2" v-if="selected.content">
                                <schema-expand :schema="selected.content.schema" />
                            </div>
                            <div class="col-auto px-2" v-if="selected.content">
                                <schema-validate :source="{ preview: { content: '', contentType: selected.content.type} }" :schema="{schema: selected.content.schema, format: 'application/vnd.oai.openapi+json;version=3.0.0'}" :name="name" />
                            </div>
                            <div class="col-auto px-2">
                                <select v-if="operation.requestBody.contents.length > 1" class="form-select form-select-sm" aria-label="Request content type" @change="selectedContentChange">
                                    <option v-for="content in operation.requestBody.contents">{{ content.type }}</option>
                                </select>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="tab-pane fade" :class="!operation.requestBody ? 'show active' : ''" id="parameters" role="tabpanel" aria-labelledby="parameters-tab">
                    <http-parameters :parameters="operation.parameters" />
                </div>
                <div class="tab-pane fade security" :class="!operation.security ? 'show active' : ''" id="security" role="tabpanel" aria-labelledby="security-tab">
                    <div class="d-flex">
                        <ul id="security-list" class="list-group nav nav-pills" v-if="!hasOneSecurityScheme">
                            <li v-for="(sec, index) of operation.security" class="list-group-item">
                                <button class="badge" v-for="(name, index2) of Object.keys(sec)" type="button" role="tab"
                                    :id="`sec-scheme-${index}-${index2}-tab`" :class="getSchemeClass(sec[name]!) + (index === 0 && index2 === 0 ? ' active' : '')"
                                    data-bs-toggle="pill" :data-bs-target="`#sec-scheme-${index}-${index2}`"
                                :aria-controls="`sec-scheme-${index}-${index2}`" :aria-selected="index === 0 && index2 === 0"
                                >
                                    {{ name + ' (' + sec[name]?.configs['type'] + ')' }}
                                </button>
                            </li>
                        </ul>
                        <div class="tab-content ms-3 ps-3 w-100">
                            <div v-for="(scheme, name) of schemes"
                                class="tab-pane fade"
                                :class="name === '0-0' ? 'show active' : ''" :id="`sec-scheme-${name}`" role="tabpanel" :aria-labelledby="`sec-scheme-${name}-tab`"
                            >
                                <div v-if="scheme.configs.type.toUpperCase() === 'APIKEY'">
                                    <div class="row">
                                        <div class="col-2">
                                            <p :id="`security-${name}-name`" class="label">Name</p>
                                            <p :aria-labelledby="`security-${name}-name`">{{ scheme.name }}</p>
                                        </div>
                                        <div class="col">
                                            <p :id="`security-${name}-type`" class="label">Type</p>
                                            <p :aria-labelledby="`security-${name}-name`">{{ scheme.configs.type }}</p>
                                        </div>
                                    </div>
                                    <div class="row">
                                        <div class="col-2">
                                            <p :id="`security-${name}-param-name`" class="label">Parameter Name</p>
                                            <p :aria-labelledby="`security-${name}-param-name`">{{ scheme.configs.name }}</p>
                                        </div>
                                        <div class="col">
                                            <p :id="`security-${name}-location`" class="label">Location</p>
                                            <p :aria-labelledby="`security-${name}-location`">{{ scheme.configs.in }}</p>
                                        </div>
                                    </div>
                                </div>

                                <div v-if="scheme.configs.type.toUpperCase() === 'OAUTH2'">
                                    <div class="row">
                                        <div class="col-2">
                                            <p :id="`security-${name}-name`" class="label">Name</p>
                                            <p :aria-labelledby="`security-${name}-name`">{{ scheme.name }}</p>
                                        </div>
                                        <div class="col">
                                            <p :id="`security-${name}-type`" class="label">Type</p>
                                            <p :aria-labelledby="`security-${name}-type`">{{ scheme.configs.type }}</p>
                                        </div>
                                    </div>
                                    <div class="row">
                                        <div class="col">
                                            <div v-if="scheme.scopes.length > 0">
                                                <p :id="`security-${name}-scopes`" class="label">Scopes</p>
                                                <p :aria-labelledby="`security-${name}-scopes`">{{ scheme.scopes.join(', ') }}</p>
                                            </div>
                                        </div>
                                    </div>
                                    <div class="row">
                                        <div class="col">
                                            <p :id="`security-${name}-flows`" class="label">Flows</p>
                                            <table class="table dataTable" :aria-labelledby="`security-${name}-flows`">
                                                <thead>
                                                    <tr>
                                                        <th scope="col" class="text-left" style="width: 15%">Type</th>
                                                        <th scope="col" class="text-left">Scopes</th>
                                                        <th scope="col" class="text-left">Authorization URL</th>
                                                        <th scope="col" class="text-left">Token URL</th>
                                                        <th scope="col" class="text-left">Refresh URL</th>
                                                    </tr>
                                                </thead>
                                                <tbody>
                                                    <tr v-for="(flow, name) of scheme.configs.flows">
                                                        <td>{{ name }}</td>
                                                        <td>
                                                            <div v-for="scope of Object.keys(flow.scopes)" :data-bs-title="flow.scopes[scope]" data-bs-toggle="tooltip" data-bs-placement="top">
                                                                {{ scope }}
                                                            </div>
                                                        </td>
                                                        <td>{{ flow.authorizationUrl }}</td>
                                                        <td>{{ flow.TokenUrl }}</td>
                                                        <td>{{ flow.RefreshUrl }}</td>
                                                    </tr>
                                                </tbody>
                                            </table>
                                        </div>
                                    </div>
                                </div>

                                <div v-if="scheme.configs.type.toUpperCase() === 'HTTP' && scheme.configs.scheme.toUpperCase() === 'BASIC'">
                                    <div class="row">
                                        <div class="col-2">
                                            <p :id="`security-${name}-name`" class="label">Name</p>
                                            <p :aria-labelledby="`security-${name}-name`">{{ scheme.name }}</p>
                                        </div>
                                        <div class="col">
                                            <p :id="`security-${name}-type`" class="label">Type</p>
                                            <p :aria-labelledby="`security-${name}-type`">{{ scheme.configs.type }}</p>
                                        </div>
                                    </div>
                                    <div class="row">
                                        <div class="col-2">
                                            <p :id="`security-${name}-scheme`" class="label">Scheme</p>
                                            <p :aria-labelledby="`security-${name}-scheme`">{{ scheme.configs.scheme }}</p>
                                        </div>
                                    </div>
                                </div>

                                <div v-if="scheme.configs.type.toUpperCase() === 'HTTP' && scheme.configs.scheme.toUpperCase() === 'BEARER'">
                                    <div class="row">
                                        <div class="col-2">
                                            <p :id="`security-${name}-name`" class="label">Name</p>
                                            <p :aria-labelledby="`security-${name}-name`">{{ scheme.name }}</p>
                                        </div>
                                        <div class="col">
                                            <p :id="`security-${name}-type`" class="label">Type</p>
                                            <p :aria-labelledby="`security-${name}-type`">{{ scheme.configs.type }}</p>
                                        </div>
                                    </div>
                                    <div class="row">
                                        <div class="col-2">
                                            <p :id="`security-${name}-scheme`" class="label">Scheme</p>
                                            <p :aria-labelledby="`security-${name}-scheme`">{{ scheme.configs.scheme }}</p>
                                        </div>
                                        <div class="col-2">
                                            <p :id="`security-${name}-format`" class="label">Format</p>
                                            <p :aria-labelledby="`security-${name}-format`">{{ scheme.configs.bearerFormat }}</p>
                                        </div>
                                    </div>
                                </div>

                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </section>
</template>

<style>

.security button.badge {
    font-size: 0.75rem;
    border-color: var(--color-datatable-border);
    border: 0;
}
.security button.badge:not(:last-child) {
    margin-bottom: 5px;
}
.security button.badge.active {
    font-size: 0.75rem;
    border: 0;
}
.security button.badge.oauth2 {
    --color: 13, 110, 253;
    background-color: rgb(var(--color));
}
.security button.badge.apiKey {
    --color: 25, 135, 84;
    background-color: rgb(var(--color));
}
.security button.badge.basic {
    --color: 255, 193, 7;
    background-color: rgb(var(--color));
}
.security button.badge.bearer {
    --color: 13, 110, 253;
    background-color: rgb(var(--color));
}
.security button.badge.active {
    outline-width: 3px;
    outline-color: rgba(var(--color), 0.5);
    outline-style: solid;
}
.security .tab-pane {
    padding: 0;
}
.security .tab-content {
    min-width: 0;
}
.security .tab-content:not(:first-child) {
    border-style: solid;
    border-width: 0;
    border-left-width: 2px;
    border-color: var(--color-datatable-border);
}
.security .row {
    margin-bottom: 0.8rem;
}
</style>

