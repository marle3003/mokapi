<script setup lang="ts">
import { useTemplateRef } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { getRouteName } from '@/composables/dashboard'
import { useMqtt } from '@/composables/mqtt'
import { useRequestLog } from '@/composables/use-request-log'
import Loading from '@/components/Loading.vue'
import Paging from '../Paging.vue'
import LogPauseButton from '../log/LogPauseButton.vue'
import MqttRequestDetailPanel from './log/MqttRequestDetailPanel.vue'
import { mqttRequestTypes } from './requests/request-types'

const props = defineProps<{
  service: MqttService,
  clientId: string
}>()

const { format: formatTime } = usePrettyDates()
const { formatType } = useMqtt()

const table = useTemplateRef<HTMLElement>('table')
const {
  paging, result, isLoading, events, rows, paused, canPause, togglePause,
  selectedId, select, onRowClick, onLinkClick,
} = useRequestLog<MqttRequestLog>('mqtt', 'request', {
  get serviceName() { return props.service.name },
  get clientId() { return props.clientId },
}, table)
</script>

<template>
  <div class="log-layout" :class="{ 'has-panel': selectedId }">
    <section class="card" aria-labelledby="request-log-title">
      <div class="card-body">
        <h2 id="request-log-title" class="card-title text-center">Connection & Subscription Requests</h2>

        <div class="log-toolbar">
          <log-pause-button v-if="canPause" :paused="paused" @toggle="togglePause" />
        </div>

        <div class="table-responsive">
          <table ref="table" class="table dataTable selectable" aria-label="Connection & Subscription Requests">
            <thead>
              <tr>
                <th scope="col" class="text-left col-2">Type</th>
                <th scope="col" class="text-left col">Summary</th>
                <th scope="col" class="text-center col-2">Time</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="{ event, data } in rows" :key="event.id"
                :aria-selected="event.id === selectedId ? 'true' : undefined" @click="onRowClick($event, event.id)">
                <td>
                  <router-link custom :to="{ name: getRouteName('mqttRequest').value, params: { service: service.name, id: event.id } }" v-slot="{ href }">
                    <a class="badge operation" :class="formatType(data.type).toLowerCase()" :href="href" @click="onLinkClick">{{ formatType(data.type) }}</a>
                  </router-link>
                </td>
                <td>
                  <component :is="mqttRequestTypes[data.type]?.summary" :request="data.request"
                    :response="data.response" />
                </td>
                <td class="text-center">{{ formatTime(event.time) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <loading v-if="isLoading && events.length === 0" />
        <Paging v-if="result"
          :count="events.length"
          :total="result.total"
          v-model:current="paging.current"
          v-model:size="paging.size"
        />
      </div>
    </section>

    <mqtt-request-detail-panel v-if="selectedId" :service="service.name" :request-id="selectedId" @close="select(undefined)" />
  </div>
</template>
