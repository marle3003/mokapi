<script setup lang="ts">
import Markdown from 'vue3-markdown-it'
import SourceView from '../SourceView.vue'
import SchemaExpand from '../SchemaExpand.vue'
import SchemaValidate from '../SchemaValidate.vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import { ref } from 'vue'

const props = defineProps<{
    topic: KafkaTopic,
}>()

const { formatSchema } = usePrettyLanguage()

const types: { [name: string]: string } = {
    'application/json': '.json',
    'application/xml': '.xml'
}

let selected = ref<KafkaMessage | null>(null)

function filename() {
    let ext = '.dat'
    if (selected.value) {   
        if (types[selected.value.contentType]) {
            ext = types[selected.value.contentType]
        }
    }

    return `${props.topic.name}-example${ext}`
}

// select first message
for (const messageId in props.topic.messages) {
    selected.value = props.topic.messages[messageId]
    break
}

function selectedMessageChange(event: any){
    for (const messageId in props.topic.messages){
        const msg = props.topic.messages[messageId]
        if (msg.name == event.target.value){
            selected.value = msg
        }
    }
}
</script>

<template>
    <div class="row" v-if="Object.keys(topic.messages).length > 1">
        <div class="col-auto">
            <div class="label">
                Messages
            </div>
            
        </div>
        <hr />
    </div>
    <div v-if="selected">
        <div class="row">
            <div class="col-auto" v-if="selected.name">
                <p id="message-name" class="label">Name</p>
                <select class="form-select form-select-sm mb-2" aria-labelledby="message-name" @change="selectedMessageChange" v-if="Object.keys(topic.messages).length > 1">
                    <option v-for="msg in topic.messages">{{ msg.name }}</option>
                </select>
                <p aria-labelledby="message-name" v-else>{{ selected.name }}</p>
            </div>
        </div>
        <div class="row mt-2" v-if="selected.title">
            <div class="col">
                <p id="message-title" class="label">Title</p>
                <p aria-labelledby="message-title">{{ selected.title }}</p>
            </div>
        </div>
        <div class="row mt-2">
            <div class="col" v-if="selected.summary">
                <p id="message-summary" class="label">Summary</p>
                <p aria-labelledby="message-summary">{{ selected.summary }}</p>
            </div>
            <div class="col" v-if="selected.description">
                <p id="message-description" class="label">Description</p>
                <markdown :source="selected.description" aria-labelledby="message-description" :html="true"></markdown>
            </div>
        </div>
        <div class="row mt-2">
            <div class="col">
                <p id="message-content-type" class="label">Content Type</p>
                <p aria-labelledby="message-content-type">{{ selected.contentType }}</p>
            </div>
            
        </div>
        <div class="row mt-2">
            <ul class="nav nav-pills tab-sm" role="tablist">
                <li class="nav-link" :class="selected.key ? '' : 'disabled'" id="tab-config-schemas-key" data-bs-toggle="pill" data-bs-target="#tabpanel-config-schemas-key" type="button" role="tab" aria-controls="tabpanel-config-schemas-key" aria-selected="false">Key</li>
                <li class="nav-link active" id="tab-config-schemas-message" data-bs-toggle="pill" data-bs-target="#tabpanel-config-schemas-message" type="button" role="tab" aria-controls="tabpanel-config-schemas-message" aria-selected="true">Value</li>
            </ul>

            <div class="tab-content" id="tab-config-schemas">
                <div class="tab-pane fade" id="tabpanel-config-schemas-key" role="tabpanel" v-if="selected.key" aria-labelledby="tab-config-schemas-key">
                    <source-view :source="JSON.stringify(selected.key)" content-type="application/json" :hide-content-type="true" />
                </div>
                <div class="tab-pane fade show active" id="tabpanel-config-schemas-message" role="tabpanel" aria-labelledby="tab-config-schemas-message">
                    <section aria-label="Schema">
                        <source-view :source="formatSchema(selected.payload)" content-type="application/json" :hide-content-type="true" :height="500" class="mb-2" :filename="topic.name+'-message.json'" />
                        <div class="row">
                            <div class="col-auto pe-2 mt-1">
                                <schema-expand :schema="selected.payload" :title="'Value - '+topic.name" :source="{filename: topic.name+'-message.json'}" />
                            </div>
                            <div class="col-auto pe-2 mt-1">
                                <schema-validate :title="'Value Validator - '+topic.name" :schema="selected.payload" :content-type="selected.contentType" :source="{filename: filename()}" />
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