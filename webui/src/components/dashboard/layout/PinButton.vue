<script setup lang="ts">
import { computed, type PropType } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import { useRouter } from '@/router'
import { useDashboardLayout } from '@/composables/dashboard-layout'
import { usePinned } from '@/composables/pinned'

const props = defineProps({
    to: { type: [String, Object] as PropType<RouteLocationRaw>, required: true },
    label: { type: String, required: true },
})

const router = useRouter()
const { layout } = useDashboardLayout()
const { isPinned, toggle } = usePinned()

const path = computed(() => router.resolve(props.to).fullPath)
const pinned = computed(() => isPinned(path.value))
</script>

<template>
    <!-- Pinned entries only exist in the new layout -->
    <button v-if="layout === 'new'" type="button" class="pin-button" :class="{ pinned }"
        :aria-label="`Pin ${label}`" :aria-pressed="pinned" :title="pinned ? 'Unpin' : 'Pin'"
        @click.stop="toggle(path)" @mousedown.stop @mouseup.stop>
        <i class="bi" :class="pinned ? 'bi-star-fill' : 'bi-star'" aria-hidden="true"></i>
    </button>
</template>
