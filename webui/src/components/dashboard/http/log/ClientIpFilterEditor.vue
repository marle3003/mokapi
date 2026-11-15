<script setup lang="ts">
const clientIP = defineModel<string>({ required: true })
</script>

<template>
    <input type="text" class="form-control form-control-sm" v-model="clientIP" aria-label="Client IP">
    <div class="form-text">
        <b>IP Filter Options:</b>
        <ul class="list-unstyled mb-0">
            <li>&bull; Multiple entries (comma-separated): 192.168.0.1,10.0.0.1</li>
            <li>&bull; Negation: prefix with '!' (e.g., !127.0.0.1)</li>
            <li>&bull; CIDR notation: 192.168.0.0/24</li>
        </ul>
    </div>
</template>
