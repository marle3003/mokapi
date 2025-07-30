import { watchEffect, ref } from 'vue'
import { useFetch, transformPath } from './fetch'

export function useMails() {
    
    function fetchMail(messageId: string) {
        const response = useFetch(`/api/services/mail/messages/${messageId}`, undefined, false)
        const mail = ref<Message | null>(null)
        const isLoading = ref<boolean>()

        watchEffect(() => {
            mail.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return { mail, isLoading }
    }

    function attachmentUrl(messageId: string, name: string): string {
        return transformPath(`/api/services/mail/messages/${messageId}/attachments/${name}`)
    }

    return { fetchMail, attachmentUrl }
}