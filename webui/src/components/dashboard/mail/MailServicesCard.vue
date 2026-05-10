<script setup lang="ts">
import { onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useMarkdown } from '@/composables/markdown';

const { dashboard } = useDashboard();
const {services, close} = dashboard.value.getServices('mail')
const {format} = usePrettyDates()
const router = useRouter()

function lastMail(service: Service){
    const n = service.metrics.mail_mail_timestamp
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
                        <td><div v-html="useMarkdown(service.description).content" class="table-markdown"></div></td>
                        <td class="text-center">{{ lastMail(service) }}</td>
                        <td class="text-center">{{ service.metrics.mail_mails_total }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>