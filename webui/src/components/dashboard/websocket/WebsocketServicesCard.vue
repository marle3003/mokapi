<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate'
import { useRouter } from 'vue-router'
import { onUnmounted } from 'vue'
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useMarkdown } from '@/composables/markdown'

const { format } = usePrettyDates()
const router = useRouter()
const { dashboard } = useDashboard();
const { services, close } = dashboard.value.getServices('websocket');

function lastMessage(service: Service){
    const n = service.metrics.websocket_message_timestamp
    if (n == 0){
        return '-'
    }
    return format(n)
}

function goToService(service: Service, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('websocketService').value,
        params: {service: service.name},
    };
    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
}

onUnmounted(() => {
    close()
})
</script>

<template>
    <section class="card" aria-labelledby="websocket-services-title">
        <div class="card-body">
            <h2 class="card-title text-center" id="websocket-services-title">Websocket Services</h2>
            <table class="table dataTable selectable" aria-labelledby="websocket-services-title">
                <thead>
                    <tr>
                        <th scope="col" class="text-left w-25">Name</th>
                        <th scope="col" class="text-left w-50">Description</th>
                        <th scope="col" class="text-center" style="width: 15%">Last Message</th>
                        <th scope="col" class="text-center">Messages</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="service in services" key="service.name" @mouseup.left="goToService(service)" @mousedown.middle="goToService(service, true)">
                        <td>
                            <router-link @click.stop class="row-link" :to="{ name: getRouteName('websocketService').value, params: {service: service.name} }">
                                {{ service.name }}
                            </router-link>
                        </td>
                        <td><div v-html="useMarkdown(service.description).content" class="table-markdown"></div></td>
                        <td class="text-center">{{ lastMessage(service) }}</td>
                        <td class="text-center">{{ service.metrics.websocket_messages_total }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </section>
</template>