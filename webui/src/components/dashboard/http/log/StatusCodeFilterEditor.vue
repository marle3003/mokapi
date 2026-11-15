<script setup lang="ts">
const statusCode = defineModel<string>({ required: true })
</script>

<template>
    <input type="text" class="form-control form-control-sm" v-model="statusCode" aria-label="Status Code" placeholder="4xx">
    <div class="form-text">
        <b>Status Code Filter:</b>
        <ul class="list-unstyled mb-0">
            <li>&bull; Multiple entries (comma-separated): 200, 404</li>
            <li>&bull; Negation: prefix with '!' (e.g., !501)</li>
            <li>&bull; Range: start-end (e.g., 200-300)</li>
            <li>&bull; Status class: 4xx for all client errors</li>
        </ul>
    </div>
</template>
