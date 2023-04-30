<script setup lang="ts">
import type { PropType } from 'vue';
import Markdown from 'vue3-markdown-it';

defineProps({
    servers: { type: Array as PropType<Array<string>>, required: true },
})
</script>

<template>
    <table class="table dataTable">
        <thead>
            <tr>
                <th scope="col" class="text-left">URL</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="server in servers" :key="server">
                <td>{{ server }}</td>
            </tr>
        </tbody>
    </table>
</template>