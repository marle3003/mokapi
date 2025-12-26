import type { AppInfoResponse } from "@/composables/appInfo"
import type { Response } from "@/composables/fetch"
import type { Ref } from "vue"

export interface Dashboard {
    getAppInfo(): AppInfoResponse

    getServices(type?: string, doRefresh?: boolean): ServicesResult
    getService(name: string, type: string): ServiceResult

    getEvents(namespace: string, ...labels: Label[]): EventsResult
    getEvent(id: string): EventResult

    getExample(request: ExampleRequest): ExampleResult

    getMailbox(service: string, mailbox: string): MailboxResult
    getMailboxMessages(service: string, mailbox: string): MailboxMessagesResult
    getMail(messageId: string): MailResult
    getAttachmentUrl(messageId: string, name: string): string

    getMetrics(query: string): Response

    getConfigs(): ConfigsResult 
    getConfig(id: string): ConfigResult
    getConfigData(id: string): ConfigDataResult
    getConfigDataUrl(id: string): string
}

export interface AppInfo {
    version: string
    activeServices: string[]
    search: Search
}

export interface Search {
    enabled: boolean
}

export interface AppInfoResponse {
    data: AppInfo
    isLoading: boolean
    error: string
    close: () => void
}

export interface ServicesResult {
    services: Ref<Service[]>
    close(): void
}

export interface ServiceResult {
    service: Ref<Service | null>
    isLoading: Ref<boolean>
    close(): void
}

export interface EventsResult {
    events: Ref<ServiceEvent[]>
    close(): void
}

export interface EventResult {
    event: Ref<ServiceEvent | null>
    isLoading: Ref<boolean>
    close(): void
}

export interface ExampleResult {
    data: Example[]
    error: string | null
    next:  () => void
}

export interface Example {
    contentType: string
    value: string
    error?: string
}

export interface ExampleRequest {
    name?: string
    schema: SchemaFormat
    contentTypes?: string[]
}

export interface MailboxResult {
    mailbox: Ref<SmtpMailbox | null>
    isLoading: Ref<boolean>
}

export interface MailboxMessagesResult {
    messages: Ref<MessageInfo[]>
    isLoading: Ref<boolean>
    close(): void
}

export interface MailResult {
    mail: Ref<Message | null>
    isLoading: Ref<boolean>
}

export interface ConfigsResult {
    data: Ref<Config[] | null>
    isLoading: Ref<boolean>
    close: () => void
}

export interface ConfigResult {
    config: Ref<Config | null>
    isLoading: Ref<boolean>
    close: () => void
}

export interface ConfigDataResult {
    data: Ref<string | null>
    isLoading: Ref<boolean>
    filename: Ref<string | undefined>
    close: () => void
}