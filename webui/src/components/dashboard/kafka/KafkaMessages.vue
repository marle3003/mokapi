<script setup lang="ts">
import { ref, computed, watch, reactive } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import router from '@/router'
import { getRouteName, useDashboard } from '@/composables/dashboard'
import { useLocalStorage } from '@/composables/local-storage'
import { usePrettyText } from '@/composables/usePrettyText'
import { usePageSize } from '@/composables/paging'
import Paging from '../Paging.vue'
import { formatKey } from '@/composables/kafka.ts'

const { format } = usePrettyDates()
const { fromBinary } = usePrettyText()

const props = defineProps<{
    service?: KafkaService,
    topicName?: string
    instanceName?: string
    clientId?: string
}>()

const emit = defineEmits<{
    (e: "loaded", count: number): void
}>();


const tags = useLocalStorage<string[]>(`kafka-${props.service?.name}-tags`, ['__all'])
const labels = computed(() => {
    const result = [{ name: 'namespace', value: 'kafka' }];
    if (props.service) {
        result.push({name: 'name', value: props.service.name})
    }
    if (props.topicName) {
        result.push({name: 'topic', value: props.topicName})
    }
    if (props.instanceName) {
        result.push({ name: 'instance', value: props.instanceName })
    }
    result.push({ name: 'type', value: 'message' })
    if (props.clientId){
        result.push({name: 'clientId', value: props.clientId})
    }
    return result;
})

const paging = reactive({
    current: 1,
    size: usePageSize(25),
})

const { dashboard } = useDashboard()
const result = ref<ServiceEventPage | null>(null);
const events = computed(() => result.value?.items ?? [])
watch(
  [() => dashboard.value, () => paging.current, () => paging.size, labels ],
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

const messages = computed(() => {
    const result = [];
    for (const event of events.value) {
        const data = eventData(event)

        if (props.service && !props.clientId && !props.topicName && !tags.value.includes('__all')) {
            const topic = props.service.topics.find(t => t.name === event.traits['topic']);
            if (!topic) {
                continue
            }
            if (!topic.tags || !topic.tags.some(tag => tags.value.some(x => x == tag.name))) {
                continue
            }
        }

        result.push({
            ...event,
            key: key(data),
            value: data.message.value ?? data.message.binary,
            deleted: data.deleted,
        });
    }
    return result;
})

const isTemplateTopic = computed(() => {
    if (!props.service || !props.topicName) {
        return false
    }
    const topic = props.service.topics.find(x => x.name === props.topicName)
    return (topic?.instances?.length ?? -1) >= 0 && !props.instanceName
})

function eventData(event: ServiceEvent): KafkaMessageData {
    return <KafkaMessageData>event.data
}

function goToMessage(eventId: string, openInNewTab = false) {
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('kafkaMessage').value,
        params: { id: eventId }
    }
    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
}
function key(data: KafkaMessageData): string {
    if (data.key.value) {
        return data.key.value
    }
    if (data?.key.binary) {
        return fromBinary(data.key.binary)
    }
    return ''
}
const showTopicColumn = computed(() => !props.topicName || isTemplateTopic.value)
</script>

<template>
    <div class="table-responsive">
        <table class="table dataTable selectable" aria-label="Recent Messages">
            <thead>
                <tr>
                    <th scope="col" class="text-left col-2" v-if="showTopicColumn === true">Topic</th>
                    <th scope="col" class="text-left col-2">Key</th>
                    <th scope="col" class="text-left col-4">Value</th>
                    <th scope="col" class="text-center col-2">Time</th>
                    
                </tr>
            </thead>
            <tbody>
                <tr v-for="msg in messages" :key="msg.id" @click="goToMessage(msg.id)" @mousedown.middle="goToMessage(msg.id, true)" :class="msg.deleted ? 'deleted': ''">
                    <td v-if="showTopicColumn === true">
                        <router-link @click.stop class="row-link"
                            :to="{ name: getRouteName('kafkaMessage').value, params: { id: msg.id } }">
                            {{ eventData(msg).topic }}
                        </router-link>
                    </td>
                    <td class="key">
                        <router-link v-if="!showTopicColumn" class="row-link"
                            @click.stop :to="{name: getRouteName('kafkaMessage').value, params: { id: msg.id }}">
                            {{ formatKey(msg.key) }}
                        </router-link>
                        <span v-else>{{ formatKey(msg.key) }}</span>
                    </td>
                    <td class="message">{{ msg.value }}</td>
                    <td class="text-center">{{ format(msg.time) }}</td>
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
.message, .key {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 0;
}
table.dataTable tr.deleted td {
    color: #5E5E5E;
}
</style>