import { computed, reactive, ref, watchEffect } from "vue";
import { transformPath, useFetch } from "./fetch";
import type { Dashboard, ExampleRequest, ExampleResult, MailboxMessagesResult, MailboxResult } from "@/types/dashboard";
import { usePrettyLanguage } from "./usePrettyLanguage";

const { formatLanguage } = usePrettyLanguage();

const demo = ref<any>(null);

const data = computed(async () => {
    let res = await fetch(transformPath('/demo/dashboard.json'));
    if (!res.ok) {
        let text = await res.text()
        throw new Error(res.statusText + ': ' + text)
    }
    return await res.json()
})

watchEffect(async () => {
    demo.value = await data.value;
})

export const dashboard: Dashboard = {

    getServices(type) {
        let result = []
        if (demo.value) {
            const response = demo.value['services']

            if (response) {
                if (type) {
                    for (let service of response) {
                        if (service.type == type){
                            result.push(service)
                        }
                    }
                }else{
                    result = response
                }
            }
        }

        const services = ref<Service[]>([])
        services.value = result.sort(compareService)
        return { services, close: () => {}}
    },

    getService(name, type) {
        let result = null;
        if (demo.value) {
            result = demo.value['service_'+name]
        }

        const service = ref<Service | null>(result)
        const isLoading = ref<boolean>(false)
        return {service, isLoading, close: () => {}} 
    },

    getEvents(namespace: string, ...labels: Label[]) {
        let events = null;
        if (demo.value) {
            events = demo.value['events']
        }

        const result = []
        for (const event of events) {
            if (event.traits.namespace !== namespace) {
                continue
            }
            let isValid = true
            for (const label of labels) {
                if (event.traits[label.name] !== label.value) {
                    isValid = false
                }
            }
            if (isValid) {
                result.push(event)
            }
        }
        
        return { events: ref<ServiceEvent[]>(result), close: () => {} }
    },

    getEvent(id: string){
        let event = null;
        if (demo.value) {
            const events = demo.value['events'];
            for (const e of events) {
                if (e.id === id) {
                    event = e;
                }
            }
        }

        return {event: ref<ServiceEvent | null>(event), isLoading: ref(false), close: () => {}}
    },

    getMetrics(query) {
        let metrics = []
        if (demo.value) {
            metrics = demo.value['metrics']
        }
        return {
            data: metrics,
            isLoading: false,
            error: null,
            close: () => {},
            refs: 1,
            refresh: () => {},
        }
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
        let mb = null;
        if (demo.value) {
            mb = demo.value[`mailbox_${mailbox}`];
        }
        return {mailbox: ref<SmtpMailbox | null>(mb), isLoading: ref(false)}
    },

    getMailboxMessages(service: string, mailbox: string): MailboxMessagesResult {
        const result = [];
        if (demo.value) {
            for (const mail of demo.value['mails']) {
                if (mail.data.to.filter((x: any) =>  x.address === mailbox).length > 0) {
                    result.push({
                        messageId: mail.data.messageId,
                        subject: mail.data.subject,
                        from: mail.data.from,
                        to: mail.data.to,
                        date: mail.data.date
                    })
                }
            }
        }

        const messages = ref<MessageInfo[]>(result)
        const isLoading = ref<boolean>(true)
        return { messages: messages, isLoading, close: () => {} }
    },

    getMail(messageId: string) {
        let mail = null;
        if (demo.value) {
            mail = demo.value['mails'].find((x: any) => x.data.messageId === messageId);
        }

        return { mail: ref<Message | null>(mail), isLoading: ref<boolean>(false) }
    },

    getAttachmentUrl(messageId: string, name: string): string {
        const result = this.getMail(messageId)
        if (!result.mail.value) {
            return '';
        }
        const mail = result.mail.value;
        const attachment = mail.data.attachments.find((x: any) => x.name === name);
        if (!attachment) {
            return '';
        }
        return transformPath(`/demo/${getFilename(attachment.contentType)}`)
    }
}

function compareService(s1: Service, s2: Service) {
    const name1 = s1.name.toLowerCase()
    const name2 = s2.name.toLowerCase()
    return name1.localeCompare(name2)
}

function getFilename(contentType: string) {
    const match = contentType.match(/name=([^;]+)/);
    if (match && match[1]) {
        return match[1];
    }
    return null;
}