<script setup lang="ts">
import { useService } from '@/composables/services';
import { useMetrics } from '@/composables/metrics';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRouter, useRoute } from 'vue-router';
import Markdown from 'vue3-markdown-it';

const {fetchServices} = useService()
const {sum} = useMetrics()
const {format} = usePrettyDates()
const route = useRoute()
const router = useRouter()
const services = fetchServices('http')

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
            <table class="table dataTable selectable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left w-25">Name</th>
                        <th scope="col" class="text-left w-50">Description</th>
                        <th scope="col" class="text-center" style="width:15%">Last Request</th>
                        <th scope="col" class="text-center">Requests / Errors</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="service in services" key="service.name" @click="goToService(service)">
                        <td>{{ service.name }}</td>
                        <td><markdown :source="service.description" class="description"></markdown></td>
                        <td class="text-center">{{ lastRequest(service) }}</td>
                        <td class="text-center">
                            <span>{{ requests(service) }}</span>
                            <span> / </span>
                            <span v-bind:class="{'text-danger': errors(service) > 0}">{{ errors(service) }}</span>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>

<style>
.description p {
    font-size: 0.75rem !important;
    margin: 0 !important;
    line-height: 1.6 !important;
}
</style>