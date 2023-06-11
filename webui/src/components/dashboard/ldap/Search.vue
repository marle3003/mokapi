<script setup lang="ts">
import { useEvents } from '@/composables/events'
import { useRoute } from 'vue-router'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { onUnmounted, computed } from 'vue'
import Actions from '../Actions.vue'
import { usePrettyDates } from '@/composables/usePrettyDate';

const { fetchById } = useEvents()
const eventId = useRoute().params.id as string
const { event, isLoading, close } = fetchById(eventId)
const { duration } = usePrettyDates()

function eventData() {
    return <LdapEventData>event.value?.data
}

function isInitLoading() {
    return isLoading.value && !event.value
}

function attributes() {
  return eventData().request.attributes.join(', ')
}

function scopeToString(scope: SearchScope): string {
    switch (scope) {
        case 3: return 'WholeSubtree'
        case 2: return 'SingleLevel'
        case 1: return 'BaseObject'
        default: return 'invalid'
    }
}

function codeToString(code: LdapResultStatus): string {
    switch (code) {
        case 0: return 'Success'
        case 1: return 'OperationsError'
        case 2: return 'ProtocolError'
        case 3: return 'SizeLimitExceeded'
        case 4: return 'AuthMethodNotSupported'
        case 121: return 'CannotCancel'
        default: return 'invalid'
    }
}

const searchResults = computed(() => {
  return eventData().response.results.sort(compareResult)
})

function compareResult(r1: LdapSearchResult, r2: LdapSearchResult) {
    const name1 = r1.dn.toLowerCase()
    const name2 = r2.dn.toLowerCase()
    return name1.localeCompare(name2)
}

const hasActions = computed(() => {
    return eventData().actions?.length > 0
})
onUnmounted(() => {
    close()
})
</script>

<template>
    <div v-if="event">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                  <div class="row">
                    <div class="col header">
                        <p class="label">Filter</p>
                        <p>{{ eventData().request.filter }}</p>
                    </div>
                  </div>
                  <div class="row">
                    <div class="col">
                        <p class="label">Base DN</p>
                        <p>{{ eventData().request.baseDN }}</p>
                    </div>
                    <div class="col">
                        <p class="label">Scope</p>
                        <p>{{ scopeToString(eventData().request.scope) }}</p>
                    </div>
                    <div class="col">
                        <p class="label">Size Limit</p>
                        <p>{{ eventData().request.sizeLimit }}</p>
                    </div>
                    <div class="col">
                        <p class="label">Time Limit</p>
                        <p>{{ eventData().request.timeLimit }}</p>
                    </div>
                  </div>
                  <div>
                    <p class="label">Attributes</p>
                    <p>{{ attributes() }}</p>
                  </div>
              </div>
            </div>
        </div>
        <div class="card-group">
          <div class="card">
            <div class="card-body">
              <div class="card-title text-center">Response</div>
              <div class="row">
                <div class="col">
                  <p class="label">Status</p>
                  <p>{{ codeToString(eventData().response.status) }}</p>
                </div>
                <div class="col">
                  <p class="label">Duration</p>
                  <p>{{ duration(eventData().duration) }}</p>
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
                    <actions :actions="eventData().actions" />
                </div>
            </div>
        </div>
    </div>
    <loading v-if="isInitLoading()"></loading>
    <div v-if="!event && !isLoading">
        <message message="Search request not found"></message>
    </div>
</template>