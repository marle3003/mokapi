<script setup lang="ts" generic="Field extends string">
import { onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue'
import FilterPopover from './FilterPopover.vue'
import type { FilterChip } from '@/composables/filter-chip'

// The chips of the active filters, the "Add filter" menu and the popover of the field being
// edited. The protocol decides what a filter is: it gives the chips and the fields still to add,
// reacts to `add` and `remove` and puts its editor into the slot.
const props = defineProps<{
    chips: FilterChip<Field>[]
    addable: { field: Field, label: string }[]
    fieldLabel: (field: Field) => string
}>()
const emit = defineEmits<{ add: [field: Field], remove: [field: Field] }>()
// The field whose popover is open. The log needs it to let Escape close the popover before the panel.
const editing = defineModel<Field | undefined>('editing')

const root = useTemplateRef<HTMLElement>('root')
const menuOpen = ref(false)

function addFilter(field: Field) {
    menuOpen.value = false
    emit('add', field)
    editing.value = field
}

// A filter without a value is not a filter: leaving its popover removes it again.
watch(editing, (_, previous) => {
    if (previous && !props.chips.some(c => c.field === previous)) {
        emit('remove', previous)
    }
})

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
    <div ref="root" class="log-filters" @keydown.esc="closeMenu">
        <ul class="filter-chips" aria-label="Active filters">
            <li v-for="chip in chips" :key="chip.field" class="filter-chip">
                <button type="button" class="filter-chip-label" :title="chip.title" @click="editing = chip.field">{{ chip.label }}</button>
                <button type="button" class="filter-chip-remove" :aria-label="`Remove ${chip.label}`"
                    @click="emit('remove', chip.field)">
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
            <slot name="editor" :field="editing" />
        </filter-popover>
    </div>
</template>
