<script setup lang="ts">
import type { PropType } from 'vue';
import type { Event } from '@/composables/events';
import { usePrettyBytes } from '@/composables/usePrettyBytes';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { usePrettyHttp } from '@/composables/http';

defineProps({
    event: { type: Object as PropType<Event>, required: true },
})

const {format, duration} = usePrettyDates()
const formatBytes = usePrettyBytes().format
const formatStatusCode = usePrettyHttp().format

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
            <div class="card-body">
                <div class="row">
                    <div class="col">
                        <p class="label">URL</p>
                        <p>{{ event.data.request.url }}</p>
                    </div>
                    <div class="col">
                        <p class="label">Time</p>
                        <p>{{ format(event.time) }}</p>
                    </div>
                </div>
                <div class="row">
                    <div class="col">
                        <p class="label">Method</p>
                        <p><span class="badge operation" :class="event.data.request.method.toLowerCase()">{{ event.data.request.method }}</span></p>
                    </div>
                    <div class="col">
                        <p class="label">Duration</p>
                        <p>{{ duration(event.data.duration) }}</p>
                    </div>
                </div>
                <div class="row">
                    <div class="col">
                        <p class="label">Status</p>
                        <p>
                            <span class="badge status-code" :class="getClassByStatusCode(event.data.response.statusCode)">
                                {{ formatStatusCode(event.data.response.statusCode) }}
                            </span>
                        </p>
                    </div>
                    <div class="col">
                        <p class="label">Size</p>
                        <p>{{ formatBytes(event.data.response.size) }}</p>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>