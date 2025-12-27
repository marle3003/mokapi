<script setup lang="ts">
import { useMetrics } from '@/composables/metrics'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { useRouter, useRoute } from 'vue-router'
import { onUnmounted } from 'vue'
import Markdown from 'vue3-markdown-it'
import { getRouteName, useDashboard } from '@/composables/dashboard';

const { sum, max } = useMetrics()
const { format } = usePrettyDates()
const route = useRoute()
const router = useRouter()
const { dashboard } = useDashboard();
const { services, close } = dashboard.value.getServices('kafka')

function lastMessage(service: Service){
    const n = max(service.metrics, 'kafka_message_timestamp')
    if (n == 0){
        return '-'
    }
    return format(n)
}

function messages(service: Service){
    return sum(service.metrics, 'kafka_messages_total')
}

function goToService(service: Service, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('kafkaService').value,
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
    <section class="card" aria-labelledby="clusters" data-testid="kafka-service-list">
        <div class="card-body">
            <h2 class="card-title text-center" id="clusters">Kafka Clusters</h2>
            <table class="table dataTable selectable">
                <caption class="visually-hidden">Kafka Clusters</caption>
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
                            <router-link @click.stop class="row-link" :to="{ name: getRouteName('kafkaService').value, params: {service: service.name} }">
                                {{ service.name }}
                            </router-link>
                        </td>
                        <td><markdown :source="service.description" class="description" :html="true"></markdown></td>
                        <td class="text-center">{{ lastMessage(service) }}</td>
                        <td class="text-center">{{ messages(service) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </section>
</template>

<style scoped>
.description p {
    font-size: 0.75rem !important;
    margin: 0 !important
}
</style>