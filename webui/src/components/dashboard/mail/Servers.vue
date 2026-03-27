<script setup lang="ts">
import { useMarkdown } from '@/composables/markdown';
import { computed, type PropType } from 'vue';

const props = defineProps({
    service: { type: Object as PropType<MailService>, required: true },
})

const servers = computed(() => {
    if (!props.service.servers) {
        return undefined
    }
    return props.service.servers.sort((x: MailServer, y: MailServer) => x.name.localeCompare(y.name));
})
</script>

<template>
    <div class="table-responsive-sm">
        <table class="table dataTable" aria-label="Servers">
            <thead>
                <tr>
                    <th scope="col" class="text-left">Name</th>
                    <th scope="col" class="text-left">Host</th>
                    <th scope="col" class="text-left">Protocol</th>
                    <th scope="col" class="text-left w-50">Description</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="server in servers" :key="server.name">
                    <td>{{ server.name }}</td>
                    <td>{{ server.host }}</td>
                    <td>{{ server.protocol }}</td>
                    <td><div v-html="useMarkdown(server.description).content" class="description"></div></td>
                </tr>
            </tbody>
        </table>
    </div>
</template>