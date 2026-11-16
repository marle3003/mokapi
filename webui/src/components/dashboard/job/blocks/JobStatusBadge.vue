<script setup lang="ts">
import type { JobStatus } from '@/composables/job'

defineProps<{ status: JobStatus }>()

const badgeClass: Record<JobStatus, string> = { Success: 'bg-success', Warning: 'bg-warning', Error: 'bg-danger' }
</script>

<template>
    <span class="badge" :class="badgeClass[status]">{{ status }}</span>
</template>
