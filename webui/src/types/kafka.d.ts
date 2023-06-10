declare interface KafkaService extends Service {
  topics: KafkaTopic[];
  groups: KafkaGroup[];
  servers: KafkaServer[];
}

declare interface KafkaServer {
  name: string;
  url: string;
  description: string;
}

declare interface KafkaTopic {
  name: string;
  description: string;
  partitions: KafkaPartition[];
  configs: KafkaTopicConfig;
}

declare interface KafkaTopicConfig {
  name: string
  title: string
  summary: string
  description: string
  key: Schema;
  message: Schema;
  header: Schema
  messageType: string;
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
  partitions: KafkaPartition[];
}

declare interface KafkaEventData {
  offset: number;
  key: string;
  message: string;
  partition: number;
}
