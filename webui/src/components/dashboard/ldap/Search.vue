<script setup lang="ts">
import { computed } from 'vue'
import Actions from '../Actions.vue'
import { usePrettyDates } from '@/composables/usePrettyDate';

const props = defineProps<{
   event: ServiceEvent
}>()

const { format, duration } = usePrettyDates()

const data = computed((): {data: LdapEventData, request: LdapSearchRequest, response: LdapSearchResponse} => {
    const data = <LdapEventData>props.event.data
    return { data: data, request: <LdapSearchRequest>data.request, response: <LdapSearchResponse>data.response }
})

function attributes() {
    if (!data.value.request.attributes) {
    return ''
    }
    return data.value.request.attributes.join(', ')
}

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
        <div class="card">
            <div class="card-body">
                <div class="row">
                    <div class="col header">
                        <p class="label">Filter</p>
                        <p>{{ data.request.filter }}</p>
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
                    <div class="col">
                        <p class="label">Base DN</p>
                        <p>{{ data.request.baseDN }}</p>
                    </div>
                </div>
                <div class="row">
                    <div class="col">
                        <p class="label">Attributes</p>
                        <p>{{ attributes() }}</p>
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
                <div class="card-title text-center">Results</div>
                <table class="table dataTable">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left" style="width: 20%">DN</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="item of searchResults" :key="item.dn">
                            <td>
                                {{ item.dn }}
                            </td>
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
</template>

<style scoped>
.row {
    padding-bottom: 10px;
}
</style>