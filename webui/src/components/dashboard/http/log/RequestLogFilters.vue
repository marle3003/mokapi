<script setup lang="ts">
import { computed } from 'vue'
import FilterChipBar from '../../log/FilterChipBar.vue'
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

// in place: the log owns the filter object
function reset(field: FilterField) {
    Object.assign(filter.value, removeFilter(filter.value, field))
}
</script>

<template>
    <filter-chip-bar v-model:editing="editing" :chips="chips" :addable="addable" :field-label="fieldLabel"
        @add="field => activateFilter(filter, field)" @remove="reset">
        <template #editor="{ field }">
            <option-filter-editor v-if="field === 'service'" v-model="filter.service" label="API Name" :options="serviceNames" />
            <method-filter-editor v-else-if="field === 'method'" v-model="filter.method" />
            <url-filter-editor v-else-if="field === 'url'" v-model="filter.url.value" />
            <header-filter-editor v-else-if="field === 'requestHeaders'" v-model="filter.request.headers" label="Request Header" />
            <status-code-filter-editor v-else-if="field === 'statusCode'" v-model="filter.response.statusCode.value" />
            <header-filter-editor v-else-if="field === 'responseHeaders'" v-model="filter.response.headers" label="Response Header" />
            <client-ip-filter-editor v-else-if="field === 'clientIP'" v-model="filter.clientIP.value" />
        </template>
    </filter-chip-bar>
</template>
