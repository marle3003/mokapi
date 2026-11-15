<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue'
import FilterPopover from './FilterPopover.vue'
import OptionFilterEditor from './OptionFilterEditor.vue'
import MethodFilterEditor from './MethodFilterEditor.vue'
import UrlFilterEditor from './UrlFilterEditor.vue'
import HeaderFilterEditor from './HeaderFilterEditor.vue'
import StatusCodeFilterEditor from './StatusCodeFilterEditor.vue'
import ClientIpFilterEditor from './ClientIpFilterEditor.vue'
import {
    activateFilter, addableFields, fieldLabel, filterChips, removeFilter,
    type FilterField, type HttpRequestFilter,
} from '@/composables/http-request-filter'

const props = defineProps<{
    // set when the log belongs to one API
    serviceName?: string
    serviceNames: string[]
}>()
const filter = defineModel<HttpRequestFilter>({ required: true })
// The field whose popover is open. The log needs it to let Escape close the popover before the panel.
const editing = defineModel<FilterField | undefined>('editing')

const context = computed(() => ({ serviceName: props.serviceName }))
const chips = computed(() => filterChips(filter.value, context.value))
const addable = computed(() => addableFields(filter.value, context.value))

const root = useTemplateRef<HTMLElement>('root')
const menuOpen = ref(false)

function addFilter(field: FilterField) {
    menuOpen.value = false
    activateFilter(filter.value, field)
    editing.value = field
}

// A filter without a value is not a filter: leaving its popover switches it off again.
watch(editing, (_, previous) => {
    if (previous && !chips.value.some(c => c.field === previous)) {
        reset(previous)
    }
})

// in place: the log owns the filter object
function reset(field: FilterField) {
    Object.assign(filter.value, removeFilter(filter.value, field))
}

function closeMenuOnOutsideClick(e: MouseEvent) {
    if (!root.value?.contains(e.target as Node)) {
        menuOpen.value = false
    }
}
function closeMenu(e: KeyboardEvent) {
    if (menuOpen.value) {
        menuOpen.value = false
        e.stopPropagation()
    }
}
onMounted(() => document.addEventListener('mousedown', closeMenuOnOutsideClick))
onUnmounted(() => document.removeEventListener('mousedown', closeMenuOnOutsideClick))
</script>

<template>
    <div ref="root" class="request-log-filters" @keydown.esc="closeMenu">
        <ul class="filter-chips" aria-label="Active filters">
            <li v-for="chip in chips" :key="chip.field" class="filter-chip">
                <button type="button" class="filter-chip-label" :title="chip.title" @click="editing = chip.field">{{ chip.label }}</button>
                <button type="button" class="filter-chip-remove" :aria-label="`Remove ${chip.label}`"
                    @click="reset(chip.field)">
                    <i class="bi bi-x" aria-hidden="true"></i>
                </button>
            </li>
        </ul>

        <div class="filter-add" v-if="addable.length > 0">
            <button type="button" class="btn btn-outline-primary btn-sm" aria-haspopup="menu" :aria-expanded="menuOpen"
                @click="menuOpen = !menuOpen">
                <i class="bi bi-plus" aria-hidden="true"></i> Add filter
            </button>
            <ul v-if="menuOpen" class="filter-menu" role="menu">
                <li v-for="field in addable" :key="field.field" role="none">
                    <button type="button" role="menuitem" @click="addFilter(field.field)">{{ field.label }}</button>
                </li>
            </ul>
        </div>

        <filter-popover v-if="editing" :label="fieldLabel(editing)" @close="editing = undefined">
            <option-filter-editor v-if="editing === 'service'" v-model="filter.service" label="API Name" :options="serviceNames" />
            <method-filter-editor v-else-if="editing === 'method'" v-model="filter.method" />
            <url-filter-editor v-else-if="editing === 'url'" v-model="filter.url.value" />
            <header-filter-editor v-else-if="editing === 'requestHeaders'" v-model="filter.request.headers" label="Request Header" />
            <status-code-filter-editor v-else-if="editing === 'statusCode'" v-model="filter.response.statusCode.value" />
            <header-filter-editor v-else-if="editing === 'responseHeaders'" v-model="filter.response.headers" label="Response Header" />
            <client-ip-filter-editor v-else-if="editing === 'clientIP'" v-model="filter.clientIP.value" />
        </filter-popover>
    </div>
</template>
