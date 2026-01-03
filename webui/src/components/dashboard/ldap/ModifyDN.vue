<script setup lang="ts">
import { computed } from 'vue'
import Actions from '../Actions.vue'
import { usePrettyDates } from '@/composables/usePrettyDate';

const props = defineProps<{
   event: ServiceEvent
}>()

const { format, duration } = usePrettyDates()

const data = computed((): {data: LdapEventData, request: LdapModifDNRequest, response: LdapResponse} => {
    const data = <LdapEventData>props.event.data
    return { data: data, request: <LdapModifDNRequest>data.request, response: <LdapResponse>data.response }
})

const hasActions = computed(() => {
    return data.value.data.actions?.length > 0
})
</script>

<template>
    <div v-if="event">
        <div class="card-group">
            <section class="card" aria-labelledby="request-title">
                <div class="card-body">
                    <h2 id="request-title" class="card-title text-center">Request</h2>
                    <div class="row">
                        <div class="col-2">
                            <p class="label">New RDN</p>
                            <p>{{ data.request.newRdn }}</p>
                        </div>
                        <div class="col-2">
                            <p class="label">Delete Old RDN</p>
                            <p>{{ data.request.deleteOldDn }}</p>
                        </div>
                        <div class="col">
                            <p class="label">New Superior DN</p>
                            <p>{{ data.request.newSuperiorDn }}</p>
                        </div>
                    </div>
                </div>
            </section>
        </div>
        <div class="card-group" v-if="hasActions">
            <div class="card">
                <div class="card-body">
                    <h2 class="card-title text-center">Actions</h2>
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