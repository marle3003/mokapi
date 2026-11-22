<script setup lang="ts">
import { sitePath } from '@/composables/site-base'
import { getRouteName } from '@/composables/dashboard'

// The tags of a job. The tag `file` links to the config of the script.
defineProps<{ tags: JobExecution['tags'] }>()
</script>

<template>
    <h5>
        Tags
        <a :href="sitePath('/docs/javascript-api/mokapi/eventhandler/scheduledeventargs')" class="ms-1" title="Learn how to define tags for your job" rel="noopener">
            <span class="bi bi-question-circle" aria-hidden="true"></span>
        </a>
    </h5>
    <table class="table dataTable" aria-label="Tags">
        <thead>
            <tr>
                <th scope="col" class="text-left">Name</th>
                <th scope="col" class="text-left w-75">Value</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="(value, key) of tags" :key="key">
                <td>{{ key }}</td>
                <td>
                    <router-link v-if="key === 'file' && tags.fileKey" :to="{ name: getRouteName('config').value, params: { id: tags.fileKey } }">
                        {{ value }}
                    </router-link>
                    <span v-else>{{ value }}</span>
                </td>
            </tr>
        </tbody>
    </table>
</template>

<style scoped>
h5 span.bi {
    font-size: 0.8rem;
    vertical-align: top;
}
</style>
