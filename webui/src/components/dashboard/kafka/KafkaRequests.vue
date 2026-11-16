<script setup lang="ts">
import { useTemplateRef } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { getRouteName } from '@/composables/dashboard'
import { useRequestLog } from '@/composables/use-request-log'
import Loading from '@/components/Loading.vue'
import Paging from '../Paging.vue'
import LogPauseButton from '../log/LogPauseButton.vue'
import KafkaRequestDetailPanel from './log/KafkaRequestDetailPanel.vue'
import { kafkaRequestTypes } from './requests/request-types'

const props = defineProps<{
  service: KafkaService,
  clientId: string
}>()

const { format: formatTime } = usePrettyDates()

const table = useTemplateRef<HTMLElement>('table')
const {
  paging, result, isLoading, events, rows, paused, canPause, togglePause,
  selectedId, select, onRowClick, onLinkClick,
} = useRequestLog<KafkaRequestLog>('kafka', 'request', {
  get serviceName() { return props.service.name },
  get clientId() { return props.clientId },
}, table)
</script>

<template>
  <div class="log-layout" :class="{ 'has-panel': selectedId }">
    <section class="card" aria-labelledby="request-log-title">
      <div class="card-body">
        <h2 id="request-log-title" class="card-title text-center">Recent Requests</h2>

        <p class="text-muted small text-center mb-3">
          This view shows requests for producer initialization, group coordination, and offset resolution
        </p>

        <div class="log-toolbar">
          <log-pause-button v-if="canPause" :paused="paused" @toggle="togglePause" />
        </div>

        <div class="table-responsive">
          <table ref="table" class="table dataTable selectable" aria-label="Recent Requests">
            <thead>
              <tr>
                <th scope="col" class="text-left col-2">API Key</th>
                <th scope="col" class="text-left col-2">Version</th>
                <th scope="col" class="text-left col">Summary</th>
                <th scope="col" class="text-left col">Error</th>
                <th scope="col" class="text-center col-2">Time</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="{ event, data } in rows" :key="event.id"
                :aria-selected="event.id === selectedId ? 'true' : undefined" @click="onRowClick($event, event.id)">
                <td>
                  <router-link custom :to="{ name: getRouteName('kafkaRequest').value, params: { service: service.name, id: event.id } }" v-slot="{ href }">
                    <a class="row-link" :href="href" @click="onLinkClick">{{ data.header.requestName }}</a>
                  </router-link>
                </td>
                <td>{{ data.header.version }}</td>
                <td>
                  <component :is="kafkaRequestTypes[data.header.requestKey]?.summary" :request="data.request"
                    :response="data.response" />
                </td>
                <td>{{ data.response.errorCode && data.response.errorCode !== '' ? data.response.errorCode : 'None' }}</td>
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

    <kafka-request-detail-panel v-if="selectedId" :service="service.name" :request-id="selectedId" @close="select(undefined)" />
  </div>
</template>
