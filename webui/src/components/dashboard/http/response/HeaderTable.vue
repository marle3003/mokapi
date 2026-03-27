<script setup lang="ts">
import type { PropType } from 'vue'
import { useSchema } from '@/composables/schema'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import SourceView from '../../SourceView.vue'
import { useMarkdown } from '@/composables/markdown'

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
                <td>
                    <span role="button" @click.stop data-bs-toggle="modal" :data-bs-target="'#modal-'+header.name" :title="'Read more about header '+header.name" tabindex="0">
                        {{ header.name }}
                    </span>
                </td>
                <td>{{ printType(header.schema) }}</td>
                <td><div v-html="useMarkdown(header.description).content" class="description"></div></td>
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
                        <div v-html="useMarkdown(header.description).content"></div>
                        <p class="label">Schema</p>
                        <div class="codeBlock">
                            <source-view 
                                :source="{ preview: { content: formatSchema(header.schema), contentType: 'application/json' }}" 
                                :deprecated="header.schema?.deprecated" 
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