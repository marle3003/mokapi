<script setup lang="ts">
import type { PropType, } from 'vue';
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import HttpParameters from './HttpParameters.vue';
import SchemaExpand from '../../SchemaExpand.vue';
import SchemaExample from '../../SchemaExample.vue';

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

            <div class="nav card-tabs" role="tablist">
              <button :class="operation.requestBody ? 'active' : 'disabled'" id="body-tab" data-bs-toggle="tab" data-bs-target="#body" type="button" role="tab" aria-controls="body" aria-selected="true">Body</button>
              <button :class="operation.parameters ? (operation.requestBody ? '' : 'active') : 'disabled'" id="parameters-tab" data-bs-toggle="tab" data-bs-target="#parameters" type="button" role="tab" aria-controls="parameters" aria-selected="false">Parameters</button>
            </div>

            <div class="tab-content" id="tabRequest">
              <div class="tab-pane fade" :class="operation.requestBody ? 'show active' : ''" id="body" role="tabpanel" aria-labelledby="body-tab" v-if="operation.requestBody">
                    <p class="codeBlock">
                        <span v-if="operation.requestBody.contents.length == 1" class="label">{{ selectContent?.type }}</span>
                        <pre v-highlightjs="formatLanguage(JSON.stringify(selectContent?.schema), 'application/json')" class="overflow-auto" style="max-height: 250px;"><code class="json"></code></pre>
                    </p>
                    <div class="row">
                        <div class="col-auto pe-2" v-if="selectContent">
                            <schema-expand :schema="selectContent.schema" />
                        </div>
                        <div class="col-auto px-2" v-if="selectContent">
                            <schema-example :schema="selectContent.schema" />
                        </div>
                        <div class="col-auto px-2">
                            <select v-if="operation.requestBody.contents.length > 1" class="form-select form-select-sm" aria-label=".form-select-sm example" style="width: 30%" @change="selectedContentChange">
                                <option v-for="content in operation.requestBody.contents">{{ content.type }}</option>
                            </select>
                        </div>
                    </div>
                </div>
                <div class="tab-pane fade" :class="!operation.requestBody ? 'show active' : ''" id="parameters" role="tabpanel" aria-labelledby="parameters-tab">
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

