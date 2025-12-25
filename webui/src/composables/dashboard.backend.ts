import { reactive, ref, watchEffect } from "vue";
import { transformPath, useFetch } from "./fetch";
import type { Dashboard, ExampleRequest, ExampleResult, MailboxMessagesResult, MailboxResult } from "@/types/dashboard";
import { usePrettyLanguage } from "./usePrettyLanguage";

const { formatLanguage } = usePrettyLanguage();

export const dashboard: Dashboard = {

    getAppInfo() {
        return useFetch('/api/info')
    },

    getServices(type, doRefresh) {
        const response = useFetch('/api/services', {}, doRefresh)
        const services = ref<Service[]>([])

        watchEffect(() => {
            if (!response.data){
                services.value = []
                return
            }
    
            let result = []
            if (type) {
                for (let service of response.data) {
                    if (service.type == type){
                        result.push(service)
                    }
                }
            }else{
                result = response.data
            }
    
            services.value = result.sort(compareService)
        })
        return { services, close: response.close }
    },

    getService(name, type) {
        const response = useFetch(`/api/services/${type}/${name}`)
        const service = ref<Service | null>(null)
        const isLoading = ref<boolean>(true)

        watchEffect(() => {
            service.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {service, isLoading, close: response.close} 
    },

    getEvents(namespace: string, ...labels: Label[]) {
        let url = `/api/events?namespace=${namespace}`
        for (let label of labels) {
            url += `&${label.name}=${label.value}`
        }
        const res = useFetch(url)
        const events = ref<ServiceEvent[]>([])

        watchEffect(() => {
            if (!res.data){
                events.value = []
                return
            }
            events.value = res.data
        })
        return { events, close: res.close }
    },

    getEvent(id: string){
        const event = ref<ServiceEvent | null>(null)
        const isLoading = ref<boolean>(true)
        const response = useFetch(`/api/events/${id}`)
        watchEffect(() =>{
            event.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {event, isLoading, close: response.close}
    },

    getMetrics(query) {
        const response = useFetch('/api/metrics?q=' + query)
        return response
    },

    getExample(request: ExampleRequest) {
        const response = useFetch('/api/schema/example', {
            method: 'POST', 
            body: JSON.stringify({name: request.name, schema: request.schema.schema, format: request.schema.format, contentTypes: request.contentTypes}), 
            headers: {'Content-Type': 'application/json', 'Accept': 'application/json'}},
            false, false)
        const res: ExampleResult =  reactive({
            data: [],
            next: () => response.refresh(),
            error: null
        })

        watchEffect(() => {
            if (response.isLoading) {
                return
            }
            if (response.error) {
                res.error = response.error
                res.data = []
                return
            }

            res.data = response.data
            for (const example of res.data) {
                example.value = atob(example.value)
                example.value = formatLanguage(example.value, example.contentType!)
            }
        })
        return res
    },

    getMailbox(service: string, mailbox: string): MailboxResult {
        const mb = ref<SmtpMailbox | null>(null)
        const isLoading = ref<boolean>(true)
        const response = useFetch(`/api/services/mail/${service}/mailboxes/${mailbox}`, {
                headers: {'Accept': 'application/json'}
            },
            false, false)
        
        watchEffect(() =>{
            mb.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {mailbox: mb, isLoading}
    },

    getMailboxMessages(service: string, mailbox: string): MailboxMessagesResult {
        const messages = ref<MessageInfo[]>([])
        const isLoading = ref<boolean>(true)
        const response = useFetch(`/api/services/mail/${service}/mailboxes/${mailbox}/messages`, {
                headers: {'Accept': 'application/json'}
            })
        
        watchEffect(() =>{
            messages.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return { messages: messages, isLoading, close: response.close }
    },

    getMail(messageId: string) {
        const response = useFetch(`/api/services/mail/messages/${messageId}`, undefined, false)
        const mail = ref<Message | null>(null)
        const isLoading = ref<boolean>(true)

        watchEffect(() => {
            mail.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return { mail, isLoading }
    },
    
    getAttachmentUrl(messageId: string, name: string): string {
        return transformPath(`/api/services/mail/messages/${messageId}/attachments/${name}`)
    }
}

function compareService(s1: Service, s2: Service) {
    const name1 = s1.name.toLowerCase()
    const name2 = s2.name.toLowerCase()
    return name1.localeCompare(name2)
}