<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate'
import { useRoute } from '@/router'
import { orDash, recipients } from '@/composables/mail'
import MailAddresses from './MailAddresses.vue'
import MailStatusBadge from './MailStatusBadge.vue'

// The head of a mail. The full view shows the mail it loaded (`mail`), the detail panel the event
// (`event`, `data`), which also tells a rejected mail and one that is not stored anymore.
defineProps<{
    mail?: Message
    event?: ServiceEvent
    data?: SmtpEventData
}>()

const { format } = usePrettyDates()
const route = useRoute()
</script>

<template>
  <div class="card-group" v-if="mail">
    <section class="card" aria-label="Info">
      <div class="card-body">
        <div class="row">
          <div class="col header">
            <p id="subject" class="label">Subject</p>
            <p aria-labelledby="subject">{{ mail.data.subject }}</p>
          </div>
          <div class="col-2">
            <p id="service" class="label">Service</p>
            <p aria-labelledby="service">
              <router-link :to="route.service(mail.service, 'mail')">
                {{ mail.service }}
              </router-link>
            </p>
          </div>
          <div class="col-2">
            <p id="date" class="label">Date</p>
            <p aria-labelledby="date">{{ format(mail.data.date) }}</p>
          </div>
        </div>
        <mail-addresses class="mt-2" label="From" :addresses="mail.data.from" />
        <mail-addresses label="To" :addresses="mail.data.to" />
        <mail-addresses label="Cc" :addresses="mail.data.cc" />
        <mail-addresses label="Bcc" :addresses="mail.data.bcc" />
      </div>
    </section>
  </div>
  <template v-else-if="event && data">
    <div class="row mb-2">
      <div class="col-10 header">
        <p class="label">Subject</p>
        <p>{{ orDash(data.subject) }}</p>
      </div>
      <div class="col">
        <p class="label">Time</p>
        <p>{{ format(event.time) }}</p>
      </div>
    </div>
    <div class="row mb-2">
      <div class="col-3">
        <p class="label">Status</p>
        <p><mail-status-badge :data="data" /></p>
      </div>
      <div class="col">
        <p class="label">From</p>
        <p>{{ orDash(data.from) }}</p>
      </div>
      <div class="col">
        <p class="label">To</p>
        <p>{{ recipients(data.to) }}</p>
      </div>
    </div>
    <div class="row mb-2" v-if="data.error">
      <div class="col">
        <p class="label">Error</p>
        <p>{{ data.error }}</p>
      </div>
    </div>
  </template>
</template>
