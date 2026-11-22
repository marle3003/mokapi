<script setup lang="ts">
import { computed } from 'vue'
import FilterChipBar from '../../log/FilterChipBar.vue'
import {
    addableFields, fieldLabel, filterChips, operations, removeFilter,
    type LdapRequestFilter, type LdapRequestFilterContext, type LdapRequestFilterField,
} from '@/composables/ldap-request-filter'

const props = defineProps<{
    context: LdapRequestFilterContext
    serviceNames: string[]
}>()
const filter = defineModel<LdapRequestFilter>({ required: true })
// The field whose popover is open. The log needs it to let Escape close the popover before the panel.
const editing = defineModel<LdapRequestFilterField | undefined>('editing')

const chips = computed(() => filterChips(filter.value, props.context))
const addable = computed(() => addableFields(filter.value, props.context))

// in place: the log owns the filter object
function reset(field: LdapRequestFilterField) {
    Object.assign(filter.value, removeFilter(filter.value, field))
}
</script>

<template>
    <filter-chip-bar v-model:editing="editing" :chips="chips" :addable="addable" :field-label="fieldLabel" @remove="reset">
        <template #editor="{ field }">
            <select v-if="field === 'service'" class="form-select form-select-sm" v-model="filter.service" aria-label="Service">
                <option v-for="name of serviceNames" :key="name">{{ name }}</option>
            </select>
            <select v-else-if="field === 'operation'" class="form-select form-select-sm" v-model="filter.operation" aria-label="Operation">
                <option value=""></option>
                <option v-for="operation of operations" :key="operation.value" :value="operation.value">{{ operation.label }}</option>
            </select>
        </template>
    </filter-chip-bar>
</template>
