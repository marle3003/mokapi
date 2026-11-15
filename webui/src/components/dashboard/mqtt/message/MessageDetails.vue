<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard'

defineProps<{
    event: ServiceEvent
    data: MqttMessageData
    // how the message configuration describes the message, unknown when the topic is gone
    contentType?: string
}>()
</script>

<template>
    <div class="row mb-2">
        <div class="col-2">
            <p id="clientId" class="label">Client</p>
            <p aria-labelledby="clientId">
                <router-link v-if="data.script" :to="{ name: getRouteName('config').value, params: { id: data.script } }">
                    {{ data.clientId }}
                </router-link>
                <router-link v-else-if="data.clientId"
                    :to="{ name: getRouteName('mqttClient').value, params: { service: event.traits.name, clientId: data.clientId } }">
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
</template>
