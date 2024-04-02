<script setup lang="ts">
import { useService } from '@/composables/services'
import { useMetrics } from '@/composables/metrics'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { useRouter, useRoute } from 'vue-router'
import { onUnmounted } from 'vue'
import Markdown from 'vue3-markdown-it'

const { fetchServices } = useService()
const { sum, max } = useMetrics()
const { format } = usePrettyDates()
const route = useRoute()
const router = useRouter()
const { services, close } = fetchServices('kafka')

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

function goToService(service: Service){
    if (getSelection()?.toString()) {
        return
    }

    router.push({
        name: 'kafkaService',
        params: {service: service.name},
        query: {refresh: route.query.refresh}
    })
}

onUnmounted(() => {
    close()
})
</script>

<template>
    <section class="card" aria-labelledby="clusters" data-testid="kafka-service-list">
        <div class="card-body">
            <div class="card-title text-center" id="clusters">Kafka Clusters</div>
            <table class="table dataTable selectable">
                <caption class="visually-hidden">Kafka Clusters</caption>
                <thead>
                    <tr>
                        <th scope="col" class="text-left w-25">Name</th>
                        <th scope="col" class="text-left w-50">Description</th>
                        <th scope="col" class="text-center" style="width: 15%">Last Record</th>
                        <th scope="col" class="text-center">Records</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="service in services" key="service.name" @click="goToService(service)">
                        <td>{{ service.name }}</td>
                        <td><markdown :source="service.description" class="description"></markdown></td>
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