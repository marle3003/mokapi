<script setup lang="ts">
import { useMetrics } from '@/composables/metrics';
import { onUnmounted } from 'vue';
import Markdown from 'vue3-markdown-it'
import { useRouter, useRoute } from 'vue-router';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { getRouteName, useDashboard } from '@/composables/dashboard';

const {sum} = useMetrics()
const { dashboard } = useDashboard();
const {services, close} = dashboard.value.getServices('mail')
const {format} = usePrettyDates()
const route = useRoute()
const router = useRouter()

function messages(service: Service){
    return sum(service.metrics, 'mail_mails_total')
}

function lastMail(service: Service){
    const n = sum(service.metrics, 'mail_mail_timestamp')
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
        name: getRouteName('mailService').value,
        params: {service: service.name},
    }
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
    <div class="card" data-testid="mail-service-list">
        <div class="card-body">
            <h2 id="mail-servers-title" class="card-title text-center">Mail Servers</h2>
            <table class="table dataTable selectable" aria-labelledby="mail-servers-title">
                <thead>
                    <tr>
                        <th scope="col" class="text-left w-25">Name</th>
                        <th scope="col" class="text-left w-50">Description</th>
                        <th scope="col" class="text-center" style="width:15%">Last Mail</th>
                        <th scope="col" class="text-center">Messages</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="service in services" key="service.name" @mouseup.left="goToService(service)" @mousedown.middle="goToService(service, true)">
                        <td>
                            <router-link @click.stop class="row-link" :to="{ name: getRouteName('mailService').value, params: {service: service.name} }">
                                {{ service.name }}
                            </router-link>
                        </td>
                        <td><markdown :source="service.description" class="description" :html="true"></markdown></td>
                        <td class="text-center">{{ lastMail(service) }}</td>
                        <td class="text-center">{{ messages(service) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>