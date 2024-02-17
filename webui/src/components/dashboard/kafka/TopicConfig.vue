<script setup lang="ts">
import Markdown from 'vue3-markdown-it'
import SourceView from '../SourceView.vue'
import SchemaExpand from '../SchemaExpand.vue'
import SchemaExample from '../SchemaExample.vue'

defineProps<{
    topic: KafkaTopic,
}>()
</script>

<template>
    <div v-if="topic.configs">
        <div class="row">
            <div class="col" v-if="topic.configs.title">
                <p id="message-title" class="label">Title</p>
                <p aria-labelledby="message-title">{{ topic.configs.title }}</p>
            </div>
            <div class="col" v-if="topic.configs.name">
                <p id="message-name" class="label">Name</p>
                <p aria-labelledby="message-name">{{ topic.configs.name }}</p>
            </div>
        </div>
        <div class="row mt-2">
            <div class="col" v-if="topic.configs.summary">
                <p id="message-summary" class="label">Summary</p>
                <p aria-labelledby="message-summary">{{ topic.configs.summary }}</p>
            </div>
            <div class="col" v-if="topic.configs.description">
                <p id="message-description" class="label">Description</p>
                <markdown :source="topic.configs.description" aria-labelledby="message-description"></markdown>
            </div>
        </div>
        <div class="row mt-2">
            <div class="col">
                <p id="message-content-type" class="label">Content Type</p>
                <p aria-labelledby="message-content-type">{{ topic.configs.messageType }}</p>
            </div>
            
        </div>
        <div class="row mt-2">
            <ul class="nav nav-pills tab-sm" role="tablist">
                <li class="nav-link" :class="topic.configs.key ? '' : 'disabled'" id="tab-config-schemas-key" data-bs-toggle="pill" data-bs-target="#tabpanel-config-schemas-key" type="button" role="tab" aria-controls="tabpanel-config-schemas-key" aria-selected="false">Key</li>
                <li class="nav-link active" id="tab-config-schemas-message" data-bs-toggle="pill" data-bs-target="#tabpanel-config-schemas-message" type="button" role="tab" aria-controls="tabpanel-config-schemas-message" aria-selected="true">Message</li>
            </ul>

            <div class="tab-content" id="tab-config-schemas">
                <div class="tab-pane fade" id="tabpanel-config-schemas-key" role="tabpanel" v-if="topic.configs.key" aria-labelledby="tab-config-schemas-key">
                    <source-view :source="JSON.stringify(topic.configs.key)" content-type="application/json" :hide-content-type="true" />
                </div>
                <div class="tab-pane fade show active" id="tabpanel-config-schemas-message" role="tabpanel" aria-labelledby="tab-config-schemas-message">
                    <section aria-label="Schema">
                        <source-view :source="JSON.stringify(topic.configs.message)" content-type="application/json" :hide-content-type="true" height="500px" class="mb-2" :filename="topic.name+'-message.json'" />
                        <div class="row">
                            <div class="col-auto pe-2 mt-1">
                                <schema-expand :schema="topic.configs.message" :title="'Message - '+topic.name" :source="{filename: topic.name+'-message.json'}" />
                            </div>
                            <div class="col-auto pe-2 mt-1">
                                <schema-example :content-type="topic.configs.messageType" :schema="topic.configs.message"/>
                            </div> 
                        </div>
                    </section>
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