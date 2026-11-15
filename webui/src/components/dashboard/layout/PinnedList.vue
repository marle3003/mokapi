<script setup lang="ts">
import { computed, type PropType } from 'vue'
import { isPinTargetAvailable, usePinned } from '@/composables/pinned'
import { useTabs } from '@/composables/use-tabs'

const props = defineProps({
    services: { type: Array as PropType<Service[]>, required: true },
    // until the services are loaded no target counts as missing
    servicesLoaded: { type: Boolean, required: true },
})

const { pinned, remove } = usePinned()
const { describe } = useTabs()

const entries = computed(() => pinned.value.map(path => ({
    path,
    ...describe(path),
    available: !props.servicesLoaded || isPinTargetAvailable(path, props.services),
})))
</script>

<template>
    <div class="sidebar-pinned">
        <h2 class="sidebar-section-title">Pinned</h2>
        <p v-if="!entries.length" class="sidebar-hint">Pin a tab with its star.</p>
        <ul v-else aria-label="Pinned">
            <li v-for="entry in entries" :key="entry.path">
                <router-link v-if="entry.available" class="sidebar-entry" :to="entry.path" :title="entry.title">
                    <span v-if="entry.method" class="badge operation" :class="entry.method">{{ entry.method.toUpperCase() }}</span>
                    <span class="sidebar-entry-name">{{ entry.label }}</span>
                </router-link>
                <div v-else class="sidebar-entry sidebar-entry-missing" title="This target does not exist anymore">
                    <!-- only the name is disabled, an aria-disabled list item would disable its remove button too -->
                    <span class="sidebar-entry-name" aria-disabled="true">{{ entry.label }}</span>
                    <button type="button" class="sidebar-remove" :aria-label="`Remove ${entry.label} from Pinned`" @click="remove(entry.path)">
                        <i class="bi bi-x" aria-hidden="true"></i>
                    </button>
                </div>
            </li>
        </ul>
    </div>
</template>
