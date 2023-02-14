<script setup lang="ts">
import type { PropType } from 'vue';
import Markdown from 'vue3-markdown-it';

defineProps({
    service: { type: Object as PropType<Service>, required: true },
    type: { type: String, required: false}
})
</script>

<template>
    <div class="card">
            <div class="card-body">
                <div class="row">
                    <div class="col header">
                        <p class="label">Name</p>
                        <p>{{ service.name }}</p>
                    </div>
                    <div class="col">
                        <p class="label">Version</p>
                        <p>{{ service.version }}</p>
                    </div>
                    <div class="col">
                        <p class="label">Contact</p>
                        <p v-if="service.contact">
                            <a :href="service.contact.url">{{ service.contact.name }}</a>
                            <a v-if="service.contact.email" :href="'mailto:'+service.contact.email" style="margin-left: 0.5em;"><i class="bi bi-envelope"></i></a>
                        </p>
                    </div>
                    <div class="col text-end">
                        <span class="badge bg-secondary">{{ type ? type : service.type }}</span>
                    </div>
                </div>
                <div class="row">
                    <div class="col">
                        <p class="label">Description</p>
                        <markdown :source="service.description"></markdown>
                    </div>
                    
                </div>
            </div>
        </div>
</template>

<style scoped>
.col .badge{
    font-size: 1em;
}
</style>