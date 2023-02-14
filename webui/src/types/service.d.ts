declare interface Service {
    name: string
    description: string
    version: string
    contact: Contact | null
    type: ServiceType
    metrics: Metric[]
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
    data: HttpEventData | KafkaEventData
    time: string
}