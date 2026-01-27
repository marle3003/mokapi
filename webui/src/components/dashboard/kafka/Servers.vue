<script setup lang="ts">
import { computed, onMounted } from 'vue'
import Markdown from 'vue3-markdown-it'
import { Popover } from 'bootstrap'

const props = defineProps<{
    servers: KafkaServer[],
}>()

onMounted(() => {
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

const servers = computed(() => {
    if (!props.servers) {
        return undefined;
    }
    return props.servers.sort((x: KafkaServer, y: KafkaServer) => {
        return x.name.localeCompare(y.name);
    })
})
</script>

<template>
    <div class="table-responsive-sm">
        <table class="table dataTable table-responsive-lg" aria-label="Servers">
            <thead>
                <tr>
                    <th scope="col" class="text-left col-2">Name</th>
                    <th scope="col" class="text-left col-3">Host</th>
                    <th scope="col" class="text-left col">Description</th>
                    <th scope="col" class="text-left col-1">Tags</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="server in servers" :key="server.host">
                    <td>{{ server.name }}</td>
                    <td v-html="server.host.replace(/([^:]*):(.*)/g, '$1<wbr>:$2')"></td>
                    <td>
                        <markdown :source="server.description" class="description" :html="true"></markdown>
                    </td>
                    <td>
                        <ul class="tags">
                            <li v-for="tag in server.tags" class="has-popover">
                                {{ tag.name }}
                                <span style="display:none">{{ tag.description }}</span>
                            </li>

                        </ul>
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
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