declare interface WebsocketService extends Service {
  channels: WebsocketChannel[];
  servers: WebsocketServer[];
  clients?: WebsocketClient[];
}

declare interface WebsocketServer {
  name: string;
  host: string;
  protocol: string
  title: string
  summary: string
  description: string;
  tags: WebsocketTag[]
  configs: { [key: string]: any }
}

declare interface WebsocketTag {
  name: string
  description: string
}

declare interface WebsocketChannel {
  name: string;
  title: string
  summary: string;
  description: string;
  messages: { [messageId: string]: WebsocketMessage }
  tags: WebsocketTag[]
  instances?: WebsocketChannelInstance[]
  metrics: {
    websocket_messages_total: number
    websocket_message_timestamp: number
  }
}

declare interface WebsocketChannelnstance {
  name: string
  parameters: Record<string, string>
}

declare interface WebsocketMessage {
  name: string
  title: string
  summary: string
  description: string
  key: SchemaFormat;
  payload: SchemaFormat;
  contentType: string;
}

declare interface WebsocketServer {
  name: string;
  addr: string;
}

declare type WebsocketEventData = WebsocketMessageData | WebsocketConnectionLog

declare interface WebsocketMessageData {
  api: string
  channel: string
  message: WebsocketMessage;
  messageId: string
  direction: 'receive' | 'send'
  client: WebsocketClientLog
  script: string
  actions: Action[]
  error?: string
}

declare interface WebsocketMessage {
  value?: string
  binary?: string
}

declare interface WebsocketClient {
  id: string
  query: Record<string, any>
  headers: Record<string, any>
  address: string
  serverAddress: string
}

declare interface WebsocketClientLog {
  id: string
  query: Record<string, any>
  headers: Record<string, any>
  address: string
  server: string
}

declare interface WebsocketConnectionLog {
  type: 'connect' | 'close'
  api: string
  channel: string
  client: WebsocketClientLog
  actions: Action[]
}

declare interface WebsocketCloseLog extends WebsocketConnectionLog {
  reason: string
  closedBy: string
}