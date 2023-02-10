<script setup lang="ts">
import { useService, ServiceType, type Service} from '@/composables/services';
import { useMetrics } from '@/composables/metrics';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRouter, useRoute } from 'vue-router';

const {fetchServices} = useService()
const {sum} = useMetrics()
const {format} = usePrettyDates()
const route = useRoute()
const router = useRouter()
const services = fetchServices(ServiceType.Http)

function lastRequest(service: Service){
    const n = sum(service.metrics, 'http_request_timestamp')
    if (n == 0){
        return '-'
    }
    return format(n)
}

function requests(service: Service){
    return sum(service.metrics, 'http_requests_total')
}

function errors(service: Service){
    return sum(service.metrics, 'http_requests_errors_total')
}

function goToService(service: Service){
    router.push({
        name: 'httpService',
        params: {service: service.name},
        query: {refresh: route.query.refresh}
    })
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">HTTP Services</div>
            <div class="card-body">
                <table class="table dataTable selectable">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left">Name</th>
                            <th scope="col" class="text-left">Last Request</th>
                            <th scope="col">Requests / Errors</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="service in services" key="service.name" @click="goToService(service)">
                            <td>{{ service.name }}</td>
                            <td>{{ lastRequest(service) }}</td>
                            <td>
                                <span>{{ requests(service) }}</span>
                                <span> / </span>
                                <span v-bind:class="{'text-danger': requests(service) > 0}">{{ errors(service) }}</span>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</template>