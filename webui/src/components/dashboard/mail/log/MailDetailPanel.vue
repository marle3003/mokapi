<script setup lang="ts">
import { computed, ref } from 'vue'
import LogDetailPanel from '../../log/LogDetailPanel.vue'
import LogTabs from '../../log/LogTabs.vue'
import Actions from '../../Actions.vue'
import MailSummary from '../blocks/MailSummary.vue'
import MailAddresses from '../blocks/MailAddresses.vue'
import MailDetails from '../blocks/MailDetails.vue'
import MailBody from '../MailBody.vue'
import MailAttachments from '../MailAttachments.vue'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { getRouteName } from '@/composables/dashboard'
import { activeMailTab, mailTabs, type MailTab } from '@/composables/mail'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { useMailEvent } from '@/composables/use-mail-event'

const props = defineProps<{ eventId: string }>()
defineEmits<{ close: [] }>()

const { duration } = usePrettyDates()
const { event, data, mail, rejected, mailGone, isLoading } = useMailEvent(() => props.eventId)
const hasActions = computed(() => (data.value?.actions?.length ?? 0) > 0)

// a rejected mail has no mail to show, so no full view either
const fullView = computed(() => data.value && !rejected.value
    ? { name: getRouteName('smtpMail').value, params: { id: data.value.messageId } }
    : undefined)

const tabs = computed(() => data.value ? mailTabs(data.value, mail.value) : [])
// The sub tab belongs to the panel, not to the mail: the next mail opens on the same tab if it has it.
const selectedTab = ref<MailTab>('Body')
const activeTab = computed({
    get: () => activeMailTab(tabs.value, selectedTab.value),
    set: (tab: MailTab) => { selectedTab.value = tab }
})
</script>

<template>
    <log-detail-panel label="Mail details" :full-view="fullView" @close="$emit('close')">
        <template v-if="event && data">
            <div class="card">
                <div class="card-body">
                    <mail-summary :event="event" :data="data" />
                </div>
            </div>

            <log-tabs v-model="activeTab" :tabs="tabs" />
            <div class="log-detail-content" role="tabpanel">
                <template v-if="activeTab !== 'Details'">
                    <loading v-if="isLoading" />
                    <message v-else-if="mailGone" message="Mail content is no longer available" />
                    <template v-else-if="mail">
                        <mail-body v-if="activeTab === 'Body'" :message-id="mail.data.messageId" :body="mail.data.body"
                            :content-type="mail.data.contentType" />
                        <mail-attachments v-else :message-id="mail.data.messageId" :attachments="mail.data.attachments" />
                    </template>
                </template>
                <div v-else class="card">
                    <div class="card-body">
                        <div class="row mb-2">
                            <div class="col">
                                <p class="label">Service</p>
                                <p>{{ event.traits.name }}</p>
                            </div>
                            <div class="col">
                                <p class="label">Duration</p>
                                <p>{{ duration(data.duration) }}</p>
                            </div>
                        </div>
                        <template v-if="mail">
                            <mail-addresses label="Cc" :addresses="mail.data.cc" />
                            <mail-addresses label="Bcc" :addresses="mail.data.bcc" />
                            <mail-details :content-type="mail.data.contentType" :encoding="mail.data.contentTransferEncoding"
                                :message-id="mail.data.messageId" :in-reply-to="mail.data.inReplyTo" />
                        </template>
                        <mail-details v-else-if="mailGone" :message-id="data.messageId" />
                        <template v-if="hasActions">
                            <p class="label mt-3">Event Handlers</p>
                            <actions :actions="data.actions" />
                        </template>
                    </div>
                </div>
            </div>
        </template>
        <loading v-if="!event && isLoading" />
        <message v-if="!event && !isLoading" message="Mail not found" />
    </log-detail-panel>
</template>
