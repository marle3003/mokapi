<script setup lang="ts">
import type { PropType, } from 'vue';
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import HttpParameters from './HttpParameters.vue';

const {formatLanguage} = usePrettyLanguage()

const props = defineProps({
    operation: { type: Object as PropType<HttpOperation>, required: true }
})
let selectContent: HttpMediaType | null = null
if (props.operation.requestBody?.contents?.length > 0){
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
            <div class="card-title text-center">Request</div>

            <ul class="nav nav-pills response-tab" role="tabList">
                <li class="nav-link" id="pills-body-tab" :class="operation.requestBody ? 'active' : 'disabled'" data-bs-toggle="pill" data-bs-target="#pills-body" type="button" role="tab" aria-controls="pills-body" aria-selected="true">Body</li>
                <li class="nav-link" id="pills-header-tab" :class="operation.parameters ? (operation.requestBody ? '' : 'active') : 'disabled'" data-bs-toggle="pill" data-bs-target="#pills-parameter" type="button" role="tab" aria-controls="pills-parameter" aria-selected="true">Parameters</li>
            </ul>
            <div class="tab-content" id="pills-tabContent">
                <div v-if="operation.requestBody" class="tab-pane fade" :class="operation.requestBody ? 'show active' : ''" id="pills-body" role="tabpanel" aria-labelledby="pills-body-tab">
                    <p class="codeBlock">
                    <span v-if="operation.requestBody.contents.length == 1" class="label">{{ selectContent?.type }}</span>
                    <pre v-highlightjs="formatLanguage(JSON.stringify(selectContent?.schema), 'application/json')" class="overflow-auto" style="max-height: 250px;"><code class="json"></code></pre>
                    <select v-if="operation.requestBody.contents.length > 1" class="form-select form-select-sm" aria-label=".form-select-sm example" style="width: 30%" @change="selectedContentChange">
                        <option v-for="content in operation.requestBody.contents">{{ content.type }}</option>
                    </select>
                    </p>
                </div>
                <div v-if="operation.parameters" class="tab-pane fade" :class="!operation.requestBody ? 'show active' : ''" id="pills-parameter" role="tabpanel" aria-labelledby="pills-parameter-tab">
                    <http-parameters :parameters="operation.parameters" />
                </div>
            </div>
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

