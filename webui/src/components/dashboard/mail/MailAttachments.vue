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
  <section aria-label="Attachments">
    <ul class="list-unstyled row row-cols-auto g-3 attachments">
      <li class="col attachment" v-for="attach of attachments" :key="attach.name" :aria-label="attach.name">
        <a :href="attachmentUrl(messageId, attach.name)" download>
          <div class="card">
            <div class="card-body d-flex">
              <div class="me-3 d-flex align-items-center">
                <i class="bi bi-paperclip fs-4"></i>
              </div>
              <div>
                <p class="fw-semibold" aria-label="Name">{{ attach.name }}</p>
                <p class="small" aria-label="Disposition">{{ attach.disposition }}</p>
                <p class="small" aria-label="Size">{{ format(attach.size) }}</p>
              </div>
             
            </div>
          </div>
        </a>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.attachments .card{
    border-color: var(--color-border);
    background-color: var(--color-background-soft);
    margin: 7px;
    margin-left: 0;
}
.attachment {
 min-width: 200px;
}
.attachments .card-body{
  padding: 0.3rem;
  padding-top: 0.5rem;
}
.attachments .card-body .small {
  margin: 0;
  font-size: 0.875rem;
}
</style>