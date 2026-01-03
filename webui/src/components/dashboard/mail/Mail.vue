<script setup lang="ts">
import { useRoute } from "@/router";
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import MailBody from './MailBody.vue'
import MailFooter from './MailFooter.vue'
import MailAttachments from './MailAttachments.vue'
import { useDashboard } from "@/composables/dashboard";
import { computed, onMounted } from "vue";
import { useMeta } from "@/composables/meta";

const { format } = usePrettyDates()
const route = useRoute()
const { dashboard, getMode } = useDashboard()

const events = computed(() => {
    return dashboard.value.getEvents('mail')
})

const id = computed(() => {
  const id = route.params.id
  if (!id) {
    return undefined
  }

  if (typeof id === 'string') {
    if (isNumber(id)) {
        const index = parseFloat(id);
        const ev = events.value.events.value[index];
        return (<SmtpEventData>ev?.data).messageId ?? null;
    } else {
        return id;
    }
  }
  return null
})

const mailResult = computed(() => id.value ? dashboard.value.getMail(id.value) : undefined)
const mail = computed(() => mailResult.value?.mail.value)
const isLoading = computed(() => mailResult.value?.isLoading.value)

const event = computed(() => {
  if (!id.value) return null
  return events.value.events.value.find(x => {
    const msg = x.data as SmtpEventData
    if (!msg) {
      return false
    }
    return msg.messageId == id.value
  })
})

function isNumber(value: string): boolean {
  return /^[0-9]+$/.test(value);
}
onMounted(() => {
    if (!event.value || getMode() !== 'demo') {
        return
    }
    const id = events.value.events.value.indexOf(event.value)
    useMeta(
        `${mail.value?.data.subject} – Mail Message Details`,
        'View detailed mail event data including sender, recipient, headers, and message content. Inspect and debug email workflows in the Mokapi Dashboard.',
        'https://mokapi.io/dashboard-demo/mail/mail/mails/' + id
    )
})
</script>

<template>
  <div v-if="mail">
    <div class="card-group">
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
          <div class="row mt-2">
            <div class="col">
              <p id="from-label" class="label">From</p>
              <ul class="list-unstyled address-list" aria-labelledby="from-label">
                <li v-for="(addr, index) of mail.data.from">
                  <span v-if="index>0">, </span>
                  <strong v-if="addr.name">{{ addr.name }}</strong>
                  <span v-if="addr.name"> &lt;{{ addr.address }}&gt;</span>
                  <span v-else>{{ addr.address }}</span>
                </li>
              </ul>
            </div>
          </div>
          <div class="row">
            <div class="col">
              <p id="to-label" class="label">To</p>
              <ul class="list-unstyled address-list" aria-labelledby="to-label">
                <li v-for="(addr, index) of mail.data.to">
                  <span v-if="index>0">, </span>
                  <strong v-if="addr.name">{{ addr.name }}</strong>
                  <span v-if="addr.name"> &lt;{{ addr.address }}&gt;</span>
                  <span v-else>{{ addr.address }}</span>
                </li>
              </ul>
            </div>
          </div>
          <div class="row" v-if="mail.data.cc">
            <div class="col">
              <p id="cc-label" class="label">Cc</p>
              <ul class="list-unstyled address-list" aria-labelledby="cc-label">
                <li v-for="(addr, index) of mail.data.cc">
                  <span v-if="index>0">, </span>
                  <strong v-if="addr.name">{{ addr.name }}</strong>
                  <span v-if="addr.name"> &lt;{{ addr.address }}&gt;</span>
                  <span v-else>{{ addr.address }}</span>
                </li>
              </ul>
            </div>
          </div>
          <div class="row" v-if="mail.data.bcc">
            <div class="col">
              <p id="bcc-label" class="label">Bcc</p>
              <ul class="list-unstyled address-list" aria-labelledby="bcc-label">
                <li v-for="(addr, index) of mail.data.bcc">
                  <span v-if="index>0">, </span>
                  <strong v-if="addr.name">{{ addr.name }}</strong>
                  <span v-if="addr.name"> &lt;{{ addr.address }}&gt;</span>
                  <span v-else>{{ addr.address }}</span>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </section>
    </div>
    <mail-body :messageId="mail.data.messageId" :body="mail.data.body" :contentType="mail.data.contentType" />
    <mail-attachments v-if="mail.data.attachments && mail.data.attachments.length > 0" :messageId="mail.data.messageId" :attachments="mail.data.attachments" />
    <mail-footer :contentType="mail.data.contentType" :encoding="mail.data.contentTransferEncoding" :messageId="mail.data.messageId" :inReplyTo="mail.data.inReplyTo" />
  </div>
  <loading v-if="isLoading"></loading>
  <div v-if="!mail && !isLoading">
      <message message="Mail not found"></message>
  </div>
</template>

<style scoped>

.address-list li {
  display: inline;
}
.address-name {
  font-weight:700;
}
.dashboard .card p.subject {
  font-size: 1.4rem;
}
</style>