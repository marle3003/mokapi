<script setup lang="ts">
import { type PropType, computed } from 'vue'
import { usePrettyBytes } from '@/composables/usePrettyBytes'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { usePrettyHttp } from '@/composables/http'
import { useRoute } from 'vue-router';

const prop = defineProps({
    event: { type: Object as PropType<ServiceEvent>, required: true },
})

const {format, duration} = usePrettyDates()
const formatBytes = usePrettyBytes().format
const {formatStatusCode, getClassByStatusCode} = usePrettyHttp()
const eventData = computed(() => <HttpEventData>prop.event.data)
const route = useRoute()

function operation(){
    const endpoint = prop.event.traits.path.substring(1).split('/')
    endpoint.push(prop.event.traits.method.toLowerCase())

    return {
        name: 'httpEndpoint',
        params: { service: prop.event.traits.name, endpoint: endpoint },
        query: { refresh: route.query.refresh }
    }
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="row">
                <div class="col header">
                    <p class="label">URL</p>
                    <p>
                        <span class="badge operation" :class="eventData.request.method.toLowerCase()">{{ eventData.request.method }}</span>
                        {{ eventData.request.url }}
                    </p>
                </div>
                <div class="col">
                    <p class="label">Time</p>
                    <p>{{ format(event.time) }}</p>
                </div>
                <div class="col" v-if="eventData.deprecated">
                    <p class="label">Warning</p>
                    <p><i class="bi bi-exclamation-triangle-fill yellow"></i> Deprecated</p>
                </div>
            </div>
            <div class="row">
                <div class="col">
                    <p class="label">Status</p>
                    <p>
                        <span class="badge status-code" :class="getClassByStatusCode(eventData.response.statusCode)">
                            {{ formatStatusCode(eventData.response.statusCode) }}
                        </span>
                    </p>
                </div>
                <div class="col">
                    <p class="label">Duration</p>
                    <p>{{ duration(eventData.duration) }}</p>
                </div>
                <div class="col">
                    <p class="label">Specification</p>
                    <router-link :to="operation()">Operation</router-link>
                </div>
            </div>
            <div class="row">
                <div class="col">
                    <p class="label">Size</p>
                    <p>{{ formatBytes(eventData.response.size) }}</p>
                </div>
            </div>
        </div>
    </div>
</template>