<script setup lang="ts">
import { computed, onMounted, ref, useTemplateRef } from 'vue'
import Actions from '../Actions.vue'
import { Modal } from 'bootstrap';

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

const selected = ref<{dn: string, keys: string[], attributes: { [name: string]: string[] }} | undefined>()
const dialogRef = useTemplateRef('dialogRef')
let dialog: Modal;

function showResult(result: LdapSearchResult) {
    if (getSelection()?.toString()) {
        return
    }
    selected.value = {
        dn: result.dn,
        keys: Object.keys(result.attributes).sort((x: string, y: string) => x.localeCompare(y)),
        attributes: result.attributes
    }
    dialog.show()
}

onMounted(() => {
    if (dialogRef.value) {
        dialog = new Modal(dialogRef.value)
    }
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
                <table class="table dataTable selectable" aria-labelledby="response">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left col">DN</th>
                            <th scope="col" class="text-center col-2">Attributes</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="item of searchResults" :key="item.dn" @click="showResult(item)">
                            <td>{{ item.dn }}</td>
                            <td class="text-center">{{ Object.keys(item.attributes).length }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </section>
    </div>
    <div class="modal fade" id="modal-response" tabindex="-1" aria-hidden="true" aria-labelledby="modal-response-title" ref="dialogRef">
        <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
            <div class="modal-content" v-if="selected">
                <div class="modal-header">
                    <h6 id="modal-response-title" class="modal-title">DN: {{ selected.dn }}</h6>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <div class="row p-2">
                        <table class="table dataTable" aria-labelledby="modal-response-title">
                            <thead>
                                <tr>
                                    <th scope="col" class="text-left col-4">Type</th>
                                    <th scope="col" class="text-left col">Value</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr v-for="name in selected.keys" :key="name">
                                    <td>{{ name }}</td>
                                    <td v-html="selected.attributes[name]?.join('<br />')"></td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
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