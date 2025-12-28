<script setup lang="ts">
import { computed } from 'vue'
import Actions from '../Actions.vue'

const props = defineProps<{
   event: ServiceEvent
}>()

const data = computed((): {data: LdapEventData, request: LdapSearchRequest, response: LdapSearchResponse} => {
    const data = <LdapEventData>props.event.data
    return { data: data, request: <LdapSearchRequest>data.request, response: <LdapSearchResponse>data.response }
})

const attributes = computed(() => {
    if (!data.value.request.attributes) {
    return ''
    }
    return data.value.request.attributes.join(', ')
})

const searchResults = computed(() => {
    const response = <LdapSearchResponse>data.value.response
    if (!response.results) {
    return []
    }
    return response.results.sort(compareResult)
})

function compareResult(r1: LdapSearchResult, r2: LdapSearchResult) {
    const name1 = r1.dn.toLowerCase()
    const name2 = r2.dn.toLowerCase()
    return name1.localeCompare(name2)
}

const hasActions = computed(() => {
    return data.value.data.actions?.length > 0
})
</script>

<template>
    <div class="card-group">
        <section class="card" aria-labelledby="request">
            <div class="card-body">
                <h2 id="request" class="card-title text-center">Request</h2>
                <div class="row">
                    <div class="col">
                        <p class="label">Base DN</p>
                        <p>{{ data.request.baseDN }}</p>
                    </div>
                </div>
                <div class="row">
                    <div class="col-2">
                        <p class="label">Scope</p>
                        <p>{{ data.request.scope }}</p>
                    </div>
                    <div class="col-2">
                        <p class="label">Size Limit</p>
                        <p>{{ data.request.sizeLimit > 0 ? data.request.sizeLimit : 'no limit' }}</p>
                    </div>
                    <div class="col-2">
                        <p class="label">Time Limit</p>
                        <p>{{ data.request.timeLimit > 0 ? data.request.timeLimit + ' [s]' : 'no limit' }}</p>
                    </div>
                </div>
                <div class="row" v-if="attributes.length > 0">
                    <div class="col">
                        <p class="label">Attributes</p>
                        <p>{{ attributes }}</p>
                    </div>
                </div>
            </div>
        </section>
    </div>
    <div class="card-group" v-if="hasActions">
        <section class="card" aria-labelledby="actions">
            <div class="card-body">
                <h2 id="actions" class="card-title text-center">Event Handlers</h2>
                <actions :actions="data.data.actions" />
            </div>
        </section>
    </div>
    <div class="card-group">
        <section class="card" aria-labelledby="response">
            <div class="card-body">
                <h2 id="response" class="card-title text-center">Response</h2>
                <table class="table dataTable" aria-labelledby="response">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left" style="width: 20%">DN</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="item of searchResults" :key="item.dn">
                            <td>{{ item.dn }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </section>
    </div>
</template>

<style scoped>
.row {
    padding-bottom: 10px;
}
</style>