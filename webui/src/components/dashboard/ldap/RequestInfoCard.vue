<script setup lang="ts">
import { type PropType, computed } from 'vue'
import { usePrettyBytes } from '@/composables/usePrettyBytes'
import { usePrettyDates } from '@/composables/usePrettyDate'

const props = defineProps({
    event: { type: Object as PropType<ServiceEvent>, required: true },
})

const { format } = usePrettyDates()
const { duration } = usePrettyDates()
const data = computed(() => <LdapEventData>props.event?.data)
const request = computed(() => {
    switch (data.value.request.operation) {
        case 'Search': return data.value.request.filter;
        case 'Add': return data.value.request.dn;
        case 'Compare': return data.value.request.dn;
        case 'Modify': return data.value.request.dn;
        case 'ModifyDN': return data.value.request.dn;
        case 'Delete': return data.value.request.dn;
    }
})
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="row mb-2">
                <div class="col-10 header">
                    <p class="label">Operation</p>
                    <p>
                        <span class="badge operation" :class="data.request.operation.toLowerCase()">{{ data.request.operation }}</span>
                        {{ request }}
                    </p>
                </div>
                <div class="col">
                    <p class="label">Time</p>
                    <p>{{ format(event.time) }}</p>
                </div>
            </div>
            <div class="row">
                <div class="col-2">
                    <p class="label">Status</p>
                    <p>{{ data.response.status }}</p>
                </div>
                <div class="col-2">
                    <p class="label">Duration</p>
                    <p>{{ duration(data.duration) }}</p>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.operation.search {
    background-color: var(--color-blue);
}
</style>