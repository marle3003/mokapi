declare module 'mokapi/kafka' {
    function produce(args: ProduceArgs): ProduceResult
}

declare interface ProduceArgs {
    cluster?: string
    topic?: string
    partition?: number
    key?: any
    value?: any
    headers?: {[key: string]: any; }
}

declare interface ProduceResult {
    cluster: string
    topic: string
    partition: number
    offset: number
    key: string
    value: string
}

type KafkaEventHandler = (record: KafkaRecord) => boolean

declare interface KafkaRecord {
    offset: number
    time: number
    key: number[]
    Value: number[]
    Headers: KafkaHeader[]
}

declare interface KafkaHeader {
    Key: string
    Value: number[]
}