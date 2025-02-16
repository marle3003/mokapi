<script setup lang="ts">
import { computed, type PropType } from 'vue'
import Actions from '../Actions.vue'
import { usePrettyDates } from '@/composables/usePrettyDate';

const props = defineProps<{
   event: ServiceEvent
}>()

const { format, duration } = usePrettyDates()

const data = computed((): {data: LdapEventData, request: LdapModifyRequest, response: LdapResponse} => {
    const data = <LdapEventData>props.event.data
    return { data: data, request: <LdapModifyRequest>data.request, response: <LdapResponse>data.response }
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
                            <p>Modify {{ data.request.dn }}</p>
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
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="card-title text-center">Modifications</div>
                    <table class="table dataTable">
                        <thead>
                            <tr>
                                <th scope="col" class="text-left" style="width: 20%">Modification</th>
                                <th scope="col" class="text-left" style="width: 20%">Type</th>
                                <th scope="col" class="text-left">Values</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="item of data.request.items">
                                <td>{{ item.modification }}</td>
                                <td>{{ item.attribute.type }}</td>
                                <td><p v-for="v of item.attribute.values">{{ v }}</p></td>
                            </tr>
                        </tbody>
                    </table>
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