import { computed, shallowRef, watch, type MaybeRefOrGetter } from 'vue'
import { useDashboard } from '@/composables/dashboard'
import { useRequestEvent } from '@/composables/request-event'
import { mailStatus } from '@/composables/mail'

// A recorded mail event by id and, for an accepted mail, the mail it names. The mail is gone when it is
// not (or no longer) in a mailbox store; the event still tells what happened. Used by the detail panel
// of the mail log.
export function useMailEvent(id: MaybeRefOrGetter<string>) {
    const { dashboard } = useDashboard()
    const { event, data, isLoading: eventLoading } = useRequestEvent<SmtpEventData>(id)

    const rejected = computed(() => !!data.value && mailStatus(data.value) === 'Rejected')
    const messageId = computed(() => rejected.value ? undefined : data.value?.messageId)

    const mailRes = shallowRef<ReturnType<typeof dashboard.value.getMail>>()
    watch([dashboard, messageId], ([db, mailId]) => {
        // getMail isn't reactive to the id, so every id gets its own request
        mailRes.value = mailId ? db.getMail(mailId) : undefined
    }, { immediate: true })

    const mail = computed(() => mailRes.value?.mail.value ?? null)
    const mailLoading = computed(() => mailRes.value?.isLoading.value ?? false)
    const mailGone = computed(() => !!data.value && !rejected.value && !mail.value && !mailLoading.value)
    const isLoading = computed(() => eventLoading.value || mailLoading.value)

    return { event, data, mail, rejected, mailGone, isLoading }
}
