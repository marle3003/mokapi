<script setup lang="ts">
import { useDashboard } from '@/composables/dashboard';
import { useLdap } from '@/composables/ldap';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { usePrettyText } from '@/composables/usePrettyText';
import { computed, onUnmounted } from 'vue';

const { dashboard } = useDashboard()
const { events: data, close } = dashboard.value.getEvents()
const { format } = usePrettyDates()
const { parseUrls } = usePrettyText()
const { targetDn } = useLdap()

onUnmounted(() => {
    close()
})

const events = computed(() => {
    return data.value.filter((e: ServiceEvent) => e.traits.type !== 'request')
})

function title(event: ServiceEvent) {
    switch (event.traits.namespace) {
        case 'http':
            return `${event.traits.method} ${event.traits.path}`
        case 'kafka':
            return `${event.traits.topic}`
        case 'mqtt':
            return `${event.traits.topic}`
        case 'mail':
            return `${event.traits.operation} ${event.traits.recipient}`
        default:
            return `${event.traits.name}`
    }
}
</script>

<template>
    <div class="card-group">
        <section class="card" aria-label="Event Results">
            <div class="card-body">
                <table class="table dataTable selectable" aria-label="Recent Messages">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left col-2">Type</th>
                            <th scope="col" class="text-left col-2">Namespace</th>
                            <th scope="col" class="text-left col-4">Name</th>
                            <th scope="col" class="text-left col-4">Event</th>
                            <th scope="col" class="text-center col-2">Time</th>

                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="event in events" :key="event.id">
                            <td>{{ event.traits.type }}</td>
                            <td>{{ event.traits.namespace }}</td>
                            <td>{{ event.traits.name }}</td>
                            <td v-if="event.traits.type === 'log'" v-for="data in [event.data as LogMessage]">
                                <span class="bi bi-exclamation-triangle-fill text-warning" v-if="data.level == 'warn'"></span>
                                <span class="bi bi-x-circle-fill text-danger" v-else-if="data.level == 'error'"></span>
                                <span class="bi bi-bug-fill text-info" v-else-if="data.level == 'debug'"></span>
                                <span class="bi bi-chat-dots text-primary" v-else></span>
                                <span class="ms-2" v-html="parseUrls(data.message)"></span>
                            </td>
                            <td v-else-if="event.traits.namespace === 'http'">
                                <span class="badge http operation me-1" :class="event.traits.method.toLocaleLowerCase()">
                                    {{ event.traits.method }}
                                </span>
                                <span>{{ event.traits.path }}</span>
                            </td>
                            <td v-else-if="event.traits.namespace === 'ldap'">
                                <span class="badge ldap operation me-1" :class="(event.data as LdapEventData).request.operation.toLocaleLowerCase()">
                                    {{ (event.data as LdapEventData).request.operation }}
                                </span>
                                <span>{{ targetDn(event.data as LdapEventData) }}</span>
                            </td>
                            <td v-else>{{ title(event) }}</td>
                            <td class="text-center">{{ format(event.time) }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </section>
    </div>
</template>