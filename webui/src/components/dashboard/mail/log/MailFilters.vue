<script setup lang="ts">
import { computed } from 'vue'
import FilterChipBar from '../../log/FilterChipBar.vue'
import {
    addableFields, fieldLabel, filterChips, removeFilter,
    type MailLogFilter, type MailLogFilterContext, type MailLogFilterField,
} from '@/composables/mail-log-filter'

const props = defineProps<{
    context: MailLogFilterContext
    serviceNames: string[]
}>()
const filter = defineModel<MailLogFilter>({ required: true })
// The field whose popover is open. The log needs it to let Escape close the popover before the panel.
const editing = defineModel<MailLogFilterField | undefined>('editing')

const chips = computed(() => filterChips(filter.value, props.context))
const addable = computed(() => addableFields(filter.value, props.context))

// in place: the log owns the filter object
function reset(field: MailLogFilterField) {
    Object.assign(filter.value, removeFilter(filter.value, field))
}
</script>

<template>
    <filter-chip-bar v-model:editing="editing" :chips="chips" :addable="addable" :field-label="fieldLabel" @remove="reset">
        <template #editor="{ field }">
            <select v-if="field === 'service'" class="form-select form-select-sm" v-model="filter.service" aria-label="Service">
                <option v-for="name of serviceNames" :key="name">{{ name }}</option>
            </select>
        </template>
    </filter-chip-bar>
</template>
