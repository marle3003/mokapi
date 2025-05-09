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
      <div class="card">
        <div class="card-body">
          <div class="row">
            <div class="col header">
              <p class="label">Subject</p>
              <p>{{ mail.subject }}</p>
            </div>
            <div class="col-2 align-self-end">
              <p class="label">Date</p>
              <p>
                {{ format(mail.time) }}
              </p>
            </div>
          </div>
          <div class="row">
            <div class="col">
              <p class="label">From</p>
              <span class="prop">
                <ul class="address-list">
                  <li v-for="addr of mail.from">
                    <span class="address-name" v-if="addr.name">{{ addr.name }}</span>
                    <span>{{ addr.address }}</span>
                  </li>
                </ul>
              </span>
            </div>
          </div>
          <div class="row">
            <div class="col">
              <p class="label">To</p>
              <span class="prop">
                <ul class="address-list">
                  <li v-for="addr of mail.to">
                    <span class="address-name" v-if="addr.name">{{ addr.name }}</span>
                    <span>{{ addr.address }}</span>
                  </li>
                </ul>
              </span>
            </div>
          </div>
          <div class="row" v-if="mail.cc">
            <div class="col">
              <p class="label">Cc</p>
              <span class="prop">
                <ul class="address-list">
                  <li v-for="addr of mail.cc">
                    <span class="address-name" v-if="addr.name">{{ addr.name }}</span>
                    <span>{{ addr.address }}</span>
                  </li>
                </ul>
              </span>
            </div>
          </div>
          <div class="row" v-if="mail.bcc">
            <div class="col">
              <p class="label">Bcc</p>
              <span class="prop">
                <ul class="address-list">
                  <li v-for="addr of mail.bcc">
                    <span class="address-name" v-if="addr.name">{{ addr.name }}</span>
                    <span>{{ addr.address }}</span>
                  </li>
                </ul>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
    <mail-body :messageId="mail.messageId" :body="mail.body" :contentType="mail.contentType" />
    <mail-attachments :messageId="mail.messageId" :attachments="mail.attachments" />
    <mail-footer :contentType="mail.contentType" :encoding="mail.contentTransferEncoding" :messageId="mail.messageId" :inReplyTo="mail.inReplyTo" />
  </div>
  <loading v-if="isLoading"></loading>
  <div v-if="!mail && !isLoading">
      <message message="Mail not found"></message>
  </div>
</template>

<style scoped>
.prop span {
  font-size: 0.8rem;
}
.address-list {
  display: inline;
  list-style: none;
  padding: 0px;
}
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
  padding-right: 0.3rem;
}
.dashboard .card p.subject {
  font-size: 1.4rem;
}
</style>