import { reactive, ref, watchEffect } from "vue";
import { transformPath, useFetch } from "./fetch";
import type { AppInfoResponse, Dashboard, ExampleRequest, ExampleResult, MailboxMessagesResult, MailboxResult } from "@/types/dashboard";
import type { Response } from "./fetch";

export function useDemoDashboard() {

    const db = useFetch('/demo/dashboard.json', undefined, false);

    const dashboard: Dashboard = {

        getAppInfo() {
            const response: AppInfoResponse = reactive({
                data: {
                    version: '0.0.0',
                    activeServices: [],
                    search: { enabled: false }
                },
                isLoading: true,
                error: '',
                close: () => {},
            })

            watchEffect(() => {
                if (!db.data) {
                    return
                }
                response.data = db.data.info
                response.isLoading = db.isLoading
            })
            return response
        },

        getServices(type) {
            const services = ref<Service[]>([])

            watchEffect(() => {
                if (!db.data) {
                    return
                }
                let result = []
                if (type) {
                    for (let service of db.data['services']) {
                        if (service.type == type) {
                            result.push(service)
                        }
                    }
                } else {
                    result = db.data['services']
                }
                services.value = result.sort(compareService)
            })
            return { services, close: () => {} } 
        },

        getService(name) {
            const service = ref<Service | null>(null)
            const isLoading = ref<boolean>(true)

            watchEffect(() => {
                if (!db.data) {
                    return
                }
                service.value = db.data['service_' + name]
                isLoading.value = db.isLoading
            })
            return { service, isLoading, close: () => {} }
        },

        getEvents(namespace: string, ...labels: Label[]) {
            const events = ref<ServiceEvent[]>([])
            const isLoading = ref<boolean>(true)

            watchEffect(() => {
                if (!db.data) {
                    return
                }
                
                const result = []
                 for (const event of db.data['events']) {
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

                events.value = result
                isLoading.value = db.isLoading
            })
            return { events, isLoading, close: () => {} }
        },

        getEvent(id: string) {
            const event = ref<ServiceEvent | null>(null)
            const isLoading = ref<boolean>(true)

            watchEffect(() => {
                if (!db.data) {
                    return
                }
                const events = db.data['events'];
                for (const e of events) {
                    if (e.id === id) {
                        event.value = e;
                    }
                }
                isLoading.value = db.isLoading
            })
            return { event, isLoading, close: () => {} }
        },

        getMetrics(query) {
            const response: Response = reactive({
                data: null,
                isLoading: false,
                error: null,
                close: () => {},
                refs: 1,
                refresh: () => { },
            })

            watchEffect(() => {
                if (!db.data) {
                    return
                }
                response.data = db.data['metrics']
                response.isLoading = db.isLoading
            })
            return response
        },

        getExample(request: ExampleRequest) {
            const response: ExampleResult = {
                data: [],
                next: () => {},
                error: 'Example is not working in demo dashboard',
            }
            return response
        },

        getMailbox(service: string, mailbox: string): MailboxResult {
            const mb = ref<SmtpMailbox | null>(null)
            const isLoading = ref<boolean>(true)

            watchEffect(() => {
                if (!db.data) {
                    return
                }
                mb.value =  db.data[`mailbox_${mailbox}`]
                isLoading.value = db.isLoading
            })
            return { mailbox: mb, isLoading }
        },

        getMailboxMessages(service: string, mailbox: string): MailboxMessagesResult {
            const messages = ref<MessageInfo[]>([])
            const isLoading = ref<boolean>(true)

            watchEffect(() => {
                if (!db.data) {
                    return
                }

                const result = [];
                for (const mail of db.data['mails']) {
                    if (mail.data.to.filter((x: any) => x.address === mailbox).length > 0) {
                        result.push({
                            messageId: mail.data.messageId,
                            subject: mail.data.subject,
                            from: mail.data.from,
                            to: mail.data.to,
                            date: mail.data.date
                        })
                    }
                }

                messages.value = result
                isLoading.value = db.isLoading
            })
            return { messages, isLoading, close: () => {} }
        },

        getMail(messageId: string) {
            const mail = ref<Message | null>(null)
            const isLoading = ref<boolean>(true)

            watchEffect(() => {
                if (!db.data) {
                    return
                }
                mail.value = db.data['mails'].find((x: any) => x.data.messageId === messageId);
                isLoading.value = db.isLoading
            })
            return { mail, isLoading, close: () => {} }
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
        },

        getConfigs() {
            const configs = ref<Config[]>([])
            const isLoading = ref<boolean>(true)

            watchEffect(() => {
                if (!db.data) {
                    return
                }
                configs.value = db.data.configs
                isLoading.value = db.isLoading
            })
            return { data: configs, isLoading, close: () => {} }
        },

        getConfig(id: string) {
            const config = ref<Config | null>(null)
            const isLoading = ref<boolean>(true)

            watchEffect(() => {
                if (!db.data) {
                    return
                }
                config.value =  db.data.configs.find((x: Config) => x.id === id);
                isLoading.value = db.isLoading
            })
            return { config, isLoading, close: () => {} }
        },

        getConfigData(id) {
            const response = useFetch(this.getConfigDataUrl(id));
            const data = ref<string | null>(null);
            const isLoading = ref<boolean>(true);
            const filename = ref<string | undefined>();

            watchEffect(() => {
                let config = null
                if (db.data) {
                    config = db.data.configs.find((x: Config) => x.id === id);
                }

                data.value = response.data ? response.data : null;
                isLoading.value = response.isLoading;
                filename.value = getFilenameFromUrl(config?.url);
            })

            return { data, isLoading, filename: filename, close: () => { } }
        },

        getConfigDataUrl(id) {
            let config = null;
            if (db.data) {
                config = db.data.configs.find((x: Config) => x.id === id);
            }
            return '/demo/' + getFilenameFromUrl(config?.url)
        },
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

    function getFilenameFromUrl(url: string): string {
        return new URL(url).pathname.split('/').pop()!;
    }

    return dashboard
}