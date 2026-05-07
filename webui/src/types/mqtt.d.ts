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

declare interface MqttTopic {
  name: string;
  description: string;
  messages: { [messageId: string]: MqttMessage }
  tags: MqttTag[]
  instances: MqttTopicInstance[]
}

declare interface MqttTopicInstance {
  name: string
  parameters: Record<string, string>
}

declare interface MqttMessage {
  name: string
  title: string
  summary: string
  description: string
  key: SchemaFormat;
  payload: SchemaFormat;
  header: SchemaFormat
  contentType: string;
}

declare interface MqttBroker {
  name: string;
  addr: string;
}

declare type MqttEventData = MqttMessageData | MqttRequestLog

declare interface MqttMessageData {
  topic: string
  message: MqttMessage;
  messageId: string
  clientId: string
  script: string
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