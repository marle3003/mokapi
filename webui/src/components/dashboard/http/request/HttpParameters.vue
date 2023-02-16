<script setup lang="ts">
import type { PropType } from 'vue';
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import { useSchema } from '@/composables/schema';
import Markdown from 'vue3-markdown-it';

const {formatLanguage} = usePrettyLanguage()
const {printType} = useSchema()

defineProps({
    parameters: { type: Array as PropType<Array<HttpParameter>> }
})
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
                <td>{{ parameter.description }}</td>
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
                                        <p class="label">Schema</p>
                                        <pre v-highlightjs="formatLanguage(JSON.stringify(parameter.schema), 'application/json')"><code class="json"></code></pre>
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