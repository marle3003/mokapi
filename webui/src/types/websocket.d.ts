declare interface WebsocketService extends Service {
  channels: WebsocketChannel[];
  servers: WebsocketServer[];
  clients: WebsocketClient[];
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
  metrics: {
    websocket_messages_total: number
    websocket_message_timestamp: number
  }
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

declare interface WebsocketMessageData {
  channel: string
  message: WebsocketMessage;
  messageId: string
  client: WebsocketClientLog
  script: string
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
  serverAddress: string
}