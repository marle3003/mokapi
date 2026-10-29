import { reactive, ref, watchEffect } from "vue";
import { transformPath, useFetch } from "./fetch";
import type { Dashboard, ExampleRequest, ExampleResult, MailboxMessagesResult, MailboxResult, EventQuery } from "@/types/dashboard";
import { usePrettyLanguage } from "./usePrettyLanguage";

const { formatLanguage } = usePrettyLanguage();

export const dashboard: Dashboard = {

    getAppInfo() {
        return useFetch('/api/info')
    },

    getServices(type, doRefresh) {
        const path = type ? `/api/services?type=${type}` : '/api/services'
        const response = useFetch(path, {}, doRefresh)
        const services = ref<Service[]>([])

        watchEffect(() => {
            if (!response.data){
                services.value = []
                return
            }
            services.value = response.data .sort(compareService)
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

    getHttpOperations(serviceName: string, path?: string, method?: string) {
        let url = `/api/services/http/${serviceName}/operations`
        if (method) {
            url += `?method=${method}`
        }
        if (path) {
            const separator = method ? '&' : '?'
            url += `${separator}path=${path}`
        }
        const response = useFetch(url)
        const operations = ref<HttpOperation[] | null>(null)
        const isLoading = ref<boolean>(true)

        watchEffect(() => {
            operations.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {operations, isLoading, close: response.close} 
    },

    getKafkaChannel(serviceName: string, channelName: string) {
        const response = useFetch(`/api/services/kafka/${serviceName}/channels/${channelName}`)
        const channel = ref<KafkaChannel | null>(null)
        const isLoading = ref<boolean>(true)

        watchEffect(() => {
            channel.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {channel, isLoading, close: response.close} 
    },

    getKafkaTopic(serviceName: string, topicName: string) {
        const response = useFetch(`/api/services/kafka/${serviceName}/topics/${topicName}`)
        const topic = ref<KafkaTopic | null>(null)
        const isLoading = ref<boolean>(true)

        watchEffect(() => {
            topic.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {topic, isLoading, close: response.close} 
    },

    getKafkaGroup(serviceName: string, groupName: string) {
        const response = useFetch(`/api/services/kafka/${serviceName}/groups/${groupName}`)
        const group = ref<KafkaGroup | null>(null)
        const isLoading = ref<boolean>(true)

        watchEffect(() => {
            group.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {group, isLoading, close: response.close} 
    },

    getKafkaClient(serviceName: string, clientId: string) {
        const response = useFetch(`/api/services/kafka/${serviceName}/clients/${clientId}`)
        const client = ref<KafkaClient | null>(null)
        const isLoading = ref<boolean>(true)

        watchEffect(() => {
            client.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {client, isLoading, close: response.close} 
    },

    getEvents(query?: EventQuery) {
        let url = `/api/events`

        if (query) {
            const params = new URLSearchParams()
            if (query.labels) {
                for (const label of query.labels) {
                    params.append(`traits[${label.name}]`, label.value)
                }
            }
            if (query.http) {
                if (query.http.methods) {
                    params.append('http[method]', query.http.methods.join(','))
                }
                if (query.http.url) {
                    params.append('http[url]', query.http.url)
                }
                for (const h of query.http.requestHeaders ?? []) {
                    params.append('http[requestHeader]', `${h.name}${h.value ? ':'+h.value : ''}`)
                }
                if (query.http.statusCode) {
                    params.append('http[statusCode]', query.http.statusCode)
                }
                for (const h of query.http.responseHeaders ?? []) {
                    params.append('http[responseHeader]', `${h.name}${h.value ? ':'+h.value : ''}`)
                }
                if (query.http.clientIP) {
                    params.append('http[clientIP]', query.http.clientIP)
                }
            }
            if (query.page !== undefined) {
                params.append('page', query.page.toString())
            }
            if (query.size !== undefined) {
                params.append('size', query.size.toString())
            }

            url += `?${params.toString()}`
        }
        

       
        const res = useFetch(url)
        const data = ref<ServiceEventPage>({ items: [], page: 1, size: 25, total: 0 })

        watchEffect(() => {
            if (!res.data){
                data.value.items = []
                data.value.total = 0
                return
            }
            data.value = res.data
        })
        return { data, close: res.close }
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

    getMetrics(type?: string) {
        const path = type ? `/api/metrics/${type}` : '/api/metrics'
        const response = useFetch(path)
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
    },

    getConfigs() {
        const response = useFetch('/api/configs')
        const configs = ref<Config[] | null>(null)
        const isLoading = ref<boolean>(true)

        watchEffect(() => {
            configs.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {data: configs, isLoading, close: response.close}
    },

    getConfig(id: string) {
        const response = useFetch(`/api/configs/${id}`)
        const config = ref<Config | null>(null)
        const isLoading = ref<boolean>(true)

        watchEffect(() => {
            config.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {config, isLoading, close: response.close}
    },

    getConfigData(id) {
        const response = useFetch(this.getConfigDataUrl(id)); 
        const data = ref<string | null>(null);
        const isLoading = ref<boolean>(true);
        const filename = ref<string | undefined>();

        watchEffect(() => {
            data.value = response.data ? response.data : null
            isLoading.value = response.isLoading
            if (response.header) {
                const h = response.header.get('Content-Disposition')!
                const matches = h.match(/filename="([^\"]*)"/)
                if (matches && matches.length > 1) {
                    filename.value = matches[1]
                }
            }
        })

        return { data, isLoading, filename, close: response.close }
    },

    getConfigDataUrl(id: string) {
        return `/api/configs/${id}/data`
    },

    getBrunoCollectionUrl(serviceName: string, params?: URLSearchParams) {
        const url = ref<string | null>(null)
        const base = `/api/services/http/${encodeURIComponent(serviceName)}`
        let fullUrl = `${base}/bruno.yaml`
        if (params) {
            const query = params.toString()
            if (query) {
                fullUrl += `?${query}`
            }
        }
        url.value = transformPath(fullUrl)
        return url
    }
}

function compareService(s1: Service, s2: Service) {
    const name1 = s1.name.toLowerCase()
    const name2 = s2.name.toLowerCase()
    return name1.localeCompare(name2)
}