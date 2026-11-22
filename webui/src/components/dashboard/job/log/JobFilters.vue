<script setup lang="ts">
import { computed, useId } from 'vue'
import FilterChipBar from '../../log/FilterChipBar.vue'
import {
    addableFields, fieldLabel, filterChips, removeFilter,
    type JobLogFilter, type JobLogFilterField,
} from '@/composables/job-log-filter'

defineProps<{
    // the job names the Name field suggests
    suggestions: string[]
}>()
const filter = defineModel<JobLogFilter>({ required: true })
// The field whose popover is open. The log needs it to let Escape close the popover before the panel.
const editing = defineModel<JobLogFilterField | undefined>('editing')

// The log can be on the page twice (a tab and the panel of another log), so the list needs its own id.
const nameOptionsId = useId()

const chips = computed(() => filterChips(filter.value, {}))
const addable = computed(() => addableFields(filter.value, {}))

// in place: the log owns the filter object
function reset(field: JobLogFilterField) {
    Object.assign(filter.value, removeFilter(filter.value, field))
}
</script>

<template>
    <filter-chip-bar v-model:editing="editing" :chips="chips" :addable="addable" :field-label="fieldLabel" @remove="reset">
        <template #editor="{ field }">
            <template v-if="field === 'name'">
                <input type="text" class="form-control form-control-sm" v-model="filter.name" aria-label="Name" :list="nameOptionsId">
                <datalist :id="nameOptionsId">
                    <option v-for="name of suggestions" :key="name" :value="name"></option>
                </datalist>
            </template>
        </template>
    </filter-chip-bar>
</template>
