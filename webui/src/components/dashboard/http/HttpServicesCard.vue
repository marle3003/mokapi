<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRoute } from '@/router';
import { onUnmounted } from 'vue';
import { useDashboard } from '@/composables/dashboard';
import { useMarkdown } from '@/composables/markdown';

const { format } = usePrettyDates()
const { service: serviceRoute, router } = useRoute()
const { dashboard } = useDashboard()
const { services, close } = dashboard.value.getServices('http')

function lastRequest(s: Service) {
    const n = s.metrics.http_request_timestamp
    if (n == 0) {
        return '-'
    }
    return format(n)
}

function requests(s: Service) {
    return s.metrics.http_requests_total
}

function errors(s: Service) {
    return s.metrics.http_requests_errors_total
}

function goToService(s: Service, openInNewTab = false) {
    if (getSelection()?.toString()) {
        return
    }

    const to = serviceRoute(s, 'http');
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
    <div class="card" data-testid="http-service-list">
        <div class="card-body">
            <h2 id="http-apis-title" class="card-title text-center">HTTP APIs</h2>
            <div class="table-responsive-sm">
                <table class="table dataTable selectable" aria-labelledby="http-apis-title">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left w-25">Name</th>
                            <th scope="col" class="text-left w-50">Description</th>
                            <th scope="col" class="text-center" style="width:15%">Last Request</th>
                            <th scope="col" class="text-center">Requests / Errors</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="service in services" key="service.name" @mouseup.left="goToService(service)"
                            @mousedown.middle="goToService(service, true)">
                            <td>
                                <router-link @click.stop class="row-link" :to="serviceRoute(service, 'http')">
                                    <i class="bi bi-x-circle-fill red" style="vertical-align: middle;" v-if="service.status !== 'valid'" role="img" aria-label="Service status invalid"></i>
                                    {{ service.name }}
                                </router-link>
                            </td>
                            <td>
                                <div v-html="useMarkdown(service.description).content" class="table-markdown"></div>
                            </td>
                            <td class="text-center">{{ lastRequest(service) }}</td>
                            <td class="text-center">
                                <span>{{ requests(service) }}</span>
                                <span> / </span>
                                <span v-bind:class="{ 'text-danger': errors(service) > 0 }">{{ errors(service) }}</span>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</template>