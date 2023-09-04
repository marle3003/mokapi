<script setup lang="ts">
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import type { PropType } from 'vue';
import hljs from 'highlight.js'
import Markdown from 'vue3-markdown-it';

defineProps({
    topic: { type: Object as PropType<KafkaTopic>, required: true },
})

const {formatLanguage} = usePrettyLanguage()

function getCode(schema: Schema){
    return hljs.highlightAuto(formatLanguage(JSON.stringify(schema), 'application/json')).value
}
</script>

<template>
    <div v-if="topic.configs">
        <div class="row">
            <div class="col" v-if="topic.configs.title">
                <p class="label">Title</p>
                <p data-testid="service-name">{{ topic.configs.title }}</p>
            </div>
            <div class="col"  v-if="topic.configs.name">
                <p class="label">Name</p>
                <p data-testid="service-version">{{ topic.configs.name }}</p>
            </div>
        </div>
        <div class="row">
            <div class="col" v-if="topic.configs.summary">
                <p class="label">Summary</p>
                <p>{{ topic.configs.summary }}</p>
            </div>
            <div class="col" v-if="topic.configs.description">
                <p class="label">Description</p>
                <markdown :source="topic.configs.description" data-testid="message-description"></markdown>
            </div>
        </div>
        <div class="row">
            <div class="col">
                <p class="label">Content Type</p>
                <p>{{ topic.configs.messageType }}</p>
            </div>
            
        </div>
        <div class="row">
            <ul class="nav nav-pills tab-sm" role="tabList">
                <li class="nav-link" :class="topic.configs.key ? '' : 'disabled'" id="pills-config-key-tab" data-bs-toggle="pill" data-bs-target="#pills-config-key" type="button" role="tab" aria-controls="'pills-config-key" aria-selected="false">Key</li>
                <li class="nav-link active" id="pills-config-tab" data-bs-toggle="pill" data-bs-target="#pills-config-message" type="button" role="tab" aria-controls="'pills-config-message" aria-selected="true">Message</li>
            </ul>

            <div class="tab-content" id="'pills-tabconfig">
                <div class="tab-pane fade" id="pills-config-key" role="tabpanel" v-if="topic.configs.key">
                    <pre><code class="json hljs" v-html="getCode(topic.configs.key)"></code></pre>
                </div>
                <div class="tab-pane fade show active" id="pills-config-message" role="tabpanel">
                    <pre><code class="json hljs" v-html="getCode(topic.configs.message)"></code></pre>
                </div>
            </div>
        </div>
    </div>
</template>