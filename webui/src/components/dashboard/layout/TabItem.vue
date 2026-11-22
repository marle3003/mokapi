<script setup lang="ts">
import type { PropType } from 'vue'
import PinButton from '@/components/dashboard/layout/PinButton.vue'
import type { TabInfo } from '@/composables/use-tabs'

defineProps({
    tab: { type: Object as PropType<TabInfo>, required: true },
    active: { type: Boolean, default: false },
    width: { type: Number, required: true },
})

defineEmits<{ activate: [], close: [] }>()
</script>

<template>
    <li class="tab-item" :class="{ active }" :style="{ width: `${width}px` }">
        <router-link class="tab-link" :to="tab.path" :title="tab.title" @click="$emit('activate')" @auxclick.middle.prevent="$emit('close')">
            <span v-if="tab.method" class="badge operation" :class="tab.method">{{ tab.method.toUpperCase() }}</span>
            <span class="tab-title">{{ tab.label }}</span>
        </router-link>
        <pin-button :to="tab.path" :label="tab.label" />
        <button type="button" class="tab-close" :aria-label="`Close ${tab.label}`" @click="$emit('close')">
            <i class="bi bi-x" aria-hidden="true"></i>
        </button>
    </li>
</template>
