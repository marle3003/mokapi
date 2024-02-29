<script setup lang="ts">
import Markdown from 'vue3-markdown-it'

defineProps<{
    servers: KafkaServer[],
}>()
</script>

<template>
    <table class="table dataTable">
        <caption class="visually-hidden">Cluster Brokers</caption>
        <thead>
            <tr>
                <th scope="col" class="text-left">Name</th>
                <th scope="col" class="text-left">URL</th>
                <th scope="col" class="text-left">Description</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="server in servers" :key="server.url">
                <td>{{ server.name }}</td>
                <td>{{ server.url }}</td>
                <td><markdown :source="server.description" class="description"></markdown></td>
            </tr>
        </tbody>
    </table>
</template>