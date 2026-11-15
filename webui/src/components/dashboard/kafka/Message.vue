<script setup lang="ts">
import { useRoute } from '@/router'
import SourceView from '../SourceView.vue'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import Actions from '../Actions.vue'
import MessageHeaderTable from './MessageHeaderTable.vue'
import MessageSummary from './message/MessageSummary.vue'
import MessageDetails from './message/MessageDetails.vue'
import { useKafkaMessageEvent } from '@/composables/kafka-message-event'

const route = useRoute()
const { event, data, isLoading, message, hasActions } = useKafkaMessageEvent(() => route.params.id as string)
</script>

<template>
  <div v-if="event && data">
    <div class="card-group">
      <section class="card" aria-label="Meta">
        <div class="card-body">
          <message-summary :event="event" :data="data" />
          <message-details :event="event" :data="data" :key-type="message?.keyType" :content-type="message?.contentTypeTitle" />
        </div>
      </section>
    </div>

    <div class="card-group" v-if="hasActions">
      <section class="card" aria-labelledby="actions">
          <div class="card-body">
              <h2 id="actions" class="card-title text-center">Event Handlers</h2>
              <actions :actions="data.actions" />
          </div>
      </section>
    </div>

    <div class="card-group">
      <section class="card" aria-labelledby="value-title">
        <div class="card-body">
          <h2 id="value-title" class="card-title text-center">Value</h2>
          <source-view v-if="message" :source="message.source" :content-type="message.contentType"
            :content-type-title="message.contentTypeTitle" />
        </div>
      </section>
    </div>

    <div class="card-group">
      <section class="card" aria-labelledby="header-title">
        <div class="card-body">
          <h2 id="header-title" class="card-title text-center">Headers</h2>
          <MessageHeaderTable :headers="data.headers" />
        </div>
      </section>
    </div>
  </div>
  <loading v-if="!event && isLoading"></loading>
  <div v-else-if="!event && !isLoading">
      <Message message="Kafka Message not found"></Message>
  </div>
</template>
