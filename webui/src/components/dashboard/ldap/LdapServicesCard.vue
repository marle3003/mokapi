<script setup lang="ts">
import { useMetrics } from '@/composables/metrics';
import { onUnmounted } from 'vue';
import Markdown from 'vue3-markdown-it'
import { useRouter, useRoute } from 'vue-router';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { getRouteName, useDashboard } from '@/composables/dashboard';

const {sum} = useMetrics()
const { dashboard } = useDashboard();
const {services, close} = dashboard.value.getServices('ldap')
const {format} = usePrettyDates()
const router = useRouter()

function lastRequest(service: Service){
    const n = sum(service.metrics, 'ldap_request_timestamp')
    if (n == 0){
        return '-'
    }
    return format(n)
}

function requests(service: Service) {
    return sum(service.metrics, 'ldap_requests_total')
}

function goToService(service: Service, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('ldapService').value,
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
    <div class="card" data-testid="smtp-service-list">
        <div class="card-body">
            <h2 id="ldap-servers-title" class="card-title text-center">LDAP Servers</h2>
            <table class="table dataTable selectable" aria-labelledby="ldap-servers-title">
                <thead>
                    <tr>
                        <th scope="col" class="text-left w-25">Name</th>
                        <th scope="col" class="text-left w-50">Description</th>
                        <th scope="col" class="text-center" style="width:15%">Last Request</th>
                        <th scope="col" class="text-center">Requests</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="service in services" key="service.name" @mouseup.left="goToService(service)" @mousedown.middle="goToService(service, true)">
                        <td>
                            <router-link @click.stop class="row-link" :to="{ name: getRouteName('ldapService').value, params: {service: service.name} }">
                                {{ service.name }}
                            </router-link>
                        </td>
                        <td><markdown :source="service.description" class="description" :html="true"></markdown></td>
                        <td class="text-center">{{ lastRequest(service) }}</td>
                        <td class="text-center">{{ requests(service) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>