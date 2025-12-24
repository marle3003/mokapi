<script setup lang="ts">
import { type PropType, computed } from 'vue';
import Markdown from 'vue3-markdown-it';

const props = defineProps({
    servers: { type: Array as PropType<Array<HttpServer>>, required: true },
})

const sortedServers = computed(() => {
    return props.servers.sort(comparePath)
})

function comparePath(s1: HttpServer, s2: HttpServer) {
    const name1 = s1.url.toLowerCase()
    const name2 = s2.url.toLowerCase()
    return name1.localeCompare(name2)
}

</script>

<template>
    <section class="card">
        <div class="card-body">
            <h2 id="servers" class="card-title text-center">Servers</h2>
                <table class="table dataTable" data-testid="servers" aria-labelledby="servers">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left">Url</th>
                            <th scope="col" class="text-left">Description</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="server in sortedServers" :key="server.url">
                            <td>{{ server.url }}</td>
                            <td><markdown :source="server.description" class="description" :html="true"></markdown></td>
                        </tr>
                    </tbody>
                </table>
            </div>
    </section>
</template>