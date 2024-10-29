<script setup lang="ts">
import { onMounted } from 'vue'
import Markdown from 'vue3-markdown-it'
import { Popover } from 'bootstrap'

defineProps<{
    servers: KafkaServer[],
}>()

onMounted(()=> {
    const elements = document.querySelectorAll('.has-popover')
    const popovers = [...elements].map(x => {
        new Popover(x, {
            customClass: 'custom-popover',
            trigger: 'hover',
            html: true,
            placement: 'left',
            content: () => x.querySelector('span')?.innerHTML ?? '',
        })
    })
})
</script>

<template>
    <table class="table dataTable">
        <caption class="visually-hidden">Cluster Brokers</caption>
        <thead>
            <tr>
                <th scope="col" class="text-left">Name</th>
                <th scope="col" class="text-left">URL</th>
                <th scope="col" class="text-left">Tags</th>
                <th scope="col" class="text-left w-50">Description</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="server in servers" :key="server.url">
                <td>{{ server.name }}</td>
                <td>{{ server.url }}</td>
                <td>
                    <ul class="tags">
                        <li v-for="tag in server.tags" class="has-popover">
                            {{ tag.name }}
                            <span style="display:none">{{ tag.description }}</span>
                        </li>
                        
                    </ul>
                </td>
                <td><markdown :source="server.description" class="description" html="true"></markdown></td>
            </tr>
        </tbody>
    </table>
</template>

<style scoped>
ul.tags {
    list-style: none; 
    padding: 0;
    margin: 0;
}
ul.tags li {
    padding-right: 0.5em;
}
</style>