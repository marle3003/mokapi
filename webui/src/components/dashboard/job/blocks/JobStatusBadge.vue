<script setup lang="ts">
import { computed } from 'vue'
import { jobStatus, type JobStatus } from '@/composables/job'

const props = defineProps<{ data: JobExecution }>()

const badgeClass: Record<JobStatus, string> = { Success: 'bg-success', Warning: 'bg-warning', Error: 'bg-danger' }
const status = computed(() => jobStatus(props.data))
</script>

<template>
    <span class="badge" :class="badgeClass[status]">{{ status }}</span>
</template>
