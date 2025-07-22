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
            <p class="label">Content-Type</p>
            <p>{{ contentType }}</p>
          </div>
          <div class="col" v-if="encoding">
            <p class="label">Encoding</p>
            <p>{{ encoding }}</p>
          </div>
        </div>
        <div class="row">
          <div class="col">
            <p class="label">Message-ID</p>
            <p>{{ messageId }}</p>
          </div>
          <div class="col" v-if="inReplyTo">
            <p class="label">In-Reply-To</p>
            <p>
              <router-link :to="{name: 'smtpMail', params: { id: messageId } }">{{ inReplyTo }}</router-link>
            </p>
          </div>
          </div>
      </div>
    </div>
  </div>
</template>