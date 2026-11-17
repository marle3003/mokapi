<script setup lang="ts">
import { computed, nextTick, ref, useTemplateRef, watch } from 'vue';
import { useMarkdown } from '@/composables/markdown';
import { useToolbarItems } from '@/composables/toolbar-items';
import { useRoute } from '@/router';
import { Dropdown } from 'bootstrap';

const props = defineProps<{
    service: Service,
    type?: string
    status?: string
}>()

const route = useRoute()

const description = computed(() => useMarkdown(props.service.description, true).content)
const toolbar = useToolbarItems()
const exportButton = useTemplateRef('export-button')
const highlightItem = ref<string | null>(null)

watch(() => [route.query.demo, exportButton.value], async ([kind]) => {
    if (kind !== 'bruno' || !exportButton.value) {
        return
    }
    await nextTick()
    Dropdown.getOrCreateInstance(exportButton.value).show()
    highlightItem.value = 'Bruno collection'
}, { immediate: true })
</script>

<template>
    <section class="card" data-testid="service-info" aria-label="Info">

        <div class="card-body">
            <div class="row">
                <div class="col header mb-3">
                    <p id="name" class="label">Name</p>
                    <div class="service-title d-flex flex-wrap align-items-center gap-2">
                        <span class="badge bg-secondary api" title="Type of API" aria-label="Type of API" data-testid="service-type">{{ type ? type : service.type }}</span>
                        <p aria-labelledby="name" data-testid="service-name">{{ service.name }}</p>
                    </div>
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
                <div class="col-auto ms-auto" v-if="toolbar.items.length > 0">
                    <div class="dropdown">
                        <button
                            type="button"
                            class="btn btn-outline-secondary btn-compact dropdown-toggle"
                            data-bs-toggle="dropdown"
                            aria-expanded="false"
                            ref="export-button"
                        >
                            <i class="bi bi-box-arrow-up me-1" aria-hidden="true"></i>Export
                        </button>
                        <ul class="dropdown-menu dropdown-menu-end export-menu">
                            <li v-for="item in toolbar.items" :key="item.id">
                                <button class="dropdown-item" @click.prevent="highlightItem = null; item.onClick()" :class="{ 'menu-highlight': highlightItem == item.label}">
                                    <i v-if="item.icon" :class="['bi', item.icon, 'me-2']" aria-hidden="true"></i>{{ item.label }}
                                </button>
                            </li>
                        </ul>
                    </div>
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
/* the badge scales with the font size of the name (1.4rem in .header p) */
.service-title {
    font-size: 1.4rem;
    margin-bottom: .3rem;
}
.dashboard .card .service-title p {
    margin: 0;
}
.service-title .badge {
    font-size: .55em;
}
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
/* same look as the "Add filter" menu (.filter-menu in dashboard.css) */
.export-menu {
    --bs-dropdown-font-size: .8125rem;
    --bs-dropdown-min-width: 12rem;
    --bs-dropdown-padding-y: .25rem;
    --bs-dropdown-border-color: var(--layout-border);
    --bs-dropdown-border-radius: 6px;
    --bs-dropdown-bg: var(--layout-popover-background);
    --bs-dropdown-box-shadow: var(--layout-popover-shadow);
    --bs-dropdown-color: var(--layout-text);
    --bs-dropdown-link-color: var(--layout-text);
    --bs-dropdown-link-hover-color: var(--layout-text);
    --bs-dropdown-link-hover-bg: var(--layout-hover-background);
    --bs-dropdown-link-active-color: var(--layout-text);
    --bs-dropdown-link-active-bg: var(--layout-hover-background);
    --bs-dropdown-item-padding-y: .3rem;
    --bs-dropdown-item-padding-x: .75rem;
}
.export-menu .dropdown-item:focus-visible {
    outline: 2px solid var(--layout-accent);
    outline-offset: -2px;
}
.menu-highlight {
  background: var(--bs-primary-bg-subtle);
  animation: menu-pulse 1.6s ease-in-out 3;
}
@keyframes menu-pulse {
  30% { background: var(--bs-primary-border-subtle); }
}
@media (prefers-reduced-motion: reduce) {
  .menu-highlight { animation: none; }
}
</style>