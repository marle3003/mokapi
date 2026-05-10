declare interface Service {
    name: string
    description: string
    version: string
    contact: Contact | null
    type: string
    metrics: Record<string, any>
    configs: Config[]
    status?: string
}

declare interface Contact {
    name: string
    url: string
    email: string
}

declare interface Metric {
    name: string
    value: string | number
}

declare interface Label {
    name: string
    value: string
}

declare interface ServiceEvent {
    id: string
    data: HttpEventData | KafkaEventData | MqttEventData | SmtpEventData | LdapEventData | JobExecution | LogMessage
    time: string
    traits: Traits
}

declare interface Traits {[name: string]: string}

declare interface Config {
    id: string
    url: string
    provider: string
    time: string
    refs: ConfigRef[]
    tags?: string[]
}

declare interface ConfigRef {
    id: string
    url: string
    provider: string
    time: string
}

declare interface LogMessage {
    message: string
    level: string
}

declare interface Error {
    message: string
}