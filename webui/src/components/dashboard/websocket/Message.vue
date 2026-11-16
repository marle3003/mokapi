<script setup lang="ts">
import { useRoute } from "@/router";
import SourceView from '../SourceView.vue'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import Actions from '../Actions.vue'
import WebsocketValidationErrors from "./WebsocketValidationErrors.vue";
import MessageSummary from './message/MessageSummary.vue'
import MessageDetails from './message/MessageDetails.vue'
import { useWebsocketMessageEvent } from '@/composables/websocket-message-event'

const route = useRoute();
const { event, data, isLoading, channel, message, value, hasActions, isClientAvailable } = useWebsocketMessageEvent(() => route.params.id as string)
</script>

<template>
  <div v-if="event && data">
    <div class="card-group">
      <section class="card" aria-label="Meta">
        <div class="card-body">
          <message-summary :event="event" :data="data" />
          <message-details :event="event" :data="data" :content-type="message?.contentTypeTitle"
            :client-known="!!data.client && isClientAvailable(data.client.id)" />
        </div>
      </section>
    </div>

    <websocket-validation-errors v-if="data.validationErrors" :validationErrors="data.validationErrors" :channel="channel" :value="data.message.value" />

    <div class="card-group" v-if="hasActions">
      <section class="card" aria-labelledby="actions">
          <div class="card-body">
              <h2 id="actions" class="card-title text-center">Event Handlers</h2>
              <actions :actions="data.actions" />
          </div>
      </section>
    </div>

    <div class="card-group" v-if="value">
      <section class="card" aria-labelledby="value-title">
        <div class="card-body">
          <h2 id="value-title" class="card-title text-center">Value</h2>
          <source-view :source="value.source" :content-type="value.contentType"
            :content-type-title="value.contentTypeTitle" :hide-content-type="value.hideContentType" />
        </div>
      </section>
    </div>
  </div>
  <loading v-if="!event && isLoading"></loading>
  <div v-else-if="!event && !isLoading">
    <Message message="WebSocket Message not found"></Message>
  </div>
</template>
