<script setup lang="ts">
import { computed, type PropType } from 'vue'
import Markdown from 'vue3-markdown-it'

const props = defineProps<{
    service: Service,
    type?: string
}>()

const developer = computed(() => {
    if (!props.service?.contact?.name || props.service?.contact?.name === '') {
        return 'the developer'
    }
    return props.service.contact.name
})
</script>

<template>
    <section class="card" data-testid="service-info" aria-label="Info">
            <div class="card-body">
                <div class="row">
                    <div class="col col-8 header mb-3">
                        <label id="name" class="label">Name</label>
                        <p aria-labelledby="name" data-testid="service-name">{{ service.name }}</p>
                    </div>
                    <div class="col">
                        <p id="version" class="label">Version</p>
                        <p aria-labelledby="version" data-testid="service-version">{{ service.version }}</p>
                    </div>
                    <div class="col-2">
                        <p id="contact" class="label">Contact</p>
                        <ul v-if="service.contact" class="contact" aria-labelledby="contact" data-testid="service-contact">
                            <li>
                                <a v-if="service.contact.url" :href="service.contact.url" :title="developer+' - Website'">{{ service.contact.name }}</a>
                                <span v-else>{{ service.contact.name }}</span>
                            </li>
                            <li>
                                <a v-if="service.contact.email" :href="'mailto:'+service.contact.email" :title="'Send email to '+developer" data-testid="service-mail">
                                    <i class="bi bi-envelope"></i>
                                </a>
                            </li>
                        </ul>
                    </div>
                    <div class="col-1 text-end" style="width:70px;">
                        <span class="badge bg-secondary api" title="Type of API" aria-label="Type of API" data-testid="service-type">{{ type ? type : service.type }}</span>
                    </div>
                </div>
                <div class="row">
                    <div class="col">
                        <p id="description" class="label">Description</p>
                        <markdown :source="service.description" data-testid="service-description" aria-labelledby="description" :html="true"></markdown>
                    </div>
                </div>
            </div>
        </section>
</template>

<style scoped>
ul.contact {
    list-style: none; 
    padding: 0;
}
ul.contact li {
    display: inline;
    padding-right: 0.5em;
}
ul.contact li i {
    vertical-align:middle;
}
</style>