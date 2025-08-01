declare interface Service {
    name: string
    description: string
    version: string
    contact: Contact | null
    type: string
    metrics: Metric[]
    configs: Config[]
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
    data: HttpEventData | KafkaEventData | SmtpEventData | LdapEventData | JobExecution
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
}

declare interface ConfigRef {
    id: string
    url: string
    provider: string
    time: string
}