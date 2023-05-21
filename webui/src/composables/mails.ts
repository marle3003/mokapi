import { watchEffect, ref } from 'vue'
import { useFetch } from './fetch'

export function useMails() {
    
    function fetchMail(messageId: string) {
        const response = useFetch(`/api/services/smtp/mails/${messageId}`, undefined, false)
        const mail = ref<Mail | null>(null)
        const isLoading = ref<boolean>()

        watchEffect(() => {
            mail.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return { mail, isLoading }
    }

    function attachmentUrl(messageId: string, name: string): string {
        return `/api/services/smtp/mails/${messageId}/attachments/${name}`
    }

    return { fetchMail, attachmentUrl }
}