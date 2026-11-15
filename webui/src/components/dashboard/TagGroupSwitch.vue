<script setup lang="ts">
import type { TagGrouping } from '@/composables/tag-group'

const grouping = defineModel<TagGrouping>({ required: true })
</script>

<template>
    <div class="btn-group btn-group-sm" role="group" aria-label="Group rows">
        <button type="button" class="btn btn-outline-secondary" :class="{ active: grouping === 'az' }"
            :aria-pressed="grouping === 'az'" @click="grouping = 'az'">A-Z</button>
        <button type="button" class="btn btn-outline-secondary" :class="{ active: grouping === 'tag' }"
            :aria-pressed="grouping === 'tag'" @click="grouping = 'tag'">By tag</button>
    </div>
</template>

<style scoped>
/* Bootstrap sets the button variables on each .btn, so they can't be inherited from the group */
.btn-group-sm > .btn {
    --bs-btn-padding-y: .1rem;
    --bs-btn-padding-x: .5rem;
    --bs-btn-font-size: .75rem;
}
</style>
