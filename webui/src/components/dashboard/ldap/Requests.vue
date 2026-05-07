<script setup lang="ts">
import { useRouter } from 'vue-router';
import { type PropType, onUnmounted } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useLdap } from '@/composables/ldap';

const props = defineProps({
    service: { type: Object as PropType<LdapService> },
})

const labels = [{ name: 'namespace', value: 'ldap' }]
if (props.service){
    labels.push({name: 'name', value: props.service!.name})
}

const router = useRouter()
const { dashboard } = useDashboard();
const { events, close } = dashboard.value.getEvents(...labels)
const { format, duration } = usePrettyDates()
const { targetDn, criteria } = useLdap()
let data: LdapEventData

function goToEvent(event: ServiceEvent, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('ldapRequest').value,
        params: {id: event.id},
    }
    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
}
function eventData(event: ServiceEvent): LdapEventData{
    return <LdapEventData>event.data
}

onUnmounted(() => {
    close()
})
</script>

<template>
    <section class="card" aria-labelledby="requests">
        <div class="card-body">
            <h2 id="requests" class="card-title text-center">Recent Requests</h2>
            <div class="table-responsive-sm">
                <table class="table dataTable selectable" aria-labelledby="requests">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left col-1">Operation</th>
                            <th scope="col" class="text-left col">DN</th>
                            <th scope="col" class="text-left col">Criteria</th>
                            <th scope="col" class="text-center col-1">Status</th>
                            <th scope="col" class="text-center col-2">Time</th>
                            <th scope="col" class="text-center col-1">Duration</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="event in events!" :key="event.id" @mouseup.left="goToEvent(event)" @mousedown.middle="goToEvent(event, true)" :set="data = eventData(event)">
                            <td>
                                <router-link @click.stop class="row-link" :to="{ name: getRouteName('ldapRequest').value, params: {id: event.id} }">
                                    <span class="badge ldap operation" :class="eventData(event).request.operation.toLocaleLowerCase()">
                                    {{ eventData(event).request.operation }}
                                    </span>
                                </router-link>
                            </td>
                            <td>
                                {{ targetDn(data) }}
                            </td>
                            <td>
                                {{ criteria(data)}}
                            </td>
                            <td class="text-center">{{ data.response?.status ?? '-' }}</td>
                            <td class="text-center">{{ format(event.time) }}</td>
                            <td class="text-center">{{ duration(data.duration) }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </section>
</template>