declare interface KafkaService extends Service {
  topics: KafkaTopic[];
  groups: KafkaGroup[];
  servers: KafkaServer[];
}

declare interface KafkaServer {
  name: string;
  host: string;
  description: string;
  tags: KafkaServerTag[]
}

declare interface KafkaServerTag {
  name: string
  description: string
}

declare interface KafkaTopic {
  name: string;
  description: string;
  partitions: KafkaPartition[];
  messages: { [messageId: string]: KafkaMessage }
}

declare interface KafkaMessage {
  name: string
  title: string
  summary: string
  description: string
  key: SchemaFormat;
  payload: SchemaFormat;
  header: SchemaFormat
  contentType: string;
}

declare interface KafkaPartition {
  id: number;
  startOffset: number;
  offset: number;
  leader: KafkaBroker;
  segments: number;
}

declare interface KafkaBroker {
  name: string;
  addr: string;
}

declare interface KafkaGroup {
  name: string;
  members: KafkaMember[];
  coordinator: string;
  leader: string;
  state: string;
  protocol: string;
  topics: string[] | null;
}

declare interface KafkaMember {
  name: string;
  addr: string;
  clientSoftwareName: string;
  clientSoftwareVersion: string;
  heartbeat: number;
  partitions: { [topicName: string]: number[] };
}

declare interface KafkaEventData {
  offset: number;
  key: KafkaValue;
  message: KafkaValue;
  schemaId: number;
  messageId: string
  partition: number;
  headers: KafkaHeader
  deleted: boolean
}

declare interface KafkaHeader { [name: string]: KafkaHeaderValue }

declare interface KafkaHeaderValue {
  value: string;
  binary: string;
}

declare interface KafkaValue {
  value?: string
  binary?: string
}
