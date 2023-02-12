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
const services = fetchServices('smtp')

function messages(service: Service){
    return sum(service.metrics, 'smtp_mails_total')
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">SMTP Servers</div>
            <table class="table dataTable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left w-25">Name</th>
                        <th scope="col" class="text-left" style="width: 65%">Description</th>
                        <th scope="col" class="text-center">Messages</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="service in services" key="service.name">
                        <td>{{ service.name }}</td>
                        <td>{{ service.description }}</td>
                        <td class="text-center">{{ messages(service) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>