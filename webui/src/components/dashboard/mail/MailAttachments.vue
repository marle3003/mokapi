<script setup lang="ts">
import { useDashboard } from '@/composables/dashboard';
import { usePrettyBytes } from '@/composables/usePrettyBytes';
import type { PropType } from 'vue';

defineProps({
    messageId: { type: String, required: true },
    attachments: { type: Object as PropType<Attachment[]>, required: true },
  }
)

const { format } = usePrettyBytes()
const { dashboard } = useDashboard()
</script>

<template>
  <section aria-labelledby="attachments-title">
    <h2 id="attachments-title" class="visually-hidden">Attachments</h2>

    <ul class="list-unstyled row row-cols-auto g-3 attachments">
      <li class="col attachment" v-for="attach of attachments" :key="attach.name">
        <a :href="dashboard.getAttachmentUrl(messageId, attach.name)"
          download
          :aria-label="`${attach.name}, ${format(attach.size)}, disposition: ${attach.disposition}`"
          >
          <div class="card">
            <div class="card-body d-flex">
              <div class="me-3 d-flex align-items-center">
                <span class="bi bi-paperclip fs-4" aria-hidden="true"></span>
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