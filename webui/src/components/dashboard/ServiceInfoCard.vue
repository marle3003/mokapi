<script setup lang="ts">
import type { PropType } from 'vue';
import Markdown from 'vue3-markdown-it';

defineProps({
    service: { type: Object as PropType<Service>, required: true },
    type: { type: String, required: false}
})
</script>

<template>
    <div class="card" data-testid="service-info">
            <div class="card-body">
                <div class="row">
                    <div class="col header">
                        <p class="label">Name</p>
                        <p data-testid="service-name">{{ service.name }}</p>
                    </div>
                    <div class="col">
                        <p class="label">Version</p>
                        <p data-testid="service-version">{{ service.version }}</p>
                    </div>
                    <div class="col">
                        <p class="label">Contact</p>
                        <p v-if="service.contact" data-testid="service-contact">
                            <a v-if="service.contact.url" :href="service.contact.url">{{ service.contact.name }}</a>
                            <span v-else>{{ service.contact.name }}</span>
                            <a v-if="service.contact.email" :href="'mailto:'+service.contact.email" style="margin-left: 0.5em;" data-testid="service-mail"><i class="bi bi-envelope"></i></a>
                        </p>
                    </div>
                    <div class="col text-end">
                        <span class="badge bg-secondary" data-testid="service-type">{{ type ? type : service.type }}</span>
                    </div>
                </div>
                <div class="row">
                    <div class="col">
                        <p class="label">Description</p>
                        <markdown :source="service.description" data-testid="service-description"></markdown>
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