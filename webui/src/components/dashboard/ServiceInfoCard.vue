<script setup lang="ts">
import { computed, ref } from 'vue';
import { useMarkdown } from '@/composables/markdown';
import { useToolbarItems } from '@/composables/toolbar-items';


const props = defineProps<{
    service: Service,
    type?: string
    status?: string
}>()

const description = computed(() => useMarkdown(props.service.description, true).content)
const toolbar = useToolbarItems()
</script>

<template>
    <section class="card position-relative" data-testid="service-info" aria-label="Info">

        <div class="dropdown card-toolbar" style="z-index: 100;" v-if="toolbar.items.length > 0">
            <button
                type="button"
                class="btn btn-sm btn-outline-secondary kebab-btn"
                data-bs-toggle="dropdown"
                aria-expanded="false"
                :aria-label="`Actions for ${service.name}`"
            >
                <i class="bi bi-three-dots-vertical"></i>
            </button>
            <ul class="dropdown-menu dropdown-menu-end">
                <li v-for="item in toolbar.items" :key="item.id">
                    <button class="dropdown-item" @click.prevent="item.onClick()">
                        <i v-if="item.icon" :class="['bi', item.icon, 'me-2']" aria-hidden="true"></i>{{ item.label }}
                    </button>

                </li>
            </ul>
        </div>

        <div class="card-body">
            <div class="row">
                <div class="col col-8 header mb-3">
                    <p id="name" class="label">Name</p>
                    <p aria-labelledby="name" data-testid="service-name">{{ service.name }}</p>
                </div>
                <div class="col-1">
                    <p id="version" class="label">Version</p>
                    <p aria-labelledby="version" data-testid="service-version">{{ service.version }}</p>
                </div>
                <div class="col-2">
                    <p id="contact" class="label">Contact</p>
                    <ul v-if="service.contact" class="contact" aria-labelledby="contact" data-testid="service-contact">
                        <li v-if="service.contact.name || service.contact.url">
                            <a v-if="service.contact.url" :href="service.contact.url">
                                <span v-if="service.contact.name">{{ service.contact.name }}</span>
                                <i v-else class="bi bi-link"></i>
                            </a>
                            <span v-else>{{ service.contact.name }}</span>
                        </li>
                        <li>
                            <a v-if="service.contact.email" :href="'mailto:'+service.contact.email" :title="service.contact.email" data-testid="service-mail">
                                <span class="bi bi-envelope"></span>
                            </a>
                        </li>
                    </ul>
                </div>
                <div class="col-1 text-end">
                    <span class="badge bg-secondary api" title="Type of API" aria-label="Type of API" data-testid="service-type">{{ type ? type : service.type }}</span>
                </div>
            </div>
            <div class="row">
                <div class="col">
                    <p id="description" class="label">Description</p>
                    <div aria-labelledby="description" v-html="description"></div>
                </div>
            </div>
            <div class="row" v-if="status && status !== 'valid'">
                <div class="col">
                    <p id="status" class="label">Status</p>
                    <div aria-labelledby="status" class="status">{{ status }}</div>
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
ul.contact li span.bi {
    vertical-align:middle;
}
.card-toolbar {
    position: absolute;
    top: -7px;
    right: 0;
    transform: translateY(-100%);
    background: var(--bs-body-bg, #fff);
    border-radius: 0.375rem;
    box-shadow: 0 1px 2px rgba(0,0,0,0.08);
}
</style>