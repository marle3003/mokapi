<script setup lang="ts">
import { useService } from '@/composables/services';
import { useMetrics } from '@/composables/metrics';
import { onUnmounted } from 'vue';
import Markdown from 'vue3-markdown-it'
import { useRouter, useRoute } from 'vue-router';
import { usePrettyDates } from '@/composables/usePrettyDate';

const {fetchServices} = useService()
const {sum} = useMetrics()
const {services, close} = fetchServices('smtp')
const {format} = usePrettyDates()
const route = useRoute()
const router = useRouter()

function messages(service: Service){
    return sum(service.metrics, 'smtp_mails_total')
}

function lastMail(service: Service){
    const n = sum(service.metrics, 'smtp_mail_timestamp')
    if (n == 0){
        return '-'
    }
    return format(n)
}

function goToService(service: Service){
    router.push({
        name: 'smtpService',
        params: {service: service.name},
        query: {refresh: route.query.refresh}
    })
}

onUnmounted(() => {
    close()
})
</script>

<template>
    <div class="card" data-testid="smtp-service-list">
        <div class="card-body">
            <div class="card-title text-center">SMTP Servers</div>
            <table class="table dataTable selectable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left w-25">Name</th>
                        <th scope="col" class="text-left w-50">Description</th>
                        <th scope="col" class="text-center" style="width:15%">Last Mail</th>
                        <th scope="col" class="text-center">Messages</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="service in services" key="service.name" @click="goToService(service)">
                        <td>{{ service.name }}</td>
                        <td><markdown :source="service.description" class="description"></markdown></td>
                        <td class="text-center">{{ lastMail(service) }}</td>
                        <td class="text-center">{{ messages(service) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>