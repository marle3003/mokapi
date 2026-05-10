<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useDashboard } from '@/composables/dashboard';

defineProps({
    includeError: Boolean,
    onlyError: Boolean
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
        result += service.metrics.http_requests_total || 0
    }
    return result;
})
const errors = computed(() => {
    if (!services.value) {
        return 0
    }

    let result = 0;
    for (const service of services.value) {
        result += service.metrics.http_requests_errors_total || 0
    }
    return result;
})
</script>

<template>
    <div class="card metric text-center" data-testid="metric-http-requests">
        <div class="card-body">
            <div class="card-title" v-if="includeError">HTTP Requests / Errors</div>
            <div class="card-title" v-else-if="onlyError">HTTP Errors</div>
            <div class="card-title" v-else>HTTP Requests</div>
            <div class="card-text">
                <span v-if="!onlyError">{{ requests }}</span>
                <span v-if="includeError"> / </span>
                <span v-bind:class="{'text-danger': errors > 0}" v-if="onlyError || includeError">{{  errors }}</span>
            </div>
        </div>
    </div>
</template>