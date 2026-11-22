<script setup lang="ts">
import { computed, useId } from 'vue'
import FilterChipBar from '../../log/FilterChipBar.vue'
import {
    addableFields, fieldLabel, filterChips, removeFilter,
    type WebsocketMessageFilter, type WebsocketMessageFilterContext, type WebsocketMessageFilterField,
} from '@/composables/websocket-message-filter'

const props = defineProps<{
    context: WebsocketMessageFilterContext
    serviceNames: string[]
    // the paths the Path field suggests
    channelNames: string[]
}>()
const filter = defineModel<WebsocketMessageFilter>({ required: true })
// The field whose popover is open. The log needs it to let Escape close the popover before the panel.
const editing = defineModel<WebsocketMessageFilterField | undefined>('editing')

// The log can be on the page twice (a tab and the panel of another log), so the list needs its own id.
const channelOptionsId = useId()

const chips = computed(() => filterChips(filter.value, props.context))
const addable = computed(() => addableFields(filter.value, props.context))

// in place: the log owns the filter object
function reset(field: WebsocketMessageFilterField) {
    Object.assign(filter.value, removeFilter(filter.value, field))
}
</script>

<template>
    <filter-chip-bar v-model:editing="editing" :chips="chips" :addable="addable" :field-label="fieldLabel" @remove="reset">
        <template #editor="{ field }">
            <select v-if="field === 'service'" class="form-select form-select-sm" v-model="filter.service" aria-label="Service">
                <option v-for="name of serviceNames" :key="name">{{ name }}</option>
            </select>
            <template v-else-if="field === 'channel'">
                <input type="text" class="form-control form-control-sm" v-model="filter.channel" aria-label="Path" :list="channelOptionsId">
                <datalist :id="channelOptionsId">
                    <option v-for="name of channelNames" :key="name" :value="name"></option>
                </datalist>
            </template>
            <input v-else-if="field === 'clientId'" type="text" class="form-control form-control-sm" v-model="filter.clientId" aria-label="Client">
        </template>
    </filter-chip-bar>
</template>
