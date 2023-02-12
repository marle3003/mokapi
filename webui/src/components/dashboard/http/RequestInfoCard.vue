<script setup lang="ts">
import { type PropType, computed } from 'vue';
import { usePrettyBytes } from '@/composables/usePrettyBytes';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { usePrettyHttp } from '@/composables/http';

const prop = defineProps({
    event: { type: Object as PropType<ServiceEvent>, required: true },
})

const {format, duration} = usePrettyDates()
const formatBytes = usePrettyBytes().format
const formatStatusCode = usePrettyHttp().format
const eventData = computed(() => <HttpEventData>prop.event.data)

function getClassByStatusCode(statusCode: number) {
    switch (true) {
        case statusCode >= 200 && statusCode < 300:
            return 'success'
        case statusCode >= 300 && statusCode < 400:
            return 'redirect'
        case statusCode >= 400 && statusCode < 500:
            return 'client-error'
        case statusCode >= 500 && statusCode < 600:
            return 'server-error'
    }
    return ''
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="row">
                <div class="col header">
                    <p class="label">URL</p>
                    <p>{{ eventData.request.url }}</p>
                </div>
                <div class="col">
                    <p class="label">Time</p>
                    <p>{{ format(event.time) }}</p>
                </div>
            </div>
            <div class="row">
                <div class="col">
                    <p class="label">Method</p>
                    <p><span class="badge operation" :class="eventData.request.method.toLowerCase()">{{ eventData.request.method }}</span></p>
                </div>
                <div class="col">
                    <p class="label">Duration</p>
                    <p>{{ duration(eventData.duration) }}</p>
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
                    <p class="label">Size</p>
                    <p>{{ formatBytes(eventData.response.size) }}</p>
                </div>
            </div>
        </div>
    </div>
</template>