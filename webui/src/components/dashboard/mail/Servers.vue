<script setup lang="ts">
import type { PropType } from 'vue';
import Markdown from 'vue3-markdown-it'

defineProps({
    service: { type: Object as PropType<MailService>, required: true },
})
</script>

<template>
    <table class="table dataTable">
        <caption class="visually-hidden">Cluster Brokers</caption>
        <thead>
            <tr>
                <th scope="col" class="text-left">Protocol</th>
                <th scope="col" class="text-left">Host</th>
                <th scope="col" class="text-left">Name</th>
                <th scope="col" class="text-left w-50">Description</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="server in service.servers" :key="server.name">
                <td>{{ server.protocol }}</td>
                <td>{{ server.host }}</td>
                <td>{{ server.name }}</td>
                <td><markdown :source="server.description" class="description" :html="true"></markdown></td>
            </tr>
        </tbody>
    </table>
</template>