<script setup lang="ts">
import { useRoute } from 'vue-router'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { useMails } from '@/composables/mails'
import MailBody from './MailBody.vue'
import MailFooter from './MailFooter.vue'
import MailAttachments from './MailAttachments.vue'

const { fetchMail } = useMails()
const { format } = usePrettyDates()
const messageId = useRoute().params.id as string
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
              <p aria-labelledby="subject">{{ mail.subject }}</p>
            </div>
            <div class="col-2 align-self-end">
              <p id="date" class="label">Date</p>
              <p aria-labelledby="date">{{ format(mail.date) }}</p>
            </div>
          </div>
          <div class="row">
            <div class="col">
              <p id="from-label" class="label">From</p>
              <ul class="list-unstyled address-list" aria-labelledby="from-label">
                <li v-for="addr of mail.from">
                  <strong v-if="addr.name">{{ addr.name }}</strong>
                  <span> &lt;{{ addr.address }}&gt;</span>
                </li>
              </ul>
            </div>
          </div>
          <div class="row">
            <div class="col">
              <p id="to-label" class="label">To</p>
              <ul class="list-unstyled address-list" aria-labelledby="to-label">
                <li v-for="addr of mail.to">
                  <strong v-if="addr.name">{{ addr.name }}</strong>
                  <span> &lt;{{ addr.address }}&gt;</span>
                </li>
              </ul>
            </div>
          </div>
          <div class="row" v-if="mail.cc">
            <div class="col">
              <p id="cc-label" class="label">Cc</p>
              <ul class="list-unstyled address-list" aria-labelledby="cc-label">
                <li v-for="addr of mail.cc">
                  <strong v-if="addr.name">{{ addr.name }}</strong>
                  <span> &lt;{{ addr.address }}&gt;</span>
                </li>
              </ul>
            </div>
          </div>
          <div class="row" v-if="mail.bcc">
            <div class="col">
              <p id="bcc-label" class="label">Bcc</p>
              <ul class="list-unstyled address-list" aria-labelledby="bcc-label">
                <li v-for="addr of mail.bcc">
                  <strong v-if="addr.name">{{ addr.name }}</strong>
                  <span> &lt;{{ addr.address }}&gt;</span>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </section>
    </div>
    <mail-body :messageId="mail.messageId" :body="mail.body" :contentType="mail.contentType" />
    <mail-attachments v-if="mail.attachments.length > 0" :messageId="mail.messageId" :attachments="mail.attachments" />
    <mail-footer :contentType="mail.contentType" :encoding="mail.contentTransferEncoding" :messageId="mail.messageId" :inReplyTo="mail.inReplyTo" />
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
.address-list li::after {
  content: ", ";
}
.address-list li:last-child::after {
    content: "";
}
.address-name {
  font-weight:700;
}
.dashboard .card p.subject {
  font-size: 1.4rem;
}
</style>