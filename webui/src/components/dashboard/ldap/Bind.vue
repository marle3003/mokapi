<script setup lang="ts">
import { computed } from 'vue'
import Actions from '../Actions.vue'

const props = defineProps<{
   event: ServiceEvent
}>()

const data = computed((): {data: LdapEventData, request: LdapBindRequest, response: LdapResponse} => {
    const data = <LdapEventData>props.event.data
    return { data: data, request: <LdapBindRequest>data.request, response: <LdapResponse>data.response }
})

const hasActions = computed(() => {
    return data.value.data.actions?.length > 0
})
</script>

<template>
    <div v-if="event">
        <div class="card-group">
            <section class="card" aria-labelledby="request">
            <div class="card-body">
                <h2 id="request" class="card-title text-center">Request</h2>
                  
                    <div class="row">
                      <div class="col-2">
                          <p class="label">Auth</p>
                          <p>{{ data.request.auth }}</p>
                      </div>
                      <div class="col-2">
                          <p class="label">Password</p>
                          <p>{{ data.request.password }}</p>
                      </div>
                  </div>
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