declare interface KafkaService extends Service {
  topics: KafkaTopic[];
  groups: KafkaGroup[];
  servers: KafkaServer[];
  clients: KafkaClient[];
}

declare interface KafkaServer {
  name: string;
  host: string;
  protocol: string
  title: string
  summary: string
  description: string;
  tags: KafkaTag[]
}

declare interface KafkaTag {
  name: string
  description: string
}

declare interface KafkaTopic {
  name: string;
  description: string;
  partitions: KafkaPartition[];
  messages: { [messageId: string]: KafkaMessage }
  tags: KafkaTag[]
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
  segments: number;
}

declare interface KafkaBroker {
  name: string;
  addr: string;
}

declare interface KafkaGroup {
  name: string;
  generation: number
  members: KafkaMember[];
  leader: string;
  state: string;
  protocol: string;
  topics: string[] | null;
}

declare interface KafkaMember {
  name: string;
  clientId: string
  addr: string;
  clientSoftwareName: string;
  clientSoftwareVersion: string;
  heartbeat: number;
  partitions: { [topicName: string]: number[] };
}

declare type KafkaEventData = KafkaMessageData | KafkaRequestLog

declare interface KafkaMessageData {
  offset: number;
  key: KafkaValue;
  message: KafkaValue;
  schemaId: number;
  messageId: string
  partition: number;
  headers: KafkaHeader
  deleted: boolean
  producerId: number
  producerEpoch: number
  sequenceNumber: number
  clientId: string
  script: string
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

declare interface KafkaClient {
  clientId: string
  address: string
  brokerAddress: string
  clientSoftwareName: string;
  clientSoftwareVersion: string;
  groups: {
    memberId: string
    group: string
  }[]
}

declare interface KafkaRequestLog {
  header: KafkaRequestHeader
  request:  KafkaJoinGroupRequest | KafkaSyncGroupRequest | KafkaFindCoordinatorRequest | KafkaInitProducerIdRequest
  response: KafkaJoinGroupResponse | KafkaSyncGroupResponse | KafkaFindCoordinatorResponse | KafkaInitProducerIdResponse
}

declare interface KafkaResponseError {
  errorCode: string
  errorMessage: string
}

declare interface KafkaRequestHeader {
  requestKey: number
  requestName: string
  version: number
}

declare interface KafkaJoinGroupRequest {
  groupName: string
  memberId: string
  protocolType: string
  protocols: string[]
}

declare interface KafkaJoinGroupResponse extends KafkaResponseError {
  generationId: number
  protocolName: string
  memberId: string
  leaderId: string
  members: string[] | undefined
}

declare interface KafkaSyncGroupRequest {
  groupName: string
  generationId: number
  memberId: string
  protocolType: string
  protocolName: string
  groupAssignments: { [name: string]: KafkaGroupAssignment }
}

declare interface KafkaSyncGroupResponse extends KafkaResponseError {
  protocolType: string
  protocolName: string
  assignment: KafkaGroupAssignment
}

declare interface KafkaGroupAssignment {
  version: number
  // topic: partition index
  topics: { [name: string]: int[] }
}

declare interface KafkaListOffsetsRequest {
  topics: { [name: string]: KafkaListOffsetsRequestPartition[] }
}

declare interface KafkaListOffsetsRequestPartition {
  partition: number
  timestamp: number
}

declare interface KafkaListOffsetsResponse {
  topics: { [name: string]: KafkaListOffsetsResponsePartition[] }
}

declare interface KafkaListOffsetsResponsePartition {
  partition: number
  timestamp: number
  offset: number
  snapshot: {
    startOffset: number
    endOffset: number
  }
}

declare interface KafkaFindCoordinatorRequest {
  key: string
  keyType: number
}

declare interface KafkaFindCoordinatorResponse extends KafkaResponseError {
  host: string
  port: number
}

declare interface KafkaInitProducerIdRequest {
  transactionalId: string
	transactionTimeoutMs: number
	producerId: number
	producerEpoch: number
	enable2PC: boolean
}

declare interface KafkaInitProducerIdResponse extends KafkaResponseError {
  producerId: number
	producerEpoch: number
	ongoingTxnProducerId: number
	ongoingTxnProducerEpoch: number
}