<script setup lang="ts">
import type { PropType, Ref } from 'vue';
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import { useSchema } from '@/composables/schema';
import Markdown from 'vue3-markdown-it';
import { useExample } from '@/composables/example';
import hljs from 'highlight.js'

const {formatLanguage} = usePrettyLanguage()
const {printType} = useSchema()
const {fetchExample} = useExample()

const props = defineProps({
    parameters: { type: Array as PropType<Array<HttpParameter>> }
})

const examples: {[key: string]: Ref<string | undefined>} = {}
if (props.parameters){
    for (let parameter of props.parameters){
        let example = fetchExample(parameter.schema)
        examples[parameter.name+parameter.type] = example
    }
}

function getExample(key: string){
    const example = examples[key]
    if (!example.value){
        return ''
    }
    return hljs.highlightAuto(example.value).value
}

</script>

<template>
    <table class="table dataTable selectable">
        <thead>
            <tr>
                <th scope="col" class="text-left">Name</th>
                <th scope="col" class="text-left">Location</th>
                <th scope="col" class="text-left">Type</th>
                <th scope="col" class="text-left">Description</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="parameter in parameters" :key="parameter.name + parameter.type" data-bs-toggle="modal" :data-bs-target="'#modal-'+parameter.name+parameter.type">
                <td>{{ parameter.name }}</td>
                <td>{{ parameter.type }}</td>
                <td>{{ printType(parameter.schema) }}</td>
                <td><markdown :source="parameter.description" class="description"></markdown></td>
            </tr>
        </tbody>
    </table>
    <div v-for="parameter in parameters" :key="parameter.name+parameter.type">
        <div class="modal fade" :id="'modal-'+parameter.name+parameter.type" tabindex="-1" aria-hidden="true">
            <div class="modal-dialog modal-dialog-centered modal-dialog-scrollable">
                <div class="modal-content">
                    <div class="modal-body">
                        <div class="card-group">
                            <div class="card">
                                <div class="card-body">
                                    <p class="label">Name</p>
                                    <p>{{ parameter.name }}</p>
                                    <p class="label">Description</p>
                                    <markdown :source="parameter.description"></markdown>
                                </div>
                            </div>
                        </div>
                        <div class="card-group">
                            <div class="card">
                                <div class="card-body">
                                    <div class="card-title text-center">Query</div>
                                    <div class="row">
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
                                    <div class="row">
                                        <ul class="nav nav-pills schema-tab" role="tabList">
                                            <li class="nav-link show active" :id="'pills-body-tab'+parameter.name+parameter.type" data-bs-toggle="pill" :data-bs-target="'#pills-schema'+parameter.name+parameter.type" type="button" role="tab" :aria-controls="'pills-schema'+parameter.name+parameter.type" aria-selected="true">Schema</li>
                                            <li class="nav-link" :id="'pills-header-tab'+parameter.name+parameter.type" data-bs-toggle="pill" :data-bs-target="'#pills-example'+parameter.name+parameter.type" type="button" role="tab" :aria-controls="'pills-example'+parameter.name+parameter.type" aria-selected="false">Example</li>
                                        </ul>

                                        <div class="tab-content" :id="'pills-tabParameter'+parameter.name+parameter.type">
                                            <div class="tab-pane fade show active" :id="'pills-schema'+parameter.name+parameter.type" role="tabpanel">
                                                <pre v-highlightjs="formatLanguage(JSON.stringify(parameter.schema), 'application/json')"><code class="json"></code></pre>
                                            </div>
                                            <div class="tab-pane fade" :id="'pills-example'+parameter.name+parameter.type" role="tabpanel">
                                                <pre><code class="json hljs" v-html="getExample(parameter.name+parameter.type)"></code></pre>
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
.schema-tab li {
    padding-right: 0;
}
.schema-tab li:not(:first-child) {
    padding-left: 0;
}
.schema-tab li + li::before {
    padding-right: 7px;
    padding-left: 7px;
    content: " | ";
}
</style>