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

function goToSearch(data: ServiceEvent){
    router.push({
        name: 'ldapSearch',
        params: {id: data.id},
    })
}
function eventData(event: ServiceEvent): LdapEventData{
    return <LdapEventData>event.data
}

function scopeToString(scope: SearchScope): string {
    switch (scope) {
        case 3: return 'WholeSubtree'
        case 2: return 'SingleLevel'
        case 1: return 'BaseObject'
        default: return 'invalid'
    }
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
                        <th scope="col" class="text-center">Duration</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="event in events!" :key="event.id" @click="goToSearch(event)">
                        <td>
                            {{ eventData(event).request.filter }}
                        </td>
                        <td>
                            {{ eventData(event).request.baseDN }}
                        </td>
                        <td>
                            {{ scopeToString(eventData(event).request.scope) }}
                        </td>
                        <td class="text-center">{{ eventData(event).request.sizeLimit }}</td>
                        <td class="text-center">{{ duration(eventData(event).duration) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>