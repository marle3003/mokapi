<script setup lang="ts">
import { computed, ref, watch, type PropType } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import MetricCard from '../MetricCard.vue'
import { useDashboard } from '@/composables/dashboard';

const props = defineProps({
    includeError: Boolean,
    onlyError: Boolean,
    to: { type: [String, Object] as PropType<RouteLocationRaw> }
})

const { dashboard } = useDashboard()
const services = ref<HttpService[] | null>(null)
watch(
    () => dashboard.value,
    (db, _, onCleanup) => {
        if (!db) {
            return
        }

        const res = db.getServices('http')

        const stop = watch(
            () => res.services.value,
            (v) => {
                services.value = v as HttpService[]
            },
            { immediate: true }
        )

        onCleanup(() => {
            stop();
            res.close();
        });
    }, { immediate: true }
)

const requests = computed(() => {
    if (!services.value) {
        return 0
    }

    let result = 0;
    for (const service of services.value) {
        result += service.metrics.requests_total || 0
    }
    return result;
})
const errors = computed(() => {
    if (!services.value) {
        return 0
    }

    let result = 0;
    for (const service of services.value) {
        result += service.metrics.requests_errors_total || 0
    }
    return result;
})

const title = computed(() => {
    if (props.includeError) {
        return 'HTTP Requests / Errors'
    }
    return props.onlyError ? 'HTTP Errors' : 'HTTP Requests'
})
</script>

<template>
    <metric-card :title="title" :to="to" data-testid="metric-http-requests">
        <template #value>
            <span v-if="!onlyError">{{ requests }}</span>
            <span v-if="includeError"> / </span>
            <span v-bind:class="{'text-danger': errors > 0}" v-if="onlyError || includeError">{{  errors }}</span>
        </template>
    </metric-card>
</template>
