<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useEvents } from '@/composables/events';
import { type PropType, onUnmounted } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';

const props = defineProps({
    service: { type: Object as PropType<LdapService> },
})

const labels = []
if (props.service){
    labels.push({name: 'name', value: props.service!.name})
}

const router = useRouter()
const { fetch } = useEvents()
const { events, close } = fetch('ldap', ...labels)
const { duration } = usePrettyDates()
let data: LdapEventData

function goToSearch(data: ServiceEvent){
    router.push({
        name: 'ldapSearch',
        params: {id: data.id},
    })
}
function eventData(event: ServiceEvent): LdapEventData{
    return <LdapEventData>event.data
}

onUnmounted(() => {
    close()
})
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Recent Searches</div>
            <table class="table dataTable selectable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left" style="width: 20%">Filter</th>
                        <th scope="col" class="text-left" style="width: 20%">Base DN</th>
                        <th scope="col" class="text-left" style="width: 20%">Scope</th>
                        <th scope="col" class="text-center" style="width:15%">Size Limit</th>
                        <th scope="col" class="text-center" style="width:15%">Results</th>
                        <th scope="col" class="text-center" style="width:15%">Status Code</th>
                        <th scope="col" class="text-center">Duration</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="event in events!" :key="event.id" @click="goToSearch(event)" :set="data = eventData(event)">
                        <td>
                            {{ data.request.filter }}
                        </td>
                        <td>
                            {{ data.request.baseDN }}
                        </td>
                        <td>
                            {{ data.request.scope }}
                        </td>
                        <td class="text-center">{{ data.request.sizeLimit }}</td>
                        <td class="text-center">{{ data.response.results?.length ?? 0 }}</td>
                        <td class="text-center">{{ data.response.status }}</td>
                        <td class="text-center">{{ duration(data.duration) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>