<script setup lang="ts">
import { useMails } from '@/composables/mails';
import { usePrettyBytes } from '@/composables/usePrettyBytes';
import type { PropType } from 'vue';

defineProps({
    messageId: { type: String, required: true },
    attachments: { type: Object as PropType<Attachment[]>, required: true },
  }
)

const { format } = usePrettyBytes()
const { attachmentUrl } = useMails()
</script>

<template>
  <div class="row row-cols-1 g-4 attachments" v-if="attachments">
    <div class="col attachment" v-for="attach of attachments">
      <a :href="attachmentUrl(messageId, attach.name)" download>
        <div class="card">
          <div class="card-body">
            <div class="row">
              <div class="col-3">
                <i class="bi bi-paperclip"></i>
              </div>
              <div class="col">
                <p class="name">{{ attach.name }}</p>
                <p>{{ attach.disposition }}</p>
                <p>{{ format(attach.size) }}</p>
              </div>
            </div>
          </div>
        </div>
      </a>
    </div>
  </div>
</template>

<style scoped>
.attachments .card{
    border-color: var(--color-border);
    background-color: var(--color-background-soft);
    margin: 7px;
    margin-left: 0;
    font-weight: 400;
    font-size: 0.7rem;
}
.attachment {
  width: 20%;
}
.attachment i {
  font-size: 1.5rem;
}
.attachments .card-body{
  padding: 0.3rem;
  padding-top: 0.5rem;
}
.attachments .card p{
    font-weight: 400;
    font-size: 0.7rem;
}

.attachments .card p.name{
    font-weight: 400;
    font-size: 0.9rem;
}
</style>