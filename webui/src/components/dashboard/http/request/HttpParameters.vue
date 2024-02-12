<script setup lang="ts">
import type { PropType, Ref } from 'vue'
import { useSchema } from '@/composables/schema'
import Markdown from 'vue3-markdown-it'
import { useExample } from '@/composables/example'
import SourceView from '../../SourceView.vue'

const {printType} = useSchema()
const { fetchExample } = useExample()

const props = defineProps({
    parameters: { type: Array as PropType<Array<HttpParameter>> }
})

const examples: {[key: string]: Ref<string | undefined>} = {}
if (props.parameters){
    for (let parameter of props.parameters){
        let example = fetchExample(parameter.schema, 'text/plain')
        examples[parameter.name+parameter.type] = example
    }
}

const sortedParameters = props.parameters?.sort((p1, p2) =>{
    const r = getParameterTypeSortOrder(p1.type) - getParameterTypeSortOrder(p2.type)
    if (r != 0) {
        return r
    }
    return p1.name.localeCompare(p2.name)
})
function getParameterTypeSortOrder(type: string): number{
    switch (type){
        case "path": return 0
        case "query": return 1
        case "header": return 2
        case "cookie": return 3
        default: return 4
    }
}
function getExample(key: string){
    const example = examples[key]
    if (!example.value){
        return ''
    }
    return example.value
}

function showWarningColumn(){
    if (!props.parameters){
        return false
    }
    for (let parameter of props.parameters){
        if (parameter.deprecated){
            return true
        }
    }
    return false
}
</script>

<template>
    <table class="table dataTable selectable">
        <thead>
            <tr>
                <th scope="col" class="text-left">Name</th>
                <th scope="col" class="text-left">Location</th>
                <th scope="col" class="text-left">Type</th>
                <th scope="col" class="text-left">Required</th>
                <th scope="col" class="text-left" v-if="showWarningColumn()">Warning</th>
                <th scope="col" class="text-left">Description</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="parameter in sortedParameters" :key="parameter.name + parameter.type" data-bs-toggle="modal" :data-bs-target="'#modal-'+parameter.name+parameter.type">
                <td>{{ parameter.name }}</td>
                <td>{{ parameter.type }}</td>
                <td>{{ printType(parameter.schema) }}</td>
                <td>{{ parameter.required }}</td>
                <td v-if="showWarningColumn()">
                    <span v-if="parameter.deprecated"><i class="bi bi-exclamation-triangle-fill yellow"></i> deprecated</span>
                </td>
                <td><markdown :source="parameter.description" class="description"></markdown></td>
            </tr>
        </tbody>
    </table>
    <div v-for="parameter in parameters" :key="parameter.name+parameter.type">
        <div class="modal fade" :id="'modal-'+parameter.name+parameter.type" tabindex="-1" aria-hidden="true">
            <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
                <div class="modal-content">
                    <div class="modal-body">
                        <div class="card-group">
                            <div class="card">
                                <div class="card-body">
                                    <div class="row">
                                        <div class="col">
                                            <p class="label">Name</p>
                                            <p>{{ parameter.name }}</p>
                                        </div>
                                        <div class="col">
                                            <p class="label">Location</p>
                                            <p>{{ parameter.type }}</p>
                                        </div>
                                        <div class="col">
                                            <p v-if="parameter.deprecated"><i class="bi bi-exclamation-triangle-fill yellow"></i> Deprecated</p>
                                        </div>
                                    </div>
                                    <div class="row mt-2">
                                        <div class="col">
                                            <p class="label">Required</p>
                                            <p>{{ parameter.required }}</p>
                                        </div>
                                        <div class="col">
                                            <p class="label">Style</p>
                                            <p>{{ parameter.style ?? 'simple' }}</p>
                                        </div>
                                        <div class="col">
                                            <p class="label">Explode</p>
                                            <p>{{ parameter.explode ?? false }}</p>
                                        </div>
                                    </div>
                                    <div class="row mt-2">
                                        <p class="label">Description</p>
                                        <markdown :source="parameter.description"></markdown>
                                    </div>
                                    <div class="row">
                                        <ul class="nav nav-pills tab-sm tab-params" role="tabList">
                                            <li class="nav-link show active" :id="'pills-body-tab'+parameter.name+parameter.type" data-bs-toggle="pill" :data-bs-target="'#pills-schema'+parameter.name+parameter.type" type="button" role="tab" :aria-controls="'pills-schema'+parameter.name+parameter.type" aria-selected="true">Schema</li>
                                            <li class="nav-link" :id="'pills-header-tab'+parameter.name+parameter.type" data-bs-toggle="pill" :data-bs-target="'#pills-example'+parameter.name+parameter.type" type="button" role="tab" :aria-controls="'pills-example'+parameter.name+parameter.type" aria-selected="false">Example</li>
                                        </ul>

                                        <div class="tab-content" :id="'pills-tabParameter'+parameter.name+parameter.type">
                                            <div class="tab-pane fade show active" :id="'pills-schema'+parameter.name+parameter.type" role="tabpanel">
                                                <source-view :source="JSON.stringify(parameter.schema)" content-type="application/json" :hide-content-type="true" />
                                            </div>
                                            <div class="tab-pane fade" :id="'pills-example'+parameter.name+parameter.type" role="tabpanel">
                                                <source-view :source="getExample(parameter.name+parameter.type)" content-type="text/plain" :hide-content-type="true" />
                                            </div>
                                        </div>
                                    </div>
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
.nav-pills li:first-child {
    padding-left: 12px;
}
.tab-sm.tab-params li + li::before {
    padding-left: 6px;
    padding-right: 6px;
}
.tab-pane {
    padding: 0;
}
</style>