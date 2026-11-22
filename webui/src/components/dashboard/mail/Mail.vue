<script setup lang="ts">
import { useRoute } from "@/router";
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import MailBody from './MailBody.vue'
import MailDetails from './blocks/MailDetails.vue'
import MailSummary from './blocks/MailSummary.vue'
import MailAttachments from './MailAttachments.vue'
import { useDashboard } from "@/composables/dashboard";
import { ref, watch } from "vue";

const route = useRoute()
const { dashboard } = useDashboard()

const mail = ref<Message | null>(null);
const isLoading = ref<boolean>(true)

// the id changes with the In-Reply-To link, which stays in this view
watch([() => dashboard.value, () => route.params.id as string],
  ([db, id], _, onCleanup) => {
    const res = db.getMail(id);

    const stop = watch(res.mail, (v) => { mail.value = v; isLoading.value = false }, { immediate: true })

    onCleanup(() => {
        stop()
    });
  },
  { immediate: true }
);
</script>

<template>
  <div v-if="mail">
    <mail-summary :mail="mail" />
    <mail-body :messageId="mail.data.messageId" :body="mail.data.body" :contentType="mail.data.contentType" />
    <mail-attachments v-if="mail.data.attachments && mail.data.attachments.length > 0" :messageId="mail.data.messageId" :attachments="mail.data.attachments" />
    <div class="card-group">
      <section class="card" aria-label="Footer">
        <div class="card-body">
          <mail-details :contentType="mail.data.contentType" :encoding="mail.data.contentTransferEncoding" :messageId="mail.data.messageId" :inReplyTo="mail.data.inReplyTo" />
        </div>
      </section>
    </div>
  </div>
  <loading v-if="isLoading"></loading>
  <div v-if="!mail && !isLoading">
      <message message="Mail not found"></message>
  </div>
</template>

<style scoped>

.address-name {
  font-weight:700;
}
.dashboard .card p.subject {
  font-size: 1.4rem;
}
</style>