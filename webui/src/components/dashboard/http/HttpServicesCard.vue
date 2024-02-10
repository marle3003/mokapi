<script setup lang="ts">
import { useService } from '@/composables/services';
import { useMetrics } from '@/composables/metrics';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRoute } from '@/router';
import Markdown from 'vue3-markdown-it';
import { onUnmounted } from 'vue';

const { fetchServices } = useService()
const { sum, max } = useMetrics()
const { format } = usePrettyDates()
const { service, router } = useRoute()
const { services, close } = fetchServices('http')

function lastRequest(s: Service){
    const n = max(s.metrics, 'http_request_timestamp')
    if (n == 0){
        return '-'
    }
    return format(n)
}

function requests(s: Service){
    return sum(s.metrics, 'http_requests_total')
}

function errors(s: Service){
    return sum(s.metrics, 'http_requests_errors_total')
}

function goToService(s: Service) {
    if (getSelection()?.toString()) {
        return
    }

    router.push(service(s))
}

onUnmounted(() => {
    close()
})
</script>

<template>
    <div class="card" data-testid="http-service-list">
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