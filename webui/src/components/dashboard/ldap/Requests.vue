<script setup lang="ts">
import { useRouter } from 'vue-router';
import { type PropType, computed, reactive, ref, watch } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { getRouteName, useDashboard } from '@/composables/dashboard';

const props = defineProps({
    service: { type: Object as PropType<LdapService> },
})

const labels = [{ name: 'namespace', value: 'ldap' }]
if (props.service){
    labels.push({name: 'name', value: props.service!.name})
}

const paging = reactive({
    current: 1,
    size: 25,
    sizeOptions: [10, 25, 50, 100]
})

const router = useRouter()
const { format, duration } = usePrettyDates()

const { dashboard } = useDashboard();
const result = ref<ServiceEventResult | null>(null);
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
function eventData(event: ServiceEventInfo): LdapEventInfo {
    return <LdapEventInfo>event.data
}
function criteria(value: string | string[]): string {
    if (Array.isArray(value)) {
        return value.join('<br />')
    }
    return value
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
                                    <span class="badge ldap operation" :class="eventData(event).operation.toLocaleLowerCase()">
                                    {{ eventData(event).operation }}
                                    </span>
                                </router-link>
                            </td>
                            <td>
                                {{ eventData(event).dn }}
                            </td>
                            <td v-html="criteria(eventData(event).criteria)"></td>
                            <td class="text-center">{{ eventData(event).status ?? '-' }}</td>
                            <td class="text-center">{{ format(event.time) }}</td>
                            <td class="text-center">{{ duration(eventData(event).duration) }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
            <!-- Paging -->
            <div class="row" v-if="result && result.total > result.size">
                 <div class="d-flex justify-content-between align-items-center mb-3">
                    
                    <div class="small text-muted">
                        Showing {{ events.length }} of {{ result.total }} requests
                    </div>

                        <Pager :current="result.page" :size="10" :total="result.total" />

                        <select
                            class="form-select form-select-sm"
                            style="width:auto"
                            v-model.number="result.size"
                        >
                            <option
                                v-for="size in paging.sizeOptions"
                                :key="size"
                                :value="size"
                            >
                                {{ size }}
                            </option>
                        </select>

                 </div>
            </div>
        </div>
    </section>
</template>