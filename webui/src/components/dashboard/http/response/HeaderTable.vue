<script setup lang="ts">
import type { PropType } from 'vue'
import { useSchema } from '@/composables/schema'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import Markdown from 'vue3-markdown-it'
import SourceView from '../../SourceView.vue'

defineProps({
    headers: { type: Array as PropType<Array<HttpParameter>> },
})

const { printType } = useSchema()
const { formatSchema } = usePrettyLanguage()
</script>

<template>
    <table class="table dataTable selectable">
        <thead>
            <tr>
                <th scope="col" class="text-left" style="width: 15%">Name</th>
                <th scope="col" class="text-left">Schema</th>
                <th scope="col" class="text-left">Description</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="header in headers!" :key="header.name" data-bs-toggle="modal" :data-bs-target="'#modal-'+header.name">
                <td>{{ header.name }}</td>
                <td>{{ printType(header.schema) }}</td>
                <td><markdown :source="header.description" class="description" :html="true"></markdown></td>
            </tr>
        </tbody>
    </table>
    <div v-for="header in headers!" :key="header.name">
        <div class="modal fade" :id="'modal-'+header.name" tabindex="-1" aria-hidden="true">
            <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
                <div class="modal-content">
                    <div class="modal-body">
                        <p class="label">Name</p>
                        <p>{{ header.name }}</p>
                        <p class="label">Description</p>
                        <markdown :source="header.description"></markdown>
                        <p class="label">Schema</p>
                        <div class="codeBlock">
                            <source-view 
                                :source="formatSchema(header.schema)" 
                                :deprecated="header.schema.deprecated" 
                                content-type="application/json"
                                :hide-content-type="true"
                                :height="250" class="mb-2">
                            </source-view>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>