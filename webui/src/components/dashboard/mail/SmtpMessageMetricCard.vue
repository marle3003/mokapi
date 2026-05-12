<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import MetricCard from '../MetricCard.vue'
import { useDashboard } from '@/composables/dashboard';

const { dashboard } = useDashboard()
const services = ref<MailService[] | null>(null)
watch(
    () => dashboard.value,
    (db, _, onCleanup) => {
        if (!db) {
            return
        }

        const res = db.getServices('mail')

        const stop = watch(
            () => res.services.value,
            (v) => {
                services.value = v as MailService[]
            },
            { immediate: true }
        )

        onCleanup(() => {
            stop();
            res.close();
        });
    }, { immediate: true }
)

const sum = computed(() => {
    if (!services.value) {
        return 0
    }

    let result = 0;
    for (const service of services.value) {
        result += service.metrics.mail_mails_total || 0
    }
    return result;
})
</script>

<template>
    <metric-card title="SMTP Mails" :value="sum" data-testid="metric-smtp-messages"></metric-card>
</template>