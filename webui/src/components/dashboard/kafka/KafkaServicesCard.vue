<script setup lang="ts">
import { useService } from '@/composables/services';
import { useMetrics } from '@/composables/metrics';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRouter, useRoute } from 'vue-router';

const {fetchServices} = useService()
const {sum} = useMetrics()
const {format} = usePrettyDates()
const route = useRoute()
const router = useRouter()
const services = fetchServices('kafka')

function lastMessage(service: Service){
    const n = sum(service.metrics, 'kafka_message_timestamp')
    if (n == 0){
        return '-'
    }
    return format(n)
}

function messages(service: Service){
    return sum(service.metrics, 'kafka_messages_total')
}

function goToService(service: Service){
    router.push({
        name: 'kafkaService',
        params: {service: service.name},
        query: {refresh: route.query.refresh}
    })
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Kafka Clusters</div>
            <table class="table dataTable selectable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left w-25">Name</th>
                        <th scope="col" class="text-left w-50">Description</th>
                        <th scope="col" class="text-center" style="width: 15%">Last Message</th>
                        <th scope="col" class="text-center">Messages</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="service in services" key="service.name" @click="goToService(service)">
                        <td>{{ service.name }}</td>
                        <td>{{ service.description }}</td>
                        <td class="text-center">{{ lastMessage(service) }}</td>
                        <td class="text-center">{{ messages(service) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>

<style scoped>
.description p {
    font-size: 0.75rem !important;
    margin: 0 !important
}
</style>