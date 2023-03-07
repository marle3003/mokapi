<script setup lang="ts">
import type { PropType } from 'vue';
import Markdown from 'vue3-markdown-it';

defineProps({
    servers: { type: Array as PropType<Array<KafkaServer>>, required: true },
})
</script>

<template>
    <table class="table dataTable">
        <thead>
            <tr>
                <th scope="col" class="text-left">Name</th>
                <th scope="col" class="text-left">Url</th>
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