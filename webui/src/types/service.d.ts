declare interface Service {
    name: string
    description: string
    version: string
    contact: Contact | null
    type: ServiceType
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
    data: HttpEventData | KafkaEventData | SmtpEventData | LdapEventData
    time: string
    traits: Traits
}

declare interface Traits {[name: string]: string}

declare interface Config {
    id: string
    url: string
    provider: string
    time: string
}