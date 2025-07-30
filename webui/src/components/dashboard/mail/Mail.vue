<script setup lang="ts">
import { useRoute } from "@/router";
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { useMails } from '@/composables/mails'
import MailBody from './MailBody.vue'
import MailFooter from './MailFooter.vue'
import MailAttachments from './MailAttachments.vue'

const { fetchMail } = useMails()
const { format } = usePrettyDates()
const route = useRoute()
const messageId = route.params.id as string
const { mail, isLoading: isLoading } = fetchMail(messageId)
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
          <div class="row">
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