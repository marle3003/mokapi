<script setup lang="ts" generic="T extends string">
// The sub tabs of a detail panel. The selected tab belongs to the panel, not to the address: the
// next row opens on the same tab. The panel puts the content of the tab below, in a `tabpanel`.
defineProps<{ tabs: readonly T[] }>()
const active = defineModel<T>({ required: true })
</script>

<template>
    <div class="nav card-tabs" role="tablist">
        <button v-for="tab in tabs" :key="tab" type="button" role="tab" :class="{ active: active === tab }"
            :aria-selected="active === tab" @click="active = tab">
            {{ tab }}
        </button>
    </div>
</template>
