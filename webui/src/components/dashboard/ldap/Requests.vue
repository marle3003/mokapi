<script setup lang="ts">
import { useRouter } from 'vue-router';
import { type PropType, computed, reactive, ref, watch } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { usePageSize } from '@/composables/paging';
import { useLdap } from '@/composables/ldap';

const props = defineProps({
    service: { type: Object as PropType<LdapService> },
})
const { targetDn, criteria } = useLdap()

const labels = [{ name: 'namespace', value: 'ldap' }]
if (props.service){
    labels.push({name: 'name', value: props.service!.name})
}

const paging = reactive({
    current: 1,
    size: usePageSize(25),
})

const router = useRouter()
const { format, duration } = usePrettyDates()

const { dashboard } = useDashboard();
const result = ref<ServiceEventPage | null>(null);
const events = computed(() => result.value?.items ?? [])
watch(
  [() => dashboard.value, () => paging.current, () => paging.size, labels ],
  ([db], _, onCleanup) => {
    const res = db.getEvents({
      page: paging.current,
      size: paging.size,
      labels: labels,
    })

    const stop = watch(res.data, (v) => { result.value = v }, { immediate: true })

    onCleanup(() => {
      stop()
      res.close()
    })
  },
  { immediate: true }
);

function goToEvent(eventId: string, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('ldapRequest').value,
        params: {id: eventId},
    }
    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
}
function eventData(event: ServiceEvent): LdapEventData {
    return <LdapEventData>event.data
}
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
                        <tr v-for="event in events!" :key="event.id" @mouseup.left="goToEvent(event.id)" @mousedown.middle="goToEvent(event.id, true)">
                            <td>
                                <router-link @click.stop class="row-link" :to="{ name: getRouteName('ldapRequest').value, params: {id: event.id} }">
                                    <span class="badge ldap operation" :class="eventData(event).request.operation.toLocaleLowerCase()">
                                    {{ eventData(event).request.operation }}
                                    </span>
                                </router-link>
                            </td>
                            <td>
                                {{ targetDn(eventData(event)) }}
                            </td>
                            <td v-html="criteria(eventData(event))"></td>
                            <td class="text-center">{{ eventData(event).response?.status ?? '-' }}</td>
                            <td class="text-center">{{ format(event.time) }}</td>
                            <td class="text-center">{{ duration(eventData(event).duration) }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
            
            <Paging v-if="result"
                :count="events.length"
                :total="result.total"
                v-model:current="paging.current"
                v-model:size="paging.size"
            />
    
        </div>
    </section>
</template>