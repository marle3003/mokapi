<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'

// The shell of the detail panel beside a log: the protocol puts its content into the slot.
defineProps<{
    // the accessible name of the panel, e.g. "Request details"
    label: string
    // the full view of the selected row, left out for a row that has none
    fullView?: RouteLocationRaw
}>()
defineEmits<{ close: [] }>()
</script>

<template>
    <section class="log-detail-panel" :aria-label="label">
        <div class="log-detail-actions">
            <router-link v-if="fullView" :to="fullView">Open as tab</router-link>
            <button type="button" class="btn-close" aria-label="Close" @click="$emit('close')"></button>
        </div>
        <slot />
    </section>
</template>
