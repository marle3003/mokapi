<script setup lang="ts">
import { useRouter } from 'vue-router';
import { type PropType, computed, reactive, ref, watch } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { getRouteName, useDashboard } from '@/composables/dashboard';

const props = defineProps({
    service: { type: Object as PropType<MailService> },
})

const labels: Label[] = [{ name: 'namespace', value: 'mail' }]
if (props.service) {
    labels.push({name: 'name', value: props.service!.name})
}

const paging = reactive({
    current: 1,
    size: 25,
    sizeOptions: [10, 25, 50, 100]
})


const router = useRouter()
const { format, duration } = usePrettyDates()

const { dashboard } = useDashboard()
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


function goToMail(evt: ServiceEventInfo, openInNewTab = false){
    const data = eventData(evt)

    const to = {
        name: getRouteName('smtpMail').value,
        params: {id: data.messageId},
    }
    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
}
function eventData(event: ServiceEventInfo): SmtpEventInfo {
    return <SmtpEventInfo>event.data
}

const filteredEvents = computed(() => {
    if (!events.value) {
        return []
    }
    return events.value.filter(e => {
        const data = eventData(e)
        return data.messageId
    })
})
</script>

<template>
    <div class="card">
        <div class="card-body">
            <h2 id="mails-title" class="card-title text-center">Recent Mails</h2>
            <div class="table-responsive-sm">
                <table class="table dataTable selectable" aria-labelledby="mails-title">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left col-4">Subject</th>
                            <th scope="col" class="text-left col-2">From</th>
                            <th scope="col" class="text-left col-2">To</th>
                            <th scope="col" class="text-center col-2">Time</th>
                            <th scope="col" class="text-center col-1">Duration</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="event in filteredEvents" :key="event.id" @mouseup.left="goToMail(event)" @mousedown.middle="goToMail(event, true)">
                            <td>
                                <router-link @click.stop class="row-link" :to="{ name: getRouteName('smtpMail').value, params: {id: eventData(event).messageId} }">
                                    {{ eventData(event).subject }}
                                </router-link>
                            </td>
                            <td>{{ eventData(event).from }}</td>
                            <td class="address-list">
                                <div v-for="address in eventData(event).to">
                                {{ address }}
                                </div>
                            </td>
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
    </div>
</template>

<style scoped>
.address-list ul {
    margin-bottom: 0;
}
</style>