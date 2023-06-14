<script setup lang="ts">
import { usePrettyHttp } from '@/composables/http';
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import { type PropType, reactive, computed } from 'vue';
import HeaderTable from './HeaderTable.vue';
import SchemaExpand from '../../SchemaExpand.vue';
import SchemaExample from '../../SchemaExample.vue';
import Markdown from 'vue3-markdown-it'

const props = defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
    path: { type: Object as PropType<HttpPath>, required: true },
    operation: { type: Object as PropType<HttpOperation>, required: true }
})

const {formatStatusCode, getClassByStatusCode} = usePrettyHttp()
const {formatLanguage} = usePrettyLanguage()

const selected = reactive({
    contents: {} as  { [statusCode: number]: HttpMediaType}
})

const responses = computed(() => {
    return props.operation.responses.sort((r1, r2) => r1.statusCode - r2.statusCode)
})

for (let response of responses.value) {
    if (!response.contents) {
        continue
    }
    selected.contents[response.statusCode] = response.contents[0]
}

function selectedContentChange(event: any, statusCode: number){
    for (let response of props.operation.responses){
        if (response.statusCode == statusCode){
            for (let content of response.contents){
                if (content.type == event.target.value){
                selected.contents[statusCode] = content
                }
            }
        }
    }
}
</script>

<template>
    <div class="card response" data-testid="http-response">
        <div class="card-body">
            <div class="card-title text-center">Response</div>
            <div class="d-flex align-items-start align-items-stretch">
                <div class="nav flex-column nav-pills" id="v-pills-tab" role="tablist" aria-orientation="vertical">
                    <button v-for="(response, index) of operation.responses" class="badge status-code" :class="getClassByStatusCode(response.statusCode) + (index==0 ? ' active' : '')" :id="'v-pills-'+response.statusCode+'-tab'" data-bs-toggle="pill" :data-bs-target="'#v-pills-'+response.statusCode" type="button" role="tab" :aria-controls="'v-pills-'+response.statusCode" aria-selected="true">{{ formatStatusCode(response.statusCode) }}</button>
                </div>
                <div class="tab-content ms-3 ps-3 responses-tab" style="width: 100%" id="v-pills-tabContent">
                    <div v-for="(response, index) of responses" class="tab-pane fade" :class="index==0 ? 'show active' : ''" :id="'v-pills-'+response.statusCode" role="tabpanel" :aria-labelledby="'v-pills-'+response.statusCode+'-tab'">
                        <p class="label">Description</p>
                        <p><markdown :source="response.description" :data-testid="'response-description-'+response.statusCode" aria-label="Response body description"></markdown></p>
                        <div v-if="response.contents || response.headers">
                            <ul class="nav nav-pills response-tab" role="tabList">
                                <li class="nav-link" id="pills-body-tab" :class="response.contents ? 'active' : 'disabled'" data-bs-toggle="pill" data-bs-target="#pills-body" type="button" role="tab" aria-controls="pills-body" aria-selected="true">Body</li>
                                <li class="nav-link" id="pills-header-tab" :class="response.headers ? (response.contents ? '' : 'active') : 'disabled'" data-bs-toggle="pill" data-bs-target="#pills-header" type="button" role="tab" aria-controls="pills-header" aria-selected="false">Headers</li>
                            </ul>
                            <div class="tab-content" id="pills-tabContent">
                                <div v-if="response.contents" class="tab-pane fade" :class="response.contents ? 'show active' : ''" id="pills-body" role="tabpanel" aria-labelledby="pills-body-tab">
                                    <p class="codeBlock">
                                        <span v-if="selected.contents[response.statusCode]" class="label">{{ selected.contents[response.statusCode].type }}</span>
                                        <span v-if="selected.contents[response.statusCode].schema?.deprecated" class="warning"><i class="bi bi-exclamation-triangle-fill yellow"></i> Deprecated</span>
                                        <pre v-highlightjs="formatLanguage(JSON.stringify(selected.contents[response.statusCode].schema), 'application/json')" class="overflow-auto" style="max-height: 250px;"><code class="json"></code></pre>
                                    </p>
                                    <div class="row">
                                        <div class="col-auto pe-2">
                                            <schema-expand :schema="selected.contents[response.statusCode].schema" />
                                        </div>
                                        <div class="col-auto px-2">
                                            <schema-example :schema="selected.contents[response.statusCode].schema" />
                                        </div>
                                        <div class="col-auto px-2">
                                            <select v-if="response.contents.length > 0" class="form-select form-select-sm" aria-label="Response content type" @change="selectedContentChange($event, response.statusCode)">
                                                <option v-for="content in response.contents">{{ content.type }}</option>
                                            </select>
                                        </div>
                                    </div>
                                </div>
                                <div v-if="response.headers" class="tab-pane fade" :class="!response.contents ? 'show active' : ''" id="pills-header" role="tabpanel" aria-labelledby="pills-header-tab">
                                    <header-table :headers="response.headers" />
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.response .tab-content.responses-tab {
    border-style: solid;
    border-width: 0;
    border-left-width: 2px;
    border-color: var(--color-datatable-border);
    min-width: 0;
}

.response .nav.nav-pills button.badge {
    font-size: 0.75rem;
    border-color: var(--color-datatable-border);
    border: 0;
    margin-bottom: 15px;
}
.response .nav.nav-pills button.badge.active {
    font-size: 0.75rem;
    border: 0;
}
.codeBlock {
  position: relative;
}
.card-body .codeBlock .label {
    position: absolute;
    right: 20px;
    top: 4px;
    font-size: 0.8rem;
}
.card-body .codeBlock .warning {
    position: absolute;
    text-align: center;
    top: 4px;
    font-size: 1.0rem;
    width: 100%
}
.response-tab li {
    padding-left: 0;
    padding-right: 0;
}
.response-tab li.active {
    background-color: transparent;
    text-shadow: none;
}

.response-tab li.disabled {
    color: var(--color-disabled);
}
.response-tab li + li::before {
    padding-right: 7px;
    padding-left: 7px;
    content: " | ";
}
.response .tab-pane {
    padding-left: 0;
    padding-top: 0.5rem;
 }
</style>