<script setup lang="ts">
import { useConnection } from '@/composables/connection'

const { isConnected } = useConnection()
</script>

<template>
    <!-- The live region exists before the banner appears, so screen readers announce it. -->
    <div class="connection-lost-region" role="status" aria-live="polite">
        <div v-if="!isConnected" class="connection-lost-banner">
            <span class="bi bi-exclamation-triangle-fill me-2" aria-hidden="true"></span>
            <span class="connection-lost-text">Connection lost. Please restart your Mokapi. Retrying...</span>
        </div>
    </div>
</template>

<style scoped>
/* --connection-banner-height is defined in dashboard.css, where the layouts make room for it */
.connection-lost-banner {
    display: flex;
    align-items: center;
    height: var(--connection-banner-height);
    padding: 0 1rem;
    font-size: 12px;
    border-bottom: 1px solid var(--bs-warning-border-subtle);
    background-color: var(--bs-warning-bg-subtle);
    color: var(--bs-warning-text-emphasis);
    white-space: nowrap;
}
.connection-lost-text {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
}
</style>
