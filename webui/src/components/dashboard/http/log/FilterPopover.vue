<script setup lang="ts">
import { onMounted, onUnmounted, useTemplateRef } from 'vue'

defineProps<{ label: string }>()
const emit = defineEmits<{ close: [] }>()

const root = useTemplateRef<HTMLElement>('root')

function closeOnOutsideClick(e: MouseEvent) {
    if (!root.value?.contains(e.target as Node)) {
        emit('close')
    }
}

onMounted(() => {
    document.addEventListener('mousedown', closeOnOutsideClick)
    root.value?.querySelector<HTMLElement>('input, select')?.focus()
})
onUnmounted(() => document.removeEventListener('mousedown', closeOnOutsideClick))
</script>

<template>
    <!-- Escape is handled here as well: the key handler of the log ignores keys typed into fields -->
    <div ref="root" class="filter-popover" role="dialog" :aria-label="label" @keydown.esc.stop="emit('close')">
        <slot />
    </div>
</template>
