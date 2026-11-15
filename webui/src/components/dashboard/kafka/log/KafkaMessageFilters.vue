<script setup lang="ts">
import { computed } from 'vue'
import FilterChipBar from '../../log/FilterChipBar.vue'
import {
    addableFields, fieldLabel, filterChips, removeFilter,
    type KafkaMessageFilter, type KafkaMessageFilterContext, type KafkaMessageFilterField,
} from '@/composables/kafka-message-filter'

const props = defineProps<{
    context: KafkaMessageFilterContext
    clusterNames: string[]
    // the topics the Topic field suggests
    topicNames: string[]
}>()
const filter = defineModel<KafkaMessageFilter>({ required: true })
// The field whose popover is open. The log needs it to let Escape close the popover before the panel.
const editing = defineModel<KafkaMessageFilterField | undefined>('editing')

const chips = computed(() => filterChips(filter.value, props.context))
const addable = computed(() => addableFields(filter.value, props.context))

// in place: the log owns the filter object
function reset(field: KafkaMessageFilterField) {
    Object.assign(filter.value, removeFilter(filter.value, field))
}

// A number input would turn the value into a number, the filter keeps text.
function setPartition(e: Event) {
    filter.value.partition = (e.target as HTMLInputElement).value
}
</script>

<template>
    <filter-chip-bar v-model:editing="editing" :chips="chips" :addable="addable" :field-label="fieldLabel" @remove="reset">
        <template #editor="{ field }">
            <select v-if="field === 'cluster'" class="form-select form-select-sm" v-model="filter.cluster" aria-label="Cluster">
                <option v-for="name of clusterNames" :key="name">{{ name }}</option>
            </select>
            <template v-else-if="field === 'topic'">
                <input type="text" class="form-control form-control-sm" v-model="filter.topic" aria-label="Topic" list="kafka-topic-options">
                <datalist id="kafka-topic-options">
                    <option v-for="name of topicNames" :key="name" :value="name"></option>
                </datalist>
            </template>
            <input v-else-if="field === 'partition'" type="number" min="0" class="form-control form-control-sm"
                :value="filter.partition" aria-label="Partition" @input="setPartition">
            <input v-else-if="field === 'clientId'" type="text" class="form-control form-control-sm" v-model="filter.clientId" aria-label="Client">
        </template>
    </filter-chip-bar>
</template>
