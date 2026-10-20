declare interface MqttService extends Service {
  topics: MqttTopic[];
  servers: MqttServer[];
  clients: MqttClient[];
}

declare interface MqttServer {
  name: string;
  host: string;
  protocol: string
  title: string
  summary: string
  description: string;
  tags: MqttTag[]
  configs: { [key: string]: any }
}

declare interface MqttTag {
  name: string
  description: string
}

declare interface MqttTopic extends BaseChannel<MqttMessage> {
  metrics: {
    messages_total: number
    message_timestamp: number
  }
}

declare interface MqttMessage extends BaseMessage {
  key: SchemaFormat;
  header: SchemaFormat
}

declare interface MqttBroker {
  name: string;
  addr: string;
}

declare type MqttEventData = MqttMessageData | MqttRequestLog

declare type MqttEventInfo = MqttMessageInfo | MqttRequestLog

declare interface MqttMessageInfo {
  topic: string
  message: MqttMessage
}

declare interface MqttMessageData {
  topic: string
  message: MqttMessage;
  retain: boolean
  messageId: string
  clientId: string
  script: string
  actions: Action[]
}

declare interface MqttMessage {
  value?: string
  binary?: string
}

declare interface MqttClient {
  clientId: string
  address: string
  brokerAddress: string
  protocolVersion: number
}

declare interface MqttRequestLog {
  type: number
  request:  MqttConnectRequest | MqttSubscribeRequest | MqttDisconnecRequest
  response: MqttConnectResponse | MqttSubscribeResponse | undefined
}

declare interface MqttConnectRequest {
  version: number
  cleanSession: boolean
  keepAlive: number
  message?: MqttPublishMessage
  username?: string
  password?: string
}

declare interface MqttPublishMessage {
  qos: number
  retain: bool
  topic: string
  message: string
}

declare interface MqttConnectResponse {
  sessionPresent: boolean
  reasonCode: { code: number, reason: string }
}

declare interface MqttSubscribeRequest {
  messageId: number
  topics: { name: string, qos: number}[]
}

declare interface MqttSubscribeResponse {
  reasonCodes: number[]
}

declare interface MqttDisconnecRequest {
  reason: number
}