<script setup lang="ts">
import { reactive, type PropType, } from 'vue'
import HttpParameters from './HttpParameters.vue'
import SchemaExpand from '../../SchemaExpand.vue'
import SchemaExample from '../../SchemaExample.vue'
import SourceView from '../../SourceView.vue'

const props = defineProps({
    operation: { type: Object as PropType<HttpOperation>, required: true }
})
const selected = reactive({
    content: {} as HttpMediaType | null
})

if (props.operation.requestBody?.contents?.length > 0){
    selected.content = props.operation.requestBody.contents[0]
}
function selectedContentChange(event: any){
    for (let content of props.operation.requestBody.contents){
        if (content.type == event.target.value){
            selected.content = content
        }
    }
}
</script>

<template>
    <div class="card" data-testid="http-request">
        <div class="card-body">
            <div class="card-title text-center">Request</div>

            <div class="nav card-tabs" role="tablist" data-testid="tabs">
              <button :class="operation.requestBody ? 'active' : 'disabled'" id="body-tab" data-bs-toggle="tab" data-bs-target="#body" type="button" role="tab" aria-controls="body" aria-selected="true">Body</button>
              <button :class="operation.parameters ? (operation.requestBody ? '' : 'active') : 'disabled'" id="parameters-tab" data-bs-toggle="tab" data-bs-target="#parameters" type="button" role="tab" aria-controls="parameters" aria-selected="false">Parameters</button>
            </div>

            <div class="tab-content" id="tabRequest">
              <div class="tab-pane fade" :class="operation.requestBody ? 'show active' : ''" id="body" role="tabpanel" aria-labelledby="body-tab" v-if="operation.requestBody">
                    <p class="label">Description</p>
                    <p>{{  operation.requestBody.description }}</p>
                    <p v-if="operation.requestBody.required">Required</p>
                    
                    <source-view 
                        :source="JSON.stringify(selected.content?.schema)" 
                        :deprecated="selected.content?.schema.deprecated" 
                        content-type="application/json"
                        :hide-content-type="true"
                        height="250px" class="mb-2">
                    </source-view>

                    <div class="row">
                        <div class="col-auto pe-2" v-if="selected.content">
                            <schema-expand :schema="selected.content.schema" />
                        </div>
                        <div class="col-auto px-2" v-if="selected.content">
                            <schema-example :schema="selected.content.schema" :content-type="selected.content.type" />
                        </div>
                        <div class="col-auto px-2">
                            <select v-if="operation.requestBody.contents.length > 1" class="form-select form-select-sm" aria-label="Request content type" @change="selectedContentChange">
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

