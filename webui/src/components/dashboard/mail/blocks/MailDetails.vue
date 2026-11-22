<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard'

// The technical fields of a mail. The full view puts them into its Footer, the panel into its Details tab.
// A mail that is not stored anymore has the Message-ID only.
defineProps<{
    contentType?: string
    encoding?: string
    messageId: string
    inReplyTo?: string
}>()
</script>

<template>
  <div class="row" v-if="contentType || encoding">
    <div class="col" v-if="contentType">
      <p id="content-type" class="label">Content-Type</p>
      <p aria-labelledby="content-type">{{ contentType }}</p>
    </div>
    <div class="col" v-if="encoding">
      <p id="encoding" class="label">Encoding</p>
      <p aria-labelledby="encoding">{{ encoding }}</p>
    </div>
  </div>
  <div class="row mt-2">
    <div class="col">
      <p id="message-id" class="label">Message-ID</p>
      <p aria-labelledby="message-id">{{ messageId }}</p>
    </div>
    <div class="col" v-if="inReplyTo">
      <p id="in-reply-to" class="label">In-Reply-To</p>
      <p aria-labelledby="in-reply-to">
        <router-link :to="{ name: getRouteName('smtpMail').value, params: { id: inReplyTo } }">{{ inReplyTo }}</router-link>
      </p>
    </div>
  </div>
</template>
