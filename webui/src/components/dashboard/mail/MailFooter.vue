<script setup lang="ts">
defineProps({
    contentType: { type: String, required: true },
    encoding: { type: String, required: false },
    messageId: { type: String, required: true },
    inReplyTo: { type: String, required: false },
  }
)
</script>

<template>
  <div class="card-group">
    <div class="card">
      <div class="card-body">
        <div class="row">
          <div class="col">
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
              <router-link :to="{name: 'smtpMail', params: { id: messageId } }">{{ inReplyTo }}</router-link>
            </p>
          </div>
          </div>
      </div>
    </div>
  </div>
</template>