<script setup lang="ts">
import { computed, onMounted } from 'vue'
import Markdown from 'vue3-markdown-it'
import { Popover } from 'bootstrap'

const props = defineProps<{
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
    <section class="card" aria-labelledby="servers">
        <div class="card-body">
            <h2 id="servers" class="card-title text-center">Brokers</h2>
         
            <table class="table dataTable" aria-labelledby="servers">
                <thead>
                    <tr>
                        <th scope="col" class="text-left" style="width: 20%">Name</th>
                        <th scope="col" class="text-left" style="width: 20%">Host</th>
                        <th scope="col" class="text-left" style="width: 20%">Tags</th>
                        <th scope="col" class="text-left w-50" style="width: 40%">Description</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="server in servers" :key="server.host">
                        <td>{{ server.name }}</td>
                        <td>{{ server.host }}</td>
                        <td>
                            <ul class="tags">
                                <li v-for="tag in server.tags" class="has-popover">
                                    {{ tag.name }}
                                    <span style="display:none">{{ tag.description }}</span>
                                </li>
                                
                            </ul>
                        </td>
                        <td><markdown :source="server.description" class="description" :html="true"></markdown></td>
                    </tr>
                </tbody>
            </table>

         </div>
    </section>
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