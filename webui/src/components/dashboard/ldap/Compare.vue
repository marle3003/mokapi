<script setup lang="ts">
import { computed } from 'vue'
import Actions from '../Actions.vue'
import { usePrettyDates } from '@/composables/usePrettyDate';

const props = defineProps<{
   event: ServiceEvent
}>()

const { format, duration } = usePrettyDates()

const data = computed((): {data: LdapEventData, request: LdapCompareRequest, response: LdapResponse} => {
    const data = <LdapEventData>props.event.data
    return { data: data, request: <LdapCompareRequest>data.request, response: <LdapResponse>data.response }
})

const hasActions = computed(() => {
    return data.value.data.actions?.length > 0
})
</script>

<template>
    <div v-if="event">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="row">
                        <div class="col header">
                            <p class="label">Operation</p>
                            <p>Compare {{ data.request.dn }}</p>
                        </div>
                        <div class="col-2">
                            <p class="label">Time</p>
                            <p>{{ format(event.time) }}</p>
                        </div>
                        <div class="col-2">
                            <p class="label">Duration</p>
                            <p>{{ duration(data.data.duration) }}</p>
                        </div>
                    </div>
                    <div class="row">
                        <div class="col-2">
                            <p class="label">Attribute</p>
                            <p>{{ data.request.attribute }}</p>
                        </div>
                        <div class="col">
                            <p class="label">Value</p>
                            <p>{{ data.request.value }}</p>
                        </div>
                    </div>
                    <div class="row">
                        <div class="col-2">
                            <p class="label">Status</p>
                            <p>{{ data.response.status }}</p>
                        </div>
                        <div class="col" v-if="data.response.message">
                            <p class="label">Message</p>
                            <p>{{ data.response.message }}</p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <div class="card-group" v-if="hasActions">
            <div class="card">
                <div class="card-body">
                    <div class="card-title text-center">Actions</div>
                    <actions :actions="data.data.actions" />
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.row {
    padding-bottom: 10px;
}
</style>