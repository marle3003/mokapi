<script setup lang="ts">
import type { PropType, } from 'vue';
import HttpRequestBodyDialog from './HttpRequestBodyDialog.vue';
import Markdown from 'vue3-markdown-it';
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';

const {formatLanguage} = usePrettyLanguage()

const props = defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
    path: { type: Object as PropType<HttpPath>, required: true },
    operation: { type: Object as PropType<HttpOperation>, required: true }
})
function getType(schema: Schema) {
    if (schema.type == 'array'){
        return `${schema.items.type}[]`
    }
    return schema.type
}
let selectContent: HttpMediaType | null = null
if (props.operation.requestBody.contents.length > 0){
    selectContent = props.operation.requestBody.contents[0]
}
function selectedContentChange(event: any){
    for (let content of props.operation.requestBody.contents){
        if (content.type == event.target.value){
            selectContent = content
        }
    }
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Request Body</div>
            <p class="label">Description</p>
            <p><markdown :source="operation.requestBody.description"></markdown></p>
            <p class="label">Body</p>
            <p class="codeBlock" v-if="selectContent">
                <span v-if="operation.requestBody.contents.length == 1" class="label">{{ selectContent.type }}</span>
                <pre v-highlightjs="formatLanguage(JSON.stringify(selectContent.schema), 'application/json')" class="overflow-auto" style="max-height: 250px;"><code class="json"></code></pre>
                <select v-if="operation.requestBody.contents.length > 1" class="form-select form-select-sm" aria-label=".form-select-sm example" style="width: 30%" @change="selectedContentChange">
                    <option v-for="content in operation.requestBody.contents">{{ content.type }}</option>
                </select>
            </p>
        </div>
    </div>
</template>

<style scoped>
.codeBlock {
  position: relative;
}
.card-body .codeBlock .label {
    position: absolute;
    right: 20px;
    top: 4px;
    font-size: 0.8rem;
  }
  </style>

