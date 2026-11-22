<script setup lang="ts">
import { usePrettyText } from '@/composables/usePrettyText'

// The error of a run on top, then what the script logged. The backend sends `logs: null` for a run without logs.
defineProps<{ data: JobExecution }>()

const { parseUrls } = usePrettyText()
</script>

<template>
    <div class="alert alert-danger" role="alert" v-if="data.error">
        <h5 class="alert-heading">Error</h5>
        <p class="mb-0">{{ data.error.message }}</p>
    </div>
    <ul class="list-group mb-3" v-if="data.logs?.length">
        <li v-for="(log, index) in data.logs" :key="index" class="list-group-item">
            <span class="text-body log">
                <span class="bi bi-exclamation-triangle-fill text-warning" v-if="log.level == 'warn'"></span>
                <span class="bi bi-x-circle-fill text-danger" v-else-if="log.level == 'error'"></span>
                <span class="bi bi-bug-fill text-info" v-else-if="log.level == 'debug'"></span>
                <span class="bi bi-chat-dots text-primary" v-else></span>
                <span class="ms-2" v-html="parseUrls(log.message)"></span>
            </span>
        </li>
    </ul>
    <p v-else>This job did not produce any logs. You can use <code>console.log()</code> or <code>console.error()</code> in your script to output information.</p>
</template>

<style scoped>
.log {
    font-family: 'Fira Code', monospace;
}
</style>
