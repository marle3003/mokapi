<script setup lang="ts">
import { computed } from 'vue'
import Actions from '../Actions.vue'

const props = defineProps<{
   event: ServiceEvent
}>()

const data = computed((): {data: LdapEventData, request: LdapAddRequest, response: LdapResponse} => {
    const data = <LdapEventData>props.event.data
    return { data: data, request: <LdapAddRequest>data.request, response: <LdapResponse>data.response }
})

const hasActions = computed(() => {
    return data.value.data.actions?.length > 0
})
const attributes = computed(() => {
    return data.value.request.attributes.sort((x: LdapAttribute, y: LdapAttribute) => x.type.localeCompare(y.type));
})
</script>

<template>
    <div v-if="event">
        <div class="card-group">
            <section class="card" aria-labelledby="request-title">
                <div class="card-body">
                    <h2 id="request-title" class="card-title text-center">Request</h2>
                    <table class="table dataTable">
                        <thead>
                            <tr>
                                <th scope="col" class="text-left" style="width: 20%">Type</th>
                                <th scope="col" class="text-left">Values</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="attr of attributes">
                                <td>{{ attr.type }}</td>
                                <td>
                                    <p v-for="v of attr.values">{{ v }}</p>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </section>
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