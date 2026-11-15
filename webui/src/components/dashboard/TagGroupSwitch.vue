<script setup lang="ts">
import type { TagGrouping } from '@/composables/tag-group'

const grouping = defineModel<TagGrouping>({ required: true })
</script>

<template>
    <div class="text-center mb-2">
        <div class="btn-group btn-group-sm" role="group" aria-label="Group rows">
            <button type="button" class="btn btn-outline-secondary" :class="{ active: grouping === 'az' }"
                :aria-pressed="grouping === 'az'" @click="grouping = 'az'">A-Z</button>
            <button type="button" class="btn btn-outline-secondary" :class="{ active: grouping === 'tag' }"
                :aria-pressed="grouping === 'tag'" @click="grouping = 'tag'">By tag</button>
        </div>
    </div>
</template>
