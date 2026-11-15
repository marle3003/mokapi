<script setup lang="ts">
import { computed, useId } from 'vue'
import FilterChipBar from '../../log/FilterChipBar.vue'
import {
    addableFields, fieldLabel, filterChips, removeFilter,
    type MqttMessageFilter, type MqttMessageFilterContext, type MqttMessageFilterField,
} from '@/composables/mqtt-message-filter'

const props = defineProps<{
    context: MqttMessageFilterContext
    clusterNames: string[]
    // the topics the Topic field suggests
    topicNames: string[]
}>()
const filter = defineModel<MqttMessageFilter>({ required: true })
// The field whose popover is open. The log needs it to let Escape close the popover before the panel.
const editing = defineModel<MqttMessageFilterField | undefined>('editing')

// The log can be on the page twice (a tab and the panel of another log), so the list needs its own id.
const topicOptionsId = useId()

const chips = computed(() => filterChips(filter.value, props.context))
const addable = computed(() => addableFields(filter.value, props.context))

// in place: the log owns the filter object
function reset(field: MqttMessageFilterField) {
    Object.assign(filter.value, removeFilter(filter.value, field))
}
</script>

<template>
    <filter-chip-bar v-model:editing="editing" :chips="chips" :addable="addable" :field-label="fieldLabel" @remove="reset">
        <template #editor="{ field }">
            <select v-if="field === 'cluster'" class="form-select form-select-sm" v-model="filter.cluster" aria-label="Cluster">
                <option v-for="name of clusterNames" :key="name">{{ name }}</option>
            </select>
            <template v-else-if="field === 'topic'">
                <input type="text" class="form-control form-control-sm" v-model="filter.topic" aria-label="Topic" :list="topicOptionsId">
                <datalist :id="topicOptionsId">
                    <option v-for="name of topicNames" :key="name" :value="name"></option>
                </datalist>
            </template>
            <input v-else-if="field === 'clientId'" type="text" class="form-control form-control-sm" v-model="filter.clientId" aria-label="Client">
        </template>
    </filter-chip-bar>
</template>
