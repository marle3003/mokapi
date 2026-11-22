<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard'

defineProps<{
    event: ServiceEvent
    data: KafkaMessageData
    // how the message configuration describes the message, unknown when the topic is gone
    keyType?: string | string[]
    contentType?: string
}>()
</script>

<template>
    <div class="row mb-2">
        <div class="col-2">
            <p id="message-key-type" class="label">Key Type</p>
            <p aria-labelledby="message-key-type">{{ keyType ?? '-' }}</p>
        </div>
        <div class="col-2">
            <p id="clientId" class="label">Client</p>
            <p aria-labelledby="clientId">
                <router-link v-if="data.script" :to="{ name: getRouteName('config').value, params: { id: data.script } }">
                    {{ data.clientId }}
                </router-link>
                <router-link v-else-if="data.clientId"
                    :to="{ name: getRouteName('kafkaClient').value, params: { service: event.traits.name, clientId: data.clientId } }">
                    {{ data.clientId }}
                </router-link>
                <span v-else>-</span>
            </p>
        </div>
        <div class="col-2">
            <p id="message-contenttype" class="label">Content Type</p>
            <p aria-labelledby="message-contenttype">{{ contentType ?? '-' }}</p>
        </div>
    </div>
    <div class="row mb-2" v-if="data.producerId > 0">
        <div class="col-2">
            <p id="message-producerId" class="label">Producer Id</p>
            <p aria-labelledby="message-producerId">{{ data.producerId }}</p>
        </div>
        <div class="col-2 mb-2">
            <p id="message-producerEpoch" class="label">Producer Epoch</p>
            <p aria-labelledby="message-producerEpoch">{{ data.producerEpoch }}</p>
        </div>
        <div class="col-2 mb-2">
            <p id="message-sequenceNumber" class="label">Sequence Number</p>
            <p aria-labelledby="message-sequenceNumber">{{ data.sequenceNumber }}</p>
        </div>
    </div>
</template>
