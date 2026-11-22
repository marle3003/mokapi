<script setup lang="ts">
defineProps<{ paused: boolean }>()
defineEmits<{ toggle: [] }>()
</script>

<template>
    <div class="log-pause">
        <span v-if="paused" class="badge text-bg-warning">Paused</span>
        <button type="button" class="btn btn-outline-secondary btn-compact" @click="$emit('toggle')">
            <i class="bi" :class="paused ? 'bi-play-fill' : 'bi-pause-fill'" aria-hidden="true"></i>
            {{ paused ? 'Resume' : 'Pause' }}
        </button>
    </div>
</template>
