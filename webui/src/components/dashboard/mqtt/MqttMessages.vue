<script setup lang="ts">
import { ref, computed, reactive, watch } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import router from '@/router'
import { getRouteName, useDashboard } from '@/composables/dashboard'
import { usePageSize } from '@/composables/paging';
import Paging from '../Paging.vue';

const props = defineProps<{
    service?: MqttService,
    topicName?: string
    instanceName?: string
    clientId?: string
}>()

const emit = defineEmits<{
    (e: "loaded", count: number): void
}>();

const labels = computed(() => {
    const result = [{ name: 'namespace', value: 'mqtt' }];
    if (props.service) {
        result.push({ name: 'name', value: props.service.name })
    }
    if (props.topicName) {
        result.push({ name: 'topic', value: props.topicName })
    }
    if (props.instanceName) {
        result.push({ name: 'instance', value: props.instanceName })
    }
    result.push({ name: 'type', value: 'message' })
    if (props.clientId) {
        result.push({ name: 'clientId', value: props.clientId })
    }
    return result;
})

const paging = reactive({
    current: 1,
    size: usePageSize(25),
})

const { format } = usePrettyDates()

const { dashboard } = useDashboard()
const result = ref<ServiceEventResult | null>(null);
const events = computed(() => result.value?.items ?? [])
watch(
    [() => dashboard.value, () => paging.current, () => paging.size, labels],
    ([db], _, onCleanup) => {
        const res = db.getEvents({
            page: paging.current,
            size: paging.size,
            labels: labels.value,
        })

        const stop = watch(res.data, (v) => { result.value = v }, { immediate: true })

        onCleanup(() => {
            stop()
            res.close()
        })
    },
    { immediate: true }
);
watch(() => events.value, (events) => emit("loaded", events.length), { immediate: true })

function eventData(event: ServiceEventInfo): MqttMessageInfo {
    return event.data as MqttMessageInfo
}

function goToMessage(eventId: string, openInNewTab = false) {
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('mqttMessage').value,
        params: { id: eventId }
    }
    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
}
const isTemplateTopic = computed(() => {
    if (!props.service || !props.topicName) {
        return false
    }
    const topic = props.service.topics.find(x => x.name === props.topicName)
    if (!topic || !topic.instances) {
        return false
    }
    return topic.instances.length > 0
})
</script>

<template>
    <div class="table-responsive">
        <table class="table dataTable selectable" aria-label="Recent Messages">
            <thead>
                <tr>
                    <th scope="col" class="text-left col-2" v-if="!topicName || isTemplateTopic">Topic</th>
                    <th scope="col" class="text-left col-4">Value</th>
                    <th scope="col" class="text-center col-2">Time</th>

                </tr>
            </thead>
            <tbody>
                <tr v-for="event in events" :key="event.id" @click="goToMessage(event.id)"
                    @mousedown.middle="goToMessage(event.id, true)">
                    <td v-if="!topicName || isTemplateTopic">
                        <router-link @click.stop class="row-link"
                            :to="{ name: getRouteName('mqttMessage').value, params: { id: event.id } }">
                            {{ eventData(event).topic }}
                        </router-link>
                    </td>
                    <td v-if="topicName" class="message">
                        <router-link @click.stop class="row-link"
                            :to="{ name: getRouteName('mqttMessage').value, params: { id: event.id } }">
                            {{ eventData(event).message.value ?? eventData(event).message.binary }}
                        </router-link>
                    </td>
                    <td v-else class="message">{{ eventData(event).message.value ?? eventData(event).message.binary }}
                    </td>
                    <td class="text-center">{{ format(event.time) }}</td>
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
</template>

<style scoped>
.message,
.key {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 0;
}

.tab-pane {
    padding: 0;
}

table.dataTable tr.deleted td {
    color: #5E5E5E;
}
</style>