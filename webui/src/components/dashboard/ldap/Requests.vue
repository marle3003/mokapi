<script setup lang="ts">
import { useRouter } from 'vue-router';
import { type PropType, onUnmounted } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { getRouteName, useDashboard } from '@/composables/dashboard';

const props = defineProps({
    service: { type: Object as PropType<LdapService> },
})

const labels = []
if (props.service){
    labels.push({name: 'name', value: props.service!.name})
}

const router = useRouter()
const { dashboard } = useDashboard();
const { events, close } = dashboard.value.getEvents('ldap', ...labels)
const { format, duration } = usePrettyDates()
let data: LdapEventData

function goToEvent(event: ServiceEvent){
    if (getSelection()?.toString()) {
        return
    }

    router.push({
        name: getRouteName('ldapRequest').value,
        params: {id: event.id},
    })
}
function eventData(event: ServiceEvent): LdapEventData{
    return <LdapEventData>event.data
}

function targetDn(data: LdapEventData) {
    switch (data.request.operation) {
        case 'Search':
            return data.request.baseDN
        case 'Modify':
        case 'Add':
        case 'Delete':
        case 'ModifyDN':
        case 'Compare':
            return data.request.dn
    }
}

onUnmounted(() => {
    close()
})
</script>

<template>
    <section class="card" aria-labelledby="requests">
        <div class="card-body">
            <h2 id="requests" class="card-title text-center">Recent Requests</h2>
            <table class="table dataTable selectable" aria-labelledby="requests">
                <thead>
                    <tr>
                        <th scope="col" class="text-left" style="width: 10%">Operation</th>
                        <th scope="col" class="text-left" style="width: 25%">DN</th>
                        <th scope="col" class="text-left" style="width: 30%">Filter</th>
                        <th scope="col" class="text-center"  style="width: 10%">Status</th>
                        <th scope="col" class="text-center" style="width:15%">Time</th>
                        <th scope="col" class="text-center" style="width: 10%">Duration</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="event in events!" :key="event.id" @click="goToEvent(event)" :set="data = eventData(event)">
                        <td>
                            {{ data.request.operation }}
                        </td>
                        <td>
                            {{ targetDn(data) }}
                        </td>
                        <td>
                            {{ data.request.operation === 'Search' ? data.request.filter : '' }}
                        </td>
                        <td class="text-center">{{ data.response.status }}</td>
                        <td class="text-center">{{ format(event.time) }}</td>
                        <td class="text-center">{{ duration(data.duration) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </section>
</template>