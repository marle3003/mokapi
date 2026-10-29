<script setup lang="ts">
import Loading from '@/components/Loading.vue';
import Message from '@/components/Message.vue';
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRoute } from '@/router';
import { computed, ref, watch } from 'vue';
import type { EventResult } from '@/types/dashboard';
import { useWebsocket } from '@/composables/websocket';
import Actions from '../Actions.vue'

const props = defineProps<{
    service: WebsocketService,
}>();

const route = useRoute();
const { format: formatTime } = usePrettyDates();
const { dashboard } = useDashboard()
const { formatType, formatAddress } = useWebsocket()

const event = ref<ServiceEvent | null>(null);
const isLoading = ref<boolean>(true)

watch(() => dashboard.value,
  (db, _, onCleanup) => {
    const res = db.getEvent(route.params.id as string);

    const stop = watch([res.event, res.isLoading], ([evt, loading]) => { event.value = evt; isLoading.value = loading }, { immediate: true })

    onCleanup(() => {
        stop()
        res.close()
    });
  },
  { immediate: true }
);

const eventData = computed(() => {
    if (!event.value) {
        return undefined
    }
    return event.value.data as WebsocketConnectionLog
})
function isClientAvailable(id: string) {
  return  props.service.clients?.find(x => x.id === id) !== undefined
}
const channel = computed(() => {
  if (!eventData.value) {
    return undefined
  }
  const name = eventData.value.channel
  return props.service.channels.find(x => x.name === name || x.instances?.find(y => y.name === name) !== undefined
  )
})
</script>

<template>
    <div v-if="event && eventData">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="row mb-2">
                        <div class="col-3 header">
                            <p id="api-key" class="label">Type</p>
                            <p aria-labelledby="api-key">
                                {{ formatType(eventData.type) }}
                            </p>
                        </div>
                        <div class="col text-end">
                            <p id="time" class="label">Time</p>
                            <p aria-labelledby="time">{{ formatTime(event.time) }}</p>
                        </div>
                        <div class="col text-end">
                            <span class="badge bg-secondary" aria-label="Service Type">Websocket</span>
                        </div>
                    </div>
                    <div class="row mb-2" v-if="channel">
                        <div class="col">
                            <p id="channel" class="label">Channel</p>
                            <div aria-labelledby="channel">
                                <router-link :to="{
                                    name: getRouteName('websocketChannel').value,
                                    params: {service: event.traits.name, channel: channel.name},
                                }">
                                {{ eventData.channel }}
                                </router-link>
                            </div>
                        </div>
                    </div>
                    <div class="row">
                        <div class="col">
                            <p id="clientId" class="label">Client</p>
                            <p aria-labelledby="clientId">
                                <router-link v-if="eventData.client && isClientAvailable(eventData.client.id)" :to="{
                                    name: getRouteName('websocketClient').value,
                                    params: {service: event.traits.name, id: eventData.client.id},
                                }">
                                {{ formatAddress(eventData.client.address) }}
                                </router-link>
                                <span v-else-if="eventData.client">{{ formatAddress(eventData.client.address) }}</span>
                                <span v-else>-</span>
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="card-group">
            <section class="card" aria-labelledby="actions">
                <div class="card-body">
                    <h2 id="actions" class="card-title text-center">Event Handlers</h2>
                    <actions :actions="eventData.actions" />
                </div>
            </section>
        </div>
        
    </div>
    <loading v-if="!event || isLoading"></loading>
    <div v-if="!event && !isLoading">
        <message message="WebSocket Event not found"></message>
    </div>
</template>