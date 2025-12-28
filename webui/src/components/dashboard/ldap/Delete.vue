<script setup lang="ts">
import { computed } from 'vue'
import Actions from '../Actions.vue'
import { usePrettyDates } from '@/composables/usePrettyDate';

const props = defineProps<{
   event: ServiceEvent
}>()

const { format, duration } = usePrettyDates()

const data = computed((): {data: LdapEventData, request: LdapAddRequest, response: LdapResponse} => {
    const data = <LdapEventData>props.event.data
    return { data: data, request: <LdapAddRequest>data.request, response: <LdapResponse>data.response }
})

const hasActions = computed(() => {
    return data.value.data.actions?.length > 0
})
</script>

<template>
    <div v-if="event">
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