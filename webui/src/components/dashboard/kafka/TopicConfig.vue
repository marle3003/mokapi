<script setup lang="ts">
import type { PropType } from 'vue'
import Markdown from 'vue3-markdown-it'
import SourceView from '../SourceView.vue'
import SchemaExpand from '../SchemaExpand.vue'
import SchemaExample from '../SchemaExample.vue'

defineProps({
    topic: { type: Object as PropType<KafkaTopic>, required: true },
})

</script>

<template>
    <div v-if="topic.configs">
        <div class="row">
            <div class="col" v-if="topic.configs.title">
                <p class="label">Title</p>
                <p data-testid="service-name">{{ topic.configs.title }}</p>
            </div>
            <div class="col" v-if="topic.configs.name">
                <p class="label">Name</p>
                <p data-testid="service-version">{{ topic.configs.name }}</p>
            </div>
        </div>
        <div class="row mt-2">
            <div class="col" v-if="topic.configs.summary">
                <p class="label">Summary</p>
                <p>{{ topic.configs.summary }}</p>
            </div>
            <div class="col" v-if="topic.configs.description">
                <p class="label">Description</p>
                <markdown :source="topic.configs.description" data-testid="message-description"></markdown>
            </div>
        </div>
        <div class="row mt-2">
            <div class="col">
                <p class="label">Content Type</p>
                <p>{{ topic.configs.messageType }}</p>
            </div>
            
        </div>
        <div class="row mt-2">
            <ul class="nav nav-pills tab-sm" role="tabList">
                <li class="nav-link" :class="topic.configs.key ? '' : 'disabled'" id="pills-config-key-tab" data-bs-toggle="pill" data-bs-target="#pills-config-key" type="button" role="tab" aria-controls="'pills-config-key" aria-selected="false">Key</li>
                <li class="nav-link active" id="pills-config-tab" data-bs-toggle="pill" data-bs-target="#pills-config-message" type="button" role="tab" aria-controls="'pills-config-message" aria-selected="true">Message</li>
            </ul>

            <div class="tab-content" id="'pills-tabconfig">
                <div class="tab-pane fade" id="pills-config-key" role="tabpanel" v-if="topic.configs.key">
                    <source-view :source="JSON.stringify(topic.configs.key)" content-type="application/json" :hide-content-type="true" />
                </div>
                <div class="tab-pane fade show active" id="pills-config-message" role="tabpanel">
                    <source-view :source="JSON.stringify(topic.configs.message)" content-type="application/json" :hide-content-type="true" height="500px" class="mb-2" />
                    <div class="row">
                        <div class="col-auto pe-2 mt-1">
                            <schema-expand :schema="topic.configs.message" />
                        </div>
                        <div class="col-auto pe-2 mt-1">
                            <schema-example :content-type="topic.configs.messageType" :schema="topic.configs.message"/>
                        </div> 
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.tab-pane {
    padding: 0;
}
</style>