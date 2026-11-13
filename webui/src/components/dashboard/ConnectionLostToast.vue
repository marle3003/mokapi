<script setup lang="ts">
import { useConnection } from '@/composables/connection'

const { isConnected } = useConnection()
</script>

<template>
    <div v-if="!isConnected" class="toast show position-fixed bottom-0 end-0 m-3" role="status" aria-live="polite" aria-atomic="true">
        <div class="toast-header">
            <span class="bi bi-exclamation-triangle-fill text-warning me-2"></span>
            <strong>Connection lost</strong>
        </div>
        <div class="toast-body">
            Please restart your Mokapi. Retrying...
        </div>
    </div>
</template>

<style scoped>
.toast {
    --bs-toast-bg: var(--color-background-soft);
    --bs-toast-color: var(--color-text);
    --bs-toast-border-color: var(--color-border);
    --bs-toast-header-color: var(--color-text);
    --bs-toast-header-bg: var(--color-background-soft);
    --bs-toast-header-border-color: var(--color-border);
}
</style>
